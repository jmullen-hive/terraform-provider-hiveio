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
				Required: true,
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

	var err error
	var task *rest.Task
	if srcPoolOk && srcFileOk {
		storage, err := client.GetStoragePool(srcPool.(string))
		if err != nil {
			return err
		}
		task, err = storage.ConvertDisk(client, srcFilename.(string), id, filename, format)
	} else if srcURLOk {
		storage, err := client.GetStoragePool(id)
		if err != nil {
			return err
		}
		task, err = storage.CopyURL(client, srcURL.(string), filename)
	} else {
		storage, err := client.GetStoragePool(id)
		if err != nil {
			return err
		}
		task, err = storage.CreateDisk(client, filename, format, size)
	}
	if err != nil {
		return err
	}

	task = task.WaitForTask(client, false)
	if task.State == "completed" {
		d.SetId(id + "-" + filename)
	} else if task.State == "failed" {
		return fmt.Errorf("Failed to Create disk: %s", task.Message)
	}
	return nil
}

func resourceDiskRead(d *schema.ResourceData, m interface{}) error {
	//TODO: Add qemu-img info to hive-rest
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
