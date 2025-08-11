FROM golang:latest

COPY scripts/go.mod go.mod
COPY scripts/go.sum go.sum

RUN go mod download

ADD scripts/pipeline.go pipeline.go
ADD resources resources

CMD [ "go", "run", "./pipeline.go" ]
ENTRYPOINT []
