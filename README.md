# gontribute

gontribute is a tool that encourages a step forward in contributing to OSS made by Go. It applies various static analysis tools to third-party packages on which Go projects depend.

## Requirements

- [goreportcard-cli](https://github.com/gojp/goreportcard/tree/master/cmd/goreportcard-cli) 
  - Install https://github.com/gojp/goreportcard#command-line-interface

## Usage

```bash
GITHUB_ACCESS_TOKEN=<your-token> GITHUB_TARGET_OWNER=<repo-owner> GITHUB_TARGET_REPO=<repo-name> go run main.go
```
