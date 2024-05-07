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
docker-e2e-playwright:
docker-e2e-playwright: ## Create a local docker deployment.
	docker run \
	-v $$PWD/ui:/app/ui \
	-e NODE_TLS_REJECT_UNAUTHORIZED=0 \
    -e PLAYWRIGHT_USE_VITE=FALSE \
    -e PLAYWRIGHT_UI_BASE_URL=http://127.0.0.1 \
    -e PLAYWRITE_SHARD=1/1 \
	-w /app/ui \
	--net=host \
	node:18 \
	bash -c "yarn && npx playwright install --with-deps chromium && npx playwright test --shard=${PLAYWRITE_SHARD} --project \"chromium\"  --reporter=line"