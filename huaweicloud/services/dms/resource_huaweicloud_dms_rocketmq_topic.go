package dms

import (
	"context"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/chnsz/golangsdk"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/jmespath/go-jmespath"
)

func ResourceDmsRocketMQTopic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsRocketMQTopicCreate,
		UpdateContext: resourceDmsRocketMQTopicUpdate,
		ReadContext:   resourceDmsRocketMQTopicRead,
		DeleteContext: resourceDmsRocketMQTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `Specifies the name of the topic.`,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z|%-_0-9]*$`),
						"the input is invalid"),
					validation.StringLenBetween(3, 64),
				),
			},
			"brokers": {
				Type:        schema.TypeList,
				Elem:        DmsRocketMQTopicBrokerRefSchema(),
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the list of associated brokers of the topic.`,
			},
			"queue_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  `Specifies the number of queues.`,
				ValidateFunc: validation.IntBetween(1, 50),
			},
			"permission": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the permissions of the topic.`,
				ValidateFunc: validation.StringInSlice([]string{
					"all", "sub", "pub",
				}, false),
			},
			"total_read_queue_num": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the total number of read queues.`,
			},
			"total_write_queue_num": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the total number of write queues.`,
			},
		},
	}
}

func DmsRocketMQTopicBrokerRefSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Indicates the name of the broker.`,
			},
			"read_queue_num": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: `Indicates the read queues number of the broker.`,
			},
			"write_queue_num": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: `Indicates the read queues number of the broker.`,
			},
		},
	}
	return &sc
}

func resourceDmsRocketMQTopicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	// createRocketmqTopic: create DMS rocketmq topic
	var (
		createRocketmqTopicHttpUrl = "v2/{project_id}/instances/{instance_id}/topics"
		createRocketmqTopicProduct = "dms"
	)
	createRocketmqTopicClient, err := config.NewServiceClient(createRocketmqTopicProduct, region)
	if err != nil {
		return diag.Errorf("error creating DmsRocketMQTopic Client: %s", err)
	}

	instanceID := d.Get("instance_id").(string)
	createRocketmqTopicPath := createRocketmqTopicClient.Endpoint + createRocketmqTopicHttpUrl
	createRocketmqTopicPath = strings.ReplaceAll(createRocketmqTopicPath, "{project_id}", createRocketmqTopicClient.ProjectID)
	createRocketmqTopicPath = strings.ReplaceAll(createRocketmqTopicPath, "{instance_id}", instanceID)

	createRocketmqTopicOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	createRocketmqTopicOpt.JSONBody = utils.RemoveNil(buildCreateRocketmqTopicBodyParams(d, config))
	createRocketmqTopicResp, err := createRocketmqTopicClient.Request("POST", createRocketmqTopicPath, &createRocketmqTopicOpt)
	if err != nil {
		return diag.Errorf("error creating DmsRocketMQTopic: %s", err)
	}

	createRocketmqTopicRespBody, err := utils.FlattenResponse(createRocketmqTopicResp)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := jmespath.Search("id", createRocketmqTopicRespBody)
	if err != nil {
		return diag.Errorf("error creating DmsRocketMQTopic: ID is not found in API response")
	}
	d.SetId(instanceID + "/" + id.(string))

	return resourceDmsRocketMQTopicUpdate(ctx, d, meta)
}

func buildCreateRocketmqTopicBodyParams(d *schema.ResourceData, config *config.Config) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"name":       utils.ValueIngoreEmpty(d.Get("name")),
		"brokers":    buildCreateRocketmqTopicBrokersChildBody(d),
		"queue_num":  utils.ValueIngoreEmpty(d.Get("queue_num")),
		"permission": utils.ValueIngoreEmpty(d.Get("permission")),
	}
	return bodyParams
}

func buildCreateRocketmqTopicBrokersChildBody(d *schema.ResourceData) []string {
	rawParams := d.Get("brokers").([]interface{})
	if len(rawParams) == 0 {
		return nil
	}
	params := make([]string, 0)
	for _, param := range rawParams {
		params = append(params, utils.PathSearch("name", param, "").(string))
	}
	return params
}

func resourceDmsRocketMQTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	updateRocketmqTopichasChanges := []string{
		"total_read_queue_num",
		"total_write_queue_num",
		"permission",
	}

	if d.HasChanges(updateRocketmqTopichasChanges...) {
		// updateRocketmqTopic: update DMS rocketmq topic
		var (
			updateRocketmqTopicHttpUrl = "v2/{project_id}/instances/{instance_id}/topics/{topic}"
			updateRocketmqTopicProduct = "dms"
		)
		updateRocketmqTopicClient, err := config.NewServiceClient(updateRocketmqTopicProduct, region)
		if err != nil {
			return diag.Errorf("error creating DmsRocketMQTopic Client: %s", err)
		}

		parts := strings.SplitN(d.Id(), "/", 2)
		if len(parts) != 2 {
			return diag.Errorf("invalid id format, must be <instance_id>/<topic>")
		}
		instanceID := parts[0]
		topic := parts[1]
		updateRocketmqTopicPath := updateRocketmqTopicClient.Endpoint + updateRocketmqTopicHttpUrl
		updateRocketmqTopicPath = strings.ReplaceAll(updateRocketmqTopicPath, "{project_id}", updateRocketmqTopicClient.ProjectID)
		updateRocketmqTopicPath = strings.ReplaceAll(updateRocketmqTopicPath, "{instance_id}", instanceID)
		updateRocketmqTopicPath = strings.ReplaceAll(updateRocketmqTopicPath, "{topic}", topic)

		updateRocketmqTopicOpt := golangsdk.RequestOpts{
			KeepResponseBody: true,
			OkCodes: []int{
				204,
			},
		}
		updateRocketmqTopicOpt.JSONBody = utils.RemoveNil(buildUpdateRocketmqTopicBodyParams(d, config))
		_, err = updateRocketmqTopicClient.Request("PUT", updateRocketmqTopicPath, &updateRocketmqTopicOpt)
		if err != nil {
			return diag.Errorf("error updating DmsRocketMQTopic: %s", err)
		}
	}
	return resourceDmsRocketMQTopicRead(ctx, d, meta)
}

func buildUpdateRocketmqTopicBodyParams(d *schema.ResourceData, config *config.Config) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"read_queue_num":  utils.ValueIngoreEmpty(d.Get("total_read_queue_num")),
		"write_queue_num": utils.ValueIngoreEmpty(d.Get("total_write_queue_num")),
		"permission":      utils.ValueIngoreEmpty(d.Get("permission")),
	}
	return bodyParams
}

func resourceDmsRocketMQTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	var mErr *multierror.Error

	// getRocketmqTopic: query DMS rocketmq topic
	var (
		getRocketmqTopicHttpUrl = "v2/{project_id}/instances/{instance_id}/topics/{topic}"
		getRocketmqTopicProduct = "dms"
	)
	getRocketmqTopicClient, err := config.NewServiceClient(getRocketmqTopicProduct, region)
	if err != nil {
		return diag.Errorf("error creating DmsRocketMQTopic Client: %s", err)
	}

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return diag.Errorf("invalid id format, must be <instance_id>/<topic>")
	}
	instanceID := parts[0]
	topic := parts[1]
	getRocketmqTopicPath := getRocketmqTopicClient.Endpoint + getRocketmqTopicHttpUrl
	getRocketmqTopicPath = strings.ReplaceAll(getRocketmqTopicPath, "{project_id}", getRocketmqTopicClient.ProjectID)
	getRocketmqTopicPath = strings.ReplaceAll(getRocketmqTopicPath, "{instance_id}", instanceID)
	getRocketmqTopicPath = strings.ReplaceAll(getRocketmqTopicPath, "{topic}", topic)

	getRocketmqTopicOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	getRocketmqTopicResp, err := getRocketmqTopicClient.Request("GET", getRocketmqTopicPath, &getRocketmqTopicOpt)

	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving DmsRocketMQTopic")
	}

	getRocketmqTopicRespBody, err := utils.FlattenResponse(getRocketmqTopicResp)
	if err != nil {
		return diag.FromErr(err)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("region", region),
		d.Set("name", utils.PathSearch("name", getRocketmqTopicRespBody, nil)),
		d.Set("total_read_queue_num", utils.PathSearch("total_read_queue_num",
			getRocketmqTopicRespBody, nil)),
		d.Set("total_write_queue_num", utils.PathSearch("total_write_queue_num",
			getRocketmqTopicRespBody, nil)),
		d.Set("permission", utils.PathSearch("permission", getRocketmqTopicRespBody, nil)),
		d.Set("brokers", flattenGetRocketmqTopicResponseBodyBrokerRef(getRocketmqTopicRespBody)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenGetRocketmqTopicResponseBodyBrokerRef(resp interface{}) []interface{} {
	if resp == nil {
		return nil
	}
	curJson := utils.PathSearch("brokers", resp, make([]interface{}, 0))
	curArray := curJson.([]interface{})
	rst := make([]interface{}, 0, len(curArray))
	for _, v := range curArray {
		rst = append(rst, map[string]interface{}{
			"name":            utils.PathSearch("broker_name", v, nil),
			"read_queue_num":  utils.PathSearch("read_queue_num", v, nil),
			"write_queue_num": utils.PathSearch("write_queue_num", v, nil),
		})
	}
	return rst
}

func resourceDmsRocketMQTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	// deleteRocketmqTopic: delete DMS rocketmq topic
	var (
		deleteRocketmqTopicHttpUrl = "v2/{project_id}/instances/{instance_id}/topics/{topic}"
		deleteRocketmqTopicProduct = "dms"
	)
	deleteRocketmqTopicClient, err := config.NewServiceClient(deleteRocketmqTopicProduct, region)
	if err != nil {
		return diag.Errorf("error creating DmsRocketMQTopic Client: %s", err)
	}

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return diag.Errorf("invalid id format, must be <instance_id>/<topic>")
	}
	instanceID := parts[0]
	topic := parts[1]
	deleteRocketmqTopicPath := deleteRocketmqTopicClient.Endpoint + deleteRocketmqTopicHttpUrl
	deleteRocketmqTopicPath = strings.ReplaceAll(deleteRocketmqTopicPath, "{project_id}", deleteRocketmqTopicClient.ProjectID)
	deleteRocketmqTopicPath = strings.ReplaceAll(deleteRocketmqTopicPath, "{instance_id}", instanceID)
	deleteRocketmqTopicPath = strings.ReplaceAll(deleteRocketmqTopicPath, "{topic}", topic)

	deleteRocketmqTopicOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			204,
		},
	}
	_, err = deleteRocketmqTopicClient.Request("DELETE", deleteRocketmqTopicPath, &deleteRocketmqTopicOpt)
	if err != nil {
		return diag.Errorf("error deleting DmsRocketMQTopic: %s", err)
	}

	return nil
}