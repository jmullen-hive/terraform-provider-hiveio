package hiveio

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceLicenseCreate,
		Read:   resourceLicenseRead,
		Delete: resourceLicenseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"license": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	return nil
}

func resourceLicenseRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceLicenseDelete(d *schema.ResourceData, m interface{}) error {
	//Not supported, so just return nil
	return nil
}
