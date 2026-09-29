# Discover the target cluster before provisioning

Identify the target and inspect its capabilities before provisioning. Use advertised Driftmower tools, direct `kubectl`, or both when a response lacks detail. CRD-specific examples belong to the target cluster advisory; examples never establish availability.

## With Driftmower MCP

Identify the connection's cluster using `get_cluster_identity` and its platform using `get_cluster_platform` when advertised. Record whether it is k3s, vanilla Kubernetes, EKS, AKS or another verified platform. Treat unknown or conflicting evidence as a question to resolve with the user, not as vanilla Kubernetes. Before using local `kubectl`, match its Kubernetes API URL to a verified kubeconfig context. Discover the tools advertised by that connection instead of assuming a particular Driftmower version exposes every tool.

| Information | Advertised tools to use |
| --- | --- |
| Platform, Kubernetes version and detection evidence | `get_cluster_platform` |
| Installed CRDs, served versions, ingress classes and issuers | `describe_cluster_capabilities` |
| A particular CRD's live schema and description | `describe_custom_resource_definition` with its full CRD name |
| Admission policies, descriptions, target rules and binding scope | `list_admission_policies` |
| StorageClass descriptions, parameters, defaults and lifecycle settings | `list_storage_classes` |
| Accessible namespaces and local guidance | `list_accessible_namespaces`, `get_cluster_advisory` |

Use additional offerings or namespace-readiness tools only if the connection actually advertises them. A capability listing does not prove controller health, permission to create a resource, or that a policy applies. The admission catalog is a discovery summary, not an evaluation of policy code: inspect matching policy and binding specifications, selectors, conditions and parameters with `kubectl` when the summary is insufficient. Missing or inaccessible information is unknown, not evidence that a policy or resource is absent.

## Without Driftmower, or when more detail is needed

Use the existing kubeconfig credentials and an explicit verified context on every cluster operation. Confirm the target API URL with the user or established project configuration if no MCP identity is available. Do not switch the current context implicitly or print raw kubeconfig credentials. Choose an existing target namespace; checking it does not transfer namespace lifecycle ownership to the application.

Replace `verified-context` and `existing-namespace` in each command with the verified deployment target:

```sh
kubectl --context verified-context api-resources
kubectl --context verified-context get namespace existing-namespace -o yaml
kubectl --context verified-context get storageclasses.storage.k8s.io -o yaml
kubectl --context verified-context --namespace existing-namespace get resourcequotas,limitranges -o yaml
```

Read annotations and descriptions as well as settings. For StorageClasses, inspect the provisioner, default annotation, parameters, reclaim policy, binding mode, expansion support and topology restrictions. Do not assume a default StorageClass exists or copy one from a different cluster.

### Determine platform and record infrastructure choices

Read the server version and, when permitted, platform-specific node labels:

```sh
kubectl --context verified-context version -o yaml
kubectl --context verified-context get nodes -L eks.amazonaws.com/nodegroup,kubernetes.azure.com/cluster
```

A k3s version suffix, EKS distribution markers or AKS-specific labels are evidence to check against the cluster's provisioning records. A generic version, cloud provider ID or missing label does not prove vanilla Kubernetes or a particular managed service. If discovery is restricted or ambiguous, confirm the platform with the user or administrator. Do not infer the platform from an MCP registration name or kubeconfig context name.

Resolve dependency and Prometheus choices through `deploying/storage` and `deploying/observability`, using the decision-recording rules in `developing/general`.

### Check each required CRD

For each selected operator API, including native-service provisioning APIs, substitute its CRD, resource name and served version:

```sh
CRD_NAME=resourceplural.operator.example
RESOURCE_NAME=resourceplural
API_VERSION=operator.example/v1
kubectl --context verified-context get customresourcedefinitions.apiextensions.k8s.io "$CRD_NAME" -o yaml
kubectl --context verified-context explain "$RESOURCE_NAME.spec" --api-version="$API_VERSION"
kubectl --context verified-context --namespace existing-namespace auth can-i create "$CRD_NAME"
```

Inspect the CRD's `Established` condition, deletion state, served versions, scope and OpenAPI schema. Use the resource's discovered API version for `kubectl explain`. Repeat for every other custom resource in the chosen example, including certificate and credential generators when referenced. A CRD alone does not establish that its controller is running, watches the target namespace or can reconcile the resource: inspect the discovered controller workload and relevant reconciliation status with the permissions available.

If the API or its controller is absent, establish how that prerequisite will be provided before deploying dependent resources. Do not silently install an operator, copy CRD definitions into an application chart or redirect to a named alternative whose availability has not been established. Distinguish `Forbidden`, discovery/network failures and actual `NotFound` results.

### Check admission configuration and bindings

Discover which admission APIs are served first:

```sh
kubectl --context verified-context api-resources --api-group=admissionregistration.k8s.io
```

Run the corresponding reads for the APIs present in that result:

```sh
kubectl --context verified-context get mutatingadmissionpolicies.admissionregistration.k8s.io -o yaml
kubectl --context verified-context get mutatingadmissionpolicybindings.admissionregistration.k8s.io -o yaml
kubectl --context verified-context get validatingadmissionpolicies.admissionregistration.k8s.io -o yaml
kubectl --context verified-context get validatingadmissionpolicybindings.admissionregistration.k8s.io -o yaml
kubectl --context verified-context get mutatingwebhookconfigurations.admissionregistration.k8s.io -o yaml
kubectl --context verified-context get validatingwebhookconfigurations.admissionregistration.k8s.io -o yaml
```

Match policies to bindings, resource rules, operations, namespace/object selectors, conditions and referenced parameter resources. Check namespace labels and validation actions; an unbound policy or a nonmatching binding does not supply defaults. Webhook configurations may establish another admission mechanism; inspect its relevant configuration and documented behavior when present. An empty API listing does not rule out other admission configuration controlled by the cluster administrator. If inspection is forbidden or incomplete, obtain the missing requirements rather than assuming there are no constraints.

### Check TLS and other prerequisites

Discover issuer and certificate APIs before querying their instances. If cert-manager's APIs are installed, inspect the available cluster-scoped and namespaced issuers:

```sh
kubectl --context verified-context get clusterissuers.cert-manager.io -o yaml
kubectl --context verified-context --namespace existing-namespace get issuers.cert-manager.io -o yaml
```

Read each issuer's own description and readiness to determine its suitability. Discover the trust bundle and the namespace's network posture from relevant ConfigMaps, NetworkPolicies and platform guidance. Inspect only the required objects and Secret key names; do not print credential values. No issuer, trust-bundle name or secret generator is implied by another resource's example.

After discovery, apply the Helm defaults and deployment verification rules in `deploying/general`.

References: [Kubernetes CRDs and schema discovery](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/), [admission policies and bindings](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/), and [admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).
