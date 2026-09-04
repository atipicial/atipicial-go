# Builder image
# Keep go version in sync with Build GA job.
FROM golang:1.27-alpine AS builder

# Display go version for information purposes.
RUN go version

RUN set -x \
    && apk add --no-cache git make \
    && mkdir -p /tmp

COPY . /atipicial-go

WORKDIR /atipicial-go

ARG REPO=repository
ARG VERSION=dev

RUN VERSION=$VERSION REPO=$REPO make build

# Executable image
FROM alpine

ARG VERSION=dev
LABEL version=$VERSION

WORKDIR /

COPY --from=builder /atipicial-go/config /config
COPY --from=builder /atipicial-go/.docker/privnet-entrypoint.sh /usr/bin/privnet-entrypoint.sh
COPY --from=builder /atipicial-go/bin/atipicial-go /usr/bin/atipicial-go
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/usr/bin/privnet-entrypoint.sh"]

CMD ["node", "--config-path", "/config", "--privnet"]
