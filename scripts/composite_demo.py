#!/usr/bin/env python3
"""
scripts/composite_demo.py — Post-processes the VHS demo recording to smoothly
reveal the GitHub Primer Web UI in a macOS browser window at the end of the walkthrough.
"""

import os
import subprocess
import shutil
from PIL import Image, ImageDraw, ImageFilter, ImageFont

WIDTH = 1380
HEIGHT = 820
FPS = 25

ROOT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
ASSETS_DIR = os.path.join(ROOT_DIR, "assets")
TMP_FRAMES_DIR = "/tmp/demo_frames"

shutil.rmtree(TMP_FRAMES_DIR, ignore_errors=True)
os.makedirs(TMP_FRAMES_DIR, exist_ok=True)

# 1. Create the Browser Window Component
def build_browser_window():
    web_img_path = os.path.join(ASSETS_DIR, "webui_pr_actual.png")
    if not os.path.exists(web_img_path):
        raise FileNotFoundError(f"Missing {web_img_path}")
    web_img = Image.open(web_img_path).convert("RGBA")

    WIN_W = 1180
    WIN_H = 680
    HEADER_H = 42

    BODY_W = WIN_W
    BODY_H = WIN_H - HEADER_H
    web_resized = web_img.resize((BODY_W, BODY_H), Image.Resampling.LANCZOS)

    window = Image.new("RGBA", (WIN_W, WIN_H), (0, 0, 0, 0))
    mask = Image.new("L", (WIN_W, WIN_H), 0)
    mask_draw = ImageDraw.Draw(mask)
    mask_draw.rounded_rectangle([0, 0, WIN_W - 1, WIN_H - 1], radius=12, fill=255)

    win_draw = ImageDraw.Draw(window)
    win_draw.rectangle([0, 0, WIN_W, HEADER_H], fill=(30, 32, 40, 255))
    window.paste(web_resized, (0, HEADER_H))
    window.putalpha(mask)

    header_draw = ImageDraw.Draw(window)
    header_draw.line([(0, HEADER_H), (WIN_W, HEADER_H)], fill=(60, 64, 75, 200), width=1)

    # Traffic lights
    traffic_y = 16
    header_draw.ellipse([18, traffic_y - 6, 30, traffic_y + 6], fill=(255, 95, 87, 255))
    header_draw.ellipse([38, traffic_y - 6, 50, traffic_y + 6], fill=(254, 188, 46, 255))
    header_draw.ellipse([58, traffic_y - 6, 70, traffic_y + 6], fill=(40, 201, 64, 255))

    try:
        font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 11)
        font_bold = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 11)
    except Exception:
        font = ImageFont.load_default()
        font_bold = font

    ADDR_W = 540
    ADDR_H = 26
    ADDR_X = (WIN_W - ADDR_W) // 2
    ADDR_Y = 8

    header_draw.rounded_rectangle(
        [ADDR_X, ADDR_Y, ADDR_X + ADDR_W, ADDR_Y + ADDR_H],
        radius=6,
        fill=(20, 22, 28, 255),
        outline=(60, 64, 75, 255),
        width=1
    )

    header_draw.text((ADDR_X + 16, ADDR_Y + 7), "🔒", fill=(160, 165, 180, 255), font=font)
    header_draw.text((ADDR_X + 38, ADDR_Y + 7), "http://127.0.0.1:8080/acme/demo/pulls", fill=(210, 215, 230, 255), font=font)
    header_draw.text((ADDR_X + ADDR_W - 130, ADDR_Y + 7), "exist.io v0.5.0", fill=(100, 110, 130, 255), font=font)

    border_overlay = Image.new("RGBA", (WIN_W, WIN_H), (0, 0, 0, 0))
    border_draw = ImageDraw.Draw(border_overlay)
    border_draw.rounded_rectangle([0, 0, WIN_W - 1, WIN_H - 1], radius=12, outline=(255, 255, 255, 50), width=1)
    window = Image.alpha_composite(window, border_overlay)

    return window

def main():
    print("Building composite demo with Web UI reveal...")
    browser_win = build_browser_window()
    base_last_frame = Image.open("/tmp/last_frame.png").convert("RGBA")

    # Final browser positioned on top
    WIN_W, WIN_H = browser_win.size
    POS_X = (WIDTH - WIN_W) // 2
    POS_Y = 28 + ((HEIGHT - 28) - WIN_H) // 2

    # Pre-render shadow
    SHADOW_BLUR = 25
    shadow = Image.new("RGBA", (WIN_W + SHADOW_BLUR * 2, WIN_H + SHADOW_BLUR * 2), (0, 0, 0, 0))
    s_draw = ImageDraw.Draw(shadow)
    s_draw.rounded_rectangle(
        [SHADOW_BLUR, SHADOW_BLUR + 6, WIN_W + SHADOW_BLUR, WIN_H + SHADOW_BLUR + 6],
        radius=14,
        fill=(0, 0, 0, 175)
    )
    shadow = shadow.filter(ImageFilter.GaussianBlur(radius=SHADOW_BLUR))

    # Full browser frame
    full_browser_frame = base_last_frame.copy()
    full_browser_frame.paste(shadow, (POS_X - SHADOW_BLUR, POS_Y - SHADOW_BLUR), shadow)
    full_browser_frame.paste(browser_win, (POS_X, POS_Y), browser_win)

    # Generate transition frames (0.4s = 10 frames)
    TRANSITION_FRAMES = 10
    frame_idx = 0

    for i in range(1, TRANSITION_FRAMES + 1):
        t = i / TRANSITION_FRAMES
        # ease-out cubic
        ease = 1 - (1 - t) ** 3
        cur_scale = 0.88 + 0.12 * ease
        cur_alpha = int(255 * ease)

        scaled_w = int(WIN_W * cur_scale)
        scaled_h = int(WIN_H * cur_scale)
        scaled_win = browser_win.resize((scaled_w, scaled_h), Image.Resampling.BILINEAR)

        # Apply alpha
        r, g, b, a = scaled_win.split()
        a = a.point(lambda p: int(p * (cur_alpha / 255.0)))
        scaled_win.putalpha(a)

        cur_x = (WIDTH - scaled_w) // 2
        cur_y = 28 + ((HEIGHT - 28) - scaled_h) // 2

        frame = base_last_frame.copy()
        # Scaled shadow
        sc_shadow = shadow.resize((scaled_w + int(SHADOW_BLUR * 2 * cur_scale), scaled_h + int(SHADOW_BLUR * 2 * cur_scale)), Image.Resampling.BILINEAR)
        sr, sg, sb, sa = sc_shadow.split()
        sa = sa.point(lambda p: int(p * (cur_alpha / 255.0)))
        sc_shadow.putalpha(sa)

        sh_x = cur_x - int(SHADOW_BLUR * cur_scale)
        sh_y = cur_y - int(SHADOW_BLUR * cur_scale)

        frame.paste(sc_shadow, (sh_x, sh_y), sc_shadow)
        frame.paste(scaled_win, (cur_x, cur_y), scaled_win)

        frame.convert("RGB").save(f"{TMP_FRAMES_DIR}/frame_{frame_idx:04d}.png")
        frame_idx += 1

    # Generate hold frames (4.5s = 112 frames)
    HOLD_FRAMES = 112
    hold_rgb = full_browser_frame.convert("RGB")
    for _ in range(HOLD_FRAMES):
        hold_rgb.save(f"{TMP_FRAMES_DIR}/frame_{frame_idx:04d}.png")
        frame_idx += 1

    print(f"Generated {frame_idx} post-walkthrough browser frames.")

    # Encode post-walkthrough segment to MP4
    post_mp4 = "/tmp/browser_reveal.mp4"
    subprocess.run([
        "/home/shiva/bin/ffmpeg", "-y",
        "-framerate", str(FPS),
        "-i", f"{TMP_FRAMES_DIR}/frame_%04d.png",
        "-c:v", "libx264",
        "-pix_fmt", "yuv420p",
        "-crf", "18",
        post_mp4
    ], check=True)

    # Concat terminal walkthrough MP4 with browser reveal MP4
    concat_list = "/tmp/concat.txt"
    with open(concat_list, "w") as f:
        f.write(f"file '{ASSETS_DIR}/demo.mp4'\n")
        f.write(f"file '{post_mp4}'\n")

    final_mp4 = os.path.join(ASSETS_DIR, "demo.mp4")
    tmp_final_mp4 = "/tmp/final_demo.mp4"

    subprocess.run([
        "/home/shiva/bin/ffmpeg", "-y",
        "-f", "concat", "-safe", "0",
        "-i", concat_list,
        "-c", "copy",
        tmp_final_mp4
    ], check=True)

    shutil.move(tmp_final_mp4, final_mp4)
    print(f"Final MP4 written to {final_mp4}")

    # Generate high quality GIF (15 fps, bayer scale 3 palette)
    final_gif = os.path.join(ASSETS_DIR, "demo.gif")
    palette_png = "/tmp/palette.png"

    print("Generating optimized palette for demo.gif...")
    subprocess.run([
        "/home/shiva/bin/ffmpeg", "-y",
        "-i", final_mp4,
        "-vf", "fps=14,scale=1380:-1:flags=lanczos,palettegen=max_colors=128:reserve_transparent=0",
        palette_png
    ], check=True)

    print("Encoding final demo.gif...")
    subprocess.run([
        "/home/shiva/bin/ffmpeg", "-y",
        "-i", final_mp4,
        "-i", palette_png,
        "-lavfi", "fps=14,scale=1380:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=3",
        final_gif
    ], check=True)

    print(f"Final GIF written to {final_gif}")
    print("Done!")

if __name__ == "__main__":
    main()
