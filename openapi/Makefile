.PHONY: build-html run-html clean

SRC = ./src
OUTPUT_DIR = ./output
OUTPUT_YAML = $(OUTPUT_DIR)/openapi.yaml
OUTPUT_HTML = $(OUTPUT_DIR)/index.html

build:
	rm -rf $(OUTPUT_DIR)
	mkdir $(OUTPUT_DIR)
	docker run -it --rm -v `pwd`:`pwd` -w `pwd` node \
	npm install

	-rm $(SRC)/paths/paths.yaml
	cat $(SRC)/paths/* > $(SRC)/paths/paths.yaml

	docker run -it --rm -v `pwd`:`pwd` -w `pwd` node \
	node_modules/.bin/swagger-cli bundle $(SRC)/openapi.yaml --outfile $(OUTPUT_YAML) --type yaml

	docker run -it --rm -v `pwd`:`pwd` -w `pwd` node \
	node_modules/.bin/swagger-cli validate $(OUTPUT_YAML)
	@echo '🎉🎉🎉🎉 done'

run: build
	docker run -it --rm -v `pwd`:`pwd` -w `pwd` -p 80:80 \
	-v `pwd`/output/openapi.yaml:/usr/share/nginx/html/openapi.yaml \
	-e SPEC_URL=openapi.yaml \
	redocly/redoc
