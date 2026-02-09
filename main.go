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

func getString(m map[string]any, key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("%s is required", key)
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", fmt.Errorf("%s must be string", key)
	}
	return s, nil
}

func CurrentTimestampTool(ctx context.Context, req *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	ts := time.Now().Format("200601021504")

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: ts},
			},
		}, map[string]any{
			"timestamp": ts,
		}, nil
}

func CountFileCharactersTool(ctx context.Context, req *mcp.CallToolRequest, input map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	path, err := getString(input, "path")
	if err != nil {
		return nil, nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	normalized := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, string(b))
	count := utf8.RuneCountInString(normalized)

	text := fmt.Sprintf("%s: %d", path, count)
	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, map[string]any{
			"path":            path,
			"character_count": count,
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
