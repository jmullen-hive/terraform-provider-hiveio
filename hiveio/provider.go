package hiveio

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

//Provider hiveio terraform provider
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
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.EnvDefaultFunc("HIO_INSECURE", false),
				Description: "Ignore SSL certificate errors",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hiveio_profile":      dataSourceProfile(),
			"hiveio_storage_pool": dataSourceStoragePool(),
			"hiveio_host":         dataSourceHost(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"hiveio_host":            resourceHost(),
			"hiveio_realm":           resourceRealm(),
			"hiveio_profile":         resourceProfile(),
			"hiveio_storage_pool":    resourceStoragePool(),
			"hiveio_disk":            resourceDisk(),
			"hiveio_template":        resourceTemplate(),
			"hiveio_guest_pool":      resourceGuestPool(),
			"hiveio_virtual_machine": resourceVM(),
			"hiveio_license":         resourceLicense(),
			"hiveio_external_guest":  resourceExternalGuest(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	log.Printf("Connecting to %s", d.Get("host").(string))

	client := &rest.Client{Host: d.Get("host").(string), Port: uint(d.Get("port").(int)), AllowInsecure: d.Get("insecure").(bool)}
	err := client.Login(d.Get("username").(string), d.Get("password").(string), d.Get("realm").(string))
	fmt.Println(err)
	return client, err
}
