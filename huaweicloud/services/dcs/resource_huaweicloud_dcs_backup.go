// ---------------------------------------------------------------
// *** AUTO GENERATED CODE ***
// @Product DCS
// ---------------------------------------------------------------

package dcs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

// @API DCS GET /v2/{project_id}/instances/{instance_id}/backups
// @API DCS POST /v2/{project_id}/instances/{instance_id}/backups
// @API DCS DELETE /v2/{project_id}/instances/{instance_id}/backups/{backup_id}
func ResourceDcsBackup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDcsBackupCreate,
		ReadContext:   resourceDcsBackupRead,
		DeleteContext: resourceDcsBackupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Specifies the ID of the DCS instance.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the description of DCS instance backup.`,
			},
			"backup_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: `Specifies the format of the DCS instance backup.`,
			},
			"backup_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Indicates the ID of the DCS instance backup.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the backup name.`,
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: `Indicates the size of the backup file (byte).`,
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the backup type.`,
			},
			"begin_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the time when the backup task is created.`,
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the time at which DCS instance backup is completed.`,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the backup status.`,
			},
			"is_support_restore": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates whether restoration is supported. Value Options: **TRUE**, **FALSE**.`,
			},
		},
	}
}

func resourceDcsBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	// createBackup: create DCS backup
	var (
		createBackupHttpUrl = "v2/{project_id}/instances/{instance_id}/backups"
		createBackupProduct = "dcs"
	)
	createBackupClient, err := cfg.NewServiceClient(createBackupProduct, region)
	if err != nil {
		return diag.Errorf("error creating DCS Client: %s", err)
	}

	instanceId := d.Get("instance_id").(string)
	createBackupPath := createBackupClient.Endpoint + createBackupHttpUrl
	createBackupPath = strings.ReplaceAll(createBackupPath, "{project_id}", createBackupClient.ProjectID)
	createBackupPath = strings.ReplaceAll(createBackupPath, "{instance_id}", instanceId)

	createBackupOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	createBackupOpt.JSONBody = utils.RemoveNil(buildCreateBackupBodyParams(d))
	var createBackupResp *http.Response
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		createBackupResp, err = createBackupClient.Request("POST", createBackupPath, &createBackupOpt)
		isRetry, err := handleOperationError(err)
		if isRetry {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("error creating DCS backup: %s", err)
	}

	createBackupRespBody, err := utils.FlattenResponse(createBackupResp)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := jmespath.Search("backup_id", createBackupRespBody)
	if err != nil {
		return diag.Errorf("error creating DCS backup: backup_id is not found in API response")
	}

	d.SetId(instanceId + "/" + id.(string))

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"waiting", "backuping"},
		Target:       []string{"succeed"},
		Refresh:      dcsBackupStatusRefreshFunc(instanceId, id.(string), createBackupClient),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for backup (%s) to become ready: %s", id.(string), err)
	}

	return resourceDcsBackupRead(ctx, d, meta)
}

func buildCreateBackupBodyParams(d *schema.ResourceData) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"remark":        utils.ValueIngoreEmpty(d.Get("description")),
		"backup_format": utils.ValueIngoreEmpty(d.Get("backup_format")),
	}
	return bodyParams
}

func resourceDcsBackupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var mErr *multierror.Error

	// getBackup: Query DCS backup
	var (
		getBackupProduct = "dcs"
	)
	getBackupClient, err := cfg.NewServiceClient(getBackupProduct, region)
	if err != nil {
		return diag.Errorf("error creating DCS Client: %s", err)
	}

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return diag.Errorf("invalid id format, must be <instance_id>/<backup_id>")
	}

	instanceID := parts[0]
	backupID := parts[1]
	backup, err := getDcsBackup(instanceID, backupID, getBackupClient)
	if err != nil {
		return diag.FromErr(err)
	}
	if backup == nil {
		return common.CheckDeletedDiag(d, golangsdk.ErrDefault404{}, "")
	}
	status := utils.PathSearch("status", backup, "")
	if status == "deleted" {
		return common.CheckDeletedDiag(d, golangsdk.ErrDefault404{}, "")
	}

	mErr = multierror.Append(
		mErr,
		d.Set("region", region),
		d.Set("backup_id", utils.PathSearch("backup_id", backup, nil)),
		d.Set("name", utils.PathSearch("backup_name", backup, nil)),
		d.Set("instance_id", utils.PathSearch("instance_id", backup, nil)),
		d.Set("size", utils.PathSearch("size", backup, nil)),
		d.Set("type", utils.PathSearch("backup_type", backup, nil)),
		d.Set("begin_time", utils.PathSearch("created_at", backup, nil)),
		d.Set("end_time", utils.PathSearch("updated_at", backup, nil)),
		d.Set("status", utils.PathSearch("status", backup, nil)),
		d.Set("description", utils.PathSearch("remark", backup, nil)),
		d.Set("is_support_restore", utils.PathSearch("is_support_restore", backup, nil)),
		d.Set("backup_format", utils.PathSearch("backup_format", backup, nil)),
	)
	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceDcsBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	// deleteBackup: Delete DCS backup
	var (
		deleteBackupHttpUrl = "v2/{project_id}/instances/{instance_id}/backups/{backup_id}"
		deleteBackupProduct = "dcs"
	)
	deleteBackupClient, err := cfg.NewServiceClient(deleteBackupProduct, region)
	if err != nil {
		return diag.Errorf("error creating DCS Client: %s", err)
	}

	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		return diag.Errorf("invalid id format, must be <instance_id>/<backup_id>")
	}
	instanceID := parts[0]
	backupID := parts[1]
	deleteBackupPath := deleteBackupClient.Endpoint + deleteBackupHttpUrl
	deleteBackupPath = strings.ReplaceAll(deleteBackupPath, "{project_id}", deleteBackupClient.ProjectID)
	deleteBackupPath = strings.ReplaceAll(deleteBackupPath, "{instance_id}", instanceID)
	deleteBackupPath = strings.ReplaceAll(deleteBackupPath, "{backup_id}", backupID)

	deleteBackupOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			204,
		},
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = deleteBackupClient.Request("DELETE", deleteBackupPath, &deleteBackupOpt)
		isRetry, err := handleOperationError(err)
		if isRetry {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("error deleting DCS backup: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"waiting", "succeed"},
		Target:       []string{"deleted"},
		Refresh:      dcsBackupStatusRefreshFunc(instanceID, backupID, deleteBackupClient),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for backup (%s) to be deleted: %s", backupID, err)
	}

	return nil
}

func dcsBackupStatusRefreshFunc(instanceId, backupId string, client *golangsdk.ServiceClient) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		backup, err := getDcsBackup(instanceId, backupId, client)
		if err != nil {
			return nil, "", err
		}
		status := utils.PathSearch("status", backup, "")
		return backup, status.(string), nil
	}
}

func getDcsBackup(instanceID, backupID string, client *golangsdk.ServiceClient) (interface{}, error) {
	// getBackup: Query DCS backup
	var (
		getBackupHttpUrl = "v2/{project_id}/instances/{instance_id}/backups"
	)

	getBackupBasePath := client.Endpoint + getBackupHttpUrl
	getBackupBasePath = strings.ReplaceAll(getBackupBasePath, "{project_id}", client.ProjectID)
	getBackupBasePath = strings.ReplaceAll(getBackupBasePath, "{instance_id}", instanceID)

	getDdmSchemasOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}

	var currentTotal int
	var getBackupPath string

	for {
		getBackupPath = getBackupBasePath + buildGetDcsBackupQueryParams(currentTotal)
		getBackupResp, err := client.Request("GET", getBackupPath, &getDdmSchemasOpt)
		if err != nil {
			return nil, fmt.Errorf("error retrieving DCS backups: %s", err)
		}
		getBackupRespBody, err := utils.FlattenResponse(getBackupResp)
		if err != nil {
			return "", err
		}
		backups := utils.PathSearch("backup_record_response", getBackupRespBody, make([]interface{}, 0)).([]interface{})
		total := utils.PathSearch("total_num", getBackupRespBody, 0)
		for _, backup := range backups {
			id := utils.PathSearch("backup_id", backup, "")
			if backupID == id {
				return backup, nil
			}
		}
		currentTotal += len(backups)
		if currentTotal >= int(total.(float64)) {
			break
		}
	}
	return nil, fmt.Errorf("error get DCS backup by backup_id (%s)", backupID)
}

func buildGetDcsBackupQueryParams(offset int) string {
	return fmt.Sprintf("?limit=10&offset=%v", offset)
}
