# Deployment observability

Use metrics for alerting, logs for triage and traces for performance tuning. Traces help locate latency and bottlenecks across service calls; correlate them with logs and metrics when investigating a slow operation. Use the platform’s available tracing pipeline, keep collection configurable, and exclude credentials and sensitive payloads.

Application instrumentation belongs to `developing/metrics`, `developing/health-checks` and `developing/logging`.

## Prometheus transport decision

Ask the user to choose **internal HTTP** or **HTTPS with TLS** before configuring scraping. Follow the decision-recording rules in `developing/general`. Explain that HTTP is unencrypted, while HTTPS requires certificates, verified CA/hostname trust and renewal/reload handling. Inspect policy constraints before presenting the choice; resolve conflicts with an existing decision rather than silently changing it.

Record the transport and, for HTTPS, its nonsecret certificate/trust provisioning approach. Do not infer metrics TLS from application TLS or fall back to HTTP when certificates are missing. Keep listener and scrape settings consistent with the decision; both modes remain internal.

## Monitoring integration

- Discover the installed monitoring APIs and controller. Use its cluster-specific resource guidance; without an operator, use the established Prometheus configuration mechanism. Do not install monitoring infrastructure through the application chart.
- Detect optional resources independently with Helm capabilities, for example `monitoring.coreos.com/v1/PodMonitor` and `monitoring.coreos.com/v1/PrometheusRule`. Match release-specific workload labels and the monitoring system's discovery selectors. Keep application metrics available when these APIs are absent.
- Follow the unset Helm defaults in `deploying/general`; provide explicit deployment values for required settings that policies or resource defaults do not supply. Inspect admitted configuration and verify scraping succeeds.
- Configure health probes on the internal listener and collect stdout/stderr through the platform log pipeline.
