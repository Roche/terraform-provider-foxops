terraform {
  required_providers {
    foxops = {
      source = "Roche/foxops"
    }
  }
}

provider "foxops" {
  endpoint = var.foxops_endpoint
  token    = var.foxops_token
}
