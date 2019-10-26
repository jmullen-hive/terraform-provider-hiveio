package hiveio

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func dataSourceStoragePool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStoragePoolRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"server": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"mount_options": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceStoragePoolRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	var err error

	id, idOk := d.GetOk("id")
	name, nameOk := d.GetOk("name")
	if idOk {
		storage, err = client.GetStoragePool(id.(string))
	} else if nameOk {
		storage, err = client.GetStoragePoolByName(name.(string))
	} else {
		return fmt.Errorf("id or name must be provided")
	}

	if err != nil {
		return err
	}
	d.SetId(storage.ID)
	d.Set("name", storage.Name)
	d.Set("server", storage.Server)
	d.Set("path", storage.Path)
	d.Set("type", storage.Type)
	d.Set("username", storage.Username)
	d.Set("roles", storage.Roles)
	return nil
}
