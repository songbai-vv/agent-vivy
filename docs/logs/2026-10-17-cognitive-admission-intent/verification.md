# Verification

Environment: Linux cloud sandbox (root), Go 1.26.4, Laputa sibling at the
pinned `4b2bec2`. `just` and PowerShell are unavailable; recipes were run as
their underlying commands.

## Status of #32 C1 and C2 at the start

- C1: already fixed by `2f4b8f6`. Every load→save path in
  `cognitive_service.go` carries the loaded version under `cogStateMu`. The
  audit's interleaving is covered by
  `TestCognitivePolicyWritePreservesAcceptedSourceHigh`.
- C2: present. `cognitiveAttempt` saved nothing between
  `ensureCognitiveSupervisor` and `StartCognitiveWorkflow`.

## New tests (`internal/runtime/cognitive_admission_intent_test.go`)

The fixture `cognitiveCrashAfterAdmission` wraps the real SQLite
`SnapshotStore` and fails the first cognitive state save made after a workflow
run exists. That is the process-loss point C2 describes.

| Test | On this branch | On the previous `cognitive_service.go` (plus the three fields so it compiles) |
|---|---|---|
| `TestCognitiveCrashAfterAdmissionAdoptsCompletedWindow` | pass | fail: `{Run:false Reason:foreground_busy}`. The unrecorded supervisor blocks evolution. |
| `TestCognitiveCrashAfterAdmissionSurvivesSupervisorReplacement` | pass | fail: `UNIQUE constraint failed: runs.id` on `run_cognitive_supervisor_1` |
| `TestCognitiveCrashAfterAdmissionRetriesFailedWindowOnce` | pass | fail: `foreground_busy`, so no retry |
| `TestCognitiveIntentWithoutAdmissionIsDropped` | pass | fail: the intent is never cleared |
| `TestCognitiveSupervisorSkipsUnsavedSequence` | pass | fail: `UNIQUE constraint failed: runs.id` |

Existing suites: `go test ./internal/runtime -run Cognitive -count=1` → ok;
`go test ./internal/app -run 'Cognitive|Evolution|Diva|DIVA' -count=1` → ok.

## Gates

| Command | Result |
|---|---|
| `gofmt -l` (tracked and new `.go` files) | clean |
| `go vet ./...` | clean |
| `go test -p 1 -timeout 35m -skip '^TestMemoryLoop' ./...` | exit 0, 84 packages ok, no failures |
| `go test -race -timeout 35m ./internal/runtime -count=1` | ok (947s), no `DATA RACE` |
| PostgreSQL 16 (local cluster): `VIVY_POSTGRES_TEST_DSN=… go test -p 1 ./internal/storage/postgres ./internal/storage/migrations -count=1` | 182 tests passed, 0 skipped, 0 failed |
| DIVA recall: bash replay of `scripts/test-diva-recall.ps1` (`generate-default` for `recipes/diva.vivy.yml`, overlay, `-tags vivy_diva_integration`) | all 6 required tests passed, none skipped |
| Memory loop: bash replay of `scripts/test-memory-loop.ps1` | step 1 (`TestMemoryLoop` outside `internal/app`) ok; DIVA `sdk pack` and `inspect-artifact` ok; `sdk/internal` overlay tests ok; `internal/app` overlay pass ok (37 passed). Skips: two subprocess helpers, which run only as children, and `TestMemoryLoopCrashC06CanonicalBeforeDerivedIndex`, which needs a separate C06 diagnostic overlay that the official script does not supply either. |

`TestDeliverableReadSnapshotWriteFailure`: run as user `nobody` from a
compiled test binary, it passes. It now skips under root (as well as on
Windows), because root bypasses the `chmod 0500` the fixture relies on.

Not run: Windows-native `just ci`, and the UI gates (no UI source changed in
this delivery).
