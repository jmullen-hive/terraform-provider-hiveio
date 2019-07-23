package hiveio

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

// type Realm struct {
// 	Enabled  bool     `json:"enabled"`
// 	FQDN     string   `json:"fqdn"`
// 	Name     string   `json:"name"`
// 	Tags     []string `json:"tags,omitempty"`
// 	Verified bool     `json:"verified"`
// }

func resourceRealm() *schema.Resource {
	return &schema.Resource{
		Create: resourceRealmCreate,
		Read:   resourceRealmRead,
		Exists: resourceRealmExists,
		Update: resourceRealmUpdate,
		Delete: resourceRealmDelete,

		Schema: map[string]*schema.Schema{
			"fqdn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"verified": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRealmCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var realm *rest.Realm
	realm = &rest.Realm{
		Name: d.Get("name").(string),
		FQDN: d.Get("fqdn").(string),
	}

	_, err := realm.Create(client)
	if err != nil {
		return err
	}
	d.SetId(realm.Name)
	return resourceRealmRead(d, m)
}

func resourceRealmRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var realm rest.Realm
	var err error
	realm, err = client.GetRealm(d.Id())
	if err != nil {
		return err
	}
	d.SetId(realm.Name)
	d.Set("name", realm.Name)
	d.Set("fqdn", realm.FQDN)
	return nil
}

func resourceRealmExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	id := d.Id()
	realm, err := client.GetRealm(id)
	if err != nil || realm.Name != d.Get("name").(string) {
		return false, err
	}
	return true, nil
}

func resourceRealmUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var realm rest.Realm
	realm.Name = d.Get("name").(string)
	realm.FQDN = d.Get("fqdn").(string)
	_, err := realm.Update(client)
	if err != nil {
		return err
	}
	return resourceRealmRead(d, m)
}

func resourceRealmDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	realm, err := client.GetRealm(d.Id())
	if err != nil {
		return err
	}
	return realm.Delete(client)
}
