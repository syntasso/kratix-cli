FROM --platform=$TARGETPLATFORM golang:1.22 AS builder
ARG TARGETARCH
ARG TARGETOS
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
COPY stages/crossplane-promise/main.go main.go
COPY stages/helm-promise/ stages/helm-promise/
COPY cmd/ cmd/
COPY internal/ internal/
RUN go mod download
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o from-api-to-crossplane-claim main.go

FROM gcr.io/distroless/cc:nonroot
WORKDIR /
COPY --from=builder /workspace/from-api-to-crossplane-claim .
USER 65532:65532
ENTRYPOINT ["/from-api-to-crossplane-claim"]
