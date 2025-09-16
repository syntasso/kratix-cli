FROM --platform=$TARGETPLATFORM golang:1.24 AS builder
ARG TARGETARCH
ARG TARGETOS
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
COPY stages/operator-promise/main.go main.go
RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o from-api-to-operator main.go

FROM gcr.io/distroless/cc:nonroot
WORKDIR /
COPY --from=builder /workspace/from-api-to-operator .
USER 65532:65532
ENTRYPOINT ["/from-api-to-operator"]
