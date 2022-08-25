variable "instance_count" {
  type        = number
  default     = 1
  description = "Number of zosmf workflow will be created"
}

variable "instance_name" {
  type        = string
  default     = "workflow-via-terraform"
  description = "Part of the workflow instance name that use to identify which workflow is provisioned via user configuration"
}

variable "workflow_dir" {
  type        = string
  default     = "/oc4z/images/ocp_live/workflows/"
  description = "Directory that holds the workflow definition file"
}

variable "workflow_file_name" {
  type        = string
  default     = "ocp_provision.xml"
  description = "Workflow definition file name and extension"
}

variable "system" {
  type        = string
  default     = "NP8"
  description = "Which system that the workflow will running on"
}
