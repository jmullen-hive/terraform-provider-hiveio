package hiveio

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostCreate,
		Read:   resourceHostRead,
		Exists: resourceHostExists,
		Update: resourceHostUpdate,
		Delete: resourceHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Description: "username",
				Default:     "admin",
				Optional:    true,
			},
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Description: "password",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceHostCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	ip := d.Get("ip_address").(string)
	task, err := client.JoinHost(d.Get("username").(string), d.Get("password").(string), ip)
	if err != nil {
		return err
	}
	task = task.WaitForTask(client, false)
	hostid := task.Ref.Host
	if task.State == "failed" {
		return fmt.Errorf("Failed to Create disk: %s", task.Message)
	}
	host, err := client.GetHost(hostid)
	if err != nil {
		return err
	}
	d.SetId(host.Hostid)
	return resourceHostRead(d, m)
}

func resourceHostRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var Host rest.Host
	var err error
	Host, err = client.GetHost(d.Id())
	if err != nil {
		return err
	}
	d.Set("hostname", Host.Hostname)
	return nil
}

func resourceHostExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	id := d.Id()
	_, err := client.GetHost(id)

	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func resourceHostUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	_, err := client.GetHost(d.Id())
	if err != nil {
		return err
	}
	//Don't change anything for now
	return nil
}

func resourceHostDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	host, err := client.GetHost(d.Id())
	if err != nil {
		return err
	}
	return host.UnjoinCluster(client)
}
