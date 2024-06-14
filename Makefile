#
# Makefile to build direktiv
#
# Help information is automatically scraped from comments beginning with 
# double-hash. Please add these lines to commands that deserve documentation.

DOCKER_REPO := localhost:5000

# gets the git hash of the actual commit
GIT_HASH := $(shell git rev-parse --short HEAD)

# adds '-dirty'
GIT_DIRTY := $(shell git diff --quiet || echo '-dirty')

# name of the release, e.g. v0.8.0
RELEASE := $(if $(RELEASE),$(RELEASE),latest)

# full version including version and git hash
RELEASE_VERSION := ${RELEASE}-${GIT_HASH}${GIT_DIRTY}

.DEFAULT_GOAL := direktiv

include make_direktiv.mk make_direktiv_ui.mk make_k3s.mk make_tests.mk make_composer.mk

.PHONY: help
help: ## Prints usage information.
	@echo "\033[36mMakefile Help\033[0m"
	@echo ""
	@echo "\033[36mTargets\033[0m"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":"}; {printf "  %-24s %s\n", $$2, $$3}'
