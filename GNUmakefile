GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
INSTALL_PATH=~/.local/share/terraform/plugins/localhost/providers/hcp/0.0.1/linux_$(GOARCH)
BUILD_ALL_PATH=${PWD}/bin
TEST?=./internal/...
GO_LINT ?= golangci-lint
GO_LINT_CONFIG_PATH ?= ./golangci-config.yml
TIMEOUT?=360m

ifeq ($(GOOS), darwin)
	INSTALL_PATH=~/Library/Application\ Support/io.terraform/plugins/localhost/providers/hcp/0.0.1/darwin_$(GOARCH)
endif
ifeq ($(GOOS), "windows")
	INSTALL_PATH=%APPDATA%/HashiCorp/Terraform/plugins/localhost/providers/hcp/0.0.1/windows_$(GOARCH)
endif

default: dev

dev:
	mkdir -p $(INSTALL_PATH)
	go build -o $(INSTALL_PATH)/terraform-provider-hcp main.go

all:
	mkdir -p $(BUILD_ALL_PATH)
	GOOS=darwin go build -o $(BUILD_ALL_PATH)/terraform-provider-hcp_darwin-amd64 main.go
	GOOS=windows go build -o $(BUILD_ALL_PATH)/terraform-provider-hcp_windows-amd64 main.go
	GOOS=linux go build -o $(BUILD_ALL_PATH)/terraform-provider-hcp_linux-amd64 main.go

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./internal

fmtcheck:
	@./scripts/gofmtcheck.sh
	$(GO_LINT) run --config $(GO_LINT_CONFIG_PATH) $(GO_LINT_ARGS)

test: fmtcheck
	go test $(TEST) $(TESTARGS) -timeout=5m -parallel=4

test-ci: fmtcheck
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

testacc: fmtcheck
	@if [ "$(TESTARGS)" = "-run=TestAccXXX" ]; then \
		echo ""; \
		echo "Error: Skipping example acceptance testing pattern. Update TESTARGS to match the test naming in the relevant *_test.go file."; \
		echo ""; \
		echo "For example if updating resource_hvn.go, use the test names in resource_hvn_test.go starting with TestAcc:"; \
		echo "make testacc TESTARGS='-run=TestAccHvn'"; \
		echo ""; \
		echo "See the contributing guide for more information: https://github.com/hashicorp/terraform-provider-hcp/blob/main/contributing/writing-tests.md"; \
		exit 1; \
	fi
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout $(TIMEOUT) -parallel=10

testacc-ci: fmtcheck
	@if [ "$(TESTARGS)" = "-run=TestAccXXX" ]; then \
		echo ""; \
		echo "Error: Skipping example acceptance testing pattern. Update TESTARGS to match the test naming in the relevant *_test.go file."; \
		echo ""; \
		echo "For example if updating resource_hvn.go, use the test names in resource_hvn_test.go starting with TestAcc:"; \
		echo "make testacc TESTARGS='-run=TestAccHvn'"; \
		echo ""; \
		echo "See the contributing guide for more information: https://github.com/hashicorp/terraform-provider-hcp/blob/main/contributing/writing-tests.md"; \
		exit 1; \
	fi
	TF_ACC=1 go test -short -coverprofile=coverage-e2e.out $(TEST) -v $(TESTARGS) -timeout $(TIMEOUT) -parallel=10
	go tool cover -html=coverage-e2e.out -o coverage-e2e.html

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)

gencheck:
	@echo "==> Checking generated source code..."
	go generate
	@git diff --compact-summary --exit-code || \
		(echo; echo "Unexpected difference in directories after code generation. Run 'go generate' command and commit."; exit 1)

.PHONY: dev all fmt fmtcheck test test-ci testacc depscheck gencheck
