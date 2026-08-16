# Anarva Cloud Platform — Platform Limitations & Non-Goals

## Overview
This document records explicit platform limitations, non-goals, and boundary conditions for the Anarva Cloud Platform.

---

## 1. Provider Execution Mode
- **Provider Abstraction Mode**: The control plane uses an abstraction layer (`ProviderRegistry`) supporting Docker and AWS execution providers. In non-production testing, provider execution operates in verified simulation/mock mode (`SIMULATED — NOT PRODUCTION EXECUTION`).

## 2. Multi-Region Active-Active Replication
- Anarva supports region selection and Multi-AZ database failover; global active-active cross-region database replication is explicitly out of scope.

## 3. Kubernetes Orchestration
- Anarva manages containerized compute workloads natively through Anarva Compute and Docker/EC2 abstraction drivers; Kubernetes cluster orchestration is not used.

## 4. Payment Gateway Integration
- Anarva includes a native metered billing, quota enforcement, and invoice generation engine (`internal/billing`); real credit-card payment gateway processing (e.g. Stripe) is handled via external webhooks.
