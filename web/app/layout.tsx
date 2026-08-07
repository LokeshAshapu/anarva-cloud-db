import './globals.css'
import React from 'react'

export const metadata = {
  title: 'Anarva Cloud DB - Enterprise Managed Database Platform',
  description: 'Serverless & Managed Cloud Database Platform with Distributed Query Engine, Automated Provisioning, and Zero-Trust Security.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className="bg-slate-950 text-slate-100 antialiased min-h-screen">
        {children}
      </body>
    </html>
  )
}
