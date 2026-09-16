import type { Metadata } from 'next';

import './globals.css';

export const metadata: Metadata = {
  title: 'C2DP Social',
  description: 'Bare-bones image post demo'
};

export default function RootLayout({
  children
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
