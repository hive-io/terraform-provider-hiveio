package hiveio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hive-io/hive-go-client/rest"
)

func resourceExternalGuest() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource can be used to add an external guest for access through the broker.",
		CreateContext: resourceExternalGuestCreate,
		ReadContext:   resourceExternalGuestRead,
		DeleteContext: resourceExternalGuestDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"address": {
				Description: "Hostname or ip address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"profile": {
				Description: "The id of a profile to use for the guest in version 8.6.0 and later.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"username": {
				Description: "The user assignment for broker access",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"ad_group": {
				Description: "The active directory group assignment for broker access",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"realm": {
				Description: "The realm of the user or ad_group.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"os": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disable_port_check": {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
			},
			"broker_default_connection": {
				Type:     schema.TypeString,
				Default:  "",
				ForceNew: true,
				Optional: true,
			},
			"broker_connection": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Default:  "",
							Optional: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disable_html5": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"gateway": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disabled": {
										Type:     schema.TypeBool,
										Default:  false,
										Optional: true,
									},
									"persistent": {
										Type:     schema.TypeBool,
										Default:  false,
										Optional: true,
									},
									"protocols": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"provider_override": &providerOverride,
		},
	}
}

func guestFromResource(d *schema.ResourceData) rest.ExternalGuest {
	guest := rest.ExternalGuest{
		GuestName:        d.Get("name").(string),
		Address:          d.Get("address").(string),
		DisablePortCheck: d.Get("disable_port_check").(bool),
	}

	if profile, ok := d.GetOk("profile"); ok {
		guest.ProfileID = profile.(string)
	}

	if username, ok := d.GetOk("username"); ok {
		guest.Username = username.(string)
	}
	if adGroup, ok := d.GetOk("ad_group"); ok {
		guest.ADGroup = adGroup.(string)
	}

	if realm, ok := d.GetOk("realm"); ok {
		guest.Realm = realm.(string)
	}

	if os, ok := d.GetOk("os"); ok {
		guest.OS = os.(string)
	}
	if guest.BrokerOptions == nil {
		guest.BrokerOptions = &rest.GuestBrokerOptions{}
	}
	if defaultConnection, ok := d.GetOk("broker_default_connection"); ok {
		guest.BrokerOptions.DefaultConnection = defaultConnection.(string)
	}
	var connections []rest.GuestBrokerConnection
	for i := 0; i < d.Get("broker_connection.#").(int); i++ {
		prefix := fmt.Sprintf("broker_connection.%d.", i)
		connection := rest.GuestBrokerConnection{
			Name:         d.Get(prefix + "name").(string),
			Description:  d.Get(prefix + "description").(string),
			Port:         uint(d.Get(prefix + "port").(int)),
			Protocol:     d.Get(prefix + "protocol").(string),
			DisableHtml5: d.Get(prefix + "disable_html5").(bool),
		}
		connection.Gateway.Disabled = d.Get(prefix + "gateway.0." + "disabled").(bool)
		connection.Gateway.Persistent = d.Get(prefix + "gateway.0." + "persistent").(bool)
		if protocolsInterface, ok := d.GetOk(prefix + "gateway.0." + "protocols"); ok {
			protocols := make([]string, len(protocolsInterface.([]interface{})))
			for i, protocol := range protocolsInterface.([]interface{}) {
				protocols[i] = protocol.(string)
			}
			connection.Gateway.Protocols = protocols
		}
		connections = append(connections, connection)
	}
	guest.BrokerOptions.Connections = connections

	return guest
}

func resourceExternalGuestCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	guest := guestFromResource(d)

	_, err = guest.Create(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(guest.GuestName)
	return resourceExternalGuestRead(ctx, d, m)
}

func resourceExternalGuestRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	guest, err := client.GetGuest(d.Id())
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", guest.Name)
	d.Set("address", guest.Address)
	d.Set("username", guest.Username)
	d.Set("ad_group", guest.ADGroup)
	d.Set("profile", guest.ProfileID)
	d.Set("realm", guest.Realm)
	d.Set("os", guest.Os)
	d.Set("disable_port_check", guest.DisablePortCheck)

	d.Set("broker_default_connection", guest.BrokerOptions.DefaultConnection)
	connection := make([]interface{}, len(guest.BrokerOptions.Connections))
	for i, conn := range guest.BrokerOptions.Connections {
		connection[i] = map[string]interface{}{
			"name":          conn.Name,
			"description":   conn.Description,
			"port":          conn.Port,
			"protocol":      conn.Protocol,
			"disable_html5": conn.DisableHtml5,
			"gateway": []interface{}{
				map[string]interface{}{
					"disabled":   conn.Gateway.Disabled,
					"persistent": conn.Gateway.Persistent,
					"protocols":  conn.Gateway.Protocols,
				},
			},
		}
	}
	d.Set("broker_connection", connection)

	return diag.Diagnostics{}
}

func resourceExternalGuestDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	guest, err := client.GetGuest(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = guest.Delete(client)
	return diag.FromErr(err)
}
