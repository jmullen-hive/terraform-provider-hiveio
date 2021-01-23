package hiveio

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		if s.Deprecated != "" {
			desc += " " + s.Deprecated
		}
		return strings.TrimSpace(desc)
	}

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

//Provider hiveio terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_USER", "admin"),
				Description: "The username to connect to the server. Defaults to admin",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_PASS", nil),
				Description: "The password to use for connection to the server.",
				Sensitive:   true,
			},
			"realm": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_REALM", "local"),
				Description: "The realm to use to connect to the server. Defaults to local",
				Sensitive:   true,
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_HOST", "admin"),
				Description: "hostname or ip address of the server.",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HIO_PORT", "8443"),
				Description: "The port to use to connect to the server. Defaults to 8443",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.EnvDefaultFunc("HIO_INSECURE", false),
				Description: "Ignore SSL certificate errors.",
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
