// Asset filenames are centralized so supplied files can be mapped without editing slides.
export const assets = {
  eagle: '/images/ai-eagle.jpg',
  tanuki: '/images/real-tanuki.jpg',
  chessPuzzle: '/images/chess-puzzle.png',
  chessSolution: '/images/chess-solution.png',
  signedAsset: '/images/signed-asset.png',
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
// Real data read from public/images/signed-asset.png's own embedded manifest
// (signed with c2patool's public development certificate, same as go-api uploads).
export const exampleManifest = {
  asset: 'ambquinn-lion-8096155_640.png',
  generator: 'C2DP upload demo / 1.0.0',
  actions: ['c2pa.opened'],
  signer: 'C2PA Test Signing Cert',
  validation: 'Signature valid — untrusted development certificate',
};
export const specificationUrl = 'https://spec.c2pa.org/specifications/specifications/2.3/specs/C2PA_Specification';
