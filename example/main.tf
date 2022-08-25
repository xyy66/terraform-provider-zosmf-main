terraform {
  required_providers {
    zosmf = {
      source  = "dlut/example"
      version = "0.1.0"
    }
  }
}

provider "zosmf" {
  zosmf_username = "ibmuser"
  zosmf_password = "asdf88"
  zosmf_url = "https://172.16.31.56/zosmf"
  allow_unverified_ssl = true
}
 

resource "zosmf_resource_liberty" "resourceLiberty" {
  count = var.liberty_count
  template_name = var.template_name
}

output "running_liberty" {
  value = zosmf_resource_liberty.resourceLiberty
  description = "The URL to visit the liberty server"
}


