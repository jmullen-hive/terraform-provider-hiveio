package hiveio

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDiskCreate,
		ReadContext:   resourceDiskRead,
		DeleteContext: resourceDiskDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"filename": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_pool": {
				Description: "The storage id for where to store the disk.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"size": {
				Description: "Size of the disk in GB",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				ForceNew:    true,
			},
			"format": {
				Description: "File format (qcow2 or raw)",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "qcow2",
				ForceNew:    true,
			},
			"src_storage": {
				Description: "The storage pool id of an existing disk to copy.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"src_filename": {
				Description: "The filename of an existing disk to copy.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"src_url": {
				Description: "HTTP url for a disk to copy into the storage pool.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"local_file": {
				Description: "A local file to upload to the storage pool.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"backing_storage": {
				Description: "The storage pool id of an existing disk to use as a backing file.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"backing_filename": {
				Description: "The filename of an existing disk to use as a backing file.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"backing_format": {
				Description: "The format of an existing disk to use as a backing file.",
				Type:        schema.TypeString,
				Default:     "qcow2",
				Optional:    true,
				ForceNew:    true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func resourceDiskCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id := d.Get("storage_pool").(string)
	filename := d.Get("filename").(string)
	format := d.Get("format").(string)
	size := uint(d.Get("size").(int))

	srcPool, srcPoolOk := d.GetOk("src_storage")
	srcFilename, srcFileOk := d.GetOk("src_filename")
	srcURL, srcURLOk := d.GetOk("src_url")
	localFile, localFileOk := d.GetOk("local_file")

	var task *rest.Task
	var storage *rest.StoragePool
	storage, err = client.GetStoragePool(id)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := storage.DiskInfo(client, filename); err == nil {
		//disk already exists
		d.SetId(id + "-" + filename)
		return resourceDiskRead(ctx, d, m)
	}
	if localFileOk {
		err = storage.Upload(client, localFile.(string), filename)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if srcPoolOk && srcFileOk {
		var srcStorage *rest.StoragePool
		srcStorage, err = client.GetStoragePool(srcPool.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		task, err = srcStorage.ConvertDisk(client, srcFilename.(string), id, filename, format)
	} else if srcURLOk {
		task, err = storage.CopyURL(client, srcURL.(string), filename)
	} else {
		var backingFile *rest.StorageDisk
		backingStorage, backingStorageOk := d.GetOk("backing_storage")
		backingFilename, backingFilenameOk := d.GetOk("backing_filename")
		if backingStorageOk && backingFilenameOk {
			backingFile = &rest.StorageDisk{
				StorageID: backingStorage.(string),
				Filename:  backingFilename.(string),
				Format:    d.Get("backing_format").(string),
			}
		}
		task, err = storage.CreateDisk(client, filename, format, size, backingFile)
	}

	if err != nil {
		return diag.FromErr(err)
	}
	if task == nil {
		return diag.Errorf("Failed to create disk: Task was not returned")
	}

	task, err = task.WaitForTaskWithContext(ctx, client, false)
	if err != nil {
		return diag.FromErr(err)
	}
	if task.State == "failed" {
		return diag.Errorf("Failed to Create disk: %s", task.Message)
	}
	disk, err := storage.DiskInfo(client, filename)
	if err != nil {
		return diag.FromErr(err)
	}
	gbSize := disk.VirtualSize / 1024 / 1024 / 1024
	if size > gbSize {
		task, err = storage.GrowDisk(client, filename, size-gbSize)
		if err != nil {
			return diag.FromErr(err)
		}
		task, err = task.WaitForTaskWithContext(ctx, client, false)
		if err != nil {
			return diag.FromErr(err)
		}
		if task.State == "failed" {
			return diag.Errorf("Failed to resize disk: %s", task.Message)
		}
	}
	d.SetId(id + "-" + filename)
	return resourceDiskRead(ctx, d, m)
}

func resourceDiskRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id := d.Get("storage_pool").(string)
	filename := d.Get("filename").(string)
	storage, err := client.GetStoragePool(id)
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	disk, err := storage.DiskInfo(client, filename)
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	d.Set("format", disk.Format)
	return diag.Diagnostics{}
}

func resourceDiskDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	id := d.Get("storage_pool").(string)
	storage, err := client.GetStoragePool(id)
	if err != nil {
		return diag.FromErr(err)
	}
	err = storage.DeleteFile(client, d.Get("filename").(string))
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return diag.Diagnostics{}
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
