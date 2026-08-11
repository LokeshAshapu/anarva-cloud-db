# ANARVA CLOUD — COMPUTE PLATFORM & ANARVA COMPUTE UNITS (ACUs)

## 1. Executive Overview

Anarva Compute Platform is a production-oriented control plane for provisioning, scaling, and managing containerized and virtual compute workloads across cloud regions.

Compute capacity is abstracted into **Anarva Compute Units (ACUs)**:
- **1.0 ACU** = 1.0 vCPU + 2.0 GB RAM + 20 GB NVMe Storage + 250 Mbps Network.
- Standard ACU Tiers: `0.5`, `1.0`, `2.0`, `4.0`, `8.0`, `16.0`, `32.0`, `64.0`, `128.0`.

> **Disclaimer**: ACU is an Anarva-defined platform compute abstraction designed for predictable workload sizing and multi-cloud scheduling.

---

## 2. Architecture & Directory Structure

The compute service follows clean domain-driven architecture within the Go API Gateway:

```
internal/compute/
├── domain/
│   ├── compute.go      # Domain models (ComputeInstance, ComputePlan, Volume, SecurityGroup)
│   └── repository.go   # ComputeRepository & VolumeRepository interfaces
├── provider/
│   └── provider.go     # ComputeProvider interface & LocalDockerComputeProvider
├── usecase/
│   └── compute_usecase.go # Lifecycle state machine, ACU validation, & volume management
└── delivery/http/
    └── compute_handler.go # REST API HTTP routes
```

---

## 3. Provider Abstraction & Development Environment

The `ComputeProvider` interface decouples the control plane from target infrastructure:
- **`LOCAL DEVELOPMENT PROVIDER` (`LocalDockerComputeProvider`)**: Executed when Docker daemon CLI is present on the host. Spawns actual container workloads with CPU (`--cpus`) and memory (`--memory`) cgroup constraints. When Docker CLI is absent, seamlessly operates in local control-plane simulation mode labeled `LOCAL SIMULATION PROVIDER`.
- **Planned Cloud Providers**: Kubernetes (K8s), AWS EC2, GCP Compute Engine, Azure VMs, Bare Metal.

---

## 4. Compute Instance Lifecycle & Provisioning State Machine

```
REQUESTED
   ↓
VALIDATING (ACU validation & quota check)
   ↓
SCHEDULING (Region & Zone placement)
   ↓
PROVISIONING (Container / VM allocation)
   ↓
CONFIGURING (Volume attachment & env injection)
   ↓
HEALTH_CHECK
   ↓
RUNNING ↔ STOPPING ↔ STOPPED
   ↓
DELETED
```

---

## 5. REST API Specifications

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/v1/compute/plans` | List standard ACU capacity plans (0.5 - 128 ACUs) |
| `GET` | `/api/v1/compute/images` | List available OS & container images |
| `GET` | `/api/v1/compute/instances` | List project compute instances |
| `POST` | `/api/v1/compute/instances` | Provision a new compute instance |
| `GET` | `/api/v1/compute/instances/:id` | Fetch instance details |
| `POST` | `/api/v1/compute/instances/:id/start` | Start a stopped instance |
| `POST` | `/api/v1/compute/instances/:id/stop` | Stop a running instance |
| `POST` | `/api/v1/compute/instances/:id/restart` | Restart instance |
| `POST` | `/api/v1/compute/instances/:id/execute` | Execute container command inside instance |
| `DELETE` | `/api/v1/compute/instances/:id` | Terminate compute instance |

---

## 6. Audit & Security Control

- **Audit Events**: `COMPUTE_CREATED`, `COMPUTE_STARTED`, `COMPUTE_STOPPED`, `COMPUTE_DELETED`, `VOLUME_CREATED`, `COMMAND_EXECUTED`.
- **Command Security**: Command payloads inside container execution drawers are sanitized, logged to audit streams, and restricted from host socket escalation.
