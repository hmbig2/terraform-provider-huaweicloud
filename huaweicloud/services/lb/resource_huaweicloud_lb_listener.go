package lb

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk/openstack/common/tags"
	"github.com/chnsz/golangsdk/openstack/elb/v2/listeners"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

// @API ELB DELETE /v2/{project_id}/elb/listeners/{id}
// @API ELB GET /v2/{project_id}/elb/listeners/{id}
// @API ELB PUT /v2/{project_id}/elb/listeners/{id}
// @API ELB POST /v2/{project_id}/elb/listeners
// @API ELB POST /v2.0/{project_id}/listeners/{id}/tags/action
// @API ELB GET /v2.0/{project_id}/listeners/{id}/tags
func ResourceListener() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceListenerV2Create,
		ReadContext:   resourceListenerV2Read,
		UpdateContext: resourceListenerV2Update,
		DeleteContext: resourceListenerV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"protocol_port": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"tenant_id": {
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
				ForceNew:   true,
				Deprecated: "tenant_id is deprecated",
			},

			"loadbalancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"default_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"connection_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			"http2_enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"default_tls_container_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"sni_container_refs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"admin_state_up": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"tags": common.TagsSchema(),
		},
	}
}

func resourceListenerV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	lbClient, err := cfg.LoadBalancerClient(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2 Client: %s", err)
	}

	// client for tags
	lbv2Client, err := cfg.ElbV2Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2.0 Client: %s", err)
	}

	lbID := d.Get("loadbalancer_id").(string)
	adminStateUp := d.Get("admin_state_up").(bool)
	http2Enable := d.Get("http2_enable").(bool)
	var sniContainerRefs []string
	if raw, ok := d.GetOk("sni_container_refs"); ok {
		for _, v := range raw.([]interface{}) {
			sniContainerRefs = append(sniContainerRefs, v.(string))
		}
	}
	createOpts := listeners.CreateOpts{
		Protocol:               listeners.Protocol(d.Get("protocol").(string)),
		ProtocolPort:           d.Get("protocol_port").(int),
		TenantID:               d.Get("tenant_id").(string),
		LoadbalancerID:         lbID,
		Name:                   d.Get("name").(string),
		DefaultPoolID:          d.Get("default_pool_id").(string),
		Description:            d.Get("description").(string),
		DefaultTlsContainerRef: d.Get("default_tls_container_ref").(string),
		SniContainerRefs:       sniContainerRefs,
		Http2Enable:            &http2Enable,
		AdminStateUp:           &adminStateUp,
	}

	if v, ok := d.GetOk("connection_limit"); ok {
		connectionLimit := v.(int)
		createOpts.ConnLimit = &connectionLimit
	}

	// Wait for LoadBalancer to become active before continuing
	timeout := d.Timeout(schema.TimeoutCreate)
	err = waitForLBV2LoadBalancer(ctx, lbClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	listener, err := listeners.Create(lbClient, createOpts).Extract()
	if err != nil {
		return diag.Errorf("error creating listener: %s", err)
	}

	// Wait for LoadBalancer to become active again before continuing
	err = waitForLBV2LoadBalancer(ctx, lbClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(listener.ID)

	// set tags
	tagRaw := d.Get("tags").(map[string]interface{})
	if len(tagRaw) > 0 {
		tagList := utils.ExpandResourceTags(tagRaw)
		if tagErr := tags.Create(lbv2Client, "listeners", listener.ID, tagList).ExtractErr(); tagErr != nil {
			return diag.Errorf("error setting tags of ELB listener %s: %s", listener.ID, tagErr)
		}
	}

	return resourceListenerV2Read(ctx, d, meta)
}

func resourceListenerV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	lbClient, err := cfg.LoadBalancerClient(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2 Client: %s", err)
	}

	// client for tags
	lbv2Client, err := cfg.ElbV2Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2.0 Client: %s", err)
	}

	listener, err := listeners.Get(lbClient, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "listener")
	}

	log.Printf("[DEBUG] Retrieved listener %s: %#v", d.Id(), listener)

	mErr := multierror.Append(nil,
		d.Set("region", cfg.GetRegion(d)),
		d.Set("name", listener.Name),
		d.Set("protocol", listener.Protocol),
		d.Set("tenant_id", listener.TenantID),
		d.Set("description", listener.Description),
		d.Set("protocol_port", listener.ProtocolPort),
		d.Set("admin_state_up", listener.AdminStateUp),
		d.Set("default_pool_id", listener.DefaultPoolID),
		d.Set("connection_limit", listener.ConnLimit),
		d.Set("http2_enable", listener.Http2Enable),
		d.Set("sni_container_refs", listener.SniContainerRefs),
		d.Set("default_tls_container_ref", listener.DefaultTlsContainerRef),
	)

	if len(listener.Loadbalancers) != 0 {
		mErr = multierror.Append(mErr, d.Set("loadbalancer_id", listener.Loadbalancers[0].ID))
	}

	// fetch tags
	if resourceTags, err := tags.Get(lbv2Client, "listeners", d.Id()).Extract(); err == nil {
		tagMap := utils.TagsToMap(resourceTags.Tags)
		mErr = multierror.Append(mErr, d.Set("tags", tagMap))
	} else {
		log.Printf("[WARN] fetching tags of ELB listener failed: %s", err)
	}

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error setting listener fields: %s", err)
	}

	return nil
}

func resourceListenerV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	lbClient, err := cfg.LoadBalancerClient(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2 client: %s", err)
	}

	// client for tags
	lbv2Client, err := cfg.ElbV2Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2.0 Client: %s", err)
	}

	//lintignore:R019
	if d.HasChanges("name", "description", "admin_state_up", "connection_limit",
		"default_tls_container_ref", "sni_container_refs", "http2_enable") {
		var updateOpts listeners.UpdateOpts
		if d.HasChange("name") {
			updateOpts.Name = d.Get("name").(string)
		}
		if d.HasChange("description") {
			desc := d.Get("description").(string)
			updateOpts.Description = &desc
		}
		if d.HasChange("connection_limit") {
			connLimit := d.Get("connection_limit").(int)
			updateOpts.ConnLimit = &connLimit
		}
		if d.HasChange("default_tls_container_ref") {
			updateOpts.DefaultTlsContainerRef = d.Get("default_tls_container_ref").(string)
		}
		if d.HasChange("sni_container_refs") {
			var sniContainerRefs []string
			if raw, ok := d.GetOk("sni_container_refs"); ok {
				for _, v := range raw.([]interface{}) {
					sniContainerRefs = append(sniContainerRefs, v.(string))
				}
			}
			updateOpts.SniContainerRefs = sniContainerRefs
		}
		if d.HasChange("admin_state_up") {
			asu := d.Get("admin_state_up").(bool)
			updateOpts.AdminStateUp = &asu
		}
		if d.HasChange("http2_enable") {
			http2Enable := d.Get("http2_enable").(bool)
			updateOpts.Http2Enable = &http2Enable
		}

		// Wait for LoadBalancer to become active before continuing
		lbID := d.Get("loadbalancer_id").(string)
		timeout := d.Timeout(schema.TimeoutUpdate)
		err = waitForLBV2LoadBalancer(ctx, lbClient, lbID, "ACTIVE", nil, timeout)
		if err != nil {
			return diag.FromErr(err)
		}

		log.Printf("[DEBUG] Updating listener %s with options: %#v", d.Id(), updateOpts)
		//lintignore:R006
		err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {
			_, err = listeners.Update(lbClient, d.Id(), updateOpts).Extract()
			if err != nil {
				return common.CheckForRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return diag.Errorf("error updating listener %s: %s", d.Id(), err)
		}

		// Wait for LoadBalancer to become active again before continuing
		err = waitForLBV2LoadBalancer(ctx, lbClient, lbID, "ACTIVE", nil, timeout)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// update tags
	if d.HasChange("tags") {
		tagErr := utils.UpdateResourceTags(lbv2Client, d, "listeners", d.Id())
		if tagErr != nil {
			return diag.Errorf("error updating tags of ELB listener:%s, err:%s", d.Id(), tagErr)
		}
	}

	return resourceListenerV2Read(ctx, d, meta)
}

func resourceListenerV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	lbClient, err := cfg.LoadBalancerClient(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating ELB v2 client: %s", err)
	}

	// Wait for LoadBalancer to become active before continuing
	lbID := d.Get("loadbalancer_id").(string)
	timeout := d.Timeout(schema.TimeoutDelete)
	err = waitForLBV2LoadBalancer(ctx, lbClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting listener %s", d.Id())
	//lintignore:R006
	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		err = listeners.Delete(lbClient, d.Id()).ExtractErr()
		if err != nil {
			return common.CheckForRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return diag.Errorf("error deleting listener %s: %s", d.Id(), err)
	}

	// Wait for LoadBalancer to become active again before continuing
	err = waitForLBV2LoadBalancer(ctx, lbClient, lbID, "ACTIVE", nil, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for Listener to delete
	err = waitForLBV2Listener(ctx, lbClient, d.Id(), "DELETED", nil, timeout)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
