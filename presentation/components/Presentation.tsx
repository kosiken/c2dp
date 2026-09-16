'use client';
import { useCallback, useEffect, useState } from 'react';
import { slides } from '@/lib/slides';

export function Presentation() {
  const [index,setIndex]=useState(0);
  const [scale,setScale]=useState(1);
  const [fullscreen,setFullscreen]=useState(false);
  const [notice,setNotice]=useState('');
  const go=useCallback((next: number)=>{
    const bounded=Math.max(0,Math.min(slides.length-1,next));
    setIndex(bounded);
    window.history.replaceState(null,'',`#${slides[bounded].id}`);
  },[]);
  const enterFullscreen=useCallback(async()=>{
    try {
      if (!document.fullscreenElement) await document.documentElement.requestFullscreen();
    } catch { setNotice('Fullscreen unavailable. Use your browser’s fullscreen command.'); }
  },[]);
  useEffect(()=>{
    const resize=()=>setScale(Math.min(window.innerWidth/1600,window.innerHeight/900));
    const readHash=()=>{ const found=slides.findIndex(s=>`#${s.id}`===window.location.hash); setIndex(found<0?0:found); };
    const syncFullscreen=()=>setFullscreen(Boolean(document.fullscreenElement));
    resize(); readHash();
    window.addEventListener('resize',resize);
    window.addEventListener('hashchange',readHash);
    document.addEventListener('fullscreenchange',syncFullscreen);
    return ()=>{window.removeEventListener('resize',resize);window.removeEventListener('hashchange',readHash);document.removeEventListener('fullscreenchange',syncFullscreen);};
  },[]);
  useEffect(()=>{
    const onKey=(event: KeyboardEvent)=>{
      const target=event.target as HTMLElement;
      if(event.altKey||event.ctrlKey||event.metaKey||target.closest('input,textarea,select,[contenteditable]:not([contenteditable="false"])')) return;
      // Preserve Space activation on focused controls while keeping arrow navigation available.
      if(event.code==='Space'&&target.closest('button,a,[role="button"]')) return;
      if(event.key==='ArrowRight'||event.code==='Space') {event.preventDefault();if(!event.repeat)go(index+1);}
      if(event.key==='ArrowLeft') {event.preventDefault();if(!event.repeat)go(index-1);}
      if(event.key.toLowerCase()==='f') {event.preventDefault();void enterFullscreen();}
      // Escape is deliberately left to the browser.
    };
    window.addEventListener('keydown',onKey);
    return ()=>window.removeEventListener('keydown',onKey);
  },[index,go,enterFullscreen]);
  const slide=slides[index];
  const Slide=slide.component;
  return <main className="viewport" aria-label="C2PA presentation">
    <div className="presentation-stage" style={{transform:`translate(-50%, -50%) scale(${scale})`}}>
      <header className="deck-header"><a href="#cover" aria-label="Go to cover" className="brand">Rise<span className="brand-dot">.</span><span className="header-divider"/>ENGINEERING</a><span className="mono">{slide.section}</span><span className="mono header-topic">C2PA / DIGITAL PROVENANCE</span></header>
      <section key={slide.id} className={`slide slide-${slide.id}`} aria-label={slide.title}><Slide /></section>
      <footer className="deck-footer"><div className="keyboard-hints mono"><span>← → <span className="muted">Navigate</span></span><span>SPACE <span className="muted">Next</span></span><button onClick={enterFullscreen} title="Enter fullscreen (F)">F <span className="muted">{fullscreen?'Fullscreen active':'Fullscreen'}</span></button></div><div className="footer-navigation"><button aria-label="Previous slide" disabled={index===0} onClick={()=>go(index-1)}>←</button><span className="mono slide-count" aria-live="polite">{index===0?'COVER':String(index).padStart(2,'0')} <span className="muted">/ {slides.length-1}</span></span><button aria-label="Next slide" disabled={index===slides.length-1} onClick={()=>go(index+1)}>→</button></div></footer>
      <div className="progress-track" role="progressbar" aria-label="Presentation progress" aria-valuemin={0} aria-valuemax={slides.length-1} aria-valuenow={index}><div style={{width:`${index/(slides.length-1)*100}%`}}/></div>
      {notice&&<button className="fullscreen-notice" onClick={()=>setNotice('')} role="status">{notice}</button>}
    </div>
  </main>;
}
