# Anarva Cloud Load Balancer & Application Delivery Platform Architecture (Phase 19)

## Overview
Anarva Cloud Phase 19 introduces a provider-neutral **Application Delivery, Load Balancing, TLS Certificate Management, Domain Verification, and Edge Routing Platform** under `internal/loadbalancer/...`. It integrates with Phase 18 Networking, Phase 13 Provisioning, Phase 8 Observability, and Phase 15 Metering & Quotas.

---

## Core Domain Models (`internal/loadbalancer/domain/domain.go`)
- **`LoadBalancer`**: Layer 7 Application & Layer 4 Network load balancers (`PUBLIC` or `INTERNAL` schemes).
- **`Listener`**: Port and protocol bindings (`HTTP`, `HTTPS`, `TCP`, `TLS`).
- **`BackendPool` & `BackendTarget`**: Target pools supporting algorithms (`ROUND_ROBIN`, `LEAST_CONNECTIONS`, `IP_HASH`, `WEIGHTED`) attached to Containers, Kubernetes Services, and Compute ACUs.
- **`LoadBalancerHealthCheck`**: Configurable health probes (`GET /health` default or custom paths).
- **`RoutingRule`**: Host-based (`api.example.com`) and path-based (`/api/*`, `/static/*`) deterministic route matching with priority conflict detection.
- **`Certificate`**: TLS certificate lifecycle (`PENDING`, `ACTIVE`, `EXPIRING`, `EXPIRED`). Private keys are strictly protected and never stored in plain text.
- **`Domain`**: Custom domain verification engine using DNS TXT / CNAME verification.
- **`Application`**: Unified application model orchestrating Network -> Container -> Load Balancer -> Domain -> TLS -> HealthCheck.

---

## Security & SSRF Protection Architecture (`internal/loadbalancer/edge/ssrf_validation.go`)
- **SSRF Origin Protection**: Origin targets and distribution endpoints are strictly validated to block probes targeting:
  - Cloud provider metadata endpoints (`169.254.169.254`, `169.254.169.253`, `169.254.0.0/16`).
  - Local loopback control plane addresses (`127.0.0.1`, `localhost`, `::1`).
  - Restricted internal control plane networks.

---

## Provider Reality Labels
- **Local Execution**: `LOCAL_LOAD_BALANCER (LIMITED_CAPABILITIES)`
- **Cloud Provider**: `PROVIDER_CONNECTED`
- **Unconfigured**: `NOT_CONFIGURED`
