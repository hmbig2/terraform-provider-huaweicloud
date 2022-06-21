---
subcategory: "IoT Device Access (IoTDA)"
---

# huaweicloud_iotda_access_credential

Manages an IoTDA access credential within HuaweiCloud.

The access credential is used by the client to connect with the platform using protocols such as AMQP.
Only one record is kept in IoT platform. If the resource is created repeatedly, the access credential will be reset
and the previous one will be invalid.

## Example Usage

```hcl
resource "huaweicloud_iotda_access_credential" "space" {
  type = "first_space"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, String, ForceNew) Specifies the region in which to create the IoTDA resource space resource.
If omitted, the provider-level region will be used. Changing this parameter will create a new resource.

* `type` - (Optional, String, ForceNew) Specifies access credential type. The valid values are **AMQP**.
Defaults to `AMQP`. Changing this parameter will create a new resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID in UUID format.

* `access_key` - The access key.

* `access_code` - The access code.

## Import

Access credentials can be imported by `type`, `access_key` and `access_code`, separated by a slash, e.g.

```
$ terraform import huaweicloud_iotda_access_credential.test AMQP/12345678/lfkPqUUl6yWjRhmD5UKM3seJP9ZUXaOn
```
