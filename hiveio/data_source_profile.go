package hiveio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func dataSourceProfile() *schema.Resource {
	return &schema.Resource{
		Description: "The profile data source can be used to retrieve settings from an existing profile.",
		ReadContext: dataSourceProfileRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ad_config": {
				Description: "active directory options",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Description: "The realm name used by guests in this profile.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"username": {
							Description: "Username for the active directory service account.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"user_group": {
							Description: "AD group for users who can login through the broker.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"ou": {
							Description: "OU for guests using this profile.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"user_volumes": {
				Type:     schema.TypeList,
				Computed: true,
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
			"broker_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_desktop_composition": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"audio_capture": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"credssp": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"disable_full_window_drag": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disable_menu_anims": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disable_printer": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disable_themes": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disable_wallpaper": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"fail_on_cert_mismatch": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"hide_authentication_failure": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"html5": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"inject_password": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"redirect_clipboard": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"redirect_disk": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"redirect_pnp": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"redirect_printer": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"redirect_smartcard": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"redirect_usb": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"smart_resize": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
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

func dataSourceProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.Errorf("id or name must be provided")
	}

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(profile.ID)
	d.Set("name", profile.Name)
	d.Set("timezone", profile.Timezone)

	if profile.AdConfig != nil {
		adConfig := []map[string]interface{}{
			{
				"domain":     profile.AdConfig.Domain,
				"username":   profile.AdConfig.Username,
				"user_group": profile.AdConfig.UserGroup,
				"ou":         profile.AdConfig.Ou,
			},
		}
		d.Set("ad_config", adConfig)
	}

	if profile.UserVolumes != nil {
		d.Set("user_volumes", []map[string]interface{}{
			{
				"repository":      profile.UserVolumes.Repository,
				"size":            profile.UserVolumes.Size,
				"backup_schedule": profile.UserVolumes.BackupSchedule,
				"target":          profile.UserVolumes.Target,
			},
		})
	}
	if profile.Backup != nil {
		d.Set("backup", []map[string]interface{}{
			{
				"enabled":   profile.Backup.Enabled,
				"frequency": profile.Backup.Frequency,
				"target":    profile.Backup.TargetStorageID,
			},
		})
	}

	if profile.BrokerOptions != nil {
		d.Set("broker_options", []map[string]interface{}{
			{
				"allow_desktop_composition":   profile.BrokerOptions.AllowDesktopComposition,
				"audio_capture":               profile.BrokerOptions.AudioCapture,
				"credssp":                     profile.BrokerOptions.RedirectCSSP,
				"disable_full_window_drag":    profile.BrokerOptions.DisableFullWindowDrag,
				"disable_menu_anims":          profile.BrokerOptions.DisableMenuAnims,
				"disable_printer":             profile.BrokerOptions.DisablePrinter,
				"disable_themes":              profile.BrokerOptions.DisableThemes,
				"disable_wallpaper":           profile.BrokerOptions.DisableWallpaper,
				"fail_on_cert_mismatch":       profile.BrokerOptions.FailOnCertMismatch,
				"hide_authentication_failure": profile.BrokerOptions.HideAuthenticationFailure,
				"html5":                       profile.BrokerOptions.EnableHTML5,
				"inject_password":             profile.BrokerOptions.InjectPassword,
				"redirect_clipboard":          profile.BrokerOptions.RedirectClipboard,
				"redirect_disk":               profile.BrokerOptions.RedirectDisk,
				"redirect_pnp":                profile.BrokerOptions.RedirectPNP,
				"redirect_printer":            profile.BrokerOptions.RedirectPrinter,
				"redirect_smartcard":          profile.BrokerOptions.RedirectSmartCard,
				"redirect_usb":                profile.BrokerOptions.RedirectUSB,
				"smart_resize":                profile.BrokerOptions.SmartResize,
			},
		})
	}
	return diag.Diagnostics{}
}
