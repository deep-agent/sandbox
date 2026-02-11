package sandbox

import (
	"testing"
)

const (
	testBaseURL = "http://localhost:8080"
	testSecret  = "sandbox-secret"
)

func newTestClient() *Client {
	return NewClient(testBaseURL, WithSecret(testSecret))
}

func TestNewClient(t *testing.T) {
	client := NewClient(testBaseURL)
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.baseURL != testBaseURL {
		t.Errorf("expected baseURL %s, got %s", testBaseURL, client.baseURL)
	}
}

func TestNewClientWithSecret(t *testing.T) {
	client := NewClient(testBaseURL, WithSecret(testSecret))
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.tokenProvider == nil {
		t.Fatal("expected tokenProvider to be non-nil")
	}

	token, err := client.tokenProvider()
	if err != nil {
		t.Fatalf("tokenProvider error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestSignToken(t *testing.T) {
	secret := []byte(testSecret)
	token, err := signToken(secret)
	if err != nil {
		t.Fatalf("signToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGetContext(t *testing.T) {
	client := newTestClient()
	ctx, err := client.GetContext()
	if err != nil {
		t.Fatalf("GetContext error: %v", err)
	}
	if ctx == nil {
		t.Fatal("expected context to be non-nil")
	}
	if ctx.Workspace == "" {
		t.Error("expected workspace to be non-empty")
	}
	if ctx.OS == "" {
		t.Error("expected os to be non-empty")
	}
	t.Logf("Context: home=%s, workspace=%s, os=%s, arch=%s", ctx.HomeDir, ctx.Workspace, ctx.OS, ctx.Arch)
}
