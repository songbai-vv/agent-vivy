# Acceptance

- Stop Vivy right after an evolution run has been admitted (before its state
  is saved) and start it again. The evolution status shows the window as
  processed once the run completes; no second evolution run appears for the
  same window, and the status does not stay stuck at "foreground busy".
- A window whose run genuinely failed is still retried once per wake, up to
  three attempts, as before.
