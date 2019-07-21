package hiveio

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

// type Profile struct {
// 	AdConfig *struct {
// 		Domain    string      `json:"domain,omitempty"`
// 		Ou        interface{} `json:"ou,omitempty"`
// 		Password  string      `json:"password,omitempty"`
// 		UserGroup string      `json:"userGroup,omitempty"`
// 		Username  string      `json:"username,omitempty"`
// 	} `json:"adConfig,omitempty"`
// 	BrokerOptions *struct {
// 		AllowDesktopComposition bool `json:"allowDesktopComposition,omitempty"`
// 		AudioCapture            bool `json:"audioCapture,omitempty"`
// 		RedirectCSSP            bool `json:"redirectCSSP,omitempty"`
// 		RedirectClipboard       bool `json:"redirectClipboard,omitempty"`
// 		RedirectDisk            bool `json:"redirectDisk,omitempty"`
// 		RedirectPNP             bool `json:"redirectPNP,omitempty"`
// 		RedirectPrinter         bool `json:"redirectPrinter,omitempty"`
// 		RedirectUSB             bool `json:"redirectUSB,omitempty"`
// 		SmartResize             bool `json:"smartResize,omitempty"`
// 	} `json:"brokerOptions,omitempty"`
// 	BypassBroker bool     `json:"bypassBroker,omitempty"`
// 	ID           string   `json:"id,omitempty"`
// 	Name         string   `json:"name"`
// 	Tags         []string `json:"tags,omitempty"`
// 	Timezone     string   `json:"timezone,omitempty"`
// 	UserVolumes  *struct {
// 		BackupSchedule int    `json:"backupSchedule,omitempty"`
// 		Repository     string `json:"repository,omitempty"`
// 		Size           int    `json:"size,omitempty"`
// 		Target         string `json:"target,omitempty"`
// 	} `json:"userVolumes,omitempty"`
// 	Vlan int `json:"vlan,omitempty"`
// }

func resourceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceProfileCreate,
		Read:   resourceProfileRead,
		Exists: resourceProfileExists,
		Update: resourceProfileUpdate,
		Delete: resourceProfileDelete,

		Schema: map[string]*schema.Schema{
			/*"profile_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},*/
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"timezone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ad_domain": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ad_ou": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_group": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ad_username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ad_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceProfileCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var profile *rest.Profile
	profile = &rest.Profile{
		Name:     d.Get("name").(string),
		Timezone: d.Get("timezone").(string),
	}
	// var adConfig *struct{
	// 	Username:  d.Get("ad_username").(string),
	// 	Password:  d.Get("ad_password").(string),
	// 	UserGroup: d.Get("ad_user_group").(string),
	// 	Domain:    d.Get("ad_domain").(string),
	// 	OU:        d.Get("ad_ou").(string),
	// }
	_, err := profile.Create(client)
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
	d.SetId(profile.ID)
	d.Set("name", profile.Name)
	d.Set("profile_id", profile.ID)
	d.Set("timezone", profile.Timezone)
	return nil
}

func resourceProfileExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*rest.Client)
	var profile *rest.Profile
	var err error
	id := d.Id()
	if id != "" {
		profile, err = client.GetProfile(id)
	} else {
		profile, err = client.GetProfileByName(d.Get("name").(string))
	}
	if err != nil || profile.Name != d.Get("name").(string) {
		return false, err
	}
	return true, nil
}

func resourceProfileUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*rest.Client)
	var profile rest.Profile
	profile.Name = d.Get("name").(string)
	profile.Timezone = d.Get("timezone").(string)
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
