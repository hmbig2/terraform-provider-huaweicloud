package dataarts

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chnsz/golangsdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func getDataConnectionResourceFunc(config *config.Config, state *terraform.ResourceState) (interface{}, error) {
	region := acceptance.HW_REGION_NAME
	// getDataConnection: Query the data connection
	var (
		getDataConnectionHttpUrl = "v1/{project_id}/data-connections/{id}"
		getDataConnectionProduct = "dataarts"
	)
	getDataConnectionClient, err := config.NewServiceClient(getDataConnectionProduct, region)
	if err != nil {
		return nil, fmt.Errorf("error creating DataConnection Client: %s", err)
	}

	getDataConnectionPath := getDataConnectionClient.Endpoint + getDataConnectionHttpUrl
	getDataConnectionPath = strings.ReplaceAll(getDataConnectionPath, "{project_id}", getDataConnectionClient.ProjectID)
	getDataConnectionPath = strings.ReplaceAll(getDataConnectionPath, "{id}", state.Primary.ID)

	getDataConnectionOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
		MoreHeaders: map[string]string{
			"Content-Type": "application/json",
			"X-Language":   "en-us",
			"workspace":    state.Primary.Attributes["worksapce"],
		},
	}
	getDataConnectionResp, err := getDataConnectionClient.Request("GET", getDataConnectionPath, &getDataConnectionOpt)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DataConnection: %s", err)
	}
	return utils.FlattenResponse(getDataConnectionResp)
}

func TestAccDataConnection_basic(t *testing.T) {
	var obj interface{}

	name := acceptance.RandomAccResourceName()
	rName := "huaweicloud_dataarts_data_connection.test"

	rc := acceptance.InitResourceCheck(
		rName,
		&obj,
		getDataConnectionResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testDataConnection_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name),
					resource.TestCheckResourceAttr(rName, "type", "DLI"),
				),
			},
			{
				Config: testDataConnection_basic_update(name + "update"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "name", name+"update"),
					resource.TestCheckResourceAttr(rName, "type", "DLI"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testDataConnectionImportState(rName),
			},
		},
	})
}

func testDataConnection_basic(name string) string {
	return fmt.Sprintf(`
resource "huaweicloud_dataarts_data_connection" "test" {
  name = "%s"
  type = "DLI"
  config = {}
}
`, name)
}

func testDataConnection_basic_update(name string) string {
	return fmt.Sprintf(`
resource "huaweicloud_dataarts_data_connection" "test" {
  name = "%s"
  type = "DLI"
  config = {}
}
`, name)
}

func testDataConnectionImportState(name string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("Resource (%s) not found: %s", name, rs)
		}
		if rs.Primary.Attributes["workspace"] == "" {
			return "", fmt.Errorf("Attribute (workspace) of Resource (%s) not found: %s", name, rs)
		}
		if rs.Primary.ID == "" {
			return "", fmt.Errorf("Attribute (ID) of Resource (%s) not found: %s", name, rs)
		}

		return rs.Primary.Attributes["workspace"] + "/" +
			rs.Primary.ID, nil
	}
}
