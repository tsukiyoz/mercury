# ==============================================================================
# Includes

include scripts/make-rules/common.mk
include scripts/make-rules/all.mk

.PHONY: format
format: tools.verify.goimports tools.verify.gofumpt
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -w
	@$(FIND) -type f -name '*.go' | $(XARGS) gofumpt -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(PRJ_SRC_PATH)
	@$(GO) mod edit -fmt

.PHONY: install-tools
install-tools:
	@echo "===========> Installing tools"
	@$(MAKE) tools.install

.PHONY: tidy
tidy:
	@$(GO) mod tidy
	
.PHONY: mock
mock:
	@go generate ./...
	@go mod tidy

.PHONY: grpc
grpc:
	@buf generate api/proto