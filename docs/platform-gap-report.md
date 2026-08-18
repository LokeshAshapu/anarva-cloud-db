# ANARVA Cloud Control Plane — Platform Gap Report

**Repository**: `LokeshAshapu/anarva-cloud-db`  
**Auditor**: ANARVA Platform Reality Validator & End-to-End QA Engineer  
**Audit Date**: August 18, 2026  

---

## Executive Summary of Platform Gaps

This gap report identifies current platform limitations, distinguishing between production local features, control-plane metadata models, and external cloud provider requirements.

---

## Top 10 Platform Gaps & Severity Ratings

| # | Gap Description | Current Status | Severity | Recommended Upgrade Path |
| :--- | :--- | :--- | :--- | :--- |
| **1** | **Host Kernel VPC Interface Mutation** | Control-Plane Only (Metadata + Docker network if present) | **P2** | Integrate Linux `ip link` / `eBPF` network namespace plugin for bare-metal network interface attachment. |
| **2** | **Edge Load Balancer Traffic Proxying** | Control-Plane Only (Metadata listener rules) | **P2** | Embed Envoy Proxy / NGINX dynamic config reloader for live TCP/HTTP proxying. |
| **3** | **Global Multi-Region Raft Clustering** | Single-Region / Multi-AZ Synchronous Standby | **P3** | Integrate CockroachDB / YugabyteDB Raft consensus cluster coordinator for multi-continent deployments. |
| **4** | **Native Graph Query Language (Cypher)** | Not Implemented | **P3** | Implement Neo4j/Cypher graph query parser over adjacency list store. |
| **5** | **Columnar OLAP Data Warehouse Engine** | Not Implemented | **P3** | Integrate DuckDB / Apache Arrow columnar engine for multi-terabyte analytical queries. |
| **6** | **Live Docker CLI Dependency for Compute** | Falls back to in-memory status if Docker CLI is absent | **P2** | Bundle lightweight Containerd / Podman runtime wrapper. |
| **7** | **Physical Hard-Disk Volume Attachment** | Disk file storage (`./data/`) | **P2** | Implement NVMe block storage attachment driver. |
| **8** | **Automatic WAL Shipping to Cloud S3** | Local snapshot file copy | **P2** | Configure `pg_receivewal` daemon for continuous WAL streaming to S3. |
| **9** | **Live Stripe Payment Processing** | Control-plane usage meter calculation | **P3** | Wire Stripe webhooks for live payment collection. |
| **10**| **Hardware GPU Allocation for AI Vector Models**| CPU-based Vector KNN search | **P3** | Add NVIDIA CUDA GPU device passthrough for embedding generation. |

---

## What Works Perfectly Today (P0 Capabilities)

1. **Stateful SQL Engine (PostgreSQL & MySQL)**: 100% persistent across server restarts and browser refreshes.
2. **MongoDB MQL Document Store**: 100% stateful BSON document query processing.
3. **Redis Key-Value In-Memory Store**: 100% stateful command execution.
4. **S3 Object Storage**: Real file upload, download, and directory traversal security validation.
5. **Database Branching**: Copy-on-write instant database sandbox cloning.
6. **CSV & JSON Exporters**: 1-click result exports.
7. **Control-Plane Operation Recovery**: Background daemon reconciling operations.
8. **IAM & API Key Management**: Secret hashing and redaction.
