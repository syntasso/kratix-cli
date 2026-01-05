FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
ARG TARGETARCH
ARG TARGETOS
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY stages/terraform-module-promise/main.go main.go
COPY internal internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o from-api-to-terraform-module main.go

FROM --platform=${TARGETARCH:-$BUILDPLATFORM} gcr.io/distroless/cc:nonroot
WORKDIR /
COPY --from=builder /workspace/from-api-to-terraform-module .
USER 65532:65532
ENTRYPOINT ["/from-api-to-terraform-module"]
