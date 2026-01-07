You are marking a MAIN TASK as complete in Kiro spec tasks.

SCOPE:
- Only modify a single tasks.md entry.
- Do not run go commands.
- Do not stage or commit anything.

WHERE TASKS LIVE:
.kiro/specs/<spec_slug>/tasks.md

PROCESS:

1) Discover specs:
- List .kiro/specs/*/tasks.md
If none exist, STOP and say no spec tasks files were found.

2) Ask user for ONE of these selectors:
A) spec slug + task number (preferred)
B) spec slug + exact task text snippet
If the user provides neither, show a numbered menu of available specs (slugs) and ask them to pick one.

3) Load the selected tasks.md and locate the task:
- If task number was given: locate that numbered task item.
- If text snippet was given: locate the single best exact match.
If there are zero matches or multiple plausible matches, STOP and show the closest matching lines and ask the user to specify more precisely.

4) Mark complete using the file’s existing style:
- If the task line uses markdown checkboxes '- [ ]' -> change to '- [x]'
- If it is a numbered list without checkboxes:
  - Prefer to convert ONLY that task line to a checkbox form without reformatting everything:
    Example: '3. Do X' -> '3. [x] Do X'
  - If that would be inconsistent with the file, append ' (done)' to that one line.
Do not rewrap, reorder, or reformat other content.

5) Confirmation gate:
- Print:
  - file path
  - the exact line BEFORE and AFTER
Ask the user to reply exactly:
CONFIRM MARK DONE
If not confirmed, do not write changes.

6) Write the updated tasks.md.
Then print:
- 'Task marked complete.'
- A reminder: run @commit-main-task to commit when ready.

SAFETY RULES:
- Never mark a task complete without explicit confirmation.
- Only change one task line.
