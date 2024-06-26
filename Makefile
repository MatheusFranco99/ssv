ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

ifndef $(HOST_ADDRESS)
    HOST_ADDRESS=$(shell dig @resolver4.opendns.com myip.opendns.com +short)
    export HOST_ADDRESS
endif

ifndef $(BUILD_PATH)
    BUILD_PATH="/go/bin/ssvnode"
    export BUILD_PATH
endif

# node command builder
NODE_COMMAND=--config=${CONFIG_PATH}

ifneq ($(SHARE_CONFIG),)
  NODE_COMMAND+= --share-config=${SHARE_CONFIG}
endif

COV_CMD="-cover"
ifeq ($(COVERAGE),true)
	COV_CMD=-coverpkg=./... -covermode="atomic" -coverprofile="coverage.out"
endif
UNFORMATTED=$(shell gofmt -s -l .)

#Lint
.PHONY: lint-prepare
lint-prepare:
	@echo "Preparing Linter"
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

.PHONY: lint
lint:
	./bin/golangci-lint run -v ./...
	@if [ ! -z "${UNFORMATTED}" ]; then \
		echo "Some files requires formatting, please run 'go fmt ./...'"; \
		exit 1; \
	fi

.PHONY: full-test
full-test:
	@echo "Running all tests"
	@go test -tags blst_enabled -timeout 20m ${COV_CMD} -p 1 -v ./...

.PHONY: integration-test
integration-test:
	@echo "Running integration tests"
	@go test -tags blst_enabled -timeout 20m ${COV_CMD} -p 1 -v ./integration/...

.PHONY: unit-test
unit-test:
	@echo "Running unit tests"
	@go test -tags blst_enabled -timeout 20m ${COV_CMD} -race -p 1 -v `go list ./... | grep -ve "spectest\|integration\|ssv/scripts/"`

.PHONY: spec-test
spec-test:
	@echo "Running spec tests"
	@go test -tags blst_enabled -timeout 15m ${COV_CMD} -race -p 1 -v `go list ./... | grep spectest`

#Test
.PHONY: docker-spec-test
docker-spec-test:
	@echo "Running spec tests in docker"
	@docker build -t ssv_tests -f tests.Dockerfile .
	@docker run --rm ssv_tests make spec-test

#Test
.PHONY: docker-unit-test
docker-unit-test:
	@echo "Running unit tests in docker"
	@docker build -t ssv_tests -f tests.Dockerfile .
	@docker run --rm ssv_tests make unit-test

.PHONY: docker-integration-test
docker-integration-test:
	@echo "Running integration tests in docker"
	@docker build -t ssv_tests -f tests.Dockerfile .
	@docker run --rm ssv_tests make integration-test

#Build
.PHONY: build
build:
	CGO_ENABLED=1 go build -o ./bin/ssvnode -ldflags "-X main.Version=`git describe --tags $(git rev-list --tags --max-count=1)`" ./cmd/ssvnode/

.PHONY: start-node
start-node:
	@echo "Build ${BUILD_PATH}"
	@echo "Build ${CONFIG_PATH}"
	@echo "Build ${CONFIG_PATH2}"
	@echo "Command ${NODE_COMMAND}"
ifdef DEBUG_PORT
	@echo "Running node-${NODE_ID} in debug mode"
	@dlv  --continue --accept-multiclient --headless --listen=:${DEBUG_PORT} --api-version=2 exec \
	 ${BUILD_PATH} start-node -- ${NODE_COMMAND}
else
	@echo "Running node on address: ${HOST_ADDRESS})"
	@${BUILD_PATH} start-node ${NODE_COMMAND}
endif

.PHONY: docker
docker:
	@echo "node ${NODES_ID}"
	@docker rm -f ssv_node && docker build -t ssv_node . && docker run -d --env-file .env --restart unless-stopped --name=ssv_node -p 13000:13000 -p 12000:12000/udp -it ssv_node make BUILD_PATH=/go/bin/ssvnode  start-node && docker logs ssv_node --follow

.PHONY: docker-image
docker-image:
	@echo "node ${NODES_ID}"
	@sudo docker rm -f ssv_node && docker run -d --env-file .env --restart unless-stopped --name=ssv_node -p 13000:13000 -p 12000:12000/udp 'bloxstaking/ssv-node:latest' make BUILD_PATH=/go/bin/ssvnode start-node

NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4
.PHONY: docker-all
docker-all:
	@echo "nodes $(NODES)"
	@docker-compose up --build $(NODES)

NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4
.PHONY: docker-local
docker-local:
	@echo "nodes $(NODES)"
	@docker-compose -f docker-compose-local.yaml up --build $(NODES)

# NODES=ssv-node-1
# .PHONY: docker-local1
# docker-local1:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-1.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2
# .PHONY: docker-local2
# docker-local2:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-2.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7 ssv-node-8
# .PHONY: docker-local8
# docker-local8:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-8.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7
# .PHONY: docker-local7
# docker-local7:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-7.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7 ssv-node-8 ssv-node-9 ssv-node-10
# .PHONY: docker-local10
# docker-local10:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-10.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7 ssv-node-8 ssv-node-9 ssv-node-10 ssv-node-11 ssv-node-12 ssv-node-13
# .PHONY: docker-local13
# docker-local13:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-13.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7 ssv-node-8 ssv-node-9 ssv-node-10 ssv-node-11 ssv-node-12 ssv-node-13 ssv-node-14 ssv-node-15 ssv-node-16 ssv-node-17 ssv-node-18 ssv-node-19
# .PHONY: docker-local19
# docker-local19:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-19.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7 ssv-node-8 ssv-node-9 ssv-node-10 ssv-node-11 ssv-node-12 ssv-node-13 ssv-node-14 ssv-node-15 ssv-node-16 ssv-node-17 ssv-node-18 ssv-node-19 ssv-node-20 ssv-node-21 ssv-node-22 ssv-node-23 ssv-node-24 ssv-node-25 ssv-node-26 ssv-node-27 ssv-node-28 ssv-node-29 ssv-node-30 ssv-node-31
# .PHONY: docker-local31
# docker-local31:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-31.yaml up --build $(NODES)

# NODES=ssv-node-1 ssv-node-2 ssv-node-3 ssv-node-4 ssv-node-5 ssv-node-6 ssv-node-7 ssv-node-8 ssv-node-9 ssv-node-10 ssv-node-11 ssv-node-12 ssv-node-13 ssv-node-14 ssv-node-15 ssv-node-16
# .PHONY: docker-local16
# docker-local16:
# 	@echo "nodes $(NODES)"
# 	@docker-compose -f docker-compose-local-16.yaml up --build $(NODES)

DEBUG_NODES=ssv-node-1-dev ssv-node-2-dev ssv-node-3-dev ssv-node-4-dev
.PHONY: docker-debug
docker-debug:
	@echo $(DEBUG_NODES)
	@docker-compose up --build $(DEBUG_NODES)

.PHONY: stop
stop:
	@docker-compose down

.PHONY: start-boot-node
start-boot-node:
	@echo "Running start-boot-node"
	${BUILD_PATH} start-boot-node

MONITOR_NODES=prometheus grafana
.PHONY: docker-monitor
docker-monitor:
	@echo $(MONITOR_NODES)
	@docker-compose up --build $(MONITOR_NODES)