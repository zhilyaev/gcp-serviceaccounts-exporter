ARG GOLANG_VERSION=1.21
ARG ALPINE_VERSION=3.18

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS builder
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV PROJECT=gcp-serviceaccounts-exporter

WORKDIR ${PROJECT}

COPY go.mod go.sum ./
RUN go mod download

# Copy src code from the host and compile it
COPY pkg pkg
RUN go build -a -o /${PROJECT} ./main.go

### Base image with shell
FROM alpine:${ALPINE_VERSION} as base-release
RUN apk --update --no-cache add ca-certificates && update-ca-certificates
ENTRYPOINT ["/bin/gcp-serviceaccounts-exporter"]

### Build with goreleaser
FROM base-release as goreleaser
COPY gcp-serviceaccounts-exporter /bin/

### Build in docker
FROM base-release as release
COPY --from=builder /gcp-serviceaccounts-exporter /bin/

### Scratch with build in docker
FROM scratch as scratch-release
COPY --from=base-release /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gcp-serviceaccounts-exporter /bin/
ENTRYPOINT ["/bin/gcp-serviceaccounts-exporter"]
CMD ["run"]
USER 65534

### Scratch with goreleaser
FROM scratch as scratch-goreleaser
COPY --from=base-release /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gcp-serviceaccounts-exporter /bin/
ENTRYPOINT ["/bin/gcp-serviceaccounts-exporter"]
CMD ["run"]
USER 65534
