# StreamForge Monorepo Makefile
# Unified build and development orchestration

# Project configuration
PROJECT_NAME := streamforge
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Directories
APPS_DIR := apps
PACKAGES_DIR := packages
INFRA_DIR := infra
DOCS_DIR := docs
EXAMPLES_DIR := examples
BENCHMARKS_DIR := benchmarks
SCRIPTS_DIR := scripts
TOOLS_DIR := tools

# Language versions
NODE_VERSION := 20
RUST_VERSION := 1.75.0
GO_VERSION := 1.21
PYTHON_VERSION := 3.11

# Docker settings
DOCKER_REGISTRY := ghcr.io/streamforge
DOCKER_TAG := $(VERSION)

# Color output
RESET := \033[0m
BOLD := \033[1m
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m

define log
	@echo "$(CYAN)$(BOLD)[$(PROJECT_NAME)]$(RESET) $(1)"
endef

define log_success
	@echo "$(GREEN)$(BOLD)✅ $(1)$(RESET)"
endef

define log_error
	@echo "$(RED)$(BOLD)❌ $(1)$(RESET)"
endef

define log_warning
	@echo "$(YELLOW)$(BOLD)⚠️  $(1)$(RESET)"
endef

.PHONY: help
help: ## Show this help message
	@echo "$(BOLD)$(MAGENTA)StreamForge Development Commands$(RESET)"
	@echo ""
	@echo "$(BOLD)🚀 Quick Start:$(RESET)"
	@echo "  make dev-setup    - Setup development environment"
	@echo "  make dev-up       - Start all development services"
	@echo "  make demo         - Run demo environment"
	@echo ""
	@echo "$(BOLD)📋 Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# =============================================================================
# 🔧 Development Environment
# =============================================================================

.PHONY: dev-setup
dev-setup: ## Setup development environment
	$(call log,"Setting up development environment...")
	@$(SCRIPTS_DIR)/setup-dev.sh
	$(call log_success,"Development environment ready!")

.PHONY: dev-up
dev-up: ## Start all development services
	$(call log,"Starting development services...")
	docker-compose -f docker-compose.dev.yml up -d
	$(call log_success,"Development services started!")
	@echo "$(BOLD)🌐 Services:$(RESET)"
	@echo "  Dashboard:     http://localhost:3000"
	@echo "  API Gateway:   http://localhost:8080"
	@echo "  Grafana:       http://localhost:3001"
	@echo "  Jaeger:        http://localhost:16686"

.PHONY: dev-down
dev-down: ## Stop all development services
	$(call log,"Stopping development services...")
	docker-compose -f docker-compose.dev.yml down
	$(call log_success,"Development services stopped!")

.PHONY: dev-logs
dev-logs: ## Show development services logs
	docker-compose -f docker-compose.dev.yml logs -f

.PHONY: dev-clean
dev-clean: ## Clean development environment
	$(call log,"Cleaning development environment...")
	docker-compose -f docker-compose.dev.yml down -v --remove-orphans
	docker system prune -f
	$(call log_success,"Development environment cleaned!")

.PHONY: demo
demo: ## Start demo environment with sample data
	$(call log,"Starting demo environment...")
	docker-compose -f docker-compose.demo.yml up -d
	@sleep 10
	@$(SCRIPTS_DIR)/generate-demo-data.sh
	$(call log_success,"Demo environment ready!")
	@echo "$(BOLD)🎯 Demo Dashboard: http://localhost:3000$(RESET)"

# =============================================================================
# 📦 Dependencies Management
# =============================================================================

.PHONY: deps-install
deps-install: deps-install-node deps-install-rust deps-install-go deps-install-python ## Install all dependencies

.PHONY: deps-install-node
deps-install-node: ## Install Node.js dependencies
	$(call log,"Installing Node.js dependencies...")
	pnpm install --frozen-lockfile
	$(call log_success,"Node.js dependencies installed!")

.PHONY: deps-install-rust
deps-install-rust: ## Install Rust dependencies
	$(call log,"Installing Rust dependencies...")
	cargo fetch --locked
	$(call log_success,"Rust dependencies installed!")

.PHONY: deps-install-go
deps-install-go: ## Install Go dependencies
	$(call log,"Installing Go dependencies...")
	@for dir in $(shell find $(APPS_DIR) $(PACKAGES_DIR) -name "go.mod" -exec dirname {} \;); do \
		echo "Installing dependencies in $$dir..."; \
		cd $$dir && go mod download && cd - > /dev/null; \
	done
	$(call log_success,"Go dependencies installed!")

.PHONY: deps-install-python
deps-install-python: ## Install Python dependencies
	$(call log,"Installing Python dependencies...")
	pip install -r requirements.txt
	pip install -r requirements-dev.txt
	$(call log_success,"Python dependencies installed!")

.PHONY: deps-update
deps-update: ## Update all dependencies
	$(call log,"Updating dependencies...")
	pnpm update
	cargo update
	@for dir in $(shell find $(APPS_DIR) $(PACKAGES_DIR) -name "go.mod" -exec dirname {} \;); do \
		cd $$dir && go get -u ./... && go mod tidy && cd - > /dev/null; \
	done
	pip-review --auto
	$(call log_success,"Dependencies updated!")

.PHONY: deps-audit
deps-audit: ## Audit dependencies for vulnerabilities
	$(call log,"Auditing dependencies...")
	pnpm audit --audit-level moderate
	cargo audit
	@for dir in $(shell find $(APPS_DIR) $(PACKAGES_DIR) -name "go.mod" -exec dirname {} \;); do \
		cd $$dir && govulncheck ./... && cd - > /dev/null; \
	done
	pip-audit
	$(call log_success,"Dependencies audit completed!")

# =============================================================================
# 🧪 Testing
# =============================================================================

.PHONY: test
test: test-unit test-integration ## Run all tests

.PHONY: test-unit
test-unit: test-unit-node test-unit-rust test-unit-go test-unit-python ## Run unit tests

.PHONY: test-unit-node
test-unit-node: ## Run Node.js unit tests
	$(call log,"Running Node.js unit tests...")
	pnpm test:unit
	$(call log_success,"Node.js unit tests passed!")

.PHONY: test-unit-rust
test-unit-rust: ## Run Rust unit tests
	$(call log,"Running Rust unit tests...")
	cargo test --workspace --lib
	$(call log_success,"Rust unit tests passed!")

.PHONY: test-unit-go
test-unit-go: ## Run Go unit tests
	$(call log,"Running Go unit tests...")
	@for dir in $(shell find $(APPS_DIR) $(PACKAGES_DIR) -name "go.mod" -exec dirname {} \;); do \
		cd $$dir && go test -v -race -coverprofile=coverage.out ./... && cd - > /dev/null; \
	done
	$(call log_success,"Go unit tests passed!")

.PHONY: test-unit-python
test-unit-python: ## Run Python unit tests
	$(call log,"Running Python unit tests...")
	pytest -v --cov=. --cov-report=xml --cov-report=html
	$(call log_success,"Python unit tests passed!")

.PHONY: test-integration
test-integration: ## Run integration tests
	$(call log,"Running integration tests...")
	docker-compose -f docker-compose.test.yml up -d
	@sleep 15
	pnpm test:integration
	docker-compose -f docker-compose.test.yml down
	$(call log_success,"Integration tests passed!")

.PHONY: test-e2e
test-e2e: ## Run E2E tests
	$(call log,"Running E2E tests...")
	pnpm playwright test
	$(call log_success,"E2E tests passed!")

.PHONY: test-benchmark
test-benchmark: ## Run performance benchmarks
	$(call log,"Running performance benchmarks...")
	cargo bench --workspace
	@for dir in $(shell find $(APPS_DIR) -name "go.mod" -exec dirname {} \;); do \
		cd $$dir && go test -bench=. -benchmem ./... && cd - > /dev/null; \
	done
	pnpm benchmark
	$(call log_success,"Performance benchmarks completed!")

.PHONY: test-coverage
test-coverage: ## Generate test coverage report
	$(call log,"Generating test coverage report...")
	@$(SCRIPTS_DIR)/generate-coverage.sh
	$(call log_success,"Coverage report generated!")

# =============================================================================
# 🔍 Code Quality
# =============================================================================

.PHONY: lint
lint: lint-node lint-rust lint-go lint-python lint-docker lint-yaml ## Run all linters

.PHONY: lint-node
lint-node: ## Run Node.js linters
	$(call log,"Running Node.js linters...")
	pnpm lint
	$(call log_success,"Node.js linting passed!")

.PHONY: lint-rust
lint-rust: ## Run Rust linters
	$(call log,"Running Rust linters...")
	cargo clippy --workspace --all-targets --all-features -- -D warnings
	$(call log_success,"Rust linting passed!")

.PHONY: lint-go
lint-go: ## Run Go linters
	$(call log,"Running Go linters...")
	@for dir in $(shell find $(APPS_DIR) $(PACKAGES_DIR) -name "go.mod" -exec dirname {} \;); do \
		cd $$dir && golangci-lint run && cd - > /dev/null; \
	done
	$(call log_success,"Go linting passed!")

.PHONY: lint-python
lint-python: ## Run Python linters
	$(call log,"Running Python linters...")
	ruff check .
	mypy .
	$(call log_success,"Python linting passed!")

.PHONY: lint-docker
lint-docker: ## Run Dockerfile linters
	$(call log,"Running Dockerfile linters...")
	@find . -name "Dockerfile*" -exec hadolint {} \;
	$(call log_success,"Dockerfile linting passed!")

.PHONY: lint-yaml
lint-yaml: ## Run YAML linters
	$(call log,"Running YAML linters...")
	yamllint .
	$(call log_success,"YAML linting passed!")

.PHONY: fmt
fmt: fmt-node fmt-rust fmt-go fmt-python ## Format all code

.PHONY: fmt-node
fmt-node: ## Format Node.js code
	$(call log,"Formatting Node.js code...")
	pnpm format
	$(call log_success,"Node.js code formatted!")

.PHONY: fmt-rust
fmt-rust: ## Format Rust code
	$(call log,"Formatting Rust code...")
	cargo fmt --all
	$(call log_success,"Rust code formatted!")

.PHONY: fmt-go
fmt-go: ## Format Go code
	$(call log,"Formatting Go code...")
	@find $(APPS_DIR) $(PACKAGES_DIR) -name "*.go" -exec gofmt -w {} \;
	@find $(APPS_DIR) $(PACKAGES_DIR) -name "*.go" -exec goimports -w {} \;
	$(call log_success,"Go code formatted!")

.PHONY: fmt-python
fmt-python: ## Format Python code
	$(call log,"Formatting Python code...")
	ruff format .
	$(call log_success,"Python code formatted!")

.PHONY: pre-commit
pre-commit: fmt lint test-unit ## Run pre-commit checks
	$(call log_success,"Pre-commit checks passed!")

# =============================================================================
# 🔨 Build
# =============================================================================

.PHONY: build
build: build-dashboard build-api-gateway build-stream-processor build-ml-engine build-collector build-operator ## Build all applications

.PHONY: build-dashboard
build-dashboard: ## Build dashboard app
	$(call log,"Building dashboard...")
	cd $(APPS_DIR)/dashboard && pnpm build
	$(call log_success,"Dashboard built!")

.PHONY: build-api-gateway
build-api-gateway: ## Build API gateway
	$(call log,"Building API gateway...")
	cd $(APPS_DIR)/api-gateway && go build -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT_SHA) -X main.date=$(BUILD_DATE)" -o bin/api-gateway ./cmd/server
	$(call log_success,"API gateway built!")

.PHONY: build-stream-processor
build-stream-processor: ## Build stream processor
	$(call log,"Building stream processor...")
	cd $(APPS_DIR)/stream-processor && cargo build --release
	$(call log_success,"Stream processor built!")

.PHONY: build-ml-engine
build-ml-engine: ## Build ML engine
	$(call log,"Building ML engine...")
	cd $(APPS_DIR)/ml-engine && python -m build
	$(call log_success,"ML engine built!")

.PHONY: build-collector
build-collector: ## Build collector
	$(call log,"Building collector...")
	cd $(APPS_DIR)/collector && go build -ldflags="-X main.version=$(VERSION)" -o bin/collector ./cmd/collector
	$(call log_success,"Collector built!")

.PHONY: build-operator
build-operator: ## Build Kubernetes operator
	$(call log,"Building Kubernetes operator...")
	cd $(APPS_DIR)/operator && go build -ldflags="-X main.version=$(VERSION)" -o bin/operator ./cmd/manager
	$(call log_success,"Kubernetes operator built!")

# =============================================================================
# 🐳 Docker
# =============================================================================

.PHONY: docker-build
docker-build: docker-build-dashboard docker-build-api-gateway docker-build-stream-processor docker-build-ml-engine docker-build-collector docker-build-operator ## Build all Docker images

.PHONY: docker-build-dashboard
docker-build-dashboard: ## Build dashboard Docker image
	$(call log,"Building dashboard Docker image...")
	docker build -t $(DOCKER_REGISTRY)/dashboard:$(DOCKER_TAG) -f $(APPS_DIR)/dashboard/Dockerfile .
	$(call log_success,"Dashboard Docker image built!")

.PHONY: docker-build-api-gateway
docker-build-api-gateway: ## Build API gateway Docker image
	$(call log,"Building API gateway Docker image...")
	docker build -t $(DOCKER_REGISTRY)/api-gateway:$(DOCKER_TAG) -f $(APPS_DIR)/api-gateway/Dockerfile .
	$(call log_success,"API gateway Docker image built!")

.PHONY: docker-build-stream-processor
docker-build-stream-processor: ## Build stream processor Docker image
	$(call log,"Building stream processor Docker image...")
	docker build -t $(DOCKER_REGISTRY)/stream-processor:$(DOCKER_TAG) -f $(APPS_DIR)/stream-processor/Dockerfile .
	$(call log_success,"Stream processor Docker image built!")

.PHONY: docker-build-ml-engine
docker-build-ml-engine: ## Build ML engine Docker image
	$(call log,"Building ML engine Docker image...")
	docker build -t $(DOCKER_REGISTRY)/ml-engine:$(DOCKER_TAG) -f $(APPS_DIR)/ml-engine/Dockerfile .
	$(call log_success,"ML engine Docker image built!")

.PHONY: docker-build-collector
docker-build-collector: ## Build collector Docker image
	$(call log,"Building collector Docker image...")
	docker build -t $(DOCKER_REGISTRY)/collector:$(DOCKER_TAG) -f $(APPS_DIR)/collector/Dockerfile .
	$(call log_success,"Collector Docker image built!")

.PHONY: docker-build-operator
docker-build-operator: ## Build operator Docker image
	$(call log,"Building operator Docker image...")
	docker build -t $(DOCKER_REGISTRY)/operator:$(DOCKER_TAG) -f $(APPS_DIR)/operator/Dockerfile .
	$(call log_success,"Operator Docker image built!")

.PHONY: docker-push
docker-push: ## Push all Docker images
	$(call log,"Pushing Docker images...")
	@for service in dashboard api-gateway stream-processor ml-engine collector operator; do \
		docker push $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG); \
	done
	$(call log_success,"Docker images pushed!")

.PHONY: docker-scan
docker-scan: ## Scan Docker images for vulnerabilities
	$(call log,"Scanning Docker images...")
	@for service in dashboard api-gateway stream-processor ml-engine collector operator; do \
		trivy image $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG); \
	done
	$(call log_success,"Docker image scanning completed!")

# =============================================================================
# ☸️  Kubernetes
# =============================================================================

.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	$(call log,"Deploying to Kubernetes...")
	helm upgrade --install streamforge $(INFRA_DIR)/helm/streamforge \
		--set image.tag=$(DOCKER_TAG) \
		--set global.environment=development
	$(call log_success,"Deployed to Kubernetes!")

.PHONY: k8s-status
k8s-status: ## Check Kubernetes deployment status
	$(call log,"Checking Kubernetes status...")
	kubectl get pods -l app.kubernetes.io/name=streamforge
	kubectl get services -l app.kubernetes.io/name=streamforge
	$(call log_success,"Kubernetes status checked!")

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs
	kubectl logs -f -l app.kubernetes.io/name=streamforge --all-containers=true

.PHONY: k8s-clean
k8s-clean: ## Clean Kubernetes deployment
	$(call log,"Cleaning Kubernetes deployment...")
	helm uninstall streamforge
	$(call log_success,"Kubernetes deployment cleaned!")

# =============================================================================
# 📚 Documentation
# =============================================================================

.PHONY: docs-dev
docs-dev: ## Start documentation development server
	$(call log,"Starting documentation server...")
	cd $(DOCS_DIR)/website && pnpm dev

.PHONY: docs-build
docs-build: ## Build documentation
	$(call log,"Building documentation...")
	cd $(DOCS_DIR)/website && pnpm build
	$(call log_success,"Documentation built!")

.PHONY: docs-generate-api
docs-generate-api: ## Generate API documentation
	$(call log,"Generating API documentation...")
	@$(SCRIPTS_DIR)/generate-api-docs.sh
	$(call log_success,"API documentation generated!")

.PHONY: docs-generate-examples
docs-generate-examples: ## Generate code examples
	$(call log,"Generating code examples...")
	@$(SCRIPTS_DIR)/generate-examples.sh
	$(call log_success,"Code examples generated!")

# =============================================================================
# 🔐 Security
# =============================================================================

.PHONY: security-scan
security-scan: ## Run security scans
	$(call log,"Running security scans...")
	@$(SCRIPTS_DIR)/security-scan.sh
	$(call log_success,"Security scans completed!")

.PHONY: sbom-generate
sbom-generate: ## Generate Software Bill of Materials
	$(call log,"Generating SBOM...")
	@$(SCRIPTS_DIR)/generate-sbom.sh
	$(call log_success,"SBOM generated!")

.PHONY: sign-artifacts
sign-artifacts: ## Sign release artifacts
	$(call log,"Signing artifacts...")
	@$(SCRIPTS_DIR)/sign-artifacts.sh
	$(call log_success,"Artifacts signed!")

# =============================================================================
# 📊 Monitoring & Health
# =============================================================================

.PHONY: health-check
health-check: ## Check system health
	$(call log,"Checking system health...")
	@$(SCRIPTS_DIR)/health-check.sh
	$(call log_success,"Health check completed!")

.PHONY: metrics
metrics: ## Show system metrics
	$(call log,"Fetching system metrics...")
	@curl -s http://localhost:8080/metrics || echo "Metrics endpoint not available"

.PHONY: trace
trace: ## Show distributed traces
	$(call log,"Opening Jaeger UI...")
	@open http://localhost:16686 2>/dev/null || echo "Jaeger UI: http://localhost:16686"

# =============================================================================
# 🧹 Cleanup
# =============================================================================

.PHONY: clean
clean: clean-build clean-test clean-deps ## Clean all generated files

.PHONY: clean-build
clean-build: ## Clean build artifacts
	$(call log,"Cleaning build artifacts...")
	@find . -name "target" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name ".next" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "build" -type d -exec rm -rf {} + 2>/dev/null || true
	$(call log_success,"Build artifacts cleaned!")

.PHONY: clean-test
clean-test: ## Clean test artifacts
	$(call log,"Cleaning test artifacts...")
	@find . -name "coverage.out" -delete 2>/dev/null || true
	@find . -name "coverage.xml" -delete 2>/dev/null || true
	@find . -name "htmlcov" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name ".coverage" -delete 2>/dev/null || true
	@find . -name "test-results" -type d -exec rm -rf {} + 2>/dev/null || true
	$(call log_success,"Test artifacts cleaned!")

.PHONY: clean-deps
clean-deps: ## Clean dependency caches
	$(call log,"Cleaning dependency caches...")
	@rm -rf node_modules
	@cargo clean
	@go clean -cache -modcache
	@pip cache purge
	$(call log_success,"Dependency caches cleaned!")

# =============================================================================
# 🚀 Release
# =============================================================================

.PHONY: release-prepare
release-prepare: ## Prepare release
	$(call log,"Preparing release $(VERSION)...")
	@$(SCRIPTS_DIR)/prepare-release.sh $(VERSION)
	$(call log_success,"Release $(VERSION) prepared!")

.PHONY: release-publish
release-publish: ## Publish release
	$(call log,"Publishing release $(VERSION)...")
	@$(SCRIPTS_DIR)/publish-release.sh $(VERSION)
	$(call log_success,"Release $(VERSION) published!")

# =============================================================================
# 🎯 Development Workflows
# =============================================================================

.PHONY: dev-dashboard
dev-dashboard: ## Start only dashboard in dev mode
	cd $(APPS_DIR)/dashboard && pnpm dev

.PHONY: dev-api-gateway
dev-api-gateway: ## Start only API gateway in dev mode
	cd $(APPS_DIR)/api-gateway && go run ./cmd/server

.PHONY: dev-stream-processor
dev-stream-processor: ## Start only stream processor in dev mode
	cd $(APPS_DIR)/stream-processor && cargo run

.PHONY: dev-ml-engine
dev-ml-engine: ## Start only ML engine in dev mode
	cd $(APPS_DIR)/ml-engine && python -m uvicorn main:app --reload

# =============================================================================
# 🔧 Utilities
# =============================================================================

.PHONY: version
version: ## Show version information
	@echo "$(BOLD)StreamForge Version Information$(RESET)"
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT_SHA)"
	@echo "Build Date: $(BUILD_DATE)"

.PHONY: check-tools
check-tools: ## Check required tools
	$(call log,"Checking required tools...")
	@$(SCRIPTS_DIR)/check-tools.sh
	$(call log_success,"Tool check completed!")

.PHONY: update-hooks
update-hooks: ## Update git hooks
	$(call log,"Updating git hooks...")
	@cp $(SCRIPTS_DIR)/hooks/* .git/hooks/
	@chmod +x .git/hooks/*
	$(call log_success,"Git hooks updated!")

# Default target
.DEFAULT_GOAL := help 