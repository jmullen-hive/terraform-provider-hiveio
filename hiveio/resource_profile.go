package hiveio

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProfileCreate,
		ReadContext:   resourceProfileRead,
		UpdateContext: resourceProfileUpdate,
		DeleteContext: resourceProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timezone": {
				Description: "A timezone to inject to guests in the profile.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "disabled",
			},
			"ad_config": {
				Description: "active directory options",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Description: "The realm name to use for guests in this profile.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"username": {
							Description: "Username for a service account to override the one in realm.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"password": {
							Description: "Password for the service account.",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							WriteOnly:   true,
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
				Description: "User Volume options.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
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
			"provider_override": &providerOverride,
		},
	}
}

func profileFromResource(d *schema.ResourceData) *rest.Profile {
	profile := &rest.Profile{
		Name:     d.Get("name").(string),
		Timezone: d.Get("timezone").(string),
	}

	if d.Id() != "" {
		profile.ID = d.Id()
	}
	if _, ok := d.GetOk("ad_config"); ok {
		var adConfig rest.ProfileADConfig
		adConfig.Domain = d.Get("ad_config.0.domain").(string)
		adConfig.UserGroup = d.Get("ad_config.0.user_group").(string)
		if ou, ok := d.GetOk("ad_config.0.ou"); ok {
			adConfig.Ou = ou.(string)
		}
		if username, ok := d.GetOk("ad_config.0.username"); ok {
			adConfig.Username = username.(string)
		}
		if password, ok := d.GetOk("ad_config.0.password"); ok {
			adConfig.Password = password.(string)
		}
		profile.AdConfig = &adConfig
	}
	if _, ok := d.GetOk("broker_options"); ok {
		var config rest.ProfileBrokerOptions
		config.AllowDesktopComposition = d.Get("broker_options.0.allow_desktop_composition").(bool)
		config.AudioCapture = d.Get("broker_options.0.audio_capture").(bool)
		config.RedirectCSSP = d.Get("broker_options.0.credssp").(bool)
		config.FailOnCertMismatch = d.Get("broker_options.0.fail_on_cert_mismatch").(bool)
		config.HideAuthenticationFailure = d.Get("broker_options.0.hide_authentication_failure").(bool)
		config.RedirectClipboard = d.Get("broker_options.0.redirect_clipboard").(bool)
		config.RedirectDisk = d.Get("broker_options.0.redirect_disk").(bool)
		config.RedirectPNP = d.Get("broker_options.0.redirect_pnp").(bool)
		config.RedirectPrinter = d.Get("broker_options.0.redirect_printer").(bool)
		config.RedirectSmartCard = d.Get("broker_options.0.redirect_smartcard").(bool)
		config.RedirectUSB = d.Get("broker_options.0.redirect_usb").(bool)
		config.SmartResize = d.Get("broker_options.0.smart_resize").(bool)
		config.EnableHTML5 = d.Get("broker_options.0.html5").(bool)
		config.DisableFullWindowDrag = d.Get("broker_options.0.disable_full_window_drag").(bool)
		config.DisableMenuAnims = d.Get("broker_options.0.disable_menu_anims").(bool)
		config.DisablePrinter = d.Get("broker_options.0.disable_printer").(bool)
		config.DisableThemes = d.Get("broker_options.0.disable_themes").(bool)
		config.DisableWallpaper = d.Get("broker_options.0.disable_wallpaper").(bool)
		config.InjectPassword = d.Get("broker_options.0.inject_password").(bool)
		profile.BrokerOptions = &config
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

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	profile := profileFromResource(d)
	_, err = profile.Create(client)
	if err != nil {
		return diag.FromErr(err)
	}
	profile, err = client.GetProfileByName(profile.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(profile.ID)
	return resourceProfileRead(ctx, d, m)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	var profile *rest.Profile
	profile, err = client.GetProfile(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", profile.Name)
	d.Set("timezone", profile.Timezone)

	if profile.AdConfig != nil {
		d.Set("ad_config", []map[string]interface{}{
			{
				"domain":     profile.AdConfig.Domain,
				"username":   profile.AdConfig.Username,
				"user_group": profile.AdConfig.UserGroup,
				"ou":         profile.AdConfig.Ou,
			},
		})
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

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	profile := profileFromResource(d)
	_, err = profile.Update(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceProfileRead(ctx, d, m)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	profile, err := client.GetProfile(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = profile.Delete(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
