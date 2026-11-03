# Site unification & Paper × Yellow redesign — execution plan

Drafted 2026-11-03. Supersedes the Astro landing build and the half-finished Blume→Fumadocs docs migration.

## Decisions (locked with the owner)

| # | Decision | Choice |
|---|----------|--------|
| 1 | Art direction | **Paper × Yellow, Fumadocs-forward** — light paper home experience, Fumadocs/print-engineering language pushed further: yellow as the primary signal, ember as secondary, status-bar footer, pixel accents |
| 2 | Restyle scope | Marketing **and** docs (one design system, one identity) |
| 3 | Architecture | **Merge marketing into `exist.io` as ONE app** (overrides two-repo plan) — single Next.js static export at `site/`, basePath `/exist.io` |
| 4 | Copy | **Rewrite the copy too** — claims audit updated in the same commits (never weakened) |
| 5 | Order | **Finish the docs migration first** (its own commits), then merge, then redesign |

## Research findings

### Reference A — fumadocs.dev ("technical print / blueprint")
- Palette (extracted from CSS): `--lp-paper:#f2f0ea`, `--lp-paper-2:#e8e5dc`, `--lp-ink:#121210`, `--lp-ink-2:#1212109e`, hairlines `#12121024` / strong `#121210d1`; yellow `#fff383`, deep orange `#af4204` (light) / `#ff7a33` (dark), on-yellow text `#121210`; light-on-yellow selection color.
- Layout tokens: `--lp-max:1400px`, `--lp-gutter:32px→16px`, menu `56px`, section `80px→64px`.
- Signature devices: registration **crosshairs at frame corners**, **ruler tick strips** (24px and 240px repeating gradients), **halftone-dot** hero artwork (yellow sun + orbiting satellites + pixel astronaut), hairline-framed content panels, **status-bar footer**, sharp (0-radius) buttons: yellow fill w/ ink text, hover inverts to ink/paper; Geist display (tight tracking), **Geist Pixel**, JetBrains Mono; dotted scatter background; grid-with-crosses backdrop.

### Reference B — clawsgo.ai ("cinematic dark AI")
- Palette (extracted from CSS): `#0a0a0a` canvas, `#171717` cards, secondary `#262626`, hairline `#ffffff1a`, `muted-fg #a1a1a1`; brand **`#f54e00`** (light) / **`#fa5f29`** (dark); success `#258651`/`#49a46e`; Morandi panel hues (sage `#b4c2b4`, clay `#cfc2b1`, dust `#d5beb9`, slate `#acb5c1`, lilac `#c2bac8`); light mode is `#ffffff`/`#0a0a0a` shadcn defaults; radius `.625rem`.
- Layout: black hero with **dot-grid**, oversized headline, white solid CTA + dark-glass secondary CTA; **perspective-tilted product plate** below the fold; "trusted by" institution logo row; alternating oversized **rounded-3xl media plates** in Morandi tones; mono log/terminal lines; full-mono "hours later…" narrative arc; footer link columns.
- Fonts: Google Sans Flex (display), Inter (body), Geist Mono (code), Noto Serif SC (accents). We will **not** adopt Inter (skills ban it) — Geist family covers display/body/pixel/mono.

### Skills installed (verified: Socket 0 alerts, Snyk low, Gen safe)
- `high-end-visual-design` — from **`leonxlnx/taste-skill-classics`** (the requested repo moved; `npx skills add …/taste-skill --skill high-end-visual-design` no longer resolves). Installed at `~/projects/.agents/skills/high-end-visual-design` (workspace-shared, outside both repos).
- `premium-web-design` — from `Lucxar/premium-web-design-skill`, same location.
- deep-research skill activation was **denied by permissions**; research was done manually (webfetch, curl, headless Chromium screenshots at `/tmp/opencode/research/`). Re-request if you want the skill path.

### Current state / breakage found
1. `.github/workflows/pages.yml` runs **`npm run validate` — which no longer exists** after the Blume→Fumadocs `package.json` rewrite → docs CI would fail on next push. Fix: `typecheck` + `check-links`.
2. Root `.gitignore` misses **`site/out/`** and **`site/.next/`** → build artifacts would be committed by accident.
3. `site/.blume/` — stale Blume cache (untracked but ignored); delete in Phase 0.
4. `.firecrawl/` — research dumps, untracked and unignored → add to `.gitignore` (keep local).
5. Marketing (`exist.os`): Astro 5, base `/exist.os`, claims audit reads `dist/index.html`, deploy via `withastro/action`; current look = paper/ink/**mint** (being replaced).
6. Docs (`exist.io/site`): Fumadocs 16 + Next 16, `output:'export'`, basePath `/exist.io`, `trailingSlash:true`, artifact already switched to `site/out`; custom paper/mint `fd-*` overrides in `app/global.css` (to be re-derived from the new system); `docs-legacy/` kept beside new `content/`; `out/` builds locally.
7. `site/scripts/check-links.mjs` + canonical logic assume docs at **root** of `/exist.io` — must learn the `/docs` prefix (Phase 1).
8. Cross-repo links to `ravaniroshan.github.io/exist.os/` and old doc URLs (`/exist.io/getting-started/` …) are scattered in READMEs, docs content, launch notes, and the marketing copy itself — grep sweep required (Phase 3).

## Target architecture — one app, one identity

- `exist.io/site/` becomes the single **Next.js App Router static export** (existing config: `output:'export'`, `basePath:'/exist.io'`, `trailingSlash:true`, unoptimized images, Node 24).
- Route groups: `app/(marketing)/page.tsx` → **`/` (landing)**; docs move to **`/docs/[[...slug]]`** (git mv of `app/(docs)` + page-tree unchanged; `content/` untouched except frontmatter if needed).
- **Old docs URLs break** (`/exist.io/getting-started/` → `/exist.io/docs/getting-started/`) → emit **redirect stubs** in a `postbuild` script for every current docs route (canonical + `<meta http-equiv="refresh">`), ~25 tiny HTML files; zero-server, Pages-safe.
- **`exist.os` repo → redirect stub** (single page → `ravaniroshan.github.io/exist.io/`, keeps Pages workflow green; old Product Hunt/launch links keep working). Alternative: archive the repo — **needs your go-ahead**; default is redirect stub.
- `sitemap.ts` / `robots.ts` / `check-links` / canonicals / docs-nav / logo link updated for two sections (`/` marketing, `/docs` docs).
- Claims audit moves into `site/scripts/audit-claims.mjs`, reads `out/index.html`; needles updated to rewritten copy; forbidden list kept intact (`Primer`, stale counts, `:3000`, `TODO`…); wired into CI.
- OG image: reuse/regenerate `og-exist-os.png` (1200×630) in new palette into `site/public/`.

## Phase 0 — Close out the docs migration (first commits)

1. `git rm -r --cached` nothing; **add** `site/out/`, `site/.next/`, `.firecrawl/` to `.gitignore`; **delete** `site/.blume/` (regenerable cache); keep `.firecrawl/` on disk.
2. Fix `pages.yml`: `npm run validate` → `npm run typecheck && npm run check-links`.
3. Verify clean: `npm ci` in `site/` → `npm run build` → `npm run check-links` → all green.
4. Commits (Conventional Commits, small):
   - `chore(site): ignore next/out artifacts and blume cache`
   - `feat(site): migrate docs from Blume to Fumadocs + Next static export` (all `site/` changes + `pages.yml` + `site/scripts/*`)
   - `chore(web): add fumadocs extraction and site screenshot tooling` (`web/scripts/*.mjs`)

## Phase 1 — Merge marketing into the app

1. `git mv` docs routes under `/docs`; fix `generateStaticParams`, canonicals, `check-links` base paths, docs nav `url:'/docs'`, search route, fumadocs `source` (unchanged).
2. Port marketing from `exist.os`: `content/marketing.ts` (rewritten in Phase 2), `public/` assets (`forge-*.png`, logos, `og-exist-os.png`, favicons), astro components → TSX client/server components; drop `og.astro` (static PNG instead); canvas hero retired for art-directed CSS/SVG + real product plates.
3. `exist.os` → redirect stub + README pointer (after your go-ahead).
4. Grep sweep in both repos: `exist.os`, `/exist.io/<old-doc-path>` → update to `/exist.io/docs/<path>` (README, CLAIMS, ROADMAP, LAUNCH_READINESS, docs content, web app links, AGENTS.md).
5. CI: `pages.yml` gains claims-audit + typecheck + check-links before upload; artifact unchanged (`site/out`).

## Phase 2 — New identity: "Exist — paper, yellow, print engineering"

**Spec first:** `site/DESIGN.md` — the extracted schema + our own deviations (so future pages stay on-system).

### Palette (own, Fumadocs-forward)
| Token | Light | Dark (docs) |
|---|---|---|
| paper / bg | `#f2f0ea` (paper-2 `#e8e5dc`, tint `#f7f5ef`) | `#0d0d0c` (panel `#171715`) |
| ink | `#121210` / ink-2 `#4c4a42` / ink-3 `#747167` | `#eceae2` / dim `#eceae299` |
| hairline | `rgba(18,18,16,.14)` / strong `.82` | `rgba(236,234,226,.13)` / strong `.75` |
| **signal yellow** | `#fff383` (selection, CTA fill, tape highlights) | same yellow |
| **ember** (links/primary/CTA-hover/active rail) | `#c2410c`→ own hue, e.g. `#b7440a`; on-dark `#ff8a3d` | `#ff8a3d` |
| pass/green (merge-gate checks only) | `#0f7a5a` | `#5ee0b0` |

Rule: yellow = never body text on paper; ember = interactive/text-safe. Yellow ↔ ink inversion on CTA hover (Fumadocs signature). No other hues; Morandi tints may be used **only** as duotone washes inside product plates.

### Type
Geist (display, `-0.04em`, 300/400) + **Geist Pixel** (logomark dot accents, status bar, eyebrow indices) + **JetBrains Mono** (code, coordinates, status bar). No Inter, no serif.

### Layout & devices
1400px max / gutter 32→16; hairline section rules; **registration crosshairs** at frame corners; **ruler tick strips** (24px) as section dividers; **halftone** hero art (SVG, reduced-motion safe); frames = 0-radius hairline boxes; **media plates = 20–28px radius only** (documented exception — kills the "uniform radius" tell); **double-bezel** plates (outer shell `ink/5` + inner plate, concentric radii); asymmetric 12-col grids / bento spans; single CTA per section; 8px grid.

### Motion
`cubic-bezier(0.32,0.72,0,1)` 500–700ms; fade-up + slight blur on entry (IntersectionObserver); press darkens, never scales; no `linear`/`ease-in-out`; `prefers-reduced-motion` = instant; blur only on sticky elements.

### Components
Yellow primary button (ink text; hover → ink fill/paper text) · outline button (strong hairline) · command chip w/ copy · eyebrow tags (mono, indexed) · **status-bar footer** (mono + dots + yellow segment, Fumadocs signature) · logo wall (grayscale, hairline dividers) · product plates: lease JSON, merge-gate interactive demo (rebuilt in React — checkboxes → merge CTA enables), architecture SVG · claims strip → CLAIMS.md.

### Landing (copy rewritten, truth-locked)
Sections: sticky hairline nav (not floating pill — reference wins over skill default) → hero (giant Geist, yellow accent word, install command chip, halftone art + forge PR plate) → protocol wall → lease→push→PR→merge loop (6 steps, mono) → asymmetric feature triptych (leases / reviewer / one-binary) → interactive merge-gate demo → quickstart (3 copyable commands) → evidence strip (claims ledger links) → status-bar footer.
Every sentence traceable to `CLAIMS.md`; rewrite reviewed against it; audit needles updated to the new copy — **audit is strengthened, not relaxed**.

### Docs restyle
Fumadocs neutral preset + our `fd-*` vars → paper/yellow system (light & dark); ember = primary/links, yellow = selection/active rail/attention; sidebar hairlines; logo = pixel-dot + wordmark (`components/logo.tsx`); `og-exist-os.png` regenerated; docs inherit status-bar footer styling (compact).

## Phase 3 — Deploy, links, cleanup

1. `pages.yml`: rename → `Build and deploy site`; steps = ci, typecheck, check-links, claims-audit, build, upload, deploy (Node 24, artifact `site/out`).
2. Redirect stubs emitted into `out/` for all legacy doc paths (postbuild).
3. `exist.os` redirect stub + README; links grep sweep; ROADMAP/LAUNCH_READINESS noted (historical references left intact, current URLs updated).
4. Delete `site/docs-legacy/` once new docs verified built (git history retains it) — **confirm before deleting**.
5. Verify: clean `npm ci` → `typecheck` → `check-links` → `claims-audit` → `build`; screenshots 1440/768/390 + reduced-motion; spot-check CLS/LCP; `go test ./...` untouched (no Go changes).

## Risks / open items
- **R1** Broken old doc URLs → redirect stubs (mitigated).
- **R2** Single app = single failure domain for both surfaces (accepted by owner decision).
- **R3** Copy rewrite drifts from claims → rewrite done against CLAIMS.md; audit updated deliberately.
- **R4** `exist.os` retirement — redirect stub recommended, **your call**.
- **R5** Launch window (target 2026-11-17) — Phase 0 is launch-safe; Phases 1–3 land as a sequence of small commits; recommend doing them on a branch → PR, or straight on `main` in exist.io (your convention).
- **Open**: regenerate OG in new palette (default yes). Marketing dark mode: default = **light primary + auto-dark via `prefers-color-scheme`** (docs keeps toggle) — say the word for toggle-on-marketing instead.
- **Open**: keep launch assets (`exist.os/memory/**`, PH packs referencing `/exist.os/`) as historical records — default yes, only live references updated.

## Commit sequence
`chore(site): ignore artifacts` → `feat(site): fumadocs migration closeout` → `chore(web): tooling` → `refactor(site): merge marketing app under /docs restructure` → `feat(site): paper-yellow design system + landing rewrite` → `feat(site): docs theme + logo + og` → `chore(site): pages workflow + redirect stubs` → `docs(repo): update links to unified site`.
