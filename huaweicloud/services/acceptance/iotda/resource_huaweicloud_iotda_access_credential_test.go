package iotda

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
)

func TestAccAccessCredential_basic(t *testing.T) {

	rName := "huaweicloud_iotda_access_credential.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      func(s *terraform.State) error { return nil },
		Steps: []resource.TestStep{
			{
				Config: testAccessCredential_basic("AMQP"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rName, "type", "AMQP"),
					resource.TestCheckResourceAttrSet(rName, "access_key"),
					resource.TestCheckResourceAttrSet(rName, "access_code"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAccessCredentialImportStateIdFunc(rName),
			},
		},
	})
}

func testAccessCredential_basic(acType string) string {
	return fmt.Sprintf(`
resource "huaweicloud_iotda_access_credential" "test" {
  type = "%s"
}
`, acType)
}

func testAccAccessCredentialImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return fmt.Sprintf("AMQP/%s/%s", r.Primary.ID, r.Primary.Attributes["access_code"]), nil
	}
}
