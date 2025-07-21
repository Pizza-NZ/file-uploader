//go:build integration

package main_test

import (
	"net/http"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	time.Sleep(1 * time.Second) // Wait for services to be up

	t.Run("Health checks", func(t *testing.T) {
		goServiceHealth(t)
		nginxHealth(t)
	})

	t.Run("Endpoint integration", func(t *testing.T) {
		uploadEndpoint(t)
	})
}

func goServiceHealth(t *testing.T) {
	resp, err := http.Get("http://go-service:2131/health")
	if err != nil {
		t.Fatalf("Failed to request go-service health check: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("go-service health check failed: expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func nginxHealth(t *testing.T) {
	resp, err := http.Get("http://nginx:80")
	if err != nil {
		t.Fatalf("Failed to request nginx health check: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("nginx health check failed: expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func uploadEndpoint(t *testing.T) {
	resp, err := http.Post("http://nginx:80/upload", "", nil)
	if err != nil {
		t.Fatalf("Failed to request upload endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest { // Expecting bad request as no file is sent
		t.Errorf("upload endpoint test failed: expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
