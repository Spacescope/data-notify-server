COMMIT_ID = $(shell git rev-parse --short HEAD)
ifeq ($(COMMIT_ID),)
COMMIT_ID = 'latest'
endif

.PHONY: test
IMAGE_PREFIX ?= 129862287110.dkr.ecr.us-east-2.amazonaws.com/data-infra
REGISTRY_SERVER ?= 129862287110.dkr.ecr.us-east-2.amazonaws.com/

help:
	@echo
	@echo "  binary - build binary"
	@echo "  build-notify-server - build docker images for centos"
	@echo "  swag - regenerate swag"
	@echo "  build-all - build docker images for centos"
	@echo "  push images to docker hub"

swag:
	swag init -g cmd/data-extraction-notify/main.go

binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/data-extraction-notify cmd/data-extraction-notify/main.go

test:
	go clean -testcache
	gotestsum --format pkgname

build-notify-server:
	docker build -t $(IMAGE_PREFIX)/data-extraction-notify-server:$(COMMIT_ID) -f build/Dockerfile .

build-all: build-dataapi

push:
	docker push $(IMAGE_PREFIX)/data-api-server:$(COMMIT_ID)
