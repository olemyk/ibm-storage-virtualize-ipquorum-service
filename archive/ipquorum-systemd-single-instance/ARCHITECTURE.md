# IP Quorum Service Architecture & Workflow

This document provides visual representations of the IP Quorum systemd service architecture and workflows.

## 📐 System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Linux System (systemd)                       │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │         ibm-virtualize-ipquorum.service (systemd unit)       │   │
│  │                                                               │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │  ExecStartPre: ipquorum-download.sh                 │    │   │
│  │  │  (Pre-start download script)                        │    │   │
│  │  │                                                      │    │   │
│  │  │  ┌──────────────────────────────────────────────┐  │    │   │
│  │  │  │  Reads: /etc/ipquorum/ipquorum.conf          │  │    │   │
│  │  │  │  - API_ENDPOINT                              │  │    │   │
│  │  │  │  - VIRTUALIZE_USERNAME                       │  │    │   │
│  │  │  │  - IPQUORUM_DOWNLOAD_TOOL (go/python/bash)   │  │    │   │
│  │  │  │  - IPQUORUM_MKQUORUMAPP_ENABLED              │  │    │   │
│  │  │  │  - IPQUORUM_TLS_VERIFY                       │  │    │   │
│  │  │  └──────────────────────────────────────────────┘  │    │   │
│  │  │                        │                            │    │   │
│  │  │                        ▼                            │    │   │
│  │  │  ┌──────────────────────────────────────────────┐  │    │   │
│  │  │  │  Reads: /etc/ipquorum/.password              │  │    │   │
│  │  │  │  (Secure password file)                      │  │    │   │
│  │  │  └──────────────────────────────────────────────┘  │    │   │
│  │  │                        │                            │    │   │
│  │  │                        ▼                            │    │   │
│  │  │  ┌──────────────────────────────────────────────┐  │    │   │
│  │  │  │  Calls Download Tool:                        │  │    │   │
│  │  │  │  • ipquorum-download-go (Go binary)          │  │    │   │
│  │  │  │  • ipquorum-download.py (Python script)      │  │    │   │
│  │  │  │  • ipquorum-restapi-download.sh (Bash)       │  │    │   │
│  │  │  └──────────────────────────────────────────────┘  │    │   │
│  │  │                        │                            │    │   │
│  │  │                        ▼                            │    │   │
│  │  │  ┌──────────────────────────────────────────────┐  │    │   │
│  │  │  │  Downloads: /opt/IBM/ip-quorum/ip_quorum.jar │  │    │   │
│  │  │  │  Logs to: /opt/IBM/ip-quorum/log/download.log│  │    │   │
│  │  │  └──────────────────────────────────────────────┘  │    │   │
│  │  └─────────────────────────────────────────────────────┘    │   │
│  │                                                               │   │
│  │  ┌─────────────────────────────────────────────────────┐    │   │
│  │  │  ExecStart: ipquorum-start.sh                       │    │   │
│  │  │  (Main service start script)                        │    │   │
│  │  │                                                      │    │   │
│  │  │  ┌──────────────────────────────────────────────┐  │    │   │
│  │  │  │  Reads: /etc/ipquorum/ipquorum.conf          │  │    │   │
│  │  │  │  - IPQUORUM_NAME                             │  │    │   │
│  │  │  │  - IPQUORUM_DEBUG                            │  │    │   │
│  │  │  │  - IPQUORUM_LOG_ROTATION                     │  │    │   │
│  │  │  │  - JAVA_BIN                                  │  │    │   │
│  │  │  └──────────────────────────────────────────────┘  │    │   │
│  │  │                        │                            │    │   │
│  │  │                        ▼                            │    │   │
│  │  │  ┌──────────────────────────────────────────────┐  │    │   │
│  │  │  │  Executes: java -jar ip_quorum.jar           │  │    │   │
│  │  │  │  With options: -name, -debug, -rotation, etc │  │    │   │
│  │  │  └──────────────────────────────────────────────┘  │    │   │
│  │  └─────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Network Communication                      │   │
│  │                                                               │   │
│  │  IP Quorum App ←→ Port 1260/TCP ←→ Storage Virtualize       │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## 🔄 Service Start Workflow

```
┌─────────────────────────────────────────────────────────────────────┐
│                    systemctl start ipquorum                          │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Step 1: ExecStartPre - ipquorum-download.sh                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1.1 Read Configuration                                        │  │
│  │     • Source /etc/ipquorum/ipquorum.conf                      │  │
│  │     • Check IPQUORUM_DOWNLOAD_ENABLED                         │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1.2 Validate Configuration                                    │  │
│  │     • Check API_ENDPOINT is set                               │  │
│  │     • Check VIRTUALIZE_USERNAME is set                        │  │
│  │     • Verify password file exists and is readable             │  │
│  │     • Validate mkquorumapp settings if enabled                │  │
│  │     • Validate TLS settings                                   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1.3 Backup Existing JAR (if enabled)                          │  │
│  │     • cp ip_quorum.jar → ip_quorum.jar.backup                 │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1.4 Call Download Tool                                        │  │
│  │     Based on IPQUORUM_DOWNLOAD_TOOL:                          │  │
│  │                                                                │  │
│  │     ┌────────────────────────────────────────────────────┐   │  │
│  │     │ Go Binary (ipquorum-download-go)                   │   │  │
│  │     │ • Fast startup (<10ms)                             │   │  │
│  │     │ • Low memory (<20MB)                               │   │  │
│  │     │ • Single binary, no dependencies                   │   │  │
│  │     └────────────────────────────────────────────────────┘   │  │
│  │                          OR                                    │  │
│  │     ┌────────────────────────────────────────────────────┐   │  │
│  │     │ Python Script (ipquorum-download.py)               │   │  │
│  │     │ • Requires Python 3.8+                             │   │  │
│  │     │ • Feature-rich with detailed logging               │   │  │
│  │     └────────────────────────────────────────────────────┘   │  │
│  │                          OR                                    │  │
│  │     ┌────────────────────────────────────────────────────┐   │  │
│  │     │ Bash Script (ipquorum-restapi-download.sh)         │   │  │
│  │     │ • Requires curl and jq                             │   │  │
│  │     │ • Most portable                                    │   │  │
│  │     └────────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1.5 Verify Downloaded File                                    │  │
│  │     • Check file exists                                       │  │
│  │     • Check file size > 1000 bytes                            │  │
│  │     • Set proper ownership (ipquorum:ipquorum)                │  │
│  │     • Set permissions (644)                                   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1.6 Success or Restore Backup                                 │  │
│  │     • If success: Continue to Step 2                          │  │
│  │     • If failure: Restore backup and exit with error          │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Step 2: ExecStart - ipquorum-start.sh                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 2.1 Read Configuration                                        │  │
│  │     • Source /etc/ipquorum/ipquorum.conf                      │  │
│  │     • Load IP Quorum application options                      │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 2.2 Build Java Command                                        │  │
│  │     • Base: java -jar /opt/IBM/ip-quorum/ip_quorum.jar       │  │
│  │     • Add -name if IPQUORUM_NAME is set                       │  │
│  │     • Add -debug if IPQUORUM_DEBUG=true                       │  │
│  │     • Add -emit if IPQUORUM_EMIT=true                         │  │
│  │     • Add -location if IPQUORUM_LOG_LOCATION is set           │  │
│  │     • Add -rotation if IPQUORUM_LOG_ROTATION is set           │  │
│  │     • Add -size if IPQUORUM_LOG_SIZE is set                   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 2.3 Execute IP Quorum Application                             │  │
│  │     • Run as user: ipquorum                                   │  │
│  │     • Listen on port: 1260/TCP                                │  │
│  │     • Log to: /opt/IBM/ip-quorum/log/                         │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Service Running                                   │
│  • IP Quorum app listening on port 1260                             │
│  • Communicating with Storage Virtualize cluster                    │
│  • Logging to /opt/IBM/ip-quorum/log/                               │
└─────────────────────────────────────────────────────────────────────┘
```

## 🔽 Download Tool Workflow (Go Binary)

```
┌─────────────────────────────────────────────────────────────────────┐
│              ipquorum-download-go Execution Flow                     │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 1: Initialization                                             │
├─────────────────────────────────────────────────────────────────────┤
│  • Parse command-line flags                                          │
│  • Read password from file                                           │
│  • Validate password file permissions (warn if world-readable)       │
│  • Initialize HTTP client with retry logic                           │
│  • Set TLS verification mode (secure/insecure)                       │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 2: Pre-flight Check                                           │
├─────────────────────────────────────────────────────────────────────┤
│  • Test connectivity to API endpoint                                 │
│  • GET https://{API_ENDPOINT}:7443/rest/v1/auth                     │
│  • Verify endpoint is reachable (200/401/404/405 acceptable)        │
│  • Exit if endpoint unreachable                                      │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 3: Authentication                                             │
├─────────────────────────────────────────────────────────────────────┤
│  • POST https://{API_ENDPOINT}:7443/rest/v1/auth                    │
│  • Headers:                                                          │
│    - X-Auth-Username: {username}                                     │
│    - X-Auth-Password: {password}                                     │
│  • Retry logic with exponential backoff (max 8 attempts)            │
│  • Handle rate limiting (429 responses)                              │
│  • Extract X-Auth-Token from response headers                        │
│  • Exit if authentication fails after all retries                    │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 4: mkquorumapp (Optional)                                     │
├─────────────────────────────────────────────────────────────────────┤
│  IF --mkquorumapp flag is set:                                       │
│                                                                       │
│  • POST https://{API_ENDPOINT}:7443/rest/v1/mkquorumapp             │
│  • Headers:                                                          │
│    - X-Auth-Token: {token}                                           │
│    - Content-Type: application/json                                  │
│  • Body:                                                             │
│    {                                                                 │
│      "ip_6": false,                                                  │
│      "nometadata": false,                                            │
│      "partnersystem": "remote_cluster",                              │
│      "partnerip6": false                                             │
│    }                                                                 │
│  • Creates new IP Quorum application on Storage Virtualize           │
│  • Requires Restricted Administrator role or higher                  │
│                                                                       │
│  ELSE:                                                               │
│  • Skip mkquorumapp call                                             │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 5: Download JAR                                               │
├─────────────────────────────────────────────────────────────────────┤
│  IF --download flag is set:                                          │
│                                                                       │
│  • POST https://{API_ENDPOINT}:7443/rest/v1/download                │
│  • Headers:                                                          │
│    - X-Auth-Token: {token}                                           │
│    - Content-Type: application/json                                  │
│  • Body:                                                             │
│    {                                                                 │
│      "prefix": "/dumps",                                             │
│      "filename": "ip_quorum.jar"                                     │
│    }                                                                 │
│  • Stream response to output file                                    │
│  • Verify file size > 0                                              │
│  • Set file permissions                                              │
│                                                                       │
│  ELSE:                                                               │
│  • Skip download                                                     │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Phase 6: Completion                                                 │
├─────────────────────────────────────────────────────────────────────┤
│  • Log success message                                               │
│  • Exit with code 0                                                  │
│                                                                       │
│  On any error:                                                       │
│  • Log error message                                                 │
│  • Exit with non-zero code                                           │
└─────────────────────────────────────────────────────────────────────┘
```

## 🔐 Security Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Security Layers                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  Layer 1: File System Security                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ /etc/ipquorum/.password                                      │   │
│  │ • Permissions: 440 (root:ipquorum) or 400 (ipquorum:ipquorum)│   │
│  │ • Not world-readable                                         │   │
│  │ • Validated by download tools                                │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  Layer 2: Systemd Security Directives                                │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ • ProtectSystem=strict (read-only /usr, /boot, /efi)        │   │
│  │ • ProtectHome=yes (no access to /home, /root, /run/user)    │   │
│  │ • NoNewPrivileges=true (cannot gain new privileges)          │   │
│  │ • PrivateTmp=yes (isolated /tmp directory)                   │   │
│  │ • ReadWritePaths=/opt/IBM/ip-quorum (limited write access)   │   │
│  │ • ReadOnlyPaths=/etc/ipquorum (protected config)             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  Layer 3: Network Security                                            │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ • TLS/HTTPS for API communication                            │   │
│  │ • Configurable certificate verification                      │   │
│  │ • Firewall rules for port 1260/TCP                           │   │
│  │ • No password in command line or logs                        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  Layer 4: User Isolation                                              │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ • Dedicated user: ipquorum                                   │   │
│  │ • Dedicated group: ipquorum                                  │   │
│  │ • No shell access (nologin)                                  │   │
│  │ • Limited file system access                                 │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## 📊 Component Interaction Diagram

```
┌──────────────────┐
│   Administrator  │
│   (root user)    │
└────────┬─────────┘
         │
         │ systemctl start/restart
         │
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                          systemd                                 │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  ibm-virtualize-ipquorum.service                           │ │
│  │                                                             │ │
│  │  EnvironmentFile=/etc/ipquorum/ipquorum.conf               │ │
│  │  User=ipquorum                                             │ │
│  │  Group=ipquorum                                            │ │
│  └────────────────────────────────────────────────────────────┘ │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ ExecStartPre
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  ipquorum-download.sh                                            │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  • Reads config from /etc/ipquorum/ipquorum.conf           │ │
│  │  • Reads password from /etc/ipquorum/.password             │ │
│  │  • Calls download tool                                     │ │
│  └────────────────────────────────────────────────────────────┘ │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ Executes
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  ipquorum-download-go (or python/bash)                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  1. Pre-flight check                                       │ │
│  │  2. Authenticate                                           │ │
│  │  3. Create quorum app (optional)                           │ │
│  │  4. Download JAR                                           │ │
│  └────────────────────────────────────────────────────────────┘ │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ HTTPS/TLS
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  IBM Storage Virtualize Cluster                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  REST API (Port 7443)                                      │ │
│  │  • /rest/v1/auth (authentication)                          │ │
│  │  • /rest/v1/mkquorumapp (create quorum app)                │ │
│  │  • /rest/v1/download (download JAR)                        │ │
│  └────────────────────────────────────────────────────────────┘ │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ Returns ip_quorum.jar
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  /opt/IBM/ip-quorum/ip_quorum.jar                                │
│  • Saved to disk                                                 │
│  • Permissions: 644 ipquorum:ipquorum                            │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ Download complete
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  systemd continues with ExecStart                                │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ ExecStart
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  ipquorum-start.sh                                               │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  • Reads config from /etc/ipquorum/ipquorum.conf           │ │
│  │  • Builds Java command with options                        │ │
│  │  • Executes: java -jar ip_quorum.jar [options]             │ │
│  └────────────────────────────────────────────────────────────┘ │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ Executes
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  IP Quorum Application (Java)                                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  • Listens on port 1260/TCP                                │ │
│  │  • Logs to /opt/IBM/ip-quorum/log/                         │ │
│  │  • Communicates with Storage Virtualize cluster            │ │
│  └────────────────────────────────────────────────────────────┘ │
└───────────┬─────────────────────────────────────────────────────┘
            │
            │ TCP/IP (Port 1260)
            ▼
┌─────────────────────────────────────────────────────────────────┐
│  IBM Storage Virtualize Cluster                                  │
│  • Uses IP Quorum for tie-breaking in split-brain scenarios      │
│  • Monitors quorum application health                            │
└─────────────────────────────────────────────────────────────────┘
```

## 🔄 Service Restart Workflow

```
systemctl restart ipquorum
         │
         ▼
┌─────────────────────┐
│  Stop Current       │
│  Service            │
│  • Graceful stop    │
│  • Wait for exit    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Download Latest    │
│  JAR (if enabled)   │
│  • Backup old JAR   │
│  • Download new JAR │
│  • Verify download  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Start Service      │
│  • Load new JAR     │
│  • Apply config     │
│  • Start listening  │
└─────────────────────┘
```

## 📁 File System Layout

```
/
├── etc/
│   └── ipquorum/
│       ├── ipquorum.conf          (644 root:root) - Main configuration
│       └── .password              (440 root:ipquorum) - Secure password
│
├── opt/
│   └── IBM/
│       └── ip-quorum/
│           ├── ip_quorum.jar      (644 ipquorum:ipquorum) - Application
│           ├── ip_quorum.jar.backup (644 ipquorum:ipquorum) - Backup
│           └── log/               (755 ipquorum:ipquorum) - Log directory
│               ├── download.log   (644 ipquorum:ipquorum) - Download logs
│               └── ipquorum*.log  (644 ipquorum:ipquorum) - App logs
│
├── usr/
│   ├── local/
│   │   └── bin/
│   │       ├── ipquorum-download-go        (755 root:root) - Go binary
│   │       ├── ipquorum-download.py        (755 root:root) - Python script
│   │       └── ipquorum-restapi-download.sh (755 root:root) - Bash script
│   │
│   └── lib/
│       └── systemd/
│           └── system/
│               └── ibm-virtualize-ipquorum.service (644 root:root)
│
└── var/
    └── lib/
        └── ipquorum/              (755 ipquorum:ipquorum) - State directory
            └── ipquorum-download.sh (755 root:root) - Download wrapper
            └── ipquorum-start.sh    (755 root:root) - Start wrapper
```

---

**Note**: All diagrams use ASCII art for maximum compatibility and can be viewed in any text editor or terminal.