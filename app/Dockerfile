FROM golang:1.17-alpine3.15 AS build
LABEL authors="Sergey Molchanov <sempost@gmail.com>"

ENV GO_DIR="/go/src/github.com/mr-yoyo/anti_bruteforce/app"

COPY ./ $GO_DIR
WORKDIR $GO_DIR

RUN set -ex && \
    apk add --update --no-cache git make && \
    make build

FROM alpine AS app

COPY --from=build /go/src/github.com/mr-yoyo/anti_bruteforce/app/bin /project

CMD ["/project/app", "http-server", "-c", "/project/config.yml"]
