You are performing a rewrite gap analysis between an original Python project and its Go rewrite.

SOURCE OF TRUTH (MUST READ):
1) ./MY_UNDERSTANDING.md
2) Current Go code in the repository
3) Specs in .kiro/specs/* (if present)

OUTPUT:
- Write a new file in the repo root named:
  REWRITE_GAPS.md

PROCESS:

1. Extract expected behavior
From MY_UNDERSTANDING.md, extract:
- Core features and workflows
- CLI commands / flags (if applicable)
- Supported environments and constraints
- Edge cases and special behaviors
- Non-obvious logic or heuristics

2. Inspect current Go implementation
Determine:
- Which features are implemented
- Which are partially implemented
- Which are missing
- Any behavior that differs from the original

Do NOT assume intent. Base conclusions on code and specs only.

3. Gap classification
For each identified gap, classify it as ONE of:
- Missing (not implemented at all)
- Partial (implemented but incomplete)
- Different (behavior does not match Python version)
- Intentional omission (explicitly documented as skipped)
- Unknown (cannot be confirmed from available sources)

4. Risk assessment
For each gap, assess:
- User impact (low / medium / high)
- Rewrite risk (low / medium / high)
- Whether it blocks parity with the original tool

5. Write REWRITE_GAPS.md with this structure:

# Rewrite Gap Analysis

## Summary
- High-level status of parity with the original project
- Major risk areas

## Feature-by-Feature Gaps
For each feature:
- Description
- Original behavior (from MY_UNDERSTANDING.md)
- Current Go status
- Gap classification
- Risk assessment
- Notes / references (files, functions, specs)

## Intentional Differences
- Features or behaviors intentionally not ported
- Rationale (from specs or steering, if present)

## Unknowns and Open Questions
- Behaviors that require manual testing or deeper investigation

## Recommended Next Steps
- Concrete actions to close high-risk gaps
- Suggested spec or task updates

6. Confirmation gate
Before writing REWRITE_GAPS.md:
- Print the file path
- Print a short outline (section headers + number of gaps found)
Ask the user to reply exactly:
CONFIRM WRITE GAP ANALYSIS

If not confirmed, do not write the file.

RULES:
- Do not invent missing features.
- If something cannot be verified, mark it Unknown.
- Prefer explicit references to files/functions/specs.
- Do not modify any files except REWRITE_GAPS.md.
