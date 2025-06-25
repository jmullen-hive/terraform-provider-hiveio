package hiveio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceGatewayHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGatewayHostCreate,
		ReadContext:   resourceGatewayHostRead,
		DeleteContext: resourceGatewayHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"hostid": {
				Description: "hostid of the host for the gateway connection",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"start_port": {
				Type:        schema.TypeInt,
				Description: "beginning of the port range for gateway connections",
				Required:    true,
				ForceNew:    true,
			},
			"end_port": {
				Type:        schema.TypeInt,
				Description: "end of the port range for gateway connections",
				Required:    true,
				ForceNew:    true,
			},
			"address": {
				Type:        schema.TypeString,
				Description: "the external address of the gateway",
				Required:    true,
				ForceNew:    true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func resourceGatewayHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.GetHost(d.Get("hostid").(string)); err != nil {
		return diag.Errorf("host %s not found", d.Get("hostid").(string))
	}
	hostid := d.Get("hostid").(string)
	clusterId, err := client.ClusterID()
	if err != nil {
		return diag.FromErr(err)
	}
	gateway, err := client.GetGateway(clusterId)
	if err != nil {
		return diag.FromErr(err)
	}
	if !gateway.Enabled {
		gateway.ClientSourceIsolation = true
		gateway.Enabled = true
	}
	host := rest.GatewayHost{
		StartPort:       uint(d.Get("start_port").(int)),
		EndPort:         uint(d.Get("end_port").(int)),
		ExternalAddress: d.Get("address").(string),
	}
	if gateway.Hosts == nil {
		gateway.Hosts = make(map[string]rest.GatewayHost)
	}
	gateway.Hosts[hostid] = host
	err = client.SetGateway(clusterId, gateway)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(hostid)
	return resourceGatewayHostRead(ctx, d, m)
}

func resourceGatewayHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	clusterId, err := client.ClusterID()
	if err != nil {
		return diag.FromErr(err)
	}
	gateway, err := client.GetGateway(clusterId)
	if err != nil {
		return diag.FromErr(err)
	}
	if gateway.Hosts == nil {
		d.SetId("")
		return diag.Diagnostics{}
	}

	host, ok := gateway.Hosts[d.Id()]
	if !ok {
		d.SetId("")
		return diag.Diagnostics{}
	}
	if err := d.Set("start_port", host.StartPort); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("end_port", host.EndPort); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("address", host.ExternalAddress); err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}

func resourceGatewayHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	hostid := d.Get("hostid").(string)
	clusterId, err := client.ClusterID()
	if err != nil {
		return diag.FromErr(err)
	}
	gateway, err := client.GetGateway(clusterId)
	if err != nil {
		return diag.FromErr(err)
	}
	if gateway.Hosts == nil {
		d.SetId("")
		return diag.Diagnostics{}
	}
	delete(gateway.Hosts, hostid)
	err = client.SetGateway(clusterId, gateway)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diag.Diagnostics{}
}
