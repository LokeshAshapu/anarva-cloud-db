# Anarva Cloud Real Provider Integration & Infrastructure Execution Layer Architecture (Phase 22)

## Overview
Anarva Cloud Phase 22 connects Anarva's control plane and provisioning engine to real infrastructure execution targets (`LOCAL_DOCKER`, `AWS`, `GOOGLE_CLOUD`) under `internal/providers/...`. It introduces Provider Registry, Capability Matrix Discovery, Provider Resource Mappings, Desired vs Observed State Reconciliation, Drift Detection Engine, Resource Import & Adoption Engine (`MANAGED = false` default), and Cloud Metadata SSRF Protections (`169.254.169.254`).

---

## Core Provider Subpackages (`internal/providers/...`)
- **`aws`**: AWS Adapter (`provider.go`, `auth.go`, `capabilities.go`, `regions.go`, `compute.go`, `network.go`, `storage.go`, `database.go`, `loadbalancer.go`, `dns.go`, `monitoring.go`, `backup.go`, `errors.go`, `mapper.go`) connecting EC2, VPC, RDS (PostgreSQL/MySQL), S3, ELB, and Route53 via AWS SDK credential chain with error mapping (`AccessDenied` -> `PROVIDER_PERMISSION_DENIED`).
- **`gcp`**: GCP Adapter (`provider.go`, `auth.go`, `capabilities.go`, `regions.go`, `zones.go`, `compute.go`, `network.go`, `storage.go`, `database.go`, `loadbalancer.go`, `dns.go`, `monitoring.go`, `backup.go`, `errors.go`, `mapper.go`) connecting Compute Engine, VPC, Cloud SQL, GCS, and Cloud Load Balancing via Application Default Credentials (ADC).
- **`registry`**: `ProviderRegistry` managing `LOCAL_DOCKER`, `AWS`, `GOOGLE_CLOUD` registration, capability matrix discovery (`GetCapabilities()`), and health verification (`CONNECTED`, `NOT_CONFIGURED`, `AUTH_FAILED`, `UNAVAILABLE`).
- **`mapping`**: `ProviderResourceMapping` maintaining authoritative links between Anarva resources and provider resource IDs.
- **`drift`**: Drift Engine comparing `DESIRED_STATE` vs `OBSERVED_STATE` (`RESOURCE_MISSING`, `STATUS_MISMATCH`, `CONFIGURATION_MISMATCH`).
- **`import`**: Resource Import, Adoption & Release orchestrator (`Import`, `Adopt`, `Release`).
- **`security`**: Metadata SSRF Protection engine blocking unauthorized access to cloud metadata endpoints (`169.254.169.254` and `metadata.google.internal`).

---

## Security & Metadata Protection Architecture (`internal/providers/security/ssrf_protection.go`)
- **Cloud Metadata Endpoint Blocklist**: Direct or indirect access to `169.254.169.254`, `169.254.169.253`, and `metadata.google.internal` is rejected with `SSRF SECURITY RISK` error.
- **Credential Storage Safety**: Credentials use IAM Role ARNs and Service Account references; secret keys are never hardcoded or displayed.

---

## Provider Reality Labels
- **Local Docker Engine**: `LOCAL_DOCKER (CONNECTED)`
- **AWS Provider**: `AWS (CONNECTED)` or `AWS (NOT_CONFIGURED)`
- **GCP Provider**: `GOOGLE_CLOUD (CONNECTED)` or `GOOGLE_CLOUD (NOT_CONFIGURED)`
