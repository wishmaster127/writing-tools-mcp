package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
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

func getInt(m map[string]any, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}

	switch n := v.(type) {
	case float64:
		if math.Trunc(n) != n {
			return 0, fmt.Errorf("%s must be integer", key)
		}
		return int(n), nil
	case int:
		return n, nil
	case int8:
		return int(n), nil
	case int16:
		return int(n), nil
	case int32:
		return int(n), nil
	case int64:
		return int(n), nil
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return 0, fmt.Errorf("%s must be number", key)
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("%s must be integer", key)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("%s must be number", key)
	}
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

func CountFileCharactersRangeTool(ctx context.Context, req *mcp.CallToolRequest, input map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	path, err := getString(input, "path")
	if err != nil {
		return nil, nil, err
	}

	startLine, err := getInt(input, "start_line")
	if err != nil {
		return nil, nil, err
	}
	if startLine < 1 {
		return nil, nil, fmt.Errorf("start_line must be >= 1")
	}

	endLine, err := getInt(input, "end_line")
	if err != nil {
		return nil, nil, err
	}
	if startLine > endLine {
		return nil, nil, fmt.Errorf("start_line must be <= end_line")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	lineNo := 0
	characters := 0
	lines := 0

	for scanner.Scan() {
		lineNo++
		if lineNo < startLine {
			continue
		}
		if lineNo > endLine {
			break
		}

		characters += utf8.RuneCountInString(scanner.Text())
		lines++
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to scan file: %w", err)
	}

	text := fmt.Sprintf("%s:%d-%d characters=%d lines=%d", path, startLine, endLine, characters, lines)
	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, map[string]any{
			"characters": characters,
			"lines":      lines,
		}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "writing-tools-mcp",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "timestamp",
		Description: "Returns the current timestamp in YYYYMMDDHHMM format",
	}, CurrentTimestampTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "count_chars",
		Description: "Count file characters excluding line break codes (LF/CR)",
	}, CountFileCharactersTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "count_chars_range",
		Description: "Count UTF-8 characters in a line range without loading entire file",
	}, CountFileCharactersRangeTool)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
