FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /ocr-server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    grpc-health-probe

COPY --from=builder /ocr-server /usr/local/bin/ocr-server

EXPOSE 50051
CMD ["ocr-server"]
