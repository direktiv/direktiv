.PHONY: k3s-wait 
k3s-wait: 
	kubectl -n direktiv wait --for=condition=ready pod -l "app=direktiv-flow"

.PHONY: k3s-uninstall
k3s-uninstall: ## Uninstall the local development k3s environment.
	./scripts/installer.sh uninstall

.PHONY: k3s-install
k3s-install: k3s-uninstall
k3s-install: ## Install the local development k3s environment.
	DEV=true ./scripts/installer.sh all
	@$(MAKE) k3s-wait

.PHONY: k3s-monitoring-install
k3s-monitoring-install: k3s-uninstall
k3s-monitoring-install: ## Install the local development k3s environment.
	WITH_MONITORING=true DEV=true ./scripts/installer.sh all
	@$(MAKE) k3s-wait

.PHONY: k3s-redeploy
k3s-redeploy: ## Upgrade the local deployment.
	DEV=true ./scripts/installer.sh all
	@$(MAKE) k3s-wait

.PHONY: k3s-reboot 
k3s-reboot: direktiv 
k3s-reboot: ## Recompile the server image and delete the existing pod in the k3s deployment to force an update.
	kubectl -n direktiv delete pod -l app.kubernetes.io/name=direktiv,app.kubernetes.io/instance=direktiv 
	@$(MAKE) k3s-wait

.PHONY: k3s-tail 
k3s-tail: k3s-wait
k3s-tail: ## Tail the logs of the direktiv server running in the local k3s environment.
	kubectl -n direktiv logs -f -l "app=direktiv-flow"