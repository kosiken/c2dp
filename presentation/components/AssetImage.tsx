'use client';
import { useState } from 'react';
export function AssetImage({ src, alt, label }: { src: string; alt: string; label: string }) {
  const [missing, setMissing] = useState(false);
  return <div className="asset-image">
    {missing ? <div className="asset-missing"><span className="image-cross">＋</span><span>{label}</span><small>Awaiting supplied image</small></div>
      // Native img supports future uploaded image URLs without a domain allowlist.
      // eslint-disable-next-line @next/next/no-img-element
      : <img src={src} alt={alt} onError={() => setMissing(true)} />}
  </div>;
}
