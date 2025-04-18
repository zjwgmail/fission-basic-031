VERSION=v1
GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...
consumer:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/server ./cmd/consumer
gw:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/server ./cmd/gw
static:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/server ./cmd/static
job:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/server ./cmd/job

docker_consumer:
	docker build --no-cache -t prod-mlbb25031-consumer-server:$(VERSION) --build-arg TARGET="consumer" .
docker_job:
	docker build --no-cache -t prod-mlbb25031-job-server:$(VERSION) --build-arg TARGET="job" .
docker_gw:
	docker build --no-cache -t prod-mlbb25031-gateway-server:$(VERSION) --build-arg TARGET="gw" .
docker_static:
	docker build --no-cache -t prod-mlbb25031-static-server:$(VERSION) --build-arg TARGET="static" .
docker_consumer_test:
	docker build --no-cache -t prod-mlbb25031-consumer-server:$(VERSION) --build-arg TARGET="consumer" --build-arg CONFIGS="test" .
docker_job_test:
	docker build --no-cache -t prod-mlbb25031-job-server:$(VERSION) --build-arg TARGET="job" --build-arg CONFIGS="test" .
docker_gw_test:
	docker build --no-cache -t prod-mlbb25031-gateway-server:$(VERSION) --build-arg TARGET="gw" --build-arg CONFIGS="test" .
docker_static_test:
	docker build --no-cache -t prod-mlbb25031-static-server:$(VERSION) --build-arg TARGET="static" --build-arg CONFIGS="test" .


.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	make api;
	make config;
	make generate;

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
