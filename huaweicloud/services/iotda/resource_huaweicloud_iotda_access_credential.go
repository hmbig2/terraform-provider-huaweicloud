package iotda

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iotda/v5/model"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceAccessCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccessCredentialCreate,
		ReadContext:   resourceAccessCredentialNil,
		DeleteContext: resourceAccessCredentialNil,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccessCredentialImport,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"AMQP"}, false),
				Default:      "AMQP",
			},

			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_code": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAccessCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := c.HcIoTdaV5Client(region)
	if err != nil {
		return diag.Errorf("error creating IoTDA v5 client: %s", err)
	}

	createOpts := model.CreateAccessCodeRequest{
		Body: &model.CreateAccessCodeRequestBody{
			Type: utils.String(d.Get("type").(string)),
		},
	}
	log.Printf("[DEBUG] Create IoTDA access credential params: %#v", createOpts)

	resp, err := client.CreateAccessCode(&createOpts)
	if err != nil {
		return diag.Errorf("error creating IoTDA access credential: %s", err)
	}

	if resp.AccessKey == nil {
		return diag.Errorf("error creating IoTDA access credential: id is not found in API response")
	}

	d.SetId(*resp.AccessKey)
	d.Set("access_key", resp.AccessKey)
	d.Set("access_code", resp.AccessCode)
	return nil
}

// resourceAccessCredentialNil returns nil. Since query API and delete API do not exist.
func resourceAccessCredentialNil(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceAccessCredentialImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData,
	error) {
	parts := strings.SplitN(d.Id(), "/", 3)
	if len(parts) != 3 {
		err := fmt.Errorf("invalid format specified for access credential. Format must be <type>/<access_key>/<access_code>")
		return nil, err
	}

	acType := parts[0]
	accessKey := parts[1]
	accessCode := parts[2]

	d.SetId(accessKey)
	d.Set("type", acType)
	d.Set("access_key", accessKey)
	d.Set("access_code", accessCode)

	return []*schema.ResourceData{d}, nil
}
