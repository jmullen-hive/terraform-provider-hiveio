package hiveio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceHostIscsi() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostIscsiCreate,
		ReadContext:   resourceHostIscsiRead,
		DeleteContext: resourceHostIscsiDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"hostid": {
				Type:        schema.TypeString,
				Description: "id of the host for the iscsi connection",
				ForceNew:    true,
				Required:    true,
			},
			"portal": {
				Description: "the iscsi portal address",
				ForceNew:    true,
				Type:        schema.TypeString,
				Required:    true,
			},
			"target": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Description: "the iscsi target to attach",
				Required:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "username to use for authentication",
				ForceNew:    true,
				Optional:    true,
				Default:     "",
			},
			"password": {
				Type:        schema.TypeString,
				Description: "password to use for authentication",
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Default:     "",
			},
			"device_name": {
				Type:        schema.TypeString,
				Description: "device name",
				Computed:    true,
			},
			"device_path": {
				Type:        schema.TypeString,
				Description: "name of the link in /dev/disk/by-path",
				Computed:    true,
			},
			"discovered_portal": {
				Type:        schema.TypeString,
				Description: "the discovered portal address",
				Computed:    true,
			},
		},
	}
}

func resourceHostIscsiCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	hostid := d.Get("hostid").(string)
	host, err := client.GetHost(hostid)
	if err != nil {
		return diag.FromErr(err)
	}
	portal := d.Get("portal").(string)
	target := d.Get("target").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	entries, err := host.IscsiDiscover(client, portal)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(entries) == 0 {
		return diag.FromErr(fmt.Errorf("no iscsi targets found"))
	}

	for _, entry := range entries {
		if entry.Target != target {
			continue
		}
		portal = entry.Portal
		d.Set("discovered_portal", portal)
	}

	// Check if the session already exists
	sessions, err := host.IscsiSessions(client, portal, target)
	if err == nil && len(sessions) > 0 {
		return resourceHostIscsiRead(ctx, d, m)
	}

	authMethod := "None"
	if username != "" && password != "" {
		authMethod = "CHAP"
	}
	sessions, err = host.IscsiLogin(client, portal, target, authMethod, username, password)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(sessions) == 0 {
		return diag.FromErr(fmt.Errorf("no iscsi sessions found"))
	}

	return resourceHostIscsiRead(ctx, d, m)
}

func resourceHostIscsiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	var err error
	host, err := client.GetHost(d.Get("hostid").(string))
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}

	sessions, err := host.IscsiSessions(client, d.Get("portal").(string), d.Get("target").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(sessions) == 0 {
		d.SetId("")
		return diag.Diagnostics{}
	}
	portal := d.Get("portal").(string)
	target := d.Get("target").(string)
	if discovered_portal, ok := d.Get("discovered_portal").(string); ok {
		portal = discovered_portal
	}

	for _, session := range sessions {
		if session.Portal != portal {
			continue
		}
		if session.Target != target {
			continue
		}

		d.SetId(fmt.Sprintf("%s/%s", session.Portal, session.Target))
		d.Set("discovered_portal", session.Portal)
		d.Set("target", session.Target)
		d.Set("device_name", session.BlockDevice.Name)
		d.Set("device_path", session.BlockDevice.Path)

		return diag.Diagnostics{}
	}

	d.SetId("")
	return diag.Diagnostics{}
}

func resourceHostIscsiDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	Host, err := client.GetHost(d.Get("hostid").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = Host.IscsiLogout(client, d.Get("portal").(string), d.Get("target").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
