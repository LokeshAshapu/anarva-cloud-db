# Auth & User Management Service (`internal/auth`)

## Architectural Overview
The Auth service manages platform identity, authentication, session tokens, API keys, and security audit logging for the **Anarva Cloud Database Platform**. Built using Clean Architecture & DDD principles, it separates concern into `domain`, `repository`, `usecase`, and `delivery` layers.

```
                    +-----------------------------+
                    |  REST Delivery / gRPC API   |
                    +--------------+--------------+
                                   |
                                   v
                    +-----------------------------+
                    |    AuthUseCase (Business)   |
                    +--------------+--------------+
                                   |
              +--------------------+--------------------+
              |                    |                    |
              v                    v                    v
        +-----------+        +-----------+        +-----------+
        | UserRepo  |        |SessionRepo|        |APIKeyRepo |
        +-----+-----+        +-----+-----+        +-----+-----+
              |                    |                    |
              +--------------------+--------------------+
                                   |
                                   v
                    +-----------------------------+
                    |    PostgreSQL Metadata DB   |
                    +-----------------------------+
```

## Security Features
- **Passwords**: Hashed with Bcrypt (cost=12). Raw passwords are never stored.
- **Tokens**: JWT access tokens (short-lived) signed with HMAC-SHA256, and refresh tokens (long-lived) stored in DB sessions.
- **API Keys**: Formatted with high entropy (`anarva_live_...`) and hashed with SHA-256 before persistence.
- **Audit Logging**: Logs every authentication attempt and key action with IP address and User Agent.

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/auth/signup` | Registers a new user and generates email verification token |
| `POST` | `/api/v1/auth/verify-email?token=...` | Verifies user email address |
| `POST` | `/api/v1/auth/login` | Authenticates credentials and returns JWT token pair |
| `POST` | `/api/v1/auth/refresh` | Rotates access token using valid refresh token |
| `POST` | `/api/v1/auth/api-keys` | Generates a new API key for the authenticated user |
| `GET` | `/api/v1/auth/api-keys` | Lists all active API keys for the user |
| `DELETE` | `/api/v1/auth/api-keys/{id}` | Revokes an existing API key |

## Database Entities (`PostgreSQL`)
- `users`: User profile, password hash, role (`OWNER`, `ADMIN`, `DEVELOPER`), status (`PENDING`, `ACTIVE`, `SUSPENDED`).
- `sessions`: Device refresh tokens, user agents, IP addresses, expiration times.
- `api_keys`: Hashed API key strings, names, prefixes, revocation flags.
- `verification_tokens`: Time-limited tokens for email activation.
- `audit_logs`: Immutable security audit trails for identity operations.
