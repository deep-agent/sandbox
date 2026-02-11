package sandbox

import (
	"testing"

	"github.com/deep-agent/sandbox/model"
)

func TestBashExec(t *testing.T) {
	client := newTestClient()

	result, err := client.BashExec(&model.BashExecRequest{
		Command: "echo hello world",
	})
	if err != nil {
		t.Fatalf("BashExec error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Output != "hello world\n" {
		t.Errorf("expected output 'hello world\\n', got %q", result.Output)
	}
	t.Logf("BashExec output: %s", result.Output)
}

func TestBashExecWithTimeout(t *testing.T) {
	client := newTestClient()

	result, err := client.BashExec(&model.BashExecRequest{
		Command:   "sleep 0.1 && echo done",
		TimeoutMS: 5000,
	})
	if err != nil {
		t.Fatalf("BashExec error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	t.Logf("BashExec output: %s", result.Output)
}

func TestBashExecNonZeroExit(t *testing.T) {
	client := newTestClient()

	result, err := client.BashExec(&model.BashExecRequest{
		Command: "exit 1",
	})
	if err != nil {
		t.Fatalf("BashExec error: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
}
