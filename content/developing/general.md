# General application development

Apply these conventions to new applications. Assess existing applications case by case using the legacy-porting guidance in `building/general`; do not assume that moving an application to Kubernetes includes rewriting its architecture.

## Decisions and configuration

- Keep one version-controlled codebase per application with many deployments. Do not maintain separate environment-specific source copies; use release configuration for differences. A monorepo can contain multiple applications, and reusable code belongs in declared dependencies.
- Read existing project decisions before choosing infrastructure or authentication. Ask only for unresolved choices: authentication architecture in `developing/architecture`, native versus portable dependencies in `deploying/storage`, and Prometheus transport in `deploying/observability`. Record each choice, rationale and nonsecret configuration in the project's `AGENTS.md` or equivalent instructions; preserve unrelated content and update decisions when the user changes them. Resolve conflicts with platform requirements before implementing dependent work.
- Configure endpoints, credentials and environment-specific behavior at runtime. Treat compatible backing services as replaceable attachments: switching their endpoint or credentials should not require a source change. Use consistent environment-variable names across environments and keep each default in one layer: application, image or deployment. For complex settings, use a configuration library that combines files and environment variables. Helm environment mappings and rollouts belong to `deploying/general`.
- Keep code at its sole call site instead of adding single-use helper wrappers. Extract helpers for actual reuse; framework handlers, callbacks and entrypoints are not wrappers.
- Write readable YAML with indented block mappings and sequences. Retain explicit empty collections where meaningful, and embedded JSON only where the consuming API requires it. End text files with a newline. Ignore generated files and secrets as described in `building/general`.
- Follow `developing/git` for Git configuration, synchronizing branches and committing changes.

## Failures and lifecycle

- Handle only known, recoverable failures at the operation that can recover. Keep try blocks narrow; do not wrap an entire request, job, loop or startup sequence. Avoid blanket exception handlers; where catch syntax is untyped, identify the expected error and rethrow every unrecognized one.
- Let unexpected failures fail the request, job or process through the runtime/framework. Never turn them into successful acknowledgements, empty results, arbitrary retries or catch-and-log-and-continue behavior. Use cleanup constructs such as `finally`, defer, context managers or RAII instead of broad catches.
- Attempt the required database operation directly. Use constraints, atomic operations and specific error handling instead of speculative preflight queries. Fetch required fields once and handle a missing row; do not precede the fetch with an existence check. Business queries using EXISTS remain legitimate. Health endpoints follow `developing/health-checks`.
- Start promptly. On SIGTERM, stop accepting requests/jobs and drain, complete or return in-flight work within the termination deadline. Also tolerate abrupt termination: make repeated jobs idempotent or transactional so retries do not duplicate side effects.

Use `building/general` for the container process model, `deploying/workloads` for scheduling and resource budgets, and `deploying/networking` for listeners and TLS. Language sections add runtime-specific details; their frameworks are examples.
