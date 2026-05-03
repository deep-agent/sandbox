package local

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/deep-agent/sandbox/types/model"
)

func TestClient_MCPTools_ListsRegisteredTools(t *testing.T) {
	ctx := context.Background()
	c := NewClient()
	tools, err := c.MCPTools(ctx)
	if err != nil {
		t.Fatalf("MCPTools failed: %v", err)
	}
	if len(tools) == 0 {
		t.Fatal("expected at least one tool, got 0")
	}

	names := toolNamesForTest(t, ctx, tools)
	for _, want := range []string{"Bash", "Read", "Write", "Edit", "Glob", "Grep"} {
		if _, ok := names[want]; !ok {
			t.Errorf("expected tool %q, missing from %v", want, keysOf(names))
		}
	}
}

func TestClient_MCPTools_CwdFromOption(t *testing.T) {
	tmp := t.TempDir()
	markerName := "cwd_marker.txt"
	if err := os.WriteFile(filepath.Join(tmp, markerName), []byte("hello"), 0o600); err != nil {
		t.Fatalf("seed marker: %v", err)
	}

	ctx := context.Background()
	c := NewClient(WithCwd(tmp))
	tools, err := c.MCPTools(ctx)
	if err != nil {
		t.Fatalf("MCPTools failed: %v", err)
	}

	bash := findInvokableForTest(t, ctx, tools, "Bash")
	args := `{"command":"pwd && ls cwd_marker.txt","run_in_background":false}`
	out, err := bash.InvokableRun(ctx, args)
	if err != nil {
		t.Fatalf("Bash InvokableRun failed: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	if !strings.Contains(out, resolved) {
		t.Errorf("expected output to contain cwd %q, got %q", resolved, out)
	}
	if !strings.Contains(out, markerName) {
		t.Errorf("expected output to contain %q, got %q", markerName, out)
	}
}

func TestClient_MCPTools_CwdFromSandboxContext(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "probe.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("seed probe: %v", err)
	}

	ctx := context.Background()
	c := NewClient(WithSandboxContext(&model.SandboxContext{HomeDir: tmp}))
	tools, err := c.MCPTools(ctx)
	if err != nil {
		t.Fatalf("MCPTools failed: %v", err)
	}

	bash := findInvokableForTest(t, ctx, tools, "Bash")
	out, err := bash.InvokableRun(ctx, `{"command":"pwd","run_in_background":false}`)
	if err != nil {
		t.Fatalf("Bash InvokableRun failed: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	if !strings.Contains(out, resolved) {
		t.Errorf("expected cwd %q from sandboxCtx.HomeDir, got %q", resolved, out)
	}
}

func toolNamesForTest(t *testing.T, ctx context.Context, tools []tool.BaseTool) map[string]struct{} {
	t.Helper()
	names := make(map[string]struct{}, len(tools))
	for _, tl := range tools {
		info, err := tl.Info(ctx)
		if err != nil {
			t.Fatalf("tool Info failed: %v", err)
		}
		names[info.Name] = struct{}{}
	}
	return names
}

func findInvokableForTest(t *testing.T, ctx context.Context, tools []tool.BaseTool, name string) tool.InvokableTool {
	t.Helper()
	for _, tl := range tools {
		info, err := tl.Info(ctx)
		if err != nil {
			t.Fatalf("tool Info failed: %v", err)
		}
		if info.Name != name {
			continue
		}
		inv, ok := tl.(tool.InvokableTool)
		if !ok {
			t.Fatalf("tool %q is not invokable", name)
		}
		return inv
	}
	t.Fatalf("tool %q not found", name)
	return nil
}

func keysOf(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
