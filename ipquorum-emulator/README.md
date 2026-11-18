# IP Quorum Emulator

Develop and test the IP quorum tooling **without an IBM FlashSystem and without a real
`ip_quorum.jar`**. The emulator provides:

| Part | What it does |
|---|---|
| **mock-flashsystem** (Go) | Emulates the Storage Virtualize REST endpoints used by `ipquorum-download-go`: `auth`, `mkquorumapp`, `download`. Hosts **one or several** clusters. |
| **fake `ip_quorum.jar`** (Java) | A stand-in JAR that prints the log format of the real one, connects to the cluster's nodes and stays running until SIGTERM. Built in CI, never committed. |
| **mock quorum nodes** (Go, TCP) | Listeners the fake JAR connects to. Node down, handshake timeout and connection drop can be simulated. |
| **control API** (HTTP, test only) | Changes a cluster's scenario or a node's behavior while everything is running. |

> **Limits of the emulation.** The real IP quorum wire protocol is not publicly documented and
> is **not** emulated. The fake JAR and the mock nodes speak a small line-based protocol of our
> own (below). Use the emulator to test the downloader, the runner container, the management
> platform and failure handling. It cannot tell you whether a real FlashSystem accepts a quorum.

## Quick start

```bash
make test           # Go unit tests (-race)
make integration    # full container test: needs podman or docker
```

Run by hand, two systems in one process, with the real downloader:

```bash
make jar build
./bin/mock-flashsystem -config scenarios/two-clusters.yaml \
    -jar-template fake-ipquorum-jar/build/fake-ip_quorum.jar &

# flash-a (REST on 8443), partner flash-b
ipquorum-download-go --api-endpoint 127.0.0.1 --api-port 8443 --user superuser \
    --pass passw0rd-a --partnersystem flash-b --output /tmp/a.jar
# flash-b (REST on 8444)
ipquorum-download-go --api-endpoint 127.0.0.1 --api-port 8444 --user superuser \
    --pass passw0rd-b --partnersystem flash-a --output /tmp/b.jar

unzip -p /tmp/a.jar quorum.properties   # each JAR lists only its own cluster's nodes
java -jar /tmp/a.jar
```

`--api-port` / `API_PORT` is an optional downloader and runner setting (default 7443). Real
systems always use 7443, so it only matters for the emulator.

## Multiple FlashSystems

Everything is per cluster, so you can emulate a HyperSwap pair, several independent systems,
or a healthy system next to a broken one.

- One mock process can host many clusters (`scenarios/two-clusters.yaml`), each on its own REST
  port. Or run one process per cluster, each on 7443 (`compose/`).
- **Tokens, credentials and state are per cluster.** A token from `flash-a` is rejected by
  `flash-b`. `download` returns 404 until `mkquorumapp` has run on that cluster.
- **Each cluster serves its own JAR** with its own node list embedded in `quorum.properties`.
- **`partnersystem` is strict**, like a real system. `mkquorumapp` is rejected (400) when the
  partner does not exist. If a cluster sets `partner_system`, only that one is accepted.
  Partners must be another cluster in the same config or listed in `external_systems`.

## Configuration

```yaml
control_listen: "127.0.0.1:9000"      # test-only control API; empty disables it
external_systems: [remote-cluster]    # partner names that exist outside this config

clusters:
  - name: flash-a                     # required, unique
    username: superuser               # default superuser
    password: passw0rd                # required
    rest_listen: "127.0.0.1:8443"     # HTTPS REST API, default ":7443"
    partner_system: flash-b           # optional, strict
    quorum_app_exists: false          # true: download works without mkquorumapp first
    scenario: healthy                 # see below, default healthy
    rate_limit_count: 2               # rate-limit-429: first N auth attempts get 429
    download_delay: 3s                # download-slow
    tls_hosts: [flash-a.example.com]  # extra names for the generated self-signed cert
    nodes:
      - address: "10.33.7.57:1260"    # host:port embedded in the generated JAR (host required)
        listen: ":1260"               # where the mock binds, default ":<port of address>"
        behavior: up                  # up | down | handshake-timeout | drop-after
        drop_after: 10s               # drop-after
```

The loader rejects unknown fields, duplicate cluster or node names, unknown scenarios and
behaviors, unknown partners and **listener port collisions** (including the control API).
Ready-made files are in [`scenarios/`](scenarios/).

### REST scenarios

| Scenario | Effect |
|---|---|
| `healthy` | Normal behavior |
| `auth-401` / `auth-403` | Every auth attempt fails with 401 / 403 |
| `rate-limit-429` | First `rate_limit_count` auth attempts get 429 with `Retry-After` |
| `mkquorumapp-error` | `mkquorumapp` returns 500 |
| `download-error` | `download` returns 500 |
| `download-empty` | `download` returns 200 with an empty body |
| `download-slow` | Body is delayed by `download_delay` |

Wrong credentials always return 403, as the downloader expects.

### Node behaviors

| Behavior | Effect |
|---|---|
| `up` | Answers the handshake and keepalives |
| `down` | Not listening: connection refused |
| `handshake-timeout` | Accepts the connection but never answers |
| `drop-after` | Works, then closes the connection after `drop_after` |

### Simplified node protocol (ours, not IBM's)

```
client -> node   HELLO protocol=3 features_bitmap=0
node   -> client HELLO-ACK protocol=3 node=<address>
client -> node   PING            (repeated every fake.keepalive.ms)
node   -> client PONG
```

The fake JAR tunes timing through system properties, which the runner passes via `JAVA_OPTS`:
`-Dfake.keepalive.ms` (2000), `-Dfake.retry.ms` (2000), `-Dfake.connect.ms` (3000),
`-Dfake.handshake.ms` (5000).

## Control API (test only, no authentication)

Never expose this outside a test setup.

```bash
curl localhost:9000/clusters                     # all clusters: state and nodes
curl localhost:9000/clusters/flash-a
curl -X POST localhost:9000/clusters/flash-a/scenario -d '{"scenario":"auth-401"}'
curl -X POST localhost:9000/clusters/flash-a/nodes/10.33.7.57:1260/behavior \
     -d '{"behavior":"drop-after","drop_after":"5s"}'
curl -X POST localhost:9000/clusters/flash-a/reset   # back to the configured state
curl localhost:9000/healthz
```

Changing one cluster or node never affects another. Changing a node's behavior closes its
open connections so a connected client sees the change immediately.

## Container integration tests

`make integration` (or `tests/run-integration.sh`) builds three images, starts
[`compose/podman-compose.test.yml`](compose/podman-compose.test.yml) (3 emulated systems and 3
runners) and checks, among other things:

- each runner downloads a JAR that lists **only its own** system's nodes
- a custom REST port (`API_PORT=8443`) and the default (7443) both work
- a node failure on `flash-a` is seen by `runner-a` only, and `runner-a` reconnects when it returns
- auth 401/403, wrong password, `mkquorumapp` error, unknown partner, download error and empty
  download all make the runner fail clearly
- 429 responses are retried, then the runner connects
- a runner exits 0 on SIGTERM

Options: `--no-build` reuses images, `--keep` leaves the stack running. Environment: `RUNTIME`
(`podman`/`docker`), `STARTUP_TIMEOUT`. The runner image is `linux/amd64`, so on Apple Silicon
it runs emulated and the first start is slower.

The runner image used in tests is the production
[`Dockerfile`](../ipquorum-management-platform/ipquorum-runner/Dockerfile), with only the
downloader and `entrypoint.sh` replaced by the ones from your working tree
([`compose/Containerfile.runner-test`](compose/Containerfile.runner-test)). Your changes are
tested, not a released binary.

### Use it with the management platform

Point an instance at the emulator, with no platform code changes: `API_ENDPOINT` = the mock's
host, `API_PORT` = its REST port (leave empty for 7443), and the mock's username and password.

## Layout

```
cmd/mock-flashsystem/   main
internal/config/        YAML model and validation
internal/cluster/       per-cluster state and registry
internal/rest/          REST API and self-signed TLS
internal/jar/           per-cluster JAR built from the template
internal/nodes/         mock quorum nodes
internal/control/       control API
fake-ipquorum-jar/      Java source of the fake ip_quorum.jar (Java 17 target)
scenarios/              example configs
compose/                container test setup
tests/                  integration test script
```

## CI

- `security-and-tests.yml`: the emulator is part of the Go module matrix (build, vet, `-race`
  tests, `govulncheck`).
- `emulator-tests.yml`: the container integration tests.
