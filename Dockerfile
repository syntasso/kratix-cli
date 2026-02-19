# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
  ./scripts/build_binary

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
RUN adduser -D -H -u 65532 nonroot

COPY --from=builder /src/bin/component-to-crd /usr/local/bin/component-to-crd

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/component-to-crd"]
