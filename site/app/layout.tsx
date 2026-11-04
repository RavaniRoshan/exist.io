import '@fontsource-variable/mona-sans';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { Provider } from '@/components/provider';
import { siteUrl } from '@/lib/shared';
import './global.css';

export const metadata: Metadata = {
  metadataBase: new URL(`${siteUrl}/`),
  title: {
    default: 'exist.io — the self-hosted forge for coding agents',
    template: '%s · exist.io',
  },
  description:
    'exist.io is a self-hosted Git forge where coding agents hold scoped, short-lived leases, open pull requests, and merge through policy. One pure-Go binary, your hardware, Apache-2.0.',
  alternates: {
    canonical: '/',
  },
  openGraph: {
    type: 'website',
    siteName: 'exist.io',
    title: 'exist.io — the self-hosted forge for coding agents',
    description:
      'Scoped leases, pull requests, and policy-gated merges for coding agents — on your hardware, in one pure-Go binary.',
    images: [
      {
        url: 'og-exist-os.png',
        width: 1200,
        height: 630,
        alt: 'exist.io — the self-hosted forge for coding agents',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
  },
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="flex min-h-screen flex-col">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
