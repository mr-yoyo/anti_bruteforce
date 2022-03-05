FROM golang:1.17-alpine3.15 AS build
LABEL authors="Sergey Molchanov <sempost@gmail.com>"

ENV GO_DIR="/go/src/github.com/mr-yoyo/anti_bruteforce/it"

COPY ./ $GO_DIR
WORKDIR $GO_DIR

RUN set -ex && \
    apk add --update --no-cache build-base

CMD ["go", "test", "-v", "./api_test.go"]
