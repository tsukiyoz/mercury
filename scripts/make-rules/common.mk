#
# These variables should not need tweaking.
#

# ==============================================================================
# Includes

# include the common make file
ifeq ($(origin MERCURY_ROOT),undefined)
MERCURY_ROOT :=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
endif

include $(MERCURY_ROOT)/scripts/make-rules/common-versions.mk

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(MERCURY_ROOT)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

# set the version number. you should not need to do this
# for the majority of scenarios.
ifeq ($(origin VERSION), undefined)
# Current version of the project.
  VERSION := $(shell git describe --tags --always --match='v*')
  ifneq (,$(shell git status --porcelain 2>/dev/null))
    VERSION := $(VERSION)-dirty
  endif
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
GOPATH ?= $(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Set a specific PLATFORM
ifeq ($(origin PLATFORM), undefined)
	ifeq ($(origin GOOS), undefined)
		GOOS := $(shell go env GOOS)
	endif
	ifeq ($(origin GOARCH), undefined)
		GOARCH := $(shell go env GOARCH)
	endif
	PLATFORM := $(GOOS)_$(GOARCH)
	# Use linux as the default OS when building images
	IMAGE_PLAT := linux_$(GOARCH)
else
	GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
	GOARCH := $(word 2, $(subst _, ,$(PLATFORM)))
	IMAGE_PLAT := $(PLATFORM)
endif

PRJ_SRC_PATH := github.com/lazywoo/mercury

FIND := find . ! -path './third_party/*' ! -path './vendor/*'
XARGS := xargs --no-run-if-empty

# Helper function to get dependency version from go.mod
get_go_version = $(shell go list -m $1 | awk '{print $$2}')
define go_install
$(info ===========> Installing $(1)@$(2))
$(GO) install $(1)@$(2)
endef

# Helper function to get dependency version from go.mod
get_go_version = $(shell go list -m $1 | awk '{print $$2}')

# Copy githook scripts when execute makefile
COPY_GITHOOK:=$(shell cp -f githooks/* .git/hooks/)

SCRIPTS_DIR=$(MERCURY_ROOT)/scripts

APIROOT ?= $(MERCURY_ROOT)/api
