package cpts

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cpts/v1/model"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
)

func getVariableResourceFunc(conf *config.Config, state *terraform.ResourceState) (interface{}, error) {
	client, err := conf.HcCptsV3Client(acceptance.HW_REGION_NAME)
	if err != nil {
		return nil, fmt.Errorf("error creating CPTS v1 client: %s", err)
	}

	id, err := strconv.ParseInt(state.Primary.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("the Task ID must be integer: %s", err)
	}

	request := &model.ShowTaskRequest{
		TaskId: int32(id),
	}

	return client.ShowTask(request)
}

func TestAccVariable_basic(t *testing.T) {
	var obj model.CreateVariableResponse

	rName := acceptance.RandomAccResourceName()
	eName := acceptance.RandomAccResourceName()
	resourceName := "huaweicloud_cpts_project_variable.test"
	resourceNameEnum := "huaweicloud_cpts_project_variable.test2"

	rc := acceptance.InitResourceCheck(
		resourceName,
		&obj,
		getVariableResourceFunc,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testVariable_basic(rName, eName),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cpts_project_id",
						"huaweicloud_cpts_project.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "value"),
					resource.TestCheckResourceAttr(resourceNameEnum, "name", rName),
					resource.TestCheckResourceAttr(resourceNameEnum, "type", "2"),
					resource.TestCheckResourceAttrPair(resourceNameEnum, "cpts_project_id",
						"huaweicloud_cpts_project.test", "id"),
					resource.TestCheckResourceAttr(resourceNameEnum, "value.#", "3"),
				),
			},
			{
				Config: testVariable_basic_update(rName, eName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cpts_project_id",
						"huaweicloud_cpts_project.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "value"),
					resource.TestCheckResourceAttr(resourceNameEnum, "name", rName),
					resource.TestCheckResourceAttr(resourceNameEnum, "type", "2"),
					resource.TestCheckResourceAttrPair(resourceNameEnum, "cpts_project_id",
						"huaweicloud_cpts_project.test", "id"),
					resource.TestCheckResourceAttr(resourceNameEnum, "value.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testVariable_basic(rName, eName string) string {
	projectTaskConfig := testProject_basic(rName, "created by acc test")
	return fmt.Sprintf(`
%s

resource "huaweicloud_cpts_project_variable" "test" {
  name            = "%s"
  cpts_project_id = huaweicloud_cpts_project.test.id
  type            = 1
  value           = ["1", "10"]
}

resource "huaweicloud_cpts_project_variable" "test2" {
  name            = "%s"
  cpts_project_id = huaweicloud_cpts_project.test.id
  type            = 2
  value           = ["enum_1", "enum_2", "enum_3"]
}
`, projectTaskConfig, rName, eName)
}

func testVariable_basic_update(rName, eName string) string {
	projectTaskConfig := testProject_basic(rName, "created by acc test")
	return fmt.Sprintf(`
%s

resource "huaweicloud_cpts_project_variable" "test" {
  name            = "%s"
  cpts_project_id = huaweicloud_cpts_project.test.id
  type            = 1
  value           = ["2", "20"]
}

resource "huaweicloud_cpts_project_variable" "test2" {
  name            = "%s"
  cpts_project_id = huaweicloud_cpts_project.test.id
  type            = 2
  value           = ["enum_1", "enum_4"]
}
`, projectTaskConfig, rName, eName)
}
