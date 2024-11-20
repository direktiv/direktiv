.PHONY: cluster-create
cluster-create: ## Creates cluster and requires kind
	cat kind-config.yaml | sed 's?MYDIR?'`pwd`'?' | kind create cluster --config -

.PHONY: cluster-delete
cluster-delete: ## Deletes cluster
	kind delete cluster --name direktiv-cluster