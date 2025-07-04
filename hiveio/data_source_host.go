package hiveio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func dataSourceHost() *schema.Resource {
	return &schema.Resource{
		Description: "A data source to retrieve host information by ip or hostname.",
		ReadContext: dataSourceHostRead,
		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"software_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func dataSourceHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	var host rest.Host

	if ip, ok := d.GetOk("ip"); ok {
		hosts, err := client.ListHosts("ip=" + ip.(string))
		if err != nil || len(hosts) != 1 {
			return diag.Errorf("Host not found")
		}
		host = hosts[0]
	} else if hostname, ok := d.GetOk("hostname"); ok {
		hosts, err := client.ListHosts("hostname=" + hostname.(string))
		if err != nil || len(hosts) != 1 {
			return diag.Errorf("Host not found")
		}
		host = hosts[0]
	} else {
		return diag.Errorf("ip_address or hostname must be provided")
	}
	d.Set("ip_address", host.IP)
	d.Set("hostname", host.Hostname)
	d.Set("hostid", host.Hostid)
	d.Set("cluster_id", host.Appliance.ClusterID)
	d.Set("software_version", host.Appliance.Firmware.Software)
	d.SetId(host.Hostid)
	return diag.Diagnostics{}
}
