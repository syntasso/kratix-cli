FROM golang:latest

WORKDIR /scripts

COPY scripts/go.mod scripts/go.sum scripts/pipeline.go ./

RUN go mod download

ADD resources resources

RUN go build -a -o /usr/bin/pipeline.go pipeline.go

CMD [ "sh", "-c", "pipeline.go" ]

ENTRYPOINT []