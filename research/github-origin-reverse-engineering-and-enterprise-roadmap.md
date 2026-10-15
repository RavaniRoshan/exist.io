# Deep Research: Reverse Engineering GitHub Enterprise Origin for exist.io

## Executive Summary
This report presents a reverse-engineered architectural blueprint of GitHub Enterprise (Cloud and Server) across repository management, repository settings, branch protection rulesets, access control, and danger zone security, juxtaposed directly against the current implementation of **exist.io**. Through parallel research agents and adversarial verification, we identified the foundational architecture required for 2nd, 3rd, and 4th layer enterprise forge capabilities. The primary findings establish that while exist.io provides functional SQLite WAL and Git bare repo primitives, substantial gaps exist in danger zone security (absence of sudo-mode / security-key reauthentication, lack of multi-step visibility-change safeguards), repository lifecycle management (missing repository rename, transfer, and archival locks), and Git push-level branch protection enforcement (hooks currently only scan for credentials rather than blocking direct pushes or force-pushes to protected refs).

---

## 1. Multiple Repositories Management & Dashboard Architecture

### 1.1 Repository Listing & Search Surface
* **Filtering Dimension Hierarchy**: In GitHub Enterprise, repository lists support multi-dimensional faceted filtering:
  - **Type Filter**: `All`, `Public`, `Private`, `Sources` (non-forks), `Forks`, `Archived`, `Mirrors`, `Templates`.
  - **Language Filter**: Dynamically populated based on detected repository languages via GitHub Linguist.
  - **Sort Ordering**: `Last updated` (via `pushed_at` timestamp), `Name` (alphabetical ascending/descending), and `Stars` (stargazer count).
  - **Search Query (`q`)**: Full-text and qualifier search (e.g. `is:private`, `archived:false`, `language:go`).
* **Repository Card Metadata**:
  - Primary title with visibility pill badge (`Public`, `Private`, `Internal`).
  - Fork attribution link (`Forked from <upstream/repo>`).
  - Description paragraph and repository topics.
  - Primary language color circle and language name.
  - SPDX License badge with Octicon scale icon.
  - Live star count and fork count.
  - Relative commit activity timestamp (`Updated 2 hours ago`).
* **Centralized Enterprise Dashboard (`/repos`)**:
  - Global cross-organization aggregation for the authenticated user.
  - Built-in views: `My repositories`, `My contributions`, `My forks`, `Admin access`.
  - Custom saved views (up to 25) with custom icon, title, and query filters.

### 1.2 Repository Creation Flow (`/new`)
* **Namespace & Owner Selector**: Allows scoping new repositories to a personal user account or an enterprise organization where the user has repository creation permissions.
* **Real-time Name Validation**: Enforces uniqueness in the selected namespace, <= 100 characters, alphanumeric with `-`, `.`, and `_`.
* **Enterprise Visibility Governance**:
  - In enterprise organizations, `Internal` is the default visibility.
  - Enterprise administrators can enforce policies disabling `Public` repository creation entirely to eliminate external leak vectors.
* **Initialization Primitives**:
  - Add `README.md`.
  - Select `.gitignore` template from language presets.
  - Select standard open-source license (MIT, Apache 2.0, GPL, BSD).

---

## 2. Repository Settings & Danger Zone Security

### 2.1 Danger Zone Operations & Consequences Matrix
GitHub isolates destructive and security-critical operations in a dedicated red-bordered section at the bottom of General Settings (`Settings -> Danger Zone`):

| Operation | Trigger Requirements | System Impact & Consequences | Safeguard UX Flow |
| :--- | :--- | :--- | :--- |
| **Change Visibility (Public -> Private)** | Admin or Org Owner + Sudo Mode / Passkey Re-auth | - Stars & Watchers **permanently erased**.<br>- Fork network detached into standalone public repos.<br>- GitHub Pages automatically unpublished.<br>- Excluded from global search and external search engine indexing.<br>- Actions runners consume private billable minutes. | 1. Modal enumerating 5 critical warnings.<br>2. Checkbox: *"I have read and understand these effects"*;.<br>3. Exact typing of repository path: `{owner}/{repo}`.<br>4. Re-auth challenge (Password or WebAuthn/FIDO2 security key). |
| **Change Visibility (Private -> Public)** | Admin or Org Owner + Sudo Mode / Passkey Re-auth | - Entire source code, commit history, and branches become world-readable.<br>- Stars & Watchers **preserved**.<br>- Private forks detached into standalone private repos.<br>- Actions runner minutes become free and unlimited.<br>- Push rulesets automatically disabled. | 1. Warning modal about intellectual property exposure.<br>2. Checkbox confirmation.<br>3. Typing `{owner}/{repo}`.<br>4. Re-auth challenge. |
| **Transfer Ownership** | Admin + Org Repo Creation Rights in Destination + Sudo Mode | - Git clone and web URLs redirect to destination.<br>- Issues, PRs, wikis, stars, releases, and commit history transfer intact.<br>- Read-only outside collaborators dropped if moving to personal account. | 1. Select destination user/org.<br>2. Warning modal explaining URL redirect loss on re-creation.<br>3. Typing `{owner}/{repo}`.<br>4. Destination acceptance (instant for owned org, 24h email invite for personal users). |
| **Archive Repository** | Admin + Sudo Mode | - Entire repository transitions to immutable read-only state.<br>- Direct git pushes, new commits, PR creation, issue filing, commenting, and collaborator modifications **blocked**.<br>- Secret scanning continues; forking and starring remain open.<br>- Unarchiving is reversible by repository admins at any time. | 1. Warning modal explaining read-only freeze.<br>2. Typing `{owner}/{repo}`.<br>3. Confirmation button. |
| **Delete Repository** | Admin or Org Owner + Sudo Mode | - Permanent deletion of git repository and SQLite metadata.<br>- Team permissions purged.<br>- Public forks detached; private forks permanently erased.<br>- GitHub provides a 90-day soft-delete restoration window. | 1. Click *"I want to delete this repository"*.<br>2. Click *"I have read and understand these effects"*.<br>3. Type `{owner}/{repo}` in verbatim red input field.<br>4. Sudo mode password/passkey challenge. |

### 2.2 Sudo Mode Security Engine
* **Purpose**: Prevents session hijacking and unauthorized changes if an authenticated browser tab is left unattended.
* **Eligible Factors**: Password, WebAuthn/FIDO2 hardware security keys, biometric passkeys (Touch ID, Windows Hello), TOTP 2FA authenticator app codes (SMS is explicitly disallowed for sudo mode).
* **Sliding Session Lifetime**: 2 hours from last successful re-authentication.

---

## 3. Branch Protection Rulesets & Enforcement Architecture

### 3.1 Modern Rulesets vs Classic Protection
GitHub Enterprise has transitioned from single-rule classic branch protection to **Rulesets**:
* **Multi-Ruleset Aggregation**: Multiple rulesets can target the same branch or ref pattern simultaneously. Rules aggregate concurrently without a single "winner"; all conditions across all matching rulesets must be satisfied.
* **Three Enforcement States**: `Active` (strictly enforced), `Evaluate` (dry-run monitoring, logging what would fail without blocking workflows), and `Disabled`.
* **Universal Scope**: Rulesets operate at Enterprise, Organization, and Repository tiers, with parent rulesets cascading downward and preventing local repository administrators from bypassing them unless explicitly permitted.

### 3.2 Core Gating Rules
1. **Require Pull Request Before Merging**:
   - Required approving reviews count (1 to 6).
   - Dismiss stale approvals when new commits are pushed to the PR branch.
   - Require approval from designated `CODEOWNERS`.
   - Require approval of the most recent push (cannot merge own commits without peer re-approval).
2. **Require Status Checks to Pass**:
   - Strict mode: PR branch must be rebased/merged with the target branch before merge.
   - Specific check source filtering: checks must originate from authorized CI apps.
3. **Require Conversation Resolution**:
   - All review comments and code discussions must be marked resolved before merging.
4. **Require Linear History**:
   - Prevents merge commits; mandates squash merge or rebase merge.
5. **Require Cryptographically Signed Commits**:
   - Commits pushed or merged must have valid GPG, SSH, or S/MIME signatures.
6. **Restrict Force Pushes & Deletions**:
   - Blocks non-fast-forward updates and `git push --delete`.
7. **Bypass Permissions**:
   - Specific roles, teams, or apps granted `always` bypass or `pull_request` only bypass (which creates an audited pull request without requiring checks).

### 3.3 Enforcement Point Gap in exist.io
* **Pull Request Merge Gate**: Implemented via `engine.EvaluateMerge` in `internal/engine/ruleset.go`.
* **Git Push Hook Enforcement (CRITICAL GAP)**:
  - On GitHub Enterprise, Git push attempts (over HTTP or SSH) trigger pre-receive and branch ruleset hooks. Pushes violating branch protection are aborted with:
    `remote: error: GH006: Protected branch update failed for refs/heads/main`
  - In **exist.io**, the `pre-receive` hook (`internal/security/hook.go`) **only executes secret scanning** (`DefaultScanner.ScanDiff`). Users can run `git push origin main` or `git push --force origin main` directly via SSH or HTTP and bypass all rulesets!

---

## 4. Access Control, Deploy Keys, Secrets & Webhooks

### 4.1 Collaborator Role Hierarchy
In GitHub Enterprise Organizations, the 5-tier role hierarchy governs permissions:
* `Read`: Clone, pull, view issues/PRs/releases, submit reviews.
* `Triage`: Read + label, assign, close/reopen issues/PRs, mark duplicates.
* `Write`: Triage + push to non-protected branches, create releases, merge PRs, edit Actions secrets.
* `Maintain`: Write + push to protected branches (if bypass granted), manage descriptions, configure merge behaviors.
* `Admin`: Full governance, manage collaborators, rulesets, webhooks, deploy keys, visibility, and deletion.
*(Note: Personal user repositories only support a single binary collaborator permission: push/pull write access).*

### 4.2 Deploy Keys
* SSH public keys registered directly at the repository level (`Settings -> Deploy keys`).
* Default to read-only (`git clone`, `git fetch`), with an optional "Allow write access" toggle for build automations.
* Unique per repository (cannot be reused across repos).

### 4.3 Webhook Subscriptions & Signature Verification
* Subscriptions for push, pull request, issues, and release events.
* Signature: `X-Hub-Signature-256` computed via `HMAC-SHA256(secret, payload)`.
* Comprehensive delivery log inspection with full headers, JSON payload, response code, and "Redeliver" capability.

---

## 5. exist.io Codebase Audit & Gap Inventory

| Feature Domain | Existing in exist.io | Gap Against GitHub Enterprise | Priority |
| :--- | :--- | :--- | :--- |
| **Multi-Repo Management** | SQLite DB table `repos`, paginated `GET /api/v1/repos`, bare git init on disk (`POST /api/v1/repos`). | - `description` is received in API payload but dropped (no column).<br>- No repository filters (Type: Public/Private/Archived, Sort: Updated/Name/Stars).<br>- No README / .gitignore / License initialization templates in creation flow. | **High** |
| **Danger Zone: Visibility** | Toggle button in `RepoSettingsView.tsx` via `PATCH /api/v1/repos/{owner}/{name}`. | - **No security confirmation modal**: changes instantly with no caution.<br>- **No sudo-mode / password challenge** before changing visibility.<br>- Does not warn about access or search implications. | **Critical** |
| **Danger Zone: Lifecycle** | Delete repo with name confirmation string. | - **No Repository Archival** (freeze commits/PRs/issues into read-only).<br>- **No Repository Rename** (with Git clone and web redirect tracking).<br>- **No Repository Transfer** (moving between owners/namespaces). | **High** |
| **Branch Protection on Push** | Rulesets table, `engine.EvaluateMerge` during PR merge API. | - **Zero protection on direct Git push**: `PreReceiveHook` only scans for secrets.<br>- Direct pushes and force-pushes to `main` over SSH/HTTP bypass all rulesets. | **Critical** |
| **Collaborator Permissions** | `grants` table with `read`, `write`, `admin` roles; management in `RepoSettingsView.tsx`. | - No pending invitation lifecycle.<br>- No team/group grants in web UI (only individual users). | **Medium** |
| **Webhooks & Deploy Keys** | System/App webhook deliveries in `WebhooksView.tsx`. | - No repository-level webhooks (`/repos/{owner}/{repo}/hooks`).<br>- No repository deploy keys. | **Medium** |

---

## 6. Implementation Roadmap for exist.io

```mermaid
flowchart TD
    subgraph Phase 1: Security & Danger Zone Parity
        A[Sudo-Mode & Re-Auth Challenge] --> B[Multi-Step Visibility Change Modal]
        B --> C[Repository Archival Read-Only Flag]
        C --> D[Repository Rename with Path Redirection]
    end
    subgraph Phase 2: Git Push Hook Enforcement
        E[Extend PreReceiveHook in cmd/existd] --> F[Ref Update Inspector]
        F --> G[Enforce Protected Branch Blocks on HTTP/SSH Push]
        G --> H[Reject Force-Pushes on Active Rulesets]
    end
    subgraph Phase 3: Multi-Repo Management & Dashboard
        I[Add description & archived to DB Schema] --> J[Repo List Filter & Sort Bar]
        J --> K[Repository Creation Template Initializer]
        K --> L[SPDX License & Gitignore Presets]
    end
    Phase 1 --> Phase 2 --> Phase 3
```

### Immediate Action Plan:
1. **Layer 1: Danger Zone Security & Visibility Protection Modal**
   - Create `ConfirmVisibilityChangeModal.tsx` in `web/src/components/modals/` requiring:
     a) Explicit review of visibility change warnings (impact on access, forks, and search).
     b) Typing `{owner}/{repo}` to confirm.
     c) Sudo-mode security challenge: requiring the user's password/auth token to authorize the modification.
   - Add backend validation in `internal/app/app.go` ensuring `PATCH /api/v1/repos/{owner}/{name}` with visibility changes verifies the caller's credentials or requires admin sudo session.
2. **Layer 2: Repository Archival & Read-Only Freeze**
   - Add migration for `archived BOOLEAN DEFAULT 0` in `repos` table.
   - Enforce archival status in `internal/app/` (blocking push, PR creation, issue editing when `archived == 1`).
   - Add "Archive this repository" action in `RepoSettingsView.tsx` Danger Zone.
3. **Layer 3: Git Push Hook Branch Protection Enforcement**
   - Hook `engine.EvaluatePush` into `internal/security/hook.go:PreReceiveHook`.
   - When a user pushes refs over SSH (`git-receive-pack`) or HTTP (`smart-http`), evaluate active rulesets for `refs/heads/<branch>`:
     - Block direct pushes to branches where `require_pull_request` is enabled.
     - Block force pushes (`non_fast_forward`) unless bypass actor.
     - Reject with Git standard `remote: error: GH006: Protected branch update failed for refs/heads/<branch>`.
4. **Layer 4: Multi-Repository Listing, Filters & Creation Templates**
   - Update `CreateRepoModal.tsx` and `POST /api/v1/repos` to store `description`, add `README.md` initialization, and add `.gitignore` and `LICENSE` templates.
   - Add Type (`All`, `Public`, `Private`, `Archived`), Language, and Sort (`Updated`, `Name`, `Stars`) filter controls in `GlobalHomeView.tsx` and user/org repo lists.
