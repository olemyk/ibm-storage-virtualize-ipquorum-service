
#!/usr/bin/env bash
set -euo pipefail

export IP="10.33.7.80"
export PORT="7443"
export USERNAME="username"
export PASSWORD="password"
export PARTNERSYSTEM="svc_cluster02"

echo "[1/3] Authenticating..."
AUTH_RESPONSE=$(curl -ks -X POST "https://${IP}:${PORT}/rest/v1/auth" \
  -H "accept: application/json" \
  -H "X-Auth-Username: ${USERNAME}" \
  -H "X-Auth-Password: ${PASSWORD}" \
  -d "")

TOKEN=$(echo "${AUTH_RESPONSE}" | jq -r '.token // .X_Auth_Token // .authToken // empty')
if [ -z "${TOKEN}" ] || [ "${TOKEN}" = "null" ]; then
  echo "ERROR: Could not parse token from auth response:"
  echo "${AUTH_RESPONSE}"
  exit 1
fi
export TOKEN
echo "Token: captured."

echo "[2/3] Creating IPQuorum..."
curl -ks -X POST "https://${IP}:${PORT}/rest/v1/mkquorumapp" \
  -H "accept: application/json" \
  -H "X-Auth-Token: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "$(jq -n --arg ps "${PARTNERSYSTEM}" '{ip_6:false, nometadata:false, partnersystem:$ps, partnerip6:false}')"

echo "[3/3] Downloading ip_quorum.jar..."
OUTPUT_FILE="ip_quorum.jar"
curl -ks -X POST "https://${IP}:${PORT}/rest/v1/download" \
  -H "accept: application/json" \
  -H "X-Auth-Token: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"prefix":"/dumps","filename":"ip_quorum.jar"}' \
  --output "${OUTPUT_FILE}"

if [ -f "${OUTPUT_FILE}" ]; then
  echo "Success: downloaded ${OUTPUT_FILE}"
else
  echo "ERROR: Download failed."
  exit 1
fi
