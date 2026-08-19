# ANARVA Phase 60 Forensic Audit: Localhost Fallback Elimination & Version Tracking

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Auditor**: Senior Staff Database Architect, Cloud Security Lead & SRE  
**Audit Date**: August 19, 2026  
**Target Endpoint**: `GET /api/v1/health/persistence`  
**Production API**: `https://anarva-cloud-db-api.onrender.com`  
**Commit**: `phase-60-cd8ca2a`  
**Status**: **VIPER MAPPING BUG FIXED, LOCALHOST ELIMINATED, COMMIT VERSION EXPOSED**  

---

## 1. Table of Configuration Sources & Production Risks

| Source | Value | Used By | Production Risk |
| :--- | :--- | :--- | :--- |
| `v.SetDefault("DATABASE.HOST", ...)` | `"localhost"` | Development Default | **HIGH**: Previously selected if `DATABASE_URL` was unmapped by Viper. |
| `v.SetDefault("DATABASE.DB_NAME", ...)` | `"anarva_cloud_db"` | Development Default | **MEDIUM**: Development DB name. |
| `v.SetDefault("DATABASE.SSL_MODE", ...)` | `"disable"` | Development Default | **HIGH**: Disables TLS. Incompatible with production PostgreSQL. |
| `os.Getenv("DATABASE_URL")` | Render DB URL | Production Store | **AUTHORITATIVE**: Must override all defaults. |
| `DatabaseConfig.URL` | Dynamic DSN String | GORM Connection Pool | **AUTHORITATIVE**: Read directly in `DSN()`. |

---

## 2. Forensic Discovery of the Exact Localhost Source

1. **Viper Struct Tag Bug**:
   In [`pkg/config/config.go`](file:///c:/Users/ASUS/Downloads/anarva-cloud-db/pkg/config/config.go#L33), `DatabaseConfig.URL` was tagged as ``mapstructure:"URL"`` while nested under `Database DatabaseConfig \`mapstructure:"DATABASE"\``. Viper looked for `DATABASE.URL` (env var `DATABASE_URL`). But without explicit `v.BindEnv("DATABASE.URL", "DATABASE_URL")`, `v.Unmarshal(&cfg)` left `cfg.Database.URL` empty.
2. **Fallback in `DSN()`**:
   When `cfg.Database.URL` was empty, `DSN()` constructed `"host=localhost port=5432 user=anarva_admin password=anarva_password dbname=anarva_cloud_db sslmode=disable"`.
3. **Execution on Render**:
   When `DATABASE_URL` was not configured in Render environment variables or when `ANARVA_ENV` defaulted to `"development"`, the gateway dialed `127.0.0.1:5432`, failing with `connection refused`.

---

## 3. Engineering Fixes & Build Commit Tracking

1. **Explicit Tag & Bind in `pkg/config/config.go`**:
   - `DatabaseConfig.URL` set to ``mapstructure:"URL"``.
   - `v.BindEnv("DATABASE.URL", "DATABASE_URL")` explicitly added in `LoadConfig`.
   - Direct environment override:
     ```go
     if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
         cfg.Database.URL = dbURL
     }
     ```
2. **DSN Priority Engine & Production Lockout**:
   - `DSN()` checks `db.URL` and `os.Getenv("DATABASE_URL")` first.
   - If `ANARVA_ENV=production` or `APP_ENV=production` and `DATABASE_URL` is missing, `DSN()` returns an empty string (`""`) and **NEVER** falls back to `localhost`.
3. **Gateway Startup Assertion**:
   - Added startup checks in [`cmd/gateway/main.go`](file:///c:/Users/ASUS/Downloads/anarva-cloud-db/cmd/gateway/main.go):
     ```go
     if appEnv == "production" {
         if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
             log.Fatal("FATAL: Production mode requires a valid DATABASE_URL environment variable")
         }
         if u, err := url.Parse(os.Getenv("DATABASE_URL")); err == nil {
             if u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" {
                 log.Fatal("FATAL: Production DATABASE_URL resolves to localhost")
             }
         }
     }
     ```
4. **Deployed Commit Tracking**:
   - Exposed `build` metadata (`gitCommit: "phase-60-cd8ca2a"`) inside `GET /api/v1/health` and `GET /api/v1/health/persistence`.
   - Allows instant verification of whether the deployed Render container is running the latest binary.

---

## 4. Render Deployment & Forensic Diagnostic Verification

1. Set Environment Variables in Render:
   ```bash
   ANARVA_ENV=production
   DATABASE_URL=postgres://anarva_user:password@dpg-xxxx-a.render.com/anarva_db?sslmode=require
   ```
2. Trigger deployment on Render.
3. Open `GET https://anarva-cloud-db-api.onrender.com/api/v1/health/persistence`.
4. Verify response:
   - `build.gitCommit`: `"phase-60-cd8ca2a"`
   - `database.configuration_source`: `"DATABASE_URL"`
   - `database.hostname`: `"dpg-xxxx-a.render.com"` (NOT `localhost`)
   - `database.sslmode`: `"require"` (NOT `disable`)
   - `diagnostics.postgres_ping`: `"PASS"`
