# C2PA and Digital Content Provenance

Standalone Next.js / TypeScript / Tailwind presentation. Cover + 17 numbered slides.

```sh
cd presentation
npm install
npm run dev
```

Open http://localhost:3001. Production: `npm run build && npm start`.

Set `NEXT_NGROK_WEB_APP_URL` to the audience-facing upload site's ngrok URL before
`npm run dev` / `npm run build` / `npm run typecheck`; each runs `scripts/generate-qr.mjs`
first (via `pre*` npm hooks), which renders it to a QR code baked into `lib/uploadQr.ts`
(gitignored, regenerated every run). Leave it unset to see Slide Thirteen's
"awaiting upload link" placeholder instead.

Set `NEXT_PUBLIC_API_URL` (see `.env.example`; defaults to `http://localhost:8080`) to
where go-api runs — normally the same machine, so the default is enough. Slide 14 fetches
recent posts from it and opens `/api/v1/posts/stream` (SSE) to show new uploads live, with
their C2PA assertions and signer, as they arrive. Slide 15's "Generate Combination" button
calls `POST /api/v1/posts/combine`, which runs `scripts/combine_signed_images.py` server-side
against the four most recently uploaded posts and returns the composite's URL and manifest.

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
Q&A prompt about NFTs. Slide 10 shows `public/images/signed-asset.png`, actually
signed with c2patool's public development certificate (same manifest shape go-api
uses); `lib/content.ts`'s `exampleManifest` is read from that image's own manifest,
not fabricated. Slide 13 (`DemoThirteen`) shows the audience-upload QR code and
URL; Slide 14 (`DemoFourteen`) streams those uploads live from go-api; Slide 15
(`DemoFifteen`) combines the four latest uploads into one new signed composite on demand.
Technical reference: https://spec.c2pa.org/specifications/specifications/2.3/specs/C2PA_Specification

## Component boundaries

- `components/Presentation.tsx`: sizing, hash navigation, keyboard, fullscreen, progress.
- `lib/slides.ts`: ordered registry of independent slide components.
- `components/slides/ContentSlides.tsx`: authored slide layouts.
- `components/slides/DemoSlides.tsx`: three demo slides, each independent.
- `lib/liveFeed.ts`: fetch + SSE client for go-api posts, used by Slide 14.
- `lib/combine.ts`: calls go-api's `/api/v1/posts/combine`, used by Slide 15.
- `components/ManifestView.tsx`: reusable structural diagram.
- `components/AssetImage.tsx`: images, including future upload/object URLs.

Only the active slide mounts. Future demos can own HTTP calls, EventSource streams,
upload state, and manifest views; close streams and abort requests on unmount.
Persist state outside an individual slide only if a future demo needs it across
navigation. There is no backend dependency or slideshow library today.

Validation: `npm run typecheck` and `npm run build`.
