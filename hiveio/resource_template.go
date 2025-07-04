package hiveio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTemplateCreate,
		ReadContext:   resourceTemplateRead,
		UpdateContext: resourceTemplateUpdate,
		DeleteContext: resourceTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Default:  2,
				Optional: true,
			},
			"mem": {
				Type:     schema.TypeInt,
				Default:  2048,
				Optional: true,
			},
			"firmware": {
				Type:     schema.TypeString,
				Default:  "uefi",
				Optional: true,
			},
			"display_driver": {
				Type:     schema.TypeString,
				Default:  "cirrus",
				Optional: true,
			},
			"os": {
				Type:     schema.TypeString,
				Required: true,
			},
			"manual_agent_install": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Default:  "Disk",
							Optional: true,
						},
						"storage_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"filename": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk_driver": {
							Type:     schema.TypeString,
							Default:  "virtio",
							Optional: true,
						},
						"format": {
							Type:     schema.TypeString,
							Default:  "qcow2",
							Optional: true,
						},
						"size": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"interface": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:     schema.TypeString,
							Required: true,
						},
						"vlan": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"emulation": {
							Type:     schema.TypeString,
							Default:  "virtio",
							Optional: true,
						},
					},
				},
			},
			"broker_default_connection": {
				Type:     schema.TypeString,
				Default:  "",
				Optional: true,
			},
			"broker_connection": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Default:  "",
							Optional: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disable_html5": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"gateway": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disabled": {
										Type:     schema.TypeBool,
										Default:  false,
										Optional: true,
									},
									"persistent": {
										Type:     schema.TypeBool,
										Default:  false,
										Optional: true,
									},
									"protocols": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"provider_override": &providerOverride,
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(time.Minute),
		},
	}
}

func templateFromResource(d *schema.ResourceData) rest.Template {
	template := rest.Template{
		Name:               d.Get("name").(string),
		Vcpu:               d.Get("cpu").(int),
		Mem:                d.Get("mem").(int),
		Firmware:           d.Get("firmware").(string),
		DisplayDriver:      d.Get("display_driver").(string),
		OS:                 d.Get("os").(string),
		ManualAgentInstall: d.Get("manual_agent_install").(bool),
	}

	if d.Id() != "" {
		template.Name = d.Id()
	}

	var disks []*rest.TemplateDisk
	for i := 0; i < d.Get("disk.#").(int); i++ {
		prefix := fmt.Sprintf("disk.%d.", i)
		disk := rest.TemplateDisk{
			DiskDriver: d.Get(prefix + "disk_driver").(string),
			Type:       d.Get(prefix + "type").(string),
			StorageID:  d.Get(prefix + "storage_id").(string),
			Filename:   d.Get(prefix + "filename").(string),
			Format:     d.Get(prefix + "format").(string),
		}
		disks = append(disks, &disk)
	}
	template.Disks = disks

	var interfaces []*rest.TemplateInterface
	for i := 0; i < d.Get("interface.#").(int); i++ {
		prefix := fmt.Sprintf("interface.%d.", i)
		iface := rest.TemplateInterface{
			Emulation: d.Get(prefix + "emulation").(string),
			Network:   d.Get(prefix + "network").(string),
			Vlan:      d.Get(prefix + "vlan").(int),
		}
		interfaces = append(interfaces, &iface)
	}
	template.Interfaces = interfaces
	if nConnections, ok := d.Get("broker_connection.#").(int); ok && nConnections > 0 {
		template.BrokerOptions = &rest.GuestBrokerOptions{}
		template.BrokerOptions.DefaultConnection = d.Get("broker_default_connection").(string)
		var connections []rest.GuestBrokerConnection
		for i := 0; i < nConnections; i++ {
			prefix := fmt.Sprintf("broker_connection.%d.", i)
			connection := rest.GuestBrokerConnection{
				Name:         d.Get(prefix + "name").(string),
				Description:  d.Get(prefix + "description").(string),
				Port:         uint(d.Get(prefix + "port").(int)),
				Protocol:     d.Get(prefix + "protocol").(string),
				DisableHtml5: d.Get(prefix + "disable_html5").(bool),
			}
			connection.Gateway.Disabled = d.Get(prefix + "gateway.0." + "disabled").(bool)
			connection.Gateway.Persistent = d.Get(prefix + "gateway.0." + "persistent").(bool)
			if protocolsInterface, ok := d.GetOk(prefix + "gateway.0." + "protocols"); ok {
				protocols := make([]string, len(protocolsInterface.([]interface{})))
				for i, protocol := range protocolsInterface.([]interface{}) {
					protocols[i] = protocol.(string)
				}
				connection.Gateway.Protocols = protocols
			}
			connections = append(connections, connection)
		}
		template.BrokerOptions.Connections = connections
	}

	return template
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	template := templateFromResource(d)
	_, err = template.Create(client)
	if err != nil {
		return diag.FromErr(err)
	}
	time.Sleep(5 * time.Second)
	template, err = client.GetTemplate(template.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	if template.State != "available" {
		return diag.FromErr(fmt.Errorf("template %s is not available", template.Name))
	}
	d.SetId(template.Name)
	return resourceTemplateRead(ctx, d, m)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	template, err := client.GetTemplate(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", template.Name)
	d.Set("cpu", template.Vcpu)
	d.Set("mem", template.Mem)
	d.Set("firmware", template.Firmware)
	d.Set("display_driver", template.DisplayDriver)
	d.Set("os", template.OS)
	d.Set("manual_agent_install", template.ManualAgentInstall)

	disks := make([]map[string]interface{}, len(template.Disks))
	for i, disk := range template.Disks {
		disks[i] = map[string]interface{}{
			"disk_driver": disk.DiskDriver,
			"type":        disk.Type,
			"storage_id":  disk.StorageID,
			"filename":    disk.Filename,
			"format":      disk.Format,
		}
	}
	d.Set("disk", disks)

	interfaces := make([]map[string]interface{}, len(template.Interfaces))
	for i, iface := range template.Interfaces {
		interfaces[i] = map[string]interface{}{
			"emulation": iface.Emulation,
			"network":   iface.Network,
			"vlan":      iface.Vlan,
		}
	}
	d.Set("interface", interfaces)

	if template.BrokerOptions != nil {
		d.Set("broker_default_connection", template.BrokerOptions.DefaultConnection)
		connections := make([]map[string]interface{}, len(template.BrokerOptions.Connections))
		for i, connection := range template.BrokerOptions.Connections {
			connections[i] = map[string]interface{}{
				"name":          connection.Name,
				"description":   connection.Description,
				"port":          connection.Port,
				"protocol":      connection.Protocol,
				"disable_html5": connection.DisableHtml5,
				"gateway": []map[string]interface{}{
					{
						"disabled":   connection.Gateway.Disabled,
						"persistent": connection.Gateway.Persistent,
						"protocols":  connection.Gateway.Protocols,
					},
				},
			}
		}
		d.Set("broker_connection", connections)
	}

	return diag.Diagnostics{}
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	template := templateFromResource(d)
	_, err = template.Update(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceTemplateRead(ctx, d, m)
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	template, err := client.GetTemplate(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = template.Delete(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
