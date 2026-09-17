# Combine four C2PA-signed images

From the repository root:

```sh
python3 -m pip install -r scripts/requirements.txt
python3 scripts/combine_signed_images.py first.png second.png third.png fourth.png \
  --ratio 16:9 --output c2pa-demo/composite-wide.png
python3 scripts/combine_signed_images.py first.png second.png third.png fourth.png \
  --ratio square --fit contain --output c2pa-demo/composite-square.png
```

Requires Python 3.10+, Pillow, and c2patool (tested with 0.27.22). Defaults to the
workspace's `go-api/bin/c2patool`; override with `--c2patool` or `C2PATOOL_PATH`.

Omit the four images to pick four random ones from `go-api/uploads` instead
(fails if fewer than four are present there).

- Layout: 2×2, in argument order: top-left, top-right, bottom-left, bottom-right.
- 16:9 outputs 1920×1080; square (or 1:1) outputs 1080×1080.
- `cover` (default) center-crops to fill each cell. `contain` adds dark padding.
- Applies EXIF orientation; flattens transparency against the dark background.
- Requires four different static, signed images with valid signatures and content
  bindings. Public test certificates are accepted despite being untrusted.
  Unsigned, tampered, duplicate, and animated inputs are rejected.
- Original files are preserved. Existing output files are never overwritten.
- Signs a new canvas with the public test certificate, with four `componentOf`
  ingredients and `c2pa.placed` actions. All source manifests are preserved.
- Records the actual composition (layout, fit, source hashes, and placement) in
  a custom signed assertion. This does not attest to original camera capture.
- Outputs the signed PNG, an extracted `.manifest.json` report, and a portable
  `.recipe.json` record. The C2PA manifest is embedded in the PNG itself; JSON
  reports are for inspection and are not independently signed sidecar manifests.
- Verifies the completed result before publishing it. No API server is required.

Inspect:

```sh
go-api/bin/c2patool c2pa-demo/composite-wide.png --tree
go-api/bin/c2patool c2pa-demo/composite-wide.png
```

Reference: https://opensource.contentauthenticity.org/docs/manifest/writing/ingredients/
