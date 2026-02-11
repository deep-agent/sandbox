package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx := context.Background()

	mcpTools := getMCPTools(ctx)
	if len(mcpTools) == 0 {
		log.Fatal("No MCP tools found")
	}

	fmt.Println("=== Available MCP Tools ===")
	for i, mcpTool := range mcpTools {
		info, err := mcpTool.Info(ctx)
		if err != nil {
			log.Printf("Failed to get tool info: %v", err)
			continue
		}
		fmt.Printf("%d. Name: %s\n", i+1, info.Name)
		// fmt.Printf("%d. Description: %s\n", i+1, info.Desc)
	}
	fmt.Println()

	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_MODEL"),
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	// callback := react.BuildAgentCallback(&template.ModelCallbackHandler{}, &template.ToolCallbackHandler{})
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		Model: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: mcpTools,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create react agent: %v", err)
	}

	fmt.Println("=== Running ReAct Agent (Streaming) ===")
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "You are a helpful assistant. Use the available tools to help answer user questions.",
		},
		{
			Role:    schema.User,
			Content: "帮我用 cdp 工具访问一下 http://localhost:8000/ 地址然后截图返回。",
		},
	}

	stream, err := agent.Stream(ctx, messages)
	// flowagent.WithComposeOptions(compose.WithCallbacks(callback)),
	// flowagent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{})),

	if err != nil {
		log.Fatalf("Failed to stream: %v", err)
	}
	defer stream.Close()

	isThinking := false
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Failed to receive chunk: %v", err)
		}

		if chunk.ReasoningContent != "" {
			if !isThinking {
				fmt.Print("\n[Thinking] ")
				isThinking = true
			}
			fmt.Print(chunk.ReasoningContent)
		}

		if chunk.Content != "" {
			if isThinking {
				fmt.Print("\n[Answer] ")
				isThinking = false
			}
			fmt.Print(chunk.Content)
		}

		if len(chunk.ToolCalls) > 0 {
			for _, tc := range chunk.ToolCalls {
				if tc.Function.Name != "" {
					fmt.Printf("\n[Tool Call: %s]\n", tc.Function.Name)
				}
				if tc.Function.Arguments != "" {
					fmt.Printf("[Arguments: %s]\n", tc.Function.Arguments)
				}
			}
		}
	}
	fmt.Println()
	fmt.Println("=== Stream Completed ===")
}

func getMCPTools(ctx context.Context) []tool.BaseTool {
	mcpURL := os.Getenv("MCP_URL")
	if mcpURL == "" {
		mcpURL = "http://127.0.0.1:8080/mcp"
	}

	fmt.Printf("Connecting to MCP server: %s\n", mcpURL)

	cli, err := client.NewStreamableHttpClient(mcpURL)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}

	err = cli.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start MCP client (check if server is running at %s): %v", mcpURL, err)
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "eino-mcp-demo",
		Version: "1.0.0",
	}

	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize MCP client: %v", err)
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: cli})
	if err != nil {
		log.Fatalf("Failed to get MCP tools: %v", err)
	}

	return tools
}

type LoggerCallback struct {
	callbacks.HandlerBuilder // 可以用 callbacks.HandlerBuilder 来辅助实现 callback
}

func (cb *LoggerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	fmt.Println("==================")
	// inputStr, _ := json.MarshalIndent(input, "", "  ")
	fmt.Printf("[OnStart] \n")
	return ctx
}

func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	fmt.Println("=========[OnEnd]=========")
	// outputStr, _ := json.MarshalIndent(output, "", "  ")
	// fmt.Println(string(outputStr))
	return ctx
}

func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	fmt.Println("=========[OnError]=========")
	fmt.Println(err)
	return ctx
}

func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	// var graphInfoName = react.GraphName

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[OnEndStream] panic err:", err)
			}
		}()

		defer output.Close() // remember to close the stream in defer

		// fmt.Println("=========[OnEndStream]=========")
		// for {
		// 	frame, err := output.Recv()
		// 	if errors.Is(err, io.EOF) {
		// 		// finish
		// 		break
		// 	}
		// 	if err != nil {
		// 		fmt.Printf("internal error: %s\n", err)
		// 		return
		// 	}

		// 	s, err := json.Marshal(frame)
		// 	if err != nil {
		// 		fmt.Printf("internal error: %s\n", err)
		// 		return
		// 	}

		// 	if info.Name == graphInfoName { // 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
		// 		fmt.Printf("%s: %s\n", info.Name, string(s))
		// 	}
		// }

	}()
	return ctx
}

func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}
