You are generating Kiro spec files from an existing project understanding document.

INPUT SOURCE (MUST READ):
- ./MY_UNDERSTANDING.md (in repo root). If it does not exist, stop and tell the user exactly what file is missing.

OUTPUT STRUCTURE (MUST CREATE):
.kiro/specs/<spec_slug>/
  requirements.md
  design.md
  tasks.md

WORKFLOW:

1) Read and extract from MY_UNDERSTANDING.md:
   - Executive summary / purpose
   - Architecture overview (components, data flows)
   - External integrations
   - Runtime/how-to-run info
   - Risks/unknowns
   - Any proposed rewrite plan (if present)

2) Ask the user for:
   - spec name (short; if not kebab-case, convert to safe kebab-case)
   - the scope boundary for THIS spec (what is included/excluded), because a whole app may be too large for one spec

3) Draft content using extracted facts (do not invent):
   A) requirements.md
      - Introduction: 5–10 bullets summarizing the capability in this spec (derived from MY_UNDERSTANDING.md)
      - User stories
      - Acceptance criteria in EARS format, e.g.:
        WHEN <condition>
        THE SYSTEM SHALL <behavior>
      Requirements must be testable.

   B) design.md
      - Context (what existing system does today)
      - Proposed target design for the Go rewrite of THIS scope
      - Components/modules (names + responsibilities)
      - Interfaces/APIs (inputs/outputs/contracts)
      - Data model and persistence (if applicable)
      - Error handling and failure modes
      - Security considerations
      - Observability (logs/metrics/tracing)
      - Add mermaid diagrams where useful

   C) tasks.md
      - Numbered tasks with clear outcomes
      - Each task maps back to one or more requirements
      - Include testing tasks and documentation tasks
      - Keep tasks small and sequential; include dependencies

4) BEFORE WRITING FILES:
   - Print the exact paths to be created/overwritten.
   - Print a 10–15 line preview summary of what will go into each file.
   - Stop and wait for the user to reply exactly:
     CONFIRM WRITE SPECS

5) AFTER CONFIRMATION:
   - Create .kiro/specs/<spec_slug>/ and write the three files.
   - Print how to reference it in chat:
     #spec:<spec_slug>

RULES:
- Use MY_UNDERSTANDING.md as the single source of truth; if something is missing, list it under 'Assumptions / Open Questions' instead of guessing.
- Do not modify any other files besides the spec files.
