# Anarva Cloud Platform — Universal Provider Architecture

## Executive Summary
This document specifies the provider-neutral execution architecture for the Anarva Control Plane. ANARVA decouples control-plane domain logic, authentication, resource lock engines, and billing from specific cloud providers.

---

## Architectural Principle

> **"AWS is a provider implementation, not the ANARVA platform."**

```
[ ANARVA Developer Interfaces ] (Console / CLI / SDK / Terraform)
               │
               ▼
[ ANARVA Control Plane ] (Gateway / IAM / Locks / Operations / Audit)
               │
               ▼
[ Universal Provider Contract ] (pkg/providers)
               │
               ▼
[ Centralized Provider Registry ] (internal/providers/registry)
               │
      ┌────────┴────────┬─────────────────┬─────────────────┐
      ▼                 ▼                 ▼                 ▼
[ Local Provider ]  [ AWS Driver ]   [ GCP Driver ]   [ Azure Driver ]
  (Reference)        (Phase 52)       (Phase 53)       (Phase 54)
```

---

## Key Principles
1. **Domain Isolation**: ANARVA domain contracts (`pkg/providers`) strictly use ANARVA-native data types (`ComputeInstanceOpts`, `DatabaseProvisionOpts`, `ObjectMetadata`, `VPCSpec`). No AWS SDK types are exposed in domain interfaces.
2. **Provider Registry**: Providers register capabilities and health check handlers dynamically.
3. **Normalized Error Mapping**: Provider errors map to standard domain errors (`ErrProviderNotFound`, `ErrResourceNotFound`, `ErrResourceAlreadyExists`, `ErrUnsupportedCapability`).
