import type { Metadata } from 'next';
import './globals.css';
export const metadata: Metadata = {
  title: 'C2PA and Digital Content Provenance',
  description: 'An engineering presentation by Kosy at Rise.',
};
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return <html lang="en"><body>{children}</body></html>;
}
