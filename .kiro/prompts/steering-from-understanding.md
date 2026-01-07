You are generating Kiro steering files in CLI-only mode.

INPUT SOURCES (MUST READ):
1) ./MY_UNDERSTANDING.md (repo root) - required
2) .kiro/specs/*/* (requirements.md/design.md/tasks.md) - optional if present; use them to refine/align
If MY_UNDERSTANDING.md is missing, STOP and say exactly that it is required.

OUTPUT (MUST CREATE):
.kiro/steering/product.md
.kiro/steering/tech.md
.kiro/steering/structure.md

INCLUSION:
Add YAML front matter at the top of each steering file:
---
inclusion: always
---

WORKFLOW:

PHASE 1 — EXTRACT (READ-ONLY)
From MY_UNDERSTANDING.md, extract:
- What the system does (current Python system)
- Primary user personas / consumers (human or system)
- Core workflows and success criteria
- External integrations and interfaces
- Current architecture modules/components
- How it runs (runtime expectations)
- Pain points driving the Go rewrite (dependency sprawl, complexity, etc.)
If specs exist, also extract the intended rewrite scope, requirements, and planned design.

PHASE 2 — DRAFT STEERING CONTENT (NO GUESSING)
Write steering docs as guidance to Kiro for the Go rewrite project. If something is unknown, write it under an 'Open Questions' section rather than inventing.

A) product.md must include:
- Product purpose and problem statement (from understanding)
- Target users / stakeholders
- Core capabilities (bullets)
- Non-goals / out-of-scope (if inferable; otherwise list as open questions)
- Success metrics / acceptance definition
- Constraints (hackathon constraints, CLI-only, etc. if relevant and known)
- Open Questions

B) tech.md must include:
- Primary language: Go
- Allowed/Preferred approach: keep dependencies minimal; prefer stdlib unless the project already decided otherwise (derive from understanding/specs)
- Tooling conventions (build/test/lint) ONLY if stated in understanding/specs; otherwise put TODOs
- Runtime/deployment assumptions (from understanding)
- External services (from understanding)
- Security/secret handling expectations (if mentioned)
- Open Questions

C) structure.md must include:
- Desired repo layout for the Go rewrite (derive from rewrite plan/specs; if missing, propose a conservative Go layout and label as 'Proposed' with rationale)
- Naming conventions and package boundaries
- How to add new features (where code goes)
- Testing layout conventions
- Logging/error-handling conventions if stated; otherwise placeholders with TODO/Open Questions
- Mapping guidance: where the quarantined Python source lives (Project_we_are_replicating/) and that it is reference-only

PHASE 3 — CONFIRMATION GATE
Before writing or overwriting any files:
- Print the exact paths that will be written.
- If any already exist, say 'WILL OVERWRITE: <path>'.
- Stop and wait for the user to reply exactly:
  CONFIRM WRITE STEERING

PHASE 4 — WRITE FILES
After confirmation:
- Create .kiro/steering/ if missing
- Write the three files
- Print a short summary and remind how steering is used (it is loaded automatically in Kiro sessions)

RULES:
- Use MY_UNDERSTANDING.md as source of truth; do not invent facts.
- Prefer concrete references (module names, workflows, integrations) when available.
- Keep the docs actionable and concise.
