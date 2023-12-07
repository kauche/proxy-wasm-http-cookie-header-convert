OS   := $(shell uname | awk '{print tolower($$0)}')
ARCH := $(shell case $$(uname -m) in (x86_64) echo amd64 ;; (aarch64) echo arm64 ;; (*) echo $$(uname -m) ;; esac)

BIN_DIR := ./.bin

TINYGO_VERSION := 0.30.0
TINYGO := $(abspath $(BIN_DIR)/tinygo-$(TINYGO_VERSION))/bin/tinygo

DOCKER_NETWORK := proxy-wasm-http-cookie-header-convert_default

tinygo: $(TINYGO)
$(TINYGO):
	@curl -sSL "https://github.com/tinygo-org/tinygo/releases/download/v$(TINYGO_VERSION)/tinygo$(TINYGO_VERSION).$(OS)-$(ARCH).tar.gz" | tar -C $(BIN_DIR) -xzv tinygo
	@mv $(BIN_DIR)/tinygo $(BIN_DIR)/tinygo-$(TINYGO_VERSION)

.PHONY: test
test:
	@cd ./test && go test -race -shuffle=on .

.PHONY: test-docker
test-docker:
	docker compose stop
	docker compose up --detach

	docker run --rm --network $(DOCKER_NETWORK) jwilder/dockerize:0.6.1 -wait tcp://envoy:8080 -timeout 30s
	docker run --rm --network $(DOCKER_NETWORK) jwilder/dockerize:0.6.1 -wait tcp://upstream:5000 -timeout 30s

	docker run \
		--rm \
		--env ENVOY_ADDRESS=envoy:8080 \
		--volume "$(shell pwd):/workspace" \
		--workdir /workspace \
		--network $(DOCKER_NETWORK) \
		golang:1.20.7-bookworm make test

.PHONY: build
build: $(TINYGO)
	@$(TINYGO) build -o $(BIN_DIR)/proxy-wasm-http-cookie-header-convert.wasm -scheduler=none -target=wasi .

.PHONY: build-docker
build-docker:
	@docker run \
		--rm \
		--env XDG_CACHE_HOME=/tmp/.cache \
		--volume "$(shell pwd):/workspace" \
		--user "$(shell id -u):$(shell id -g)" \
		--workdir /workspace \
		golang:1.21.5-bookworm \
		make build
