# Handoff — 2026-10-17 upstream-issue session

Branch `claude/pensive-shannon-4kk6fx` on the fork `songbai-vv/agent-vivy`.
It carries two deliveries:

1. `docs/logs/2026-10-17-workflow-editor-launch-and-inputs/`: closes
   [ProjectViVy/agent-vivy#21](https://github.com/ProjectViVy/agent-vivy/issues/21)
   items V1 (stale launch parent) and V2 (node inputs written to
   `config.inputs`).
2. This directory: Eino v0.9.21, the upgrade recovery policy (ADR-008
   amendment), and the Plan-review settlement fix.

No pull request was opened, and no upstream issue was edited from this session.

## For whoever picks this up

### Bookkeeping on the upstream repository

- Open a PR from this branch to `ProjectViVy/agent-vivy` when the owner wants
  it.
- Tick #21 V1 and V2 on the upstream issue after merge.
- Mention on [#73](https://github.com/ProjectViVy/agent-vivy/issues/73) that
  core Eino is now v0.9.21, the version its S0 plan targets. Its checkpoint
  concern (rollback across engine versions) is now answered by the ADR-008
  amendment: fail closed, no bypass.

### Checks this sandbox could not run

Linux cloud container, running as root, with no PowerShell or `just`:

- Full `just ci` on Windows, including `diva-recall-test` and the
  memory-loop second pass (`scripts/test-memory-loop.ps1`).
- PostgreSQL conformance (`just test-postgres`).
- `TestDeliverableReadSnapshotWriteFailure` fails here on both the base
  commit `5715950` and this branch, because root ignores the test's `chmod
  0500`. It needs a non-root run.
- `ui/.npmrc` points at `registry.npmmirror.com`, which cloud sessions cannot
  reach. Install with `--registry=https://registry.npmjs.org/` there; the
  lockfile is registry-neutral.

### Next upstream items that look actionable

From the 2026-10-17 triage of open upstream issues:

- **#32 C1**: stale state written with a freshly fetched compare-and-set
  version (`internal/runtime/cognitive_service.go`). Reproduce first.
- **#32 C2**: admission happens before the durable intent, so completed
  results become retries (`internal/runtime/inofy_admission.go`, admission
  and recovery path). Reproduce first.
- **#21 V3–V9**: need product decisions before coding (governed workflow
  invocation, switch/repeat subset, typed outputs, wait/resume, persistent
  child sessions).
- `DEEPSEEK-REASONING-CONTENT` (`docs/TODO.md`) is still open. It is an
  eino-ext adapter gap, not affected by the core bump.

Waiting on others: #72 and #64 (open PR #69 and a Laputa PR), #58 (PRs #70 and
#71), #31 (Wails branches, a missing archive tag, Windows audio). Everything
else is a proposal awaiting an owner decision.

### Known leftovers, deliberately not changed

- Historical docs and research notes cite Eino v0.9.13 line numbers
  (`docs/eino-capability-verify.md`, `docs/architecture/PLAN-GOAL-PREDESIGN.md`
  and others). They record what was inspected at the time.
- Checkpoint labels such as `"v0.9.13"` in runtime tests are arbitrary
  engine-version strings and do not need to track the pin.
- Stored workflow drafts that still hold `config.inputs` are repaired only
  when that node's inputs are next committed in the editor; there is no bulk
  migration.
