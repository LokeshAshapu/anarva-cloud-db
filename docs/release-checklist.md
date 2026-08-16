# Anarva Cloud Platform — Release Checklist

Use this checklist prior to cutting a production release of the Anarva Cloud Platform.

- [x] **Configuration Validated**: `ValidateProductionConfig` passes with fail-closed rules.
- [x] **Secrets Verified**: Zero raw API keys, passwords, JWT tokens, or credentials exposed in logs or errors.
- [x] **Database Migrations Verified**: 0 destructive migrations; PostgreSQL connection pool settings verified.
- [x] **Backend Go Tests**: 100% Go backend tests passing (`go test ./internal/...`, `go test ./pkg/...`).
- [x] **Provider Tests**: 29/29 AWS provider integration tests passing (`go test -v ./internal/providers/...`).
- [x] **SDK Tests**: 6/6 TypeScript SDK tests passing (`go test ./pkg/sdk/anarva/...`).
- [x] **Terraform Tests**: 7/7 Terraform provider tests passing (`go test -v ./internal/terraform/...`).
- [x] **Console Production Build**: Next.js 14 build passes with all 42/42 routes statically compiled (`npm run build`).
- [x] **CLI Build**: Anarva CLI binary compiles cleanly (`go build -o bin/anarva ./cmd/anarva`).
- [x] **Security Release Gate**: 10/10 security test categories passing (`go test -v ./internal/security/...`).
- [x] **Health Endpoint**: `/health` liveness probe returns HTTP 200 `{"status":"UP"}`.
- [x] **Readiness Endpoint**: `/readiness` probe returns HTTP 200 `{"status":"READY"}` when database pool is healthy.
- [x] **Operation Recovery Test**: Operation state persistence, lease renewal, and restart recovery verified.
- [x] **No Secrets in Logs**: Redaction regex utility active across all logger outputs.
- [x] **No Dev Fallback in Prod**: In-memory repositories disabled when `ENVIRONMENT=production`.
- [x] **Version Synchronized**: `pkg/version.ANARVA_VERSION` set to `0.1.0`.
- [x] **Release Artifacts Generated**: `bin/gateway`, `bin/anarva`, and Next.js `.next` bundle built successfully.
