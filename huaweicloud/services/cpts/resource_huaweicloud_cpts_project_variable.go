package cpts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cpts/v1/model"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
)

const (
	variableTypeInteger = 1
	variableTypeEnum    = 2
)

func ResourceProjectVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceProjectVariableCreate,
		UpdateContext: ResourceProjectVariableUpdate,
		ReadContext:   ResourceProjectVariableRead,
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

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cpts_project_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 2),
			},

			"value": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"is_quoted": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"variable_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func ResourceProjectVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := c.HcCptsV3Client(region)
	if err != nil {
		return diag.Errorf("error creating CPTS v1 client: %s", err)
	}

	projectId := int32(d.Get("cpts_project_id").(int))
	variableType := int32(d.Get("type").(int))

	rangeValue, err := buildRangeValueParams(variableType, d.Get("value").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	createOpts := &model.CreateVariableRequest{
		TestSuiteId: projectId,
		Body: &[]model.CreateVariableRequestBody{
			{
				Name:         d.Get("name").(string),
				VariableType: variableType,
				Variable:     rangeValue,
			},
		},
	}

	response, err := client.CreateVariable(createOpts)
	if err != nil {
		return diag.Errorf("error creating CPTS variable: %s", err)
	}

	if response.Json == nil && response.Json.VariableId == nil {
		return diag.Errorf("error creating CPTS variable: id not found in api response")
	}

	id := fmt.Sprintf("%d/%d", projectId, *response.Json.VariableId)
	d.SetId(id)
	return ResourceProjectVariableRead(ctx, d, meta)
}

func ResourceProjectVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := c.HcCptsV3Client(region)
	if err != nil {
		return diag.Errorf("error creating CPTS v1 client: %s", err)
	}

	projectId, variableId, err := ParseVariableInfoFromId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := client.ListVariables(&model.ListVariablesRequest{
		VariableType: 0,
		TestSuiteId:  projectId,
	})
	if err != nil {
		return common.CheckDeletedDiag(d, err, "the project is not found")
	}

	variableDetail := filterVariables(response.VariableList, variableId)

	if variableDetail == nil {
		return common.CheckDeletedDiag(d, nil, "the variable is not found in the project")
	}

	mErr := multierror.Append(nil,
		d.Set("region", region),
		d.Set("name", variableDetail.Name),
		d.Set("cpts_project_id", int(projectId)),
		d.Set("type", int(*variableDetail.VariableType)),
		d.Set("value", flattenVariablesItems(variableDetail.Variable)),
		d.Set("is_quoted", *variableDetail.IsQuoted),
		d.Set("variable_id", *variableDetail.Id),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func ResourceProjectVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := c.HcCptsV3Client(region)
	if err != nil {
		return diag.Errorf("error creating CPTS v1 client: %s", err)
	}

	variableType := int32(d.Get("type").(int))
	rangeValue, err := buildRangeValueParams(variableType, d.Get("value").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.UpdateVariable(&model.UpdateVariableRequest{
		TestSuiteId: int32(d.Get("cpts_project_id").(int)),
		Body: &[]model.UpdateVariableRequestBody{
			{
				Id:           int32(d.Get("variable_id").(int)),
				Name:         d.Get("name").(string),
				VariableType: variableType,
				Variable:     rangeValue,
			},
		},
	})

	if err != nil {
		return diag.Errorf("error updating the CPTS variables %q: %s", d.Id(), err)
	}

	return ResourceProjectVariableRead(ctx, d, meta)
}

func ResourceProjectVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	_, err := c.HcCptsV3Client(region)
	if err != nil {
		return diag.Errorf("error creating CPTS v1 client: %s", err)
	}

	//TODO no API support

	return nil
}

func ParseVariableInfoFromId(id string) (int32, int32, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return 0, 0, fmt.Errorf("invalid format specified for CPTS variable. Format must be <project id>/<variable id>")
	}

	projectId, err := strconv.ParseInt(idParts[0], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("the project ID must be integer: %s", err)
	}
	variableId, err := strconv.ParseInt(idParts[1], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("the variable ID must be integer: %s", err)
	}

	return int32(projectId), int32(variableId), nil
}

func filterVariables(variableList *[]model.VariableDetail, variableId int32) *model.VariableDetail {
	for _, v := range *variableList {
		if *v.Id == variableId {
			return &v
		}
	}
	return nil
}

func flattenVariablesItems(variableList *[]interface{}) []string {
	rs := make([]string, len(*variableList))
	for _, v := range *variableList {
		rs = append(rs, fmt.Sprint(v))
	}
	return rs
}

func buildRangeValueParams(variableType int32, raw []interface{}) ([]interface{}, error) {
	var rangeValue []interface{}
	if variableType == variableTypeInteger {
		if len(raw) != 2 {
			return nil, fmt.Errorf("just need two items of variables to specify value range when type is integer")
		}
		min, err := strconv.Atoi(raw[0].(string))
		if err != nil {
			return nil, fmt.Errorf("error parsing the value range")
		}
		max, err := strconv.Atoi(raw[1].(string))
		if err != nil {
			return nil, fmt.Errorf("error parsing the value range")
		}
		rangeValue = []interface{}{min, max}
	} else {
		rangeValue = raw
	}
	return rangeValue, nil
}
