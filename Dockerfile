############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/vodeno
COPY . .
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/main cmd/main.go
############################
# STEP 2 build a small image
############################
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=builder /go/bin/main /go/bin/main
COPY --from=builder /go/src/vodeno/config/config.yaml config.yaml

ENV CONFIG_PATH=config.yaml
# Run the hello binary.
ENTRYPOINT ["/go/bin/main"]