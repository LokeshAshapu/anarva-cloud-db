# Anarva Cloud Platform — Release Candidate Specification (RC-1)

## Overview
This document specifies the Release Candidate 1 (RC-1) build artifact specification and verification parameters.

---

## Release Artifacts

| Component | Build Artifact | Version | Verification Status |
| :--- | :--- | :---: | :---: |
| **API Gateway** | `bin/gateway` | `0.1.0` | **PASS** |
| **Anarva CLI** | `bin/anarva` | `0.1.0` | **PASS** |
| **TypeScript SDK** | `pkg/sdk/anarva` | `0.1.0` | **PASS** (6/6 tests) |
| **Terraform Provider** | `internal/terraform/provider` | `0.1.0` | **PASS** (7/7 tests) |
| **Console UI** | `web/.next` bundle | `0.1.0` | **PASS** (42/42 routes) |

---

## Verification Commands
```bash
# 1. Go Backend Test Suite
go test ./pkg/...
go test ./internal/...

# 2. AWS Provider Integration Tests
go test -v ./internal/providers/...

# 3. Terraform Provider Tests
go test -v ./internal/terraform/...

# 4. Reliability & Security Test Suites
go test -v ./internal/reliability/...
go test -v ./internal/security/...
go test -v ./internal/observability/...

# 5. Build Gateway & CLI Binaries
go build -o bin/gateway ./cmd/gateway
go build -o bin/anarva ./cmd/anarva

# 6. Build Next.js Console
cd web && npm run build
```
