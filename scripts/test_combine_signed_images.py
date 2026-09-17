"""Integration checks using the real c2patool and disposable image fixtures."""
import hashlib
import json
from pathlib import Path
import tempfile
import unittest

from PIL import Image
from combine_signed_images import ROOT, combine, run_tool


class CompositeTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.tool = str(ROOT / "go-api/bin/c2patool")
        if not Path(cls.tool).exists():
            raise unittest.SkipTest("local c2patool is required")
        cls.temp = tempfile.TemporaryDirectory(prefix="c2dp-composite-test-")
        cls.root = Path(cls.temp.name)
        cls.settings = cls.root / "settings.json"
        cls.settings.write_text("{}")
        cls.definition = cls.root / "manifest.json"
        cls.definition.write_text(json.dumps({"title": "Generated test fixture"}))
        cls.inputs = []
        cls.colors = ["red", "green", "blue", "yellow"]
        for index, color in enumerate(cls.colors):
            source = cls.root / f"source-{index}.png"
            signed = cls.root / f"signed-{index}.png"
            Image.new("RGB", (80 + index * 10, 60), color).save(source)
            run_tool(cls.tool, cls.settings, source, "--create", "digitalArt", "-m", cls.definition, "-o", signed)
            cls.inputs.append(signed)
        cls.hashes = [hashlib.sha256(p.read_bytes()).hexdigest() for p in cls.inputs]

    @classmethod
    def tearDownClass(cls):
        cls.temp.cleanup()

    def test_ratios_and_provenance(self):
        for ratio, fit, dimensions in [("16:9", "cover", (1920, 1080)), ("square", "contain", (1080, 1080))]:
            with self.subTest(ratio=ratio):
                output = self.root / f"composite-{fit}.png"
                report_path, recipe_path = combine(self.inputs, output, ratio, fit, self.tool)
                with Image.open(output) as result:
                    self.assertEqual(result.size, dimensions)
                    for index, color in enumerate(self.colors):
                        x = int(((index % 2) + .5) * dimensions[0] / 2)
                        y = int(((index // 2) + .5) * dimensions[1] / 2)
                        self.assertEqual(result.getpixel((x, y)), Image.new("RGB", (1, 1), color).getpixel((0, 0)))
                report = json.loads(report_path.read_text())
                self.assertEqual(len(report["manifests"]), 5)
                active = report["manifests"][report["active_manifest"]]
                self.assertEqual(len(active["ingredients"]), 4)
                assertions = {a["label"]: a["data"] for a in active["assertions"]}
                actions = assertions["c2pa.actions.v2"]["actions"]
                self.assertEqual(actions[0]["action"], "c2pa.created")
                self.assertEqual(sum(a["action"] == "c2pa.placed" for a in actions), 4)
                for action in actions[1:]:
                    self.assertTrue(action["parameters"]["ingredients"])
                self.assertEqual(assertions["org.c2dp.collage"], json.loads(recipe_path.read_text()))
                with self.assertRaisesRegex(ValueError, "overwrite"):
                    combine(self.inputs, output, ratio, fit, self.tool)
        self.assertEqual(self.hashes, [hashlib.sha256(p.read_bytes()).hexdigest() for p in self.inputs])

    def test_reject_unsigned_and_duplicates(self):
        with self.assertRaises(ValueError):
            combine([self.root / "source-0.png", *self.inputs[1:]], self.root / "unsigned.png", "square", "cover", self.tool)
        with self.assertRaisesRegex(ValueError, "distinct"):
            combine([self.inputs[0]] * 4, self.root / "duplicates.png", "square", "cover", self.tool)
        self.assertFalse((self.root / "unsigned.png").exists())

    def test_reject_modified_image(self):
        # Change image bytes while retaining embedded C2PA metadata. Change pHYs
        # if available, otherwise alter IDAT, and regenerate that chunk's CRC.
        import struct
        import zlib
        data = bytearray(self.inputs[0].read_bytes())
        offset = 8
        while offset < len(data):
            length = struct.unpack(">I", data[offset:offset+4])[0]
            if data[offset+4:offset+8] == b"IDAT":
                data[offset+8+length//2] ^= 1
                crc = zlib.crc32(data[offset+4:offset+8+length])
                data[offset+8+length:offset+12+length] = struct.pack(">I", crc)
                break
            offset += length + 12
        tampered = self.root / "tampered.png"
        tampered.write_bytes(data)
        with self.assertRaises(ValueError):
            combine([tampered, *self.inputs[1:]], self.root / "rejected.png", "square", "cover", self.tool)
        self.assertFalse((self.root / "rejected.png").exists())


if __name__ == "__main__":
    unittest.main()
