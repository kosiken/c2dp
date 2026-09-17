'use client';
// Replace each component independently with its own HTTP/SSE lifecycle and cleanup.
// Demo state belongs here, not in the presentation navigation shell.
import { useEffect, useState } from 'react';
import { uploadQrDataUri, uploadUrl } from '@/lib/uploadQr';
import { assertionsFor, listRecentPosts, signerFor, streamPosts, type LivePost } from '@/lib/liveFeed';
import { combineImages, type CombineResult } from '@/lib/combine';

const steps = ['Take a photo on your phone', 'Upload it to the site', 'Make a post'];
export function DemoThirteen() {
  return <div className="demo-slide qr-slide split">
    <div>
      <p className="eyebrow">Your turn</p>
      <h1>You're the content<br /><em>creators</em> now.</h1>
      <div className="qr-steps">{steps.map((step, i) => <div key={step}><span className="mono accent">0{i + 1}</span><p>{step}</p></div>)}</div>
      <p className="bottom-note">Your phone becomes the beginning of the provenance pipeline.</p>
    </div>
    <div className="qr-panel">
      {uploadUrl
        ? <><img src={uploadQrDataUri} alt="QR code linking to the image upload site" className="qr-image" /><span className="mono qr-url">{uploadUrl}</span></>
        : <div className="asset-missing"><span className="image-cross">＋</span><span>Upload link</span><small>Awaiting NEXT_NGROK_WEB_APP_URL</small></div>}
    </div>
  </div>;
}
const FEED_LIMIT = 4;
export function DemoFourteen() {
  const [posts, setPosts] = useState<LivePost[]>([]);
  useEffect(() => {
    let cancelled = false;
    listRecentPosts(FEED_LIMIT).then((initial) => { if (!cancelled) setPosts(initial); }).catch(() => {});
    const stop = streamPosts((post) => {
      setPosts((current) => [post, ...current.filter((p) => p.id !== post.id)].slice(0, FEED_LIMIT));
    });
    return () => { cancelled = true; stop(); };
  }, []);
  return <div className="demo-slide feed-slide">
    <p className="eyebrow">Live from the room</p>
    <h1>Watching provenance<br /><em>arrive.</em></h1>
    {posts.length === 0
      ? <div className="demo-center"><span className="status-dot" /><p>Waiting for uploads</p><span className="mono muted">Scan the QR code on the previous slide</span></div>
      : <div className="feed-grid">{posts.map((post) => {
          const assertions = assertionsFor(post);
          const signer = signerFor(post);
          return <div className="feed-card" key={post.id}>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src={post.image_url} alt={post.caption || 'Uploaded photo'} />
            <div className="feed-meta">
              <span className="mono accent">@{post.user.username}</span>
              <span className="feed-line">{assertions.map((a) => a.label).join(' · ') || 'No assertions'}</span>
              <span className="feed-line">{post.manifest_data?.validation_state ?? 'Unverified'}{signer ? ` · ${signer}` : ''}</span>
              <span className="mono muted">{new Date(post.created_at).toLocaleTimeString()}</span>
            </div>
          </div>;
        })}</div>}
  </div>;
}
type CombineState = { status: 'idle' | 'loading' | 'error'; result: CombineResult | null; error: string };
const initialCombineState: CombineState = { status: 'idle', result: null, error: '' };
export function DemoFifteen() {
  const [state, setState] = useState<CombineState>(initialCombineState);
  const generate = () => {
    setState({ status: 'loading', result: null, error: '' });
    combineImages()
      .then((result) => setState({ status: 'idle', result, error: '' }))
      .catch((err) => setState({ status: 'error', result: null, error: err instanceof Error ? err.message : 'Could not generate combination' }));
  };
  const active = state.result?.manifest.active_manifest;
  const manifest = active ? state.result?.manifest.manifests?.[active] : undefined;
  const sources = (manifest?.ingredients ?? []).filter((i) => i.relationship === 'componentOf').map((i) => i.title).filter(Boolean);
  return <div className="demo-slide combine-slide split">
    <div>
      <p className="eyebrow">One more transformation</p>
      <h1>Provenance through<br /><em>transformation.</em></h1>
      <p className="lead muted">Combines the four most recent uploads into one new signed composite.</p>
      <button className="text-button generate-button" onClick={generate} disabled={state.status === 'loading'}>{state.status === 'loading' ? 'Generating…' : 'Generate Combination'} <span>↗</span></button>
      {state.status === 'error' && <p className="combine-error">{state.error}</p>}
      {state.result && <div className="combine-summary">
        <span className="mono muted">SOURCES COMBINED</span>
        <p>{sources.length ? sources.join(' · ') : `${manifest?.ingredients?.length ?? 0} images`}</p>
        <span className="mono muted">VALIDATION</span>
        <p>{state.result.manifest.validation_state ?? 'Unknown'}</p>
      </div>}
    </div>
    <div className="combine-panel">
      {state.result
        // eslint-disable-next-line @next/next/no-img-element
        ? <img src={state.result.url} alt="Combined composite image from four audience uploads" className="combine-image" />
        : <div className="asset-missing"><span className="image-cross">＋</span><span>Composite</span><small>Not generated yet</small></div>}
    </div>
  </div>;
}
