package hiveio

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

/*type providerConfiguration struct {
	Client   *rest.Client
	Host     string
	Port     uint
	Username string
	Password string
	Realm    string
	Insecure bool
}*/

func Provider() *schema.Provider {
	return &schema.Provider{

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_USER", "admin"),
				Description: "username",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_PASS", nil),
				Description: "password",
				Sensitive:   true,
			},
			"realm": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_REALM", "local"),
				Description: "secret",
				Sensitive:   true,
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_HOST", "admin"),
				Description: "hostname or ip address of the server",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_PORT", "8443"),
				Description: "hostname or ip address of the server",
			},
			"hio_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.EnvDefaultFunc("HIO_INSECURE", false),
				Description: "Ignore SSL certificate errors",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			// "nutanix_image":           dataSourceNutanixImage(),
			// "nutanix_subnet":          dataSourceNutanixSubnet(),
			// "nutanix_cluster":         dataSourceNutanixCluster(),
			// "nutanix_clusters":        dataSourceNutanixClusters(),
			// "nutanix_virtual_machine": dataSourceNutanixVirtualMachine(),
			// "nutanix_category_key":    dataSourceNutanixCategoryKey(),
			// "nutanix_network_security_rule": dataSourceNutanixNetworkSecurityRule(),
			// "nutanix_volume_group":           dataSourceNutanixVolumeGroup(),
			// "nutanix_volume_groups":          dataSourceNutanixVolumeGroups(),
		},
		ResourcesMap: map[string]*schema.Resource{
			//"proxmox_vm_qemu": resourceVmQemu(),
			// TODO - storage_iso
			// TODO - bridge
			// TODO - vm_qemu_template
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	log.Printf("Connecting to %s", d.Get("host").(string))

	client := &rest.Client{Host: d.Get("host").(string), Port: d.Get("port").(uint), AllowInsecure: d.Get("insecure").(bool)}
	err := client.Login(d.Get("user").(string), d.Get("password").(string), d.Get("realm").(string))
	return client, err
}
