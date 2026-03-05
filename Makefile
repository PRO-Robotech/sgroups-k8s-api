SHELL := /bin/bash

# Project metadata
MODULE   := sgroups.io/sgroups-k8s-api
BIN_DIR  := bin

# Tools
GOLANGCI_LINT := golangci-lint

# Build flags
GOFLAGS  ?=
LDFLAGS  ?=

# All cmd/ subdirectories become build targets
CMDS := $(notdir $(wildcard cmd/*))

.DEFAULT_GOAL := help

##@ Build

.PHONY: build
build: $(CMDS) ## Build all binaries into bin/

.PHONY: $(CMDS)
$(CMDS):
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$@ ./cmd/$@

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)

##@ Quality

.PHONY: lint
lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	$(GOLANGCI_LINT) run --fix

.PHONY: fmt
fmt: ## Format code (gofmt + goimports)
	$(GOLANGCI_LINT) fmt

.PHONY: test
test: ## Run tests with race detector
	go test -race ./...

.PHONY: verify
verify: lint fmt test verify-openapi-spec ## Run all checks: lint + fmt + test + swagger

##@ OpenAPI

.PHONY: openapi-spec
openapi-spec: ## Generate api/openapi-spec/swagger.json
	go run ./hack/openapi-spec > api/openapi-spec/swagger.json

.PHONY: verify-openapi-spec
verify-openapi-spec: ## Verify swagger.json is up to date
	@diff <(go run ./hack/openapi-spec) api/openapi-spec/swagger.json || \
	  (echo "api/openapi-spec/swagger.json is stale — run 'make openapi-spec'" && exit 1)

##@ Code Generation

.PHONY: generate
generate: ## Run all code generators (hack/update-all.sh)
	./hack/update-all.sh

##@ Docker

IMAGE_APISERVER ?= sgroups-k8s-apiserver:latest
IMAGE_MOCK      ?= sgroups-mock-backend:latest
KIND_CLUSTER    ?= sgroups-test
CERT_MANAGER_VERSION ?= v1.17.2

.PHONY: docker-build-apiserver
docker-build-apiserver: ## Build apiserver Docker image (run from parent dir)
	docker build -f apiserver.Dockerfile -t $(IMAGE_APISERVER) ..

.PHONY: docker-build-mock
docker-build-mock: ## Build mock backend Docker image (run from parent dir)
	docker build -f mock.Dockerfile -t $(IMAGE_MOCK) ..

.PHONY: docker-build
docker-build: docker-build-apiserver docker-build-mock ## Build all Docker images

##@ Setup

.PHONY: kind-create
kind-create: ## Create a kind cluster
	@if kind get clusters 2>/dev/null | grep -qx '$(KIND_CLUSTER)'; then \
		echo "Kind cluster '$(KIND_CLUSTER)' already exists"; \
	else \
		kind create cluster --name $(KIND_CLUSTER) --wait 60s; \
	fi

.PHONY: cert-manager
cert-manager: ## Install cert-manager into the cluster
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/$(CERT_MANAGER_VERSION)/cert-manager.yaml
	kubectl -n cert-manager wait --for=condition=Available deployment/cert-manager --timeout=120s
	kubectl -n cert-manager wait --for=condition=Available deployment/cert-manager-webhook --timeout=120s
	kubectl -n cert-manager wait --for=condition=Available deployment/cert-manager-cainjector --timeout=120s

.PHONY: setup
setup: kind-create cert-manager deploy ## Full setup: kind + cert-manager + deploy

.PHONY: teardown
teardown: undeploy ## Remove resources and delete kind cluster
	kind delete cluster --name $(KIND_CLUSTER)

##@ Deploy (kind)

.PHONY: kind-load
kind-load: ## Load Docker images into kind cluster
	kind load docker-image $(IMAGE_APISERVER) --name $(KIND_CLUSTER)
	kind load docker-image $(IMAGE_MOCK) --name $(KIND_CLUSTER)

.PHONY: deploy
deploy: docker-build kind-load ## Build, load, apply manifests, and restart pods
	kubectl apply -k config/
	kubectl -n sgroups-system rollout restart deployment/sgroups-k8s-apiserver deployment/sgroups-backend
	kubectl -n sgroups-system rollout status deployment/sgroups-k8s-apiserver --timeout=60s
	kubectl -n sgroups-system rollout status deployment/sgroups-backend --timeout=60s

.PHONY: deploy-apiserver
deploy-apiserver: docker-build-apiserver ## Build and redeploy only the apiserver
	kind load docker-image $(IMAGE_APISERVER) --name $(KIND_CLUSTER)
	kubectl apply -k config/
	kubectl -n sgroups-system rollout restart deployment/sgroups-k8s-apiserver
	kubectl -n sgroups-system rollout status deployment/sgroups-k8s-apiserver --timeout=60s

.PHONY: undeploy
undeploy: ## Remove all resources from cluster
	kubectl delete -k config/ --ignore-not-found

##@ E2E

.PHONY: smoke-test
smoke-test: ## Run Postman/Newman smoke tests (requires running cluster)
	@kubectl proxy --port=8001 >/dev/null 2>&1 & PROXY_PID=$$!; \
	sleep 2; \
	npx newman run api/postman/sgroups-api.postman_collection.json \
		-e api/postman/kubectl-proxy.postman_environment.json; \
	RC=$$?; \
	kill $$PROXY_PID 2>/dev/null || true; \
	exit $$RC

##@ Help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)