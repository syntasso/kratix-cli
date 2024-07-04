FROM alpine:latest

RUN apk add --no-cache yq

COPY pipeline.sh /pipeline.sh

ENTRYPOINT ["/pipeline.sh"]