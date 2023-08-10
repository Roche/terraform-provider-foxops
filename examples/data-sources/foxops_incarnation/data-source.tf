data "foxops_incarnation" "example" {
  id = "1234"

  wait_for_mr_status = {
    status  = "merged"
    timeout = "5m"
  }
}