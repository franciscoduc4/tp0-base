FROM golang:1.17 AS builder

RUN mkdir -p /build
WORKDIR /build/
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -o bin/client github.com/7574-sistemas-distribuidos/docker-compose-init/client

FROM alpine:latest

# Instalo nc (netcat) en el cliente
RUN apk add --no-cache netcat-openbsd

COPY --from=builder /build/bin/client /client

ENTRYPOINT ["/client"]
