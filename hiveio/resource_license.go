package hiveio

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceLicenseCreate,
		Read:   resourceLicenseRead,
		Exists: resourceLicenseExists,
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

func resourceLicenseExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	clusterID, err := client.ClusterID()
	if err != nil {
		return false, err
	}
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		return false, err
	}
	_, _, err = cluster.GetLicenseInfo(client)

	if err != nil {
		return false, nil
	}
	return true, nil
}

func resourceLicenseDelete(d *schema.ResourceData, m interface{}) error {
	//Not supported, so just return nil
	return nil
}
