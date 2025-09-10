package main

import (
	"context"
	"fmt"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox"
	"github.com/lemonlyue/code-sandbox-mcp/sandbox/docker"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {

	s := server.NewMCPServer("Sandbox Server", "1.0.0")
	sandboxTool := mcp.NewTool(
		"execute_code_in_sandbox",
		mcp.WithDescription("在沙盒环境执行代码 | Execute the code in a sandbox environment"),
		mcp.WithString("language", mcp.Required(), mcp.Description("编程语言 | Programming language")),
		mcp.WithString("code", mcp.Required(), mcp.Description("需要执行的代码 | The code to be executed")),
		mcp.WithString("version", mcp.Description("编程语言版本 | Programming language version")),
	)

	s.AddTool(sandboxTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
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

		config := &sandbox.Config{
			Language: language,
			Version:  version,
		}
		sb, err := factory.Create(context.Background(), config)
		if err != nil {
			panic(err)
		}
		execute, err := sb.Execute(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("failed to execute in sandbox: %w", err)
		}

		return mcp.NewToolResultText(execute.Stdout), nil
	})

	if err := server.ServeStdio(s); err != nil {
		// todo
	}
}
