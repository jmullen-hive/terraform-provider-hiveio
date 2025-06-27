package hiveio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVersion() *schema.Resource {
	return &schema.Resource{
		Description: "A data source to retrieve host information by ip or hostname.",
		ReadContext: dataSourceVersionRead,
		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"major": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"minor": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"patch": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func dataSourceVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	version, err := client.HostVersion()
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("version", version.Version)
	d.Set("major", version.Major)
	d.Set("minor", version.Minor)
	d.Set("patch", version.Patch)

	d.SetId(version.Version)
	return diag.Diagnostics{}
}
