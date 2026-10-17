# Acceptance

- `go list -m github.com/cloudwego/eino` prints `v0.9.21`.
- `go test -race ./internal/runtime -run '^TestPlanGoalProbeReportFencesLaterToolInSameBatch$' -count=1`
  passes. On v0.9.13 it fails with `DATA RACE` at
  `eino@v0.9.13/compose/tool_node.go:1253`.
- Leave a turn waiting for approval, a question or a Plan review, stop the
  server, and start a binary built with a different Eino version. The turn
  shows as failed with "interrupted by a server restart and could not be
  recovered"; the pending tool never runs. A pending Plan review shows as
  cancelled while Plan mode stays on. A Goal whose round was paused shows as
  blocked ("goal round failed") and does not start another round.
- New runs behave as before.
