#!/usr/bin/env bash
set -euo pipefail

HYDRA_ADMIN="http://localhost:4445"

register_client() {
  local client_id="$1"
  local client_secret="$2"

  existing=$(curl -s -o /dev/null -w "%{http_code}" "${HYDRA_ADMIN}/admin/clients/${client_id}")
  if [ "$existing" = "200" ]; then
    echo "Client '${client_id}' already exists, updating..."
    curl -s -X PUT "${HYDRA_ADMIN}/admin/clients/${client_id}" \
      -H "Content-Type: application/json" \
      -d "{
        \"client_id\": \"${client_id}\",
        \"client_secret\": \"${client_secret}\",
        \"grant_types\": [\"client_credentials\"],
        \"token_endpoint_auth_method\": \"client_secret_post\",
        \"scope\": \"service\"
      }" > /dev/null
    echo "Client '${client_id}' updated."
  else
    echo "Registering client '${client_id}'..."
    curl -s -X POST "${HYDRA_ADMIN}/admin/clients" \
      -H "Content-Type: application/json" \
      -d "{
        \"client_id\": \"${client_id}\",
        \"client_secret\": \"${client_secret}\",
        \"grant_types\": [\"client_credentials\"],
        \"token_endpoint_auth_method\": \"client_secret_post\",
        \"scope\": \"service\"
      }" > /dev/null
    echo "Client '${client_id}' registered."
  fi
}

register_client "bff-client" "bff-secret"
register_client "ec-site-client" "ec-site-secret"

echo "Hydra setup complete."
