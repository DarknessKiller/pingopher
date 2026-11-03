---
name: change-verification
description: Prove a finished change before handoff, commit, push, or PR. Spawns a behavior-prover and a standards-reviewer subagent in parallel, reviews every changed line, and attributes every failing check before any gate is skipped.
---
# Change verification

1. **Record the baseline.** `git rev-parse HEAD` is `$BASE`; `git status --porcelain` lists the changed paths. Both go to every lane below. A new environment variable lands in `internal/config/config.go`, `.env.example`, and the README configuration table in the same change.

2. **Launch both lanes at once.** Use whatever subagent tool the harness provides (`Agent` in pi, `Task` in Claude Code, the equivalent elsewhere), one lane per subagent, both started before you wait for either. A harness without subagents runs the two prompts here, one after the other. The lane prompt is the whole message; fill its placeholders.

   **Behavior lane**:
   ```
   You are the behavior lane of change verification. Base: <base>. Changed: <paths>. Boundary: <route, command, or flow the change crosses>.

   1. Read the change: `git diff <base>` plus `git status --porcelain`.
   2. Exercise the change where production runs it: the registered route, the real command, the real decoder. Cover success, rejection, and the absence of an upstream side effect. Typed mocks and test-only routes prove nothing here.
   3. Run the Verify block in `CODING_STANDARDS.md` for the changed files, plus the idioms list for each touched Go file:
      sh .agents/skills/use-modern-go/scripts/run-tool.sh list --file-path <file>

   Report:
   lane: behavior
   verdict: pass | fail | blocked
   checks: `<command>` pass | fail: <first failing line>
   findings: path:line <severity> <problem>. <fix>.
   gaps: <check skipped and why>
   ```

   **Standards lane**:
   ```
   You are the standards lane of change verification. Base: <base>. Changed: <paths>.

   1. Read `git diff <base>` and `CODING_STANDARDS.md`.
   2. For each changed line, name the section it falls under and check it against that section. Read the canonical file the section points at, and the same pattern elsewhere in the repo, before reporting.
   3. Run the idioms list for each touched Go file and check the listed idioms:
      sh .agents/skills/use-modern-go/scripts/run-tool.sh list --file-path <file>
   4. Where the code is right and a rule is wrong, report the doc fix in `CODING_STANDARDS.md`.

   Report:
   lane: standards
   verdict: pass | fail | blocked
   findings: path:line <severity> <problem>. <fix>.
   gaps: <check skipped and why>
   ```

   Read the target files and plan fixes while they run.

3. **Merge the reports.** Each lane ends with `verdict: pass | fail | blocked`.
   - `fail` naming a line this change owns: fix it, then rerun that lane.
   - `blocked`: handoff stops until the lane runs.
   - both `pass`: continue.

4. **Run the gates on the final tree.** Run the Verify block in `CODING_STANDARDS.md` for the changed files, after every fix. A previous pass does not clear a later failure.

5. **Attribute every failure.** `git diff $BASE -- <paths>` shows what the change reaches. A failure in a path the diff does not touch is unrelated: record the command, the failing path, and that path's absence from the diff. A failure in a touched path is owned and gets fixed. An unattributed failure blocks handoff.

6. **Hand off.** Report the lane verdicts, gate commands with results, findings fixed, and any check skipped. Skipping a gate takes the user's explicit go-ahead in this session, with the evidence and the residual risk stated.

Complete when both lanes return `pass`, the gates are green on the final tree, every failure is fixed or lies outside the changed paths, and skips are disclosed. Otherwise report blocked.
