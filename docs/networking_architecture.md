# ANARVA CLOUD — NETWORKING ARCHITECTURE & VPC SPECIFICATION

## Overview
Phase 11 introduces the **Anarva Cloud Networking Platform**, providing software-defined Virtual Private Clouds (VPCs), CIDR address management (IPAM), Subnet isolation (`PUBLIC`, `PRIVATE`), Route Tables, Internet Gateways, Security Groups, Private DNS, and Load Balancers.

---

## Architecture & Provider Layers

```
                      Public Internet
                             │
                      Internet Gateway
                             │
                      ANARVA VPC (10.0.0.0/16)
                             │
      ┌──────────────────────┴──────────────────────┐
      │                                             │
Public Subnet (10.0.1.0/24)             Private Subnet (10.0.2.0/24)
  - Application Load Balancer             - PostgreSQL Database (10.0.2.14)
  - ACE Compute Instances                 - Private DNS (db.anarva.internal)
```

### Feature Implementation Matrix

| Component | Status | Implementation Details |
| :--- | :--- | :--- |
| **VPC Control Plane** | REAL | Software-defined VPC network lifecycle management |
| **CIDR Validation** | REAL | Strict IPv4 CIDR math (`net.ParseCIDR`) and subnet containment checks |
| **IPAM Manager** | REAL | Private / Public IP address allocation and tracking |
| **Local Docker Driver** | LOCAL DEVELOPMENT | Docker bridge network creation (`docker network create`) when daemon is present |
| **Security Groups** | REAL | Strict backend port and protocol ingress/egress firewall evaluation |
| **Private DNS** | REAL | Internal service discovery hostnames (`*.anarva.internal`) |
| **Load Balancers** | REAL | Application and Network Load Balancers targeting Compute & Container IP endpoints |

---

## API Endpoints

- `GET /api/v1/networks`: List VPC networks
- `POST /api/v1/networks`: Create new VPC network with CIDR block
- `GET /api/v1/networks/:id`: Inspect VPC details
- `DELETE /api/v1/networks/:id`: Delete VPC network
- `POST /api/v1/networks/:id/subnets`: Add Public/Private Subnet
- `GET /api/v1/networks/:id/dns`: Retrieve Private DNS zones & records
- `POST /api/v1/networks/:id/load-balancers`: Provision Application Load Balancer
