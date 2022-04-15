package hiveio

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceExternalGuest() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource can be used to add an external guest for access through the broker.",
		CreateContext: resourceExternalGuestCreate,
		ReadContext:   resourceExternalGuestRead,
		DeleteContext: resourceExternalGuestDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"address": {
				Description: "Hostname or ip address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"username": {
				Description: "The user the guest will be assigned to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"realm": {
				Description: "The realm of the user",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"os": {
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

func resourceExternalGuestCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	guest := guestFromResource(d)

	_, err := guest.Create(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(guest.GuestName)
	return resourceExternalGuestRead(ctx, d, m)
}

func resourceExternalGuestRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	guest, err := client.GetGuest(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", guest.Name)
	d.Set("address", guest.Address)
	d.Set("username", guest.Username)
	d.Set("realm", guest.Realm)
	d.Set("os", guest.Os)

	return diag.Diagnostics{}
}

func resourceExternalGuestDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	guest, err := client.GetGuest(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = guest.Delete(client)
	return diag.FromErr(err)
}
