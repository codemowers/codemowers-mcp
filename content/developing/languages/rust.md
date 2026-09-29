# Developing Rust applications

Apply `developing/general` for failure handling and lifecycle. Choose a framework and async runtime for the application's libraries and workload.

- Use typed errors and propagate failures with context.
- Keep blocking work out of async request paths. Bound tasks and connections; propagate cancellation and join application tasks during shutdown.
- Validate required configuration at startup; listener requirements belong to `deploying/networking`.

## Example: Rocket

Rocket applications may use `rocket::build()` with mounted handlers and configure through `Rocket.toml` or `ROCKET_*` variables. Set the deployment listener explicitly rather than inheriting a development bind address.

When using `rocket_prometheus`, attach instrumentation and export from the same registry, with the internal listener required by `developing/metrics`. Other frameworks use their equivalent hooks. See `building/languages/rust` for packaging guidance and `building/general` for the lolcatz Dockerfile examples.
