.PHONY: build_frontend
SHELL := /bin/bash

IMAGE_REPOSITORY ?= "samrreynolds4/"
DOCKERFILES ?= "./dockerfiles"
IMAGE_TAG := $(shell git name-rev --name-only HEAD 2>/dev/null || echo "unknown")
IMAGE_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo $(IMAGE_TAG))

build_all:
	make build_base

build_base:
	docker buildx build --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}base:${IMAGE_TAG} -f ${DOCKERFILES}/base.dockerfile .

ci_build_base:
	docker build -t ${IMAGE_REPOSITORY}base:${IMAGE_TAG} -f ${DOCKERFILES}/base.dockerfile .

push_base:
	docker push ${IMAGE_REPOSITORY}base:${IMAGE_TAG}
test:
	go test ./...