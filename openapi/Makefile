.PHONY: build-html run-html clean

SRC = ./src
OUTPUT_DIR = ./output
OUTPUT_YAML = $(OUTPUT_DIR)/openapi.yaml
OUTPUT_HTML = $(OUTPUT_DIR)/index.html

build:
	rm -rf $(OUTPUT_DIR)
	mkdir $(OUTPUT_DIR)
	npm install

	node_modules/.bin/swagger-cli bundle $(SRC)/openapi.yaml --outfile $(OUTPUT_YAML) --type yaml
	node_modules/.bin/swagger-cli validate $(OUTPUT_YAML)
	@echo '🎉🎉🎉🎉 done'

run: build
	npx @redocly/cli preview-docs $(OUTPUT_YAML)
	@echo '🎉🎉🎉🎉 done'
