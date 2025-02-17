ARG UPSTREAM_REPO
ARG UPSTREAM_TAG
ARG GO_VER
ARG IMAGE_REPO

FROM golang:${GO_VER:-1.23}-alpine3.20 AS golang

# build MTK
ARG MTK_VERSION
ENV MTK_VERSION=v2.1.0

WORKDIR /go/src/github.com/skpr
RUN apk add --virtual --update-cache git && \
	rm -rf /tmp/* /var/tmp/* /var/cache/apk/* /var/cache/distfiles/*
ADD https://github.com/skpr/mtk.git#$MTK_VERSION ./mtk

WORKDIR /go/src/github.com/skpr/mtk

# compile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -o bin/mtk-dump github.com/skpr/mtk/cmd/mtk

# build database-image-task
WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum

COPY main.go main.go
COPY cmd/ cmd/
COPY internal/ internal/

ARG BUILD
ARG GO_VER
ARG VERSION
ENV BUILD=${BUILD} \
    GO_VER=${GO_VER} \
    VERSION=${VERSION}

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Do not force rebuild of up-to-date packages (do not use -a) and use the compiler cache folder
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build \
    -ldflags="-s -w \
    -X github.com/uselagoon/database-image-task/cmd.dbitBuild=${BUILD} \
    -X github.com/uselagoon/database-image-task/cmd.goVersion=${GO_VER} \
    -X github.com/uselagoon/database-image-task/cmd.dbitVersion=${VERSION} \
    -extldflags '-static'" \
    -o /app/database-image-task .

FROM ${IMAGE_REPO:-uselagoon}/commons AS commons

# Put in some labels so people know what this image is for
LABEL org.opencontainers.image.authors="The Lagoon Authors" maintainer="The Lagoon Authors"

COPY --from=golang /go/src/github.com/skpr/mtk/bin/mtk-dump /usr/local/bin/mtk-dump
COPY --from=golang /app/database-image-task /usr/local/bin/database-image-task

# Install necessary packages
# -	perl for docker-login
# -	bash for image-builder
RUN apk add --virtual --update-cache perl bash docker-cli jq && \
	rm -rf /tmp/* /var/tmp/* /var/cache/apk/* /var/cache/distfiles/*

# Put in docker credentials so we can do docker pushes
RUN mkdir $HOME/.docker

# Put in config we're going to use for mtk
# COPY etc /usr/local/etc/dsb

# Put in needed scripts (in reverse order of mutability
COPY image-builder-entry /usr/local/bin/image-builder-entry
COPY mariadb-image-builder /usr/local/bin/mariadb-image-builder

WORKDIR /builder

COPY builder/mariadb.Dockerfile /builder/mariadb.Dockerfile

RUN chmod a+x /usr/local/bin/mariadb-image-builder /usr/local/bin/mtk-dump /usr/local/bin/image-builder-entry /usr/local/bin/database-image-task

# Ensure the syntax is correct bash before actually pushing, etc
RUN bash -n /usr/local/bin/mariadb-image-builder

# Set up what to run
ENTRYPOINT ["/sbin/tini", "--", "/lagoon/entrypoints.bash"]
CMD ["/usr/local/bin/image-builder-entry"]
