# General deployment guidance

Discover the target using `deploying/discovery`, select dependencies using `deploying/storage`, and deploy immutable releases as described in `building/ci`. Keep configuration in version control without credentials. Configure Helm image prefixes and Skaffold publishing destinations according to `building/ci`. YAML formatting and project decision records follow `developing/general`.

## Release ownership and isolation

- Namespace lifecycle belongs to the platform. Never add `kind: Namespace` to application charts, manifests or GitOps inventories, or adopt an existing namespace into application ownership.
- Before removing an existing Namespace declaration, inspect live ownership/tracking metadata and Helm/GitOps release inventories. Detach it safely from application ownership and pruning first; labels alone do not remove inventory entries. If safe detachment is unresolved, retain the declaration until the controller/release can be migrated. Never delete/recreate the namespace.
- Support independent instances across namespaces and multiple Helm releases within one namespace. Derive application-owned names and references from `.Release.Name` using consistent chart helpers. Include release identity in Pod labels and workload/Service selectors. Cover dependency requests, identity registrations, Secrets, ConfigMaps and PVCs so releases cannot collide or select each other's Pods.
- Omit ordinary `metadata.namespace`; let the release target namespace apply. Use `.Release.Namespace` only for required explicit references, such as RBAC ServiceAccount subjects or namespace-qualified addresses. Cluster-scoped resources never have a namespace. Configure instance-specific hostnames through values.

## Helm configuration and admission defaults

- Expose TLS and other platform settings through Helm values, but leave their defaults unset. This includes certificate/issuer references, trust, listeners, storage classes, replicas and security settings as appropriate. Render fields only when explicitly supplied; do not reproduce platform defaults with Helm's `default` function.
- Discover applicable mutation policies and resource defaults before relying on them. Kubernetes 1.36 enables the stable [MutatingAdmissionPolicy mechanism](https://kubernetes.io/docs/reference/access-authn-authz/mutating-admission-policy/); it does not install platform policies or bindings. Supply required settings through deployment-specific values when no applicable policy or resource default provides them. Keep resource-specific conventions in their own cluster descriptions.
- Applicable admission policies may configure IPv6 listeners for databases and other dependencies, database TLS, CA-certificate mounts and trust paths, Pod/container security contexts, dependency storage classes, and ingress conventions. Discover each policy’s target resources, bindings and requirements independently; none of these capabilities implies the others. Keep concrete defaults in the responsible policy/resource description. Verify the admitted configuration and the operator/application’s resulting behavior; a mounted CA bundle still needs the client to use it. See `deploying/networking`, `deploying/security` and `deploying/storage` for the corresponding requirements.
- Unset means omitted, not disabled or empty. Do not render nulls, empty strings or empty collections as placeholders. Check key presence to preserve explicit `false` and `0`; retain empty collections only where the API requires them. Example values are not chart defaults.
- Detect optional APIs with Helm capability checks, separately for each kind; do not use a feature flag as proof that infrastructure exists. Monitoring integration follows `deploying/observability`; required dependencies must be resolved rather than silently omitted.
- Render the chart, inspect a server-side dry run where supported, then verify admitted resources and runtime readiness. Omission alone does not prove the platform supplied a setting.

## Environment mappings and rollouts

- Generate an explicit `env:` entry for each configured variable; do not use `envFrom`. Render nonsecret values as quoted `value:` strings in the Pod template. Use explicit `secretKeyRef` or `configMapKeyRef` mappings for referenced data; never inline generated credentials into Helm values or manifests.
- A changed Pod template triggers the Deployment's rollout. Changes to referenced Secret/ConfigMap data alone do not: use a configuration checksum annotation for chart-owned data, or the established reload/restart mechanism for operator-owned data.
- Apply unset semantics to platform environment settings too. Omit unspecified variables and omit `env` when there are none. Preserve explicitly supplied booleans and numbers as strings.

Example rendered entries for explicitly supplied values:

```yaml
env:
  - name: LOG_LEVEL
    value: "debug"
  - name: TLS_ENABLED
    value: "true"
```
