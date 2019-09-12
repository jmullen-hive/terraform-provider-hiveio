package hiveio

import (
	"sort"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
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

	var err error
	if srcPoolOk && srcFileOk {
		storage, err := client.GetStoragePool(srcPool.(string))
		if err != nil {
			return err
		}
		err = storage.ConvertDisk(client, srcFilename.(string), id, filename, format)
	} else {
		storage, err := client.GetStoragePool(id)
		if err != nil {
			return err
		}
		err = storage.CreateDisk(client, filename, format, size)
	}
	if err == nil {
		d.SetId(id + "-" + filename)
	}
	return nil
}

func resourceDiskRead(d *schema.ResourceData, m interface{}) error {
	//TODO: Add virsh vol-info to hive-rest
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
	files, err := storage.Browse(client)
	if err != nil {
		return false, err
	}
	if sort.SearchStrings(files, d.Get("filename").(string)) == len(files) {
		return false, nil
	}
	return true, nil
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
