# Shared Packages (`pkg/`)

This directory contains foundational, cross-cutting Go libraries used across both Control Plane and Data Plane microservices.

## Packages Overview

### 1. `pkg/errors`
- **Purpose**: Unified error handling with domain taxonomy (`ErrorCode`).
- **Features**: Structured context details (`WithDetail`), HTTP status mapping, and gRPC status code mapping.

### 2. `pkg/config`
- **Purpose**: Centralized configuration management using Viper.
- **Features**: Defaults for development, `.env` / YAML file support, and standard environment variable overrides.

### 3. `pkg/logger`
- **Purpose**: High-performance structured logging engine built on Uber Zap.
- **Features**: Context propagation (`trace_id`, `user_id`), environment-dependent log formatters (JSON for prod, colored console for dev).

### 4. `pkg/metrics`
- **Purpose**: Prometheus instrumentation metrics registry.
- **Features**: HTTP request counter/histogram, DB query counters, active connection gauges, and standard `/metrics` handler.

### 5. `pkg/security`
- **Purpose**: Cryptographic & authentication security primitives.
- **Features**: Bcrypt password hashing, JWT token pair generation & validation, AES-256-GCM symmetric payload encryption, and API key generation.

### 6. `pkg/database`
- **Purpose**: PostgreSQL database connection pool initializer built on GORM & pgx.
- **Features**: Dynamic pool bounds (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) and health check ping.
