package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const toolPrefix = "wt_"

type TimestampOutput struct {
	Timestamp string `json:"timestamp" jsonschema:"current timestamp in YYYYMMDDHHMM format"`
}

type CountFileCharactersInput struct {
	Path string `json:"path" jsonschema:"absolute or relative file path to count characters from"`
}

type CountFileCharactersOutput struct {
	Path      string `json:"path" jsonschema:"file path that was counted"`
	Character int    `json:"character_count" jsonschema:"character count excluding line break codes"`
}

func CurrentTimestampTool(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, TimestampOutput, error) {
	ts := time.Now().Format("200601021504")

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: ts},
			},
		}, TimestampOutput{
			Timestamp: ts,
		}, nil
}

func CountFileCharactersTool(ctx context.Context, req *mcp.CallToolRequest, input CountFileCharactersInput) (*mcp.CallToolResult, CountFileCharactersOutput, error) {
	if strings.TrimSpace(input.Path) == "" {
		return nil, CountFileCharactersOutput{}, fmt.Errorf("path is required")
	}

	b, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, CountFileCharactersOutput{}, fmt.Errorf("failed to read file: %w", err)
	}

	withoutLF := strings.ReplaceAll(string(b), "\n", "")
	withoutLineBreaks := strings.ReplaceAll(withoutLF, "\r", "")
	count := utf8.RuneCountInString(withoutLineBreaks)

	text := fmt.Sprintf("%s: %d", input.Path, count)
	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, CountFileCharactersOutput{
			Path:      input.Path,
			Character: count,
		}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "writing-tools-mcp",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "current_timestamp",
		Description: "Returns the current timestamp in YYYYMMDDHHMM format",
	}, CurrentTimestampTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "count_file_characters",
		Description: "Count file characters excluding line break codes (LF/CR)",
	}, CountFileCharactersTool)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
