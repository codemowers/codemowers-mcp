# Developing Python applications

Apply `developing/general` for configuration, failures and lifecycle, and `building/languages/python` for packaging.

- Prefer async I/O with compatible libraries throughout the request path. Do not call blocking database, storage, image-processing or inference operations directly on the event loop. For synchronous applications, retain a single execution thread until concurrency can be introduced correctly.
- Justify worker threads with measurements and the actual runtime/library behavior. On GIL-enabled runtimes, CPU parallelism requires operations that release the GIL; importing a library alone proves nothing. Bound tasks, thread pools and connection pools; avoid unbounded executors.
- Follow the single-process contract in `building/general`; do not add uWSGI worker pools or process supervisors for new services.

## Example: Sanic

Sanic is one async HTTP option. Use `sanic server:app --host=:: --port=8000 --single-process` or `app.run(host="::", port=8000, single_process=True)` to disable its worker manager and auto-reloader. Its ordinary one-worker mode still includes a manager process. See [Running Sanic](https://sanic.dev/en/guide/running/running.html).
