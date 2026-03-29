//go:build integration

package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

var testInfra *Infra

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("starting test infrastructure...")
	var err error
	testInfra, err = StartInfra(ctx)
	if err != nil {
		log.Fatalf("failed to start infrastructure: %v", err)
	}

	log.Println("running migrations...")
	if err := RunMigrations(testInfra.PostgresDSN); err != nil {
		testInfra.Teardown(ctx)
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("creating opensearch indices...")
	if err := CreateOpenSearchIndex(ctx, testInfra.OpenSearchAddr); err != nil {
		testInfra.Teardown(ctx)
		log.Fatalf("failed to create opensearch index: %v", err)
	}
	if err := CreateNotificationsIndex(ctx, testInfra.OpenSearchAddr); err != nil {
		testInfra.Teardown(ctx)
		log.Fatalf("failed to create notifications index: %v", err)
	}

	log.Println("initializing stores...")
	if err := InitStores(ctx, testInfra); err != nil {
		testInfra.Teardown(ctx)
		log.Fatalf("failed to initialize stores: %v", err)
	}

	log.Println("initializing engines...")
	if err := InitEngines(); err != nil {
		testInfra.Teardown(ctx)
		log.Fatalf("failed to initialize engines: %v", err)
	}

	log.Println("infrastructure ready")
	fmt.Printf("  postgres: %s\n", testInfra.PostgresDSN)
	fmt.Printf("  opensearch: %s\n", testInfra.OpenSearchAddr)
	fmt.Printf("  minio: %s\n", testInfra.MinIOEndpoint)
	fmt.Printf("  redis: %s\n", testInfra.RedisAddr)

	code := m.Run()

	log.Println("tearing down infrastructure...")
	testInfra.Teardown(context.Background())

	os.Exit(code)
}

// TestInfraReady is a smoke test that verifies all containers started.
func TestInfraReady(t *testing.T) {
	if testInfra.PostgresDSN == "" {
		t.Fatal("postgres not available")
	}
	if testInfra.OpenSearchAddr == "" {
		t.Fatal("opensearch not available")
	}
	if testInfra.MinIOEndpoint == "" {
		t.Fatal("minio not available")
	}
	if testInfra.RedisAddr == "" {
		t.Fatal("redis not available")
	}
}
