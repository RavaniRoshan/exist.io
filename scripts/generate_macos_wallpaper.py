#!/usr/bin/env python3
"""
Generate an authentic macOS Sequoia / Sonoma dark desktop canvas wallpaper
with a native macOS top menu bar and vector system icons for VHS demo recording.
"""
from PIL import Image, ImageDraw, ImageFilter, ImageFont
import math

WIDTH = 1380
HEIGHT = 820

# 1. Base gradient (deep cosmic twilight)
img = Image.new("RGBA", (WIDTH, HEIGHT), (10, 14, 26, 255))
base_draw = ImageDraw.Draw(img)

# Multi-stop vertical gradient
for y in range(HEIGHT):
    t = y / HEIGHT
    r = int(10 * (1 - t) + 24 * t + 10 * math.sin(t * math.pi))
    g = int(14 * (1 - t) + 16 * t + 8 * math.sin(t * math.pi))
    b = int(28 * (1 - t) + 48 * t + 25 * math.sin(t * math.pi))
    base_draw.line([(0, y), (WIDTH, y)], fill=(r, g, b, 255))

# 2. Add organic Apple-style glowing auras
glow_layer = Image.new("RGBA", (WIDTH, HEIGHT), (0, 0, 0, 0))
glow_draw = ImageDraw.Draw(glow_layer)

# Vibrant cyan/teal wave on top-right
glow_draw.ellipse([WIDTH * 0.55, -HEIGHT * 0.2, WIDTH * 1.15, HEIGHT * 0.7], fill=(0, 150, 200, 110))
# Royal violet / magenta glow on bottom-left
glow_draw.ellipse([-WIDTH * 0.15, HEIGHT * 0.35, WIDTH * 0.6, HEIGHT * 1.1], fill=(130, 40, 180, 100))
# Warm twilight amber accent in center-bottom
glow_draw.ellipse([WIDTH * 0.3, HEIGHT * 0.6, WIDTH * 0.85, HEIGHT * 1.2], fill=(210, 90, 40, 70))
# Deep indigo mid-tone orb
glow_draw.ellipse([WIDTH * 0.2, HEIGHT * 0.1, WIDTH * 0.7, HEIGHT * 0.8], fill=(60, 50, 160, 90))

# Heavily blur the glows to produce silky smooth Apple-style wallpaper
blurred_glow = glow_layer.filter(ImageFilter.GaussianBlur(radius=95))
img = Image.alpha_composite(img, blurred_glow)

# 3. macOS Top Menu Bar
MENU_HEIGHT = 28
menu_bar = Image.new("RGBA", (WIDTH, MENU_HEIGHT), (18, 20, 28, 215))
menu_draw = ImageDraw.Draw(menu_bar)

# 1px bottom border
menu_draw.line([(0, MENU_HEIGHT - 1), (WIDTH, MENU_HEIGHT - 1)], fill=(255, 255, 255, 30))

img.paste(menu_bar, (0, 0), menu_bar)

draw = ImageDraw.Draw(img)
try:
    font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 12)
    font_bold = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 12)
except Exception:
    font = ImageFont.load_default()
    font_bold = font

# Vector Apple Logo
apple_x, apple_y = 16, 7
draw.ellipse([apple_x, apple_y + 3, apple_x + 10, apple_y + 13], fill=(240, 240, 245, 240))
draw.ellipse([apple_x + 5, apple_y + 3, apple_x + 13, apple_y + 13], fill=(240, 240, 245, 240))
draw.ellipse([apple_x + 11, apple_y + 4, apple_x + 17, apple_y + 11], fill=(18, 20, 28, 215)) # bite
draw.ellipse([apple_x + 6, apple_y, apple_x + 9, apple_y + 4], fill=(240, 240, 245, 240)) # leaf

# Left menu items
draw.text((40, 7), "Terminal", fill=(255, 255, 255, 255), font=font_bold)
draw.text((115, 7), "Shell", fill=(220, 220, 230, 220), font=font)
draw.text((165, 7), "Edit", fill=(220, 220, 230, 220), font=font)
draw.text((205, 7), "View", fill=(220, 220, 230, 220), font=font)
draw.text((250, 7), "Window", fill=(220, 220, 230, 220), font=font)
draw.text((315, 7), "Help", fill=(220, 220, 230, 220), font=font)

# Right menu items: Date Time
right_text = "Tue Oct 17  9:41 AM"
draw.text((WIDTH - 145, 7), right_text, fill=(240, 240, 245, 230), font=font_bold)

# Vector Control Center switch
cc_x = WIDTH - 175
draw.rounded_rectangle([cc_x, 8, cc_x + 16, 20], radius=3, outline=(220, 220, 235, 200), width=1)
draw.line([cc_x + 4, 11, cc_x + 12, 11], fill=(220, 220, 235, 200), width=1)
draw.line([cc_x + 4, 17, cc_x + 12, 17], fill=(220, 220, 235, 200), width=1)
draw.ellipse([cc_x + 5, 9, cc_x + 9, 13], fill=(220, 220, 235, 220))
draw.ellipse([cc_x + 8, 15, cc_x + 12, 19], fill=(220, 220, 235, 220))

# Vector WiFi
wifi_x = WIDTH - 205
draw.arc([wifi_x, 7, wifi_x + 14, 21], start=210, end=330, fill=(220, 220, 235, 220), width=2)
draw.arc([wifi_x + 2, 11, wifi_x + 12, 21], start=210, end=330, fill=(220, 220, 235, 220), width=2)
draw.ellipse([wifi_x + 6, 17, wifi_x + 8, 19], fill=(220, 220, 235, 240))

# Vector Battery (100% charged green)
bat_x = WIDTH - 235
draw.rounded_rectangle([bat_x, 9, bat_x + 18, 19], radius=2, outline=(220, 220, 235, 200), width=1)
draw.line([bat_x + 19, 12, bat_x + 19, 16], fill=(220, 220, 235, 200), width=1)
draw.rectangle([bat_x + 2, 11, bat_x + 16, 17], fill=(52, 199, 89, 230))

# Vector Spotlight Search
spot_x = WIDTH - 260
draw.ellipse([spot_x, 8, spot_x + 8, 16], outline=(220, 220, 235, 200), width=1)
draw.line([spot_x + 7, 15, spot_x + 11, 19], fill=(220, 220, 235, 200), width=2)

# Save output
output_path = "assets/macos_wallpaper.png"
img.save(output_path, "PNG")
print(f"Generated {output_path} ({WIDTH}x{HEIGHT}) with vector macOS menu bar.")
