# General container builds

## Reproducible images

- Pin base-image versions. Declare all application dependencies and required system libraries explicitly; do not rely on host installations. Commit lockfiles where supported and use locked/frozen installation that fails on manifest disagreement.
- Install dependencies and compile assets during image builds. Use multiple stages and copy only required runtime files into the final image. Ship a static executable where supported; otherwise include compatible runtime libraries.
- Keep `.dockerignore` at the actual build-context root. Exclude Git metadata, host dependencies/build outputs, reports, caches, editor files, local configuration, credentials and unrelated files; retain sources, manifests, lockfiles and required assets. Inspect the context when moving or adding a Dockerfile.
- Maintain `.gitignore` separately for generated artifacts and private local files. Ignore rules do not untrack files: inspect accidental tracked outputs before removing them from the index, preserving needed local files. Never copy registry credentials or SSH keys into image layers.
- Use `developing/local-development` for build tooling and `building/ci` for publication and promotion.

## Separate executable code from data

- Package ML model weights, map datasets and similar large static assets in separate data-only OCI images. These assets are inputs to an executable service; do not bake them into the application image or run a container merely to keep the files available.
- Version application and data images independently, recording compatible digests together in the release. Keep data preparation recipes and immutable source checksums in Git so the artifacts can be reproduced under `building/ci`. Include only the required data and metadata in a data image; no server, runtime or entrypoint is needed.
- Mount these artifacts into the consuming workload using the image-volume guidance in `deploying/workloads`. Separating data does not establish trust in its format or loader; use the application's supported asset formats.

## Dockerfile examples

Use the [codemowers/lolcatz repository](https://github.com/codemowers/lolcatz) for sample Dockerfiles. The Go, Node.js and Python sections link to published examples. Adapt build contexts, dependencies, toolchains and runtime assets to the application; the examples are not a substitute for the shared build rules.

## Runtime contract

- Run one application process per container, optionally with threads sharing its address space. Launch it directly using JSON exec-form ENTRYPOINT/CMD so it receives SIGTERM as PID 1. If a setup script is unavoidable, end it with `exec`. Lifecycle behavior belongs to `developing/general`.
- Follow `deploying/workloads` for orchestration. Do not add init wrappers such as dumb-init or tini to new applications. If the application launches subprocesses, it must reap them and propagate cancellation.
- Supply UID/GID at deployment time, preferably through applicable platform policy. Do not create accounts or bake an application UID into Dockerfile USER. Keep code root-owned and readable/executable by the runtime identity; do not recursively chown it. Use a read-only root filesystem with explicit writable mounts; see `deploying/security`.
- CA certificates belong in runtime trust configuration, not application images. This includes public and private CA bundles, even for `scratch` images. Follow `deploying/networking` for namespace trust distribution, Pod mounts and client configuration; trust changes must not require rebuilding the application image.

## Porting legacy applications

These conventions target newly designed applications. Porting an existing application to Kubernetes is not a promise that it already follows them, nor a requirement to rewrite it before deployment. Evaluate each legacy port case by case rather than treating every deviation as a defect.

- Inspect the actual process model, identity and filesystem assumptions, state, configuration, networking, startup/shutdown behavior and supported deployment model. Establish what can change without breaking application behavior or vendor support.
- Agree on the scope of the port. Preserve necessary compatibility, for example an existing master/worker arrangement, startup wrapper, fixed UID or writable directory, when replacing it is outside that scope. Identify which changes are required for the target cluster and which are optional modernization work.
- Record each retained deviation, its reason, operational consequences and any compensating configuration. Verify startup, health, termination, upgrades and data recovery for the resulting deployment. A documented legacy constraint does not establish that the cluster permits it; resolve applicable policy requirements with the platform owner.
- Keep exceptions local to that application. Do not turn a legacy workaround into the default for new services or silently expand a port into a redesign.
