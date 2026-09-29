# Application logging

Use logs for triage: record enough context to investigate what happened and why. Use `developing/metrics` for alerting on measurable symptoms.

- Disable ANSI color codes and terminal control sequences in Kubernetes logs, including framework and library output. Use the selected logger’s plain-output configuration; do not force terminal colors in containers. Colored output may remain an option for interactive local development.
- Write to stdout/stderr; let the runtime collect, rotate and retain logs. Do not write container-local log files or duplicate runtime timestamps.
- Choose plain text or structured JSON for the application's consumers; plain text alone is not a defect. Use meaningful severity levels; exclude credentials and sensitive payloads.
- Measure repetitive events with metrics instead of producing a log record for every occurrence; see `developing/metrics`.
