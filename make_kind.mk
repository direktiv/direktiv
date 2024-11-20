CLUSTER_NAME_FOR_E2E := direktiv-e2e
CIDR_FOR_E2E := 172.22.2.0/28
CIDR_FOR_DEV := 172.22.1.0/28

KIND_SCRIPTS_DIR := ./scripts/kind
KIND_INSTALL_DEPENDENCIES := $(KIND_SCRIPTS_DIR)/install-dependencies.sh
KIND_INSTALL_DIREKTIV := $(KIND_SCRIPTS_DIR)/install-direktiv.sh
KIND_API_TESTS_DIREKTIV := $(KIND_SCRIPTS_DIR)/run-api-test.sh

# Default cluster name
DEFAULT_CLUSTER_NAME := direktiv

# Phony targets to prevent Make from using files with these names
.PHONY: setup-kind setup-direktiv e2e

setup-host:
	sudo sysctl fs.inotify.max_user_watches=524288
	sudo sysctl fs.inotify.max_user_instances=512
	docker network inspect kind | jq -r '.[0].IPAM.Config[0].Subnet'

# Target for setting up the Kind cluster
setup-kind:
	@echo "Setting up Kind cluster with name $(DEFAULT_CLUSTER_NAME)..."
	bash $(KIND_INSTALL_DEPENDENCIES) $(DEFAULT_CLUSTER_NAME) $(CIDR_FOR_DEV)

# Target for installing Direktiv
setup-direktiv:
	@echo "Installing Direktiv in the $(DEFAULT_CLUSTER_NAME) cluster..."
	bash $(KIND_INSTALL_DIREKTIV) $(DEFAULT_CLUSTER_NAME)

# Target for running E2E tests
e2e:
	@echo "Setting up Kind cluster with random name $(CLUSTER_NAME_FOR_E2E) for E2E tests..."
	bash $(KIND_INSTALL_DEPENDENCIES) $(CLUSTER_NAME_FOR_E2E) $(CIDR_FOR_E2E)
	@echo "Installing Direktiv in the $(CLUSTER_NAME_FOR_E2E) cluster..."
	bash $(KIND_INSTALL_DIREKTIV) $(CLUSTER_NAME_FOR_E2E)
	@echo "Launching E2E tests on the $(CLUSTER_NAME_FOR_E2E) cluster..."
	bash $(KIND_API_TESTS_DIREKTIV) $(CLUSTER_NAME_FOR_E2E)
	@echo "Cleaning up the Kind cluster $(CLUSTER_NAME_FOR_E2E)..."
	kind delete cluster --name $(CLUSTER_NAME_FOR_E2E)

# Clean default cluster if needed
clean-cluster:
	@echo "Cleaning up Kind cluster $(DEFAULT_CLUSTER_NAME)..."
	kind delete cluster --name $(DEFAULT_CLUSTER_NAME)

build-load-image:
	@echo "Building the Docker image and tagging it with 'dev'..."
	# Build the Docker image
	docker build -t direktiv/direktiv:dev .
	# Load the image into the Kind cluster
	kind load docker-image direktiv/direktiv:dev --name $(DEFAULT_CLUSTER_NAME)
	@echo "Docker image 'direktiv/direktiv:dev' has been built and loaded into Kind cluster $(DEFAULT_CLUSTER_NAME)."

kill-flow:
	@echo "Killing all flow pods using the 'direktiv/direktiv:dev' image..."
	kubectl delete pod -l app.kubernetes.io/instance=direktiv,app.kubernetes.io/name=direktiv --namespace default
	@echo "All pods using the 'direktiv/direktiv:dev' image have been killed."

kill-sidecar:
	@echo "Killing all pods using the 'direktiv/direktiv:dev' image in sidecar..."
	kubectl delete pod -l direktiv-app=direktiv --namespace direktiv-services-direktiv
	@echo "All pods using the 'direktiv/direktiv:dev' image have been killed."

delete-all-clusters:
	@echo "Deleting all Kind clusters..."
	# Get all Kind clusters and delete them one by one
	for cluster in $(shell kind get clusters); do \
		echo "Deleting cluster $$cluster..."; \
		kind delete cluster --name $$cluster; \
	done
	@echo "All Kind clusters have been deleted."

print_postgress:
	kubectl get secret --namespace postgres postgres-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d