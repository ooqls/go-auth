.PHONY: build_frontend
SHELL := /bin/bash

IMAGE_REPOSITORY ?= "samrreynolds4/"
DOCKERFILES ?= "./dockerfiles"
IMAGE_TAG := $(shell git name-rev --name-only HEAD 2>/dev/null || echo "unknown")
IMAGE_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo $(IMAGE_TAG))

PROTO_PATH := proto
GOOGLEAPIS := /tmp/googleapis
PROTO_INCLUDE := /usr/local/include
PROTO_OUT := gen/grpc

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

# --- Protobuf / gRPC code generation ---
.PHONY: proto proto_common proto_auth proto_users proto_roles proto_permissions proto_resources proto_rolebindings

proto: proto_common proto_auth proto_users proto_roles proto_permissions proto_resources proto_rolebindings
	@echo "All protobuf code generated."

proto_common:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		proto/common/v1/common.proto

proto_auth:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		proto/auth/v1/auth.proto

proto_users:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		proto/users/v1/users.proto

proto_roles:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		proto/roles/v1/roles.proto

proto_permissions:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		proto/permissions/v1/permissions.proto

proto_resources:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		proto/resources/v1/resources.proto

proto_rolebindings:
	protoc -I $(PROTO_PATH) -I $(GOOGLEAPIS) -I $(PROTO_INCLUDE) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		proto/rolebindings/v1/rolebindings.proto