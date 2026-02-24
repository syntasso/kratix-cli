FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
ARG TARGETARCH
ARG TARGETOS
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN --mount=type=cache,id=go-mod-cache,target=/go/pkg/mod,sharing=locked \
  --mount=type=cache,id=go-build-cache,target=/root/.cache/go-build,sharing=locked \
  go mod download
COPY stages/pulumi-promise/main.go main.go
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -o from-api-to-pulumi-pko-program main.go

FROM --platform=${TARGETARCH:-$BUILDPLATFORM} gcr.io/distroless/cc:nonroot
WORKDIR /
COPY --from=builder /workspace/from-api-to-pulumi-pko-program .
USER 65532:65532
ENTRYPOINT ["/from-api-to-pulumi-pko-program"]
