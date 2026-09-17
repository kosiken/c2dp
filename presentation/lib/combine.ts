import { apiUrl, type ManifestReport } from '@/lib/liveFeed';

export type CombineResult = { url: string; manifest: ManifestReport };

export async function combineImages(): Promise<CombineResult> {
  const response = await fetch(`${apiUrl}/api/v1/posts/combine`, { method: 'POST' });
  if (!response.ok) {
    const error = await response.json().catch(() => null);
    throw new Error(error?.message ?? `Request failed with ${response.status}`);
  }
  return response.json() as Promise<CombineResult>;
}
