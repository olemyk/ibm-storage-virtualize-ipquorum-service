# IP Quorum Emulator Plan

## Top-Level Overview

**Goal:** Develop and test the IP quorum tooling (downloader, runner container, management platform) without a real IBM FlashSystem, a real `ip_quorum.jar`, or network access to a storage system.

**Approach:** Add a new top-level directory `ipquorum-emulator/` that emulates one **or several** FlashSystem clusters:

- **mock-flashsystem (Go):** emulates the REST endpoints that `ipquorum-download-go` calls (`/rest/v1/auth`, `/rest/v1/mkquorumapp`, `/rest/v1/download`). It hosts N virtual clusters, each with its own name, credentials, partner system, node list and failure scenario.
- **fake `ip_quorum.jar` (Java):** a small program built in CI (no binary JAR is committed). It prints log output in the real `ip_quorum.jar` format (`=== IP quorum ===`, `Trying to open socket`, `Handshaking`, `Connected to <ip>`, `Quorum FINE: Node ...`), stays running and exits on SIGTERM.
- **mock nodes (TCP):** listeners per cluster node that the fake JAR connects to. They use a simplified protocol of our own, because the real wire protocol is not publicly documented.
- **Per-cluster JAR:** each cluster serves a JAR with its own node list embedded, as the real JAR does. The mock builds it at download time from a template fake JAR, by adding a config entry to the ZIP. This means two runner containers pointed at two clusters connect to different node sets.

The management platform needs no code changes. Setting `API_ENDPOINT` to a mock cluster is enough.

### Multi-cluster requirements (from user)

- One mock process can host several clusters, defined in a config file (name, credentials, partner, nodes, scenario).
- Several mock processes or containers can run side by side, for example `flash-a` and `flash-b` in compose, each on the default REST port 7443.
- A cluster can name another mock cluster as its `partnersystem`. This covers the HyperSwap and partner-system case.
- Failure scenarios are set per cluster, so one cluster can be healthy while another has a node down.
- State is per cluster. For example, `download` before `mkquorumapp` returns an error until the quorum app has been created.

### Decisions already made

| Topic | Decision |
|---|---|
| Emulation depth | REST mock, fake JAR and mock TCP nodes on port 1260 (simplified own protocol) |
| REST port | Add optional `--api-port` / `API_PORT` to the downloader and runner (default stays 7443). The compose test setup keeps 7443 |
| Location | New top-level `ipquorum-emulator/` |
| Fake JAR | Built by a multi-stage Containerfile with a JDK, built in CI, nothing committed |
| Config format | YAML |
| Partner system check | Strict: `mkquorumapp` rejects a `partnersystem` that is not a known mock cluster (matches a real system) |

### Architecture

```
ipquorum-emulator/
  mock-flashsystem/     Go: multi-cluster REST mock + mock node listeners + control API
  fake-ipquorum-jar/    Java: fake ip_quorum.jar template
  scenarios/            example multi-cluster config files (healthy, node-down, auth-fail, ...)
  compose/              podman-compose.test.yml (flash-a, flash-b, runners)
  tests/                integration test scripts
  README.md
```

---

## Sub-Tasks

### 1. Downloader and runner: configurable REST port

- **Intent:** The downloader hardcodes port 7443, so a mock cannot run on another port. Add an optional override and keep the default unchanged.
- **Expected Outcomes:**
  - `ipquorum-download-go --api-port 8443` and the env var `API_PORT` work.
  - The default is still 7443, and existing behavior is unchanged.
  - The runner `entrypoint.sh` forwards `API_PORT` when set.
- **Todo List:**
  1. Add the `--api-port` flag in `main.go`, defaulting to the `API_PORT` env var and then to 7443.
  2. Validate the port range in `Config.Validate()`.
  3. In `entrypoint.sh`, read `API_PORT` and pass `--api-port` only if set. Add `API_PORT` to the startup banner and to the Dockerfile `ENV` defaults.
  4. Add a unit test for the config validation.
- **Relevant Context:**
  - [`ipquorum-download-go/pkg/config/config.go`](ipquorum-download-go/pkg/config/config.go:41) (`APIPort`, default 7443)
  - [`ipquorum-download-go/pkg/client/client.go`](ipquorum-download-go/pkg/client/client.go:42) (`baseURL`)
  - [`ipquorum-download-go/main.go`](ipquorum-download-go/main.go:58) (flag definitions)
  - [`ipquorum-management-platform/ipquorum-runner/entrypoint.sh`](ipquorum-management-platform/ipquorum-runner/entrypoint.sh)
  - [`ipquorum-management-platform/ipquorum-runner/Dockerfile`](ipquorum-management-platform/ipquorum-runner/Dockerfile)
- **Status:** [x] done
- **Implementation notes (for later sub-tasks):**
  - `--api-port` flag in `main.go`; env `API_PORT` is its default; `config.DefaultAPIPort = 7443`; range 1-65535 checked in `Config.Validate()`.
  - `entrypoint.sh` passes `--api-port` only when `API_PORT` is non-empty; `Dockerfile` has `API_PORT=""` in `ENV`.
  - Unit tests: `ipquorum-download-go/pkg/config/config_test.go`.
  - The downloader still uses HTTPS only. The mock must serve TLS.
  - Runner `build.sh` default downloader is now `v3.1.1`, which does not have `--api-port`. Sub-task 8 must build the runner with the local downloader binary.

### 2. Emulator skeleton and multi-cluster config model

- **Intent:** Define how several clusters, their nodes and their scenarios are described, before writing any server code.
- **Expected Outcomes:**
  - `ipquorum-emulator/` exists with its own `go.mod`.
  - A documented config schema (YAML) with a list of `clusters`. Each cluster has `name`, `rest_listen` (host and port), `username`, `password`, `partner_system` (optional), `nodes` (list of `host:port`) and `scenario`.
  - Example configs in `scenarios/` for: single healthy cluster, two clusters with partner, node down, auth failure, rate limiting (429), slow download.
  - The config loader validates the schema, including unique cluster names, no listener port collisions, and a `partner_system` that references a known cluster (strict).
- **Todo List:**
  1. Create the directory layout and `go.mod`.
  2. Define the config structs and the loader with validation.
  3. Define the scenario enum and parameters (for example `healthy`, `auth-401`, `auth-403`, `rate-limit-429`, `mkquorumapp-error`, `download-empty`, `download-slow`, `node-down`, `handshake-timeout`, `node-drop-after`).
  4. Write the example scenario files.
  5. Unit-test the loader and validation.
- **Relevant Context:** Follow the Go layout used in `ipquorum-download-go` (`pkg/`, `internal/`). Cluster examples can use the 4-node setup (10.33.7.57 to .60, port 1260) from the real log you shared.
- **Status:** [x] done

### 3. mock-flashsystem REST API (multi-cluster)

- **Intent:** Emulate the three REST calls the downloader makes, for any number of clusters.
- **Expected Outcomes:**
  - `POST /rest/v1/auth` checks the cluster credentials (`X-Auth-Username` / `X-Auth-Password`) and returns an `X-Auth-Token`. `GET /rest/v1/auth` returns 401/405 so the pre-flight check passes.
  - `POST /rest/v1/mkquorumapp` requires a valid token, validates the payload (`ip_6`, `nometadata`, `partnersystem`, `partnerip6`) and records per-cluster state "quorum app created". An unknown `partnersystem` is rejected with an error (strict).
  - `POST /rest/v1/download` requires a valid token and the created quorum app, and returns the per-cluster JAR (see sub-task 4).
  - Tokens are scoped to a cluster, so a token from cluster A is rejected by cluster B.
  - Clusters are reachable in two ways: one process with one TLS listener per cluster on different ports, or one process per cluster (one cluster in the config) on 7443 for compose.
  - TLS uses a generated self-signed certificate, matching the downloader's default `--insecure` behavior.
  - The scenario setting changes the responses (401, 403, 429 with `Retry-After`, 5xx, empty body, slow body).
- **Todo List:**
  1. HTTP servers and a per-cluster router.
  2. Self-signed certificate generation at startup (or a mounted cert).
  3. Auth, token store and per-cluster state.
  4. `mkquorumapp` and `download` handlers with scenario hooks.
  5. Request logging that identifies the cluster.
  6. Handler unit tests using `httptest`.
- **Relevant Context:**
  - [`ipquorum-download-go/pkg/client/client.go`](ipquorum-download-go/pkg/client/client.go): auth headers, token extraction, status code handling
  - [`ipquorum-download-go/pkg/client/quorum.go`](ipquorum-download-go/pkg/client/quorum.go): payload and accepted status codes (200 and 201)
  - [`ipquorum-download-go/pkg/client/download.go`](ipquorum-download-go/pkg/client/download.go): payload (`prefix=/dumps`, `filename=ip_quorum.jar`)
- **Status:** [x] done

### 4. Fake ip_quorum.jar and per-cluster JAR generation

- **Intent:** Provide a JAR that passes the runner's checks (non-empty, valid ZIP, runnable with `java -jar`) and behaves like the real one in logs and lifecycle.
- **Expected Outcomes:**
  - The fake JAR reads its node list from an embedded config entry (for example `quorum.properties`), prints the real log format, and tries to connect to every node.
  - It logs `Connected to <ip>` on success and a failure line when a node is unreachable, retrying as the scenario requires.
  - It stays running until SIGTERM, then exits cleanly.
  - mock-flashsystem builds a per-cluster JAR at download time from the template, using Go `archive/zip` to add the cluster's node list. Two clusters therefore produce two different JARs.
  - The template JAR is built by a multi-stage Containerfile (JDK build stage) and is not committed.
- **Todo List:**
  1. Write the Java source: config parsing, log output, socket connections, signal handling.
  2. Write the Maven or plain `javac` build and the multi-stage Containerfile that outputs the template JAR into the mock image.
  3. Implement the JAR rewrite in mock-flashsystem.
  4. Verify the result with `unzip -t` and `java -jar` inside the runner image.
- **Relevant Context:**
  - [`ipquorum-management-platform/ipquorum-runner/entrypoint.sh`](ipquorum-management-platform/ipquorum-runner/entrypoint.sh): size check, `unzip -t`, `java -jar`, signal trap
  - [`ipquorum-management-platform/ipquorum-runner/Dockerfile`](ipquorum-management-platform/ipquorum-runner/Dockerfile): JRE only (`eclipse-temurin:17-jre-alpine`), so the Java target version must be 17 or lower
  - Log format reference: the `podman logs specv-ipquorum` output in the original request
- **Status:** [x] done

### 5. Mock quorum nodes (TCP) and simplified protocol

- **Intent:** Let the fake JAR make real TCP connections so connection and failure behavior can be tested.
- **Expected Outcomes:**
  - mock-flashsystem opens a TCP listener for each node of each cluster (`host:port` from the config).
  - A minimal handshake is implemented in both the mock node and the fake JAR (for example a short request and response message). It is documented as a simplified protocol of our own, not the real one.
  - Node behaviors are driven by the scenario: accept normally, refuse (node down), accept but never answer (handshake timeout), and drop the connection after N seconds.
  - Node state is per cluster.
- **Todo List:**
  1. Define the handshake message format and document it in the README.
  2. Implement the node listeners in Go.
  3. Implement the client side in the Java fake JAR.
  4. Implement the node behaviors and add Go tests for them.
- **Relevant Context:** Within one container, nodes need distinct ports, so a node entry is `host:port`. In compose, each node can also be its own hostname on the default 1260.
- **Status:** [x] done

### 6. Runtime control API (optional per-cluster failure injection)

- **Intent:** Let automated tests change a cluster's scenario while it runs, for example take a node down mid-test.
- **Expected Outcomes:**
  - A small control endpoint (on a separate port) can list clusters, show their state, set the scenario, and start or stop individual nodes.
  - Tests can change the behavior of one cluster without affecting the others.
- **Todo List:**
  1. Add the control HTTP API (JSON, no auth, documented as test-only).
  2. Wire it to the scenario and node controllers.
  3. Add unit tests.
- **Relevant Context:** Scenario and node controllers from sub-tasks 3 and 5.
- **Status:** [x] done

### 7. Go unit tests for the downloader against the mock

- **Intent:** Close the `pkg/client` and `pkg/config` coverage gaps listed in `TESTING.md` and run the downloader against emulated clusters.
- **Expected Outcomes:**
  - `pkg/client` tests cover auth success, 401/403 (fail fast), 429 retry with `Retry-After`, token in header vs JSON body, `mkquorumapp` success and failure, download success and empty body, and pre-flight status handling.
  - Tests use `httptest.NewTLSServer`, with a short `BaseDelay` so they stay quick.
  - At least one test runs the downloader against two clusters in sequence and checks that the JARs differ.
- **Todo List:**
  1. Add shared test helpers (config builder, fake server builders).
  2. Write `client_test.go`, `quorum_test.go` and `download_test.go`.
  3. Update `TESTING.md` to remove the completed items from "Future Test Additions".
- **Relevant Context:**
  - [`ipquorum-download-go/TESTING.md`](ipquorum-download-go/TESTING.md) (existing conventions: table-driven tests, `make test`)
  - [`ipquorum-download-go/Makefile`](ipquorum-download-go/Makefile)
- **Status:** [x] done

### 8. Compose test environment and runner integration tests

- **Intent:** Test the complete chain (runner entrypoint, downloader, fake JAR, mock nodes) with several clusters, with no real hardware.
- **Expected Outcomes:**
  - `compose/podman-compose.test.yml` starts at least two mock clusters (for example `flash-a`, `flash-b`, each on 7443) and two runner containers, one per cluster. `API_ENDPOINT` is set to the mock hostname.
  - Integration scripts check, using `podman logs`, that each runner downloads its JAR, starts, and reports `Connected to` for its own nodes only.
  - Scripts cover the failure scenarios (auth failure makes the runner exit non-zero, node down is logged, JAR corruption is detected).
  - Scripts work with both podman and docker.
- **Todo List:**
  1. Write the compose file, including the mock image Containerfile.
  2. Write the integration test scripts with clear pass or fail exit codes.
  3. Add a `make` or script entry point that builds the images, runs the tests and cleans up.
- **Relevant Context:**
  - [`ipquorum-management-platform/ipquorum-runner/build.sh`](ipquorum-management-platform/ipquorum-runner/build.sh): the runtime detection pattern to reuse
  - [`ipquorum-management-platform/podman-compose.yml`](ipquorum-management-platform/podman-compose.yml): compose conventions
  - Note: the runner Dockerfile downloads the downloader binary from a GitHub release. The test build needs a build arg or alternative to use the local downloader build with the new `--api-port` flag.
- **Status:** [x] done

### 9. CI workflow

- **Intent:** Run the emulator and integration tests automatically on every change.
- **Expected Outcomes:**
  - A new workflow under `.github/workflows/` runs the Go tests (downloader and emulator), builds the emulator and runner images, and runs the compose integration tests.
  - It follows the existing workflow conventions.
- **Todo List:**
  1. Review the existing workflows for conventions.
  2. Add `emulator-tests.yml`.
  3. Verify the trigger paths cover `ipquorum-download-go/`, `ipquorum-emulator/` and `ipquorum-management-platform/ipquorum-runner/`.
- **Relevant Context:**
  - [`.github/workflows/container-build.yml`](.github/workflows/container-build.yml)
  - [`.github/workflows/release-go-binary.yml`](.github/workflows/release-go-binary.yml)
- **Status:** [x] done

### 10. Documentation

- **Intent:** Make the emulator usable by other developers.
- **Expected Outcomes:**
  - `ipquorum-emulator/README.md` covers usage, the config schema, scenarios, the multi-cluster setup, the simplified protocol, and the limits of the emulation (this is not the real quorum protocol).
  - The top-level `readme.md` and `TESTING.md` link to it.
  - `CHANGELOG.md` has an entry for the `--api-port` change.
- **Todo List:**
  1. Write the emulator README with a quick start for a two-cluster setup.
  2. Add links and the changelog entry.
- **Relevant Context:** [`readme.md`](readme.md), [`CHANGELOG.md`](CHANGELOG.md), [`ipquorum-download-go/TESTING.md`](ipquorum-download-go/TESTING.md)
- **Status:** [x] done

---

## Open points to confirm

- None. All open points are resolved.

## Status

All 10 sub-tasks are implemented (user asked to run through all of them in one go).

### Implementation notes / deviations from the plan
- Config gained a `quorum_app_exists` field and an `external_systems` list (partner systems that are valid but not mocked).
- The REST port is per cluster via `rest_listen`; mock nodes use `address` plus `listen`.
- The runner integration tests need an overlay (`compose/Containerfile.runner-test`) because the released downloader lacks `--api-port`. It uses `FROM --platform=linux/amd64` to work around a podman local-image bug on arm64 hosts.
- The emulator is also part of the Go matrix in `security-and-tests.yml`. `emulator-tests.yml` deliberately omits buildx.
- The wire protocol on port 1260 is our own simplified protocol; the real one is not public.
- CI workflows were validated locally only (YAML parse), not run on GitHub.
