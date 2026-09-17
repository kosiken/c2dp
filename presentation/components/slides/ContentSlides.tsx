'use client';
import { useState } from 'react';
import { AssetImage } from '../AssetImage';
import { ManifestView } from '../ManifestView';
import { assets, journey, lifecycle, terms, exampleManifest, specificationUrl } from '@/lib/content';

export function Cover() {
  return <div className="cover split"><div><p className="eyebrow">An engineering perspective</p><h1>C2PA <span className="muted">and</span><br />Digital Content<br /><em>Provenance.</em></h1><p className="cover-sub">Every piece of content has a story.<br />How much of it can we verify?</p><p className="byline"><strong>Kosy</strong><span>Developer, Rise</span></p></div><div className="cover-diagram"><ManifestView compact /><span className="diagram-caption">CONTENT → CONTEXT → CREDENTIALS</span></div></div>;
}
export function About() {
  return <div className="about split"><div><p className="eyebrow">A little about me</p><h1>Hi, I’m<br /><em>Kosy.</em></h1><p className="lead">Developer at <strong>Rise</strong>.</p></div><div className="interest-list"><p className="mono muted">THINGS I’M INTO</p>{['Sports', 'Software', 'Cool engineering things'].map((v,i)=><div key={v}><span className="mono">0{i+1}</span><h2>{v}</h2></div>)}</div></div>;
}
export function History() {
  return <><p className="eyebrow">Content, then and now</p><h1>From a few sources.<br />To <em>almost anything.</em></h1><div className="history-layout"><div><span className="mono muted">25–30 YEARS AGO</span><h2>Concentrated<br />publishing power</h2><p>Governments · Newsrooms · Corporate PR</p></div><span className="large-arrow">→</span><div><span className="mono accent">TODAY</span><div className="source-grid">{['AI models', 'Anyone with a phone', 'Security cameras', 'Editing software', 'Social platforms', 'An accidental screenshot'].map(x=><span key={x}>{x}</span>)}</div></div></div><p className="bottom-note">The number of producers has exploded. So has the content.</p></>;
}
export function Authenticity() {
  return <><p className="eyebrow">The authenticity problem</p><h1>More content.<br /><em>More questions.</em></h1><div className="questions">{['Where did this come from?', 'Has it been altered?', 'Is what it claims actually true?'].map((q,i)=><div key={q}><span className="mono">0{i+1}</span><h2>{q}</h2><span className={i<2?'accent':'muted'}>{i<2?'↗':'?'}</span></div>)}</div></>;
}
export function Journey() {
  return <><p className="eyebrow">Following the journey</p><h1>Before it reached you.</h1><p className="lead muted">One image. Many hands.</p><div className="journey">{journey.map((step,i)=><div className="journey-step" key={step}><span className="journey-node mono">{String(i+1).padStart(2,'0')}</span><h3>{step}</h3>{i<journey.length-1&&<span className="journey-arrow">→</span>}</div>)}</div><div className="journey-ending"><span className="mono accent">provenance()</span><p>The history of how content reached its current form.</p></div><p className="bottom-note">An illustrative journey; credentials record participating steps, not every viewer or share.</p></>;
}
export function Provenance() {
  return <><p className="eyebrow">Defining provenance</p><div className="definition"><span className="mono accent">provenance / noun</span><h1>The facts about<br />the <em>history</em> of<br />digital content.</h1></div><div className="definition-footer"><span className="mono muted">THE ASSET</span><p>Image <span> / </span> Video <span> / </span> Audio <span> / </span> Document</p></div></>;
}
export function Audience() {
  const [revealed,setRevealed]=useState(false);
  return <><div className="title-row"><div><p className="eyebrow">A quick experiment</p><h1>Which one is real?</h1></div><button className="text-button" onClick={()=>setRevealed(!revealed)}>{revealed?'Hide answer':'Reveal answer'} <span>↗</span></button></div><div className="image-pair"><figure><AssetImage src={assets.eagle} alt="Candidate A" label="Image A"/><figcaption><span className="mono">A</span><span>{revealed?'AI-generated eagle':'Look closely.'}</span></figcaption></figure><figure><AssetImage src={assets.tanuki} alt="Candidate B" label="Image B"/><figcaption><span className="mono">B</span><span>{revealed?'Real tanuki photograph':'Trust your eyes?'}</span></figcaption></figure></div><p className="audience-note" aria-live="polite">{revealed?'Staring at pixels to decide what’s real is part of the problem.':'Take a moment. Then cast your vote: A or B.'}</p></>;
}
export function Lifecycle() {
  return <><p className="eyebrow">How Content Credentials work</p><h1>From creation<br />to <em>verification.</em></h1><div className="pipeline">{lifecycle.map((step,i)=><div key={step}><span className="mono accent">{String(i+1).padStart(2,'0')}</span><h3>{step}</h3>{i<7&&<span className="pipeline-arrow">→</span>}</div>)}</div><div className="pipeline-caption"><span>CREATE & SIGN</span><span>DELIVER & INSPECT</span></div></>;
}
export function Terms() {
  return <><p className="eyebrow">The vocabulary</p><h1>Five pieces of the picture.</h1><div className="terms">{terms.map(([term,description],i)=><div key={term}><span className="mono muted">0{i+1}</span><h2>{term}</h2><p>{description}</p></div>)}</div></>;
}
export function Manifest() {
  return <><p className="eyebrow">Inside the structure</p><div className="split manifest-layout"><div><h1>One manifest.<br /><em>Three parts.</em></h1><p className="lead muted">Statements about the asset.<br />References to those statements.<br />A signature over the claim.</p><p className="bottom-note">Content bindings connect the provenance data to the asset.</p><a className="source-link" href={specificationUrl} target="_blank" rel="noreferrer">C2PA specification ↗</a></div><ManifestView /></div></>;
}
export function SignedAsset() {
  return <><p className="eyebrow">An asset and its credentials</p><div className="title-row"><h1>The pixels. <em>The context.</em></h1><span className="mono sample-label">C2PA TEST CERTIFICATE</span></div><div className="asset-layout"><AssetImage src={assets.signedAsset} alt="A real C2PA-signed image, alongside its credentials" label="Signed asset"/><div className="credential-panel"><div className="credential-heading"><span className="cc-mark">cr</span><h2>Content Credentials</h2></div><dl><dt>ASSET</dt><dd>{exampleManifest.asset}</dd><dt>CLAIM GENERATOR</dt><dd>{exampleManifest.generator}</dd><dt>RECORDED ACTIONS</dt><dd className="mono">{exampleManifest.actions.join(' → ')}</dd><dt>SIGNER</dt><dd>{exampleManifest.signer}</dd></dl><p className="validation-note">{exampleManifest.validation}</p></div></div></>;
}
export function ChessPuzzle() {
  return <div className="chess-slide split"><div><p className="eyebrow">A second experiment</p><h1>Your move.</h1><p className="lead muted">Study the position.<br />What would you play?</p><span className="mono accent">Take your time. No spoilers.</span></div><AssetImage src={assets.chessPuzzle} alt="Chess puzzle position, without its solution" label="Chess puzzle"/></div>;
}
export function ChessSolution() {
  return <div className="chess-slide split"><div><p className="eyebrow">The reveal</p><h1>Now, the<br /><em>solution.</em></h1><p className="lead muted">What changed when you<br />had more information?</p></div><AssetImage src={assets.chessSolution} alt="Chess puzzle solution" label="Chess solution"/></div>;
}
export function Closing() {
  return <><p className="eyebrow">The distinction that matters</p><div className="truth-equation"><span>Provenance</span><span className="not-equal">≠</span><span>Truth</span></div><div className="truth-labels"><span>Origin. History. Integrity.</span><span>Accuracy. Context. Judgment.</span></div><blockquote>C2PA doesn’t tell us what to believe.<br /><em>It gives us more evidence about where<br />something came from and what happened to it.</em></blockquote><p className="closing-footer mono">BETTER EVIDENCE. MORE INFORMED JUDGMENT.</p></>;
}

export function Questions() {
  return <div className="qa-slide">
    <p className="eyebrow">Over to you</p>
    <div className="qa-heading"><h1>Time for<br /><em>questions.</em></h1><span className="qa-mark" aria-hidden="true">?</span></div>
    <div className="qa-prompt"><span className="mono muted">A QUESTION TO GET US STARTED</span><blockquote>“Isn’t this what NFTs tried to solve?”</blockquote></div>
  </div>;
}
