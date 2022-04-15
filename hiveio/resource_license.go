package hiveio

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceLicense() *schema.Resource {
	return &schema.Resource{
		Description: "Add a license for a new cluster",
		Create:      resourceLicenseCreate,
		Read:        resourceLicenseRead,
		Delete:      resourceLicenseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
		},
	}
}

func resourceLicenseCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	clusterID, err := client.ClusterID()
	if err != nil {
		return err
	}
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		return err
	}
	err = cluster.SetLicense(client, d.Get("license").(string))
	if err != nil {
		return err
	}
	d.SetId(d.Get("license").(string))
	return resourceLicenseRead(d, m)
}

func resourceLicenseRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	clusterID, err := client.ClusterID()
	if err != nil {
		return err
	}
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		return err
	}
	if cluster.License == nil {
		d.SetId("")
		return nil
	}
	d.Set("type", cluster.License.Type)
	d.Set("expiration", cluster.License.Expiration)
	d.Set("max_guests", cluster.License.MaxGuests)
	return nil
}

func resourceLicenseDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
