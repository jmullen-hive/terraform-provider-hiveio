package hiveio

import (
	"sort"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

/*type Disk struct {
	ID           string   `json:"id,omitempty"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Server       string   `json:"server"`
	Path         string   `json:"path"`
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	Key          string   `json:"key,omitempty"`
	MountOptions []string `json:"mountOptions,omitempty"`
	Roles        []string `json:"roles,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}*/

func resourceDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceDiskCreate,
		Read:   resourceDiskRead,
		Exists: resourceDiskExists,
		Delete: resourceDiskDelete,

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
		},
	}
}

func resourceDiskCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	id := d.Get("storage_pool").(string)
	storage, err := client.GetStoragePool(id)
	if err != nil {
		return err
	}
	filename := d.Get("filename").(string)
	format := d.Get("format").(string)
	size := uint(d.Get("size").(int))
	err = storage.CreateDisk(client, filename, format, size)
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
