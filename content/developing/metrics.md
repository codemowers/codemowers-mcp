# Application metrics

Use metrics for alerting: measure symptoms such as failure rates, latency and stalled work, with thresholds tied to actionable operational needs. Use `developing/logging` for the context needed to triage an alert.

- Expose Prometheus metrics through the chosen framework's integration or client library. Use one shared, thread-safe registry for all application execution in the container; a scrape must not represent only the worker that answered.
- Serve metrics and health checks on an internal listener separate from the public application listener. Do not register these routes on the public listener or expose their port through public Ingress or externally exposed Services. Scrape transport and monitoring discovery belong to `deploying/observability`.
- Exclude sensitive data from names, labels, values, help text and exemplars: no identities, emails, IP addresses, credentials, tokens, bodies, raw URLs or query strings. Use bounded categories such as route templates, operations and status codes.
- Measure request latency, request/response sizes, failures and useful business activity such as import throughput. Bound label cardinality. Leave container memory and filesystem usage to runtime monitoring.
