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
				Optional: true,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
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
			"log_level": {
				Type:        schema.TypeString,
				Description: "set the host log level",
				Optional:    true,
				Computed:    true,
			},
			"max_clone_density": {
				Type:        schema.TypeInt,
				Description: "set the max clone density for the host",
				Optional:    true,
				Computed:    true,
			},
			"timezone": {
				Type:        schema.TypeString,
				Description: "set the timezone for the host",
				Optional:    true,
				Computed:    true,
			},
			"ntp_servers": {
				Type:        schema.TypeString,
				Description: "set the ntp servers for the host as a comma separated list",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	var hostIP string
	if ip, ok := d.GetOk("ip"); ok {
		hostIP = ip.(string)
	} else if hostname, ok := d.GetOk("hostname"); ok {
		//try adding by hostname
		hostIP = hostname.(string)
	} else {
		return diag.Errorf("ip_address or hostname must be provided")
	}
	hosts, err := client.ListHosts("")
	if err != nil {
		return diag.FromErr(err)
	}
	for _, host := range hosts {
		if host.IP == hostIP || host.Hostname == hostIP {
			d.SetId(host.Hostid)
			return resourceHostRead(ctx, d, m)
		}
	}
	task, err := client.JoinHost(d.Get("username").(string), d.Get("password").(string), hostIP)
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
	// Add a delay to ensure the host is fully joined
	time.Sleep(5 * time.Second)
	if gatewayOnly, ok := d.Get("gateway_only").(bool); ok && gatewayOnly {
		if gatewayOnly && host.Appliance.Role != "gateway" {
			host.ChangeGatewayMode(client, true)
		} else if !gatewayOnly && host.Appliance.Role == "gateway" {
			host.ChangeGatewayMode(client, false)
		}
	} else {
		task, err = host.SetState(client, "available")
		if err != nil {
			return diag.FromErr(err)
		}
	}

	task, err = task.WaitForTaskWithContext(ctx, client, false)
	if err != nil {
		return diag.FromErr(err)
	}
	if task.State == "failed" {
		return diag.Errorf("Failed to set host state: %s", task.Message)
	}
	time.Sleep(10 * time.Second)
	host, err = client.GetHost(hostid)
	if err != nil {
		return diag.FromErr(err)
	}
	updateAppliance := false
	if logLevel, ok := d.Get("log_level").(string); ok {
		if logLevel != "" && host.Appliance.Loglevel != logLevel {
			updateAppliance = true
			host.Appliance.Loglevel = logLevel
		}
	}
	if mcd, ok := d.Get("max_clone_density").(int); ok {
		if mcd != 0 && host.Appliance.MaxCloneDensity != mcd {
			updateAppliance = true
			host.Appliance.MaxCloneDensity = mcd
		}
	}
	if ntpServers, ok := d.Get("ntp_servers").(string); ok {
		if host.Appliance.Ntp != ntpServers {
			updateAppliance = true
			host.Appliance.Ntp = ntpServers
		}
	}
	if timezone, ok := d.Get("timezone").(string); ok {
		if host.Appliance.Timezone != timezone {
			updateAppliance = true
			host.Appliance.Timezone = timezone
		}
	}

	if updateAppliance {
		_, err = host.UpdateAppliance(client)
		if err != nil {
			return diag.FromErr(err)
		}
		//wait for appliance configure task to finish since the resonse does not include the id
		time.Sleep(5 * time.Second)
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
	d.Set("ip_address", host.IP)
	d.Set("cluster_id", host.Appliance.ClusterID)
	d.Set("log_level", host.Appliance.Loglevel)
	d.Set("max_clone_density", host.Appliance.MaxCloneDensity)
	d.Set("ntp_servers", host.Appliance.Ntp)
	d.Set("timezone", host.Appliance.Timezone)
	return diag.Diagnostics{}
}

func resourceHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	host, err := client.GetHost(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	updateAppliance := false
	if logLevel, ok := d.Get("log_level").(string); ok {
		if logLevel != "" && host.Appliance.Loglevel != logLevel {
			updateAppliance = true
			host.Appliance.Loglevel = logLevel
		}
	}
	if mcd, ok := d.Get("max_clone_density").(int); ok {
		if mcd != 0 && host.Appliance.MaxCloneDensity != mcd {
			updateAppliance = true
			host.Appliance.MaxCloneDensity = mcd
		}
	}
	if ntpServers, ok := d.Get("ntp_servers").(string); ok {
		if host.Appliance.Ntp != ntpServers {
			updateAppliance = true
			host.Appliance.Ntp = ntpServers
		}
	}
	if timezone, ok := d.Get("timezone").(string); ok {
		if host.Appliance.Timezone != timezone {
			updateAppliance = true
			host.Appliance.Timezone = timezone
		}
	}

	if updateAppliance {
		_, err = host.UpdateAppliance(client)
		if err != nil {
			return diag.FromErr(err)
		}
		//sleep since the api doesn't return the task
		time.Sleep(5 * time.Second)
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
	if host.State == "unreachable" {
		//Host is unreachable, just delete the record
		err = host.Delete(client)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Diagnostics{}
	}

	if host.State == "available" {
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
		//services might still be restarting from maintenance mode
		time.Sleep(10 * time.Second)
	}

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
