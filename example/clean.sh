#!/bin/bash
terraform destroy --auto-approve
rm -rf .terraform.lock.hcl
rm -rf ./plugins/*
rm terraform.tfstate
rm terraform.tfstate.backup
