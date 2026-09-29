# Application health checks

- Serve HTTP health/readiness endpoints such as `/health` and `/ready` on the internal listener defined in `developing/metrics`, separate from public application traffic. Wire them into Compose’s `healthcheck` directive for local use and the corresponding Kubernetes probes. Use a check command available in the image; do not assume a shell or curl exists in a minimal runtime.
- Keep checks local, cheap and deterministic. Do not query databases or other backing services; a dependency outage must not trigger mass application restarts.
- Readiness reports whether initialization is complete and the process accepts work; mark it unavailable when draining. Return a success status only while ready and a failure status otherwise.
- Use a startup probe only when initialization needs a separate allowance before ordinary probes begin. Avoid liveness probes by default; add one only for a demonstrated local failure that restarting can repair. Check process progress, not dependency availability.
