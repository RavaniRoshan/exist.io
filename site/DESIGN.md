# Exist — one design system for the forge, the site, and the docs

The product (`web/`), the landing (`site/`), and the documentation (`site/docs`)
share one palette, one typeface, and one set of surfaces. There is no separate
"marketing theme". A visitor can move from the landing into the docs and flip
light/dark without the identity changing underneath them.

## Source of truth

| Layer | File | Owns |
|---|---|---|
| Product tokens | `web/src/styles/tokens.css` | OKLCH palette, radii, font stacks, motion |
| Product semantics | `web/src/styles/theme.css` | shadcn-style semantic aliases |
| Site mirror | `site/app/global.css` | the same palette plus landing/docs devices |
| Site fonts | `site/app/layout.tsx` | Mona Sans Variable via `@fontsource-variable/mona-sans` |

If a colour changes, it changes in `web/src/styles/tokens.css` first, then the
matching values in `site/app/global.css`. The site deliberately keeps legacy
names (`--paper`, `--ink`, `--signal`, `--ember`) as aliases so older marketing
components resolve to the new palette instead of drifting.

## The palette — Patina on Primer structure

Primer supplies the *structure*: canvas / slab / raised surfaces, hairline
borders, muted text, 6px controls. Patina (oxidized copper) is the signal;
brass is the warm secondary. Light is the default, dark is a real token flip.

| Token | Light | Dark | Role |
|---|---|---|---|
| `--canvas` | `oklch(0.972 0.002 250)` ≈ `#f6f8fa` | `oklch(0.155 0.012 258)` ≈ `#10151c` | page |
| `--slab` | `oklch(1 0 0)` | `oklch(0.195 0.014 258)` | cards, plates |
| `--raised` | `oklch(0.955 0.003 250)` | `oklch(0.225 0.013 258)` | hover fills, bars |
| `--hairline` | `oklch(0.885 0.006 255)` | `oklch(0.285 0.01 255)` | 1px borders |
| `--text-1` | `oklch(0.238 0.008 265)` | `oklch(0.935 0.012 250)` | primary text |
| `--text-2` | `oklch(0.44 0.012 258)` | `oklch(0.635 0.013 252)` | body |
| `--text-3` | `oklch(0.575 0.012 258)` | `oklch(0.485 0.012 252)` | captions, meta |
| `--patina` | `oklch(0.45 0.065 184)` ≈ `#2E6F63` | `oklch(0.82 0.06 186)` ≈ `#8CD2C3` | primary signal |
| `--patina-hover` | `oklch(0.39 0.07 184)` | `oklch(0.75 0.06 186)` | primary hover |
| `--patina-ink` | `oklch(0.965 0.012 181)` | `oklch(0.16 0.02 184)` | text on patina |
| `--patina-soft` | `oklch(0.93 0.025 184)` | `oklch(0.28 0.03 186)` | selected rows, badges |
| `--brass` | `oklch(0.63 0.115 86)` | `oklch(0.78 0.1 88)` | secondary / warning |
| `--brass-soft` | `oklch(0.94 0.045 88)` | `oklch(0.3 0.045 86)` | brass wash |
| `--danger` | `oklch(0.55 0.2 25)` | same | destructive |

Rules:

- Patina carries primary actions, active states, pass states, and the one
  accented phrase in a headline. It is not a background for body copy.
- Brass is for "denied", warnings, and the second window dot. Never for primary
  actions.
- Hairlines are 1px. Strong frames use `color-mix(text-1 82%, transparent)`;
  dark mode uses 75%.
- No other hues. Product screenshots may contain their own interior colours.
- Selection is patina with `--patina-ink`; focus is a 2px patina ring at 3px.

### Dark mode

The Fumadocs provider (`components/provider.tsx`) owns a `next-themes`
provider with `attribute: "class"`. The `.dark` class flips every token,
including the `--color-fd-*` docs variables and the shadcn aliases. There is no
`prefers-color-scheme` CSS branch. `ThemeToggle` in the marketing nav calls the
same provider, so the landing and docs stay in sync and the choice persists.

## Typography

- **Mona Sans Variable** — everything sans. Weights: 300 for display, 400 for
  body, 500–600 for UI labels and the wordmark. Display tracking `-0.035em` to
  `-0.047em`; display line-height `0.98–1.02`.
- **`ui-monospace`** — eyebrows, coordinates, code, status bars, command chips,
  figure captions, terminal. No second webfont.
- Headlines are `font-light` with one patina phrase: "From lease to **merge.**",
  "**Policy** at every step", "Built for **agents.**", "**Evidence,** not
  adjectives." Do not colour whole headlines.
- Body copy is 15–18px with `text-wrap: pretty`; headings use `balance`.
- Mono meta is uppercase with `0.14em–0.22em` tracking, 9–11px.

## Layout

- `.container-site`: max 1400px, 16px gutters, 32px from `md`.
- Sections: `py-16` mobile / `py-24` desktop. Major sections end in a hairline.
- 12-column grids, asymmetric spans (`5/4/3`, `7/5`). Never three equal cards.
- Landing order: `Nav → Hero → SwarmBand → Loop → Product → Platform →
  Quickstart → Evidence → FinalCta → StatusFooter`.
- Every section starts with a numbered mono eyebrow (`01 — the agent loop`).
- Rules, crosshairs, rulers, and dither seams provide structure instead of
  shadows. Cards are flat; only media plates get a slight elevation.

## Surfaces and devices

| Class | Use |
|---|---|
| `.frame` / `.frame-strong` | flat hairline panels (lease JSON, ref isolation, layers) |
| `.window` / `.window-bar` / `.window-dots` | browser/terminal chrome around live demos and screenshots |
| `.plate` / `.plate-inner` | double-bezel media plate, 14px outer / 9px inner |
| `.btn` / `.btn-signal` / `.btn-ink` / `.btn-ghost` | 44px controls, 6px radius, 1px border, hover inversions; disabled resolves to `--raised` + `--text-3` |
| `.btn-chip` | circular trailing icon; nudges on hover |
| `.cmd` / `.cmd-scroll` / `.cmd-copy` | install chip; horizontal scroll with mask, copy button |
| `.tab` | mono tab strip; patina underline on `aria-selected="true"` |
| `.kbd` | keyboard key, mono 10px |
| `.eyebrow` / `.marker` | indexed label; patina wash behind one headline phrase |
| `.ruler` / `.crosshair` / `.halftone` / `.grid-field` / `.hatch` / `.dither` | print-engineering instruments |
| `.statusbar` / `.pixel` | footer rail and square mark |
| `.gate-item` / `.gate-box` | merge-gate checklist rows |

Frames are 6–8px radius; media windows 10px. Controls are 6px. No pill-shaped
cards, no glassmorphism, no gradients on text.

## Motion

- Easing: `--ease-out: cubic-bezier(0.2, 0, 0, 1)`; entrances use
  `--ease-spring: cubic-bezier(0.32, 0.72, 0, 1)`.
- Entrance: `.reveal` = opacity 0 → 1, `translateY(28px)` → 0, `blur(6px)` → 0,
  760ms, driven by an IntersectionObserver in `Reveal.tsx`.
- State changes: 140–200ms colour/border transitions. Press translates 1px; it
  never scales.
- Looping devices: marquee 38s, caret 1.05s step-end, status dot pulse 2.4s,
  scanline 1.1s, pulse ring 2s, `rise-in` 420ms.
- Hover inversions: platform cards, evidence rows, and primary buttons flip
  foreground/background rather than adding a shadow.
- `prefers-reduced-motion` disables reveals, marquees, carets, pulses,
  scanlines, and smooth scrolling; content renders in its final state.

## Interactive modules

- **`ForgeDemo`** (hero) — auto-sequences lease → push → PR → checks → merge in
  5 rows, with pause/play, replay, clickable progress segments, an aria-live
  narration, and a "merged" stamp. Reduced motion jumps to the completed state.
- **`Loop`** (01) — six-step tabbed pipeline. Auto-advances every 4.6s until
  the pointer or focus enters; supports Arrow keys, Home, End; each step types
  a terminal transcript with `rise-in` lines and a caret.
- **`ProductPreview`** (02) — keyboard-accessible tabs over the three real
  screenshots; changes trigger a scanline and a caption/badge swap.
- **`MergeGate`** — check the three policy items to enable the merge button;
  merging shows an inverted success panel with a pulse ring and reset.
- **`QuickstartTerminal`** (04) — scroll-triggered, types install → bootstrap →
  run, prints verified output lines, supports skip/run-again, copies all
  commands. The visible typing is `aria-hidden`; an `sr-only` block keeps the
  real commands in the initial HTML for the claims gate.
- **`PixelField` + `SwarmBand`** — canvas blueprint dots with agent cells,
  trails, and a pointer wake; pauses off-screen and under reduced motion.
- **`ScrollProgress`** — 1px patina rail at the bottom of the sticky nav.
- **`LiveClock`** — UTC clock in the status bar; hydrates without mismatch.
- **`ThemeToggle`** — one switch for landing + docs.

## Screenshots

All product imagery is captured from a real seeded `existd`, not mocked in the
site:

```bash
bash scripts/capture/seed.sh /tmp/exist-capture 18081
# Playwright: 1440×900 CSS viewport, deviceScaleFactor 2, dark scheme,
# token injected into localStorage, onboarding tour suppressed
```

| Asset | View |
|---|---|
| `site/public/product/forge-home.png` | repository overview + activity feed |
| `site/public/product/forge-agents.png` | agent console / MCP + leases |
| `site/public/product/forge-pr.png` | pull request + checks + merge gate |
| `site/public/forge-pr.png` | docs hero (same capture) |
| `site/public/og-exist-os.png` | 1200×630 social card, dark Patina |

The captures are 2880×1800 (2× DPR) but are rendered in a 1440×900 box, so
`next/image` dimensions stay `1440×900`. Figures are numbered 1–4 because the
hero demo is fig. 1. Captions state the view and "real ui".

## Documentation

- Fumadocs tokens are derived from the same palette:
  `--color-fd-background: var(--canvas)`, `--color-fd-primary: var(--patina)`,
  etc. The docs sidebar, TOC, search, and code blocks inherit the identity.
- `lib/layout.shared.tsx` and every marketing link use `asset()` /
  `basePath` so production URLs stay under `/exist.io/`.
- Docs MDX exposes the shared `Kbd` primitive. Frames use `Frame` with
  `/forge-pr.png`.
- The theme toggle in the docs sidebar and the marketing `ThemeToggle` write the
  same preference.

## Accessibility

- Tabs use `role="tablist"`, `aria-selected`, roving `tabIndex`, and full
  keyboard support.
- The merge demo exposes `aria-live` state; decorative canvases are
  `aria-hidden`.
- Focus rings are visible on both palettes; hover inversions are duplicated
  with `focus-visible` styles.
- The terminal's typed surface is decorative; commands remain available to
  assistive tech and to the claims audit.
- Contrast: body copy uses `--text-2` on `--canvas`; small mono meta uses
  `--text-3` only for non-essential captions. Patina fills always pair with
  `--patina-ink`.

## Guardrails

- No second palette, no Tailwind default colours in new marketing code.
- No stale screenshots, fake product UI, stock photos, or invented metrics.
- No claim without a `CLAIMS.md` entry; `npm run claims-audit` reads the built
  HTML and must stay green. The forbidden list includes "primer" and stale
  counts.
- No animation library; native CSS keyframes + canvas are enough.
- Every push gate: `site npm run verify`, `web npm test && npm run build`,
  `go test ./...`.
