# ==============================================================================
# Makefile helper functions for generate necessary files
#

SERVICES ?= $(filter-out tools,$(foreach service,$(wildcard ${MERCURY_ROOT}/cmd/*),$(notdir ${service})))