# Build all by default, even if it's not first
.DEFAULT_GOAL := help

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
ifeq ($(ALL),1)
	$(MAKE) format.protobuf
endif

.PHONY: format.protobuf
format.protobuf: tools.verify.buf ## Lint protobuf files.
	@echo "===========> Formating protobuf files"
	@for f in $(shell find $(APIROOT) -name *.proto) ; do                  \
	  buf format -w $$f ;                                                  \
	done

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
	@buf generate $(APIROOT)
	
.PHONY: help
help:
	@echo $(SERVICES)