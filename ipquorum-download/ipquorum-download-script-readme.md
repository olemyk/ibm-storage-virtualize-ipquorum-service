
# IPQuorum Download Script

A bash utility to **download the IBM Virtualize IPQuorum JAR** via REST API and optionally **create a new Quorum App** (mkquorumapp).

> Script name: `ipquorum-restapi-download.sh`

---

## Table of Contents
- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Options](#options)
  - [General](#general)
  - [mkquorumapp Payload](#mkquorumapp-payload)
- [Examples](#examples)
- [Security & Best Practices](#security--best-practices)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [License](#license)

---

## Overview
This script interacts with the IBM Storage Virtualize REST API to:
1. **Download** the IPQuorum JAR artifact to a local file.
2. Optionally **create a fresh Quorum App** in the target system through the `mkquorumapp` API call.

Both actions are configurable via command-line flags.

---

## Prerequisites
- A Linux/Unix environment with **bash**.
- Curl and JQ installed
- Network access to the IBM Storage Virtualize API endpoint on port 7443
- Remote System in PBHA (MANDATORY if option mkquorumapp )
- Valid credentials:
  - **Monitor** role for IPQuorum JAR download.
  - **Minimum Restricted Administrator** for creating a Quorum App.
- Ensure the script is executable:
  ```bash
  chmod +x ipquorum-restapi-download.sh
  ```

> **Note:** Use `--secure` in production to enforce strict TLS verification if you have trusted certificates

---

## Quick Start
Download the IPQuorum JAR only:
```bash
./ipquorum-restapi-download.sh \
  --no-mkquorumapp --download \
  --output ip_quorum.jar \
  --user <username> --pass <password>
```

Create a Quorum App and download the IPQuorum JAR in one run:
```bash
./ipquorum-restapi-download.sh \
  --mkquorumapp --partnersystem <remote_system_name> \
  --ip6=false --partnerip6=false --nometadata=false \
  --download --insecure \
  --user <username> --pass <password>
```

---

## Usage
```bash
ipquorum-restapi-download.sh [options]
```

You can combine **general** options with **mkquorumapp payload** options. When `--mkquorumapp` is enabled, `--partnersystem` becomes **mandatory**.

---

## Options

### General
```
--mkquorumapp / --no-mkquorumapp     Enable/disable mkquorumapp call (default: enabled)
--download / --no-download           Enable/disable jar download (default: enabled)
--insecure / --secure                Use insecure TLS (-k) or strict TLS (default: insecure)
--api-endpoint <host>                Set API endpoint IP of the IBM Storage Virtualize
--user <username>                    Auth username - Monitor Role for download or Minimum Restricted Administrator to create new Quorum App.
--pass <password>                    Auth password (REQUIRED)
--output <file>                      Output jar filename (default: ip_quorum.jar)
```

### mkquorumapp Payload - To Create a fresh IPQuorum- 
```
--ip6[=true|false] or --ip_6[=true|false]   Set IPv6 flag (default: false)
--nometadata[=true|false]                   Set nometadata flag (default: false)
--partnersystem <name>                      Set Partnersystem - Remote System name in PBHA (MANDATORY if mkquorumapp is enabled). this is the name of the system
--partnerip6[=true|false]                   Set partner IPv6 flag (default: false)
```

---

## Examples
Create Quorum App + Download JAR:
```bash
./ipquorum-restapi-download.sh \
  --mkquorumapp --partnersystem svc_cluster02 \
  --ip6=false --partnerip6=false --nometadata=false \
  --download --insecure \
  --user superuser --pass password
```

Download JAR only:
```bash
./ipquorum-restapi-download.sh \
  --no-mkquorumapp --download \
  --output ip_quorum.jar \
  --user superuser --pass password
```

---

## Security & Best Practices
- **Protect credentials**: Avoid hardcoding passwords. Prefer environment variables or a secrets manager.
- **TLS settings**: Use `--secure` in production. Only use `--insecure` for testing.
- **Least privilege**: Use the minimal role required (Monitor for download; Restricted Admin for mkquorumapp).
- **Audit & logs**: If the script outputs logs, store them securely for auditing.

---

## Troubleshooting
- **Permission denied**: Make the script executable: `chmod +x ipquorum-restapi-download.sh`.
- **Authentication errors**: Verify username/password and that your role grants access to the chosen operations.
- **Connection issues**: Check reachability to the API endpoint (firewall, routing, DNS). Example quick check:
  ```bash
    curl -vk https://<api-endpoint>:7442/rest/v1/auth
  ```
- **TLS failures**: If strict TLS fails, verify the server certificate chain or temporarily use `--insecure` for testing.

---

## FAQ
**Q: Do I need admin rights to download the JAR?**  
A: No. The **Monitor** role is sufficient.

**Q: Is `--partnersystem` required?**  
A: Yes, when `--mkquorumapp` is enabled, `--partnersystem <name>` is **mandatory**.

**Q: Can I change the output filename?**  
A: Yes, use `--output <file>`. The default is `ip_quorum.jar`.

**Q: Does the script support IPv6?**  
A: Yes. Use `--ip6=true` and/or `--partnerip6=true` when needed.

---
## Documentation

- [IPQuorum Info](https://www.ibm.com/support/pages/ibm-storage-virtualize-ip-quorum-application-requirements-1)
- [IP quorum application](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)
- [Storage Virtualize RESTful API](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=interface-storage-virtualize-restful-api)
---

## Maintainers
- Ole Kristian Myklebust

