# Upstream drafts — ready to paste

This session cannot write to `ProjectViVy/agent-vivy`, so these are drafted
here for the owner to post. The repository has no pull request template.

## Pull request

- **From:** `songbai-vv/agent-vivy:claude/pensive-shannon-4kk6fx`
- **To:** `ProjectViVy/agent-vivy:main`
- **Title:** Workflow editor fixes, Eino v0.9.21 with fail-closed upgrade policy, cognitive admission intent

**Body:**

```markdown
Three deliveries, each with its own iteration log under `docs/logs/`.

## 1. Workflow editor: launch parent and node inputs (#21 V1, V2)
- Run starts a workflow only under the active session's in-flight run; with no
  such run the button is disabled instead of failing with `-32603`.
- Node bindings are written to `node.inputs` (the runtime rejected
  `config.inputs` as an unknown field); old drafts are repaired on the next edit.
- Log: `docs/logs/2026-10-17-workflow-editor-launch-and-inputs/`

## 2. Eino v0.9.21 and the upgrade recovery policy
- Pins `github.com/cloudwego/eino v0.9.21`, which fixes the upstream data race
  in `ToolsNode.Stream` (`EINO-TOOLSNODE-ERR-RACE`, `PG4-EINO-TOOLS-RACE`).
- ADR-008 amendment: an engine change is an ordinary restart. No drain, no
  checkpoint conversion, no version bypass; paused runs fail closed and their
  Work state settles.
- Fix: a Plan review whose run failed during recovery is now cancelled instead
  of staying `pending`.
- Log: `docs/logs/2026-10-17-eino-v0.9.21/`

## 3. Cognitive admission write-ahead intent (#32 C2)
- `cognitiveAttempt` saves the exact parent and operation key before starting
  the workflow, so a crash before the state save no longer re-admits a
  completed window or leaves evolution stuck at `foreground_busy`.
- #32 C1 was already fixed by `2f4b8f6`.
- Log: `docs/logs/2026-10-17-cognitive-admission-intent/`

## Verification
Linux sandbox: full Go suite, `-race` on `internal/runtime`, PostgreSQL
(182 tests), DIVA recall and the memory-loop second pass, UI typecheck and
tests, and a browser smoke on the split pair. Windows-native `just ci` was not
run. Details are in each log's `verification.md`.
```

## Comment on #21

```markdown
V1 and V2 are fixed on `songbai-vv/agent-vivy` branch `claude/pensive-shannon-4kk6fx`
(commit "fix(workflow-ui): launch under the active run and author node inputs"):

- V1: launches bind the active session's in-flight run; Run is disabled when
  there is none.
- V2: bindings go to `node.inputs`; drafts holding `config.inputs` are moved
  on the next edit.

Browser smoke and runtime reproduction: `docs/logs/2026-10-17-workflow-editor-launch-and-inputs/`.
V3–V9 are unchanged.
```

## Comment on #73

```markdown
FYI: core Eino is now pinned at v0.9.21 on `songbai-vv/agent-vivy`
(`claude/pensive-shannon-4kk6fx`), the version S0 targets.

The checkpoint concern in the review revisions is answered by an ADR-008
amendment (`docs/AGENT-VIVY-ARCHITECTURE-V0.md`): an engine change is an
ordinary restart, with no converter and no version bypass. A run paused across
the change fails closed during recovery, and its Plan review or Goal round
settles. Tests: `TestServiceRecoverApprovalAcrossEngineUpgradeFailsClosed`,
`TestPlanGoalPendingReviewAcrossEngineUpgradeSettles`,
`TestPlanGoalPausedRoundAcrossEngineUpgradeSettles`.

The eino-ext component pins (claude v0.1.25, openai v0.1.13) are unchanged,
so the AgenticModel migration itself is still open.
```

## Comment on #32

```markdown
Status update for the VIVY kernel findings:

- **C1** was already fixed on main by `2f4b8f6` ("serialize cognitive state
  admission and policy updates"): every write carries the loaded version under
  `cogStateMu`, covered by `internal/runtime/cognitive_concurrency_test.go`.
- **C2** is fixed on `songbai-vv/agent-vivy` branch
  `claude/pensive-shannon-4kk6fx`: a write-ahead intent (exact parent and
  operation key) is saved before the workflow starts, and the next attempt
  adopts and settles that run. Tests:
  `internal/runtime/cognitive_admission_intent_test.go`. Log:
  `docs/logs/2026-10-17-cognitive-admission-intent/`.
```
