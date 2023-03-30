FROM golang:1.20

# Checkout, build and install IPFS Kubo
RUN git clone https://github.com/ipfs/kubo.git
WORKDIR kubo
RUN git checkout v0.19.0
RUN make install
RUN ipfs --version

# Initialize IPFS
RUN ipfs init

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