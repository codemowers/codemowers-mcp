# Continuous integration and releases

## Separate development and production publishing

- Developers push locally built images only to a separate development registry. Never push or copy a developer workstation's build into the production registry, even when it carries a Git-looking tag.
- Only trusted CI running from a protected Git branch may publish production images. Restrict production write credentials to that pipeline; developer credentials and untrusted pull-request jobs must have no production push access. Protect the build and workflow definitions along with application code.
- Make registry/repository prefixes explicit in Helm values and Skaffold environment configuration. Define a chart value such as `image.registryPrefix` and apply it consistently to application and data-image references; Helm does not supply a standard registry-prefix value. Local development must select the development registry and must never fall back to production.
- Persist Skaffold's development `default-repo` in its configuration file with `skaffold config set`, scoped to a verified development kubecontext. Pass Skaffold's resolved image references into Helm rather than constructing a different destination there. The protected CI pipeline selects the production prefix in its own configuration. See [Skaffold image repository handling](https://skaffold.dev/docs/environment/image-registries/). Prefixes route builds; registry permissions enforce the boundary.

Example: write a development registry prefix to Skaffold's config, then run against that context. Replace the illustrative context and registry with verified values:

```sh
skaffold config set --kube-context verified-dev-context default-repo registry.dev.example.com/team
skaffold dev --kube-context verified-dev-context
```

This is Skaffold's per-user configuration (normally `~/.skaffold/config`), not a `default-repo` field in `skaffold.yaml`. Avoid a global production default on developer machines.

## Reproducible, immutable production artifacts

- Every production image must be reproducible from its recorded Git commit and declared inputs. Commit Dockerfiles, build configuration and lockfiles; pin base images and external build inputs by digest or checksum, and retain them. Never depend on uncommitted files, workstation outputs or unversioned downloads. Large model/data inputs may live outside Git, but their immutable references, checksums and preparation instructions belong in Git.
- Automate builds, checks and vulnerability scanning in CI. Run integration checks through Compose with production-aligned dependencies, and verify the packaged image in Kubernetes. Build once per source revision and promote that verified CI artifact through staging and production without rebuilding or replacing it with a local build.
- Attach the applicable [OCI image annotations](https://github.com/opencontainers/image-spec/blob/main/annotations.md) to every built image, including its source URL, revision, creation time, version, title and description. Set license metadata only when the project's license is established. Derive values from the same immutable source revision and release metadata used for the image tag.
- Enforce immutable production tags in the registry and deploy by image digest. Never overwrite a published production tag. Mutable tags such as `latest` belong only to development publishing/defaults; release values must resolve to immutable references. Apply the same rules to separate data images.
- Record source, application/data-image digests and deployment configuration for each release. Retain artifacts needed to redeploy or roll back. Supply environment configuration at deployment time; check schema compatibility before rollback and coordinate migrations as described in `deploying/workloads`.
