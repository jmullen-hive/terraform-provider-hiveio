package hiveio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceHostNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostNetworkCreate,
		ReadContext:   resourceHostNetworkRead,
		UpdateContext: resourceHostNetworkUpdate,
		DeleteContext: resourceHostNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "network name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostid": {
				Type:        schema.TypeString,
				Description: "id of the host for the network",
				Required:    true,
			},
			"interface": {
				Type:        schema.TypeString,
				Description: "physical network interface",
				Optional:    true,
			},
			"vlan": {
				Type:        schema.TypeInt,
				Description: "vlan to use for the network",
				Optional:    true,
			},
			"dhcp": {
				Type:        schema.TypeBool,
				Description: "enable dhcp",
				Optional:    true,
			},
			"ip": {
				Type:        schema.TypeString,
				Description: "ip address for the host",
				Optional:    true,
			},
			"mask": {
				Type:        schema.TypeString,
				Description: "netmask",
				Optional:    true,
			},
			// "dns": {
			// 	Type:        schema.TypeString,
			// 	Description: "dns server",
			// 	Default:     "",
			// 	Optional:    true,
			// },
			// "search": {
			// 	Type:        schema.TypeString,
			// 	Description: "dns search path",
			// 	Default:     "",
			// 	Optional:    true,
			// },
			"provider_override": &providerOverride,
		},
	}
}

func HostNetworkFromResource(d *schema.ResourceData) rest.HostNetwork {
	hostNetwork := rest.HostNetwork{
		Name: d.Get("name").(string),
	}
	if net, ok := d.GetOk("interface"); ok {
		hostNetwork.Interface = net.(string)
	}

	if vlan, ok := d.GetOk("vlan"); ok {
		hostNetwork.VLAN = vlan.(int)
	}

	if dhcp, ok := d.GetOk("dhcp"); ok {
		hostNetwork.DHCP = dhcp.(bool)
	}
	if ip, ok := d.GetOk("ip"); ok {
		hostNetwork.IP = ip.(string)
	}
	if mask, ok := d.GetOk("mask"); ok {
		hostNetwork.Mask = mask.(string)
	}
	return hostNetwork
}

func resourceHostNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	hostNetwork := HostNetworkFromResource(d)
	hostid := d.Get("hostid").(string)
	host, err := client.GetHost(hostid)
	if err != nil {
		return diag.FromErr(err)
	}
	err = host.SetNetwork(client, hostNetwork)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s", hostid, hostNetwork.Name))
	return resourceHostNetworkRead(ctx, d, m)
}

func resourceHostNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	host, err := client.GetHost(d.Get("hostid").(string))
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	hostNetwork, err := host.GetNetwork(client, d.Get("name").(string))
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s", host.Hostid, hostNetwork.Name))
	d.Set("interface", hostNetwork.Interface)
	d.Set("vlan", hostNetwork.VLAN)
	d.Set("dhcp", hostNetwork.DHCP)
	d.Set("ip", hostNetwork.IP)
	d.Set("mask", hostNetwork.Mask)
	return diag.Diagnostics{}
}

func resourceHostNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	hostNetwork := HostNetworkFromResource(d)
	hostid := d.Get("hostid").(string)
	host, err := client.GetHost(hostid)
	if err != nil {
		return diag.FromErr(err)
	}
	err = host.SetNetwork(client, hostNetwork)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceHostNetworkRead(ctx, d, m)
}

func resourceHostNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	host, err := client.GetHost(d.Get("hostid").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = host.DeleteNetwork(client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
