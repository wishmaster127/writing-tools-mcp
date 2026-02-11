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

func analyzeDialogueAndNarration(text string) (int, int) {
	openerToCloser := map[rune]rune{
		'「': '」',
		'『': '』',
	}

	closerStack := make([]rune, 0, 8)
	dialogue := 0
	narration := 0

	for _, r := range text {
		if r == '\n' || r == '\r' {
			continue
		}

		if len(closerStack) > 0 && r == closerStack[len(closerStack)-1] {
			closerStack = closerStack[:len(closerStack)-1]
			continue
		}

		if closer, ok := openerToCloser[r]; ok {
			closerStack = append(closerStack, closer)
			continue
		}

		if len(closerStack) > 0 {
			dialogue++
		} else {
			narration++
		}
	}

	return dialogue, narration
}

func buildDialogueRatioResult(path string, dialogue int, narration int) (*mcp.CallToolResult, map[string]any, error) {
	total := dialogue + narration
	dialogueRatio := 0.0
	narrationRatio := 0.0
	if total > 0 {
		dialogueRatio = float64(dialogue) / float64(total)
		narrationRatio = float64(narration) / float64(total)
	}

	text := fmt.Sprintf("%s dialogue=%d narration=%d dialogue_ratio=%.4f narration_ratio=%.4f", path, dialogue, narration, dialogueRatio, narrationRatio)
	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, map[string]any{
			"path":            path,
			"dialogue_chars":  dialogue,
			"narration_chars": narration,
			"total_chars":     total,
			"dialogue_ratio":  dialogueRatio,
			"narration_ratio": narrationRatio,
		}, nil
}

func DialogueNarrationRatioTool(ctx context.Context, req *mcp.CallToolRequest, input map[string]any) (*mcp.CallToolResult, map[string]any, error) {
	path, err := getString(input, "path")
	if err != nil {
		return nil, nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	dialogue, narration := analyzeDialogueAndNarration(string(b))
	return buildDialogueRatioResult(path, dialogue, narration)
}

func DialogueNarrationRatioRangeTool(ctx context.Context, req *mcp.CallToolRequest, input map[string]any) (*mcp.CallToolResult, map[string]any, error) {
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

	var sb strings.Builder
	lineNo := 0
	lines := 0

	for scanner.Scan() {
		lineNo++
		if lineNo < startLine {
			continue
		}
		if lineNo > endLine {
			break
		}

		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
		lines++
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to scan file: %w", err)
	}

	dialogue, narration := analyzeDialogueAndNarration(sb.String())
	result, meta, err := buildDialogueRatioResult(fmt.Sprintf("%s:%d-%d", path, startLine, endLine), dialogue, narration)
	if err != nil {
		return nil, nil, err
	}
	meta["lines"] = lines
	return result, meta, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "writing-tools-mcp",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "timestamp",
		Description: "Returns the current timestamp in YYYYMMDDHHMM format",
		InputSchema: map[string]any{
			"$schema":    "https://json-schema.org/draft/2020-12/schema",
			"type":       "object",
			"properties": map[string]any{
				// no arguments
			},
			"additionalProperties": false,
		},
	}, CurrentTimestampTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "count_chars",
		Description: "Count UTF-8 characters in an entire text file (excluding line breaks).\nUse this tool when you need the total character count of a file.\n\nInput:\n- path: file path to a UTF-8 text file (most common text files)\n\nArguments must be provided as JSON object like the example below.\nExample:\n{\n  \"path\": \"novel.txt\"\n}\n\nThe file must exist and be readable by the server process.\nBoth absolute and relative paths are supported.\nRelative paths are resolved from the server process working directory.\n\nFor counting a specific section or scene, use wt_count_chars_range.",
		InputSchema: map[string]any{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"type":    "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to a text file. The file must exist and be readable by the server process.",
					"minLength":   1,
				},
			},
			"required":             []string{"path"},
			"additionalProperties": false,
		},
	}, CountFileCharactersTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "count_chars_range",
		Description: "Count UTF-8 characters between start_line and end_line in a text file.\nUse this tool when you need character count of a specific section or scene.\n\nInput:\n- path: file path to a UTF-8 text file (most common text files)\n- start_line: starting line number (1-based)\n- end_line: ending line number (inclusive)\n\nArguments must be provided as JSON object like the example below.\nExample:\n{\n  \"path\": \"novel.txt\",\n  \"start_line\": 120,\n  \"end_line\": 180\n}\n\nThe file must exist and be readable by the server process.\nBoth absolute and relative paths are supported.\nRelative paths are resolved from the server process working directory.\nLine numbers are 1-based and inclusive; if end_line exceeds file length, existing lines are counted.",
		InputSchema: map[string]any{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"type":    "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to a text file. The file must exist and be readable by the server process.",
					"minLength":   1,
				},
				"start_line": map[string]any{
					"type":        "integer",
					"description": "Starting line number (1-based).",
					"minimum":     1,
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Ending line number (inclusive).",
					"minimum":     1,
				},
			},
			"required":             []string{"path", "start_line", "end_line"},
			"additionalProperties": false,
		},
	}, CountFileCharactersRangeTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "dialogue_ratio",
		Description: "Analyze an entire text file and return dialogue/narration character counts and ratios.\nDialogue is text enclosed by Japanese quotes such as 「...」 and 『...』. Quote symbols and line breaks are excluded from counting.\n\nInput:\n- path: file path to a UTF-8 text file\n\nExample:\n{\n  \"path\": \"novel.txt\"\n}\n\nFor scene-level analysis by line range, use wt_dialogue_ratio_range.",
		InputSchema: map[string]any{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"type":    "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to a text file. The file must exist and be readable by the server process.",
					"minLength":   1,
				},
			},
			"required":             []string{"path"},
			"additionalProperties": false,
		},
	}, DialogueNarrationRatioTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        toolPrefix + "dialogue_ratio_range",
		Description: "Analyze dialogue/narration character counts and ratios between start_line and end_line in a text file.\nDialogue is text enclosed by Japanese quotes such as 「...」 and 『...』. Quote symbols and line breaks are excluded from counting.\n\nInput:\n- path: file path to a UTF-8 text file\n- start_line: starting line number (1-based)\n- end_line: ending line number (inclusive)\n\nExample:\n{\n  \"path\": \"novel.txt\",\n  \"start_line\": 120,\n  \"end_line\": 180\n}",
		InputSchema: map[string]any{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"type":    "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to a text file. The file must exist and be readable by the server process.",
					"minLength":   1,
				},
				"start_line": map[string]any{
					"type":        "integer",
					"description": "Starting line number (1-based).",
					"minimum":     1,
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Ending line number (inclusive).",
					"minimum":     1,
				},
			},
			"required":             []string{"path", "start_line", "end_line"},
			"additionalProperties": false,
		},
	}, DialogueNarrationRatioRangeTool)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
