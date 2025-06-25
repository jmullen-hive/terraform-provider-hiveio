package hiveio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLicense() *schema.Resource {
	return &schema.Resource{
		Description:   "Add a license for a new cluster",
		CreateContext: resourceLicenseCreate,
		ReadContext:   resourceLicenseRead,
		DeleteContext: resourceLicenseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"license": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_guests": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func resourceLicenseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	clusterID, err := client.ClusterID()
	if err != nil {
		return diag.FromErr(err)
	}
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = cluster.SetLicense(client, d.Get("license").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(d.Get("license").(string))
	return resourceLicenseRead(ctx, d, m)
}

func resourceLicenseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	clusterID, err := client.ClusterID()
	if err != nil {
		return diag.FromErr(err)
	}
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		return diag.FromErr(err)
	}
	if cluster.License == nil {
		d.SetId("")
		return diag.Diagnostics{}
	}
	d.Set("type", cluster.License.Type)
	d.Set("expiration", cluster.License.Expiration)
	d.Set("max_guests", cluster.License.MaxGuests)
	return diag.Diagnostics{}
}

func resourceLicenseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Diagnostics{}
}
