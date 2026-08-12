# Anarva Cloud — Phase 14: Developer Platform Architecture

## 1. Overview

The **Anarva Developer Platform** turns Anarva Cloud into an accessible developer platform with standardized REST APIs (`/api/v1/`), API key authentication (`ank_live_...`), service accounts for CI/CD automation, a Go CLI binary (`anarva-cli`), an official Go SDK (`pkg/sdk/anarva`), HMAC-signed webhooks, and an interactive Developer Portal (`/console/developer`).

---

## 2. API Response Standard & Correlation

All public developer API endpoints return standardized JSON structures:

### Success Response (HTTP 200 OK / HTTP 201 Created)
```json
{
  "data": {
    "id": "acu-instance-8f12",
    "name": "ace-worker-node-01",
    "status": "RUNNING",
    "acu": 1.0
  },
  "meta": {
    "requestId": "req_172348102930"
  }
}
```

### Error Response (HTTP 4xx / 5xx)
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Compute instance 'invalid-id' not found in project context",
    "requestId": "req_172348102930"
  }
}
```

---

## 3. Developer API Keys & Security Model

- **Key Prefixes**: `ank_live_` (Production), `ank_test_` (Development).
- **Key Hashing**: Secrets are hashed with **SHA-256** zero-trust encryption (`keyHash`). Plaintext secrets are displayed **only once** upon creation and are never logged or stored.
- **Permissions**: Scoped permissions (`compute.read`, `compute.create`, `database.read`, `storage.read`, `network.read`, `provisioning.read`).
- **Rate Limits**: Server-side headers (`RateLimit-Limit: 100`, `RateLimit-Remaining: 99`, `RateLimit-Reset: 60`).

---

## 4. Webhook Engine & SSRF Protection

- **Payload Signatures**: Every webhook request includes an **HMAC-SHA256** signature in the `X-Anarva-Signature` header.
- **SSRF Protection**: Webhook targets pointing to loopback (`127.0.0.0/8`, `::1`), private networks (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), or cloud metadata endpoints (`169.254.169.254`) are strictly blocked server-side.

---

## 5. Developer Go CLI (`anarva`)

### Installation & Configuration
```bash
# Configure API key session
anarva configure --key ank_live_8f3a921b...

# Check identity
anarva whoami

# List compute instances (text or JSON)
anarva compute list
anarva compute list --json
```

---

## 6. Official Go SDK (`anarva`)

```go
package main

import (
    "context"
    "fmt"
    "github.com/anarva-cloud/anarva-cloud-db/pkg/sdk/anarva"
)

func main() {
    client := anarva.NewClient("ank_live_8f3a921b...", "http://localhost:8080")
    instances, err := client.Compute.List(context.Background(), "proj-default")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Compute instances: %v\n", instances)
}
```
