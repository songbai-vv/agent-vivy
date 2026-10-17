# Acceptance

1. Start the split pair (`just dev`), send a chat message and let the turn
   finish. Open **Workflows**. The **Run** button is disabled and its tooltip
   says workflow runs start inside an active session run. (Before: it was
   enabled and clicking it failed with `-32603`.)
2. In the editor, open a workflow id, select a node and enter inputs such as
   `{"note": {"literal": "hello"}}`, then **Save draft**. Reopen the draft:
   the inputs are still shown.
3. While a chat turn is still running, **Run** is enabled. Clicking it starts
   a workflow run that reaches `run_succeeded` in the **Runs** tab. (Before:
   a node with authored inputs was rejected with `unknown field "inputs"`.)
