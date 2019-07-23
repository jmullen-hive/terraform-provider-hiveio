package hiveio

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

/*type StoragePool struct {
	ID           string   `json:"id,omitempty"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Server       string   `json:"server"`
	Path         string   `json:"path"`
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	Key          string   `json:"key,omitempty"`
	MountOptions []string `json:"mountOptions,omitempty"`
	Roles        []string `json:"roles,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}*/

func resourceStoragePool() *schema.Resource {
	return &schema.Resource{
		Create: resourceStoragePoolCreate,
		Read:   resourceStoragePoolRead,
		Exists: resourceStoragePoolExists,
		Delete: resourceStoragePoolDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"key": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"mount_options": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true, //TODO: add update that only changes this field
			},
		},
	}
}

func resourceStoragePoolCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	storage = &rest.StoragePool{
		Name:   d.Get("name").(string),
		Server: d.Get("server").(string),
		Path:   d.Get("path").(string),
		Type:   d.Get("type").(string),
	}

	roles := []string{}
	for _, value := range d.Get("roles").([]interface{}) {
		roles = append(roles, value.(string))
	}
	storage.Roles = roles

	if username, ok := d.GetOk("username"); ok {
		storage.Username = username.(string)
	}
	if password, ok := d.GetOk("password"); ok {
		storage.Password = password.(string)
	}
	if key, ok := d.GetOk("key"); ok {
		storage.Key = key.(string)
	}

	_, err := storage.Create(client)
	if err != nil {
		return err
	}
	storage, err = client.GetStoragePoolByName(storage.Name)
	if err != nil {
		return err
	}
	d.SetId(storage.ID)
	return resourceStoragePoolRead(d, m)
}

func resourceStoragePoolRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	var err error
	storage, err = client.GetStoragePool(d.Id())
	if err != nil {
		return err
	}
	d.SetId(storage.ID)
	d.Set("name", storage.Name)
	d.Set("server", storage.Server)
	d.Set("path", storage.Path)
	d.Set("type", storage.Type)
	d.Set("username", storage.Username)
	d.Set("password", storage.Password)
	d.Set("key", storage.Key)
	d.Set("roles", storage.Roles)
	return nil
}

func resourceStoragePoolExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	var err error
	if d.Id() != "" {
		storage, err = client.GetStoragePool(d.Id())
	} else {
		storage, err = client.GetStoragePoolByName(d.Get("name").(string))
	}
	if err != nil || storage.Name != d.Get("name").(string) {
		return false, err
	}
	return true, nil
}

func resourceStoragePoolDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	storage, err := client.GetStoragePool(d.Id())
	if err != nil {
		return err
	}
	return storage.Delete(client)
}
