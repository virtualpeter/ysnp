DOCKER_ORG ?= virtualpete
REGISTRY ?= docker.io
IMAGE ?= $(REGISTRY)/$(DOCKER_ORG)/ysnp
VERSION ?= $(shell cat VERSION 2>/dev/null || echo 0.0.0)

.PHONY: build test vet run docker compose-up

build:
	go build -o bin/ysnp ./cmd/ysnp

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/ysnp

docker:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

compose-up: docker
	IMAGE=$(IMAGE):latest docker compose up
