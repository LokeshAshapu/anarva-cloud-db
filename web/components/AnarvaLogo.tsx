import React from 'react'

export function AnarvaLogo({ className = "h-12 w-12" }: { className?: string }) {
  return (
    <div className={`relative inline-flex items-center justify-center ${className}`}>
      <img
        src="/anarva-trident.png"
        alt="Anarva Trident Logo"
        className="w-full h-full object-contain filter invert drop-shadow-[0_0_12px_rgba(59,130,246,0.6)]"
      />
    </div>
  )
}
