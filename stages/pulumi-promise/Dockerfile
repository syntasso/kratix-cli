FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
ARG TARGETARCH
ARG TARGETOS
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
COPY stages/pulumi-promise/main.go main.go
RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o from-api-to-pulumi-pko-program main.go

FROM --platform=${TARGETARCH:-$BUILDPLATFORM} gcr.io/distroless/cc:nonroot
WORKDIR /
COPY --from=builder /workspace/from-api-to-pulumi-pko-program .
USER 65532:65532
ENTRYPOINT ["/from-api-to-pulumi-pko-program"]
