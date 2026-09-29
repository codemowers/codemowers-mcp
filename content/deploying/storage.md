# Application dependencies and persistence

## Native or portable services

Determine the platform using `deploying/discovery`, then ask the user to choose **provider-native managed services** or **portable Kubernetes operators**. Record the platform, approach, selected service/operator and rationale using `developing/general`; reuse existing decisions.

For PostgreSQL on AWS, Amazon RDS versus CloudNativePG illustrates the tradeoff: provider operations versus a reusable Kubernetes deployment model with operator, storage, backup and recovery responsibilities. Check availability, permissions, connectivity, costs and support before recommending either. A platform name does not establish service availability. Resolve a missing prerequisite with the user/platform administrator rather than silently changing the selected approach.

## Ownership and provisioning

- Prefer custom resources for application databases, caches and OpenID Connect registrations. Passmower, CloudNativePG (CNPG) and Dragonfly are examples, subject to `deploying/discovery`. For native services, prefer an available Kubernetes provisioning interface such as Crossplane; verify its providers, credentials, permissions and resource scope.
- The platform team administers databases, caches and identity services cluster-wide, including operators, access policy, upgrades, backups and recovery. Application releases own per-instance requests and registrations. Never install shared operators, identity infrastructure or external dependency charts through the application chart. Do not handwrite StatefulSets, Deployments or Pods for these backing services; operator-generated workloads stay under operator management.
- Treat the cluster as shared. Keep per-instance dependency requests, identity registrations and credential references in the application's namespace, even when the requested service runs remotely. For a cluster-scoped API, use an available namespaced request interface or administrator-provisioned resources with scoped connection information.
- Apply the release naming rules in `deploying/general` to support multiple instances in one namespace. Share a service only through an explicit recorded choice, with isolated databases, cache access and identity registrations. Credential lifecycle belongs to `deploying/security`.

## Data lifecycle

- Keep application containers stateless. Store session state needed across requests in an appropriate shared backing service or supported client-side session design; do not rely on a replica’s memory/filesystem or sticky routing for correctness. Choose capacity, storage class, retention, backups and restore procedures for the workload; verify recovery. A retained PVC or Helm keep annotation is not a backup. Use bounded disposable volumes for temporary files.
- Configure internal and browser-facing object-storage endpoints separately. Follow `developing/architecture` for direct signed transfers and `deploying/networking` for TLS.
- Obtain custom-resource schemas and examples from the target cluster's Driftmower advisory or installed operator documentation. This general guide does not define operator defaults.
