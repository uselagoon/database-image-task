ARG MTK_VERSION

# build MTK from source
FROM golang:1.23-alpine as builder

ENV MTK_VERSION=v2.0.2

WORKDIR /go/src/github.com/skpr
RUN apk add --virtual --update-cache git && \
	rm -rf /tmp/* /var/tmp/* /var/cache/apk/* /var/cache/distfiles/*
ADD https://github.com/skpr/mtk.git#$MTK_VERSION ./mtk

WORKDIR /go/src/github.com/skpr/mtk

# compile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -o bin/mtk-dump github.com/skpr/mtk/cmd/mtk

ARG IMAGE_REPO
FROM ${IMAGE_REPO:-uselagoon}/commons as commons

# Put in some labels so people know what this image is for
LABEL org.opencontainers.image.authors="The Lagoon Authors" maintainer="The Lagoon Authors"

COPY --from=builder /go/src/github.com/skpr/mtk/bin/mtk-dump /usr/local/bin/mtk-dump

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

RUN chmod a+x /usr/local/bin/mariadb-image-builder /usr/local/bin/mtk-dump /usr/local/bin/image-builder-entry

# Ensure the syntax is correct bash before actually pushing, etc
RUN bash -n /usr/local/bin/mariadb-image-builder

# Set up what to run
ENTRYPOINT ["/sbin/tini", "--", "/lagoon/entrypoints.bash"]
CMD ["/usr/local/bin/image-builder-entry"]
