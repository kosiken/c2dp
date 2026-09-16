# C2PA and Digital Content Provenance

Standalone Next.js / TypeScript / Tailwind presentation. Cover + 17 numbered slides.

```sh
cd presentation
npm install
npm run dev
```

Open http://localhost:3001. Production: `npm run build && npm start`.

- Right / Space: next; Left: previous; F: enter fullscreen; Escape: native browser exit.
- Visible previous/next and fullscreen controls support mouse and keyboard focus.
- Slide URLs use hashes, e.g. `/#manifest` or `/#chess-puzzle`.
- A fixed 1600×900 stage scales to fit the viewport, retaining composition without scrolling.
- Reduced-motion preferences disable transitions.

## Content and assets

`lib/content.ts` holds shared data and image paths. Supply originals under
`public/images/` as described in its README. Images were not present in the source
workspace; missing slots are labeled, with no fabricated replacement imagery.

Slides 10–16 follow the initial task instructions. Slide 17 adds the outline’s
Q&A prompt about NFTs. Demo titles are provisional (`Demo 01`–`Demo 03`) pending the remaining
outline. Slide 10 data is illustrative and never presented as verified credentials.
Technical reference: https://spec.c2pa.org/specifications/specifications/2.3/specs/C2PA_Specification

## Component boundaries

- `components/Presentation.tsx`: sizing, hash navigation, keyboard, fullscreen, progress.
- `lib/slides.ts`: ordered registry of independent slide components.
- `components/slides/ContentSlides.tsx`: authored slide layouts.
- `components/slides/DemoSlides.tsx`: three isolated placeholders; no backend behavior.
- `components/ManifestView.tsx`: reusable structural diagram.
- `components/AssetImage.tsx`: images, including future upload/object URLs.

Only the active slide mounts. Future demos can own HTTP calls, EventSource streams,
upload state, and manifest views; close streams and abort requests on unmount.
Persist state outside an individual slide only if a future demo needs it across
navigation. There is no backend dependency or slideshow library today.

Validation: `npm run typecheck` and `npm run build`.
