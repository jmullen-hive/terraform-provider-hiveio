package hiveio

import (
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceExternalGuest() *schema.Resource {
	return &schema.Resource{
		Create: resourceExternalGuestCreate,
		Read:   resourceExternalGuestRead,
		Exists: resourceExternalGuestExists,
		Delete: resourceExternalGuestDelete,
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
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"realm": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"os": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func guestFromResource(d *schema.ResourceData) rest.ExternalGuest {
	guest := rest.ExternalGuest{
		GuestName: d.Get("name").(string),
		Address:   d.Get("address").(string),
		Username:  d.Get("username").(string),
		Realm:     d.Get("realm").(string),
	}

	if os, ok := d.GetOk("os"); ok {
		guest.OS = os.(string)
	}

	return guest
}

func resourceExternalGuestCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	guest := guestFromResource(d)

	_, err := guest.Create(client)
	if err != nil {
		return err
	}
	d.SetId(guest.GuestName)
	return resourceExternalGuestRead(d, m)
}

func resourceExternalGuestRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	guest, err := client.GetGuest(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", guest.Name)
	d.Set("address", guest.Address)
	d.Set("username", guest.Username)
	d.Set("realm", guest.Realm)
	d.Set("os", guest.Os)

	return nil
}

func resourceExternalGuestExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var err error
	id := d.Id()
	_, err = client.GetGuest(id)
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	}
	return true, err
}

func resourceExternalGuestDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	guest, err := client.GetGuest(d.Id())
	if err != nil {
		return err
	}
	err = guest.Delete(client)
	return err
}
