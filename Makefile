include Makefile.const

#Config variables

#Override this varible if you want to work with a specific target.
BUILD_TARGETS?=$(CURRENT_BIN_TARGETS)

#Version must be overrided in the CI 
VERSION?=local

# Docker options
TARGET_DOCKER_REGISTRY ?= $$USER

# Kubernetes options
TARGET_K8S_NAMESPACE ?= napptive


# Variables
BUILD_FOLDER=$(CURDIR)/build
BIN_FOLDER=$(BUILD_FOLDER)/bin
DOCKER_FOLDER=$(BUILD_FOLDER)/docker
K8S_FOLDER=$(BUILD_FOLDER)/k8s
TEMP_FOLDER=$(CURDIR)/temp

# Obtain the last commit hash
COMMIT=$(shell git log -1 --pretty=format:"%H")

# Tools
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_GENERATE=$(GO_CMD) generate
GO_TEST=$(GO_CMD) test
GO_LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X=main.Commit=$(COMMIT)"

UNAME := $(shell uname)

ifeq ($(UNAME), Darwin)
	SED := gsed
else
	SED := sed
endif

#Docker command
DOCKERCMD?=docker





.PHONY: all
all: clean test build 

.PHONY: clean
# Remove build files
clean:
	@echo "Cleaining build folder: $(BUILD_FOLDER)"
	@rm -rf $(BUILD_FOLDER)

.PHONY: test
# Test all golang files in the curdir
test:
	@echo "Executing golang tests"
	@$(GO_TEST) -v ./...

.PHONY: generate
# Generate the code objects.
generate:
	@echo "Generating golang extra files"
	@$(GO_GENERATE) -v ./...

.PHONY: coverage
# Create a coverage report for all golang files in the curdir
coverage:
	@echo "Creating golang test coverage report: $(BUILD_FOLDER)/coverage.out"
	@mkdir -p $(BUILD_FOLDER)
	@$(GO_TEST) -v ./... -coverprofile=$(BUILD_FOLDER)/cover.out

.PHONY: build
# Build target for local environment default
build: $(addsuffix .local,$(BUILD_TARGETS))

.PHONY: build-darwin
# Build target for darwin
build-darwin: $(addsuffix .darwin,$(BUILD_TARGETS))

.PHONY: build-linux
# Build target for linux
build-linux: $(addsuffix .linux,$(BUILD_TARGETS))

# Trigger the build operation for the local environment. Notice that the suffix is removed.
%.local:
	@echo "Building local binary $@"
	@$(GO_BUILD) $(GO_LDFLAGS) -o $(BIN_FOLDER)/local/$(basename $@) ./cmd/$(basename $@)/main.go

# Trigger the build operation for darwin. Notice that the suffix is removed as it is only used for Makefile expansion purposes.
%.darwin:
	@echo "Building darwin binary $@"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o $(BIN_FOLDER)/darwin/$(basename $@) ./cmd/$(basename $@)/main.go

# Trigger the build operation for linux. Notice that the suffix is removed as it is only used for Makefile expansion purposes.
%.linux:
	@ echo "Building linux binary $@"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) $(GO_LDFLAGS) -o $(BIN_FOLDER)/linux/$(basename $@) ./cmd/$(basename $@)/main.go

.PHONY: artifacts
artifacts: clean $(addsuffix .pkg-darwin,$(BUILD_TARGETS)) $(addsuffix .pkg-linux,$(BUILD_TARGETS))

%.pkg-darwin: %.darwin
	@echo "Packaging darwin binary"
	@mkdir -p build/compressed/darwin/$(basename $@)
	@mv build/bin/darwin/* build/compressed/darwin/$(basename $@)/.
	@cp README_public.md build/compressed/darwin/$(basename $@)/README.md
	@cd build/compressed/darwin && tar cvzf playground_$(VERSION).tgz playground

%.pkg-linux: %.linux
	@echo "Packaging linux binary"
	@mkdir -p build/compressed/linux/$(basename $@)
	@mv build/bin/linux/* build/compressed/linux/$(basename $@)/.
	@cp README_public.md build/compressed/linux/$(basename $@)/README.md
	@cd build/compressed/linux && tar cvzf playground_$(VERSION).tgz playground

.PHONY: docker-prep
docker-prep: $(addsuffix .docker-prep, $(BUILD_TARGETS))
%.docker-prep: %.linux
	@if [ -f docker/$(basename $@)/Dockerfile ]; then\
		echo "Preparing docker file for "$(basename $@);\
		rm -r $(DOCKER_FOLDER)/$(basename $@) || true;\
		mkdir -p $(DOCKER_FOLDER)/$(basename $@);\
		cp docker/$(basename $@)/* $(DOCKER_FOLDER)/$(basename $@)/.;\
		cp $(BIN_FOLDER)/linux/$(basename $@) $(DOCKER_FOLDER)/$(basename $@)/.;\
	fi

.PHONY: docker-build
docker-build: $(addsuffix .docker-build, $(BUILD_TARGETS))

%.docker-build: %.docker-prep
	@if [ -f $(DOCKER_FOLDER)/$(basename $@)/Dockerfile ]; then\
		echo "Building docker file for "$(basename $@);\
		$(DOCKERCMD) build --platform linux/amd64 $(DOCKER_FOLDER)/$(basename $@) -t $(TARGET_DOCKER_REGISTRY)/$(basename $@):$(VERSION);\
	fi

.PHONY: docker-push
docker-push: $(addsuffix .docker-push, $(BUILD_TARGETS))
%.docker-push: %.docker-build
	@echo Pushing $(basename $@) Docker image to DockerHub
	@if [ -f $(DOCKER_FOLDER)/$(basename $@)/Dockerfile ]; then\
		{ $(DOCKERCMD) push $(TARGET_DOCKER_REGISTRY)/$(basename $@):$(VERSION) || exit 1; } ; \
	fi

.PHONY: k8s
k8s:
	@if [ ! -d "deployments" ]; then \
		echo "Skipping k8s, no deployments found"; exit 0;\
	else \
		rm -r $(K8S_FOLDER) || true ; \
		mkdir -p $(K8S_FOLDER); \
		cp deployments/*.yaml $(K8S_FOLDER)/. ; \
		$(SED) -i 's/TARGET_K8S_NAMESPACE/$(TARGET_K8S_NAMESPACE)/' $(K8S_FOLDER)/*.yaml ;\
		$(SED) -i 's/TARGET_DOCKER_REGISTRY/'$(TARGET_DOCKER_REGISTRY)'/' $(K8S_FOLDER)/*.yaml ;\
		$(SED) -i 's/VERSION/$(VERSION)/' $(K8S_FOLDER)/*.yaml ;\
		echo "Kubernetes files ready at $(K8S_FOLDER)/"; \
	fi

.PHONY: k8s-kind
k8s-kind:
	@if [ ! -d "deployments" ]; then \
		echo "Skipping k8s, no deployments found"; exit 0;\
	else \
		rm -r $(K8S_FOLDER) || true ; \
		mkdir -p $(K8S_FOLDER); \
		cp deployments/*.yaml $(K8S_FOLDER)/. ; \
		rm $(K8S_FOLDER)/*.gcp.*.yaml ; \
		rm $(K8S_FOLDER)/*.aks.*.yaml ; \
		rm $(K8S_FOLDER)/*.onpremise.*.yaml ; \
		$(SED) -i 's/TARGET_K8S_NAMESPACE/$(TARGET_K8S_NAMESPACE)/' $(K8S_FOLDER)/*.yaml ;\
		$(SED) -i 's/TARGET_DOCKER_REGISTRY/'$(TARGET_DOCKER_REGISTRY)'/' $(K8S_FOLDER)/*.yaml ;\
		$(SED) -i 's/VERSION/$(VERSION)/' $(K8S_FOLDER)/*.yaml ;\
		echo "Kubernetes files ready at $(K8S_FOLDER)/"; \
	fi


.PHONY: release

release: clean build-darwin build-linux k8s
	@mkdir -p $(BUILD_FOLDER)
	@cp README.md $(BUILD_FOLDER)
	@if [ -d "deployments" ]; then \
		tar -czvf $(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz -C $(BUILD_FOLDER) bin k8s README.md; \
	elif [ -d $(BIN_FOLDER) ]; then \
		tar -czvf $(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz -C $(BUILD_FOLDER)  bin README.md; \
	else \
		tar -czvf $(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz -C $(BUILD_FOLDER)  README.md; \
	fi
	@echo "::set-output name=release_file::$(BUILD_FOLDER)/$(PROJECT_NAME)_$(VERSION).tar.gz"
	@echo "::set-output name=release_name::$(PROJECT_NAME)_$(VERSION).tar.gz"
	
.PHONY: prepare-integration-test-environment
prepare-integration-test-environment:
	@echo "Preparing integration tests environment"
	@echo "Launch elastic docker container..."
	@docker run -d --name elastic-env -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.13.1
	@echo "Wait several seconds until elastic is ready"
	@sleep 10
	@echo "Elastic ready!"
	@docker run -d --name local-postgres -e POSTGRES_PASSWORD=Pass2020! -p 5432:5432 postgres:13-alpine
	@echo "Database ready!"

.PHONY: it-load-db-schema
it-load-db-schema:
	@mkdir $(TEMP_FOLDER) || true
	@kubectl create -f deployments/$(PROJECT_NAME).005.configmap.all.all.yaml --dry-run=client -o "jsonpath={.data['$(PROJECT_NAME)-sql-script\.yaml']}" > $(TEMP_FOLDER)/db.yaml
	@docker run -it --network host -v $(TEMP_FOLDER):/it napptive/rdbms:latest schema load --scriptLoadPath=/it/db.yaml
	@echo "database schema has been loaded"

.PHONY: remove-integration-test-environment
remove-integration-test-environment:
	@echo "Removing integration tests environment"
	@docker stop elastic-env
	@docker rm elastic-env
	@echo "elastic removed!"
	@docker stop local-postgres
	@docker rm local-postgres
	@echo "postgres removed!"

.PHONY: all-tests
all-tests:
	@echo "Executing unit and integration tests"
	RUN_INTEGRATION_TEST=all IT_RUN_LOCAL=true $(GO_TEST) -v ./...