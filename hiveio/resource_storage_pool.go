package hiveio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceStoragePool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStoragePoolCreate,
		ReadContext:   resourceStoragePoolRead,
		DeleteContext: resourceStoragePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"server": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"key": {
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

func resourceStoragePoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.FromErr(err)
	}
	storage, err = client.GetStoragePoolByName(storage.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(storage.ID)
	return resourceStoragePoolRead(ctx, d, m)
}

func resourceStoragePoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	var err error
	storage, err = client.GetStoragePool(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
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
	return diag.Diagnostics{}
}

func resourceStoragePoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*rest.Client)
	storage, err := client.GetStoragePool(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	//{"error": 423, "message": {"code":"LockedError","message":"Storage pool vms is in use and can not be deleted"}}
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		err = storage.Delete(client)
		if err != nil && strings.Contains(err.Error(), "\"error\": 423") {
			time.Sleep(2 * time.Second)
			return retry.RetryableError(fmt.Errorf("storage Pool %s is in use", d.Id()))
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
