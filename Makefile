##### Basic go commands
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod


# Binary names
BINARY_EXE=$(shell go env GOEXE)
BINARY_CLIENT_NAME=dcenter_bridge_client$(BINARY_EXE)
BINARY_SERVER_NAME=dcenter_bridge_server$(BINARY_EXE)
BINARY_HEALTH_CHECK=health_check$(BINARY_EXE)
SWAG_BIN=$(shell go env GOPATH)/bin/swag

all: build


.PHONY: dcenter_bridge
dcenter_bridge:
	$(GOBUILD) -o bin/$(BINARY_CLIENT_NAME) -v -gcflags '-N -l' examples/client/client.go
	$(GOBUILD) -o bin/$(BINARY_SERVER_NAME) -v -gcflags '-N -l' examples/server/main.go
	$(GOBUILD) -o bin/$(BINARY_HEALTH_CHECK) -v -ldflags="-s -w" -gcflags '-N -l' health/health_check.go

.PHONY: docs
docs:
	$(SWAG_BIN) init -g main.go -o docs/swagger

.PHONY: build
build: dcenter_bridge

.PHONY: run
run:
	$(GORUN) -v examples/server/main.go

.PHONY: dock
dock: dock-build dock-push

.PHONY: dock-build
dock-build:
	docker-compose -f docker-compose.yml build online

.PHONY: dock-push
dock-push:
	docker-compose -f docker-compose.yml push online

.PHONY: dock-run
dock-run:
	docker-compose -f docker-compose.yml up -d

.PHONY: dock-rmi
dock-rmi:
	#docker images | grep dcenter_bridge_client-go_online | grep -v v1.0.0 | awk -F " " '{print $3}' | xargs docker rmi
	docker rmi -f $(docker images | grep dcenter_bridge_client-go_online | grep -v v1.0.0 | awk '{print $3}')
	docker image prune -f
	docker rmi $(docker images -f "dangling=true" -q)

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: bench
bench:
	$(GOTEST) -v -bench=. ./...

.PHONY: deps
deps:
	$(GOMOD) tidy -v

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
