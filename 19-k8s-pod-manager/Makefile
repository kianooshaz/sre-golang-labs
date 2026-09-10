# Local Kubernetes (minikube) + pod-manager app for the SRE Go class.
# Usage: see "make help" or the README.

APP_NAME     := pod-manager
GO           ?= go
BIN          := bin/$(APP_NAME)
PROFILE      := sre-lab
CLUSTER_NAME := $(PROFILE)
MEMORY       := 4096
CPUS         := 2
KUBECTL      := kubectl

.DEFAULT_GOAL := help
.PHONY: help build run tidy vet test k8s-start k8s-stop k8s-status k8s-deploy k8s-pods k8s-logs clean clean-all

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}'

## ---------- Go application ----------

build: ## Build the app binary into bin/
	$(GO) build -o $(BIN) .

tidy: ## Download Go dependencies
	$(GO) mod tidy

vet: ## Run go vet
	$(GO) vet ./...

test: ## Run tests
	$(GO) test ./...

run: ## Run the app against the local cluster (needs k8s-start)
	$(GO) run .

## ---------- Local Kubernetes (minikube) ----------

k8s-start: ## Start the minikube cluster
	minikube start --profile $(PROFILE) --memory $(MEMORY) --cpus $(CPUS)
	minikube profile $(PROFILE)
	@echo "waiting for cluster nodes to be ready..."
	$(KUBECTL) --context $(CLUSTER_NAME) wait --for=condition=ready node --all --timeout=180s

k8s-stop: ## Stop the minikube cluster
	minikube stop --profile $(PROFILE)

k8s-status: ## Show cluster status
	minikube status --profile $(PROFILE)

k8s-deploy: ## Create namespace and the 2 demo pods
	$(KUBECTL) apply -f k8s/namespace.yaml
	$(KUBECTL) apply -f k8s/pods.yaml
	@echo "waiting for demo pods..."
	$(KUBECTL) -n pod-manager wait --for=condition=Ready pod --all --timeout=180s

k8s-pods: ## List pods in the pod-manager namespace
	$(KUBECTL) -n pod-manager get pods -o wide

k8s-logs: ## Tail logs of the demo-web pod
	$(KUBECTL) -n pod-manager logs -f demo-web

## ---------- Cleanup ----------

clean: ## Remove build artifacts
	rm -rf bin

clean-all: clean ## Delete the demo pods and the minikube cluster
	-$(KUBECTL) delete namespace pod-manager --ignore-not-found --wait=false
	minikube delete --profile $(PROFILE)
