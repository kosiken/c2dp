#!/usr/bin/env python3
"""Make a signed 2×2 collage while retaining the four source manifests."""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import random
import shutil
import subprocess
import sys
import tempfile
import warnings

from PIL import Image, ImageOps

ROOT = Path(__file__).resolve().parents[1]
RATIOS = {"16:9": (1920, 1080), "square": (1080, 1080), "1:1": (1080, 1080)}
UPLOAD_DIR = ROOT / "go-api/uploads"
IMAGE_SUFFIXES = {".jpg", ".jpeg", ".png", ".webp", ".gif"}


def pick_random_uploads(count: int = 4) -> list[Path]:
    candidates = [p for p in UPLOAD_DIR.iterdir() if p.is_file() and p.suffix.lower() in IMAGE_SUFFIXES]
    if len(candidates) < count:
        raise ValueError(f"{UPLOAD_DIR} has only {len(candidates)} image(s); need at least {count}")
    return random.sample(candidates, count)


def run_tool(tool: str, settings: Path, *args: str | Path) -> str:
    # Use public test credentials consistently with the Go upload signer.
    env = {key: value for key, value in os.environ.items() if not key.startswith("C2PA")}
    result = subprocess.run(
        [tool, *map(str, args), "--settings", str(settings)],
        capture_output=True, text=True, env=env, timeout=120,
    )
    if result.returncode:
        raise ValueError(f"c2patool failed: {(result.stderr or result.stdout).strip()}")
    return result.stdout


def verify(report: dict, label: str) -> None:
    active = report.get("active_manifest")
    if not active or active not in report.get("manifests", {}):
        raise ValueError(f"{label}: no embedded C2PA manifest")
    validation = report.get("validation_results", {}).get("activeManifest", {})
    success = {item["code"] for item in validation.get("success", [])}
    if "claimSignature.validated" not in success or not success.intersection(
        {"assertion.dataHash.match", "assertion.bmffHash.match", "assertion.boxesHash.match"}
    ):
        raise ValueError(f"{label}: signature or content binding did not validate")
    # Untrusted development certificates are expected; integrity failures are not.
    failures = list(validation.get("failure", []))
    for delta in report.get("validation_results", {}).get("ingredientDeltas", []):
        failures.extend(delta.get("validationDeltas", {}).get("failure", []))
    failures.extend(report.get("validation_status", []))
    invalid = {item["code"] for item in failures if item["code"] != "signingCredential.untrusted"}
    if invalid:
        raise ValueError(f"{label}: validation failed: {', '.join(sorted(invalid))}")


def combine(inputs: list[Path], output: Path, ratio: str, fit: str, tool: str) -> tuple[Path, Path]:
    inputs = [path.resolve(strict=True) for path in inputs]
    if len(inputs) != 4:
        raise ValueError("Exactly four signed input images are required")
    if len(set(inputs)) != 4:
        raise ValueError("Provide four distinct source files")
    output = output.resolve()
    if output.suffix.lower() != ".png":
        raise ValueError("Output must end in .png")
    if output in inputs:
        raise ValueError("Output must not replace an input")
    report_path = output.with_suffix(".manifest.json")
    recipe_path = output.with_suffix(".recipe.json")
    for target in (output, report_path, recipe_path):
        if target.exists():
            raise ValueError(f"Refusing to overwrite {target}")
    output.parent.mkdir(parents=True, exist_ok=True)
    width, height = RATIOS[ratio]
    cell = (width // 2, height // 2)
    with tempfile.TemporaryDirectory(prefix="c2dp-collage-", dir=output.parent) as temporary:
        work = Path(temporary)
        settings = work / "settings.json"
        settings.write_text("{}")
        ingredients, ingredient_paths, placements, source_manifests, source_digests = [], [], [], set(), set()
        canvas = Image.new("RGB", (width, height), "#111512")
        # Snapshot inputs so validation, pixels, and provenance use identical bytes.
        for index, original in enumerate(inputs):
            snapshot = work / f"source-{index}{original.suffix.lower()}"
            shutil.copyfile(original, snapshot)
            digest = hashlib.sha256(snapshot.read_bytes()).hexdigest()
            if digest in source_digests:
                raise ValueError("Provide four distinct images, not copies of the same file")
            source_digests.add(digest)
            report = json.loads(run_tool(tool, settings, snapshot))
            verify(report, original.name)
            source_manifests.update(report["manifests"])
            ingredient_dir = work / f"ingredient-{index}"
            run_tool(tool, settings, snapshot, "--ingredient", "--output", ingredient_dir)
            ingredient = json.loads((ingredient_dir / "ingredient.json").read_text())
            ingredient["title"] = original.name
            ingredient["relationship"] = "componentOf"
            ingredient["label"] = f"source-{index + 1}"
            # c2patool resolves resource identifiers relative to this ingredient.json's
            # own directory, so leave the paths it exported as-is (untouched, relative).
            (ingredient_dir / "ingredient.json").write_text(json.dumps(ingredient))
            ingredient_paths.append(str(ingredient_dir / "ingredient.json"))
            ingredients.append(ingredient)
            with warnings.catch_warnings():
                warnings.simplefilter("error", Image.DecompressionBombWarning)
                with Image.open(snapshot) as image:
                    if getattr(image, "is_animated", False):
                        raise ValueError(f"{original.name}: animated images are not supported")
                    oriented = ImageOps.exif_transpose(image).convert("RGBA")
                    # Flatten transparency before resizing onto the collage background.
                    flattened = Image.new("RGBA", oriented.size, "#111512")
                    flattened.alpha_composite(oriented)
                    rgb = flattened.convert("RGB")
                    if fit == "cover":
                        tile = ImageOps.fit(rgb, cell, method=Image.Resampling.LANCZOS)
                    else:
                        tile = ImageOps.pad(rgb, cell, method=Image.Resampling.LANCZOS, color="#111512")
                    x, y = (index % 2) * cell[0], (index // 2) * cell[1]
                    canvas.paste(tile, (x, y))
                    placements.append({"source": original.name, "sha256": digest,
                                       "ingredient": ingredient["label"], "x": x, "y": y,
                                       "width": cell[0], "height": cell[1],
                                       "oriented_source_size": list(rgb.size)})
        unsigned = work / "canvas.png"
        canvas.save(unsigned)
        actions = [{"action": "c2pa.placed", "parameters": {"ingredientIds": [item["label"]]}}
                   for item in ingredients]
        recipe = {"layout": "2x2", "width": width, "height": height, "fit": fit,
                  "orientation": "EXIF orientation applied", "background": "#111512",
                  "resampling": "Lanczos", "placements": placements}
        manifest = {
            "title": output.name,
            "base_path": str(work),
            "claim_generator_info": [{"name": "C2DP collage demo", "version": "1.0.0"}],
            "ingredient_paths": ingredient_paths,
            "assertions": [
                {"label": "c2pa.actions.v2", "data": {"actions": actions}, "created": True},
                {"label": "org.c2dp.collage", "data": recipe, "created": True},
            ],
        }
        definition = work / "manifest.json"
        definition.write_text(json.dumps(manifest, indent=2))
        signed = work / "signed.png"
        # A new empty canvas with four placed components, rather than claiming
        # the newly assembled pixels are an original camera capture.
        run_tool(tool, settings, unsigned, "--create", "empty", "--manifest", definition, "--output", signed)
        result = json.loads(run_tool(tool, settings, signed))
        verify(result, "Composite")
        active = result["manifests"][result["active_manifest"]]
        components = [i for i in active.get("ingredients", []) if i.get("relationship") == "componentOf"]
        if len(components) != 4 or not source_manifests.issubset(result["manifests"]):
            raise ValueError("Composite did not preserve all four ingredients and their manifests")
        # Publish only after validating the complete signed output. Exclusive
        # creation prevents a concurrent invocation from overwriting an artifact.
        published = []
        try:
            with output.open("xb") as destination:
                published.append(output)
                with signed.open("rb") as source:
                    shutil.copyfileobj(source, destination)
            for path, data in ((report_path, result), (recipe_path, recipe)):
                with path.open("x") as destination:
                    published.append(path)
                    json.dump(data, destination, indent=2)
                    destination.write("\n")
        except Exception:
            for path in published:
                path.unlink(missing_ok=True)
            raise
    return report_path, recipe_path


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("images", type=Path, nargs="*",
                        help="Top-left, top-right, bottom-left, bottom-right. "
                             f"Omit to pick 4 random images from {UPLOAD_DIR}")
    parser.add_argument("-o", "--output", type=Path, required=True, help="Signed output PNG (must not already exist)")
    parser.add_argument("--ratio", choices=RATIOS, default="16:9")
    parser.add_argument("--fit", choices=("cover", "contain"), default="cover",
                        help="cover: center crop to fill; contain: retain whole image with padding")
    parser.add_argument("--c2patool", default=os.environ.get("C2PATOOL_PATH", str(ROOT / "go-api/bin/c2patool")))
    args = parser.parse_args()
    if args.images and len(args.images) != 4:
        parser.error("provide zero images (random) or exactly four")
    try:
        tool = shutil.which(args.c2patool)
        if not tool:
            raise ValueError("c2patool not found; pass --c2patool /path/to/c2patool")
        images = args.images or pick_random_uploads()
        if not args.images:
            print(f"No images given; picked randomly from {UPLOAD_DIR}:")
            for image in images:
                print(f"  {image.name}")
        report, recipe = combine(images, args.output, args.ratio, args.fit, str(Path(tool).resolve()))
    except (OSError, ValueError, subprocess.TimeoutExpired, Image.DecompressionBombError, Image.DecompressionBombWarning) as error:
        print(f"Error: {error}", file=sys.stderr)
        return 1
    print(f"Signed image: {args.output.resolve()}\nManifest report: {report}\nComposition recipe: {recipe}")
    print("Verified signature, content binding, and four source ingredients. Signer uses an untrusted public test certificate.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
