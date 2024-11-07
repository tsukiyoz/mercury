# ==============================================================================
# Makefile helper functions for golang
#

GO := go
GO_MINIMUM_VERSION := 1.22

GOPATH := $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

.PHONY: go.build.verify
go.build.verify: ## Verify supported go versions.
ifneq ($(shell $(GO) version|awk -v min=$(GO_MINIMUM_VERSION) '{gsub(/go/,"",$$3);if($$3 >= min){print 0}else{print 1}}'), 0)
	$(error unsupported go version. Please install a go version which is greater than or equal to '$(GO_MINIMUM_VERSION)')
endif

# .PHONY: go.build.%
# go.build.%: ## Build specified applications with platform, os and arch.
# 	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
# 	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
# 	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
# 	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
# 	#@EPOCH_ZERO_GIT_VERSION=$(VERSION) $(SCRIPTS_DIR)/build.sh $(COMMAND) $(PLATFORM)
# 	@if grep -q "func main()" $(EPOCH_ZERO_ROOT)/cmd/$(COMMAND)/*.go &>/dev/null; then \
# 		echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)" ; \
# 		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GO_BUILD_FLAGS) \
# 		-o $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(COMMAND)$(GO_OUT_EXT) $(PRJ_SRC_PATH)/cmd/$(COMMAND) ; \
# 	fi

.PHONY: go.build
go.build: $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS))) ## Build all applications.
