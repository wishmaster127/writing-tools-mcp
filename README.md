# writing-tools-mcp

Goで実装したMCPサーバーです。`stdio` transportで動作します。

## 提供ツール

- `wt_current_timestamp`  
  現在時刻を `YYYYMMDDHHMM` 形式で返します。
- `wt_count_file_characters`  
  指定したファイルの文字数を返します。改行コード（`\n`, `\r`）はカウントしません。

## 動作要件

- Go 1.25.3

## ビルド方法

### 4ターゲットを一括ビルド（推奨）

次の1コマンドで Linux / macOS(Intel) / macOS(Apple Silicon) / Windows の4種類を `dist/` に出力します。

```bash
go run ./cmd/buildall
```

### 現在の環境向けにビルド

```bash
go build -o writing-tools-mcp .
```

### OS別にクロスコンパイル

```bash
mkdir -p dist

# Linux (x86_64)
GOOS=linux GOARCH=amd64 go build -o dist/writing-tools-mcp-linux-amd64 .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o dist/writing-tools-mcp-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dist/writing-tools-mcp-darwin-arm64 .

# Windows (x86_64)
GOOS=windows GOARCH=amd64 go build -o dist/writing-tools-mcp-windows-amd64.exe .
```

## 実行方法

```bash
go run .
```

またはビルド済みバイナリを実行します。

```bash
./writing-tools-mcp
```

## MCP設定への追加方法

`stdio` で起動するサーバーとして登録します。  
`command` にはビルド済みバイナリの絶対パスを指定してください。

例: `C:\\path\\to\\writing-tools-mcp-windows-amd64.exe`

### Gemini CLI (`~/.gemini/settings.json`)

```json
{
  "mcpServers": {
    "writing-tools-mcp": {
      "command": "C:\\path\\to\\writing-tools-mcp-windows-amd64.exe",
      "args": []
    }
  }
}
```

### Claude Desktop (`claude_desktop_config.json`)

`mcpServers` に以下を追加します。

```json
{
  "mcpServers": {
    "writing-tools-mcp": {
      "command": "C:\\path\\to\\writing-tools-mcp-windows-amd64.exe",
      "args": []
    }
  }
}
```

Claude Desktop の設定ファイル例:
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\\Claude\\claude_desktop_config.json`
