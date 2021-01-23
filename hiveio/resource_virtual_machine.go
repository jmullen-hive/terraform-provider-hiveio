package hiveio

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"gpu": {
				Type:     schema.TypeBool,
				Default:  false,
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
			"inject_agent": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"state": {
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
							ForceNew: true,
						},
						"storage_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"filename": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
							ForceNew: true,
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
			"backup": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"frequency": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cloudinit_enabled": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"cloudinit_userdata": {
				Type:     schema.TypeString,
				Default:  "",
				Optional: true,
			},
			"cloudinit_networkconfig": {
				Type:     schema.TypeString,
				Default:  "",
				Optional: true,
			},
			"allowed_hosts": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func vmFromResource(d *schema.ResourceData) *rest.Pool {
	pool := rest.Pool{
		Name:        d.Get("name").(string),
		InjectAgent: d.Get("inject_agent").(bool),
		Type:        "standalone",
		Density:     []int{1, 1},
	}

	guestProfile := rest.PoolGuestProfile{
		OS:         d.Get("os").(string),
		Firmware:   d.Get("firmware").(string),
		Vga:        d.Get("display_driver").(string),
		Gpu:        d.Get("gpu").(bool),
		Persistent: true,
	}

	if cpu, ok := d.GetOk("cpu"); ok {
		guestProfile.CPU = []int{cpu.(int), cpu.(int)}
	}
	if mem, ok := d.GetOk("memory"); ok {
		guestProfile.Mem = []int{mem.(int), mem.(int)}
	}
	if cloudInitEnabled := d.Get("cloudinit_enabled").(bool); cloudInitEnabled {
		cloudInit := rest.PoolCloudInit{
			Enabled:       cloudInitEnabled,
			UserData:      d.Get("cloudinit_userdata").(string),
			NetworkConfig: d.Get("cloudinit_networkconfig").(string),
		}
		guestProfile.CloudInit = &cloudInit
	}
	pool.GuestProfile = &guestProfile

	if d.Id() != "" {
		pool.ID = d.Id()
	}

	var disks []*rest.PoolDisk
	for i := 0; i < d.Get("disk.#").(int); i++ {
		prefix := fmt.Sprintf("disk.%d.", i)
		disk := rest.PoolDisk{
			DiskDriver: d.Get(prefix + "disk_driver").(string),
			Type:       d.Get(prefix + "type").(string),
			StorageID:  d.Get(prefix + "storage_id").(string),
			Filename:   d.Get(prefix + "filename").(string),
		}
		disks = append(disks, &disk)
	}
	pool.GuestProfile.Disks = disks

	var interfaces []*rest.PoolInterface
	for i := 0; i < d.Get("interface.#").(int); i++ {
		prefix := fmt.Sprintf("interface.%d.", i)
		iface := rest.PoolInterface{
			Emulation: d.Get(prefix + "emulation").(string),
			Network:   d.Get(prefix + "network").(string),
			Vlan:      d.Get(prefix + "vlan").(int),
		}
		interfaces = append(interfaces, &iface)
	}
	pool.GuestProfile.Interfaces = interfaces

	if _, ok := d.GetOk("backup"); ok {
		var backup rest.PoolBackup
		backup.Enabled = d.Get("backup.0.enabled").(bool)
		backup.Frequency = d.Get("backup.0.frequency").(string)
		backup.TargetStorageID = d.Get("backup.0.target").(string)
		pool.Backup = &backup
	}

	if allowedHosts, ok := d.GetOk("allowed_hosts"); ok {
		var affinity rest.PoolAffinity
		hosts := make([]string, len(allowedHosts.([]interface{})))
		for i, host := range allowedHosts.([]interface{}) {
			hosts[i] = host.(string)
		}
		affinity.AllowedHostIDs = hosts
		pool.PoolAffinity = &affinity
	}

	return &pool
}

func resourceVMCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool := vmFromResource(d)

	_, err := pool.Create(client)
	if err != nil {
		return err
	}
	pool, err = client.GetPoolByName(pool.Name)
	if err != nil {
		return err
	}

	guestName := strings.ToUpper(pool.Name)
	guestName = strings.ReplaceAll(guestName, " ", "_")
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		guest, err := client.GetGuest(guestName)
		if err != nil {
			if strings.Contains(err.Error(), "\"error\": 404") {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(fmt.Errorf("Building pool %s", pool.ID))
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		}
		for _, v := range guest.TargetState {
			if v == guest.GuestState {
				return nil
			}
		}
		err = guest.WaitForGuest(client, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	d.SetId(pool.ID)
	return resourceVMRead(d, m)
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool, err := client.GetPool(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	d.Set("name", pool.Name)
	d.Set("cpu", pool.GuestProfile.CPU[0])
	d.Set("memory", pool.GuestProfile.Mem[0])
	d.Set("gpu", pool.GuestProfile.Gpu)
	d.Set("inject_agent", pool.InjectAgent)
	d.Set("state", pool.State)
	d.Set("os", pool.GuestProfile.OS)
	d.Set("firmware", pool.GuestProfile.Firmware)
	d.Set("display_driver", pool.GuestProfile.Vga)

	for i, disk := range pool.GuestProfile.Disks {
		prefix := fmt.Sprintf("disk.%d.", i)
		d.Set(prefix+"disk_driver", disk.DiskDriver)
		d.Set(prefix+"type", disk.Type)
		d.Set(prefix+"storage_id", disk.StorageID)
		d.Set(prefix+"filename", disk.Filename)
	}

	for i, iface := range pool.GuestProfile.Interfaces {
		prefix := fmt.Sprintf("interface.%d.", i)
		d.Set(prefix+"emulation", iface.Emulation)
		d.Set(prefix+"network", iface.Network)
		d.Set(prefix+"vlan", iface.Vlan)
	}

	if pool.GuestProfile.CloudInit != nil {
		d.Set("cloudinit_enabled", pool.GuestProfile.CloudInit.Enabled)
		d.Set("cloudinit_userdata", pool.GuestProfile.CloudInit.UserData)
		d.Set("cloudinit_networkconfig", pool.GuestProfile.CloudInit.NetworkConfig)
	}

	if pool.Backup != nil {
		d.Set("backup.0.enabled", pool.Backup.Enabled)
		d.Set("backup.0.frequency", pool.Backup.Frequency)
		d.Set("backup.0.target", pool.Backup.TargetStorageID)
	}

	if pool.PoolAffinity != nil && len(pool.PoolAffinity.AllowedHostIDs) > 0 {
		d.Set("allowed_hosts", pool.PoolAffinity.AllowedHostIDs)
	}

	return nil
}

func resourceVMUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool := vmFromResource(d)
	_, err := pool.Update(client)
	if err != nil {
		return err
	}
	return resourceVMRead(d, m)
}

func resourceVMDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool, err := client.GetPool(d.Id())
	if err != nil {
		return err
	}
	err = pool.Delete(client)
	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		pool, err := client.GetPool(d.Id())
		if err == nil && pool.State == "deleting" {
			time.Sleep(5 * time.Second)
			return resource.RetryableError(fmt.Errorf("Deleting pool %s", d.Id()))
		}
		if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
			time.Sleep(5 * time.Second)
			return nil
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
}
