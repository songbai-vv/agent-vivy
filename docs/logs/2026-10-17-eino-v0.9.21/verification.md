# Verification

Environment: Linux cloud sandbox (running as root), Go 1.26.4 with the race
toolchain. `just` and PowerShell are unavailable, so the `just ci` recipes
were run as their underlying commands (below). The Laputa sibling checkout
was prepared at the pinned `4b2bec2` from `laputa-source.lock.json`.

## Race: before and after

```text
# v0.9.13 (base 5715950)
go test -race ./internal/runtime -run '^TestPlanGoalProbeReportFencesLaterToolInSameBatch$' -count=1
WARNING: DATA RACE
      .../eino@v0.9.13/compose/tool_node.go:1253
--- FAIL: TestPlanGoalProbeReportFencesLaterToolInSameBatch
    testing.go:1712: race detected during execution of test

# v0.9.21
ok  	agent-vivy/internal/runtime	2.682s
```

Upstream source check (module zips from proxy.golang.org):
`eino@v0.9.13/compose/tool_node.go:1253` is
`ret[index].UserInputMultiContent, err = tr.ToMessageInputParts()`;
`eino@v0.9.21/compose/tool_node.go:1365` is
`parts, convertErr := tr.ToMessageInputParts()`.

## Dependency change

`go get github.com/cloudwego/eino@v0.9.21` added no new modules, but it and
`go mod tidy` both reorder unrelated `require` lines (`go mod tidy` also
dropped local plugin requirements). The pin was applied as a three-line
`go.mod` / `go.sum` edit and confirmed with `go build ./...` and
`go list -m github.com/cloudwego/eino` → `v0.9.21`. No plugin or face module
pins Eino in its own `go.mod` / `go.sum`.

## Upgrade-policy tests

Each new test was shown to discriminate:

| Test | Passes on this branch | Fails when |
|---|---|---|
| `TestServiceRecoverApprovalAcrossEngineUpgradeFailsClosed` | yes | the restart uses the same engine version (run completes) |
| `TestPlanGoalPendingReviewAcrossEngineUpgradeSettles` | yes | same engine version (run stays active), or without the `failUnrecoverable` fix (review stays `pending`) |
| `TestPlanGoalPausedRoundAcrossEngineUpgradeSettles` | yes | same engine version (run stays active) |

The Goal test passes without the fix: recovery already rebuilds the Goal
mapping (`rememberRecoveredGoalRun`) before failing the run. The existing
`TestPlanGoalIntegratedPendingReviewRecovery` and every `TestServiceRecover*`
test still pass.

## Gates

| Command | Result |
|---|---|
| `gofmt -l` over tracked and new `.go` files | clean |
| `go vet ./...` | clean |
| `go test -run '^$' -tags vivy_headless ./cmd/vivy ./cmd/vivy-code ./ui` | ok |
| `plugin-ci` loop (`go vet` + `go test` in each of 22 `plugins/*` and `faces/*` modules) | all pass |
| `go test -p 1 -timeout 35m -skip '^TestMemoryLoop' ./...` | 82 packages ok. Fails: `internal/runtime` (only `TestDeliverableReadSnapshotWriteFailure`, see below) and `sdk/internal` (4 pack tests reporting "source changed during build": files under `internal/` were being edited while the run was in progress). Both packages were rerun clean below. |
| `go test -race -timeout 35m ./internal/runtime -count=1` | no `DATA RACE` anywhere in the package (before: race at `tool_node.go:1253`); the only failure is `TestDeliverableReadSnapshotWriteFailure` |
| `go test -p 1 -skip '^TestMemoryLoop' ./internal/app ./internal/runtime ./sdk/internal -count=1` (clean rerun with no edits in progress) | `internal/app` ok, `sdk/internal` ok, `internal/runtime` fails only `TestDeliverableReadSnapshotWriteFailure` |

`TestDeliverableReadSnapshotWriteFailure` fails in this sandbox on both the
base commit `5715950` (checked in a temporary worktree) and this branch. The
test makes a directory unwritable with `chmod 0500`, which root ignores, so it
is environmental and unrelated to this change.

Not run: Windows `just ci` (including `diva-recall-test` and the memory-loop
second pass), PostgreSQL conformance, and the UI gates (no UI source changed
in this delivery).
