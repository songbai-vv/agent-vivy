# Handoff — 2026-10-17 upstream-issue session

Branch `claude/pensive-shannon-4kk6fx` on the fork `songbai-vv/agent-vivy`
carries three deliveries:

1. `docs/logs/2026-10-17-workflow-editor-launch-and-inputs/`:
   [ProjectViVy/agent-vivy#21](https://github.com/ProjectViVy/agent-vivy/issues/21)
   V1 (stale launch parent) and V2 (node inputs written to `config.inputs`).
2. This directory: Eino v0.9.21, the upgrade recovery policy (ADR-008
   amendment), and the Plan-review settlement fix.
3. `docs/logs/2026-10-17-cognitive-admission-intent/`:
   [#32](https://github.com/ProjectViVy/agent-vivy/issues/32) C2 (write-ahead
   cognitive admission intent). C1 turned out to be already fixed by
   `2f4b8f6`.

## Done since the first handoff

- #32 C2 fixed; C1 confirmed already fixed. The first handoff wrongly listed
  C1 as open.
- Checks earlier recorded as "could not run" have now run on Linux:
  PostgreSQL (182 tests pass), the DIVA recall gate (6/6), and the memory-loop
  second pass (all required tests pass). Results are in
  `docs/logs/2026-10-17-cognitive-admission-intent/verification.md`.
- `TestDeliverableReadSnapshotWriteFailure` passes as a non-root user and now
  skips under root, where the permission fixture cannot work.
- Upstream PR body and issue comments drafted in `upstream-drafts.md`.

## Still open

- **Post upstream** (owner action): open the PR and post the #21, #73 and #32
  comments from `upstream-drafts.md`. This session has no write access to
  ProjectViVy.
- **Windows-native `just ci`**: not runnable in a Linux sandbox. Every recipe
  it calls has been run here as its underlying command; a Windows run is still
  the formal gate.
- **#21 V3–V9**: need product decisions before coding (governed workflow
  invocation, switch/repeat subset, typed outputs, wait/resume, persistent
  child sessions).
- **`DEEPSEEK-REASONING-CONTENT`** (`docs/TODO.md`): blocked upstream.
  eino-ext `components/model/openai` v0.1.13 was still the newest release on
  2026-10-17.
- **`DIVA-MEM-FAILED-WINDOW`** (`docs/TODO.md`): its own remaining items
  (same-operation receipt recovery, the S10 crash matrix) are unchanged.
- Waiting on others: #72 and #64 (open PR #69 plus a Laputa PR), #58 (PRs #70
  and #71), #31 (Wails branches, a missing archive tag, Windows audio).

## Notes for the next session

- Cloud sandboxes cannot reach `registry.npmmirror.com` (`ui/.npmrc`). Install
  with `pnpm install --frozen-lockfile --registry=https://registry.npmjs.org/`;
  the lockfile is registry-neutral.
- Build needs the Laputa sibling at `../laputa`, pinned in
  `laputa-source.lock.json`.
- PostgreSQL 16 binaries live in `/usr/lib/postgresql/16/bin`. Run
  `initdb`/`pg_ctl` as the `postgres` user and set `VIVY_POSTGRES_TEST_DSN`.
- Do not edit files under `internal/` while `sdk/internal` tests run: they
  hash the source tree and fail with "source changed during build".
- Historical docs that cite Eino v0.9.13 line numbers, and the `"v0.9.13"`
  checkpoint labels in runtime tests, are intentionally unchanged.
- Stored workflow drafts with `config.inputs` are repaired only when that
  node's inputs are next edited in the editor.
