#!/usr/bin/env bash
# Integration tests: real runner image + real downloader + real `java -jar`
# against emulated FlashSystems. No IBM hardware needed.
#
# Usage: tests/run-integration.sh [--no-build] [--keep]
#   --no-build   reuse the images from a previous run
#   --keep       leave the stack running afterwards (for debugging)
#
# Environment:
#   RUNTIME                       podman|docker (default: auto-detect)
#   IPQUORUM_DOWNLOADER_VERSION   release used for the runner base image build (default v3.1.1);
#                                 the downloader is then replaced by the one from this working tree
#   STARTUP_TIMEOUT               seconds to wait for a runner to connect (default 180; the
#                                 runner image is linux/amd64 and may run emulated)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EMU_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${EMU_DIR}/.." && pwd)"
COMPOSE_FILE="${EMU_DIR}/compose/podman-compose.test.yml"
PROJECT=ipqemu
NETWORK=ipqemu-net

EMULATOR_IMAGE=localhost/ipquorum-emulator:latest
RUNNER_BASE_IMAGE=localhost/ipquorum-runner:emulator-base
RUNNER_IMAGE=localhost/ipquorum-runner:emulator-test
DOWNLOADER_VERSION="${IPQUORUM_DOWNLOADER_VERSION:-v3.1.1}"
STARTUP_TIMEOUT="${STARTUP_TIMEOUT:-180}"

BUILD=true
KEEP=false
for arg in "$@"; do
    case "$arg" in
        --no-build) BUILD=false ;;
        --keep) KEEP=true ;;
        -h|--help) sed -n '2,15p' "${BASH_SOURCE[0]}"; exit 0 ;;
        *) echo "unknown option: $arg" >&2; exit 2 ;;
    esac
done

# --- runtime detection ------------------------------------------------------
RUNTIME="${RUNTIME:-}"
if [[ -z "$RUNTIME" ]]; then
    if command -v podman >/dev/null 2>&1; then RUNTIME=podman
    elif command -v docker >/dev/null 2>&1; then RUNTIME=docker
    else echo "neither podman nor docker found" >&2; exit 2; fi
fi
if [[ "$RUNTIME" == podman ]]; then
    if command -v podman-compose >/dev/null 2>&1; then COMPOSE=(podman-compose)
    else COMPOSE=(podman compose); fi
else
    if docker compose version >/dev/null 2>&1; then COMPOSE=(docker compose)
    else COMPOSE=(docker-compose); fi
fi
compose() { "${COMPOSE[@]}" -f "$COMPOSE_FILE" -p "$PROJECT" "$@"; }

if ! "$RUNTIME" info >/dev/null 2>&1; then
    echo "cannot talk to $RUNTIME (is the podman machine / docker daemon running?)" >&2
    exit 2
fi

# --- tiny test framework ----------------------------------------------------
PASSED=0
FAILED=0
FAILURES=()
section() { printf '\n== %s\n' "$*"; }
pass() { PASSED=$((PASSED + 1)); printf '  PASS  %s\n' "$*"; }
fail() { FAILED=$((FAILED + 1)); FAILURES+=("$*"); printf '  FAIL  %s\n' "$*"; }
check() { # description, command...
    local desc="$1"; shift
    if "$@" >/dev/null 2>&1; then pass "$desc"; else fail "$desc"; fi
}

logs() { "$RUNTIME" logs "$1" 2>&1; }
count_log() { local out; out="$(logs "$1")"; grep -Ec -- "$2" <<<"$out" || true; }
wait_count() { # container pattern minimum timeout
    local end=$((SECONDS + $4))
    while (( SECONDS < end )); do
        if (( $(count_log "$1" "$2") >= $3 )); then return 0; fi
        sleep 1
    done
    return 1
}
has_log() { local out; out="$(logs "$1")"; grep -Eq -- "$2" <<<"$out"; }
# node addresses the fake JAR was built with, e.g. "flash-a:1260 flash-a:1261"
jar_nodes() { logs "$1" | sed -n 's/.*Quorum FINE: Node \([^ ]*\) (.*/\1/p' | sort | tr '\n' ' '; }

control() { # port method path [json]
    local port="$1" method="$2" path="$3" body="${4:-}"
    if [[ -n "$body" ]]; then
        curl -fsS -X "$method" "http://127.0.0.1:${port}${path}" -d "$body"
    else
        curl -fsS -X "$method" "http://127.0.0.1:${port}${path}"
    fi
}
CTL_A=19001; CTL_B=19002; CTL_C=19003

cleanup() {
    local rc=$?
    "$RUNTIME" rm -f ipqemu-oneshot >/dev/null 2>&1 || true
    if [[ "$KEEP" == true ]]; then
        echo; echo "Stack left running (project $PROJECT). Stop with:"
        echo "  ${COMPOSE[*]} -f $COMPOSE_FILE -p $PROJECT down -v"
    else
        compose down -v >/dev/null 2>&1 || true
    fi
    exit $rc
}
trap cleanup EXIT

# --- build ------------------------------------------------------------------
if [[ "$BUILD" == true ]]; then
    section "Building images"
    echo "  emulator"
    "$RUNTIME" build -q -f "${EMU_DIR}/Containerfile" -t "$EMULATOR_IMAGE" "$EMU_DIR" >/dev/null || { echo "emulator image build failed" >&2; exit 2; }
    echo "  runner base (production Dockerfile, downloader ${DOWNLOADER_VERSION})"
    "$RUNTIME" build -q --platform linux/amd64 --build-arg "IPQUORUM_DOWNLOADER_VERSION=${DOWNLOADER_VERSION}" \
        -f "${REPO_ROOT}/ipquorum-management-platform/ipquorum-runner/Dockerfile" -t "$RUNNER_BASE_IMAGE" \
        "${REPO_ROOT}/ipquorum-management-platform/ipquorum-runner" >/dev/null || { echo "runner base image build failed" >&2; exit 2; }
    echo "  runner test overlay (downloader + entrypoint from this working tree)"
    "$RUNTIME" build -q --platform linux/amd64 --build-arg "BASE=${RUNNER_BASE_IMAGE}" \
        -f "${EMU_DIR}/compose/Containerfile.runner-test" -t "$RUNNER_IMAGE" "$REPO_ROOT" >/dev/null || { echo "runner test image build failed" >&2; exit 2; }
fi

# --- start ------------------------------------------------------------------
section "Starting emulators and runners"
compose down -v >/dev/null 2>&1 || true
compose up -d >/dev/null 2>&1 || { echo "compose up failed" >&2; compose up -d; exit 2; }

for port in $CTL_A $CTL_B $CTL_C; do
    for _ in $(seq 1 30); do
        curl -fsS "http://127.0.0.1:${port}/healthz" >/dev/null 2>&1 && break
        sleep 1
    done
done

# --- 1. each runner connects to ITS OWN system -------------------------------
section "Multiple systems: each runner connects only to its own nodes"
for r in a b; do
    wait_count "ipqemu-runner-$r" "^Connected to flash-$r\$" 2 "$STARTUP_TIMEOUT" \
        && pass "runner-$r connected to both nodes of flash-$r" \
        || fail "runner-$r connected to both nodes of flash-$r"
done
wait_count ipqemu-runner-c '^Connected to flash-c$' 1 "$STARTUP_TIMEOUT" \
    && pass "runner-c connected to flash-c" || fail "runner-c connected to flash-c"

[[ "$(jar_nodes ipqemu-runner-a)" == "flash-a:1260 flash-a:1261 " ]] \
    && pass "runner-a JAR lists exactly the flash-a nodes" || fail "runner-a JAR lists exactly the flash-a nodes (got: $(jar_nodes ipqemu-runner-a))"
[[ "$(jar_nodes ipqemu-runner-b)" == "flash-b:1260 flash-b:1261 " ]] \
    && pass "runner-b JAR lists exactly the flash-b nodes" || fail "runner-b JAR lists exactly the flash-b nodes (got: $(jar_nodes ipqemu-runner-b))"
check "runner-a reports 2 nodes" has_log ipqemu-runner-a 'found 2 nodes'

section "API_PORT: runner-c uses a non-default REST port, runner-a/b the default"
check "runner-c log shows API_PORT 8443" has_log ipqemu-runner-c 'API_PORT: 8443'
check "runner-a log shows the default port" has_log ipqemu-runner-a 'API_PORT: default \(7443\)'

section "Emulator state (control API)"
STATE_A="$(control $CTL_A GET /clusters/flash-a 2>/dev/null || true)"
check "flash-a: quorum app created and 1 download" grep -Eq '"quorum_app_created":true.*"downloads":1' <<<"$STATE_A"
check "flash-a: mkquorumapp received partnersystem flash-b" grep -q '"partnersystem":"flash-b"' <<<"$STATE_A"
STATE_B="$(control $CTL_B GET /clusters/flash-b 2>/dev/null || true)"
check "flash-b: mkquorumapp received partnersystem flash-a" grep -q '"partnersystem":"flash-a"' <<<"$STATE_B"

# --- 2. node failure on one system only ------------------------------------
section "Node failure on flash-a does not affect flash-b"
B_BEFORE=$(count_log ipqemu-runner-b 'Failed to open socket|Lost connection')
control $CTL_A POST /clusters/flash-a/nodes/flash-a:1261/behavior '{"behavior":"down"}' >/dev/null
wait_count ipqemu-runner-a 'Failed to open socket to flash-a:1261' 1 60 \
    && pass "runner-a notices node flash-a:1261 is down" || fail "runner-a notices node flash-a:1261 is down"
sleep 3
B_AFTER=$(count_log ipqemu-runner-b 'Failed to open socket|Lost connection')
[[ "$B_BEFORE" == "$B_AFTER" ]] && pass "runner-b saw no failures" || fail "runner-b saw no failures (before=$B_BEFORE after=$B_AFTER)"

CONNECTED_BEFORE=$(count_log ipqemu-runner-a '^Connected to flash-a$')
control $CTL_A POST /clusters/flash-a/nodes/flash-a:1261/behavior '{"behavior":"up"}' >/dev/null
wait_count ipqemu-runner-a '^Connected to flash-a$' $((CONNECTED_BEFORE + 1)) 60 \
    && pass "runner-a reconnects when the node is back" || fail "runner-a reconnects when the node is back"

section "handshake-timeout on a node"
control $CTL_B POST /clusters/flash-b/nodes/flash-b:1260/behavior '{"behavior":"handshake-timeout"}' >/dev/null
wait_count ipqemu-runner-b 'Handshake timed out for flash-b:1260|Lost connection to flash-b:1260' 1 60 \
    && pass "runner-b reports the unresponsive node" || fail "runner-b reports the unresponsive node"
control $CTL_B POST /clusters/flash-b/reset >/dev/null

# --- 3. REST failure scenarios (one-shot runners against flash-c) ----------
oneshot() { # env assignments... ; prints output, sets ONESHOT_RC
    local envs=() kv
    for kv in "$@"; do envs+=(-e "$kv"); done
    ONESHOT_OUT="$("$RUNTIME" run --rm --name ipqemu-oneshot --platform linux/amd64 --network "$NETWORK" \
        -e VIRTUALIZE_USERNAME=superuser -e VIRTUALIZE_PASSWORD=passw0rd-c \
        -e API_ENDPOINT=flash-c -e API_PORT=8443 -e PARTNERSYSTEM=flash-a \
        -e CREATE_QUORUM_APP=true -e DOWNLOAD_JAR=true ${envs[@]+"${envs[@]}"} "$RUNNER_IMAGE" 2>&1)"
    ONESHOT_RC=$?
}
expect_failure() { # description scenario output-pattern [env...]
    local desc="$1" scenario="$2" pattern="$3"; shift 3
    control $CTL_C POST /clusters/flash-c/scenario "{\"scenario\":\"${scenario}\"}" >/dev/null
    oneshot "$@"
    if (( ONESHOT_RC != 0 )) && grep -Eq -- "$pattern" <<<"$ONESHOT_OUT"; then pass "$desc"
    else fail "$desc (rc=$ONESHOT_RC, wanted /$pattern/)"; printf '%s\n' "$ONESHOT_OUT" | tail -5 | sed 's/^/        /'; fi
    control $CTL_C POST /clusters/flash-c/reset >/dev/null
}

section "REST failure scenarios: the runner must fail clearly"
expect_failure "auth 401 fails fast"                     auth-401          'Invalid username or password'
expect_failure "auth 403 fails fast"                     auth-403          'Insufficient permissions'
expect_failure "wrong password is rejected"              healthy           'Insufficient permissions or invalid credentials' VIRTUALIZE_PASSWORD=wrong
expect_failure "mkquorumapp error"                       mkquorumapp-error 'mkquorumapp failed with status 500'
expect_failure "unknown partnersystem is rejected (strict)" healthy        'mkquorumapp failed with status 400' PARTNERSYSTEM=ghost
expect_failure "download error"                          download-error    'Download failed with status 500'
expect_failure "empty download is detected"              download-empty    'JAR file is empty'

section "Rate limiting (429) is retried and then succeeds"
control $CTL_C POST /clusters/flash-c/scenario '{"scenario":"rate-limit-429"}' >/dev/null
"$RUNTIME" rm -f ipqemu-oneshot >/dev/null 2>&1
"$RUNTIME" run -d --name ipqemu-oneshot --platform linux/amd64 --network "$NETWORK" \
    -e VIRTUALIZE_USERNAME=superuser -e VIRTUALIZE_PASSWORD=passw0rd-c -e API_ENDPOINT=flash-c -e API_PORT=8443 \
    -e PARTNERSYSTEM=flash-a -e CREATE_QUORUM_APP=true -e DOWNLOAD_JAR=true \
    -e "JAVA_OPTS=-Xmx128m -Dfake.keepalive.ms=500" "$RUNNER_IMAGE" >/dev/null
wait_count ipqemu-oneshot '^Connected to flash-c$' 1 "$STARTUP_TIMEOUT" \
    && pass "runner connected after being rate limited" || fail "runner connected after being rate limited"
check "runner logged the 429 responses" has_log ipqemu-oneshot 'Rate limited \(429\)'
"$RUNTIME" rm -f ipqemu-oneshot >/dev/null 2>&1
control $CTL_C POST /clusters/flash-c/reset >/dev/null

# --- 4. graceful shutdown ---------------------------------------------------
section "Graceful shutdown"
"$RUNTIME" stop -t 40 ipqemu-runner-b >/dev/null 2>&1
EXIT_CODE="$("$RUNTIME" inspect -f '{{.State.ExitCode}}' ipqemu-runner-b 2>/dev/null)"
[[ "$EXIT_CODE" == 0 ]] && pass "runner-b exits with 0 on SIGTERM" || fail "runner-b exits with 0 on SIGTERM (exit code ${EXIT_CODE:-?})"

# --- summary ----------------------------------------------------------------
printf '\n== Result: %d passed, %d failed\n' "$PASSED" "$FAILED"
if (( FAILED > 0 )); then
    printf 'Failed:\n'; printf '  - %s\n' "${FAILURES[@]}"
    for c in ipqemu-runner-a ipqemu-runner-b ipqemu-runner-c ipqemu-flash-a ipqemu-flash-b ipqemu-flash-c; do
        printf '\n---- last log lines of %s\n' "$c"; logs "$c" | tail -15
    done
    exit 1
fi
