# Web Dashboard Application (`web/`)

## Features & Technology Stack
Built using **Next.js 14 App Router**, **React 18**, **TypeScript**, and **Tailwind CSS**.

### Key Pages & Features
1. **Overview Dashboard (`/dashboard`)**: Real-time platform metrics, active database count, query latency telemetry, and storage quota utilization.
2. **Projects & Organizations (`/dashboard/projects`)**: Multi-tenancy hierarchy management and region selection (`us-east-1`, `eu-central-1`, `ap-southeast-1`).
3. **Managed Databases (`/dashboard/databases`)**: Interactive database provisioner console with Start, Stop, Delete controls and connection string reveal modals.
4. **SQL Query Console (`/dashboard/query`)**: Live SQL editor with statement safety parser validation and dynamic query results table.
5. **Backups & PITR (`/dashboard/backups`)**: Manual & automated snapshot triggers, backup list, and one-click restore.
6. **API Keys & Security (`/dashboard/apikeys`)**: Management of SHA-256 hashed API access keys.

## Development Commands
```bash
cd web
npm install
npm run dev
```
