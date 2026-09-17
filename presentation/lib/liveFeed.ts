export type ManifestAssertion = { label: string; data?: Record<string, unknown> };
export type ManifestIngredient = { title?: string; relationship?: string };
export type ManifestReport = {
  active_manifest?: string;
  manifests?: Record<string, {
    assertions?: ManifestAssertion[];
    ingredients?: ManifestIngredient[];
    signature_info?: { issuer?: string };
  }>;
  validation_state?: string;
};
export type LivePost = {
  id: number;
  user: { username: string };
  caption: string;
  image_url: string;
  manifest_data?: ManifestReport;
  created_at: string;
};

export const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';
const API_URL = apiUrl;

// The presenter can fetch uploaded images directly from the local API,
// even when the API publishes ngrok URLs for audience phones.
export function localImageUrl(imageUrl: string): string {
  const source = new URL(imageUrl, 'http://localhost:8080');
  return `http://localhost:8080${source.pathname}${source.search}${source.hash}`;
}

export async function listRecentPosts(limit = 4): Promise<LivePost[]> {
  const response = await fetch(`${API_URL}/api/v1/posts`);
  if (!response.ok) throw new Error(`Failed to list posts: ${response.status}`);
  const posts = (await response.json()) as LivePost[];
  return posts.slice(0, limit);
}

// The go-api /posts/stream endpoint sends no backlog, only posts created
// while connected — pair with listRecentPosts for the initial view.
export function streamPosts(onPost: (post: LivePost) => void): () => void {
  const source = new EventSource(`${API_URL}/api/v1/posts/stream`);
  source.onmessage = (event) => {
    try {
      onPost(JSON.parse(event.data) as LivePost);
    } catch {
      // Ignore a malformed event rather than tearing down the stream.
    }
  };
  return () => source.close();
}

export function assertionsFor(post: LivePost): ManifestAssertion[] {
  const manifest = post.manifest_data;
  if (!manifest?.active_manifest) return [];
  return manifest.manifests?.[manifest.active_manifest]?.assertions ?? [];
}

export function signerFor(post: LivePost): string | undefined {
  const manifest = post.manifest_data;
  if (!manifest?.active_manifest) return undefined;
  return manifest.manifests?.[manifest.active_manifest]?.signature_info?.issuer;
}
