.PHONY: build_frontend
SHELL := /bin/bash

IMAGE_REPOSITORY ?= "samrreynolds4/"
DOCKERFILES ?= "./dockerfiles"
IMAGE_TAG := $(shell git name-rev --name-only HEAD 2>/dev/null || echo "unknown")
IMAGE_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo $(IMAGE_TAG))

build_all:
	make build_base
	make build_seed
	make build_auth
	make build_roles
	make build_users

build_base:
	docker buildx build --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}base:${IMAGE_TAG} -f ${DOCKERFILES}/base.dockerfile .
push_base:
	docker push ${IMAGE_REPOSITORY}base:${IMAGE_TAG}
build_seed: build_base
	docker buildx build --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}seed:${IMAGE_TAG} -f ${DOCKERFILES}/seed.dockerfile .
build_auth: build_base
	docker buildx build --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}auth:${IMAGE_TAG} -f ${DOCKERFILES}/auth.dockerfile .
build_users: build_base
	docker buildx build --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}users:${IMAGE_TAG} -f ${DOCKERFILES}/users.dockerfile .

build_roles: build_base
	docker buildx build --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}roles:${IMAGE_TAG} -f ${DOCKERFILES}/roles.dockerfile .

push_users:
	docker push ${IMAGE_REPOSITORY}users:${IMAGE_TAG}

push_auth:
	docker push ${IMAGE_REPOSITORY}auth:${IMAGE_TAG}

push_roles:
	docker push ${IMAGE_REPOSITORY}roles:${IMAGE_TAG}

push_seed:
	docker push ${IMAGE_REPOSITORY}seed:${IMAGE_TAG}

push_all:
	make push_base
	make push_auth
	make push_roles
	make push_users

test_auth:
	go test ./...