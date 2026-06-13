# WebGUI OIDC Login

Purpose: this document explains how to enable OpenID Connect (OIDC) login for the Fast Note Sync WebGUI. Read this when you want users to sign in to the WebGUI through an external identity provider such as Dex, Keycloak, or Casdoor. This does not cover MCP OAuth resource-server authorization; that is configured under `oauth`.

## What It Does

The `oidc` configuration enables SSO login for the WebGUI only.

When enabled:

- the WebGUI login page fetches `/api/user/auth/oidc/config`;
- if OIDC is enabled, the login page shows one OIDC login button for a single provider or one button per configured provider;
- `/api/user/auth/oidc/start` creates a state, nonce, and PKCE verifier, then redirects the browser to the selected provider;
- the provider redirects back to its configured `redirect-url`;
- the service verifies the `id_token`, maps the OIDC subject to a local FNS user, and issues the normal WebGUI login token.

## Configuration

For new WebGUI deployments, configure one or more providers under `oidc.providers`:

```yaml
oidc:
  enabled: true
  callback-path: "/api/user/auth/oidc/callback"
  auto-register: false
  user-mapping:
    subject-claim: "sub"
    email-claim: "email"
    username-claim: "preferred_username"
    display-name-claim: "name"
  providers:
    - id: "dex"
      display-name: "Login with Dex"
      issuer: "https://dex.example.com/dex"
      client-id: "fns-webgui"
      client-secret: "change-me"
      redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
      scopes:
        - openid
        - profile
        - email
    - id: "casdoor"
      display-name: "Login with Casdoor"
      issuer: "https://casdoor.example.com"
      client-id: "fns-webgui"
      client-secret: "change-me"
      redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
      scopes:
        - openid
        - profile
        - email
      user-mapping:
        display-name-claim: "displayName"
    - id: "keycloak"
      display-name: "Login with Keycloak"
      issuer: "https://keycloak.example.com/realms/fns"
      client-id: "fns-webgui"
      client-secret: "change-me"
      redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

Each provider has a stable `id`. Use lowercase letters, numbers, and hyphens, and keep it stable so provider selection and callback handling remain predictable.

Required for each provider:

- `issuer`
- `client-id`
- `client-secret`
- `redirect-url`

Provider-level `user-mapping` overrides only the claims that differ from the global `oidc.user-mapping`. This is useful when most providers use `name`, but one provider uses a different display-name claim such as Casdoor `displayName`.

Defaults:

- `display-name`: `Login with OIDC`
- `callback-path`: `/api/user/auth/oidc/callback`
- `scopes`: `openid`, `profile`, `email`
- `subject-claim`: `sub`
- `email-claim`: `email`
- `username-claim`: `preferred_username`
- `display-name-claim`: `name`

Keep `client-secret` out of Git-managed public configuration.

### Backward-Compatible Single Provider

Existing single-provider deployments can keep the historical top-level provider fields:

```yaml
oidc:
  enabled: true
  display-name: "Login with SSO"
  issuer: "https://idp.example.com"
  client-id: "fns-webgui"
  client-secret: "change-me"
  redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
  callback-path: "/api/user/auth/oidc/callback"
  scopes:
    - openid
    - profile
    - email
  auto-register: false
  user-mapping:
    subject-claim: "sub"
    email-claim: "email"
    username-claim: "preferred_username"
    display-name-claim: "name"
```

This form behaves like a single entry in `oidc.providers`. Prefer `providers` when you want multiple WebGUI login buttons.

## User Mapping

FNS stores OIDC bindings in `user_oidc_identity`.

Login resolution order:

1. If `(issuer, subject)` is already bound, FNS logs in the bound local user.
2. If no binding exists and the OIDC email matches an existing local user, FNS creates the binding and logs in that user.
3. If no user matches and `auto-register: true`, FNS creates a local user and then creates the binding.
4. If no user matches and `auto-register: false`, login fails.

For safer rollout, start with `auto-register: false`, create local users first, and let first login bind them by email.

When `auto-register: true`, the local username is generated from the first usable value in this order:

1. `username-claim` such as `preferred_username`
2. `display-name-claim` such as `name`
3. the email local part before `@`
4. `oidc_` plus the OIDC subject

The value is normalized to FNS username rules: letters, numbers, and underscores, 3 to 20 characters. If the username already exists, FNS appends a numeric suffix.

## Provider Setup

The WebGUI OIDC login uses standard OIDC discovery, authorization code flow, PKCE, and `id_token` verification. Google, Microsoft Entra ID, Auth0, Okta, Zitadel, and similar providers can work as long as they provide a normal OIDC issuer, client ID, client secret, redirect URL, and claims compatible with the configured user mapping.

GitHub is different: GitHub OAuth Apps are OAuth 2.0 providers and do not behave like a plain OIDC login provider with discovery and `id_token` in the same way. For GitHub login, usually put Dex, Keycloak, or Casdoor in front as an OIDC broker, or use a separate OAuth adapter that translates GitHub OAuth into the login flow you need.

### Dex

Create a confidential client:

- Client ID: `fns-webgui`
- Client secret: same as the provider `client-secret`
- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Scopes: `openid`, `profile`, `email`

Use the Dex issuer URL as the provider `issuer`, for example:

```yaml
providers:
  - id: "dex"
    display-name: "Login with Dex"
    issuer: "https://dex.example.com/dex"
    client-id: "fns-webgui"
    client-secret: "change-me"
    redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

### Keycloak

Create an OpenID Connect confidential client:

- Client ID: `fns-webgui`
- Client authentication: enabled
- Standard flow: enabled
- Valid redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- PKCE: `S256` is supported

Use the realm issuer as the provider `issuer`:

```yaml
providers:
  - id: "keycloak"
    display-name: "Login with Keycloak"
    issuer: "https://keycloak.example.com/realms/fns"
    client-id: "fns-webgui"
    client-secret: "change-me"
    redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

### Casdoor

Create or update an application:

- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Grant type: `authorization_code`
- Client ID and secret: match the provider `client-id` and `client-secret`
- Scopes: `openid`, `profile`, `email`

Use the Casdoor origin as the provider `issuer`:

```yaml
providers:
  - id: "casdoor"
    display-name: "Login with Casdoor"
    issuer: "https://casdoor.example.com"
    client-id: "fns-webgui"
    client-secret: "change-me"
    redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

Casdoor commonly uses `displayName` rather than `name`. If needed, set:

```yaml
providers:
  - id: "casdoor"
    user-mapping:
      display-name-claim: "displayName"
```

## Public URL and Reverse Proxy

`redirect-url` must be the externally reachable callback URL seen by the provider and the browser. Behind a reverse proxy, use the public HTTPS origin, not the internal container address.

Example:

```yaml
redirect-url: "https://notes.example.com/api/user/auth/oidc/callback"
```

If you run the WebGUI on a separate port, the callback still belongs to the API route. Configure the provider with the callback URL that reaches the FNS service.

## Verification

The repository includes a Docker-backed smoke test:

```bash
scripts/oidc-smoke-test.sh
```

It starts local Dex, Keycloak, and Casdoor instances and validates provider compatibility.

Regular tests do not start Docker:

```bash
go test ./...
```

The provider smoke test uses a build tag internally:

```bash
go test -tags oidc_integration ./internal/oidc -run TestOIDCIntegrationProvider
```

## Troubleshooting

- `oidc provider discovery failed`: verify the provider `issuer` and `/.well-known/openid-configuration`.
- `OIDC state is invalid or expired`: restart login; the callback was reused, expired, or generated by another service instance.
- `OIDC token exchange failed`: verify client ID, client secret, redirect URL, and PKCE support.
- Login succeeds at the provider but fails in FNS: verify `email` and `sub` claims, and check whether `auto-register` should be enabled.
- The OIDC button is not shown: verify `oidc.enabled: true` and that the WebGUI can call `/api/user/auth/oidc/config` with `X-Client: WebGui`.
