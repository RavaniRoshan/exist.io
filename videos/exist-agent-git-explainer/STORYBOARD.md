---
format: 1920x1080
duration: 74s
message: "exist.io is the neutral, self-hosted Git forge built specifically for autonomous coding agents"
arc: "concept-explainer with process"
audience: "software engineers, AI builders, and engineering leads"
mode: autonomous
music: none
---

## Video direction

A high-craft developer editorial visual system (`code-editorial` preset). Charcoal and obsidian dark canvas (`#0D1117`), electric cyan accent (`#00ADD8`), emerald success highlights (`#2EA043`), and crisp monospace typography (`JetBrains Mono`, `Inter`). Visual rhythm alternates between high-velocity kinetic typography, abstract architectural sequence diagrams, and terminal-grade code telemetry. The video never front-loads: each visual element reveals precisely as the voiceover names it, ending on held, authoritative states with zero camera drift.

---

## Frame 1 — The Agent Dilemma

- status: animated
- src: compositions/frames/01-agent-dilemma.html
- duration: 11.912s
- transition_in: cut
- scene: Bold kinetic typography opens on an AI agent terminal with a catastrophic red git push error warning
- voiceover: "Autonomous coding agents are rewriting software engineering. But give an agent your personal Git token, and one hallucinated push can wipe out production."
- type: hook
- blueprint: titlecard-reveal (Adapt)
- focal: Warning terminal prompt with glowing destructive force-push command
- roles: warning-badge = top anchor · code-prompt = center hero · destruction-alert = bottom warning pulse
- sfx: terminal-keystrokes, alert-chime

Scene 1 (0.0–2.8s): Clean dark canvas (#0D1117); bold headline "AUTONOMOUS CODING AGENTS" reveals word-by-word in crisp monospace. Subhead "Claude Code · Cursor · Codex · Cline" glides upward underneath.
Scene 2 (2.8–5.2s): Canvas shifts into a floating terminal block: `git push origin main --force` types out; an intense amber-to-crimson warning beacon pulses as "PERSONAL ACCESS TOKEN EXPOSED" flashes in red.
Scene 3 (5.2–7.0s): Red error badge locks dead-center: "FATAL: PRODUCTION OVERWRITTEN"; camera holds still, allowing the high-stakes risk to sink in completely.

---

## Frame 2 — 15-Year-Old Forges

- status: animated
- src: compositions/frames/02-legacy-bottleneck.html
- duration: 12.121s
- transition_in: crossfade
- scene: Three architectural fracture lines illustrating why human-era web forges break under parallel agent swarms
- voiceover: "Traditional Git platforms were built fifteen years ago for humans in web browsers. They have no ref isolation, bottleneck parallel swarms, and treat agents like untrusted intruders."
- type: problem_statement
- blueprint: grid-card-assemble (Adapt)
- focal: Triple split-card showing the three critical agent failure modes
- roles: card-1 = No Ref Isolation · card-2 = Swarm Bottleneck · card-3 = Human PR Lock
- sfx: glitch-transition, data-hum

Scene 1 (0.0–2.5s): Headline "15-YEAR-OLD FORGES" punches in with subtle CRT scanline effect. An antique browser window chrome dissolves, contrasting human browser assumptions against agent speed.
Scene 2 (2.5–5.5s): Three distinct diagnostic cards slide in sequentially: Card 1 reveals "No Ref Isolation" (branch collisions), Card 2 reveals "Rate-Limited Swarms" (agents choked on HTTP 429), Card 3 reveals "Human Bottleneck" (PR queue halted).
Scene 3 (5.5–8.0s): An overarching red diagonal strikethrough strikes across the cards as "UNTRUSTED INTRUDERS" stamps across the baseline. Frame holds firm in crisp contrast.

---

## Frame 3 — Sovereign Agent Forge: exist.io

- status: animated
- src: compositions/frames/03-solution-exist.html
- duration: 10.919s
- transition_in: wipe
- scene: Sleek exist.io brand reveal with glowing cyan emblem and core architecture banner
- voiceover: "Meet exist.io: the sovereign, self-hosted Git forge engineered from first principles for autonomous coding agents and human teams."
- type: solution_intro
- blueprint: logo-assemble-lockup (Adapt)
- focal: Glowing exist.io golden Git branch logo with electric cyan energy rings
- roles: brand-logo = center hero · forge-spec = floating orbital badges · backdrop = architectural blueprint grid
- sfx: energy-whoosh, metallic-click

Scene 1 (0.0–2.4s): Deep space obsidian background activates with subtle cyan wireframe coordinate grid. The exist.io golden emblem scales smoothly into center position, radiating a soft 12px outer glow.
Scene 2 (2.4–5.0s): Crisp title "exist.io" types out in bold typography with badge "Sovereign Git Forge". Three floating pill badges orbit into place: "Pure-Go SSH :2222", "Smart-HTTP :8080", and "Native MCP /mcp".
Scene 3 (5.0–7.0s): Badges lock into a unified architectural triumvirate; a subtle electric cyan pulse sweeps horizontally across the title and holds in pristine stillness.

---

## Frame 4 — Ephemeral Capability Leases

- status: animated
- src: compositions/frames/04-capability-leases.html
- duration: 14.864s
- transition_in: crossfade
- scene: Cryptographic HMAC-SHA256 capability lease token issuance with strict ref-prefix namespace isolation
- voiceover: "Instead of static credentials, agents request ephemeral HMAC-SHA256 capability leases. Every write is strictly isolated to agent branch namespaces. Zero privilege escalation."
- type: mechanism_lease
- blueprint: agent-progress-theater (Adapt)
- focal: Dynamic token authorization card showing TTL countdown and refs/heads/agent/* isolation lock
- roles: token-card = hero card · lock-icon = security anchor · namespace-pill = ref barrier
- sfx: digital-blip, cyber-lock

Scene 1 (0.0–3.0s): Top banner illuminates: `POST /mcp -> lease_request`. A sleek cryptographic token card materialized in center with glowing HMAC-SHA256 hash digits dynamically counting down `TTL: 15m`.
Scene 2 (3.0–6.5s): A branching Git ref tree animates on the right: `main` is shielded by a green lock barrier, while branch `refs/heads/agent/feature` lights up green with permitted write scope. An attempted push outside the prefix hits an illuminated shield barrier: "BLOCKED (403)".
Scene 3 (6.5–9.0s): Green verification badge stamps: "ZERO PRIVILEGE ESCALATION"; the ref isolation diagram locks in glowing clarity and holds still.

---

## Frame 5 — Native MCP & Security Hooks

- status: animated
- src: compositions/frames/05-native-mcp.html
- duration: 13.949s
- transition_in: crossfade
- scene: Model Context Protocol JSON-RPC tool matrix interacting with in-memory pre-receive secret scanner
- voiceover: "With a native Model Context Protocol server mounting thirty-five tools, agents search AST symbols and open PRs effortlessly, while pre-receive hooks kill secret leaks before commits hit disk."
- type: mechanism_mcp
- blueprint: comparison-split (Adapt)
- focal: Dual-panel architecture with MCP JSON-RPC protocol stream on left and pre-receive hook pipeline on right
- roles: mcp-panel = left terminal · security-pipeline = right inspector · data-bridge = center conduit
- sfx: data-stream, laser-pass

Scene 1 (0.0–3.5s): Left panel slides in: "MCP JSON-RPC 2.0 (:8080/mcp)" listing tools `repo_tree`, `symbol_search`, `pull_request_create`, `pull_request_merge` with 35 active tool indicators glowing.
Scene 2 (3.5–7.0s): Right panel slides in: "Pre-Receive Security Hook". A simulated commit stream passes through AST regex filters; a red leaked API key is intercepted in flight: "SECRET BLOCKED (0.04s) — PUSH REJECTED".
Scene 3 (7.0–10.0s): Both panels connect via high-speed emerald data stream; summary tag "Sub-Millisecond Policy Guard" illuminates and holds without movement.

---

## Frame 6 — Install in 5 Minutes & CTA

- status: animated
- src: compositions/frames/06-install-cta.html
- duration: 10.266s
- transition_in: wipe
- scene: Single-binary terminal installation command with GitHub link and Product Hunt launch badge
- voiceover: "Pure Go. Zero host dependencies. Deploy in under five minutes with existd bootstrap, and give your coding agents a forge of their own."
- type: call_to_action
- blueprint: cta-morph-press (Adapt)
- focal: High-contrast terminal box running `curl -sSL https://exist.io/install.sh | bash`
- roles: terminal = hero center · specs = bottom pill metrics · cta-button = final call
- sfx: terminal-enter, success-fanfare

Scene 1 (0.0–3.0s): Dark terminal window opens centered: `$ curl -sSL https://exist.io/install.sh | bash` types rapidly with green confirmation: `✓ existd v0.5.0-beta installed (CGO_ENABLED=0)`.
Scene 2 (3.0–6.0s): `$ existd bootstrap --namespace acme` executes; three metrics animate below: "100% Pure Go", "Zero Host OpenSSH", "< 5 Minute Setup".
Scene 3 (6.0–9.0s): Terminal resolves to grand headline: "Give Your Agents a Sovereign Forge." GitHub repo URL and Product Hunt badge illuminate and hold in crisp stillness.
