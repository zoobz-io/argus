// Package main runs the OCR gRPC sidecar that wraps Tesseract.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/zoobz-io/argus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	port := os.Getenv("OCR_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	proto.RegisterOCRServiceServer(srv, &ocrServer{})

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hsrv)

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("serve failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	srv.GracefulStop()
}

type ocrServer struct {
	proto.UnimplementedOCRServiceServer
}

func (s *ocrServer) ExtractText(_ context.Context, req *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
	if len(req.Document) == 0 {
		return nil, fmt.Errorf("empty document")
	}

	// Determine file extension from MIME type.
	ext := mimeToExt(req.MimeType)

	// Write document bytes to a temp file.
	tmpDir, err := os.MkdirTemp("", "ocr-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input"+ext)
	if err := os.WriteFile(inputPath, req.Document, 0600); err != nil {
		return nil, fmt.Errorf("writing temp file: %w", err)
	}

	outputBase := filepath.Join(tmpDir, "output")

	// Build tesseract command.
	args := []string{inputPath, outputBase}
	lang := req.Language
	if lang == "" {
		lang = "eng"
	}
	args = append(args, "-l", lang)

	cmd := exec.Command("tesseract", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tesseract failed: %w", err)
	}

	// Read output text.
	text, err := os.ReadFile(outputBase + ".txt")
	if err != nil {
		return nil, fmt.Errorf("reading tesseract output: %w", err)
	}

	return &proto.ExtractTextResponse{
		Text:       strings.TrimSpace(string(text)),
		Confidence: 0.0, // Tesseract CLI doesn't easily expose confidence per-run.
		Pages:      1,
	}, nil
}

func mimeToExt(mime string) string {
	switch mime {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/tiff":
		return ".tiff"
	case "image/bmp":
		return ".bmp"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	default:
		return ".bin"
	}
}
