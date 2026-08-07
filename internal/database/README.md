# Database Engine & Provisioner Service (`internal/database`)

## Architectural Overview
The Database Engine & Provisioner Service handles automated provisioning, lifecycle orchestration (Start, Stop, Terminate), resource allocation, credential generation, and AES-256 encrypted storage of managed database instances.

```
                    +------------------------------------+
                    | REST / gRPC Provisioner API        |
                    +------------------+-----------------+
                                       |
                                       v
                    +------------------------------------+
                    |    DatabaseUseCase (Orchestrator)  |
                    +------------------+-----------------+
                                       |
                   +-------------------+-------------------+
                   |                                       |
                   v                                       v
     +---------------------------+           +---------------------------+
     | ProvisionerDriver         |           | InstanceRepository        |
     | (Docker / Kubernetes)     |           | (PostgreSQL Metadata DB)  |
     +---------------------------+           +---------------------------+
```

## Features & Isolation
- **Encrypted Credentials**: Master DB passwords are automatically generated with 128-bit entropy and stored using AES-256-GCM symmetric encryption.
- **Port Allocation**: Dynamic isolated TCP port assignment (`15000 - 25000`).
- **Quota Safeguards**: Hard limit of **5 active managed databases per project**.

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/databases` | Provisions a new managed database instance |
| `GET` | `/api/v1/databases/{id}` | Retrieves instance status & specifications |
| `GET` | `/api/v1/projects/{project_id}/databases` | Lists all database instances for a project |
| `POST` | `/api/v1/databases/{id}/start` | Starts a stopped database container |
| `POST` | `/api/v1/databases/{id}/stop` | Stops a running database container |
| `DELETE` | `/api/v1/databases/{id}` | Terminates and deletes database resources |
| `GET` | `/api/v1/databases/{id}/connection-string` | Decrypts credentials and returns connection string |
