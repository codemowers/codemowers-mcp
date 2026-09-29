# Developing Go applications

Apply `developing/general` for configuration and lifecycle.

- Propagate request/job cancellation through contexts and stop background goroutines during shutdown. Bound concurrent work and connection pools.
- Handle expected errors at the operation that can recover; return unexpected errors with context. Do not ignore errors or use broad panic recovery to keep failed work running.
- Use one metrics registry across handlers and workers in the process; follow `developing/metrics` for its listener. Frameworks such as Gin are optional integrations, not deployment requirements.
