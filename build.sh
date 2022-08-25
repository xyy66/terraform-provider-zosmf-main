#!/bin/bash
VERSION=0.1.0
LOCAL_PROVIDERS="D:\\terraform\\.terraform.d\\plugins_local"
BINARY_PATH="registry.terraform.io/dlut/example/${VERSION}/$(go env GOOS)_$(go env GOARCH)/terraform-provider-example_v${VERSION}"

echo "building terraform-provider-example_v${VERSION}"
go build -o ${LOCAL_PROVIDERS}/${BINARY_PATH}
