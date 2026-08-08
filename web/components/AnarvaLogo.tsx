import React from 'react'

export function AnarvaLogo({ className = "h-8 w-8" }: { className?: string }) {
  return (
    <div className={`relative inline-flex items-center justify-center ${className}`}>
      <svg
        viewBox="0 0 100 120"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className="w-full h-full drop-shadow-[0_0_12px_rgba(59,130,246,0.5)]"
      >
        <path
          d="M50 5 L56 35 L62 5 L76 30 C74 58 64 68 56 74 L56 100 L60 105 L50 115 L40 105 L44 100 L44 74 C36 68 26 58 24 30 L38 5 L44 35 Z"
          fill="url(#trident-gradient)"
          stroke="#60A5FA"
          strokeWidth="3"
          strokeLinejoin="round"
        />
        <defs>
          <linearGradient id="trident-gradient" x1="50" y1="5" x2="50" y2="115" gradientUnits="userSpaceOnUse">
            <stop stopColor="#3B82F6" />
            <stop offset="0.5" stopColor="#06B6D4" />
            <stop offset="1" stopColor="#6366F1" />
          </linearGradient>
        </defs>
      </svg>
    </div>
  )
}
