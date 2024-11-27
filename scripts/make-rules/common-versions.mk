# ==============================================================================
# Versions used by all Makefiles
#

PROTOC_GEN_GO_VERSION ?= v1.32.0
PROTOC_GEN_GO_GRPC_VERSION ?= v1.3.0
BUF_VERSION ?= v1.40.1
GO_IMPORTS_VERSION ?= v0.26.0
GO_FUMPT_VERSION ?= v0.7.0
GO_GIT_LINT_VERSION ?= v1.1.1
GOLANGCI_LINT_VERSION := v1.55.2
WIRE_VERSION ?= $(call get_go_version,github.com/google/wire)
MOCKGEN_VERSION ?= $(call get_go_version,github.com/golang/mock)