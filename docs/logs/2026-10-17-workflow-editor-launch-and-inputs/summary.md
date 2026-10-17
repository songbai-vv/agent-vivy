# Workflow editor: launch parent and node inputs

Closes items V1 and V2 of the upstream tracker
[ProjectViVy/agent-vivy#21](https://github.com/ProjectViVy/agent-vivy/issues/21)
(workflow follow-ups after the INOFY cutover). Both were recorded as known
gaps in `docs/logs/2026-09-29-inofy-cutover/acceptance.md` and
`verification.md`.

## What changed

- **V1: launch parent.** The editor and the Workflows list bound
  `state.currentRun` as `parent_run_id`. That is the last-viewed run, which is
  often finished or belongs to another session, so `inofy.startRun` failed
  with `-32603`. The backend admits a workflow only under an *active* run of
  the caller's session (`currentChildAuthorizer` in
  `internal/runtime/child_sessions.go`; session check in
  `internal/runtime/workflow_product.go`). The new `launchParentRunId` in
  `face-bridge.ts` returns the current run only when it is `active` and
  belongs to the active session. The bridge uses it for `parent_run_id`, and
  `WorkflowPage` uses it to enable the Run buttons, so Run is disabled with
  the existing "needs an active session run" tooltip instead of failing.
  The cutover log left two options open: bind the session's active run, or
  resolve the parent backend-side. This takes the first, because it changes
  no RPC or runtime contract.
- **V2: node inputs.** The node panel wrote authored bindings to
  `config.inputs`. Bindings are the node-level `inputs` field; the
  `vivy.child-task@1` admission decodes `config` with
  `DisallowUnknownFields`, so any node with authored inputs was rejected at
  run time (`node "…" config: json: unknown field "inputs"`). The panel now
  reads and writes `node.inputs`. A draft saved before this fix still shows
  its `config.inputs` and moves them to `node.inputs` on the next commit.
  An empty object clears `node.inputs`.

Files: `plugins/vivy-workflow/ui/vivy-workflow/src/face-bridge.ts`,
`WorkflowPage.tsx`, `editor/NodePanel.tsx`, tests `face-bridge.test.ts` and
new `editor/NodePanel.test.tsx`; `ui/src/generated/assembly.ts` regenerated
by `stage-ui` (source hash only).

## Not done

- No backend or RPC change. Resolving the parent server-side remains a
  possible future product decision; it is not needed for correctness.
- No automatic migration of stored drafts that hold `config.inputs`; they are
  repaired when the node's inputs are next committed in the editor.
- The external issue #21 is not updated from this branch.
- #21 V3–V9 are unchanged.
