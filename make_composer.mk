.PHONY: docker-stop
docker-stop: ## Stop an existing docker deployment.
	docker rm -f $$(docker ps -q -f "label=direktiv.io/object-type=container") || true
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev docker compose down --remove-orphans -v

.PHONY: docker-build-api
docker-build-api:
	docker build -t direktiv-dev .

.PHONY: docker-build-ui
docker-build-ui:
	cd ui && docker build -t direktiv-ui-dev .

.PHONY: docker-start
docker-start: docker-build-api docker-build-ui docker-stop
docker-start: ## Create a local docker deployment.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose up -d --scale e2e-api=0

.PHONY: docker-headless
docker-headless: docker-build-api docker-stop
docker-headless: ## Create a local docker deployment without an included UI container.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose up -d --scale ui=0 --scale e2e-api=0

.PHONY: docker-tail
docker-tail: ## Tail the logs for the direktiv container in the docker deployment.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose logs -f

.PHONY: docker-e2e-api
docker-e2e-api: docker-stop docker-build-api
docker-e2e-api: ## Perform backend end-to-end tests against the docker deployment.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose run e2e-api

.PHONY: docker-e2e-playwright
docker-e2e-playwright: docker-build-api docker-stop
docker-e2e-playwright: ## Create a local docker deployment.
	cd ui && docker build -t direktiv-ui-dev .
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose run e2e-playwright
