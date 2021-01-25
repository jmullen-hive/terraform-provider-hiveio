package hiveio

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Add a ldap user or group with admin or readonly access.",
		Create:      resourceUserCreate,
		Read:        resourceUserRead,
		Update:      resourceUserUpdate,
		Delete:      resourceUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	user, err := userFromResource(d)
	if err != nil {
		return err
	}

	_, err = user.Create(client)
	if err != nil {
		return err
	}
	d.SetId(user.ID)
	return resourceUserRead(d, m)
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var user *rest.User
	var err error
	user, err = client.GetUser(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}
	d.SetId(user.ID)
	if user.Username != "" {
		d.Set("username", user.Username)
	} else if user.GroupName != "" {
		d.Set("groupname", user.GroupName)
	}

	d.Set("realm", user.Realm)
	d.Set("role", user.Role)
	return nil
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	user, err := userFromResource(d)
	if err != nil {
		return err
	}
	_, err = user.Update(client)
	if err != nil {
		return err
	}
	return resourceUserRead(d, m)
}

func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	user, err := client.GetUser(d.Id())
	if err != nil {
		return err
	}
	return user.Delete(client)
}
