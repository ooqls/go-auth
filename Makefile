.PHONY: build_frontend
SHELL := /bin/bash

IMAGE_REPOSITORY ?= "samrreynolds4/"
DOCKERFILES ?= "./dockerfiles"
IMAGE_TAG := $(shell git name-rev --name-only HEAD 2>/dev/null || echo "unknown")
IMAGE_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo $(IMAGE_TAG))

build_auth:
	docker buildx build --platform linux/amd64 -t ${IMAGE_REPOSITORY}auth:${IMAGE_TAG} -f ${DOCKERFILES}/auth.dockerfile .

push_auth:
	echo "$$DOCKER_REGISTRY_TOKEN" | docker login -u samrreynolds4 --password-stdin
	docker push ${IMAGE_REPOSITORY}auth:${IMAGE_TAG}