Perform a complete code review of the current Go codebase.

Scope:
- Review the entire repo, focusing on Go source, tests, build config, and tooling.
- Do not modify files unless the user explicitly asks for fixes.

Process:
1) Repo inventory
- Identify module layout (go.mod, packages, cmd/, internal/, pkg/, etc.)
- Identify entry points and main binaries.
- Identify CI/lint config if present (golangci-lint config, Makefile, scripts).

2) Correctness and reliability
- Error handling patterns (wrapping, sentinel errors, context usage).
- Concurrency safety (goroutines, channels, mutexes, context cancellation).
- Resource management (files, network connections, defer correctness).
- Input validation and edge cases.

3) Go best practices
- Package boundaries and import hygiene.
- Naming, exported API design, doc comments.
- Avoiding anti-patterns (global state, init complexity, excessive interfaces).
- Simplicity and standard library preference.

4) Performance (only where relevant)
- Obvious hot paths, unnecessary allocations, inefficient IO.
- Avoid premature micro-optimizations; focus on clear wins.

5) Testing and quality
- Test coverage of critical paths.
- Table-driven tests, test helpers, fixtures.
- Determinism, race potential, missing negative tests.
- Recommend go test flags and patterns.

6) Tooling alignment
- Ensure formatting is gofmt-compliant.
- Ensure 'go vet' and 'golangci-lint' findings are anticipated.
- If hooks are configured, verify they match repo reality.

Output format:
- Start with a short overall assessment (5–10 lines).
- Then provide findings grouped by severity:
  - Blockers (must fix)
  - Major
  - Minor
  - Suggestions
For each finding:
- Quote the file path and symbol name (function/type) where applicable.
- Explain why it matters.
- Provide an actionable recommendation.
If you cannot locate the relevant code, say so explicitly.

Do not invent project requirements. If needed, reference steering/spec files for intent:
- .kiro/steering/*
- .kiro/specs/*

Do not modify files.
