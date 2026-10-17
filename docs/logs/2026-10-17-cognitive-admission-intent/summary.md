# Cognitive admission: write-ahead intent (#32 C2)

Fixes finding C2 of
[ProjectViVy/agent-vivy#32](https://github.com/ProjectViVy/agent-vivy/issues/32):
"Admission recovery can replay a completed cognitive window". Finding C1 of the
same issue (stale snapshot write) was already fixed by `2f4b8f6` before this
work, together with `internal/runtime/cognitive_concurrency_test.go`; no
change was needed for it.

## The bug

`cognitiveAttempt` (`internal/runtime/cognitive_service.go`) started the
evolution workflow first and saved `ActiveRunID` / `PendingThrough` only
afterwards. `ensureCognitiveSupervisor` had also changed the supervisor in
memory only. If the process stopped between those two points:

- The supervisor run was never recorded. The next attempt minted a new
  supervisor, and because workflow dedup is scoped to the parent run, the
  admitted window ran a second time under a new effect identity.
- The unrecorded supervisor stayed an active primary run, so automatic
  evolution reported `foreground_busy` until the next restart closed it.
- When dedup did return the earlier run, a *completed* run went through the
  failure branch (`Attempt++`) instead of advancing the watermark.
- After a crash between creating the supervisor run and saving the sequence,
  the next supervisor reused the same id and failed with a UNIQUE violation.

## What changed

- `cognitiveState` gains `IntentOpKey`, `IntentParentRunID` and
  `IntentThrough` (JSON, `omitempty`; older snapshots decode unchanged; no
  schema change).
- Before starting the workflow, the attempt saves the intent (exact parent and
  operation key) together with the supervisor fields.
- `reconcileCognitiveIntent` runs first on every attempt. It looks the intent
  up with `GetWorkflowRevisionByOperation` and adopts the admitted run, or
  drops the intent when nothing was admitted.
- The existing settlement block moved unchanged into
  `settleCognitiveActiveRun`. After a start, the returned run is always
  adopted, and a terminal dedup result is settled immediately: completed
  advances the watermark, a transient failure counts one attempt, and
  cancelled or unknown outcomes block.
- `ensureCognitiveSupervisor` skips supervisor ids that already exist.
- The retry contract for genuinely failed runs is unchanged: a new attempt
  suffix (`a1`, `a2`), bounded at `cognitiveMaxAttempts`.

Eino capability check: not applicable. This is Vivy-owned INOFY workflow
admission and snapshot durability, with no model, tool or checkpoint
orchestration.

## Not done

- Laputa effect keys are unchanged; they stay stable because the same run is
  adopted rather than re-admitted.
- `DIVA-MEM-FAILED-WINDOW` stays open for its own remaining items
  (same-operation receipt recovery, the S10 crash matrix).
- Nothing reloads and retries on `ErrVersionConflict` across processes; within
  one process `cogStateMu` serializes every write.
