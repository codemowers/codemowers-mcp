# Codemowers MCP

Software engineering guidance at **https://mcp.codemowers.io**, focused on bare-metal Kubernetes clusters and useful to developers deploying to Kubernetes environments of all kinds, including managed cloud and local clusters.
The homepage renders the bundled Markdown with **Zen of Kubernetes** at the top.
There is no browser explorer or authentication. Agents use the same documents
through **https://mcp.codemowers.io/mcp** (Streamable HTTP).

Go's `net/http`, the official MCP Go SDK and Goldmark provide the server and
Markdown rendering. Markdown, HTML and CSS are embedded in one binary; the
container runs that binary as an unprivileged user in a `scratch` image. No Node,
nginx, database, cluster credentials or runtime document downloads are needed.
HTML is rendered once at startup and requires no browser JavaScript.

## Run locally with Docker

Images are published under `ghcr.io/codemowers/codemowers-mcp`. Once a release is published, replace `VERSION` below with its published tag (or use a digest-pinned reference):

The [image workflow](.github/workflows/image.yaml) runs Go tests and vet on pull
requests and pushes to `main`. After checks pass, pushes to protected `main`
publish using `GITHUB_TOKEN` with package write permission. Enable branch
protection for `main`; publishing is skipped while the branch is unprotected.
Tags use `sha-<commit>-<run-id>-<attempt>` to avoid overwriting earlier builds.
The workflow summary records the digest reference for deployment.

```sh
docker run --rm --name codemowers-mcp \
  -p 127.0.0.1:3000:3000 \
  -e PUBLIC_URL=http://localhost:3000 \
  ghcr.io/codemowers/codemowers-mcp:VERSION
```

Open http://localhost:3000 for the guide, or use `http://localhost:3000/mcp` in your MCP client. No authentication is required. Stop the foreground container with Ctrl-C.

To run unpublished changes from this source directory:

```sh
docker build -t codemowers-mcp:local .
docker run --rm --name codemowers-mcp \
  -p 127.0.0.1:3000:3000 \
  -e PUBLIC_URL=http://localhost:3000 \
  codemowers-mcp:local
```

## Development

Requires Go 1.26.5 or later.

```sh
go test ./...
go vet ./...
PUBLIC_URL=http://localhost:3000 go run .
```

Open http://localhost:3000. Alternatively, use `docker compose up --build`.

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `PUBLIC_URL` | `https://mcp.codemowers.io` | Public origin used in connection instructions and MCP Origin validation |
| `HTTP_ADDR` | `[::]:3000` | Website and MCP listener |
| `HEALTH_ADDR` | `[::]:3001` | Internal `/healthz` and `/readyz` listener |

SIGTERM stops readiness and drains requests for up to 20 seconds. The public
listener does not expose health endpoints. The service does not collect client
analytics or log request bodies.

## Advisory API

Call `get_engineering_advisory` without arguments for an index, or with
`{"path":"developing/logging"}` for a document. Resources use
`engineering://advisory/index` and `engineering://advisory/<path>`. Raw Markdown
is also available at `/advisory/<path>.md`.

The server is stateless and returns JSON over HTTP, without sessions or sticky
routing. It has no cluster access. General guidance does not establish that an
operator, issuer or admission policy is installed; discover the target cluster's
capabilities and applicable policy bindings before using the examples.

## Content

The guide applies [The Twelve-Factor App](https://12factor.net/) principles to Kubernetes: explicit configuration and dependencies, separate build/release/run stages, stateless processes, disposable workloads and development/production parity. It adds cluster discovery, operator-managed dependencies, admission policy integration and Helm release isolation. The [original Codemowers best-practices notes](https://pad.codemowers.io/p/best-practices) provide additional background.

Edit `content/**/*.md`; every file starts with `# Title`. Paths become both MCP
lookup keys and website anchors. `index` is generated and reserved. Rebuild the
binary/image after editing embedded content. Shared rules have one primary document;
language sections add runtime-specific details. Zen principles provide short hover/focus
explanations linking to those documents. This service contains shared development,
build and deployment conventions. CRD-specific requirements and YAML examples live in Driftmower's cluster advisory. Discover the target cluster's APIs before relying on them.

This content was extracted from `../driftmower-mcp/src/advisory/content`.
Driftmower retains CRD examples, platform discovery and administrator guidance through
`get_cluster_advisory` and `cluster://advisory/` resources. Existing administrator
documents stay local to each Driftmower instance; this service never fetches them.

## Deployment

The Helm chart configures `mcp.codemowers.io` and exposes only port 3000 through
its Service and Ingress. It creates no namespace, RBAC or authentication resources.
Use an existing namespace and the cluster's default ingress class, or set an
available class directly in the Ingress template. Provision DNS and the
`codemowers-mcp-tls` Secret separately. Deployment settings are inlined in the
chart templates.

```sh
helm template codemowers-mcp ./chart --namespace YOUR_NAMESPACE
```

Publish production images to `ghcr.io/codemowers/codemowers-mcp` only through
protected-branch CI, following `content/building/ci.md`. Set the Deployment image
to the immutable published reference and deploy with an explicit verified
Kubernetes context. The chart uses `codemowers-mcp` as its image name;
Skaffold builds that image and substitutes its resolved reference in the rendered
manifests.

TLS terminates at the selected ingress controller. The application listener uses
HTTP; clusters requiring encrypted ingress-to-Pod traffic need that transport
provided by their deployment infrastructure.
