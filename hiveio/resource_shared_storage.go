package hiveio

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceSharedStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceSharedStorageCreate,
		Read:   resourceSharedStorageRead,
		Delete: resourceSharedStorageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"minimum_set_size": {
				Type:        schema.TypeInt,
				Description: "minimum number of hosts required to increase shared storage",
				Default:     3,
				Optional:    true,
				ForceNew:    true,
			},
			"utilization": {
				Type:     schema.TypeInt,
				Default:  75,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "name",
				Computed:    true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
	}
}

func resourceSharedStorageCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)

	// roles := []string{}
	// for _, value := range d.Get("roles").([]interface{}) {
	// 	roles = append(roles, value.(string))
	// }
	// storage.Roles = roles
	setSize := d.Get("minimum_set_size").(int)
	utilization := d.Get("utilization").(int)
	clusterID, err := client.ClusterID()
	if err != nil {
		return err
	}
	cluster, err := client.GetCluster(clusterID)
	if err != nil {
		return err
	}
	task, err := cluster.EnableSharedStorage(client, utilization, setSize)
	if err != nil {
		return err
	}
	task, err = task.WaitForTask(client, false)
	if err != nil {
		return err
	}
	if task.State == "failed" {
		return fmt.Errorf("failed to Enable Shared storage: %s", task.Message)
	}
	storage, err := client.GetStoragePoolByName("HF_Shared")
	if err != nil {
		return err
	}
	d.SetId(storage.ID)
	d.Set("name", "HF_Shared")
	return resourceSharedStorageRead(d, m)
}

func resourceSharedStorageRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var storage *rest.StoragePool
	var err error
	storage, err = client.GetStoragePool(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}
	d.SetId(storage.ID)
	d.Set("name", storage.Name)
	return nil
}

func resourceSharedStorageDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		clusterID, err := client.ClusterID()
		if err != nil {
			return resource.NonRetryableError(err)
		}
		cluster, err := client.GetCluster(clusterID)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		task, err := cluster.DisableSharedStorage(client)
		if err != nil {
			return resource.RetryableError(err)
		}
		task, err = task.WaitForTask(client, false)
		if err != nil {
			return resource.RetryableError(err)
		}
		if task.State == "failed" {
			return resource.NonRetryableError(fmt.Errorf("failed to Disable Shared storage: %s", task.Message))
		}
		return nil
	})
}
