GO_EXECUTABLE ?= go
BUILD_VERSION ?= $(shell git describe --tags)
SCRIPT_DIR ?= /opt/deploy-tools/tools
BUILD_TIME = `date +%FT%T%z`
BUILD_NAME = log-colorizer
MAIN_FILE = log-colorizer.go

export PATH := $(PATH):$(GOPATH)/bin

init:
	${GO_EXECUTABLE} get github.com/Masterminds/glide
	${GOPATH}/bin/glide ${GLIDEOPS} install

build: init
	${GO_EXECUTABLE} build \
	-o build/${BUILD_NAME} \
	-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
	.

run: init
	${GO_EXECUTABLE} run ${MAIN_FILE}

test: init
	${GO_EXECUTABLE} test .*

bootstrap-dist:
	${GO_EXECUTABLE} get github.com/Ak-Army/gox

build-all: init bootstrap-dist
	${GOPATH}/bin/gox -verbose \
	-ldflags="-X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME}" \
	-output="build/${BUILD_VERSION}/${BUILD_NAME}-{{.OS}}-{{.Arch}}" .

build-deb: build
	${SCRIPT_DIR}/go-deb -version ${BUILD_VERSION} config/config.json

.PHONY: init build test build-all deploy build-deb
