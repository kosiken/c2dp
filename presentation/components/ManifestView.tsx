export function ManifestView({ compact = false }: { compact?: boolean }) {
  return <div className={`manifest ${compact ? 'compact' : ''}`} aria-label="Manifest containing assertions, a claim referencing those assertions, and a signature over the claim">
    <div className="manifest-top"><span>C2PA MANIFEST</span><span>01</span></div>
    <div className="manifest-assertions"><span className="mono">ASSERTION STORE</span><div className="assertion-tags"><span>Actions</span><span>Ingredients</span><span>Content binding</span></div></div>
    <div className="binding-line">↓ <span>hashed references</span></div>
    <div className="manifest-claim"><span>Claim</span><small>References the assertions</small></div>
    <div className="binding-line">↓ <span>digitally signed</span></div>
    <div className="manifest-signature"><span>Claim signature</span><span>↗</span></div>
  </div>;
}
