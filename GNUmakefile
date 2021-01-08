default: testacc
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
INSTALL_PATH=~/.local/share/terraform/plugins/localhost/providers/hcp/0.0.1/linux_$(GOARCH)
BUILD_ALL_PATH=${PWD}/bin

ifeq ($(GOOS), darwin)
	INSTALL_PATH=~/Library/Application\ Support/io.terraform/plugins/localhost/providers/hcp/0.0.1/darwin_$(GOARCH)
endif
ifeq ($(GOOS), "windows")
	INSTALL_PATH=%APPDATA%/HashiCorp/Terraform/plugins/localhost/providers/hcp/0.0.1/windows_$(GOARCH)
endif

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

dev:
	mkdir -p $(INSTALL_PATH)
	go build -o $(INSTALL_PATH)/terraform-provider-hcp main.go

all:
	mkdir -p $(BUILD_ALL_PATH)
	GOOS=darwin go build -o $(BUILD_ALL_PATH)/terraform-provider-hcp_darwin-amd64 main.go
	GOOS=windows go build -o $(BUILD_ALL_PATH)/terraform-provider-hcp_windows-amd64 main.go
	GOOS=linux go build -o $(BUILD_ALL_PATH)/terraform-provider-hcp_linux-amd64 main.go
