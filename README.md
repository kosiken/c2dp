# C2PA and Digital Content Provenance

A demo built for a talk on C2PA / Content Credentials. Audience members upload
photos from their phones, each is signed with a real (public test) C2PA
certificate, and the presentation shows them arriving live and lets you fold
four of them into one signed composite.

## Pieces

- **`go-api/`** — Go/Echo/Postgres backend. Signs every upload with c2patool,
  stores its manifest, and serves a live feed plus the four-image combine.
- **`web/`** — Next.js upload site the audience's phones hit (via the QR code
  on Slide 13). Talks to `go-api` directly.
- **`presentation/`** — the Next.js slide deck itself. Slides 13–15 talk to
  `go-api` too (QR/URL, live feed, combine button).
- **`scripts/`** — `combine_signed_images.py`, invoked by `go-api`'s
  `/api/v1/posts/combine` endpoint (also runnable standalone).

Each has its own README with setup details; this one is about running them
together.

## Running the full demo

1. **Postgres + go-api**, on the host machine directly (not
   `docker compose up --build` — see the gotcha below):
   ```sh
   cd go-api
   docker compose up -d postgres
   cp .env.example .env
   go run ./cmd/api
   ```
2. **web**, the audience-facing upload app:
   ```sh
   cd web
   cp .env.example .env.local
   npm install && npm run dev
   ```
   Runs on `http://localhost:3000`.
3. **Tunnel `web` with ngrok** so phones can reach it:
   ```sh
   ngrok http 3000
   ```
   Copy the `https://...ngrok...` URL it prints.
4. **presentation**, with that URL set:
   ```sh
   cd presentation
   NEXT_NGROK_WEB_APP_URL=<ngrok url from step 3> npm install && npm run dev
   ```
   Runs on `http://localhost:3001`. (`presentation` and `go-api` talk over
   `localhost:8080` directly — no tunnel needed for that leg.)

During the talk: Slide 13 shows the QR code / URL from step 3 → audience
scans it, uploads via `web` → `go-api` signs and stores the post → Slide 14
shows it arriving live → Slide 15 combines the four latest into one signed
composite on demand.

## Gotcha: combine only works with go-api on the host

`POST /api/v1/posts/combine` shells out to `scripts/combine_signed_images.py`,
which needs Python 3 + Pillow (`python3 -m pip install -r scripts/requirements.txt`)
+ c2patool. The `go-api` Docker image doesn't include any of those (it only
builds the Go binary), so combine will fail if you run `go-api` via
`docker compose up --build` instead of `go run`.

## Signing

All uploads and composites are signed with c2patool's **public development
test certificate** — signatures and content bindings verify, but the
certificate itself is untrusted. This demonstrates the mechanism, not
uploader identity or camera capture.
