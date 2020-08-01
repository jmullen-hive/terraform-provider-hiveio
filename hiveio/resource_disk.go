package hiveio

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceDiskCreate,
		Read:   resourceDiskRead,
		Exists: resourceDiskExists,
		Delete: resourceDiskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"filename": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_pool": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
				ForceNew: true,
			},
			"format": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "qcow2",
				ForceNew: true,
			},
			"src_storage": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"src_filename": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"src_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"local_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDiskCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	id := d.Get("storage_pool").(string)
	filename := d.Get("filename").(string)
	format := d.Get("format").(string)
	size := uint(d.Get("size").(int))

	srcPool, srcPoolOk := d.GetOk("src_storage")
	srcFilename, srcFileOk := d.GetOk("src_filename")
	srcURL, srcURLOk := d.GetOk("src_url")
	localFile, localFileOk := d.GetOk("local_file")

	var err error
	var task *rest.Task
	var storage *rest.StoragePool
	storage, err = client.GetStoragePool(id)
	if err != nil {
		return err
	}
	if localFileOk {
		err = storage.Upload(client, localFile.(string), filename)
	}
	if srcPoolOk && srcFileOk {
		srcStorage, err := client.GetStoragePool(srcPool.(string))
		if err != nil {
			return err
		}
		task, err = srcStorage.ConvertDisk(client, srcFilename.(string), id, filename, format)
	} else if srcURLOk {
		task, err = storage.CopyURL(client, srcURL.(string), filename)
	} else {
		task, err = storage.CreateDisk(client, filename, format, size)
	}
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("Failed to create disk: Task was not returned")
	}
	task = task.WaitForTask(client, false)
	if task.State == "failed" {
		return fmt.Errorf("Failed to Create disk: %s, %s", task.Message)
	}
	disk, err := storage.DiskInfo(client, filename)
	if err != nil {
		return err
	}
	gbSize := disk.VirtualSize / 1024 / 1024 / 1024
	if (size - gbSize) > 0 {
		task, err = storage.GrowDisk(client, filename, size-gbSize)
		if err != nil {
			return err
		}
		task = task.WaitForTask(client, false)
		if task.State == "failed" {
			return fmt.Errorf("Failed to resize disk: %s, %s", task.Message)
		}
	}
	d.SetId(id + "-" + filename)
	return resourceDiskRead(d, m)
}

func resourceDiskRead(d *schema.ResourceData, m interface{}) error {
	//TODO: Add qemu-img info to hive-rest
	client := m.(*rest.Client)
	id := d.Get("storage_pool").(string)
	filename := d.Get("filename").(string)
	storage, err := client.GetStoragePool(id)
	if err != nil {
		return err
	}
	disk, err := storage.DiskInfo(client, filename)
	if err != nil {
		return err
	}
	d.Set("size", disk.VirtualSize/1024/1024/1024)
	d.Set("format", disk.Format)
	return nil
}

func resourceDiskExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	id := d.Get("storage_pool").(string)
	storage, err := client.GetStoragePool(id)
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	} else if err != nil {
		return false, err
	}
	_, err = storage.DiskInfo(client, d.Get("filename").(string))
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	}
	return true, err
}

func resourceDiskDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	id := d.Get("storage_pool").(string)
	storage, err := client.GetStoragePool(id)
	if err != nil {
		return err
	}
	return storage.DeleteFile(client, d.Get("filename").(string))
}
