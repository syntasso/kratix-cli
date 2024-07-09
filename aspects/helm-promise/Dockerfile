FROM "alpine"

RUN apk update && apk add --no-cache yq helm

ADD pipeline.sh /usr/bin/pipeline.sh

CMD [ "sh", "-c", "pipeline.sh" ]
ENTRYPOINT []
