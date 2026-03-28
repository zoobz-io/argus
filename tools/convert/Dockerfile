FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy proto module first (local dependency).
COPY proto/ /proto/

# Copy convert sidecar source.
COPY tools/convert/go.mod tools/convert/go.sum ./
RUN go mod download
COPY tools/convert/ .
RUN CGO_ENABLED=0 go build -o /convert-server ./cmd/server

# Install grpc-health-probe for container health checks.
RUN go install github.com/grpc-ecosystem/grpc-health-probe@latest

FROM alpine:3.21

RUN apk add --no-cache \
    libreoffice \
    font-noto

COPY --from=builder /convert-server /usr/local/bin/convert-server
COPY --from=builder /go/bin/grpc-health-probe /usr/local/bin/grpc-health-probe

EXPOSE 50052
CMD ["convert-server"]
