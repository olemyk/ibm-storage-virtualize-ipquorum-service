# IPQuorum Download Script - Workflow Diagram

## Complete Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    START: Run Python Script                      │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 1: Parse Command-Line Arguments                           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ • --api-endpoint (required)                              │  │
│  │ • --user (username)                                      │  │
│  │ • --pass / --pass-file / --pass-prompt                   │  │
│  │ • --mkquorumapp / --no-mkquorumapp                       │  │
│  │ • --download / --no-download                             │  │
│  │ • --partnersystem (if mkquorumapp enabled)               │  │
│  │ • --ip6, --nometadata, --partnerip6                      │  │
│  │ • --insecure / --secure                                  │  │
│  │ • --output (jar filename)                                │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 2: Setup Logging                                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ • Configure log level (INFO or DEBUG)                    │  │
│  │ • Setup log format with timestamps                       │  │
│  │ • Enable password masking in logs                        │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 3: Handle Password Input                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Password provided?                                      │  │
│  │  ├─ --pass-file → Read from file                         │  │
│  │  ├─ --pass-prompt → Interactive prompt (hidden input)    │  │
│  │  ├─ --pass → Use provided password                       │  │
│  │  └─ None → Auto-prompt interactively                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 4: Validate Configuration                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ ✓ Password provided?                                     │  │
│  │ ✓ API endpoint specified?                                │  │
│  │ ✓ If mkquorumapp: partnersystem specified?               │  │
│  │ ✓ Boolean flags valid (true/false)?                      │  │
│  │                                                           │  │
│  │ ❌ Validation fails → Exit with error code 2             │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 5: Initialize IPQuorumClient                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ • Create HTTP session                                    │  │
│  │ • Configure SSL verification (secure/insecure)           │  │
│  │ • Set base URL: https://<endpoint>:7443                  │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 6: Pre-flight Check                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ GET https://<endpoint>:7443/rest/v1/auth                 │  │
│  │                                                           │  │
│  │ Response Status:                                         │  │
│  │ ✓ 200, 401, 404, 405 → Endpoint reachable               │  │
│  │ ❌ Connection error → Exit with error code 3             │  │
│  │ ❌ SSL/TLS error → Exit with error code 3                │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 7: Authentication (with Retry Logic)                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ POST https://<endpoint>:7443/rest/v1/auth                │  │
│  │ Headers:                                                 │  │
│  │   X-Auth-Username: <username>                            │  │
│  │   X-Auth-Password: <password>                            │  │
│  │                                                           │  │
│  │ Retry Logic (max 8 attempts):                            │  │
│  │ ┌────────────────────────────────────────────────────┐  │  │
│  │ │ Response Status:                                   │  │  │
│  │ │ ✓ 200/201 + Token → Success, continue             │  │  │
│  │ │ ❌ 401 → Invalid credentials, FAIL IMMEDIATELY     │  │  │
│  │ │ ❌ 403 → Insufficient permissions, FAIL IMMEDIATELY│  │  │
│  │ │ ⚠️  429 → Rate limited, wait & retry               │  │  │
│  │ │    • Check Retry-After header                      │  │  │
│  │ │    • Use exponential backoff                       │  │  │
│  │ │ ⚠️  Network error → Wait & retry                   │  │  │
│  │ └────────────────────────────────────────────────────┘  │  │
│  │                                                           │  │
│  │ Token Extraction:                                        │  │
│  │ • Try X-Auth-Token header                                │  │
│  │ • Try Authorization header                               │  │
│  │ • Try JSON body: token, access_token, authToken, etc.   │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    ┌────────┴────────┐
                    │ mkquorumapp     │
                    │ enabled?        │
                    └────┬───────┬────┘
                         │       │
                    YES  │       │  NO
                         │       │
                         ▼       └──────────────────┐
┌─────────────────────────────────────────────┐    │
│  STEP 8a: Create Quorum App (Optional)      │    │
│  ┌──────────────────────────────────────┐   │    │
│  │ POST /rest/v1/mkquorumapp            │   │    │
│  │ Headers:                             │   │    │
│  │   X-Auth-Token: <token>              │   │    │
│  │ Payload:                             │   │    │
│  │ {                                    │   │    │
│  │   "ip_6": false,                     │   │    │
│  │   "nometadata": false,               │   │    │
│  │   "partnersystem": "svc_cluster02",  │   │    │
│  │   "partnerip6": false                │   │    │
│  │ }                                    │   │    │
│  │                                      │   │    │
│  │ Response:                            │   │    │
│  │ ✓ 200/201 → Quorum app created      │   │    │
│  │ ❌ Other → Error, exit code 1        │   │    │
│  └──────────────────────────────────────┘   │    │
└────────────────────┬────────────────────────┘    │
                     │                             │
                     └──────────┬──────────────────┘
                                │
                                ▼
                       ┌────────┴────────┐
                       │ download        │
                       │ enabled?        │
                       └────┬───────┬────┘
                            │       │
                       YES  │       │  NO
                            │       │
                            ▼       └──────────────────┐
┌──────────────────────────────────────────────────┐  │
│  STEP 8b: Download JAR File (Optional)           │  │
│  ┌───────────────────────────────────────────┐   │  │
│  │ POST /rest/v1/download                    │   │  │
│  │ Headers:                                  │   │  │
│  │   X-Auth-Token: <token>                   │   │  │
│  │ Payload:                                  │   │  │
│  │ {                                         │   │  │
│  │   "prefix": "/dumps",                     │   │  │
│  │   "filename": "ip_quorum.jar"             │   │  │
│  │ }                                         │   │  │
│  │                                           │   │  │
│  │ Download Process:                         │   │  │
│  │ • Stream response to file                 │   │  │
│  │ • Save as: <output_file>                  │   │  │
│  │ • Verify file size > 0                    │   │  │
│  │                                           │   │  │
│  │ Response:                                 │   │  │
│  │ ✓ 200/201 → File downloaded successfully │   │  │
│  │ ❌ Other → Error, exit code 1             │   │  │
│  └───────────────────────────────────────────┘   │  │
└────────────────────┬─────────────────────────────┘  │
                     │                                │
                     └────────────┬───────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│  STEP 9: Success                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Log: "Operation completed successfully"                  │  │
│  │ Exit code: 0                                             │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘


═══════════════════════════════════════════════════════════════════
                         ERROR HANDLING
═══════════════════════════════════════════════════════════════════

┌─────────────────────────────────────────────────────────────────┐
│  ValidationError (Exit Code 2)                                  │
│  • Missing required parameters                                  │
│  • Invalid configuration                                        │
│  • partnersystem missing when mkquorumapp enabled               │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  NetworkError (Exit Code 3)                                     │
│  • Cannot reach endpoint                                        │
│  • SSL/TLS errors                                               │
│  • Connection timeout                                           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  AuthenticationError (Exit Code 1)                              │
│  • Invalid username/password (401)                              │
│  • Insufficient permissions (403)                               │
│  • Failed after max retries                                     │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  APIError (Exit Code 1)                                         │
│  • mkquorumapp failed                                           │
│  • Download failed                                              │
│  • Unexpected API response                                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  KeyboardInterrupt (Exit Code 130)                              │
│  • User pressed Ctrl+C                                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## Simplified Flow Diagram

```
┌──────────────┐
│ Parse Args   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Get Password │ ◄── Interactive prompt / File / Command line
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Validate     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Pre-flight   │ ◄── Check endpoint reachability
│ Check        │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Authenticate │ ◄── Get token (with retry logic)
└──────┬───────┘
       │
       ▼
┌──────────────┐     YES
│ mkquorumapp? ├─────────► Create Quorum App
└──────┬───────┘
       │ NO
       ▼
┌──────────────┐     YES
│ download?    ├─────────► Download ip_quorum.jar
└──────┬───────┘
       │ NO
       ▼
┌──────────────┐
│   Success    │
└──────────────┘
```

---

## API Endpoints Used

| Step | Method | Endpoint | Purpose |
|------|--------|----------|---------|
| Pre-flight | GET | `/rest/v1/auth` | Check endpoint reachability |
| Authentication | POST | `/rest/v1/auth` | Get authentication token |
| Create Quorum | POST | `/rest/v1/mkquorumapp` | Create new IP Quorum app |
| Download | POST | `/rest/v1/download` | Download ip_quorum.jar file |

---

## Data Flow

```
┌─────────────┐
│   User      │
└──────┬──────┘
       │ Credentials
       ▼
┌─────────────────────────────────────┐
│  Python Script                      │
│  ┌───────────────────────────────┐  │
│  │ 1. Validate inputs            │  │
│  │ 2. Get authentication token   │  │
│  │ 3. Create quorum app (opt)    │  │
│  │ 4. Download JAR file (opt)    │  │
│  └───────────────────────────────┘  │
└──────┬──────────────────────────────┘
       │ HTTPS REST API
       ▼
┌─────────────────────────────────────┐
│  IBM Storage Virtualize             │
│  (FlashSystem / SVC)                │
│  ┌───────────────────────────────┐  │
│  │ • Authenticate user           │  │
│  │ • Create quorum app in /dumps │  │
│  │ • Serve ip_quorum.jar file    │  │
│  └───────────────────────────────┘  │
└──────┬──────────────────────────────┘
       │ ip_quorum.jar
       ▼
┌─────────────┐
│ Local File  │
│ System      │
└─────────────┘
```

---

## Security Flow

```
Password Input Methods:
┌─────────────────────────────────────────────────────┐
│                                                     │
│  1. Interactive Prompt (--pass-prompt)              │
│     └─► getpass.getpass() → Hidden input           │
│                                                     │
│  2. Password File (--pass-file)                     │
│     └─► Read from file with chmod 400               │
│                                                     │
│  3. Environment Variable (VIRTUALIZE_PASSWORD)      │
│     └─► Read from environment                       │
│                                                     │
│  4. Command Line (--pass) [INSECURE]                │
│     └─► Direct input (visible in history)          │
│                                                     │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│  Password Masking in Logs                           │
│  • Debug logs show: password='********'             │
│  • Never logs plain text password                   │
└─────────────────────────────────────────────────────┘
```

---

## Retry Logic Detail

```
Authentication Attempt:
┌─────────────────────────────────────────────────────┐
│  Attempt 1                                          │
│  ├─ 200/201 + Token → SUCCESS                       │
│  ├─ 401 → FAIL IMMEDIATELY (wrong credentials)      │
│  ├─ 403 → FAIL IMMEDIATELY (no permissions)         │
│  ├─ 429 → Wait (Retry-After or exponential backoff) │
│  └─ Network error → Wait & retry                    │
└─────────────────────────────────────────────────────┘
                         │
                         ▼ (if retry needed)
┌─────────────────────────────────────────────────────┐
│  Attempt 2-8 (same logic)                           │
│  • Exponential backoff: base_delay * attempt        │
│  • Random jitter: +0-2 seconds                      │
│  • Max attempts: 8                                  │
└─────────────────────────────────────────────────────┘
                         │
                         ▼ (if all fail)
┌─────────────────────────────────────────────────────┐
│  AuthenticationError                                │
│  Exit code: 1                                       │
└─────────────────────────────────────────────────────┘
```

This diagram shows the complete workflow from start to finish!