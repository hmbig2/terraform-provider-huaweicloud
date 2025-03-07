package iam

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/chnsz/golangsdk/openstack/identity/v3.0/users"
	identity_users "github.com/chnsz/golangsdk/openstack/identity/v3/users"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

// @API IAM GET /v3.0/OS-USER/users/{userID}
// @API IAM PUT /v3.0/OS-USER/users/{userID}
// @API IAM POST /v3.0/OS-USER/users
// @API IAM DELETE /v3/users/{userID}
func ResourceIdentityUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityUserCreate,
		ReadContext:   resourceIdentityUserRead,
		UpdateContext: resourceIdentityUserUpdate,
		DeleteContext: resourceIdentityUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"country_code"},
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9]{0,32}$"),
					"the phone number must have a maximum of 32 digits"),
			},
			"country_code": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"phone"},
			},
			"external_identity_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"external_identity_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"external_identity_id"},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"pwd_reset": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"access_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"default", "programmatic", "console",
				}, false),
			},
			"password_strength": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_login": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func buildExternalIdentityType(d *schema.ResourceData) string {
	// external_identity_type is valid only when external_identity_id is specified.
	if _, ok := d.GetOk("external_identity_id"); !ok {
		return ""
	}

	if v, ok := d.GetOk("external_identity_type"); ok {
		return v.(string)
	}

	// the default value of external_identity_type is TenantIdp
	return "TenantIdp"
}

func resourceIdentityUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	iamClient, err := cfg.IAMV3Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating IAM client: %s", err)
	}

	if cfg.DomainID == "" {
		return diag.Errorf("the domain_id must be specified in the provider configuration")
	}

	enabled := d.Get("enabled").(bool)
	reset := d.Get("pwd_reset").(bool)
	createOpts := users.CreateOpts{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Email:         d.Get("email").(string),
		Phone:         d.Get("phone").(string),
		AreaCode:      d.Get("country_code").(string),
		AccessMode:    d.Get("access_type").(string),
		XUserID:       d.Get("external_identity_id").(string),
		XUserType:     buildExternalIdentityType(d),
		Enabled:       &enabled,
		PasswordReset: &reset,
		DomainID:      cfg.DomainID,
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	// Add password here so it wouldn't go in the above log entry
	createOpts.Password = d.Get("password").(string)

	user, err := users.Create(iamClient, createOpts).Extract()
	if err != nil {
		return diag.Errorf("error creating IAM user: %s", err)
	}

	d.SetId(user.ID)
	return resourceIdentityUserRead(ctx, d, meta)
}

func resourceIdentityUserRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	iamClient, err := cfg.IAMV3Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating IAM client: %s", err)
	}

	user, err := users.Get(iamClient, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "user")
	}

	log.Printf("[DEBUG] Retrieved IAM user: %#v", user)
	mErr := multierror.Append(nil,
		d.Set("enabled", user.Enabled),
		d.Set("name", user.Name),
		d.Set("description", user.Description),
		d.Set("email", user.Email),
		d.Set("phone", normalizePhoneNumber(user.Phone)),
		d.Set("country_code", user.AreaCode),
		d.Set("access_type", user.AccessMode),
		d.Set("password_strength", user.PasswordStrength),
		d.Set("pwd_reset", user.PasswordStatus),
		d.Set("create_time", user.CreateAt),
		d.Set("last_login", user.LastLogin),
		d.Set("external_identity_id", user.XUserID),
		d.Set("external_identity_type", user.XUserType),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error setting IAM user fields: %s", err)
	}
	return nil
}

func normalizePhoneNumber(raw string) string {
	phone := raw

	rawList := strings.Split(raw, "-")
	if len(rawList) > 1 {
		phone = rawList[1]
	}

	return phone
}

func resourceIdentityUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	iamClient, err := cfg.IAMV3Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating IAM client: %s", err)
	}

	var updateOpts users.UpdateOpts
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		updateOpts.Description = utils.String(d.Get("description").(string))
	}

	if d.HasChange("email") {
		updateOpts.Email = d.Get("email").(string)
	}

	if d.HasChanges("country_code", "phone") {
		updateOpts.AreaCode = d.Get("country_code").(string)
		updateOpts.Phone = d.Get("phone").(string)
	}

	if d.HasChanges("external_identity_id", "external_identity_type") {
		updateOpts.XUserID = utils.String(d.Get("external_identity_id").(string))
		updateOpts.XUserType = utils.String(buildExternalIdentityType(d))
	}

	if d.HasChange("access_type") {
		updateOpts.AccessMode = d.Get("access_type").(string)
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		updateOpts.Enabled = &enabled
	}

	if d.HasChange("pwd_reset") {
		reset := d.Get("pwd_reset").(bool)
		updateOpts.PasswordReset = &reset
	}

	log.Printf("[DEBUG] Update Options: %#v", updateOpts)

	// Add password here so it wouldn't go in the above log entry
	if d.HasChange("password") {
		updateOpts.Password = d.Get("password").(string)
	}

	_, err = users.Update(iamClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return diag.Errorf("error updating IAM user: %s", err)
	}

	return resourceIdentityUserRead(ctx, d, meta)
}

func resourceIdentityUserDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	identityClient, err := cfg.IdentityV3Client(cfg.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating IAM client: %s", err)
	}

	err = identity_users.Delete(identityClient, d.Id()).ExtractErr()
	if err != nil {
		return diag.Errorf("error deleting IAM user: %s", err)
	}

	return nil
}
