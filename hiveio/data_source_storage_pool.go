package hiveio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func dataSourceStoragePool() *schema.Resource {
	return &schema.Resource{
		Description: "The storage pool data source can be used to retrieve settings from an existing storage pool",
		ReadContext: dataSourceStoragePoolRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
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

func dataSourceStoragePoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.Errorf("id or name must be provided")
	}

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(storage.ID)
	d.Set("name", storage.Name)
	d.Set("server", storage.Server)
	d.Set("path", storage.Path)
	d.Set("type", storage.Type)
	d.Set("username", storage.Username)
	d.Set("roles", storage.Roles)
	return diag.Diagnostics{}
}
