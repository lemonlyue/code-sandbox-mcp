package main

import (
	"context"
	"fmt"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox/docker"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcp "trpc.group/trpc-go/trpc-mcp-go"
)

func main() {
	// Create SSE server.
	server := mcp.NewSSEServer(
		"SSE Code Sandbox Server",           // Server name.
		"1.0.0",                             // Server version.
		mcp.WithSSEEndpoint("/sse"),         // Explicitly set SSE endpoint.
		mcp.WithMessageEndpoint("/message"), // Explicitly set message endpoint.
	)

	// Register tools.
	sandboxTool := mcp.NewTool("execute_code_in_sandbox",
		mcp.WithDescription("在沙盒环境执行代码 | Execute the code in a sandbox environment"),
		mcp.WithString("language", mcp.Required(), mcp.Description("编程语言 | Programming language")),
		mcp.WithString("code", mcp.Required(), mcp.Description("需要执行的代码 | The code to be executed")),
		mcp.WithString("version", mcp.Description("编程语言版本 | Programming language version")),
	)
	server.RegisterTool(sandboxTool, sandboxHandler)

	log.Printf("Registered tools: execute_code_in_sandbox")
	log.Printf("SSE endpoint: /sse")
	log.Printf("Message endpoint: /message")

	// Set graceful exit.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Start server.
	log.Printf("Starting SSE server on port 4000...")
	go func() {
		if err := server.Start(":4000"); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for exit signal.
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Graceful exit.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// sandboxHandler handles greet tool callback function.
func sandboxHandler(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	select {
	case <-ctx.Done():
		return mcp.NewErrorResult("Request cancelled"), ctx.Err()
	default:

	}

	args, ok := interface{}(request.Params.Arguments).(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments format, expected a map")
	}
	language, languageOk := args["language"].(string)
	version, _ := args["version"].(string)
	code, codeOk := args["code"].(string)
	if !languageOk {
		return nil, fmt.Errorf("missing required argument: 'language'")
	}
	if !codeOk {
		return nil, fmt.Errorf("missing required argument: 'code'")
	}

	dockerCreatorFunc := docker.NewDockerSandboxCreator()

	factory := sandbox.NewFactory(
		sandbox.WithDockerCreator(dockerCreatorFunc),
	)

	configManager, err := sandbox.NewConfigManager()
	if err != nil {
		return nil, err
	}
	languageConfig := configManager.GetLanguageConfig(language)

	config := &sandbox.Config{
		Language:   language,
		Version:    version,
		Image:      languageConfig.BaseImage,
		BaseImage:  languageConfig.DefaultImage,
		Entrypoint: languageConfig.Entrypoint,
		Suffix:     languageConfig.Suffix,
	}
	sb, err := factory.Create(context.Background(), config)
	if err != nil {
		panic(err)
	}
	execute, err := sb.Execute(ctx, code)
	log.Printf("result: %+v", execute)
	if err != nil {
		return nil, fmt.Errorf("failed to execute in sandbox: %w", err)
	}

	return mcp.NewTextResult(execute.Stdout), nil
}
