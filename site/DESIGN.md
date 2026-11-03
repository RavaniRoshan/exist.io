# Exist — site design system

One identity for the marketing landing (`/`) and the docs (`/docs`).

## References & extraction

- **fumadocs.dev** — the "print engineering" reference. Extracted: paper `#f2f0ea`, paper-2 `#e8e5dc`, ink `#121210`, hairlines at 14%/82% ink, layout 1400px / gutter 32→16px, yellow `#fff383` with ink type on it, deep orange accent, crosshair registration marks, ruler tick strips, halftone dot artwork, hard 0-radius frames, status-bar footer, Geist + JetBrains Mono.
- **clawsgo.ai** — the "cinematic dark AI" reference. Extracted: near-black canvas `#0a0a0a`, `#171717` panels, `#ffffff1a` hairlines, dot-grid hero, perspective product plate, Morandi muted panels, white primary CTA + dark-glass secondary, oversized radius on media only, institution/logo wall, mono log narratives.

## The synthesis — "paper, yellow, print engineering"

Light paper is the canonical experience (docs included). Dark mode is a token flip via `prefers-color-scheme`, not a separate design. Yellow is the only loud color; ember carries everything interactive on light; green is reserved for pass/check states.

### Color tokens

| Token | Light | Dark | Use |
|---|---|---|---|
| `--paper` | `#f2f0ea` | `#0d0d0c` | page background |
| `--paper-2` | `#e8e5dc` | `#171715` | panels, muted fills |
| `--paper-3` | `#f7f5ef` | `#131312` | cards on paper |
| `--ink` | `#121210` | `#eceae2` | primary text |
| `--ink-2` | `#4c4a42` | `#b9b6ab` | body/lede text |
| `--ink-3` | `#747167` | `#8b887d` | captions, mono meta |
| `--line` | `rgba(18,18,16,.14)` | `rgba(236,234,226,.13)` | hairlines, dividers |
| `--line-strong` | `rgba(18,18,16,.82)` | `rgba(236,234,226,.75)` | frame borders, inputs |
| `--signal` | `#fff383` | `#fff383` | **primary signal** — selection, primary CTA fill, tape/highlight, active rail |
| `--signal-deep` | `#e9d457` | `#e9d457` | signal borders & pressed state |
| `--ember` | `#b23a0e` | `#ff8a3d` | links, interactive text, active states (contrast-safe on both papers) |
| `--ember-2` | `#8f2d09` | `#ffab6b` | ember hover |
| `--pass` | `#0f7a5a` | `#5ee0b0` | check/merge-pass states only |

Rules:
- **Never** yellow body text on paper; yellow never carries small text except on ink.
- Interactive text on light = ember; hover = `--ember-2`.
- Yellow ↔ ink inversion is the CTA hover signature (Fumadocs).
- No other hues. Morandi tints are allowed only as low-opacity washes inside product plates, never as flat fills.
- Focus ring: 2px ember, 3px offset. Selection: signal background, ink text (both modes).

### Typography

- **Geist** — display & body. Display tracking `-0.04em`, line-height ≤ 1.0, weight 400/500. Body 400/450.
- **JetBrains Mono** — code, coordinates, eyebrow indices, status bar, command chips.
- **Pixel accents are SVG**, not a typeface (logomark, crosshair ticks) — zero font risk.
- No Inter, no serif, no system-ui display.

### Layout & devices

- Container 1400px max; gutter 32px (desktop) → 16px (mobile); section padding 80–112px (desktop) / 56–72px (mobile).
- **Hairline everywhere**: section rules, frames, tables, status bar. 1px, `--line`.
- **Registration crosshairs**: `+` marks at the four corners of hero & major plates (`.crosshair`, pseudo-elements).
- **Ruler tick strips** (`.ruler`): repeating 24px hairlines; used at hero bottom and section dividers.
- **Halftone dot field** (`.halftone`): radial-gradient dot pattern, fixed, pointer-events none.
- **Frames are 0-radius**. **Media plates are 16–24px radius only** (documented exception; concentric double-bezel wrapper: outer `ink/6` shell p-1.5 + inner plate with inset highlight).
- Asymmetric 12-col grids and bento spans; never 3 equal columns.
- Single primary CTA per section; secondary is outline or text link with arrow.
- 8px grid; vertical rhythm locked to it.

### Motion

- Easing `cubic-bezier(0.32, 0.72, 0, 1)`; durations 500–700ms for entrances, 150–250ms for state changes.
- Entrances: fade + translate-y 24–40px + blur(6px) → resolve; IntersectionObserver, stagger ≤ 70ms (max 280ms).
- Press darkens, never scales. Hover inversions (signal↔ink) are discrete flips with 200ms color transition.
- `prefers-reduced-motion`: all entrance motion disabled; content visible immediately.
- No `linear`/`ease-in-out`; no layout-property animation; blur only on sticky surfaces.

### Components

- `ButtonSignal` — signal fill, ink text, 1px `--line-strong` border, 0 radius, height 44px, hover → ink fill + paper text; `ButtonOutline` — transparent, `--line-strong` border, ink text, hover → paper-2 fill; `ButtonInk` — ink fill, paper text (inverted sections).
- `CommandChip` — mono, paper-3 fill, hairline border, copy button (copied state ✓ 1.5s).
- `Eyebrow` — mono 11px uppercase, tracked 0.18em, optional index (`01`) and ember dot.
- `StatusBar` — footer: mono 11–12px, `v0.6.0-beta · PURE GO · ZERO CGO · APACHE-2.0`, pixel dots, hairline top, yellow segment mark.
- `Plate` — double-bezel wrapper; inner plate `--paper-3` with hairline border; captions below in mono (`fig. 1 — pr view`).
- `MergeGate` — interactive demo: three checks gate the merge CTA; success = ink panel, paper text, `--pass` accent.
- `Crosshair`, `.ruler`, `.halftone` — CSS-only decorative devices.

### Logo

`exist.io` wordmark in Geist 600, tracking-tight; to its left a 16px mark: signal-yellow rounded-square (3px radius), 1.5px ink border, centered 4×4 ink pixel. Docs variant: mono `/docs` tag appended (hairline pill). SVG only — no raster dependency.

### Icon

`app/icon.svg` — the mark alone. Favicon, apple-touch, and docs all derive from it.

### Anti-slop guardrails (enforced in review)

- No uniform border-radius across components.
- No purple/blue gradients; no gradient text.
- No emoji in UI copy; no exclamation marks.
- No symmetrical 3-col card rows; no stock imagery; only real product screenshots (`public/product/`) or CSS/SVG art.
- No `Inter`, no fake logos, no "enterprise-ready", no stale counts (see CLAIMS.md; machine-enforced for README + site sources).
