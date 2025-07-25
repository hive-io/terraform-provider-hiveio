package hiveio

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceHostIscsi() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostIscsiCreate,
		ReadContext:   resourceHostIscsiRead,
		DeleteContext: resourceHostIscsiDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Adds an iscsi disk to a host in the Hive cluster.",
		Schema: map[string]*schema.Schema{
			"hostid": {
				Type:        schema.TypeString,
				Description: "id of the host for the iscsi connection",
				ForceNew:    true,
				Required:    true,
			},
			"portal": {
				Description: "the iscsi portal address",
				ForceNew:    true,
				Type:        schema.TypeString,
				Required:    true,
			},
			"target": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Description: "the iscsi target to attach",
				Required:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "username to use for authentication",
				ForceNew:    true,
				Optional:    true,
				Default:     "",
			},
			"password": {
				Type:        schema.TypeString,
				Description: "password to use for authentication",
				ForceNew:    true,
				Optional:    true,
				Sensitive:   true,
				Default:     "",
			},
			"block_devices": {
				Type:        schema.TypeList,
				Description: "list of block devices",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "name of the block device",
							Computed:    true,
						},
						"path": {
							Type:        schema.TypeString,
							Description: "path of the block device",
							Computed:    true,
						},
						"fstype": {
							Type:        schema.TypeString,
							Description: "filesystem type of the block device",
							Computed:    true,
						},
						"model": {
							Type:        schema.TypeString,
							Description: "model of the block device",
							Computed:    true,
						},
						"vendor": {
							Type:        schema.TypeString,
							Description: "vendor of the block device",
							Computed:    true,
						},
						"serial": {
							Type:        schema.TypeString,
							Description: "serial number of the block device",
							Computed:    true,
						},
						"size": {
							Type:        schema.TypeString,
							Description: "size of the block device in bytes",
							Computed:    true,
						},
						"label": {
							Type:        schema.TypeString,
							Description: "label of the block device",
							Computed:    true,
						},
					},
				},
			},
			"discovered_portal": {
				Type:        schema.TypeString,
				Description: "the discovered portal address",
				Computed:    true,
			},
			"provider_override": &providerOverride,
		},
	}
}

func resourceHostIscsiCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	hostid := d.Get("hostid").(string)
	host, err := client.GetHost(hostid)
	if err != nil {
		return diag.FromErr(err)
	}
	portal := d.Get("portal").(string)
	target := d.Get("target").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	entries, err := host.IscsiDiscover(client, portal)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(entries) == 0 {
		return diag.FromErr(fmt.Errorf("no iscsi targets found"))
	}

	for _, entry := range entries {
		if entry.Target != target {
			continue
		}
		portal = entry.Portal
		d.Set("discovered_portal", portal)
	}

	// Check if the session already exists
	sessions, err := host.IscsiSessions(client, portal, target)
	if err == nil && len(sessions) > 0 {
		return resourceHostIscsiRead(ctx, d, m)
	}

	authMethod := "None"
	if username != "" && password != "" {
		authMethod = "CHAP"
	}
	sessions, err = host.IscsiLogin(client, portal, target, authMethod, username, password)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(sessions) == 0 {
		return diag.FromErr(fmt.Errorf("no iscsi sessions found"))
	}

	return resourceHostIscsiRead(ctx, d, m)
}

func resourceHostIscsiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	host, err := client.GetHost(d.Get("hostid").(string))
	if err != nil && strings.Contains(err.Error(), "\"error\": 404") {
		d.SetId("")
		return diag.Diagnostics{}
	} else if err != nil {
		return diag.FromErr(err)
	}

	sessions, err := host.IscsiSessions(client, d.Get("portal").(string), d.Get("target").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(sessions) == 0 {
		d.SetId("")
		return diag.Diagnostics{}
	}
	portal := d.Get("portal").(string)
	target := d.Get("target").(string)
	if discovered_portal, ok := d.Get("discovered_portal").(string); ok {
		portal = discovered_portal
	}

	for _, session := range sessions {
		if session.Portal != portal {
			continue
		}
		if session.Target != target {
			continue
		}

		d.SetId(fmt.Sprintf("%s/%s", session.Portal, session.Target))
		if err := d.Set("discovered_portal", session.Portal); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("target", session.Target); err != nil {
			return diag.FromErr(err)
		}
		blockDevices := make([]interface{}, len(session.BlockDevices))
		for i, device := range session.BlockDevices {
			blockDevices[i] = map[string]interface{}{
				"name":   device.Name,
				"path":   device.Path,
				"fstype": device.Fstype,
				"model":  device.Model,
				"vendor": device.Vendor,
				"serial": device.Serial,
				"size":   device.Size,
				"label":  device.Label,
			}
		}
		err = d.Set("block_devices", blockDevices)
		if err != nil {
			return diag.FromErr(err)
		}

		return diag.Diagnostics{}
	}

	d.SetId("")
	return diag.Diagnostics{}
}

func resourceHostIscsiDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := getClient(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	Host, err := client.GetHost(d.Get("hostid").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = Host.IscsiLogout(client, d.Get("portal").(string), d.Get("target").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
