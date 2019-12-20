package hiveio

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceGuestPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceGuestPoolCreate,
		Read:   resourceGuestPoolRead,
		Exists: resourceGuestPoolExists,
		Update: resourceGuestPoolUpdate,
		Delete: resourceGuestPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"density": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 2,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"gpu": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"persistent": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
				ForceNew: true,
			},
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"profile": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"seed": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_type": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "disk",
				Optional: true,
				ForceNew: true,
			},
			"storage_id": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "disk",
				Optional: true,
				ForceNew: true,
			},
			"cloudinit_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"cloudinit_userdata": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "",
				Optional: true,
			},
		},
	}
}

func poolFromResource(d *schema.ResourceData) *rest.Pool {
	pool := rest.Pool{
		Name:        d.Get("name").(string),
		ProfileID:   d.Get("profile").(string),
		Seed:        d.Get("seed").(string),
		InjectAgent: true,
		StorageID:   d.Get("storage_id").(string),
		StorageType: d.Get("storage_type").(string),
		Type:        "vdi",
		Density:     []int{d.Get("density.0").(int), d.Get("density.1").(int)},
	}

	guestProfile := rest.PoolGuestProfile{
		Persistent:   d.Get("persistent").(bool),
		TemplateName: d.Get("template").(string),
		Gpu:          d.Get("gpu").(bool),
	}

	if cpu, ok := d.GetOk("cpu"); ok {
		guestProfile.CPU = []int{cpu.(int), cpu.(int)}
	}
	if mem, ok := d.GetOk("memory"); ok {
		guestProfile.Mem = []int{mem.(int), mem.(int)}
	}
	if cloudInitEnabled := d.Get("cloudinit_enabled").(bool); cloudInitEnabled {
		cloudInit := rest.PoolCloudInit{
			Enabled:  cloudInitEnabled,
			UserData: d.Get("cloudinit_userdata").(string),
		}
		guestProfile.CloudInit = &cloudInit
	}

	pool.GuestProfile = &guestProfile

	if d.Id() != "" {
		pool.ID = d.Id()
	}

	return &pool
}

func resourceGuestPoolCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool := poolFromResource(d)

	template, err := client.GetTemplate(pool.GuestProfile.TemplateName)
	if err != nil {
		return err
	}
	pool.GuestProfile.OS = template.OS
	pool.GuestProfile.Vga = template.DisplayDriver
	if len(pool.GuestProfile.CPU) != 2 {
		pool.GuestProfile.CPU = []int{template.Vcpu, template.Vcpu}
	}
	if len(pool.GuestProfile.Mem) != 2 {
		pool.GuestProfile.Mem = []int{template.Mem, template.Mem}
	}

	_, err = pool.Create(client)
	if err != nil {
		return err
	}
	pool, err = client.GetPoolByName(pool.Name)
	if err != nil {
		return err
	}
	d.SetId(pool.ID)
	return resourceGuestPoolRead(d, m)
}

func resourceGuestPoolRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool, err := client.GetPool(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", pool.Name)
	d.Set("cpu", pool.GuestProfile.CPU[0])
	d.Set("memory", pool.GuestProfile.Mem[0])
	d.Set("gpu", pool.GuestProfile.Gpu)
	d.Set("persistent", pool.GuestProfile.Persistent)
	d.Set("inject_agent", pool.InjectAgent)
	d.Set("template", pool.GuestProfile.TemplateName)
	d.Set("profile", pool.ProfileID)
	d.Set("seed", pool.Seed)
	d.Set("state", pool.State)
	d.Set("storage_type", pool.StorageType)
	d.Set("storage_id", pool.StorageID)
	d.Set("density.0", pool.Density[0])
	d.Set("density.1", pool.Density[1])
	if pool.GuestProfile.CloudInit != nil {
		d.Set("cloudinit_enabled", pool.GuestProfile.CloudInit.Enabled)
		d.Set("cloudinit_userdata", pool.GuestProfile.CloudInit.UserData)
	}
	return nil
}

func resourceGuestPoolExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var err error
	id := d.Id()
	_, err = client.GetPool(id)
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	}
	return true, err
}

func resourceGuestPoolUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	pool := poolFromResource(d)
	_, err := pool.Update(client)
	if err != nil {
		return err
	}
	return resourceGuestPoolRead(d, m)
}

func resourceGuestPoolDelete(d *schema.ResourceData, m interface{}) error {
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
			return resource.NonRetryableError(nil)
		}
		return resource.NonRetryableError(err)
	})
}
