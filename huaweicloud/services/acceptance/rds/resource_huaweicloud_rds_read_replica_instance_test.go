package rds

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/chnsz/golangsdk/openstack/rds/v3/instances"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/acceptance"
)

func TestAccReadReplicaInstance_basic(t *testing.T) {
	var replica instances.RdsInstanceResponse
	name := acceptance.RandomAccResourceName()
	updateName := acceptance.RandomAccResourceName()
	resourceType := "huaweicloud_rds_read_replica_instance"
	resourceName := "huaweicloud_rds_read_replica_instance.test"
	dbPwd := fmt.Sprintf("%s%s%d", acctest.RandString(5),
		acctest.RandStringFromCharSet(2, "!#%^*"), acctest.RandIntRange(10, 99))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckRdsInstanceDestroy(resourceType),
		Steps: []resource.TestStep{
			{
				Config: testAccReadReplicaInstance_basic(name, dbPwd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsInstanceExists(resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "test_description"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_instance_id",
						"huaweicloud_rds_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "Replica"),
					resource.TestCheckResourceAttrPair(resourceName, "flavor",
						"data.huaweicloud_rds_flavors.replica", "flavors.0.name"),
					resource.TestCheckResourceAttr(resourceName, "type", "Replica"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id",
						"data.huaweicloud_networking_secgroup.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "fixed_ip", "192.168.0.210"),
					resource.TestCheckResourceAttr(resourceName, "ssl_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "db.0.port", "8888"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.type", "CLOUDSSD"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.size", "50"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.limit_size", "400"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.trigger_threshold", "10"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.name", "connect_timeout"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.value", "14"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				Config: testAccReadReplicaInstance_update(name, updateName, dbPwd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsInstanceExists(resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, "name", updateName),
					resource.TestCheckResourceAttr(resourceName, "description", "test_description_update"),
					resource.TestCheckResourceAttrPair(resourceName, "primary_instance_id",
						"huaweicloud_rds_instance.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "flavor",
						"data.huaweicloud_rds_flavors.replica", "flavors.0.name"),
					resource.TestCheckResourceAttr(resourceName, "type", "Replica"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id",
						"data.huaweicloud_networking_secgroup.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "fixed_ip", "192.168.0.220"),
					resource.TestCheckResourceAttr(resourceName, "ssl_enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "db.0.port", "8889"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.type", "CLOUDSSD"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.size", "60"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.limit_size", "500"),
					resource.TestCheckResourceAttr(resourceName, "volume.0.trigger_threshold", "15"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.name", "div_precision_increment"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.value", "12"),
					resource.TestCheckResourceAttr(resourceName, "tags.key_update", "value_update"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo_update", "bar_update"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ssl_enable", "parameters",
				},
			},
		},
	})
}

func TestAccReadReplicaInstance_withEpsId(t *testing.T) {
	var replica instances.RdsInstanceResponse
	name := acceptance.RandomAccResourceName()
	resourceType := "huaweicloud_rds_read_replica_instance"
	resourceName := "huaweicloud_rds_read_replica_instance.test"
	dbPwd := fmt.Sprintf("%s%s%d", acctest.RandString(5),
		acctest.RandStringFromCharSet(2, "!#%^*"), acctest.RandIntRange(10, 99))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
			acceptance.TestAccPreCheckEpsID(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckRdsInstanceDestroy(resourceType),
		Steps: []resource.TestStep{
			{
				Config: testAccReadReplicaInstance_withEpsId(name, dbPwd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRdsInstanceExists(resourceName, &replica),
					resource.TestCheckResourceAttr(resourceName, "enterprise_project_id", acceptance.HW_ENTERPRISE_PROJECT_ID_TEST),
				),
			},
		},
	})
}

func testAccReadReplicaInstance_basic(name, dbPwd string) string {
	return fmt.Sprintf(`
%s

data "huaweicloud_rds_flavors" "replica" {
  db_type       = "MySQL"
  db_version    = "8.0"
  instance_mode = "replica"
  group_type    = "dedicated"
  memory        = 4
  vcpus         = 2
}

resource "huaweicloud_rds_read_replica_instance" "test" {
  name                = "%s"
  description         = "test_description"
  flavor              = data.huaweicloud_rds_flavors.replica.flavors[0].name
  primary_instance_id = huaweicloud_rds_instance.test.id
  availability_zone   = data.huaweicloud_availability_zones.test.names[0]
  security_group_id   = data.huaweicloud_networking_secgroup.test.id
  fixed_ip            = "192.168.0.210"
  ssl_enable          = true

  db {
    port = 8888
  }

  volume {
    type              = "CLOUDSSD"
    size              = 50
    limit_size        = 400
    trigger_threshold = 10
  }

  parameters {
    name  = "connect_timeout"
    value = "14"
  }

  tags = {
    key = "value"
    foo = "bar"
  }
}
`, testAccRdsInstance_mysql_step1(name, dbPwd), name)
}

func testAccReadReplicaInstance_update(name, updateName, dbPwd string) string {
	return fmt.Sprintf(`
%s

data "huaweicloud_rds_flavors" "replica" {
  db_type       = "MySQL"
  db_version    = "8.0"
  instance_mode = "replica"
  group_type    = "dedicated"
  memory        = 8
  vcpus         = 2
}

resource "huaweicloud_rds_read_replica_instance" "test" {
  name                = "%s"
  description         = "test_description_update"
  flavor              = data.huaweicloud_rds_flavors.replica.flavors[0].name
  primary_instance_id = huaweicloud_rds_instance.test.id
  availability_zone   = data.huaweicloud_availability_zones.test.names[0]
  security_group_id   = data.huaweicloud_networking_secgroup.test.id
  fixed_ip            = "192.168.0.220"
  ssl_enable          = false

  db {
    port = 8889
  }

  volume {
    type              = "CLOUDSSD"
    size              = 60
    limit_size        = 500
    trigger_threshold = 15
  }

  parameters {
    name  = "div_precision_increment"
    value = "12"
  }

  tags = {
    key_update = "value_update"
    foo_update = "bar_update"
  }
}
`, testAccRdsInstance_mysql_step1(name, dbPwd), updateName)
}

func testAccReadReplicaInstance_withEpsId(name, dbPwd string) string {
	return fmt.Sprintf(`
%s

data "huaweicloud_rds_flavors" "replica" {
  db_type       = "MySQL"
  db_version    = "8.0"
  instance_mode = "replica"
  group_type    = "dedicated"
  memory        = 4
  vcpus         = 2
}

resource "huaweicloud_rds_read_replica_instance" "test" {
  name                  = "%s"
  flavor                = data.huaweicloud_rds_flavors.replica.flavors[0].name
  primary_instance_id   = huaweicloud_rds_instance.test.id
  availability_zone     = data.huaweicloud_availability_zones.test.names[0]
  enterprise_project_id = "%s"

  volume {
    type              = "CLOUDSSD"
    size              = 40
    limit_size        = 300
    trigger_threshold = 10
  }
}
`, testAccRdsInstance_mysql_step1(name, dbPwd), name, acceptance.HW_ENTERPRISE_PROJECT_ID_TEST)
}
