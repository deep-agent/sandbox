package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/deep-agent/sandbox/internal/services/bash"
	"github.com/deep-agent/sandbox/model"
)

type BashHandler struct {
	executor *bash.Executor
}

func NewBashHandler(executor *bash.Executor) *BashHandler {
	return &BashHandler{executor: executor}
}

func (h *BashHandler) ExecCommand(ctx context.Context, c *app.RequestContext) {
	var req model.BashExecRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if req.Cwd == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "cwd is required",
		})
		return
	}

	if req.RunInBackground {
		timeout := 10 * time.Minute
		if req.TimeoutMS > 0 {
			timeout = time.Duration(req.TimeoutMS) * time.Millisecond
		}
		result, err := h.executor.ExecuteBackground(ctx, req.Command, req.Cwd, timeout)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    500,
				Message: "execution failed: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, model.Response{
			Code: 0,
			Data: model.BashExecResult{
				Output:     result.Output,
				ExitCode:   result.ExitCode,
				OutputFile: result.OutputFile,
			},
		})
		return
	}

	if req.TimeoutMS > 0 {
		h.executor.SetTimeout(time.Duration(req.TimeoutMS) * time.Millisecond)
	}

	result, err := h.executor.Execute(ctx, req.Command, req.Cwd, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "execution failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BashExecResult{
			Output:   result.Output,
			ExitCode: result.ExitCode,
		},
	})
}

type StreamEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type StreamChunkData struct {
	Data   string `json:"data"`
	Source string `json:"source"`
}

type StreamDoneData struct {
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	TimedOut   bool   `json:"timed_out"`
	Truncated  bool   `json:"truncated"`
}

func (h *BashHandler) ExecCommandStream(ctx context.Context, c *app.RequestContext) {
	var req model.BashExecRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if req.Cwd == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "cwd is required",
		})
		return
	}

	if req.TimeoutMS > 0 {
		h.executor.SetTimeout(time.Duration(req.TimeoutMS) * time.Millisecond)
	}

	c.SetStatusCode(consts.StatusOK)
	c.Response.Header.Set("Content-Type", "text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("X-Accel-Buffering", "no")

	sendEvent := func(event string, data interface{}) {
		jsonData, _ := json.Marshal(StreamEvent{Event: event, Data: data})
		c.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData)))
		c.Flush()
	}

	onChunk := func(chunk bash.StreamChunk) {
		sendEvent("chunk", StreamChunkData{
			Data:   chunk.Data,
			Source: chunk.Source,
		})
	}

	result, err := h.executor.ExecuteStream(ctx, req.Command, req.Cwd, onChunk, nil)

	if err != nil {
		sendEvent("error", map[string]string{"message": err.Error()})
		return
	}

	sendEvent("done", StreamDoneData{
		Output:     result.Output,
		ExitCode:   result.ExitCode,
		DurationMs: result.DurationMs,
		TimedOut:   result.TimedOut,
		Truncated:  result.Truncated,
	})
}
