# Anarva Cloud Platform — Enterprise Design System Specification (Phase 48)

## Executive Summary
This document specifies the enterprise design language, color tokens, typography hierarchy, border rules, surface levels, component styles, and spacing scale for the ANARVA Cloud Control Plane console.

---

## 1. Surface & Color System

Anarva uses a restrained, dark enterprise interface engineered for operational clarity and high contrast.

```css
/* Color Tokens */
:root {
  /* Surfaces */
  --anarva-bg-root:       #080C14; /* Deep dark background */
  --anarva-surface-1:     #0B1220; /* Primary panel & sidebar fill */
  --anarva-surface-2:     #0F172A; /* Container card & table row fill */
  --anarva-surface-3:     #111827; /* Input field & hover fill */

  /* Borders */
  --anarva-border:        #1E293B; /* Subtle Slate-800 divider */
  --anarva-border-active: #334155; /* Element focus border */

  /* Typography */
  --anarva-text-primary:   #F8FAFC; /* Slate-50 high contrast text */
  --anarva-text-secondary: #94A3B8; /* Slate-400 secondary text */
  --anarva-text-muted:     #64748B; /* Slate-500 metadata text */

  /* Primary Action Accents */
  --anarva-accent-primary: #0284C7; /* Sky-600 action button */
  --anarva-accent-hover:   #0369A1; /* Sky-700 hover state */
  --anarva-accent-bg:      rgba(2, 132, 199, 0.1);

  /* Status Colors */
  --anarva-status-healthy:     #22C55E; /* Emerald-500 */
  --anarva-status-degraded:    #F59E0B; /* Amber-500 */
  --anarva-status-unavailable: #EF4444; /* Red-500 */
  --anarva-status-info:        #38BDF8; /* Sky-400 */
}
```

---

## 2. Typography Specification

- **Primary UI Font**: `Inter`, `-apple-system`, `BlinkMacSystemFont`, `sans-serif`
  - Applied to: Navigation, page headings, button text, table headers, form labels, tooltips.
- **Infrastructure Code & Data Font**: `JetBrains Mono`, `ui-monospace`, `SFMono-Regular`, `monospace`
  - Applied to: Resource IDs (`arnv:anarva:...`), Request IDs (`req-sys-...`), IP addresses, ports, API keys, DSN connection strings, logs, CLI commands, and JSON payloads.

### Typographic Scale
- `H1 / Page Title`: `20px` (`1.25rem`), `font-semibold`, `tracking-tight`
- `H2 / Section Title`: `16px` (`1rem`), `font-semibold`, `tracking-tight`
- `H3 / Card Header`: `14px` (`0.875rem`), `font-semibold`
- `Body Text`: `13px` (`0.8125rem`), `font-normal`, `leading-relaxed`
- `Caption / Mono`: `12px` (`0.75rem`), `font-mono`
- `Micro / Status Badge`: `11px` (`0.6875rem`), `font-mono`, `font-bold`, `uppercase`

---

## 3. Status Badge System

Standardized status codes across 10 platform states:

| Status Code | Icon | Badge Text Color | Background Fill | Border Color |
| :--- | :---: | :--- | :--- | :--- |
| `HEALTHY` | `●` | `#22C55E` (Emerald) | `rgba(34, 197, 94, 0.1)` | `rgba(34, 197, 94, 0.2)` |
| `COMPLETED` | `✓` | `#22C55E` (Emerald) | `rgba(34, 197, 94, 0.1)` | `rgba(34, 197, 94, 0.2)` |
| `DEGRADED` | `▲` | `#F59E0B` (Amber) | `rgba(245, 158, 11, 0.1)` | `rgba(245, 158, 11, 0.2)` |
| `RECOVERING` | `⟳` | `#F59E0B` (Amber) | `rgba(245, 158, 11, 0.1)` | `rgba(245, 158, 11, 0.2)` |
| `UNAVAILABLE` | `✕` | `#EF4444` (Red) | `rgba(239, 68, 68, 0.1)` | `rgba(239, 68, 68, 0.2)` |
| `FAILED` | `✕` | `#EF4444` (Red) | `rgba(239, 68, 68, 0.1)` | `rgba(239, 68, 68, 0.2)` |
| `PROVISIONING` | `⟳` | `#38BDF8` (Sky) | `rgba(56, 189, 248, 0.1)` | `rgba(56, 189, 248, 0.2)` |
| `RUNNING` | `⟳` | `#38BDF8` (Sky) | `rgba(56, 189, 248, 0.1)` | `rgba(56, 189, 248, 0.2)` |
| `PENDING` | `○` | `#94A3B8` (Slate) | `rgba(148, 163, 184, 0.1)`| `rgba(148, 163, 184, 0.2)`|
| `UNKNOWN` | `?` | `#64748B` (Slate) | `rgba(100, 116, 139, 0.1)`| `rgba(100, 116, 139, 0.2)`|

---

## 4. Component Design Specifications

- **Container Panels**: Fill `#0F172A`, Border 1px `#1E293B`, Radius `8px` (`rounded-lg`).
- **Primary Buttons**: Fill `#0284C7`, Hover `#0369A1`, Text `#F8FAFC`, Font `Inter 12px font-semibold`.
- **Secondary Buttons**: Fill `#111827`, Border 1px `#1E293B`, Hover `#1F2937`, Text `#F8FAFC`.
- **Data Tables**: Header fill `#0B1220`, Row fill `#0F172A`, Row hover `#1E293B/50`, Row height `36px`, Dividers 1px `#1E293B`.
- **Inputs & Selects**: Fill `#111827`, Border 1px `#1E293B`, Focus border `#0284C7`, Text `#F8FAFC`.
