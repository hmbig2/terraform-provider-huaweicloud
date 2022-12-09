---
subcategory: "DataArts Studio"
---

# huaweicloud_dataarts_data_connection

Manages a data connection resource of DataArts Studio within HuaweiCloud.  

## Example Usage

```HCL
variable "data_connection_name" {}
variable "workspace" {}

resource "huaweicloud_dataarts_data_connection" "test" {
  name      = var.data_connection_name
  type      = "DLI"
  workspace = var.workspace
  config    = {}
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, String, ForceNew) Specifies the region in which to create the resource.
  If omitted, the provider-level region will be used. Changing this parameter will create a new resource.

* `name` - (Required, String) The data connect name.  
  The name can contain 1 to 128 characters, only letters, digits, hyphens (-) and underscores (_).

* `type` - (Required, String) The type of data connect.  
  The options are as follows:
    + **DWS**: DWS connection.
    + **DLI**: DLI connection.
    + **ORACLE**: Oracle connection.
    + **RDS**: RDS connection.
    + **DIS**: DIS connection.
    + **SHELL**: Host Connection.
    + **MRS Hive**: MRS Hive connection.
    + **MRS HBase**: MRS HBase connection.
    + **MRS Kafka**: MRS Kafka connection.
    + **MRS Ranger**: MRS Ranger connection.
    + **MRS Spark**: MRS Spark connection.
    + **MRS Presto**: MRS Presto connection.

* `config` - (Required, Map) The config of the data connect.  
  For more detail, please see[Creating Data Connections](https://support.huaweicloud.com/intl/en-us/usermanual-dataartsstudio/dataartsstudio_01_0009.html).

* `workspace` - (Required, String, ForceNew) The worksapce to which the data connect belongs.
   Changing this parameter will create a new resource.

* `agent_id` - (Optional, String) The agent ID which used by the data connect.  

* `agent_name` - (Optional, String) The agent name which used by the data connect.  

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID.

## Import

The data connection can be imported using
`workspace`, `id`, separated by slashes, e.g.

```
$ terraform import huaweicloud_dataarts_data_connection.test <workspace>/<id>
```
