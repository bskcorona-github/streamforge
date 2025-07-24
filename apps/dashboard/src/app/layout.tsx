import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'
import { Toaster } from '@/components/ui/toaster'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'StreamForge Dashboard',
  description: 'Real-time distributed system observability platform with AI-driven anomaly detection',
  keywords: ['observability', 'monitoring', 'real-time', 'streaming', 'anomaly-detection'],
  authors: [{ name: 'StreamForge Team' }],
  creator: 'StreamForge Team',
  publisher: 'StreamForge',
  robots: 'index, follow',
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://dashboard.streamforge.dev',
    title: 'StreamForge Dashboard',
    description: 'Real-time distributed system observability platform',
    siteName: 'StreamForge',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'StreamForge Dashboard',
    description: 'Real-time distributed system observability platform',
    creator: '@streamforge',
  },
  viewport: {
    width: 'device-width',
    initialScale: 1,
    maximumScale: 1,
  },
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#0f172a' },
  ],
  manifest: '/manifest.json',
  icons: {
    icon: '/favicon.ico',
    apple: '/apple-touch-icon.png',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <Providers>
          {children}
          <Toaster />
        </Providers>
      </body>
    </html>
  )
} 