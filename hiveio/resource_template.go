package hiveio

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Exists: resourceTemplateExists,
		Update: resourceTemplateUpdate,
		Delete: resourceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  2,
				Optional: true,
			},
			"mem": &schema.Schema{
				Type:     schema.TypeInt,
				Default:  2048,
				Optional: true,
			},
			"firmware": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "uefi",
				Optional: true,
			},
			"display_driver": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "cirrus",
				Optional: true,
			},
			"os": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"manual_agent_install": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_message": &schema.Schema{
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
							Required: true,
						},
						"emulation": {
							Type:     schema.TypeString,
							Default:  "virtio",
							Optional: true,
						},
					},
				},
			},
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

	return template
}

func resourceTemplateCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	template := templateFromResource(d)
	_, err := template.Create(client)
	if err != nil {
		return err
	}
	d.SetId(template.Name)
	return resourceTemplateRead(d, m)
}

func resourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	template, err := client.GetTemplate(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", template.Name)
	d.Set("cpu", template.Vcpu)
	d.Set("mem", template.Mem)
	d.Set("firmware", template.Firmware)
	d.Set("display_driver", template.DisplayDriver)
	d.Set("os", template.OS)
	d.Set("manual_agent_install", template.ManualAgentInstall)

	for i, disk := range template.Disks {
		prefix := fmt.Sprintf("disk.%d.", i)
		d.Set(prefix+"disk_driver", disk.DiskDriver)
		d.Set(prefix+"type", disk.Type)
		d.Set(prefix+"storage_id", disk.StorageID)
		d.Set(prefix+"filename", disk.Filename)
		d.Set(prefix+"format", disk.Format)
	}

	for i, iface := range template.Interfaces {
		prefix := fmt.Sprintf("interface.%d.", i)
		d.Set(prefix+"emulation", iface.Emulation)
		d.Set(prefix+"network", iface.Network)
		d.Set(prefix+"vlan", iface.Vlan)
	}

	return nil
}

func resourceTemplateExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var err error
	name := d.Id()
	_, err = client.GetTemplate(name)
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func resourceTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	template := templateFromResource(d)
	_, err := template.Update(client)
	if err != nil {
		return err
	}
	return resourceTemplateRead(d, m)
}

func resourceTemplateDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	template, err := client.GetTemplate(d.Id())
	if err != nil {
		return err
	}
	return template.Delete(client)
}
