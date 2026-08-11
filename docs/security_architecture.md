# Anarva Cloud Enterprise IAM & Security Architecture Documentation

## 1. Security Hierarchy & Model

Anarva Cloud enforces zero-trust authorization at every boundary:

```
AUTHENTICATED USER
└── ORGANIZATION (Anarva Systems)
    └── PROJECT (Anarva Cloud Platform)
        └── RESOURCE (arnv:db:ap-hyderabad-1:proj-default:database/production-db)
            └── ACTION (database.query / database.delete)
```

---

## 2. Password Security Audit

- **Authentication Provider**: Powered by Supabase Auth (`@supabase/supabase-js`, `@supabase/ssr`).
- **Backend Hashing**: Password hashes are securely computed on Supabase Auth infrastructure using **bcrypt / Argon2**.
- **Client Hashing Removal**: Client-side plain SHA-256 pre-hashing was removed to prevent weak hash key collisions and rely on Supabase Auth's native salted password pipeline.

---

## 3. API Key Security & SHA-256 Masking

- **Key Generation**: API secret keys generate a unique prefix (`ak_...`) and full secret (`anarva_live_ak_...`).
- **Backend Storage**: The full secret is returned **ONLY ONCE** to the user upon creation. The application database stores only the `sha256(fullSecret)` hash.
- **Revocation**: Instant server-side revocation sets `revoked_at` timestamp.

---

## 4. Threat Model & Mitigation Matrix

| Asset | Threat Actor | Potential Threat | Mitigation Strategy |
| :--- | :--- | :--- | :--- |
| **API Endpoints** | Malicious IP | Brute-force & DoS | RateLimiter Middleware (100 req/sec per IP) |
| **Tenant Data** | Cross-Tenant User | IDOR / Unauthorized Access | Backend OrganizationID & ProjectID validation on all CRUD |
| **Database Credentials** | External Sniffer | Secret Exposure | Passwords masked in UI, TLS 1.3 enforced |
| **Client Browsers** | Malicious Script | XSS / Clickjacking | Next.js Security Headers (`CSP`, `X-Frame-Options: DENY`, `HSTS`) |

---

## 5. Security Headers (`web/next.config.js`)

- `Strict-Transport-Security`: `max-age=31536000; includeSubDomains`
- `X-Content-Type-Options`: `nosniff`
- `X-Frame-Options`: `DENY`
- `X-XSS-Protection`: `1; mode=block`
- `Referrer-Policy`: `strict-origin-when-cross-origin`
- `Permissions-Policy`: `camera=(), microphone=(), geolocation=()`
