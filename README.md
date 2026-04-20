# DingTalk Dialog

> [中文文档](README.zh.md)

`github.com/nuln/agent-dialog-dingtalk` — DingTalk Dialog plugin for [nuln/agent-core](https://github.com/nuln/agent-core).

## Overview

| Field | Value |
|-------|-------|
| **Plugin Type** | `dialog` |
| **Module** | `github.com/nuln/agent-dialog-dingtalk` |
| **Key Dependency** | `github.com/open-dingtalk/dingtalk-stream-sdk-go` |

## Installation

```bash
go get github.com/nuln/agent-dialog-dingtalk
```

Import the package in your `main.go` (side-effect import triggers `init()`):

```go
import _ "github.com/nuln/agent-dialog-dingtalk"
```

## Configuration

Configure via environment variables or the Web UI.  
See `RegisterPluginConfigSpec` in the plugin source for the full field list.

## Development

```bash
make fmt     # format code
make lint    # run golangci-lint
make test    # run tests
make build   # go build ./...
```

## License

MIT
