// ---------------------------------------------------------------
// *** AUTO GENERATED CODE ***
// @Product DataArts
// ---------------------------------------------------------------

package dataarts

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/chnsz/golangsdk"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/jmespath/go-jmespath"
)

func ResourceDataConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataConnectionCreate,
		UpdateContext: resourceDataConnectionUpdate,
		ReadContext:   resourceDataConnectionRead,
		DeleteContext: resourceDataConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDataConnectionImportState,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `The data connect name.`,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z-_0-9]*$`),
						"the input is invalid"),
					validation.StringLenBetween(1, 128),
				),
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `The type of data connect.`,
				ValidateFunc: validation.StringInSlice([]string{
					"DWS", "DLI", "ORACLE", "RDS", "DIS", "SHELL", "MRS Hive", "MRS HBase", "MRS Kafka", "MRS Ranger", "MRS Spark", "MRS Presto",
				}, false),
			},
			"config": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: `The config of the data connect.`,
			},
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `The worksapce to which the data connect belongs.`,
			},
			"agent_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `The agent ID which used by the data connect.`,
			},
			"agent_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `The agent name which used by the data connect.`,
			},
		},
	}
}

func resourceDataConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	// createDataConnection: create a data connection.
	var (
		createDataConnectionHttpUrl = "v1/{project_id}/data-connections"
		createDataConnectionProduct = "dataarts"
	)
	createDataConnectionClient, err := config.NewServiceClient(createDataConnectionProduct, region)
	if err != nil {
		return diag.Errorf("error creating DataConnection Client: %s", err)
	}

	createDataConnectionPath := createDataConnectionClient.Endpoint + createDataConnectionHttpUrl
	createDataConnectionPath = strings.ReplaceAll(createDataConnectionPath, "{project_id}", createDataConnectionClient.ProjectID)

	createDataConnectionOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
		MoreHeaders: map[string]string{
			"Content-Type": "application/json",
			"X-Language":   "en-us",
			"workspace":    d.Get("workspace").(string),
		},
	}
	createDataConnectionOpt.JSONBody = utils.RemoveNil(buildCreateDataConnectionBodyParams(d, config))
	createDataConnectionResp, err := createDataConnectionClient.Request("POST", createDataConnectionPath, &createDataConnectionOpt)
	if err != nil {
		return diag.Errorf("error creating DataConnection: %s", err)
	}

	createDataConnectionRespBody, err := utils.FlattenResponse(createDataConnectionResp)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := jmespath.Search("data_connection_id", createDataConnectionRespBody)
	if err != nil {
		return diag.Errorf("error creating DataConnection: ID is not found in API response")
	}
	d.SetId(id.(string))

	return resourceDataConnectionRead(ctx, d, meta)
}

func buildCreateDataConnectionBodyParams(d *schema.ResourceData, config *config.Config) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"data_source_vos": []map[string]interface{}{{
			"dw_name":    utils.ValueIngoreEmpty(d.Get("name")),
			"dw_type":    utils.ValueIngoreEmpty(d.Get("type")),
			"dw_config":  utils.ValueIngoreEmpty(d.Get("config")),
			"agent_id":   utils.ValueIngoreEmpty(d.Get("agent_id")),
			"agent_name": utils.ValueIngoreEmpty(d.Get("agent_name")),
		},
		},
	}
	return bodyParams
}

func resourceDataConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	var mErr *multierror.Error

	// getDataConnection: Query the data connection
	var (
		getDataConnectionHttpUrl = "v1/{project_id}/data-connections/{id}"
		getDataConnectionProduct = "dataarts"
	)
	getDataConnectionClient, err := config.NewServiceClient(getDataConnectionProduct, region)
	if err != nil {
		return diag.Errorf("error creating DataConnection Client: %s", err)
	}

	getDataConnectionPath := getDataConnectionClient.Endpoint + getDataConnectionHttpUrl
	getDataConnectionPath = strings.ReplaceAll(getDataConnectionPath, "{project_id}", getDataConnectionClient.ProjectID)
	getDataConnectionPath = strings.ReplaceAll(getDataConnectionPath, "{id}", d.Id())

	getDataConnectionOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
		MoreHeaders: map[string]string{
			"Content-Type": "application/json",
			"X-Language":   "en-us",
			"workspace":    d.Get("workspace").(string),
		},
	}
	getDataConnectionResp, err := getDataConnectionClient.Request("GET", getDataConnectionPath, &getDataConnectionOpt)

	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving DataConnection")
	}

	getDataConnectionRespBody, err := utils.FlattenResponse(getDataConnectionResp)
	if err != nil {
		return diag.FromErr(err)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("region", region),
		d.Set("name", utils.PathSearch("dw_name", getDataConnectionRespBody, nil)),
		d.Set("type", utils.PathSearch("dw_type", getDataConnectionRespBody, nil)),
		d.Set("config", utils.PathSearch("dw_config", getDataConnectionRespBody, nil)),
		d.Set("agent_id", utils.PathSearch("agent_id", getDataConnectionRespBody, nil)),
		d.Set("agent_name", utils.PathSearch("agent_name", getDataConnectionRespBody, nil)),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceDataConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	updateDataConnectionhasChanges := []string{
		"name",
		"type",
		"config",
		"agent_id",
		"agent_name",
	}

	if d.HasChanges(updateDataConnectionhasChanges...) {
		// updateDataConnection: update the data connection
		var (
			updateDataConnectionHttpUrl = "v1/{project_id}/data-connections/{id}"
			updateDataConnectionProduct = "dataarts"
		)
		updateDataConnectionClient, err := config.NewServiceClient(updateDataConnectionProduct, region)
		if err != nil {
			return diag.Errorf("error creating DataConnection Client: %s", err)
		}

		updateDataConnectionPath := updateDataConnectionClient.Endpoint + updateDataConnectionHttpUrl
		updateDataConnectionPath = strings.ReplaceAll(updateDataConnectionPath, "{project_id}", updateDataConnectionClient.ProjectID)
		updateDataConnectionPath = strings.ReplaceAll(updateDataConnectionPath, "{id}", d.Id())

		updateDataConnectionOpt := golangsdk.RequestOpts{
			KeepResponseBody: true,
			OkCodes: []int{
				200,
			},
			MoreHeaders: map[string]string{
				"Content-Type": "application/json",
				"X-Language":   "en-us",
				"workspace":    d.Get("workspace").(string),
			},
		}
		updateDataConnectionOpt.JSONBody = utils.RemoveNil(buildUpdateDataConnectionBodyParams(d, config))
		_, err = updateDataConnectionClient.Request("PUT", updateDataConnectionPath, &updateDataConnectionOpt)
		if err != nil {
			return diag.Errorf("error updating DataConnection: %s", err)
		}
	}
	return resourceDataConnectionRead(ctx, d, meta)
}

func buildUpdateDataConnectionBodyParams(d *schema.ResourceData, config *config.Config) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"data_source_vos": []map[string]interface{}{{
			"dw_name":    utils.ValueIngoreEmpty(d.Get("name")),
			"dw_type":    utils.ValueIngoreEmpty(d.Get("type")),
			"dw_config":  utils.ValueIngoreEmpty(d.Get("config")),
			"agent_id":   utils.ValueIngoreEmpty(d.Get("agent_id")),
			"agent_name": utils.ValueIngoreEmpty(d.Get("agent_name")),
		},
		},
	}
	return bodyParams
}

func resourceDataConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)

	// deleteDataConnection: missing operation notes
	var (
		deleteDataConnectionHttpUrl = "v1/{project_id}/data-connections/{id}"
		deleteDataConnectionProduct = "dataarts"
	)
	deleteDataConnectionClient, err := config.NewServiceClient(deleteDataConnectionProduct, region)
	if err != nil {
		return diag.Errorf("error creating DataConnection Client: %s", err)
	}

	deleteDataConnectionPath := deleteDataConnectionClient.Endpoint + deleteDataConnectionHttpUrl
	deleteDataConnectionPath = strings.ReplaceAll(deleteDataConnectionPath, "{project_id}", deleteDataConnectionClient.ProjectID)
	deleteDataConnectionPath = strings.ReplaceAll(deleteDataConnectionPath, "{id}", d.Id())

	deleteDataConnectionOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			204,
		},
		MoreHeaders: map[string]string{
			"Content-Type": "application/json",
			"X-Language":   "en-us",
			"workspace":    d.Get("workspace").(string),
		},
	}
	_, err = deleteDataConnectionClient.Request("DELETE", deleteDataConnectionPath, &deleteDataConnectionOpt)
	if err != nil {
		return diag.Errorf("error deleting DaddtaConnection: %s", err)
	}

	return nil
}

func resourceDataConnectionImportState(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format specified for import id, must be <workspace>/<id>")
	}

	d.Set("workspace", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
