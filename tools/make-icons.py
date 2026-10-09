"""Builds every icon file the app needs from one picture.

    python tools/make-icons.py

The source is assets/mvd-icon-1024.png when it exists (a 1024 x 1024 picture gives the
sharpest result), else assets/mvd-icon.png (128 x 128: sizes up to 128 are exact, larger
ones are scaled up and a little soft).

It writes:

  assets/icons/mvd.ico               Windows: 16 to 256, embedded in the .exe by the .syso
  assets/icons/mvd.icns              macOS: the bundle and the disk image
  assets/icons/mvd-<size>.png        Linux and anywhere else a PNG is wanted
  apps/mvd/appwindow/window_icon.png the window's own icon (256)
  apps/mvd/macbundle/mvd.icns        the copy the app embeds to build its own MVD.app
  apps/mvd/install/linux_icon.png    the copy the app embeds for the Linux menu entry (256)

The Windows resource file is made from the .ico with:

    go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico assets/icons/mvd.ico -o apps/mvd/rsrc_windows_amd64.syso

Needs Pillow (pip install pillow).
"""

import os
import shutil
import sys

from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ICON_SIZES = [16, 24, 32, 48, 64, 128, 256]
PNG_SIZES = [16, 32, 48, 64, 128, 256, 512, 1024]


def source() -> Image.Image:
    for name in ("assets/mvd-icon-1024.png", "assets/mvd-icon.png"):
        path = os.path.join(ROOT, name)
        if os.path.exists(path):
            print("source:", name)
            return Image.open(path).convert("RGBA")
    sys.exit("no source picture: put assets/mvd-icon-1024.png (or mvd-icon.png) in place")


def scaled(image: Image.Image, size: int) -> Image.Image:
    if image.width == size:
        return image.copy()
    return image.resize((size, size), Image.LANCZOS)


def main() -> None:
    base = source()
    if base.width != base.height:
        sys.exit(f"the picture must be square, it is {base.width} x {base.height}")

    out = os.path.join(ROOT, "assets", "icons")
    os.makedirs(out, exist_ok=True)

    for size in PNG_SIZES:
        scaled(base, size).save(os.path.join(out, f"mvd-{size}.png"))

    # Windows: one file with every size, each frame scaled from the source on its own.
    frames = [scaled(base, s) for s in ICON_SIZES]
    frames[-1].save(os.path.join(out, "mvd.ico"), format="ICO", sizes=[(s, s) for s in ICON_SIZES], append_images=frames[:-1])

    # macOS: the sizes an .icns holds.
    scaled(base, 1024).save(os.path.join(out, "mvd.icns"), format="ICNS")

    shutil.copyfile(os.path.join(out, "mvd-256.png"), os.path.join(ROOT, "apps", "mvd", "appwindow", "window_icon.png"))
    shutil.copyfile(os.path.join(out, "mvd.icns"), os.path.join(ROOT, "apps", "mvd", "macbundle", "mvd.icns"))
    shutil.copyfile(os.path.join(out, "mvd-256.png"), os.path.join(ROOT, "apps", "mvd", "install", "linux_icon.png"))
    print("wrote", os.path.relpath(out, ROOT))


if __name__ == "__main__":
    main()
