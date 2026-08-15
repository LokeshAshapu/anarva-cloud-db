# @anarva/sdk

Official TypeScript SDK for the **Anarva Cloud Platform API (v1)**.

---

## 🔒 Security Guidelines

> **IMPORTANT:**
> API Keys (`anarva_live_...` or `anarva_test_...`) are **server-side credentials**.
> Never embed live API keys into client-side browser bundles, Next.js `NEXT_PUBLIC_` variables, or frontend code.
> Always initialize `AnarvaClient` in Node.js server environments or API routes.

---

## 🚀 Quick Start

### Installation (Local Workspace / Package Link)

```bash
npm install ./sdk
```

### Basic Usage (Server-Side)

```typescript
import { AnarvaClient, AnarvaError } from '@anarva/sdk';

const client = new AnarvaClient({
  apiKey: process.env.ANARVA_API_KEY, // Defaults to process.env.ANARVA_API_KEY
  apiUrl: process.env.ANARVA_API_URL, // Defaults to https://anarva-cloud-db-api.onrender.com
});

async function main() {
  try {
    // 1. List Managed PostgreSQL Databases
    const databases = await client.databases.list();
    console.log('Active Databases:', databases);

    // 2. Poll Control-Plane Asynchronous Operations
    const operation = await client.operations.wait('op-101', {
      timeoutMs: 60000,
      intervalMs: 2000,
    });
    console.log('Operation Status:', operation.status);
  } catch (err) {
    if (err instanceof AnarvaError) {
      console.error('Anarva API Error:', err.code, err.message, err.requestId);
    } else {
      console.error('Unexpected Error:', err);
    }
  }
}

main();
```

---

## 💻 Module Reference

| Module | Method | Description |
| :--- | :--- | :--- |
| **`client.organizations`** | `list()`, `get(id)` | List authorized organizations |
| **`client.projects`** | `list()`, `get(id)` | List projects in active organization |
| **`client.compute`** | `list()`, `create(params)` | Provision & inspect AWS EC2 compute |
| **`client.databases`** | `list()`, `get(id)`, `create(params)`, `failover(id)` | Manage RDS PostgreSQL & Multi-AZ failover |
| **`client.databases.backups`** | `list(dbId)`, `create(dbId, snapName)` | Manage database snapshot backups |
| **`client.databases.ha`** | `get(dbId)` | Inspect RDS Multi-AZ status |
| **`client.storage`** | `list()` | List Amazon S3 object storage buckets |
| **`client.metrics`** | `get(resourceId)` | Fetch real AWS CloudWatch metrics |
| **`client.billing`** | `invoices()` | Inspect customer invoices & usage breakdown |
| **`client.operations`** | `get(id)`, `wait(id, options)` | Track & poll background provisioning jobs |
