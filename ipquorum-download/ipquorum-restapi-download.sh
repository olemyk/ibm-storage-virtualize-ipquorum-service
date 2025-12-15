
#!/usr/bin/env bash
set -euo pipefail

# ===== Defaults (can be overridden via CLI flags) =====
API_ENDPOINT="${API_ENDPOINT:-}"
VIRTUALIZE_USERNAME="${VIRTUALIZE_USERNAME:-}"
VIRTUALIZE_PASSWORD="${VIRTUALIZE_PASSWORD:-}"     # require via --pass
IPQ_OUTPUT_FILE="${IPQ_OUTPUT_FILE:-ip_quorum.jar}"
DOWNLOADIPQ="${DOWNLOADIPQ:-yes}"
insecure_api="${insecure_api:---insecure}" # set to "" if you trust TLS
API_PORT="${API_PORT:-7443}"
INSECURE="${INSECURE:---insecure}"  # set to "" if you trust TLS

# Retry for API Call.
MAX_RETRIES="${MAX_RETRIES:-8}"
BASE_DELAY="${BASE_DELAY:-3}" # seconds

# mkquorumapp default only if unset (CLI can override)
mkquorumapp="${mkquorumapp:-yes}"

# mkquorumapp payload defaults
ip6="${ip6:-false}"                # --ip6 / --ip_6: true|false 
nometadata="${nometadata:-false}"  # --nometadata: true|false
partnersystem="${partnersystem:-}" # --partnersystem: <string>  (MANDATORY if mkquorumapp=yes)
partnerip6="${partnerip6:-false}"  # --partnerip6: true|false

## Partnersystem is the Remote System in PBHA

# ===== Helpers =====
to_bool() {
  # Normalize truthy/falsey inputs to "true"/"false"
  local v="${1:-}"
  v="$(printf '%s' "$v" | tr '[:upper:]' '[:lower:]' | awk '{$1=$1}1')"
  case "$v" in
    true|1|yes|y|on) echo "true" ;;
    false|0|no|n|off|"") echo "false" ;;
    *) echo "false" ;;
  esac
}

print_usage() {
  cat <<'EOF'
Usage: ipquorum-restapi-download.sh [options]

General:
  --mkquorumapp / --no-mkquorumapp     Enable/disable mkquorumapp call (default: enabled)
  --download / --no-download           Enable/disable jar download (default: enabled)
  --insecure / --secure                Use insecure TLS (-k) or strict TLS (default: insecure)
  --api-endpoint <host>                Set API endpoint
  --user <username>                    Auth username - Monitor Role for download or Minmum Restricted Administrator to Create new Quorum App. 
  --pass <password>                    Auth password (REQUIRED)
  --output <file>                      Output jar filename (default: ip_quorum.jar)

mkquorumapp payload:
  --ip6[=true|false] or --ip_6[=true|false]   Set IPv6 flag (default: false)
  --nometadata[=true|false]                   Set nometadata flag (default: false)
  --partnersystem <name>                      Set Partnersystem - Remote System in PBHA (MANDATORY if mkquorumapp is enabled)
  --partnerip6[=true|false]                   Set partner IPv6 flag (default: false)

Examples:
  ./ipquorum-restapi-download.sh \
    --mkquorumapp --partnersystem svc_cluster02 \
    --ip6=false --partnerip6=false --nometadata=false \
    --download --insecure --user superuser --pass password

  ./ipquorum-restapi-download.sh \
    --no-mkquorumapp --download --output ip_quorum.jar --user superuser --pass password
EOF
}

# ===== CLI parsing =====

while [[ $# -gt 0 ]]; do
  case "$1" in
    --help|-h) print_usage; exit 0 ;;
    --no-mkquorumapp) mkquorumapp='no'; shift ;;
    --mkquorumapp) mkquorumapp='yes'; shift ;;
    --download) DOWNLOADIPQ='yes'; shift ;;
    --no-download) DOWNLOADIPQ='no'; shift ;;
    --insecure) insecure_api="--insecure"; shift ;;
    --secure) insecure_api=""; shift ;;
    --api-endpoint|--api-endpoint=*)
      if [[ "$1" == *=* ]]; then API_ENDPOINT="${1#*=}"; shift; else API_ENDPOINT="${2:-}"; shift 2; fi ;;
    --user|--user=*)
      if [[ "$1" == *=* ]]; then VIRTUALIZE_USERNAME="${1#*=}"; shift; else VIRTUALIZE_USERNAME="${2:-}"; shift 2; fi ;;
    --pass|--pass=*)
      if [[ "$1" == *=* ]]; then VIRTUALIZE_PASSWORD="${1#*=}"; shift; else VIRTUALIZE_PASSWORD="${2:-}"; shift 2; fi ;;
    --output|--output=*)
      if [[ "$1" == *=* ]]; then IPQ_OUTPUT_FILE="${1#*=}"; shift; else IPQ_OUTPUT_FILE="${2:-}"; shift 2; fi ;;
    --ip6|--ip6=*|--ip_6|--ip_6=*)
      if [[ "$1" == *=* ]]; then ip6="$(to_bool "${1#*=}")"; shift; else ip6="$(to_bool "${2:-}")"; shift 2; fi ;;
    --nometadata|--nometadata=*)
      if [[ "$1" == *=* ]]; then nometadata="$(to_bool "${1#*=}")"; shift; else nometadata="$(to_bool "${2:-}")"; shift 2; fi ;;
    --partnersystem|--partnersystem=*)
      if [[ "$1" == *=* ]]; then partnersystem="${1#*=}"; shift; else partnersystem="${2:-}"; shift 2; fi ;;
    --partnerip6|--partnerip6=*)
      if [[ "$1" == *=* ]]; then partnerip6="$(to_bool "${1#*=}")"; shift; else partnerip6="$(to_bool "${2:-}")"; shift 2; fi ;;
    *) echo "Unknown option: $1" >&2; print_usage; exit 2 ;;
  esac
done

# Normalize mkquorumapp flag
mkq_norm=$(printf '%s' "${mkquorumapp:-}" | tr '[:upper:]' '[:lower:]' | awk '{$1=$1}1')

echo "DEBUG mkquorumapp='${mkquorumapp}' (norm='${mkq_norm}') DOWNLOADIPQ='${DOWNLOADIPQ}' insecure_api='${insecure_api:-}'"

# ===== Required fields check =====
if [[ -z "${VIRTUALIZE_PASSWORD}" ]]; then
  echo "Error: --pass <password> is required." >&2
  exit 2
fi


echo "DEBUG mkquorumapp='${mkquorumapp}' (norm='${mkq_norm}') DOWNLOADIPQ='${DOWNLOADIPQ}' insecure_api='${insecure_api:-}'"
echo "DEBUG payload ip6='${ip6}' nometadata='${nometadata}' partnersystem='${partnersystem}' partnerip6='${partnerip6}'"
echo "DEBUG user='${VIRTUALIZE_USERNAME}' endpoint='${API_ENDPOINT}'"

# ===== Pre-flight validation =====
preflight() {
  echo "Pre-flight: validating inputs & endpoint reachability ..."

  if [[ "${mkq_norm}" == "yes" && -z "${partnersystem}" ]]; then
    echo "Error: --partnersystem <name> is MANDATORY when --mkquorumapp is enabled." >&2
    exit 2
  fi

  for k in ip6 nometadata partnerip6; do
    v="${!k}"
    if [[ "$v" != "true" && "$v" != "false" ]]; then
      echo "Error: $k must be true|false (got '$v')." >&2
      exit 2
    fi
  done

  local hdr_file; hdr_file=$(mktemp)
  # Check endpoint reachability; parse status from headers (no --write-out)
  curl -sS ${insecure_api:-} \
    -X GET "https://${API_ENDPOINT}:7443/rest/v1/auth" \
    -D "$hdr_file" \
    -o /dev/null || true

  echo "Pre-flight: endpoint headers:"
  sed 's/^/  /' "$hdr_file"
  local http_code; http_code=$(awk 'NR==1{print $2}' "$hdr_file")
  echo "Pre-flight: endpoint status: ${http_code:-unknown}"
  rm -f "$hdr_file"

  case "${http_code}" in
    200|401|404|405) echo "Pre-flight: endpoint is reachable."; ;;
    "")
      echo "Pre-flight: no HTTP status returned (network/TLS error). Try --insecure or verify connectivity." >&2
      exit 3
      ;;
    *)
      echo "Pre-flight: unexpected status ${http_code}. Proceeding may fail." >&2
      ;;
  esac
}
preflight


#---------

echo "Get Token, please wait ..." >&2

# ---- Require jq ----
if ! command -v jq >/dev/null 2>&1; then
  echo "Error: 'jq' is required to parse JSON. Please install jq." >&2
  exit 2
fi

get_token_with_retry() {
  local attempt=1
  while (( attempt <= MAX_RETRIES )); do
    echo "Auth attempt ${attempt}/${MAX_RETRIES} ..." >&2
    local hdr_file body_file curl_rc http_code resp_type token retry_after delay
    hdr_file="$(mktemp)"
    body_file="$(mktemp)"

    curl -sS ${insecure_api} -L \
      -X POST "https://${API_ENDPOINT}:${API_PORT}/rest/v1/auth" \
      -H "Accept: application/json" \
      -H "Content-Type: application/json" \
      -H "X-Auth-Username: ${VIRTUALIZE_USERNAME}" \
      -H "X-Auth-Password: ${VIRTUALIZE_PASSWORD}" \
      -D "${hdr_file}" \
      -o "${body_file}"
    curl_rc=$?

    http_code="$(awk '/^HTTP\/[0-9.]+ /{code=$2} END{print code}' "${hdr_file}")"
    resp_type="$(jq -r 'type' "${body_file}" 2>/dev/null || echo "unknown")"

    if [[ "${resp_type}" == "array" ]] && jq -e 'map(tostring)|join(",")|test("Too Many Requests")' "${body_file}" >/dev/null 2>&1; then
      http_code="429"
    fi

    token="$(
      awk '
        BEGIN{IGNORECASE=1}
        /^X-Auth-Token:/   {sub(/^[^:]*:[[:space:]]*/, "", $0); print $0}
        /^Authorization:/  {sub(/^[^:]*:[[:space:]]*/, "", $0); print $0}
      ' "${hdr_file}" | head -n1
    )"

    if [[ -z "${token}" || "${token}" == "null" ]]; then
      if [[ "${resp_type}" == "object" ]]; then
        token="$(jq -r '
          .token // .access_token // .data.token // .result.token // .authToken // .session // empty
        ' "${body_file}" 2>/dev/null || true)"
      else
        token=""
      fi
    fi

    if [[ "${curl_rc}" -eq 0 && ( "${http_code}" == "200" || "${http_code}" == "201" ) && -n "${token}" && "${token}" != "null" ]]; then
      rm -f "${hdr_file}" "${body_file}"
      # IMPORTANT: token to stdout only
      echo "${token}"
      return 0
    fi

    if [[ "${http_code}" == "429" ]]; then
      retry_after="$(
        awk '
          BEGIN{IGNORECASE=1}
          /^Retry-After:/ {sub(/^[^:]*:[[:space:]]*/, "", $0); gsub(/\r/,""); print $0}
        ' "${hdr_file}" | tail -n1
      )"
      if [[ -n "${retry_after}" ]]; then
        delay="${retry_after}"
      else
        delay=$(( BASE_DELAY * attempt + (RANDOM % 3) ))
      fi
      echo "Rate limited (429). Sleeping ${delay}s and retrying ..." >&2
      rm -f "${hdr_file}" "${body_file}"
      sleep "${delay}"
      attempt=$(( attempt + 1 ))
      continue
    fi

    echo "Attempt ${attempt} failed (curl_rc=${curl_rc}, http_code=${http_code}, resp_type=${resp_type}). Retrying ..." >&2
    rm -f "${hdr_file}" "${body_file}"
    sleep $(( BASE_DELAY + attempt + (RANDOM % 2) ))
    attempt=$(( attempt + 1 ))
  done

  echo "ERROR: Failed to obtain token after ${MAX_RETRIES} attempts." >&2
  return 1
}

ACCESS_TOKEN="$(get_token_with_retry)" || {
  echo "Failed to retrieve the access token." >&2
  exit 1
}

# If you want to print it, echo to stdout (no extra text)
echo "${ACCESS_TOKEN}"




# ===== mkquorumapp (NO --write-out; parse headers) =====

# Assumes these vars are already set:
#   API_ENDPOINT, access_token, ip6, nometadata, partnersystem, partnerip6, insecure_api

if [[ "${mkq_norm}" == "yes" ]]; then
  echo "Creating IP-Quorum app via /rest/v1/mkquorumapp ..."

  # Build a clean JSON payload
  mkq_payload=$(printf '{ "ip_6": %s, "nometadata": %s, "partnersystem": "%s", "partnerip6": %s }' \
                "${ip6}" "${nometadata}" "${partnersystem}" "${partnerip6}")

  echo "DEBUG payload: ${mkq_payload}"

  # Simple verbose curl call (no --write-out, no header/body files)
  curl -k ${insecure_api:-} \
    -X POST "https://${API_ENDPOINT}:7443/rest/v1/mkquorumapp" \
    -H 'accept: application/json' \
    -H "X-Auth-Token: ${ACCESS_TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "${mkq_payload}"
  curl_rc=$?


  if [[ "${curl_rc}" -eq 0 ]]; then
    echo "mkquorumapp call completed (curl exit code=0). Inspect the verbose output above for HTTP status/body."
  else
    echo "mkquorumapp call failed (curl exit code=${curl_rc}). See verbose output above."
    exit 1
  fi
else
  echo "Condition not met. Skipping the Create new IP-Quorum app call."
fi


# ===== Download block (NO --write-out; parse headers) =====
if [[ "${DOWNLOADIPQ}" == "yes" ]]; then
  echo "Proceeding to download ip_quorum.jar ..."
  

  hdr_file=$(mktemp)

  # Save body directly to file, headers to hdr_file; no --write-out
  curl -sS ${insecure_api:-} \
    --fail \
    -X POST "https://${API_ENDPOINT}:7443/rest/v1/download" \
    -H "accept: application/json" \
    -H "X-Auth-Token: ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"prefix":"/dumps","filename":"ip_quorum.jar"}' \
    -D "$hdr_file" \
    --output "${IPQ_OUTPUT_FILE}"
  curl_rc=$?

  echo "Download curl exit code: ${curl_rc}"
  echo "Download Headers:"
  sed 's/^/  /' "$hdr_file"

  # Parse status from first header line (HTTP/1.1 200 or HTTP/2 200)
  dl_http=$(awk 'NR==1{print $2}' "$hdr_file")
  rm -f "$hdr_file"

  if [[ -z "${dl_http}" ]]; then
    echo "Download: could not parse HTTP status code."
    exit 1
  fi

  if [[ "$curl_rc" -eq 0 && ( "$dl_http" == "200" || "$dl_http" == "201" ) ]]; then
    if [[ -s "${IPQ_OUTPUT_FILE}" ]]; then
      file_size=$(wc -c < "${IPQ_OUTPUT_FILE}")
      echo "Downloaded ${IPQ_OUTPUT_FILE} (size: ${file_size} bytes)"
    else
      echo "Success status but empty file. Check server response."
    fi
  else
    echo "Download failed: curl_rc=${curl_rc}, HTTP=${dl_http}"
    exit 1
  fi
else
  echo "Condition not met. Skipping the downloading of quorumapp."
fi
