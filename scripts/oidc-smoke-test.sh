#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"

DEX_PORT="${DEX_PORT:-15556}"
KEYCLOAK_PORT="${KEYCLOAK_PORT:-18081}"
CASDOOR_PORT="${CASDOOR_PORT:-18000}"
CASDOOR_MYSQL_PORT="${CASDOOR_MYSQL_PORT:-13306}"
REDIRECT_URL="${REDIRECT_URL:-http://localhost:18080/callback}"
CLIENT_ID="${CLIENT_ID:-fns-webgui}"
CLIENT_SECRET="${CLIENT_SECRET:-fns-secret}"
LOGIN_EMAIL="${LOGIN_EMAIL:-oidc@example.com}"
LOGIN_USERNAME="${LOGIN_USERNAME:-oidc-user}"
LOGIN_PASSWORD="${LOGIN_PASSWORD:-password}"

cleanup() {
  docker rm -f fns-oidc-dex fns-oidc-keycloak >/dev/null 2>&1 || true
  if [ -f "$WORK_DIR/casdoor/compose.yml" ]; then
    (cd "$WORK_DIR/casdoor" && docker compose down -v >/dev/null 2>&1 || true)
  fi
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

wait_url() {
  local url="$1"
  local name="$2"
  for _ in $(seq 1 90); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  echo "Timed out waiting for ${name}: ${url}" >&2
  return 1
}

run_auth_code_test() {
  local name="$1"
  local issuer="$2"
  local login="$3"
  local login_field="$4"

  echo "==> Testing ${name} authorization-code flow"
  (
    cd "$ROOT_DIR"
    OIDC_INTEGRATION_METHOD=auth_code \
      OIDC_INTEGRATION_ISSUER="$issuer" \
      OIDC_INTEGRATION_CLIENT_ID="$CLIENT_ID" \
      OIDC_INTEGRATION_CLIENT_SECRET="$CLIENT_SECRET" \
      OIDC_INTEGRATION_REDIRECT_URL="$REDIRECT_URL" \
      OIDC_INTEGRATION_LOGIN="$login" \
      OIDC_INTEGRATION_PASSWORD="$LOGIN_PASSWORD" \
      OIDC_INTEGRATION_LOGIN_FIELD="$login_field" \
      go test -tags oidc_integration ./internal/oidc -run TestOIDCIntegrationProvider -count=1
  )
}

run_password_test() {
	local name="$1"
	local issuer="$2"
	local token_url="$3"
	local jwks_url="$4"
	local client_id="$5"
	local client_secret="$6"
	local login="$7"
	local password="${8:-$LOGIN_PASSWORD}"

  echo "==> Testing ${name} password-grant token compatibility"
  (
    cd "$ROOT_DIR"
    OIDC_INTEGRATION_METHOD=password \
      OIDC_INTEGRATION_ISSUER="$issuer" \
      OIDC_INTEGRATION_TOKEN_URL="$token_url" \
      OIDC_INTEGRATION_JWKS_URL="$jwks_url" \
		OIDC_INTEGRATION_CLIENT_ID="$client_id" \
		OIDC_INTEGRATION_CLIENT_SECRET="$client_secret" \
		OIDC_INTEGRATION_LOGIN="$login" \
		OIDC_INTEGRATION_PASSWORD="$password" \
		go test -tags oidc_integration ./internal/oidc -run TestOIDCIntegrationProvider -count=1
	)
}

start_dex() {
	mkdir -p "$WORK_DIR/dex"
	cat > "$WORK_DIR/dex/config.yaml" <<EOF
issuer: http://localhost:${DEX_PORT}/dex
storage:
  type: sqlite3
  config:
    file: /tmp/dex.db
web:
  http: 0.0.0.0:5556
oauth2:
  skipApprovalScreen: true
staticClients:
  - id: ${CLIENT_ID}
    name: Fast Note Sync WebGUI
    secret: ${CLIENT_SECRET}
    redirectURIs:
      - ${REDIRECT_URL}
enablePasswordDB: true
staticPasswords:
  - email: ${LOGIN_EMAIL}
    hash: "\$2a\$10\$tu9Wibjqjmpdvuxq82yn6.1V0l0rH1YzZdwFSah1LEPZ6M3WOV/me"
    username: ${LOGIN_USERNAME}
    userID: oidc-user-1
EOF
  chmod 644 "$WORK_DIR/dex/config.yaml"
  docker rm -f fns-oidc-dex >/dev/null 2>&1 || true
  docker run -d --name fns-oidc-dex -p "${DEX_PORT}:5556" \
    -v "$WORK_DIR/dex/config.yaml:/etc/dex/cfg/config.yaml:ro" \
    ghcr.io/dexidp/dex:v2.44.0 dex serve /etc/dex/cfg/config.yaml >/dev/null
  wait_url "http://localhost:${DEX_PORT}/dex/.well-known/openid-configuration" "Dex"
}

start_keycloak() {
  mkdir -p "$WORK_DIR/keycloak"
  cat > "$WORK_DIR/keycloak/fns-realm.json" <<EOF
{
  "realm": "fns",
  "enabled": true,
  "clients": [
    {
      "clientId": "${CLIENT_ID}",
      "secret": "${CLIENT_SECRET}",
      "enabled": true,
      "protocol": "openid-connect",
      "publicClient": false,
      "clientAuthenticatorType": "client-secret",
      "standardFlowEnabled": true,
      "directAccessGrantsEnabled": true,
      "redirectUris": ["${REDIRECT_URL}"],
      "webOrigins": ["*"],
      "attributes": {
        "pkce.code.challenge.method": "S256"
      }
    }
  ],
	"users": [
		{
			"username": "${LOGIN_USERNAME}",
			"firstName": "OIDC",
			"lastName": "User",
			"email": "${LOGIN_EMAIL}",
			"enabled": true,
			"emailVerified": true,
			"requiredActions": [],
			"credentials": [
        {"type": "password", "value": "${LOGIN_PASSWORD}", "temporary": false}
      ]
    }
  ]
}
EOF
  chmod 644 "$WORK_DIR/keycloak/fns-realm.json"
	docker rm -f fns-oidc-keycloak >/dev/null 2>&1 || true
	docker run -d --name fns-oidc-keycloak -p "${KEYCLOAK_PORT}:8080" \
		-e KC_BOOTSTRAP_ADMIN_USERNAME=admin \
		-e KC_BOOTSTRAP_ADMIN_PASSWORD=admin \
		-v "$WORK_DIR/keycloak/fns-realm.json:/opt/keycloak/data/import/fns-realm.json:ro" \
		quay.io/keycloak/keycloak:26.4.5 \
		start-dev --import-realm --http-port=8080 --hostname="http://localhost:${KEYCLOAK_PORT}" --hostname-strict=false >/dev/null
  wait_url "http://localhost:${KEYCLOAK_PORT}/realms/fns/.well-known/openid-configuration" "Keycloak"
}

start_casdoor() {
  mkdir -p "$WORK_DIR/casdoor"
  cat > "$WORK_DIR/casdoor/compose.yml" <<EOF
services:
  mysql:
    image: mysql:8.4
    environment:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: casdoor
    ports:
      - "${CASDOOR_MYSQL_PORT}:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1", "-uroot", "-p123456"]
      interval: 5s
      timeout: 3s
      retries: 20
  casdoor:
    image: casbin/casdoor:v2.13.0
    depends_on:
      mysql:
        condition: service_healthy
    ports:
      - "${CASDOOR_PORT}:8000"
    volumes:
      - ./app.conf:/conf/app.conf:ro
EOF
  cat > "$WORK_DIR/casdoor/app.conf" <<EOF
appname = casdoor
httpport = 8000
runmode = dev
copyrequestbody = true
driverName = mysql
dataSourceName = root:123456@tcp(mysql:3306)/
dbName = casdoor
tableNamePrefix =
showSql = false
redisEndpoint =
defaultStorageProvider =
isCloudIntranet = false
authState = "casdoor"
socks5Proxy =
verificationCodeTimeout = 10
initScore = 0
logPostOnly = true
isUsernameLowered = false
origin = "http://localhost:${CASDOOR_PORT}"
originFrontend = "http://localhost:${CASDOOR_PORT}"
staticBaseUrl = "https://cdn.casbin.org"
isDemoMode = false
batchSize = 100
enableErrorMask = false
enableGzip = true
inactiveTimeoutMinutes =
ldapServerPort = 0
ldapsCertId = ""
ldapsServerPort = 0
radiusServerPort = 0
radiusDefaultOrganization = "built-in"
radiusSecret = "secret"
quota = {"organization": -1, "user": -1, "application": -1, "provider": -1}
logConfig = {"adapter":"console"}
initDataNewOnly = false
initDataFile = "./init_data.json"
frontendBaseDir = "../cc_0"
EOF
  chmod 644 "$WORK_DIR/casdoor/app.conf"
  (cd "$WORK_DIR/casdoor" && docker compose up -d >/dev/null)
  wait_url "http://localhost:${CASDOOR_PORT}/.well-known/openid-configuration" "Casdoor"

  local mysql_container client_id client_secret
  mysql_container="$(cd "$WORK_DIR/casdoor" && docker compose ps -q mysql)"
  docker exec "$mysql_container" mysql -uroot -p123456 casdoor -e "update application set grant_types='[\"authorization_code\",\"password\"]', redirect_uris='[\"${REDIRECT_URL}\"]' where name='app-built-in';" >/dev/null
  client_id="$(docker exec "$mysql_container" mysql -N -uroot -p123456 casdoor -e "select client_id from application where name='app-built-in';" 2>/dev/null)"
  client_secret="$(docker exec "$mysql_container" mysql -N -uroot -p123456 casdoor -e "select client_secret from application where name='app-built-in';" 2>/dev/null)"
  echo "${client_id}" > "$WORK_DIR/casdoor/client_id"
  echo "${client_secret}" > "$WORK_DIR/casdoor/client_secret"
}

start_dex
run_auth_code_test "Dex" "http://localhost:${DEX_PORT}/dex" "$LOGIN_EMAIL" "login"

start_keycloak
run_auth_code_test "Keycloak" "http://localhost:${KEYCLOAK_PORT}/realms/fns" "$LOGIN_USERNAME" "username"

start_casdoor
CASDOOR_CLIENT_ID="$(cat "$WORK_DIR/casdoor/client_id")"
CASDOOR_CLIENT_SECRET="$(cat "$WORK_DIR/casdoor/client_secret")"
run_password_test "Casdoor" \
  "http://localhost:${CASDOOR_PORT}" \
  "http://localhost:${CASDOOR_PORT}/api/login/oauth/access_token" \
  "http://localhost:${CASDOOR_PORT}/.well-known/jwks" \
	"$CASDOOR_CLIENT_ID" \
	"$CASDOOR_CLIENT_SECRET" \
	"admin" \
	"123"

echo "All OIDC smoke tests passed."
