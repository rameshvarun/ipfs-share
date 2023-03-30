FROM golang:1.20

WORKDIR /usr/src/app

# Fetch and build dependencies.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build application source.
COPY server.go .
COPY public/ ./public
COPY templates ./templates
RUN go build -v -o /usr/local/bin/app ./...

# Run application.
CMD app