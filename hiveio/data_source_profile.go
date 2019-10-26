package hiveio

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func dataSourceProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceProfileRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"timezone": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ad_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ou": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_volumes": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_schedule": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"repository": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"target": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceProfileRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var profile *rest.Profile
	var err error

	id, idOk := d.GetOk("id")
	name, nameOk := d.GetOk("name")
	if idOk {
		profile, err = client.GetProfile(id.(string))
	} else if nameOk {
		profile, err = client.GetProfileByName(name.(string))
	} else {
		return fmt.Errorf("id or name must be provided")
	}

	if err != nil {
		return err
	}
	d.SetId(profile.ID)
	d.Set("name", profile.Name)
	d.Set("timezone", profile.Timezone)

	if _, ok := d.GetOk("ad_config"); ok {
		var adConfig rest.ProfileADConfig
		adConfig.Domain = d.Get("ad_config.0.domain").(string)
		adConfig.Username = d.Get("ad_config.0.username").(string)
		adConfig.Password = d.Get("ad_config.0.password").(string)
		adConfig.UserGroup = d.Get("ad_config.0.user_group").(string)
		if ou, ok := d.GetOk("ad_config.0.ou"); ok {
			adConfig.Ou = ou.(string)
		}
		profile.AdConfig = &adConfig
	}

	if profile.AdConfig != nil {
		d.Set("ad_config.0.domain", profile.AdConfig.Domain)
		d.Set("ad_config.0.username", profile.AdConfig.Domain)
		d.Set("ad_config.0.user_group", profile.AdConfig.UserGroup)
		d.Set("ad_config.0.ou", profile.AdConfig.Ou)
	}

	if profile.UserVolumes != nil {
		d.Set("user_volumes.0.repository", profile.UserVolumes.Repository)
		d.Set("user_volumes.0.size", profile.UserVolumes.Size)
		d.Set("user_volumes.0.backup_schedule", profile.UserVolumes.BackupSchedule)
		d.Set("user_volumes.0.target", profile.UserVolumes.Target)
	}
	return nil
}
