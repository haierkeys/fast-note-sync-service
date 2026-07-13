"""Generate fnOS app icons for fast-note-sync-service.

Renders a branded rounded-square icon with the sync/wifi arc glyph
(matching the project's original icon.svg) at the sizes fnOS requires:
- ICON.PNG and ICON_256.PNG  -> package root (256x256)
- app/ui/images/icon_64.png  -> 64x64
- app/ui/images/icon_256.png -> 256x256
"""
import os
from PIL import Image, ImageDraw

# Package root is the parent of this scripts/ directory.
OUT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

BG = (91, 108, 255, 255)      # indigo brand
BG_DARK = (64, 80, 230, 255)
STROKE = (255, 255, 255, 255)


def rounded_gradient(size: int) -> Image.Image:
    """A rounded square with a vertical indigo gradient."""
    img = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    px = img.load()
    for y in range(size):
        t = y / max(size - 1, 1)
        r = int(BG[0] * (1 - t) + BG_DARK[0] * t)
        g = int(BG[1] * (1 - t) + BG_DARK[1] * t)
        b = int(BG[2] * (1 - t) + BG_DARK[2] * t)
        for x in range(size):
            px[x, y] = (r, g, b, 255)

    # Round the corners with a mask.
    radius = int(size * 0.22)
    mask = Image.new("L", (size, size), 0)
    md = ImageDraw.Draw(mask)
    md.rounded_rectangle((0, 0, size - 1, size - 1), radius=radius, fill=255)
    rounded = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    rounded.paste(img, (0, 0), mask)
    return rounded


def draw_glyph(canvas: Image.Image) -> None:
    """Draw the wifi/sync arc glyph (two arcs + dot) centered on the canvas."""
    w, h = canvas.size
    draw = ImageDraw.Draw(canvas)

    # Geometry mirrors the source SVG (viewBox 24x24, translated -0.5 y).
    # Arc 1: M5 13 a10 10 0 0 1 14 0   -> big arc
    # Arc 2: M8.5 16.5 a5 5 0 0 1 7 0  -> small arc
    # Dot:   x12 y20 r~0.6
    def to_px(x_svg: float, y_svg: float):
        # Map SVG 24x24 space into a centered box occupying ~74% of canvas.
        box = min(w, h) * 0.74
        ox = (w - box) / 2
        oy = (h - box) / 2
        return (ox + x_svg / 24 * box, oy + y_svg / 24 * box)

    lw = max(2, int(min(w, h) * 0.045))

    # Big arc: from (5,13) sweeping to (19,13) via a 10-radius bulge upward.
    p1 = to_px(5, 13)
    p2 = to_px(19, 13)
    # Bounding box of a circle r=10 centered at (12,13): (2,3)-(22,23).
    bb1 = (to_px(2, 3)[0], to_px(2, 3)[1], to_px(22, 23)[0], to_px(22, 23)[1])
    draw.arc(bb1, 200, 340, fill=STROKE, width=lw)

    # Small arc: from (8.5,16.5) to (15.5,16.5), r=5, center (12,16.5).
    bb2 = (to_px(7, 11.5)[0], to_px(7, 11.5)[1], to_px(17, 21.5)[0], to_px(17, 21.5)[1])
    draw.arc(bb2, 200, 340, fill=STROKE, width=lw)

    # Dot near (12, 20).
    dot_c = to_px(12, 20)
    r = max(1, int(min(w, h) * 0.022))
    draw.ellipse((dot_c[0] - r, dot_c[1] - r, dot_c[0] + r, dot_c[1] + r), fill=STROKE)


def make(size: int, path: str) -> None:
    canvas = rounded_gradient(size)
    draw_glyph(canvas)
    canvas.save(path, "PNG")
    print("wrote", path, size)


if __name__ == "__main__":
    root = OUT
    ui = os.path.join(root, "app", "ui", "images")
    os.makedirs(ui, exist_ok=True)
    make(256, os.path.join(root, "ICON.PNG"))
    make(256, os.path.join(root, "ICON_256.PNG"))
    make(64, os.path.join(ui, "icon_64.png"))
    make(256, os.path.join(ui, "icon_256.png"))
