# Eino core v0.9.13 → v0.9.21 and the upgrade recovery policy

## What changed

- `go.mod` / `go.sum`: `github.com/cloudwego/eino` v0.9.13 → v0.9.21. No other
  module changes; v0.9.21 adds no new requirements. `go get` and `go mod tidy`
  also reorder unrelated `require` lines, so the three-line pin change was
  applied by hand and checked with `go build ./...`.
- Fixes the upstream data race recorded as `EINO-TOOLSNODE-ERR-RACE` and
  `PG4-EINO-TOOLS-RACE`. In v0.9.13 the parallel enhanced-tool converter in
  `(*ToolsNode).Stream` wrote the function-scope `err`
  (`compose/tool_node.go:1253`); v0.9.21 uses a local `convertErr`
  (`compose/tool_node.go:1365`). No Vivy workaround was needed.
- **Upgrade policy** (ADR-008 amendment, `docs/AGENT-VIVY-ARCHITECTURE-V0.md`):
  an Eino pin change is an ordinary restart. No drain, no checkpoint
  conversion, no version bypass. Recovery closes a run paused across the
  version change (approval, question, Plan review, child approval) before the
  server listens, and settles Work state tied to it.
- **Fix: stranded Plan review.** `Service.failUnrecoverable`
  (`internal/runtime/service.go`) now cancels a pending Plan review that the
  failed run originated. Before, `emitTerminal` cancelled reviews only for
  runs registered in this process (`runSessions`), which a recovered run is
  not, so the review stayed `pending` on a run that could never continue.
  `cancelPlanReviewForRun` only acts on a pending review whose origin is that
  run, so the call is a no-op everywhere else.
- Tests:
  - `TestServiceRecoverApprovalAcrossEngineUpgradeFailsClosed`
    (`internal/runtime/recovery_test.go`): a run paused on an approval is
    `failed` right after `Recover`, ends with exactly one `run.failed` with
    the restart-recovery message, and a late approval never starts the tool.
    `restartService` now delegates to `restartServiceWithEngine`.
  - `TestPlanGoalPendingReviewAcrossEngineUpgradeSettles`
    (`internal/app/plan_goal_e2e_test.go`): real App, crash during a pending
    Plan review, restart under another engine version → run `failed`, review
    `cancelled`, Plan still active.
  - `TestPlanGoalPausedRoundAcrossEngineUpgradeSettles`: Goal round paused on
    `ask_user` → run `failed`, Goal `blocked` with reason
    `goal round failed`, no extra round admitted.
- Test pins `pinnedEinoVersion` in `internal/app/realsmoke_test.go` and
  `internal/codeface/launch_test.go`, and the pin comment on
  `EngineFactory` in `internal/runtime/engine.go`, follow the new version.
- Boards: `EINO-CORE-UPGRADE` (from `docs/DEFER.MD`),
  `EINO-TOOLSNODE-ERR-RACE` (from `docs/TODO.md` and `docs/DEFER.MD`) and
  `PG4-EINO-TOOLS-RACE` (from `docs/TODO.md`) moved to `docs/COMPLETE.MD`.

## Decision record

`docs/DEFER.MD` held a user decision (2026-09-24) to stay on v0.9.13 until a
checkpoint drain or migration policy existed, because a version mismatch
fails a paused run but may not settle its Goal/Plan state. That conflict was
raised before committing. On 2026-10-17 the user chose to upgrade and asked
for the migration policy to be updated in full. Checking the named risk found
one real gap (Plan review left pending; fixed here) and confirmed the Goal
path already settles.

## Eino capability check

Upstream diff v0.9.13 → v0.9.21 (non-test files) touches `adk/cancel.go`,
`adk/interrupt.go`, `adk/react.go`, `adk/turn_loop.go`, ADK middlewares,
`compose/checkpoint.go`, `compose/graph*.go`, `compose/stream_concat.go`,
`compose/tool_node.go`, `internal/core/interrupt.go`,
`components/model/callback_extra.go`, `schema/message.go` and
`schema/agentic_message.go`. Relevant to Vivy:

- **Checkpoint encoding changed.** `adk/interrupt.go` adds
  `InfoDataSourceInterruptID` and compacts `ChatModelAgentInterruptInfo.Data`
  into the interrupt state. v0.9.21 still reads v0.9.13 bytes, but a v0.9.13
  binary reading v0.9.21 bytes would lose that data. Vivy's
  `VersionedCheckpointStore` already rejects any engine-version mismatch in
  both directions (`ErrCheckpointEngineVersionMismatch`). This delivery keeps
  that guard and adds no converter or bypass.
- `PromptTokenDetails` gains `CacheWriteTokens` (additive).
- `adk.StreamCanceledError` gains `Is` for `core.ErrStreamCanceled`
  (additive).
- The APIs `EngineFactory` names (`adk.NewChatModelAgent`, `adk.NewRunner`,
  `Runner.Run/Query`, `Runner.ResumeWithParams`) compile unchanged.

## Not done

- No drain tooling and no checkpoint conversion, by policy. A paused turn is
  lost on upgrade; operators keep it by answering it before upgrading.
- Child runs paused on an approval follow the existing
  `child checkpoint not readable` recovery branch (`child.failed`); no new
  version-specific test was added for them.
- eino-ext components stay at their current pins (claude v0.1.25, openai
  v0.1.13, tool/mcp v0.0.9). The Agentic-model migration in
  [ProjectViVy/agent-vivy#73](https://github.com/ProjectViVy/agent-vivy/issues/73)
  is separate.
- Historical docs that cite v0.9.13 line numbers are left as written.
