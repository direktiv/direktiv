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

include make_protobuf.mk make_direktiv.mk make_direktiv_ui.mk make_k3s.mk make_tests.mk make_composer.mk


