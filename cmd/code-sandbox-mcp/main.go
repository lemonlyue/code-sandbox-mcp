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

	mcp "trpc.group/trpc-go/trpc-mcp-go"
)

func main() {
	log.Printf("Starting basic example server...")

	mcpServer := mcp.NewServer(
		"Sandbox-Server",
		"0.1.0",
		mcp.WithServerAddress(":3000"),
		mcp.WithServerPath("/mcp"),
		mcp.WithServerLogger(mcp.GetDefaultLogger()),
	)

	sandboxTool := mcp.NewTool("execute_code_in_sandbox",
		mcp.WithDescription("在沙盒环境执行代码 | Execute the code in a sandbox environment"),
		mcp.WithString("language", mcp.Required(), mcp.Description("编程语言 | Programming language")),
		mcp.WithString("code", mcp.Required(), mcp.Description("需要执行的代码 | The code to be executed")),
		mcp.WithString("version", mcp.Description("编程语言版本 | Programming language version")),
	)

	sandboxHandler := func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			BaseImage:  languageConfig.BaseImage,
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

	mcpServer.RegisterTool(sandboxTool, sandboxHandler)
	log.Printf("Registered basic sandbox tool: sandbox")

	// Set up a graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("MCP server started, listening on port 3000, path /mcp")
		if err := mcpServer.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Printf("Shutting down server...")
}
