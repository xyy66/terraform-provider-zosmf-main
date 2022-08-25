# Terraform Provider zOSMF

## Test sample configuration

`cd example` folder holds ready to use example you can do a quick test

**Note** You must fill in correct zOSMF credential in `/example/terraform.tfvars` before modify resource

First, run the build script if you want to test any source code changes, the script will automatically init the plugin and plan the dry run of your configuration.

```shell
./prep.sh
```

Then, run the following command to apply resource changes by sample configuration.

```shell
terraform apply
```
"# crm"
