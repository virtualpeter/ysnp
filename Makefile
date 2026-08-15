DOCKER_ORG ?= virtualpete
REGISTRY ?= docker.io
IMAGE ?= $(REGISTRY)/$(DOCKER_ORG)/ysnp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)
LDFLAGS ?= -X main.version=$(VERSION)

.PHONY: build test vet run docker compose-up

build:
	go build -ldflags "$(LDFLAGS)" -o bin/ysnp ./cmd/ysnp

test:
	go test ./...

vet:
	go vet ./...

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/ysnp

docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

compose-up: docker
	IMAGE=$(IMAGE):latest docker compose up
