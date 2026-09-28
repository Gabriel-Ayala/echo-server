.DEFAULT_GOAL := help

GO ?= go
BINARY := bin/server

# Docker is used for image builds and the Kind cluster.
KIND_CLUSTER_NAME ?= echo
IMAGE ?= docker.io/library/web-server-eccho:local
PULUMI ?= pulumi
PULUMI_STACK ?= dev
# Local development only; matches the existing local stack.
PULUMI_CONFIG_PASSPHRASE ?= local-development
PORT ?= 8080
ACT ?= act
ACT_EVENT ?= push
ACT_IMAGE ?= catthehacker/ubuntu:act-latest
ACT_FLAGS ?=

export KIND_CLUSTER_NAME IMAGE
export PULUMI_CONFIG_PASSPHRASE PORT

.PHONY: help build run test fmt vet ci act act-list pulumi-setup preview deploy destroy outputs clean

help:
	@printf '%s\n' \
	  'make build  Build the server into bin/server' \
	  'make run    Run the server on port 8080' \
	  'make test   Run all Go tests' \
	  'make fmt    Format Go source files' \
	  'make vet    Run Go static analysis' \
	  'make ci     Test, create Kind if needed, build and load the image' \
	  'make act    Run the GitHub Actions workflow locally with Docker' \
	  'make act-list  List the workflow jobs' \
	  'make pulumi-setup  Initialize local state and configure the stack' \
	  'make preview       Preview infrastructure changes' \
	  'make deploy        Build/load the image and deploy with Pulumi' \
	  'make outputs       Show deployment, namespace, and service names' \
	  'make destroy       Remove stack resources (keeps the Kind cluster)' \
	  'make clean  Remove the compiled server' \
	  '' \
	  'Defaults: Docker; Kind cluster echo' \
	  'Image tag: make ci IMAGE=docker.io/library/web-server-eccho:v2'

build:
	@mkdir -p bin
	$(GO) build -o $(BINARY) ./cmd/echo-server

run:
	$(GO) run ./cmd/echo-server

test:
	$(GO) test ./...
	cd infra && $(GO) test ./...

fmt:
	$(GO) fmt ./...
	cd infra && $(GO) fmt ./...

vet:
	$(GO) vet ./...
	cd infra && $(GO) vet ./...

ci:
	./scripts/ci.sh

act:
	@git rev-parse --verify HEAD >/dev/null 2>&1 || { echo 'Create an initial Git commit before running make act (CI uses the commit SHA as its image tag).' >&2; exit 1; }
	$(ACT) $(ACT_EVENT) -W .github/workflows/ci.yml -P ubuntu-latest=$(ACT_IMAGE) --network host $(ACT_FLAGS)

act-list:
	$(ACT) --list -W .github/workflows/ci.yml

pulumi-setup:
	@mkdir -p infra/.pulumi
	cd infra && $(PULUMI) stack select "$(PULUMI_STACK)" --create --non-interactive
	cd infra && $(PULUMI) config set image "$(IMAGE)" --stack "$(PULUMI_STACK)"
	cd infra && $(PULUMI) config set clusterName "$(KIND_CLUSTER_NAME)" --stack "$(PULUMI_STACK)"
	cd infra && $(PULUMI) config set port "$(PORT)" --stack "$(PULUMI_STACK)"

preview: pulumi-setup
	cd infra && $(PULUMI) preview --stack "$(PULUMI_STACK)"

deploy: ci
	$(MAKE) pulumi-setup
	cd infra && $(PULUMI) up --yes --stack "$(PULUMI_STACK)"

outputs:
	cd infra && $(PULUMI) stack output --stack "$(PULUMI_STACK)"

destroy:
	cd infra && $(PULUMI) destroy --stack "$(PULUMI_STACK)"

clean:
	rm -f $(BINARY)
