// Asset filenames are centralized so supplied files can be mapped without editing slides.
export const assets = {
  eagle: '/images/ai-eagle.jpg',
  tanuki: '/images/real-tanuki.jpg',
  chessPuzzle: '/images/chess-puzzle.png',
  chessSolution: '/images/chess-solution.png',
  signedAsset: '/images/signed-asset.jpg',
};
export const journey = ['Camera', 'Photographer', 'Photo Editing', 'News Editor', 'News Website', 'X', 'Viewer'];
export const lifecycle = ['Content generation', 'Assertion generation', 'Manifest generation', 'Claim signing', 'Embedding or linking', 'Distribution', 'Verification', 'Presentation'];
export const terms = [
  ['Actor', 'An entity involved in the lifecycle of an asset.'],
  ['Asset', 'The digital content itself.'],
  ['Assertion', 'A statement about an asset or its provenance.'],
  ['Claim', 'A structure that references assertions and is digitally signed.'],
  ['Manifest', 'Assertions, a claim, and its signature, collected together.'],
];
// Illustrative UI data only: this is neither a real manifest nor a validation result.
export const exampleManifest = {
  asset: 'field-study.jpg',
  generator: 'Example camera / 1.0',
  actions: ['c2pa.created', 'c2pa.edited'],
  signer: 'Example publisher',
  validation: 'Not performed — illustrative data',
};
export const specificationUrl = 'https://spec.c2pa.org/specifications/specifications/2.3/specs/C2PA_Specification';
