package hiveio

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

var clients map[string]*rest.Client

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
	clients = make(map[string]*rest.Client)
}

var providerSchema = map[string]*schema.Schema{
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
}

var providerOverride = schema.Schema{
	Type:        schema.TypeList,
	Description: "Override the provider configuration for this resource.  This can be used to connect to a different cluster or change credentials",
	Optional:    true,
	MaxItems:    1,
	ForceNew:    true,
	Elem: &schema.Resource{
		Schema: providerSchema,
	},
}

func getClient(d *schema.ResourceData, m interface{}) (*rest.Client, error) {
	if override, ok := d.GetOk("provider_override"); ok {
		settings := override.([]interface{})[0].(map[string]interface{})
		host := settings["host"].(string)
		port := settings["port"].(int)
		username := settings["username"].(string)
		password := settings["password"].(string)
		realm := settings["realm"].(string)
		allowInsecure := settings["insecure"].(bool)
		key := fmt.Sprintf("%s:%d:%s:%s:%t", host, port, username, realm, allowInsecure)
		if client, ok := clients[key]; ok {
			return client, nil
		}
		var client = &rest.Client{
			Host:          host,
			Port:          uint(port),
			AllowInsecure: allowInsecure,
		}

		err := client.Login(
			username,
			password,
			realm,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to login with provider override: %w", err)
		}
		clients[key] = client
		return client, nil
	}
	client, ok := m.(*rest.Client)
	if !ok {
		return nil, fmt.Errorf("expected *rest.Client, got %T", m)
	}

	return client, nil

}

// Provider hiveio terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{

		Schema: providerSchema,
		DataSourcesMap: map[string]*schema.Resource{
			"hiveio_profile":      dataSourceProfile(),
			"hiveio_storage_pool": dataSourceStoragePool(),
			"hiveio_host":         dataSourceHost(),
			"hiveio_host_network": dataSourceHostNetwork(),
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
			"hiveio_user":            resourceUser(),
			"hiveio_shared_storage":  resourceSharedStorage(),
			"hiveio_host_network":    resourceHostNetwork(),
			"hiveio_host_iscsi":      resourceHostIscsi(),
			"hiveio_gateway_host":    resourceGatewayHost(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	log.Printf("Connecting to %s", d.Get("host").(string))

	client := &rest.Client{
		Host:          d.Get("host").(string),
		Port:          uint(d.Get("port").(int)),
		AllowInsecure: d.Get("insecure").(bool),
	}
	err := client.Login(d.Get("username").(string), d.Get("password").(string), d.Get("realm").(string))
	if err == nil {
		key := fmt.Sprintf("%s:%d:%s:%s:%t", client.Host, client.Port, d.Get("username").(string), d.Get("realm").(string), client.AllowInsecure)
		clients[key] = client
	}
	return client, err
}
