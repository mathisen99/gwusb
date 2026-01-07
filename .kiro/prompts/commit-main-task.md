You are finalizing a MAIN TASK and preparing a git commit.

SOURCE OF TRUTH:
- Use `git diff --stat` and `git diff` to understand what changed.
- Do NOT invent intent. Derive it from the diff only.

PROCESS:

1) QUALITY GATES (FAIL FAST)
Run, in order:
- go fmt ./...
- go vet ./...
- golangci-lint run ./...
- go test ./...

If any command fails, STOP and report the failure clearly.
Do not stage or commit anything.

2) CHANGE ANALYSIS
Run:
- git diff --stat
- git diff

From the diff, determine:
- Primary intent of the change (e.g. refactor, new feature, bug fix)
- Key areas/modules affected
- Scope (small/medium/large)

3) COMMIT MESSAGE GENERATION
Generate a commit message in this format:

<type>: <concise summary>

<optional body>
- bullet points describing notable changes
- derived strictly from the diff

Rules:
- Use conventional types where applicable: feat, fix, refactor, chore, test, docs
- Keep the summary under 72 characters
- Do NOT reference tasks, specs, or assumptions unless visible in the diff

4) CONFIRMATION
Print:
- git diff --stat
- the proposed commit message

Then ask:
Type EXACTLY 'CONFIRM COMMIT' to proceed.

5) COMMIT
Only after confirmation:
- git add -A
- git commit -m "<generated message>" (include body if present)

SAFETY RULES:
- Never commit without explicit confirmation.
- Never modify files other than via git add/commit.
- If the working tree is clean, STOP and say so.

