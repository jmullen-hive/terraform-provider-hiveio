package hiveio

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceProfileCreate,
		Read:   resourceProfileRead,
		Exists: resourceProfileExists,
		Update: resourceProfileUpdate,
		Delete: resourceProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"timezone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ad_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"user_group": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ou": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"user_volumes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_schedule": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"repository": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:      schema.TypeInt,
							Required:  true,
							Sensitive: true,
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"backup": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"frequency": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func profileFromResource(d *schema.ResourceData) *rest.Profile {
	var profile *rest.Profile
	profile = &rest.Profile{
		Name:     d.Get("name").(string),
		Timezone: d.Get("timezone").(string),
	}

	if d.Id() != "" {
		profile.ID = d.Id()
	}
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

	if _, ok := d.GetOk("user_volumes"); ok {
		var uv rest.ProfileUserVolumes
		uv.Repository = d.Get("user_volumes.0.repository").(string)
		uv.Size = d.Get("user_volumes.0.size").(int)
		if backup, ok := d.GetOk("user_volumes.0.backup_schedule"); ok {
			uv.BackupSchedule = backup.(int)
		}
		if target, ok := d.GetOk("user_volumes.0.target"); ok {
			uv.Target = target.(string)
		}
		profile.UserVolumes = &uv
	}

	if _, ok := d.GetOk("backup"); ok {
		var backup rest.ProfileBackup
		backup.Enabled = d.Get("backup.0.enabled").(bool)
		backup.Frequency = d.Get("backup.0.frequency").(string)
		backup.TargetStorageID = d.Get("backup.0.target").(string)
		profile.Backup = &backup
	}
	return profile
}

func resourceProfileCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	profile := profileFromResource(d)
	_, err := profile.Create(client)
	if err != nil {
		return err
	}
	profile, err = client.GetProfileByName(profile.Name)
	if err != nil {
		return err
	}
	d.SetId(profile.ID)
	return resourceProfileRead(d, m)
}

func resourceProfileRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var profile *rest.Profile
	var err error
	profile, err = client.GetProfile(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", profile.Name)
	d.Set("timezone", profile.Timezone)

	if profile.AdConfig != nil {
		d.Set("ad_config.0.domain", profile.AdConfig.Domain)
		d.Set("ad_config.0.username", profile.AdConfig.Domain)
		//d.Set("ad_config.0.password", profile.AdConfig.Password)
		d.Set("ad_config.0.user_group", profile.AdConfig.UserGroup)
		d.Set("ad_config.0.ou", profile.AdConfig.Ou)
	}

	if profile.UserVolumes != nil {
		d.Set("user_volumes.0.repository", profile.UserVolumes.Repository)
		d.Set("user_volumes.0.size", profile.UserVolumes.Size)
		d.Set("user_volumes.0.backup_schedule", profile.UserVolumes.BackupSchedule)
		d.Set("user_volumes.0.target", profile.UserVolumes.Target)
	}

	if profile.Backup != nil {
		d.Set("backup.0.enabled", profile.Backup.Enabled)
		d.Set("backup.0.frequency", profile.Backup.Frequency)
		d.Set("backup.0.target", profile.Backup.TargetStorageID)
	}
	return nil
}

func resourceProfileExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var err error
	id := d.Id()
	if id != "" {
		_, err = client.GetProfile(id)
	} else {
		_, err = client.GetProfileByName(d.Get("name").(string))
	}
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func resourceProfileUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	profile := profileFromResource(d)
	_, err := profile.Update(client)
	if err != nil {
		return err
	}
	return resourceProfileRead(d, m)
}

func resourceProfileDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	profile, err := client.GetProfile(d.Id())
	if err != nil {
		return err
	}
	return profile.Delete(client)
}
