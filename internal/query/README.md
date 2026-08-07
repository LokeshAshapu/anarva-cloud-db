# Distributed Query Engine & API Gateway (`cmd/gateway` & `internal/query`)

## Architectural Overview
The API Gateway serves as the single entrypoint for all platform traffic, managing security middleware, authentication, rate limiting, and SQL Query Execution.

```
                    +------------------------------------+
                    | REST Client / CLI / Web Dashboard  |
                    +------------------+-----------------+
                                       |
                                       v
                    +------------------------------------+
                    |           API Gateway              |
                    | (CORS -> RateLimit -> Auth JWT)    |
                    +------------------+-----------------+
                                       |
                   +-------------------+-------------------+
                   |                                       |
                   v                                       v
     +---------------------------+           +---------------------------+
     | Microservice Router       |           | Distributed Query Engine  |
     | (Auth, Project, DB Svc)   |           | (SQL Parser -> Executor)  |
     +---------------------------+           +---------------------------+
```

## Security & Parsing Pipeline
1. **CORS Middleware**: Manages cross-origin headers.
2. **Rate Limiter**: Token-bucket algorithm (default: 100 requests / min).
3. **Auth Middleware**: Validates JWT access tokens & API keys.
4. **SQL Parser & Sanitizer**: Validates statement types (`SELECT`, `INSERT`, `UPDATE`, `DELETE`, `DDL`) and blocks dangerous DDL (`DROP DATABASE`, `ALTER SYSTEM`).

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/query` | Parses, validates, and executes SQL query against managed DB instance |
| `GET` | `/health` | Gateway health check |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint |
