#EJ2
FROM golang:1.17 AS builder

# First stage: compile golang binary
RUN mkdir -p /build
WORKDIR /build/
COPY . .

# CGO_ENABLED must be disabled to run go binary in Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -o bin/client github.com/7574-sistemas-distribuidos/docker-compose-init/client

# Second stage: create lightweight deployment image
FROM busybox:latest

# Copy the compiled binary from the builder stage
COPY --from=builder /build/bin/client /client

# Use /config.yaml from volume mount instead of copying into the image
ENTRYPOINT ["/client"]

