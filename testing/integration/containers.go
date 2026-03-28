//go:build integration

// Package integration provides testcontainer-based infrastructure for
// integration tests that require live PostgreSQL, OpenSearch, MinIO, and Redis.
package integration

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	minioclient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Infra holds running testcontainer instances and their connection details.
type Infra struct {
	PostgresDSN    string
	OpenSearchAddr string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	RedisAddr      string
	OCRAddr        string
	ConvertAddr    string

	containers []testcontainers.Container
}

// StartInfra spins up all required containers and returns their connection details.
func StartInfra(ctx context.Context) (*Infra, error) {
	infra := &Infra{}

	// PostgreSQL
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("argus_test"),
		postgres.WithUsername("argus"),
		postgres.WithPassword("argus"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("starting postgres: %w", err)
	}
	infra.containers = append(infra.containers, pg)

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("getting postgres DSN: %w", err)
	}
	infra.PostgresDSN = dsn

	// OpenSearch — use generic container to avoid rlimit issues on rootless Docker.
	osReq := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.19.1",
		ExposedPorts: []string{"9200/tcp"},
		Env: map[string]string{
			"discovery.type":          "single-node",
			"DISABLE_SECURITY_PLUGIN": "true",
			"OPENSEARCH_JAVA_OPTS":    "-Xms256m -Xmx256m",
			"bootstrap.memory_lock":   "false",
		},
		WaitingFor: wait.ForHTTP("/_cluster/health").
			WithPort("9200/tcp").
			WithStartupTimeout(120 * time.Second),
	}
	osContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: osReq,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("starting opensearch: %w", err)
	}
	infra.containers = append(infra.containers, osContainer)

	osHost, err := osContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting opensearch host: %w", err)
	}
	osPort, err := osContainer.MappedPort(ctx, "9200/tcp")
	if err != nil {
		return nil, fmt.Errorf("getting opensearch port: %w", err)
	}
	infra.OpenSearchAddr = fmt.Sprintf("http://%s:%s", osHost, osPort.Port())

	// MinIO
	mn, err := minio.Run(ctx, "minio/minio:latest",
		minio.WithUsername("argus"),
		minio.WithPassword("argustest"),
	)
	if err != nil {
		return nil, fmt.Errorf("starting minio: %w", err)
	}
	infra.containers = append(infra.containers, mn)

	mnEndpoint, err := mn.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting minio endpoint: %w", err)
	}
	infra.MinIOEndpoint = mnEndpoint
	infra.MinIOAccessKey = "argus"
	infra.MinIOSecretKey = "argustest"

	// Redis
	rd, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("starting redis: %w", err)
	}
	infra.containers = append(infra.containers, rd)

	rdEndpoint, err := rd.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting redis endpoint: %w", err)
	}
	infra.RedisAddr = rdEndpoint

	// OCR sidecar — build from project root context.
	projectRoot := projectRootDir()
	ocrContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    projectRoot,
				Dockerfile: "tools/ocr/Dockerfile",
			},
			ExposedPorts: []string{"50051/tcp"},
			WaitingFor: wait.ForLog("listening on :50051").
				WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("starting ocr: %w", err)
	}
	infra.containers = append(infra.containers, ocrContainer)

	ocrHost, err := ocrContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting ocr host: %w", err)
	}
	ocrPort, err := ocrContainer.MappedPort(ctx, "50051/tcp")
	if err != nil {
		return nil, fmt.Errorf("getting ocr port: %w", err)
	}
	infra.OCRAddr = fmt.Sprintf("%s:%s", ocrHost, ocrPort.Port())

	// Convert sidecar (LibreOffice) — build from project root context.
	convertContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    projectRoot,
				Dockerfile: "tools/convert/Dockerfile",
			},
			ExposedPorts: []string{"50052/tcp"},
			WaitingFor: wait.ForLog("listening on :50052").
				WithStartupTimeout(180 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("starting convert: %w", err)
	}
	infra.containers = append(infra.containers, convertContainer)

	convertHost, err := convertContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting convert host: %w", err)
	}
	convertPort, err := convertContainer.MappedPort(ctx, "50052/tcp")
	if err != nil {
		return nil, fmt.Errorf("getting convert port: %w", err)
	}
	infra.ConvertAddr = fmt.Sprintf("%s:%s", convertHost, convertPort.Port())

	return infra, nil
}

// projectRootDir returns the absolute path to the project root.
func projectRootDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// UploadObject uploads content to MinIO in the test bucket.
func (i *Infra) UploadObject(ctx context.Context, objectKey string, content []byte) error {
	client, err := minioclient.New(i.MinIOEndpoint, &minioclient.Options{
		Creds:  credentials.NewStaticV4(i.MinIOAccessKey, i.MinIOSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("creating minio client: %w", err)
	}

	bucket := "argus-test"
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("checking bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minioclient.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("creating bucket: %w", err)
		}
	}

	_, err = client.PutObject(ctx, bucket, objectKey, bytes.NewReader(content), int64(len(content)), minioclient.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("uploading object: %w", err)
	}
	return nil
}

// Teardown stops and removes all containers.
func (i *Infra) Teardown(ctx context.Context) {
	for j := len(i.containers) - 1; j >= 0; j-- {
		_ = i.containers[j].Terminate(ctx)
	}
}
