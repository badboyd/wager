ARG GO_VERSION=1.14

FROM golang:$GO_VERSION-alpine AS build

RUN apk add --no-cache \
    gcc \
    git \
    musl-dev \
    nodejs \
    openssl \
    postgresql-client \
    yarn

WORKDIR /wager

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN go install -v ./cmd

FROM alpine:3.10

RUN apk add --no-cache \
    ca-certificates \
    openssl \
    postgresql-client

RUN mkdir -p /etc/wager && \
    mkdir -p /var/opt/wager

COPY --from=build /go/bin/wager /
COPY config/*.yml /etc/wager/config/

EXPOSE 8080

CMD ["./wager"]