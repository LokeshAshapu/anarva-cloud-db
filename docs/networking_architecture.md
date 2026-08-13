# Anarva Cloud Networking Architecture (Phase 18)

## Overview
Anarva Cloud Phase 18 introduces a provider-neutral **Virtual Private Cloud (VPC), Networking, Security Group, IPAM, DNS, and SSRF-Protected Connectivity Platform** under `internal/networking/...`. It extends Phase 11 networking capabilities with strict tenant isolation, CIDR overlap detection, database port security validation, and state reconciliation.

---

## Core Domain Models (`internal/networking/domain/domain.go`)
- **`VirtualNetwork`**: Multi-tenant virtual private cloud metadata including CIDR, region, DNS resolution flags, and reality labels (`LOCAL_NETWORK (LIMITED_CAPABILITIES)`).
- **`Subnet`**: Subnet partitioning with explicit types (`PUBLIC`, `PRIVATE`, `ISOLATED`). Private subnets block direct inbound internet routing.
- **`RouteTable` & `Route`**: Destination CIDR targets (`LOCAL`, `INTERNET_GATEWAY`, `NAT_GATEWAY`, `NETWORK_INTERFACE`, `PEERING`).
- **`SecurityGroup` & `SecurityRule`**: Stateful firewall policies with default `DENY` inbound and `ALLOW` outbound rules.
- **`NetworkInterface`**: Virtual network interfaces bound to Compute ACUs, PostgreSQL instances, and Load Balancers.
- **`ConnectivityTest`**: Real-time reachability and latency probe with **SSRF Protection**.
- **`DNSZone` & `DNSRecord`**: Private and Public DNS zone resolution supporting `A`, `AAAA`, `CNAME`, `TXT`, `MX`, `NS`.

---

## Security & Protection Architecture

### 1. Database Port 5432 Inbound Protection
- PostgreSQL instances default to `PRIVATE` network attachment.
- Security group rule validation rejects ingress rules allowing `0.0.0.0/0` on port `5432` unless explicitly authorized by platform administrators.

### 2. SSRF Protection Engine (`internal/networking/connectivity/connectivity_service.go`)
- Connectivity testing enforces strict SSRF filters blocking:
  - Cloud provider metadata endpoints (`169.254.169.254`, `169.254.169.253`, link-local ranges `169.254.0.0/16`).
  - Local loopback control plane addresses (`127.0.0.1`, `localhost`, `::1`).
  - Cross-tenant network boundaries.

---

## Provider Reality Labels
- **Local Network**: `LOCAL_NETWORK (LIMITED_CAPABILITIES)`
- **Cloud Provider**: `PROVIDER_NETWORK_CONNECTED`
- **Unconfigured**: `PROVIDER_NETWORK_NOT_CONFIGURED`
