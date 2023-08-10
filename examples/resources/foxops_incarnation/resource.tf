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