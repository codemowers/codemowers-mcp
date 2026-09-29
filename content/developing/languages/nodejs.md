# Developing Node.js applications

Apply `developing/general` for configuration, failures and shutdown, and `building/languages/nodejs` for the production entrypoint.

- Keep blocking or CPU-intensive work off the request event loop. Bound outstanding promises, queue consumption and connection pools; an async function does not make synchronous work nonblocking.
- Await or explicitly supervise asynchronous work so failures reach the request/job owner. Do not consume arbitrary promise rejections with catch-and-continue handlers.
- Close listeners and drain tracked work on SIGTERM. Choose the framework for application needs and keep health/metrics on the internal listener in `developing/metrics`.
