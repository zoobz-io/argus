package boot

import (
	"context"
	"fmt"
	"log"

	"github.com/zoobz-io/sum"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zoobz-io/argus/config"
	"github.com/zoobz-io/argus/proto"
)

// OCR creates a gRPC connection and OCR service client from config.
// Caller must defer conn.Close().
func OCR(ctx context.Context) (*grpc.ClientConn, proto.OCRServiceClient, error) {
	cfg := sum.MustUse[config.OCR](ctx)
	conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to ocr service: %w", err)
	}
	log.Println("ocr service connected")
	return conn, proto.NewOCRServiceClient(conn), nil
}

// Convert creates a gRPC connection and Convert service client from config.
// Caller must defer conn.Close().
func Convert(ctx context.Context) (*grpc.ClientConn, proto.ConvertServiceClient, error) {
	cfg := sum.MustUse[config.Convert](ctx)
	conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to convert service: %w", err)
	}
	log.Println("convert service connected")
	return conn, proto.NewConvertServiceClient(conn), nil
}

// Classify creates a gRPC connection and Classify service client from config.
// Caller must defer conn.Close().
func Classify(ctx context.Context) (*grpc.ClientConn, proto.ClassifyServiceClient, error) {
	cfg := sum.MustUse[config.Classify](ctx)
	conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to classify service: %w", err)
	}
	log.Println("classify service connected")
	return conn, proto.NewClassifyServiceClient(conn), nil
}
