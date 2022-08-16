package hiveio

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostCreate,
		ReadContext:   resourceHostRead,
		UpdateContext: resourceHostUpdate,
		DeleteContext: resourceHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Default:  "admin",
				Optional: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"gateway_only": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"hostid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license": {
				Type:        schema.TypeString,
				Description: "unused field to add a license as a dependency",
				Optional:    true,
			},
		},
	}
}

func resourceHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	ip := d.Get("ip_address").(string)
	task, err := client.JoinHost(d.Get("username").(string), d.Get("password").(string), ip)
	if err != nil {
		return diag.FromErr(err)
	}
	task, err = task.WaitForTaskWithContext(ctx, client, false)
	if err != nil {
		return diag.FromErr(err)
	}
	hostid := task.Ref.Host
	if task.State == "failed" {
		return diag.Errorf("Failed to Add Host: %s", task.Message)
	}
	host, err := client.GetHost(hostid)
	if err != nil {
		return diag.FromErr(err)
	}
	state := "available"
	if d.Get("gateway_only").(bool) {
		state = "broker"
	}
	task, err = host.SetState(client, state)
	if err != nil {
		return diag.FromErr(err)
	}
	task, err = task.WaitForTaskWithContext(ctx, client, false)
	if err != nil {
		return diag.FromErr(err)
	}
	if task.State == "failed" {
		return diag.Errorf("Failed to set host state: %s", task.Message)
	}
	d.SetId(host.Hostid)
	return resourceHostRead(ctx, d, m)
}

func resourceHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	var host rest.Host
	var err error
	host, err = client.GetHost(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	d.Set("gateway_only", host.Appliance.Role == "gateway")
	d.Set("hostname", host.Hostname)
	d.Set("hostid", d.Id())
	return diag.Diagnostics{}
}

func resourceHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	_, err := client.GetHost(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	//Don't change anything for now
	return diag.Diagnostics{}
}

func resourceHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	host, err := client.GetHost(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if host.State != "maintenance" {
		task, err := host.SetState(client, "maintenance")
		if err != nil {
			return diag.FromErr(err)
		}
		task, err = task.WaitForTaskWithContext(ctx, client, false)
		if err != nil {
			return diag.FromErr(err)
		}
		if task.State == "failed" {
			return diag.Errorf("Failed to enter maintenance mode: %s", task.Message)
		}
	}
	//services might still be restarting from maintenance mode
	time.Sleep(10 * time.Second)
	task, err := host.UnjoinCluster(client)
	if err != nil {
		return diag.FromErr(err)
	}
	task, err = task.WaitForTaskWithContext(ctx, client, false)
	if err != nil {
		return diag.FromErr(err)
	}
	if task.State == "failed" {
		return diag.Errorf("Failed to remove host: %s", task.Message)
	}
	return diag.Diagnostics{}
}
