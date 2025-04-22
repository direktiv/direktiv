.PHONY: tests-scan
tests-scan: direktiv
	trivy image --exit-code 1 --ignore-unfixed localhost:5000/direktiv

.PHONY: tests-scan-ui
tests-scan-ui: direktiv-ui
	trivy image --exit-code 1 --ignore-unfixed localhost:5000/frontend


# DIREKTIV_HOST := $(shell kubectl -n direktiv get services direktiv-ingress-nginx-controller --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
.PHONY: tests-api
tests-api: ## Runs end-to-end tests. DIREKTIV_HOST=128.0.0.1 make tests-api [JEST_PREFIX=/tests/namespaces]	
	kubectl wait --for=condition=ready pod -l "app=direktiv-flow"
	docker run -it --rm \
	-v `pwd`/tests:/tests \
	-e 'DIREKTIV_HOST=http://${DIREKTIV_HOST}' \
	-e 'NODE_TLS_REJECT_UNAUTHORIZED=0' \
	--network=host \
	node:lts-alpine3.18 npm --prefix "/tests" run jest -- ${JEST_PREFIX} --runInBand

# DIREKTIV_HOST := $(shell kubectl -n direktiv get services direktiv-ingress-nginx-controller --output jsonpath='{.status.loadBalancer.ingress[0].ip}')
.PHONY: tests-ee-api
tests-ee-api: ## Runs end-to-end tests. DIREKTIV_HOST=128.0.0.1 make tests-api [JEST_PREFIX=/tests/namespaces]
	kubectl wait --for=condition=ready pod -l "app=direktiv-flow"
	docker run -it --rm \
	-v `pwd`/direktiv-ee/tests:/tests \
	-e 'DIREKTIV_HOST=http://${DIREKTIV_HOST}' \
	-e 'NODE_TLS_REJECT_UNAUTHORIZED=0' \
	--network=host \
	node:lts-alpine3.18 npm --prefix "/tests" run jest -- ${JEST_PREFIX} --runInBand


TEST_PACKAGES := $(shell find . -type f -name '*_test.go' | sed -e 's/^\.\///g' | sed -r 's|/[^/]+$$||'  |sort |uniq)
UNITTEST_PACKAGES := $(shell echo ${TEST_PACKAGES} | sed 's/ /\n/g' | awk '{print "github.com/direktiv/direktiv/" $$0}')

.PHONY: tests-unittest
tests-unittest: ## Runs all Go unit tests. Or, you can run a specific set of unit tests by defining UNITTEST_PACKAGES relative to the root directory.	
	go test -cover -timeout 60s ${UNITTEST_PACKAGES}

.PHONY: tests-license-check 
tests-license-check: ## Scans dependencies looking for blacklisted licenses.
	go install github.com/google/go-licenses@latest
	go-licenses check --ignore=github.com/bbuck/go-lexer,github.com/xi2/xz,modernc.org/mathutil ./... --disallowed_types forbidden,unknown,restricted

.PHONY: tests-godoc
tests-godoc: ## Hosts a godoc server for the project on http port 6060.
	go install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060

.PHONY: tests-lint 
tests-lint: VERSION="v2.1"
tests-lint: ## Runs very strict linting on the project.
	docker run \
	--rm \
	--tty \
	-v `pwd`:/app \
	-w /app \
	-v /tmp/golangci:/root/.cache:rw \
	-e GOCACHE=/root/.cache/  \
	-e GOLANGCI_LINT_CACHE=/root/.cache/ \
	-e GOMODCACHE=/root/.cache/go/pkg/mod  \
	golangci/golangci-lint:${VERSION} golangci-lint run --verbose --config /app/.golangci.yaml
#  -e GOCACHE=/root/.cache/  -e GOLANGCI_LINT_CACHE=/root/.cache/ -e GOMODCACHE=/root/.cache/go/pkg/mod 
.PHONY: docker-playwright
docker-e2e-playwright:
	docker run \
	-v $$PWD/ui:/app/ui \
	-e NODE_TLS_REJECT_UNAUTHORIZED=0 \
    -e PLAYWRIGHT_USE_VITE=FALSE \
    -e PLAYWRIGHT_UI_BASE_URL=http://127.0.0.1 \
    -e PLAYWRIGHT_SHARD=1/1 \
    -e PLAYWRIGHT_CI=TRUE \
	-w /app/ui \
	--net=host \
	node:20-slim \
	bash -c "\
		corepack enable && \
		corepack prepare pnpm@9.15.4 --activate && \
		pnpm install && \
		pnpm exec playwright install --with-deps chromium && \
		pnpm exec playwright test --shard=$${PLAYWRIGHT_SHARD} --project \"chromium\" --reporter=line"
