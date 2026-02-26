package bash

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/deep-agent/sandbox/pkg/safe"
)

type TruncateOptions struct {
	MaxLines int
	MaxBytes int
}

type StreamChunk struct {
	Data   string `json:"data"`
	Source string `json:"source"`
}

type StreamCallback func(chunk StreamChunk)

type Executor struct {
	timeout time.Duration
}

type ExecResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	TimedOut   bool   `json:"timed_out"`
	Truncated  bool   `json:"truncated"`
	Metadata   string `json:"metadata,omitempty"`
	OutputFile string `json:"output_file,omitempty"`
}

func NewExecutor() *Executor {
	return &Executor{
		timeout: 30 * time.Second,
	}
}

func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

func (e *Executor) ExecuteBackground(ctx context.Context, command string, workDir string, timeout time.Duration) (*ExecResult, error) {
	startTime := time.Now()

	outputDir := filepath.Join(workDir, ".logs", "background_outputs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	outputFile := filepath.Join(outputDir, fmt.Sprintf("bg_%d.log", startTime.UnixNano()))

	wrappedCommand := fmt.Sprintf("script -q -f -c %q %q", command, outputFile)
	cmd := exec.Command("bash", "-c", wrappedCommand)
	cmd.Dir = workDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		os.Remove(outputFile)
		return nil, fmt.Errorf("failed to start background command: %w", err)
	}

	safe.Go(func() {
		done := make(chan struct{})
		go func() {
			cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(timeout):
			if cmd.Process != nil {
				syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			}
			<-done
		}
	})

	return &ExecResult{
		Output:     fmt.Sprintf("Command started in background (PID: %d)", cmd.Process.Pid),
		ExitCode:   0,
		DurationMs: time.Since(startTime).Milliseconds(),
		OutputFile: outputFile,
	}, nil
}

func (e *Executor) Execute(ctx context.Context, command string, workDir string, truncateOpts *TruncateOptions) (*ExecResult, error) {
	startTime := time.Now()

	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = workDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	durationMs := time.Since(startTime).Milliseconds()

	return e.buildResultWithTruncate(ctx, stdout.String(), stderr.String(), durationMs, err, truncateOpts)
}

func (e *Executor) buildResultWithTruncate(ctx context.Context, stdoutStr, stderrStr string, durationMs int64, err error, truncateOpts *TruncateOptions) (*ExecResult, error) {
	combinedOutput := stdoutStr + stderrStr

	result := &ExecResult{
		Stdout:     stdoutStr,
		Stderr:     stderrStr,
		Output:     combinedOutput,
		ExitCode:   0,
		DurationMs: durationMs,
		TimedOut:   false,
		Truncated:  false,
	}

	var metadataLines []string

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
			metadataLines = append(metadataLines, fmt.Sprintf("command timed out after %v", e.timeout))
		} else {
			metadataLines = append(metadataLines, fmt.Sprintf("command execution failed: %v", err))
		}
	}

	if truncateOpts != nil && truncateOpts.MaxLines > 0 && truncateOpts.MaxBytes > 0 {
		truncatedOutput, wasTruncated := truncateOutput(result.Output, truncateOpts.MaxLines, truncateOpts.MaxBytes)
		if wasTruncated {
			result.Truncated = true
			result.Output = truncatedOutput
			metadataLines = append(metadataLines, "output was truncated due to size limits")
		}
	}

	if len(metadataLines) > 0 {
		result.Metadata = formatMetadata(metadataLines)
		result.Output = result.Output + "\n" + result.Metadata
	}

	return result, nil
}

func (e *Executor) ExecuteStream(ctx context.Context, command string, workDir string, onChunk StreamCallback, truncateOpts *TruncateOptions) (*ExecResult, error) {
	startTime := time.Now()

	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = workDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	var stdout, stderr bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	streamReader := func(pipe io.ReadCloser, buf *bytes.Buffer, source string) {
		defer wg.Done()
		reader := bufio.NewReader(pipe)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				buf.WriteString(line)
				if onChunk != nil {
					onChunk(StreamChunk{
						Data:   line,
						Source: source,
					})
				}
			}
			if err != nil {
				if err != io.EOF {
					buf.WriteString(fmt.Sprintf("read error: %v\n", err))
				}
				break
			}
		}
	}

	safe.Go(func() { streamReader(stdoutPipe, &stdout, "stdout") })
	safe.Go(func() { streamReader(stderrPipe, &stderr, "stderr") })

	wg.Wait()
	err = cmd.Wait()
	durationMs := time.Since(startTime).Milliseconds()

	return e.buildResultWithTruncate(ctx, stdout.String(), stderr.String(), durationMs, err, truncateOpts)
}

func truncateOutput(output string, maxLines, maxBytes int) (string, bool) {
	if len(output) <= maxBytes {
		lines := countLines(output)
		if lines <= maxLines {
			return output, false
		}
	}

	lines := splitLines(output)
	var result bytes.Buffer
	lineCount := 0
	byteCount := 0

	for _, line := range lines {
		lineBytes := len(line) + 1
		if byteCount+lineBytes > maxBytes || lineCount >= maxLines {
			break
		}
		if result.Len() > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(line)
		lineCount++
		byteCount += lineBytes
	}

	totalLines := len(lines)
	totalBytes := len(output)
	truncatedLines := totalLines - lineCount
	truncatedBytes := totalBytes - result.Len()

	hint := fmt.Sprintf("\n\n...%d lines (%d bytes) truncated...\n\nUse Grep to search the full content or Read with offset/limit to view specific sections.",
		truncatedLines, truncatedBytes)

	return result.String() + hint, true
}

func countLines(s string) int {
	if len(s) == 0 {
		return 0
	}
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

func splitLines(s string) []string {
	var lines []string
	var current bytes.Buffer

	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current.String())
			current.Reset()
		} else {
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

func formatMetadata(lines []string) string {
	var result bytes.Buffer
	result.WriteString("<bash_metadata>\n")
	for _, line := range lines {
		result.WriteString(line)
		result.WriteByte('\n')
	}
	result.WriteString("</bash_metadata>")
	return result.String()
}
