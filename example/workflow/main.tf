terraform {
  required_providers {
    zosmf = {
      source  = "dlut/example"
      # version = "0.1.0"
    }
  }
}

provider "zosmf" {
  zosmf_username = "ibmuser"
  zosmf_password = "asdf88"
  zosmf_url = "https://172.16.31.56/zosmf"
  allow_unverified_ssl = true
}

resource "zosmf_workflow" "workflow" {
  instance_name    = "Terraform - ${var.instance_name} - ${count.index}"
  count = var.instance_count

  workflow_dir = var.workflow_dir
  workflow_file_name = var.workflow_file_name
  system = var.system
} 

output "workflow_result" {
  value = zosmf_workflow.workflow
}
