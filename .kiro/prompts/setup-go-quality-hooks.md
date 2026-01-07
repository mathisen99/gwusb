Create a local Kiro CLI agent config at .kiro/agents/go-rewrite.json with Go quality hooks.

Requirements:
- The agent must have hooks that run Go quality checks automatically.
- Use:
  - gofmt (or go fmt)
  - go vet
  - golangci-lint
- Run checks after code is written and also at end of each assistant turn.
- If golangci-lint is missing, print a clear error and fail (non-zero exit) so the user sees a warning.
- Do not commit automatically in hooks.

Write exactly this JSON (adjust only if necessary for correctness):

{
  "name": "go-rewrite",
  "description": "Go rewrite agent with automatic formatting + vet + lint hooks",
  "tools": ["@builtin"],
  "hooks": {
    "postToolUse": [
      {
        "matcher": "write",
        "command": "go fmt ./... && go vet ./... && (command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || (echo 'ERROR: golangci-lint not installed' 1>&2; exit 1))"
      }
    ],
    "stop": [
      {
        "command": "go fmt ./... && go vet ./... && (command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || (echo 'ERROR: golangci-lint not installed' 1>&2; exit 1))"
      }
    ]
  }
}

Before writing, print the target path and whether it will overwrite anything, then wait for the user to reply exactly: CONFIRM WRITE AGENT.
After confirmation, write the file and tell the user how to run with it.
