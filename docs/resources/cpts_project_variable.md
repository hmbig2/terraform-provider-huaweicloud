---
subcategory: "Cloud Performance Test Service (CPTS)"
---

# huaweicloud_cpts_project_variable

Manages a global variables resource within HuaweiCloud CPTS.

## Example Usage

### Create a integer variable which range 1 to 10

```hcl
resource "huaweicloud_cpts_project" "example" {
  name = "tf_demo_project"
}

resource "huaweicloud_cpts_project_variable" "example" {
  name            = "tf_demo_variable"
  cpts_project_id = huaweicloud_cpts_project.example.id
  type            = 1
  value           = ["1", "10"]
}
```

### Create a Enumerated variable with 3 values

```hcl
resource "huaweicloud_cpts_project" "example" {
  name = "tf_demo_project"
}

resource "huaweicloud_cpts_project_variable" "example" {
  name            = "tf_demo_variable"
  cpts_project_id = huaweicloud_cpts_project.example.id
  type            = 2
  value           = ["enum_1", "enum_2", "enum_3"]
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional, String, ForceNew) Specifies the region in which to create the project resource. If omitted, the
  provider-level region will be used. Changing this parameter will create a new resource.

* `name` - (Required, String) Specifies a name for the project, which can contain a maximum of 42 characters.

* `cpts_project_id` - (Required, Int, ForceNew) Specifies the CPTS project ID which the variable belongs to.
 Changing this parameter will create a new resource.

* `type` - (Optional, Int) Specifies the type of variable. The options are as follows:
  + **1**: Integer.
  + **2**: Enumerated.

* `value` - (Optional, List) Specifies the value range of the variable when the variable type is Integer,
 or the variable value when the variable type is Enumerated.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The resource ID. It is composed of the CPTS project ID which variable belongs and the variable ID,
 separated by a slash.

* `variable_id` -The variable ID.

* `is_quoted` - Whether this value is quoted.

## Import

The variable can be imported by `id`. It is composed of the CPTS project ID which variable belongs and the variable ID,
 separated by a slash. For example,

```
terraform import huaweicloud_cpts_project_variable.example <cpts_project_id>/<variable_id>
```
