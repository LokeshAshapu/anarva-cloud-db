# Anarva Cloud Platform — Security Model & Threat Boundary

## Executive Summary
Anarva is designed around a Zero-Trust security model. All requests—whether originating from the Anarva Console, CLI, TypeScript SDK, Terraform Provider, or external API consumers—must present valid, authenticated credentials and pass server-side authorization, tenant isolation, rate limiting, and input validation checks.

---

## 1. Authentication & Identity Boundary
- **Password Security**: Passwords are hashed using bcrypt with cost 12 (`pkg/security.HashPassword`).
- **JWT Manager**: API sessions issue HMAC-SHA256 signed JSON Web Tokens (JWTs) containing `user_id`, `org_id`, and `role`. Tokens expire in 1 hour; refresh tokens expire in 24 hours.
- **API Keys**: API keys use prefixes (`anarva_live_` / `anarva_test_`). The raw secret is displayed ONLY ONCE upon creation. The control plane stores only a SHA-256 hash (`pkg/security.HashAPIKey`).

---

## 2. Multi-Tenant Isolation
- **Strict Query Isolation**: All control-plane database queries use parameterized GORM queries with explicit organization boundaries:
  `SELECT * FROM instances WHERE organization_id = ? AND project_id = ? AND id = ?`
- **Safe Rejection**: Attempts to access resources outside a tenant's authorized organization return `HTTP 404 NOT_FOUND` with error code `RESOURCE_NOT_FOUND` to prevent resource existence probing.

---

## 3. Webhook & Signature Security
- **Constant-Time Verification**: Webhook payload signatures are computed using HMAC-SHA256 (`ComputeHMACSignature`). Verification uses `crypto/subtle.ConstantTimeCompare` (`VerifyHMACSignature`) to defend against timing attacks.

---

## 4. SSRF & Storage Path Traversal Protection
- **SSRF Engine (`internal/providers/security`)**: Blocks webhook and external integration requests targeting:
  - Loopback (`127.0.0.1`, `localhost`, `::1`)
  - Private CIDRs (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`)
  - Link-local addresses (`169.254.0.0/16`)
  - Cloud metadata endpoints (`169.254.169.254`, `metadata.google.internal`)
- **Storage Path Traversal (`internal/storage/provider`)**: `ValidateObjectKey()` rejects keys containing `../`, `..\`, null bytes (`\x00`), leading slashes, and URL-encoded traversal patterns (`%2e%2e%2f`).

---

## 5. Defense-in-Depth Middleware Stack
1. **Security Headers Middleware**: Injects `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy`, and `Content-Security-Policy`.
2. **Correlation Middleware**: Assigns a unique `X-Request-ID` header for end-to-end audit tracing.
3. **CORS Middleware**: Restricts allowed origins in production mode with strict credential support.
4. **Rate Limit Middleware**: Enforces bucket limits, returning `HTTP 429 TOO_MANY_REQUESTS` with `Retry-After: 60` and logging security events.
5. **Auth Middleware**: Validates JWTs or API keys, setting request context user ID, org ID, and role.

---

## 6. Secret Redaction Guarantee
`pkg/security.RedactSecrets()` automatically redacts API keys, webhook secrets, passwords, Bearer headers, JWT tokens, DSN passwords, and AWS credentials across all loggers, audit events, CLI output, SDK errors, and Terraform diagnostics.
