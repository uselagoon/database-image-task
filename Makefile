VERSION=$(shell echo $(shell git describe --abbrev=0 --tags)+$(shell git rev-parse --short=8 HEAD))
BUILD=$(shell date +%FT%T%z)
GO_VER=1.23
GOOS=linux
GOARCH=amd64
PKG=github.com/uselagoon/database-image-task
LDFLAGS=-w -s -X ${PKG}/cmd.dbitVersion=${VERSION} -X ${PKG}/cmd.dbitBuild=${BUILD} -X "${PKG}/cmd.goVersion=${GO_VER}"

.PHONY: test
test: fmt vet
	go clean -testcache && go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: run
run: fmt vet
	go run ./main.go

.PHONY: build
build: fmt vet
	CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -v .

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --pull --build-arg GO_VER=${GO_VER} --build-arg VERSION=${VERSION} --build-arg BUILD=${BUILD} --rm -f Dockerfile -t lagoon/database-image-task:local .
	docker run --entrypoint /bin/bash lagoon/database-image-task:local -c 'database-image-task version && database-image-task dump'