FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Copy proto module first (local dependency).
COPY proto/ /proto/

# Copy OCR sidecar source.
COPY tools/ocr/go.mod tools/ocr/go.sum ./
RUN go mod download
COPY tools/ocr/ .
RUN CGO_ENABLED=0 go build -o /ocr-server ./cmd/server

# Install grpc-health-probe for container health checks.
RUN go install github.com/grpc-ecosystem/grpc-health-probe@latest

FROM alpine:3.21

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng

RUN adduser -D -u 1000 sidecar

COPY --from=builder /ocr-server /usr/local/bin/ocr-server
COPY --from=builder /go/bin/grpc-health-probe /usr/local/bin/grpc-health-probe

USER sidecar
EXPOSE 50051
CMD ["ocr-server"]
