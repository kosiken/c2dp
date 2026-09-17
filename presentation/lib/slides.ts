import type { ComponentType } from 'react';
import { Cover, About, History, Authenticity, Journey, Provenance, Audience, Lifecycle, Terms, Manifest, SignedAsset, ChessPuzzle, ChessSolution, Closing, Questions } from '@/components/slides/ContentSlides';
import { DemoThirteen, DemoFourteen, DemoFifteen } from '@/components/slides/DemoSlides';
export type SlideDefinition = { id: string; title: string; section: string; component: ComponentType };
export const slides: SlideDefinition[] = [
  {id:'cover',title:'C2PA and Digital Content Provenance',section:'Engineering / Rise',component:Cover},
  {id:'about',title:'A little about me',section:'01 / Introduction',component:About},
  {id:'history',title:'Content generation and consumption',section:'01 / Introduction',component:History},
  {id:'authenticity',title:'The authenticity problem',section:'01 / Introduction',component:Authenticity},
  {id:'journey',title:'Following the journey',section:'01 / Introduction',component:Journey},
  {id:'provenance',title:'Defining provenance',section:'01 / Introduction',component:Provenance},
  {id:'which-is-real',title:'Which one is real?',section:'02 / Look closer',component:Audience},
  {id:'lifecycle',title:'How Content Credentials work',section:'03 / Inside C2PA',component:Lifecycle},
  {id:'terms',title:'Core C2PA terms',section:'03 / Inside C2PA',component:Terms},
  {id:'manifest',title:'What is in a C2PA Manifest?',section:'03 / Inside C2PA',component:Manifest},
  {id:'signed-asset',title:'An asset and its credentials',section:'03 / Inside C2PA',component:SignedAsset},
  {id:'chess-puzzle',title:'Your move',section:'04 / Context matters',component:ChessPuzzle},
  {id:'chess-solution',title:'The solution',section:'04 / Context matters',component:ChessSolution},
  {id:'demo-13',title:"You're the content creators now",section:'05 / In practice',component:DemoThirteen},
  {id:'demo-14',title:'Watching provenance arrive',section:'05 / In practice',component:DemoFourteen},
  {id:'demo-15',title:'Provenance through transformation',section:'05 / In practice',component:DemoFifteen},
  {id:'closing',title:'Provenance is not truth',section:'06 / A final thought',component:Closing},
  {id:'questions',title:'Time for questions you may have',section:'07 / Discussion',component:Questions},
];
