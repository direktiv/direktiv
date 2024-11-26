.PHONY: docker-start
docker-start:
docker-start: ## Create a local docker deployment.
	docker compose down --remove-orphans -v
	docker build -t direktiv-dev .
	docker compose up -d --scale e2e-api=0

.PHONY: docker-e2e
docker-e2e-api:
docker-e2e-api: ## Perform backend end-to-end tests against the docker deployment.
	docker compose down --remove-orphans -v
	docker build -t direktiv-dev .
	docker compose run e2e-api

.PHONY: docker-playwright
docker-e2e-playwright:
docker-e2e-playwright: ## Perform ui e2e tests with playwright platform.
	docker run \
	-v $$PWD/ui:/app/ui \
	-e NODE_TLS_REJECT_UNAUTHORIZED=0 \
    -e PLAYWRIGHT_USE_VITE=FALSE \
    -e PLAYWRIGHT_UI_BASE_URL=http://127.0.0.1 \
    -e PLAYWRIGHT_SHARD=1/1 \
    -e PLAYWRIGHT_CI=TRUE \
	-w /app/ui \
	--net=host \
	node:18 \
	bash -c "yarn && npx playwright install --with-deps chromium && npx playwright test --shard=${PLAYWRIGHT_SHARD} --project \"chromium\"  --reporter=line"