# MCP OAuth and Stytch runbook

Languages: [English](mcp-oauth-stytch.md) | [简体中文](mcp-oauth-stytch.zh-CN.md) | [繁體中文](mcp-oauth-stytch.zh-TW.md)

Purpose: configure Fast Note Sync Service as an OAuth-protected MCP resource for ChatGPT and other MCP clients that support OAuth.

Read this when you need to enable MCP OAuth, configure Stytch as the authorization server, configure an FNS deployment, or troubleshoot ChatGPT connector authorization.

This document does not cover WebGUI OIDC SSO login. The current implementation does not add a "Login with OIDC", "Login with Stytch", or "Login with Casdoor" button to the WebGUI, and it does not implement a WebGUI callback route such as `/api/user/auth/callback`.

## What this implements

The current OAuth implementation is for MCP access control.

When `oauth.enabled` is true, `/api/mcp` and `/api/mcp/sse` become OAuth-aware protected resources:

1. An unauthenticated MCP request receives `401 Unauthorized` with a `WWW-Authenticate` header.
2. The challenge points the client to protected resource metadata, usually `/.well-known/oauth-protected-resource/api/mcp`.
3. The metadata advertises this FNS deployment as the protected resource and lists the configured authorization server.
4. ChatGPT or another OAuth-capable MCP client completes OAuth with that authorization server.
5. The client sends later MCP requests with `Authorization: Bearer <access-token>`.
6. FNS verifies the bearer JWT locally with the configured issuer, JWKS URL, audience/resource, and scopes.
7. FNS maps the verified JWT subject to an FNS user and maps OAuth scopes to FNS MCP permissions.

This matches the ChatGPT Apps SDK authentication model: after OAuth, ChatGPT attaches the access token to MCP requests, while the MCP server remains responsible for signature validation, issuer and audience checks, expiry checks, and scope enforcement.

References:

- OpenAI Apps SDK authentication: <https://developers.openai.com/apps-sdk/build/auth#implementing-token-verification>
- OpenAI Apps SDK MCP overview: <https://developers.openai.com/apps-sdk/concepts/mcp-server#why-apps-sdk-standardises-on-mcp>
- Stytch MCP Server: <https://stytch.com/docs/resources/workspace-management/stytch-mcp-server>
- Stytch custom Connected Apps flow: <https://stytch.com/docs/connected-apps/build-custom-flow/getting-started-api>
- Stytch B2B Start OAuth Authorization API: <https://stytch.com/docs/api-reference/b2b/api/connected-apps/consent-management/start-oauth-authorization>

## What this does not implement

This is not WebGUI OIDC SSO.

The following are not supported by this implementation:

- A WebGUI login button for generic OIDC providers such as Casdoor.
- A WebGUI OIDC callback endpoint such as `/api/user/auth/callback`.
- Automatic WebGUI user provisioning from an external OIDC login.
- Replacing `/api/user/login` with Stytch, Casdoor, Pocket ID, or another OIDC provider.

The `oauth` config block verifies bearer tokens for MCP requests. It does not change WebGUI login behavior.

## Code paths

The main code paths are:

- `internal/config/oauth.go`: config schema, defaults, validation, and protected resource metadata.
- `internal/routers/router_oauth_metadata.go`: OAuth protected resource metadata routes.
- `internal/routers/router_mcp.go`: MCP routes protected by OAuth-aware middleware.
- `internal/middleware/mcp_oauth.go`: mixed static-token/OAuth MCP authentication.
- `internal/oauth/verifier.go`: JWT verification using issuer, audience, JWKS, expiry, and scopes.
- `internal/oauth/subject_mapper.go`: JWT subject to FNS UID mapping.
- `internal/oauth/scope_mapper.go`: OAuth scope to FNS MCP permission mapping.
- `internal/oauth/stytch_client.go`: Stytch authorize start/submit API client.
- `internal/routers/api_router/handler_stytch_oauth.go`: authenticated FNS endpoints that bridge the authorize page to Stytch.
- `frontend/oauth-authorize.html`: OAuth authorize page built from the WebGUI.

## Public endpoints

With OAuth enabled, the deployment must expose these endpoints over HTTPS:

| Endpoint | Purpose |
| --- | --- |
| `/api/mcp` | Streamable HTTP MCP endpoint. |
| `/api/mcp/sse` | Legacy SSE MCP endpoint. |
| `/.well-known/oauth-protected-resource/api/mcp` | Protected resource metadata for `/api/mcp`. |
| `/.well-known/oauth-protected-resource` | Generic protected resource metadata endpoint. |
| `/oauth/authorize` | Browser page used during the Stytch Connected App authorization flow. |
| `/api/oauth/stytch/authorize/start` | Authenticated FNS API endpoint that calls Stytch authorize start. |
| `/api/oauth/stytch/authorize/submit` | Authenticated FNS API endpoint that calls Stytch authorize submit. |

The two `/api/oauth/stytch/*` endpoints require an existing FNS WebGUI login token. The authorize page asks the user to log in to FNS first if `localStorage.token` is missing.

## Configuration reference

Add or update the `oauth` block in `config.yaml`.

```yaml
oauth:
  enabled: true
  resource: "https://fns.example.com/api/mcp"
  authorization-servers:
    - "https://example.customers.stytch.com"
  jwks-url: "https://example.customers.stytch.com/.well-known/jwks.json"
  issuer: "https://example.customers.stytch.com"
  audience: []
  scopes-supported:
    - notes:read
    - notes:write
    - files:read
    - files:write
    - vaults:read
  required-scopes:
    - notes:read
    - files:read
    - vaults:read
  resource-name: "Fast Note Sync MCP"
  allow-static-fns-token: true
  default-client: "ChatGPT"
  default-client-name: "ChatGPT"
  default-client-version: ""
  default-vault-name: "main"
  default-fns-scope: ""
  subject-mapping:
    mode: "email_or_fixed_uid"
    claim: "email"
    fixed-uid: 1
  stytch:
    enabled: true
    kind: "b2b"
    domain: "https://example.customers.stytch.com"
    project-id: "project-live-..."
    secret: "${FNS_STYTCH_SECRET}"
    organization-id: "organization-live-..."
    member-id: "member-live-..."
```

### Core OAuth fields

| Field | Required when enabled | Description |
| --- | --- | --- |
| `enabled` | yes | Enables OAuth-aware MCP authentication and protected resource metadata. |
| `resource` | yes | Protected resource identifier. For ChatGPT MCP, use the public MCP endpoint URL, for example `https://fns.example.com/api/mcp`. |
| `authorization-servers` | yes | Authorization server issuer origins advertised to the MCP client. For Stytch, use the Stytch customer domain origin. |
| `jwks-url` | yes | JWKS endpoint used by FNS to verify JWT signatures. |
| `issuer` | yes | Expected JWT issuer. Must match the token `iss` claim. |
| `audience` | no | Optional accepted JWT audiences. If empty, FNS uses `resource` as the expected audience. |
| `scopes-supported` | no | Scopes advertised in protected resource metadata. Defaults to the standard FNS MCP scopes. |
| `required-scopes` | no | Scopes every OAuth token must contain. Defaults to `notes:read`, `files:read`, and `vaults:read`. |
| `resource-name` | no | Human-readable resource name in metadata. |
| `allow-static-fns-token` | no | Keeps existing FNS static token MCP clients working when OAuth is enabled. Defaults to true. |

Use `required-scopes: []` only when the authorization server or client does not issue the custom FNS scopes and you intentionally rely on FNS permission mapping or another deployment-level control. For stricter public deployments, keep explicit required scopes.

### FNS identity and permission mapping

| Field | Description |
| --- | --- |
| `subject-mapping.mode` | `email`, `fixed_uid`, or `email_or_fixed_uid`. |
| `subject-mapping.claim` | JWT claim used when matching an existing FNS user by email. Defaults to `email`. |
| `subject-mapping.fixed-uid` | FNS UID used by `fixed_uid` or as fallback for `email_or_fixed_uid`. |
| `default-fns-scope` | If set, bypasses OAuth scope mapping and grants this FNS scope to verified OAuth requests. |
| `default-client`, `default-client-name`, `default-client-version` | Default MCP client headers used when the request or token does not provide them. |
| `default-vault-name` | Default vault name used by MCP operations when the request does not provide a vault. |

For a single-user deployment, `fixed_uid` is the simplest option:

```yaml
subject-mapping:
  mode: fixed_uid
  fixed-uid: 1
```

For a multi-user deployment, prefer email mapping:

```yaml
subject-mapping:
  mode: email
  claim: email
```

The email claim must exist in the verified access token, and the corresponding FNS user must already exist.

## Stytch setup

Use Stytch Connected Apps as the authorization server. The FNS server does not implement a full OAuth authorization server by itself; it acts as an MCP protected resource and uses Stytch to issue tokens.

### 1. Create or select a Stytch project

Use either:

- B2B project: recommended for controlled organization/member access.
- Consumer project: usable for user-based access when organization/member identifiers are not needed.

Record these values:

- Stytch project ID.
- Stytch project secret.
- Stytch customer domain, for example `https://example.customers.stytch.com`.
- For B2B, organization ID and member ID, unless the authorize flow will identify the member through a Stytch session.

Store the project secret only in a secret manager, environment variable, or another private configuration mechanism. Do not commit the real secret to Git.

### 2. Configure the Connected App

In Stytch, create a Connected App for ChatGPT or the MCP client.

Configure:

- Redirect URI: the redirect URI required by the MCP client or ChatGPT connector setup.
- Allowed scopes: include the FNS MCP scopes that the client will request, such as `notes:read`, `notes:write`, `files:read`, `files:write`, and `vaults:read`.
- Authorization entry point: point the Connected App custom authorization flow to the FNS authorize page, for example `https://fns.example.com/oauth/authorize`.

The authorize page parses OAuth query parameters such as `client_id`, `redirect_uri`, `response_type`, `scope`, `state`, `nonce`, `code_challenge`, and `code_challenge_method`. It then calls the authenticated FNS Stytch bridge endpoints.

### Optional: configure Stytch through Stytch MCP

Stytch also publishes its own MCP server at:

```text
https://mcp.stytch.dev/mcp
```

Stytch's MCP server can be used from an MCP-capable development environment to manage Stytch project configuration with natural language instead of clicking through the Dashboard. It still requires OAuth authorization for the Stytch workspace before it can read or modify workspace resources.

When available, use Stytch MCP to inspect or configure:

- Existing projects and whether the target project is B2B or Consumer.
- Connected App client IDs and names.
- Connected App redirect URIs.
- Project public tokens and SDK settings, when relevant.
- Workspace-level settings needed by Connected Apps.

Even when Stytch MCP is available, keep the FNS `config.yaml` values as the source of truth for the FNS deployment. Stytch MCP configures Stytch; it does not update FNS service configuration.

### 3. Match Stytch values in FNS

For B2B:

```yaml
oauth:
  stytch:
    enabled: true
    kind: "b2b"
    domain: "https://example.customers.stytch.com"
    project-id: "project-live-..."
    secret: "${FNS_STYTCH_SECRET}"
    organization-id: "organization-live-..."
    member-id: "member-live-..."
```

For Consumer:

```yaml
oauth:
  stytch:
    enabled: true
    kind: "consumer"
    domain: "https://example.customers.stytch.com"
    project-id: "project-live-..."
    secret: "${FNS_STYTCH_SECRET}"
    user-id-prefix: "fns:"
```

Consumer mode uses `stytch.user-id` if configured. Otherwise, it sends `stytch.user-id-prefix + <FNS UID>` to Stytch.

## Step-by-step deployment

1. Deploy FNS with HTTPS.

   The public base URL must be stable, for example `https://fns.example.com`.

2. Configure `server.ext-api-url`.

   ```yaml
   server:
     ext-api-url: "https://fns.example.com"
   ```

3. Configure the Stytch Connected App and collect the domain, project ID, secret, and B2B organization/member IDs if used.

4. Add the `oauth` config block.

   Set `resource` to `https://fns.example.com/api/mcp`. Set `issuer`, `jwks-url`, and `authorization-servers` to the Stytch customer domain and JWKS URL.

5. Store the Stytch secret securely.

   Use your normal secret-management mechanism or a local uncommitted config file. Do not commit the real Stytch secret.

6. Restart the FNS service.

7. Verify protected resource metadata:

   ```bash
   curl -sS https://fns.example.com/.well-known/oauth-protected-resource/api/mcp
   ```

   Expected result:

   ```json
   {
     "resource": "https://fns.example.com/api/mcp",
     "authorization_servers": ["https://example.customers.stytch.com"],
     "scopes_supported": ["notes:read", "notes:write", "files:read", "files:write", "vaults:read"],
     "bearer_methods_supported": ["header"],
     "resource_name": "Fast Note Sync MCP",
     "jwks_uri": "https://example.customers.stytch.com/.well-known/jwks.json"
   }
   ```

8. Verify the MCP challenge:

   ```bash
   curl -i https://fns.example.com/api/mcp
   ```

   Expected result:

   ```http
   HTTP/2 401
   WWW-Authenticate: Bearer resource_metadata="https://fns.example.com/.well-known/oauth-protected-resource/api/mcp" resource="https://fns.example.com/api/mcp", error="invalid_token"
   ```

9. Add the MCP server in ChatGPT or another OAuth-capable MCP client.

   Use the MCP endpoint:

   ```text
   https://fns.example.com/api/mcp
   ```

   The client should discover the protected resource metadata and begin the OAuth flow with Stytch.

10. During authorization, log in to FNS if the authorize page asks for an FNS login.

    This is required because the FNS authorize bridge endpoints are protected by the existing FNS WebGUI token.

11. After consent, confirm the MCP client can list and call tools.

## Static token compatibility

If `allow-static-fns-token: true`, existing MCP clients can continue using an FNS static token:

```http
Authorization: Bearer <fns-token>
```

This compatibility path is useful during migration because OAuth-capable clients and older token-only clients can coexist.

If you want OAuth-only MCP access, set:

```yaml
oauth:
  allow-static-fns-token: false
```

## Troubleshooting

### ChatGPT says the MCP server does not implement OAuth

Check that:

- `oauth.enabled` is true.
- `https://fns.example.com/.well-known/oauth-protected-resource/api/mcp` returns HTTP 200.
- `https://fns.example.com/api/mcp` returns `401` with a `WWW-Authenticate` header that includes `resource_metadata`.
- `oauth.resource` exactly matches the public MCP endpoint URL.
- The ingress exposes `/.well-known/oauth-protected-resource/api/mcp`.

### Metadata is missing `authorization_servers`

Check `oauth.authorization-servers`. It must contain at least one authorization server origin.

### Token verification fails with `invalid_token`

Check:

- The token `iss` claim matches `oauth.issuer`.
- `oauth.jwks-url` is reachable by the FNS pod.
- The signing key appears in the JWKS response.
- The token has not expired.
- `oauth.audience` or `oauth.resource` matches the token audience/resource.

### Token verification fails with `insufficient_scope`

Check:

- The token contains every scope listed in `oauth.required-scopes`.
- The Stytch Connected App is allowed to issue the requested FNS scopes.
- `default-fns-scope` is intentionally unset or set to the expected FNS permission scope.

### The WebGUI has no OIDC login button

This is expected. The current implementation does not provide WebGUI OIDC SSO. It only protects MCP routes and provides an OAuth authorization bridge for MCP clients.

### `/api/user/auth/callback` returns 404

This is expected. That endpoint is not implemented.

### Stytch authorize start or submit fails

Check:

- `oauth.stytch.enabled` is true.
- `oauth.stytch.domain`, `project-id`, and `secret` are configured.
- For `kind: b2b`, both `organization-id` and `member-id` are configured.
- The Stytch Connected App allows the incoming `client_id`, `redirect_uri`, and scopes.
- The user is logged in to FNS before using `/oauth/authorize`.
