package hiveio

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func dataSourceHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHostRead,
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
		},
	}
}

func dataSourceHostRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var host rest.Host
	ip, ipOk := d.GetOk("ip")
	hostname, hostnameOk := d.GetOk("hostname")

	if ipOk {
		hosts, err := client.ListHosts("ip=" + ip.(string))
		if err != nil || len(hosts) != 1 {
			return fmt.Errorf("Host not found")
		}
		host = hosts[0]
	} else if hostnameOk {
		hosts, err := client.ListHosts("hostname=" + hostname.(string))
		if err != nil || len(hosts) != 1 {
			return fmt.Errorf("Host not found")
		}
		host = hosts[0]
	} else {
		return fmt.Errorf("ip_address or hostname must be provided")
	}
	d.Set("ip_address", host.IP)
	d.Set("hostname", host.Hostname)
	d.Set("hostid", host.Hostid)
	d.Set("cluster_id", host.Appliance.ClusterID)
	d.Set("software_version", host.Appliance.Firmware.Software)
	return nil
}
