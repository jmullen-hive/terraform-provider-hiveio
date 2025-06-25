package hiveio

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHostNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHostNetworkRead,

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
				Computed:    true,
			},
			"vlan": {
				Type:        schema.TypeInt,
				Description: "vlan to use for the network",
				Computed:    true,
			},
			"dhcp": {
				Type:        schema.TypeBool,
				Description: "enable dhcp",
				Computed:    true,
			},
			"ip": {
				Type:        schema.TypeString,
				Description: "ip address for the host",
				Computed:    true,
			},
			"mask": {
				Type:        schema.TypeString,
				Description: "netmask",
				Computed:    true,
			},
			"dns": {
				Type:        schema.TypeString,
				Description: "dns server",
				Computed:    true,
			},
			"search": {
				Type:        schema.TypeString,
				Description: "dns search path",
				Computed:    true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func dataSourceHostNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	d.SetId(hostNetwork.Name)
	d.Set("interface", hostNetwork.Interface)
	d.Set("vlan", hostNetwork.VLAN)
	d.Set("dhcp", hostNetwork.DHCP)
	d.Set("ip", hostNetwork.IP)
	d.Set("mask", hostNetwork.Mask)
	d.Set("dns", hostNetwork.DNS)
	d.Set("search", hostNetwork.Search)
	return diag.Diagnostics{}
}
