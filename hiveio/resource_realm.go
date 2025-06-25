package hiveio

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceRealm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRealmCreate,
		ReadContext:   resourceRealmRead,
		UpdateContext: resourceRealmUpdate,
		DeleteContext: resourceRealmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"fqdn": {
				Description: "fully qualified domain nam",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "netbios name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"site": {
				Type:        schema.TypeString,
				Description: "Active directory site to use instead of Default-First-Site-Name",
				Optional:    true,
			},
			"alias": {
				Type:        schema.TypeString,
				Description: "Alias for the fqdn for broker login",
				Optional:    true,
			},
			"verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tls": {
				Type:        schema.TypeBool,
				Description: "Require tls for the ldap connection",
				Optional:    true,
				Default:     false,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "Service Account username",
				Optional:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "Service Account password",
				Optional:    true,
				Sensitive:   true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func realmFromResource(d *schema.ResourceData) rest.Realm {
	realm := rest.Realm{
		Name:     d.Get("name").(string),
		FQDN:     d.Get("fqdn").(string),
		ForceTLS: d.Get("tls").(bool),
		ServiceAccount: &rest.RealmServiceAccount{
			Username: d.Get("username").(string),
			Password: d.Get("password").(string),
		},
	}
	if site, ok := d.GetOk("site"); ok {
		realm.Site = site.(string)
	}

	if alias, ok := d.GetOk("alias"); ok {
		realm.Alias = alias.(string)
	}

	if verified, ok := d.GetOk("verified"); ok {
		realm.Verified = verified.(bool)
	}
	return realm
}

func resourceRealmCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realm := realmFromResource(d)
	_, err = realm.Create(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(realm.Name)
	return resourceRealmRead(ctx, d, m)
}

func resourceRealmRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	var realm rest.Realm
	realm, err = client.GetRealm(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(realm.Name)
	d.Set("name", realm.Name)
	d.Set("fqdn", realm.FQDN)
	d.Set("username", realm.ServiceAccount.Username)
	d.Set("site", realm.Site)
	d.Set("alias", realm.Alias)
	d.Set("tls", realm.ForceTLS)
	return diag.Diagnostics{}
}

func resourceRealmUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realm := realmFromResource(d)
	_, err = realm.Update(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceRealmRead(ctx, d, m)
}

func resourceRealmDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	realm, err := client.GetRealm(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = realm.Delete(client)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
