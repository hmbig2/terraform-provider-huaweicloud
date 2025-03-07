package dws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/dws"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func getClusterResourceFunc(cfg *config.Config, state *terraform.ResourceState) (interface{}, error) {
	region := acceptance.HW_REGION_NAME
	// getDwsCluster: Query the DWS cluster.
	var (
		getDwsClusterHttpUrl = "v1.0/{project_id}/clusters/{id}"
		getDwsClusterProduct = "dws"
	)
	getDwsClusterClient, err := cfg.NewServiceClient(getDwsClusterProduct, region)
	if err != nil {
		return nil, fmt.Errorf("error creating DWS Client: %s", err)
	}

	getDwsClusterPath := getDwsClusterClient.Endpoint + getDwsClusterHttpUrl
	getDwsClusterPath = strings.ReplaceAll(getDwsClusterPath, "{project_id}", getDwsClusterClient.ProjectID)
	getDwsClusterPath = strings.ReplaceAll(getDwsClusterPath, "{id}", state.Primary.ID)

	getDwsClusterOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=UTF-8",
		},
	}

	getDwsClusterResp, err := getDwsClusterClient.Request("GET", getDwsClusterPath, &getDwsClusterOpt)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DwsCluster: %s", err)
	}

	getDwsClusterRespBody, err := utils.FlattenResponse(getDwsClusterResp)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DwsCluster: %s", err)
	}

	return getDwsClusterRespBody, nil
}

func TestAccResourceCluster_basicV1(t *testing.T) {
	var obj interface{}

	resourceName := "huaweicloud_dws_cluster.test"
	name := acceptance.RandomAccResourceName()

	rc := acceptance.InitResourceCheck(
		resourceName,
		&obj,
		getClusterResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDwsCluster_basic(name, 3, dws.PublicBindTypeAuto, "cluster123@!", "bar"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "number_of_node", "3"),
					resource.TestCheckResourceAttr(resourceName, "logical_cluster_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "val"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				Config: testAccDwsCluster_basic(name, 6, dws.PublicBindTypeAuto, "cluster123@!u", "bar"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "number_of_node", "6"),
					resource.TestCheckResourceAttr(resourceName, "logical_cluster_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "val"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_pwd", "number_of_cn", "volume", "endpoints", "logical_cluster_enable"},
			},
		},
	})
}

func testAccDwsCluster_basic(rName string, numberOfNode int, publicIpBindType, password, tag string) string {
	baseNetwork := common.TestBaseNetwork(rName)

	return fmt.Sprintf(`
%s

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_dws_cluster" "test" {
  name                   = "%s"
  node_type              = "dwsk2.xlarge"
  number_of_node         = %d
  vpc_id                 = huaweicloud_vpc.test.id
  network_id             = huaweicloud_vpc_subnet.test.id
  security_group_id      = huaweicloud_networking_secgroup.test.id
  availability_zone      = data.huaweicloud_availability_zones.test.names[0]
  user_name              = "test_cluster_admin"
  user_pwd               = "%s"
  logical_cluster_enable = true

  public_ip {
    public_bind_type = "%s"
  }

  tags = {
    key = "val"
    foo = "%s"
  }
}
`, baseNetwork, rName, numberOfNode, password, publicIpBindType, tag)
}

func TestAccResourceCluster_basicV2(t *testing.T) {
	var obj interface{}

	resourceName := "huaweicloud_dws_cluster.test"
	name := acceptance.RandomAccResourceName()

	rc := acceptance.InitResourceCheck(
		resourceName,
		&obj,
		getClusterResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccDwsCluster_basicV2(name, 3, dws.PublicBindTypeAuto, "cluster123@!", "bar", 100),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "number_of_node", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "val"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.capacity", "100"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				Config: testAccDwsCluster_basicV2(name, 6, dws.PublicBindTypeAuto, "cluster123@!u", "cat", 150),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "number_of_node", "6"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "val"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "cat"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.capacity", "150"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_pwd", "number_of_cn", "volume", "endpoints"},
			},
		},
	})
}

func testAccDwsCluster_basicV2(rName string, numberOfNode int, publicIpBindType, password, tag string, volumeCap int) string {
	baseNetwork := common.TestBaseNetwork(rName)

	return fmt.Sprintf(`
%s

data "huaweicloud_availability_zones" "test" {}

data "huaweicloud_dws_flavors" "test" {
  vcpus = 4
  memory = 32
  datastore_type = "dws"
}

resource "huaweicloud_dws_cluster" "test" {
  name              = "%s"
  node_type         = "dwsk2.xlarge"
  number_of_node    = %d
  vpc_id            = huaweicloud_vpc.test.id
  network_id        = huaweicloud_vpc_subnet.test.id
  security_group_id = huaweicloud_networking_secgroup.test.id
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  user_name         = "test_cluster_admin"
  user_pwd          = "%s"
  version           = data.huaweicloud_dws_flavors.test.flavors[0].datastore_version
  number_of_cn      = 3

  public_ip {
    public_bind_type = "%s"
  }

  volume {
    type     = "SSD"
    capacity = %d
  }

  tags = {
    key = "val"
    foo = "%s"
  }
}
`, baseNetwork, rName, numberOfNode, password, publicIpBindType, volumeCap, tag)
}

func TestAccResourceCluster_BindingElb(t *testing.T) {
	var obj interface{}

	resourceName := "huaweicloud_dws_cluster.test"
	rName := acceptance.RandomAccResourceName()

	rc := acceptance.InitResourceCheck(
		resourceName,
		&obj,
		getClusterResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccCluster_bindingElb(rName, dws.PublicBindTypeAuto, "cluster123@!u"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "elb.0.name", rName+"_elb1"),
				),
			},
			{
				Config: testAccCluster_bindingElb_update(rName, dws.PublicBindTypeAuto, "cluster123@!u"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "elb.0.name", rName+"_elb2"),
				),
			},
			{
				Config: testAccCluster_bindingElb_null(rName, dws.PublicBindTypeAuto, "cluster123@!u"),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(resourceName, "elb.0.name", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_pwd", "number_of_cn", "volume", "endpoints", "elb_id"},
			},
		},
	})
}

func testAccCluster_bindingElb(rName, publicIpBindType, password string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_dws_flavors" "test" {
  vcpus          = 4
  memory         = 32
  datastore_type = "dws"
}

resource "huaweicloud_dws_cluster" "test" {
  name              = "%[2]s"
  node_type         = "dwsk2.xlarge"
  number_of_node    = 3
  vpc_id            = huaweicloud_vpc.test.id
  network_id        = huaweicloud_vpc_subnet.test.id
  security_group_id = huaweicloud_networking_secgroup.test.id
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  user_name         = "test_cluster_admin"
  user_pwd          = "%[3]s"
  version           = data.huaweicloud_dws_flavors.test.flavors[0].datastore_version
  number_of_cn      = 3
  elb_id            = huaweicloud_elb_loadbalancer.test1.id

  public_ip {
    public_bind_type = "%[4]s"
  }

  volume {
    type     = "SSD"
    capacity = 150
  }

  tags = {
    key = "val"
    foo = "bar"
  }
}
`, testAccElbV3LoadBalancerConfig_basic(rName), rName, password, publicIpBindType)
}

func testAccCluster_bindingElb_update(rName, publicIpBindType, password string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_dws_flavors" "test" {
  vcpus          = 4
  memory         = 32
  datastore_type = "dws"
}

resource "huaweicloud_dws_cluster" "test" {
  name              = "%[2]s"
  node_type         = "dwsk2.xlarge"
  number_of_node    = 3
  vpc_id            = huaweicloud_vpc.test.id
  network_id        = huaweicloud_vpc_subnet.test.id
  security_group_id = huaweicloud_networking_secgroup.test.id
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  user_name         = "test_cluster_admin"
  user_pwd          = "%[3]s"
  version           = data.huaweicloud_dws_flavors.test.flavors[0].datastore_version
  number_of_cn      = 3
  elb_id            = huaweicloud_elb_loadbalancer.test2.id

  public_ip {
    public_bind_type = "%[4]s"
  }

  volume {
    type     = "SSD"
    capacity = 150
  }

  tags = {
    key = "val"
    foo = "bar"
  }
}
`, testAccElbV3LoadBalancerConfig_basic(rName), rName, password, publicIpBindType)
}

func testAccCluster_bindingElb_null(rName, publicIpBindType, password string) string {
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_dws_flavors" "test" {
  vcpus          = 4
  memory         = 32
  datastore_type = "dws"
}

resource "huaweicloud_dws_cluster" "test" {
  name              = "%[2]s"
  node_type         = "dwsk2.xlarge"
  number_of_node    = 3
  vpc_id            = huaweicloud_vpc.test.id
  network_id        = huaweicloud_vpc_subnet.test.id
  security_group_id = huaweicloud_networking_secgroup.test.id
  availability_zone = data.huaweicloud_availability_zones.test.names[0]
  user_name         = "test_cluster_admin"
  user_pwd          = "%[3]s"
  version           = data.huaweicloud_dws_flavors.test.flavors[0].datastore_version
  number_of_cn      = 3

  public_ip {
    public_bind_type = "%[4]s"
  }

  volume {
    type     = "SSD"
    capacity = 150
  }

  tags = {
    key = "val"
    foo = "bar"
  }
}
`, testAccElbV3LoadBalancerConfig_basic(rName), rName, password, publicIpBindType)
}

func testAccElbV3LoadBalancerConfig_basic(rName string) string {
	baseNetwork := common.TestBaseNetwork(rName)
	return fmt.Sprintf(`
%[1]s

data "huaweicloud_availability_zones" "test" {}

resource "huaweicloud_elb_loadbalancer" "test1" {
  name           = "%[2]s_elb1"
  vpc_id         = huaweicloud_vpc.test.id
  ipv4_subnet_id = huaweicloud_vpc_subnet.test.ipv4_subnet_id
	
  availability_zone = [
    data.huaweicloud_availability_zones.test.names[0]
  ]

  backend_subnets = [
    huaweicloud_vpc_subnet.test.id
  ]

  protection_status = "nonProtection"

  tags = {
    key   = "value"
    owner = "terraform"
  }
}
resource "huaweicloud_elb_loadbalancer" "test2" {
  name           = "%[2]s_elb2"
  vpc_id         = huaweicloud_vpc.test.id
  ipv4_subnet_id = huaweicloud_vpc_subnet.test.ipv4_subnet_id
	
  availability_zone = [
    data.huaweicloud_availability_zones.test.names[0]
  ]

  backend_subnets = [
    huaweicloud_vpc_subnet.test.id
  ]

  protection_status = "nonProtection"

  tags = {
    key   = "value"
    owner = "terraform"
  }
}
`, baseNetwork, rName)
}
