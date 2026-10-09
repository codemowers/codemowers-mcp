# Application architecture and identity

## Choose the authentication boundary

Before implementing authentication, ask the user to choose **a backend for frontend (BFF)** or **direct OIDC/OAuth API access with custom scopes**. Explain the tradeoffs and follow the decision-recording rules in `developing/general`. Record clients, provider/access proxy, audiences/scopes, token/session ownership and the signed-transfer boundary. Wait for unresolved choices before implementing dependent authentication work.

| Pattern | Fits | Responsibilities |
| --- | --- | --- |
| BFF: browser session cookie; backend holds tokens and calls APIs | A first-party browser UI; centralized sessions and API composition | Secure HttpOnly cookies, CSRF protection and session lifecycle. XSS can still act through the session. Adds a backend hop. |
| Direct API: clients hold access tokens and call protected APIs | Independent browser, mobile, CLI or integration clients; delegated access | Audience/scope design, token lifecycle and CORS. Browser tokens are exposed to malicious JavaScript; public clients cannot keep secrets. |

Both use an external OIDC provider. Use Authorization Code with PKCE for interactive sign-in. A BFF may also use scoped downstream tokens; token placement and API callers distinguish the patterns. See the [IETF browser application guidance](https://datatracker.ietf.org/doc/draft-ietf-oauth-browser-based-apps/).

Recommend based on actual client requirements and obtain the unresolved choice before implementing it. Keep business authorization in the backend; split services only when ownership, scaling or isolation warrants it.

## Delegate identity, enforce authorization

- Use an external identity provider for accounts and authentication. Do not embed an IdP in the application or bundle one into its release; follow `deploying/storage` for platform ownership and per-application registrations. Do not implement password collection, storage, hashing, verification or reset/recovery flows in the application. Redirect authentication and account lifecycle to the provider; do not add a parallel local login system.
- Discover provisioning interfaces through `deploying/discovery` and follow `deploying/storage` for ownership. [Passmower](https://github.com/passmower/passmower) and [authentik](https://docs.goauthentik.io/) illustrate OIDC providers; an access proxy such as [Pomerium](https://www.pomerium.com/docs) is a separate role, paired with a provider. None is assumed installed.
- Associate application data with the verified `(iss, sub)` identity. Enforce tenant membership, object ownership and business permissions; authentication alone does not grant record access.
- Agree on API audiences and minimal scopes such as `documents:read`. Verify how the provider grants them: a client's requested scope does not prove a token carries that permission.
- Validate access tokens with the issuer's supported mechanism, checking issuer, audience, expiry and scopes, then enforce object authorization. ID tokens are not API access tokens. A BFF must enforce the same business permissions rather than proxying broad service credentials without restriction.

## Keep signed transfers direct

- After authorizing an object operation, issue a short-lived, narrowly scoped presigned URL for direct client transfer. Keep storage credentials server-side and configure endpoint CORS for the actual browser origins and operations.
- Sign for the client-reachable HTTPS endpoint. Preserve the signed method, host, path, query, headers and payload constraints; never rewrite an internal signed URL for public use. See [S3 Signature Version 4 query authentication](https://docs.aws.amazon.com/AmazonS3/latest/developerguide/sigv4-query-string-auth.html).
- Do not relay signed-request services such as MinIO/S3 through the BFF or application API. Resolve missing client endpoints instead of adding a proxy workaround. Infrastructure routing must preserve signed requests; other signed services require their own supported direct-client authorization mechanism.
