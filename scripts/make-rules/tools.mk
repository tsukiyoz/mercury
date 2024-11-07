# ==============================================================================
#  Makefile helper functions for tools
#
# Rules name starting with `_` mean that it is not recommended to call directly through make command, 
# like `make _install.gotests`, you should run `make tools.install.gotests` instead.

CI_WORKFLOW_TOOLS := golangci-lint goimports gofumpt
OTHER_TOOLS := go-gitlint

.PHONY: tools.install
tools.install: _install.other

.PHONY: _tools.install.%
_tools.install.%: ## Install a specified tool.
	@echo "===========> Installing $*"
	@$(MAKE) _install.$*

.PHONY: tools.verify.%
tools.verify.%: ## Verify a specified tool.
	@if ! which $* > /dev/null; then $(MAKE) tools.install.$*; fi

.PHONY: _install.other
_install.other: $(addprefix _tools.install., $(OTHER_TOOLS))

.PHONY: _install.grpc
_install.grpc: ## Install grpc.
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	
.PHONY: _install.goimports
_install.goimports: ## Install goimports.
	@$(GO) install golang.org/x/tools/cmd/goimports@$(GO_IMPORTS_VERSION)
	
.PHONY: _install.go-gitlint
_install.go-gitlint: ## Install go-gitlint.
	@$(GO) install github.com/marmotedu/go-gitlint/cmd/go-gitlint@$(GO_GIT_LINT_VERSION)
	
.PHONY: _install.gofumpt
_install.gofumpt: ## Install gofumpt.
	@$(GO) install mvdan.cc/gofumpt@$(GO_FUMPT_VERSION)

.PHONY: _install.buf
_install.buf: ## Install buf command line tool.
	@$(GO) install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

