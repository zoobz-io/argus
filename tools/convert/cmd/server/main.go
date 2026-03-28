// Package main runs the document conversion gRPC sidecar that wraps LibreOffice.
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
	"syscall"

	"github.com/zoobz-io/argus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	port := os.Getenv("CONVERT_PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	proto.RegisterConvertServiceServer(srv, &convertServer{})

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

type convertServer struct {
	proto.UnimplementedConvertServiceServer
}

func (s *convertServer) ConvertDocument(_ context.Context, req *proto.ConvertRequest) (*proto.ConvertResponse, error) {
	if len(req.Document) == 0 {
		return nil, fmt.Errorf("empty document")
	}

	inputExt := mimeToInputExt(req.MimeType)
	if inputExt == "" {
		return nil, fmt.Errorf("unsupported input MIME type: %s", req.MimeType)
	}

	outputFormat := mimeToOutputFormat(req.MimeType)
	outputMime := mimeToOutputMime(req.MimeType)

	// Write input to temp file.
	tmpDir, err := os.MkdirTemp("", "convert-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input"+inputExt)
	if err := os.WriteFile(inputPath, req.Document, 0600); err != nil {
		return nil, fmt.Errorf("writing temp file: %w", err)
	}

	// Run LibreOffice headless conversion.
	cmd := exec.Command("soffice",
		"--headless",
		"--norestore",
		"--convert-to", outputFormat,
		"--outdir", tmpDir,
		inputPath,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("libreoffice conversion failed: %w", err)
	}

	// Find the output file.
	outputExt := formatToExt(outputFormat)
	outputPath := filepath.Join(tmpDir, "input"+outputExt)
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("reading converted file: %w", err)
	}

	return &proto.ConvertResponse{
		Document: data,
		MimeType: outputMime,
	}, nil
}

func mimeToInputExt(mime string) string {
	switch mime {
	case "application/msword":
		return ".doc"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	default:
		return ""
	}
}

func mimeToOutputFormat(mime string) string {
	switch mime {
	case "application/msword":
		return "docx"
	case "application/vnd.ms-excel":
		return "xlsx"
	case "application/vnd.ms-powerpoint":
		return "pptx"
	default:
		return ""
	}
}

func mimeToOutputMime(mime string) string {
	switch mime {
	case "application/msword":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "application/vnd.ms-excel":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "application/vnd.ms-powerpoint":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return ""
	}
}

func formatToExt(format string) string {
	return "." + format
}
