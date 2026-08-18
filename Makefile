.PHONY: build_frontend
SHELL := /bin/bash

IMAGE_REPOSITORY ?= "samrreynolds4/"
DOCKERFILES ?= "./dockerfiles"
IMAGE_TAG := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null | sed 's|remotes/origin/||' || echo "unknown")
IMAGE_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo $(IMAGE_TAG))

build_prod: build_base_prod

build_dev: build_base_dev

build_base_prod:
	docker buildx build --no-cache --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}base:${IMAGE_TAG} -f ${DOCKERFILES}/prod/base.dockerfile .

build_base_dev:
	docker buildx build --no-cache --platform linux/amd64,linux/arm64/v8 -t ${IMAGE_REPOSITORY}base:${IMAGE_TAG} -f ${DOCKERFILES}/dev/base.dockerfile .

ci_build_base:
	docker build -t ${IMAGE_REPOSITORY}base:${IMAGE_TAG} -f ${DOCKERFILES}/prod/base.dockerfile .

push:
	docker push ${IMAGE_REPOSITORY}base:${IMAGE_TAG}
test:
	go test ./...

update_schemas:
	@for api in authentication permissions resources roles users; do \
		echo "  copying -> v1/$$api/api/schemas.yaml"; \
		cp v1/schemas.yaml v1/$$api/api/docs/schemas.yaml; \
	done

sqlc_generate:
	@for api in authentication permissions resources roles users; do \
		echo "  generating -> internal/$$api/v1/datagen"; \
		sqlc generate -f internal/$${api}v1/datagen/sqlc.yaml; \
	done