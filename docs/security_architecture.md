# Anarva Security Architecture

## Principles
1. **Anarva-Native Identity & Authorization**: The security model is native to the Anarva Control Plane.
2. **Zero-Trust & Fail-Closed**: Unauthenticated or unauthorized requests fail closed.
3. **Secret Protection & Redaction**: Secrets (API keys, JWT secrets, passwords, Bearer tokens) are never printed in logs, errors, CLI `--debug`, SDK errors, or Terraform diagnostics.
4. **Tenant Isolation**: Server-side enforced queries prevent cross-tenant access.

## Architecture Layers
- **Authentication**: JWT Manager (HMAC-SHA256) + SHA-256 Hashed API Keys.
- **Authorization**: Role-Based Access Control (`OWNER`, `ADMIN`, `DEVELOPER`, `VIEWER`, `BILLING_ADMIN`, `AUDITOR`).
- **Security Health Engine**: `/api/v1/security/status` and `/api/v1/security/events`.
- **SSRF & Storage Defense**: IP CIDR validation and object key path traversal checks.
- **Webhook Security**: Constant-time HMAC-SHA256 signature verification (`crypto/subtle.ConstantTimeCompare`).
- **Response Hardening**: Security headers, strict CORS origin matching, rate limiting (HTTP 429 + `Retry-After: 60`).
