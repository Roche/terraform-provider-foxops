---
title: "foxops_incarnation"
subcategory: ""
description: |-
  Use this resource to create and manage incarnations.
---

Use this resource to create and manage incarnations.

## Example Usage
```terraform
resource "foxops_incarnation" "example" {
  incarnation_repository      = "my-org/my-repository"
  target_directory            = "./some-folder"
  template_repository         = "https://github.com/my-org/my-repository"
  template_repository_version = "v1.2.3"
  auto_merge_on_update        = true

  template_data = {
    hello = "World!"
  }

  wait_for_mr_status_on_update = {
    status  = "merged"
    timeout = "5m"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `incarnation_repository` (String) The repository in which the incarnation will be created.
- `template_repository` (String) The repository containing the template used to create the incarnation.
- `template_repository_version` (String) A tag, commit or branch of the template repository to use for the incarnation.

### Optional

- `auto_merge_on_update` (Boolean) Whether merge request should automatically merged after update of the incarnation.
- `target_directory` (String) The folder in which the incarnation will be created. Default: `.`.
- `template_data` (Map of String) An object containing variables used to generate the incarnation. These variables should match those declared in the `fengine.yaml` file of the template
- `wait_for_mr_status_on_update` (Attributes) Wait for the status of the last merge request to reach a status before completing the current operation. This field only affects incarnation that have been updated as it requires a merge request to exist. (see [below for nested schema](#nestedatt--wait_for_mr_status_on_update))

### Read-Only

- `commit_sha` (String) The hash of the last commit created for the incarnation.
- `commit_url` (String) The url of the last commit created for the incarnation.
- `id` (String) The `id` of the incarnation.
- `merge_request_id` (String) The id of the last merge request created for the incarnation. This property will be `null` after the creation of the incarnation and only populated after updates.
- `merge_request_status` (String) The status of the last merge request created for the incarnation. This property will be `null` after the creation of the incarnation and only populated after updates. It will be one of `open`, `merged`, `closed` or `unknown`.
- `merge_request_url` (String) The url of the latest merge request created for the incarnation. This property will be `null` after the creation of the incarnation and only populated after updates.

<a id="nestedatt--wait_for_mr_status_on_update"></a>
### Nested Schema for `wait_for_mr_status_on_update`

Required:

- `status` (String) The expected status for the merge request. Can be one of `open`, `merge`, `closed` or `unknown`.

Optional:

- `timeout` (String) The amount of time to wait for the expected status to be reached. It should be a sequence of numbers followed by a unit suffix (`s`, `m` or `h`). Example: `1m30s`. Default: `10s`.

## Import
Import is supported using the following syntax:
```shell
terraform import foxops_incarnation.example "<edge_cluster_id>"
```
