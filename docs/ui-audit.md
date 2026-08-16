# Anarva Cloud Platform — UI/UX Audit (Phase 48)

## Executive Summary
This document provides a comprehensive UI/UX audit of the ANARVA web console (`web/`). It evaluates visual hierarchy, spacing, typography, information density, contrast, component consistency, status indicators, and state handling across all 42 console routes.

---

## 1. Audit Findings Matrix

| Audit Area | Observation / Problem Discovered | Impact | Corrective Strategy |
| :--- | :--- | :--- | :--- |
| **Information Density** | Overuse of `rounded-2xl` containers and `p-6`/`p-8` padding creates sparse, floating cards rather than compact infrastructure control views. | Low operational efficiency for dense resource management. | Replace oversized card wrappers with compact, border-defined panels (`#0B1220`, `#1E293B`, `rounded-lg`). |
| **Typography Hierarchy**| Mixed font assignments (`font-sans` vs `font-mono`) without strict role definitions. Resource IDs and IPs occasionally render in variable-width fonts. | Inconsistent technical scanning. | Enforce `Inter` for UI labels/headings and `JetBrains Mono` for Resource IDs, Request IDs, IPs, ports, commands, and logs. |
| **Branding & Terminology**| External provider badges (`EC2`, `RDS`, `S3`, `ALB`, `VPC`) in navigation sidebar dilute the native ANARVA platform identity. | Confuses product positioning. | Replace all external cloud provider badges with clean ANARVA-native infrastructure labels. |
| **Color System** | Inconsistent accent colors across primary buttons (`bg-blue-600`, `bg-cyan-500`, `bg-slate-900`) and arbitrary status badge shades. | Visual noise; lacks semantic color clarity. | Implement a restrained enterprise color palette (`#080C14` root, `#0284C7` accent blue, `#22C55E` success, `#F59E0B` warning, `#EF4444` danger). |
| **Data Tables** | Table rows lack column sorting indicators, compact height options, and inline action menus. | Impairs quick resource filtering. | Upgrade `CloudTable.tsx` to support column sorting, compact 36px row heights, inline status badges, and action dropdowns. |
| **State Handling** | Resource tables use generic text shimmers during loading; empty states lack documentation links and call-to-action buttons. | Sub-optimal developer experience during async fetches. | Enhance `CloudSkeleton.tsx` with structured skeleton variants (`card`, `table`, `metric`, `detail`, `text`) and `CloudEmptyState.tsx` with docs links. |

---

## 2. Preservation Map

- **PRESERVE (KEEP)**:
  - All 42/42 Next.js App Router static routes (`/console/*`).
  - All API client integrations (`/api/v1/...`), authentication handlers, and RBAC permission guards.
  - ANARVA Trident logo motif and core dark technical aesthetic.
  - Operation timeline modal, audit log correlation search, readiness health probes, and quota meters.

- **IMPROVE**:
  - Restructure navigation sidebar into 8 clear section groups (OVERVIEW, RESOURCES, PLATFORM, SECURITY, DEVELOPER, FINANCE, FEEDBACK, SETTINGS).
  - Standardize status badges across 10 unified status codes (`HEALTHY`, `DEGRADED`, `UNAVAILABLE`, `RECOVERING`, `PROVISIONING`, `PENDING`, `RUNNING`, `COMPLETED`, `FAILED`, `UNKNOWN`).
  - High-density data tables, metric summary bars, and resource detail tabs.

- **REMOVE**:
  - Oversized floating `rounded-2xl` containers.
  - External provider badges (`EC2`, `RDS`, `S3`, `ALB`, `VPC`).
  - Neon glowing background gradients and marketing clutter inside the console.
