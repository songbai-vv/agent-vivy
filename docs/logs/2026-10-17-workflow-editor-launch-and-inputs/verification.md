# Verification

Environment: Linux cloud sandbox, Go 1.26.4, Node 22, pnpm 10.28. `just` and
PowerShell are unavailable here, so the `just ci` recipes were run as their
underlying commands (listed below). The Laputa sibling checkout was prepared
at the pinned commit `4b2bec2` from `laputa-source.lock.json`, as
`scripts/ensure-laputa.ps1` does. `ui/.npmrc` points at
`registry.npmmirror.com`, which this sandbox cannot reach, so the frozen
lockfile was installed with `--registry=https://registry.npmjs.org/`
(no file changed).

## Reproduction before the fix

V2, backend admission (temporary test calling `validateINOFYDefinition`,
deleted after use):

```text
config.inputs (UI today) -> err=runtime: node "summary" config: json: unknown field "inputs"
node.inputs (fixed)      -> err=<nil>
```

V1 was already reproduced in `docs/logs/2026-09-29-inofy-cutover/` (`-32603`
when the last-viewed run is stale).

The new tests were run against the old sources and fail there:
11 `face-bridge` cases (stale/foreign/missing parent, non-active statuses,
no active session) and all 3 `NodePanel` cases.

## Gates

| Command | Result |
|---|---|
| `node ui/scripts/stage-ui-assembly.mjs` | regenerated `ui/src/generated/assembly.ts` (one source-hash line) |
| `npx vitest run src/generated/ui/vivy-workflow` (in `ui/`) | 5 files, 41 tests passed |
| `npx tsc --noEmit` (in `ui/`) | clean |
| `npx vitest run` (in `ui/`) | 93 files, 705 tests passed |
| `node scripts/check-i18n-completeness.js` | pass |
| `node --test scripts/check-i18n-cross-face.test.js` | pass |
| `node scripts/check-i18n-cross-face.js` | pass |
| `npx vite build` (in `ui/`) | pass |
| `go test ./internal/bundlecontract/... ./sdk/internal/assembly/... ./ui/...` | ok |

Not run: the full Go `test`/`vet` recipes (no Go source changed in this
delivery) and the Windows-only `diva-recall-test`.

## Real-path smoke (split pair, `127.0.0.1:3015`)

Backend `go build ./cmd/vivy` on `:8787` with SQLite in a scratch directory
(not `data/`), Vite on `:3015`, and a local OpenAI-compatible mock on
`:11434` (`providers.active: vllm`). The mock answers immediately, except
that a user message containing `HOLD` keeps its stream open for 120 s so the
parent run stays active. Driven by Playwright (Chromium):

1. Sent `hello`; the turn completed. On **Workflows**, opened `smoke-wf`,
   selected node `begin`, entered `{"note":{"literal":"from-the-editor"}}`,
   and saved. Header showed `completed`; **Run** was disabled with the
   "active session run" tooltip.
2. Sent `HOLD this turn`, returned to **Workflows** in-app, and clicked
   **Run**. The workflow run went `run_admitted → node_completed(begin) →
   run_succeeded`, and the mock received the child task call.
3. In the scratch database, the saved draft and the admitted run descriptor
   both hold the binding at node level:
   `{"id":"begin",...,"config":{"task":"smoke-wf"},"inputs":{"note":{"literal":"from-the-editor"}}}`.

Temporary scripts, the mock, the scratch database and the servers were
removed or stopped after the run.
