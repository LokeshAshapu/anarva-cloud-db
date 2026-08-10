// Anarva Cloud Platform — Centralized Design Tokens

export const tokens = {
  colors: {
    bg: {
      primary: '#020617', // slate-950
      secondary: '#0f172a', // slate-900
      tertiary: '#1e293b', // slate-800
      hover: '#1e293b',
    },
    brand: {
      blue: '#3b82f6', // electric blue
      violet: '#8b5cf6', // violet
      cyan: '#06b6d4', // cyan
    },
    status: {
      success: '#10b981', // emerald-500
      warning: '#f59e0b', // amber-500
      danger: '#ef4444', // red-500
      info: '#3b82f6',
    },
    border: {
      subtle: 'rgba(51, 65, 85, 0.7)', // slate-700/70
      muted: 'rgba(30, 41, 59, 0.9)', // slate-800/90
      active: 'rgba(59, 130, 246, 0.4)', // blue-500/40
    },
    text: {
      primary: '#f8fafc', // slate-50
      secondary: '#94a3b8', // slate-400
      muted: '#64748b', // slate-500
      brand: '#60a5fa', // blue-400
    },
  },
  typography: {
    fontSans: 'var(--font-inter), system-ui, -apple-system, sans-serif',
    fontMono: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
  },
  radius: {
    sm: '0.375rem',
    md: '0.5rem',
    lg: '0.75rem',
    xl: '1rem',
    full: '9999px',
  },
  shadows: {
    card: '0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.4)',
    glow: '0 0 20px -3px rgba(59, 130, 246, 0.25)',
  },
  transitions: {
    fast: 'all 150ms cubic-bezier(0.4, 0, 0.2, 1)',
    normal: 'all 250ms cubic-bezier(0.4, 0, 0.2, 1)',
  },
} as const
