ARG MTK_VERSION

# build MTK from source
FROM golang:1.18-alpine as builder

WORKDIR /go/src/github.com/skpr
RUN apk add --virtual --update-cache git && \
	rm -rf /tmp/* /var/tmp/* /var/cache/apk/* /var/cache/distfiles/*
RUN git clone https://github.com/skpr/mtk.git && cd mtk && git checkout $MTK_VERSION
WORKDIR /go/src/github.com/skpr/mtk/dump

# compile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/mtk-dump github.com/skpr/mtk/dump

ARG IMAGE_REPO
FROM ${IMAGE_REPO:-uselagoon}/commons as commons

# Put in some labels so people know what this image is for
LABEL org.opencontainers.image.authors="The Lagoon Authors" maintainer="The Lagoon Authors"

COPY --from=builder /go/src/github.com/skpr/mtk/dump/bin/mtk-dump /usr/local/bin/mtk-dump

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
