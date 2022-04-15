package hiveio

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Add a ldap user or group with admin or readonly access.",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"groupname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"realm": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Description: "readonly or admin",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func userFromResource(d *schema.ResourceData) (*rest.User, error) {
	user := rest.User{
		ID:    uuid.New().String(),
		Realm: d.Get("realm").(string),
		Role:  d.Get("role").(string),
	}

	if username, ok := d.GetOk("username"); ok {
		user.Username = username.(string)
	} else if groupname, ok := d.GetOk("groupname"); ok {
		user.GroupName = groupname.(string)
	} else {
		return nil, fmt.Errorf("username or groupname must be provided")
	}
	return &user, nil
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	user, err := userFromResource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = user.Create(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.ID)
	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	var user *rest.User
	var err error
	user, err = client.GetUser(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	if user.Username != "" {
		d.Set("username", user.Username)
	} else if user.GroupName != "" {
		d.Set("groupname", user.GroupName)
	}

	d.Set("realm", user.Realm)
	d.Set("role", user.Role)
	return diag.Diagnostics{}
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	user, err := userFromResource(d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = user.Update(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	user, err := client.GetUser(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = user.Delete(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
