package hiveio

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceStoragePool() *schema.Resource {
	return &schema.Resource{
		Create: resourceStoragePoolCreate,
		Read:   resourceStoragePoolRead,
		Exists: resourceStoragePoolExists,
		Delete: resourceStoragePoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"server": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
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
			"s3_access_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"s3_secret_access_key": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"s3_region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
	}
}

func resourceStoragePoolCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	storage = &rest.StoragePool{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
	}

	roles := []string{}
	for _, value := range d.Get("roles").([]interface{}) {
		roles = append(roles, value.(string))
	}
	storage.Roles = roles

	if server, ok := d.GetOk("server"); ok {
		storage.Server = server.(string)
	}
	if path, ok := d.GetOk("path"); ok {
		storage.Path = path.(string)
	}
	if url, ok := d.GetOk("url"); ok {
		storage.URL = url.(string)
	}

	if username, ok := d.GetOk("username"); ok {
		storage.Username = username.(string)
	}
	if password, ok := d.GetOk("password"); ok {
		storage.Password = password.(string)
	}
	if key, ok := d.GetOk("key"); ok {
		storage.Key = key.(string)
	}

	if S3AccessKeyID, ok := d.GetOk("s3_access_key_id"); ok {
		storage.S3AccessKeyID = S3AccessKeyID.(string)
	}
	if s3SecretAccessKey, ok := d.GetOk("s3_secret_access_key"); ok {
		storage.S3SecretAccessKey = s3SecretAccessKey.(string)
	}
	if s3Region, ok := d.GetOk("s3_region"); ok {
		storage.S3Region = s3Region.(string)
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
	d.Set("url", storage.URL)
	d.Set("type", storage.Type)
	d.Set("username", storage.Username)
	//d.Set("password", storage.Password)
	//d.Set("key", storage.Key)
	d.Set("roles", storage.Roles)
	d.Set("s3_acess_key_id", storage.S3AccessKeyID)
	d.Set("s3_region", storage.S3Region)
	return nil
}

func resourceStoragePoolExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var err error
	if d.Id() != "" {
		_, err = client.GetStoragePool(d.Id())
	} else {
		_, err = client.GetStoragePoolByName(d.Get("name").(string))
	}
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	} else if err != nil {
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
	//{"error": 423, "message": {"code":"LockedError","message":"Storage pool vms is in use and can not be deleted"}}
	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err = storage.Delete(client)
		if err != nil && strings.Contains(err.Error(), "\"error\": 423") {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(fmt.Errorf("Storage Pool %s is in use", d.Id()))
		}
		return resource.NonRetryableError(err)
	})
}
