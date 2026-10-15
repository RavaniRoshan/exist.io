import React, { useState, useEffect } from 'react'
import {
  Settings,
  Users,
  AlertTriangle,
  GitBranch,
  Trash2,
  Lock,
  Globe,
  Check,
  Loader2,
  Plus,
  AlertCircle,
  GitMerge,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody, BoxRow, BoxFooter } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Dialog } from '@/components/ui/Dialog'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import {
  type RepoInfo,
  type BranchItem,
  type CollaboratorItem,
  getBranches,
  updateRepoSettings,
  getCollaborators,
  addCollaborator,
  removeCollaborator,
  deleteRepository,
  defaultRepo,
} from '@/lib/api'

export interface RepoSettingsViewProps {
  repo?: RepoInfo
  onRepoUpdated?: () => void
  onRepoDeleted?: () => void
}

export const RepoSettingsView: React.FC<RepoSettingsViewProps> = ({
  repo = defaultRepo,
  onRepoUpdated,
  onRepoDeleted,
}) => {
  const owner = repo.ownerSlug || 'acme'
  const repoName = repo.name || 'rocket'

  const [activeSection, setActiveSection] = useState<'general' | 'collaborators' | 'danger'>('general')

  // General Settings State
  const [branches, setBranches] = useState<BranchItem[]>([])
  const [defaultBranch, setDefaultBranch] = useState(repo.defaultBranch || 'main')
  const [allowSquash, setAllowSquash] = useState(true)
  const [allowMergeCommit, setAllowMergeCommit] = useState(true)
  const [deleteBranchOnMerge, setDeleteBranchOnMerge] = useState(false)
  const [savingGeneral, setSavingGeneral] = useState(false)
  const [generalFeedback, setGeneralFeedback] = useState<string | null>(null)

  // Collaborators State
  const [collaborators, setCollaborators] = useState<CollaboratorItem[]>([])
  const [loadingCollabs, setLoadingCollabs] = useState(false)
  const [addCollabOpen, setAddCollabOpen] = useState(false)
  const [newCollabLogin, setNewCollabLogin] = useState('')
  const [newCollabRole, setNewCollabRole] = useState<'write' | 'read' | 'admin'>('write')
  const [addingCollab, setAddingCollab] = useState(false)
  const [collabError, setCollabError] = useState<string | null>(null)

  // Danger Zone / Visibility & Deletion State
  const [isPrivate, setIsPrivate] = useState(Boolean(repo.isPrivate))
  const [togglingVisibility, setTogglingVisibility] = useState(false)
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [confirmDeleteInput, setConfirmDeleteInput] = useState('')
  const [deletingRepo, setDeletingRepo] = useState(false)
  const [dangerError, setDangerError] = useState<string | null>(null)

  useEffect(() => {
    getBranches(owner, repoName).then((data) => {
      setBranches(data)
      if (data.length > 0 && !defaultBranch) {
        setDefaultBranch(data[0].name)
      }
    }).catch(() => {})

    loadCollaborators()
  }, [owner, repoName])

  const loadCollaborators = async () => {
    setLoadingCollabs(true)
    try {
      const data = await getCollaborators(owner, repoName)
      setCollaborators(data)
    } finally {
      setLoadingCollabs(false)
    }
  }

  const handleSaveGeneral = async (e: React.FormEvent) => {
    e.preventDefault()
    setSavingGeneral(true)
    setGeneralFeedback(null)
    try {
      await updateRepoSettings(owner, repoName, {
        default_branch: defaultBranch,
        allow_squash_merge: allowSquash,
        allow_merge_commit: allowMergeCommit,
        delete_branch_on_merge: deleteBranchOnMerge,
      })
      setGeneralFeedback('Repository general settings updated successfully.')
      if (onRepoUpdated) onRepoUpdated()
      setTimeout(() => setGeneralFeedback(null), 3500)
    } catch (err: any) {
      setGeneralFeedback(err.message || 'Failed to update settings')
    } finally {
      setSavingGeneral(false)
    }
  }

  const handleToggleVisibility = async () => {
    setTogglingVisibility(true)
    setDangerError(null)
    const newVis = isPrivate ? 'public' : 'private'
    try {
      await updateRepoSettings(owner, repoName, {
        visibility: newVis,
      })
      setIsPrivate(!isPrivate)
      if (onRepoUpdated) onRepoUpdated()
    } catch (err: any) {
      setDangerError(err.message || 'Failed to toggle visibility')
    } finally {
      setTogglingVisibility(false)
    }
  }

  const handleAddCollaborator = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newCollabLogin.trim()) return
    setAddingCollab(true)
    setCollabError(null)
    try {
      const created = await addCollaborator(owner, repoName, newCollabLogin.trim(), newCollabRole)
      setCollaborators((prev) => [...prev, created])
      setAddCollabOpen(false)
      setNewCollabLogin('')
    } catch (err: any) {
      setCollabError(err.message || 'Failed to add collaborator')
    } finally {
      setAddingCollab(false)
    }
  }

  const handleRemoveCollaborator = async (login: string) => {
    if (!confirm(`Are you sure you want to remove @${login} from this repository?`)) return
    try {
      await removeCollaborator(owner, repoName, login)
      setCollaborators((prev) => prev.filter((c) => c.login !== login))
    } catch (err: any) {
      alert(err.message || 'Failed to remove collaborator')
    }
  }

  const handleDeleteRepo = async () => {
    const requiredConfirmation = `${owner}/${repoName}`
    if (confirmDeleteInput.trim() !== requiredConfirmation) {
      setDangerError(`Confirmation string does not match "${requiredConfirmation}"`)
      return
    }

    setDeletingRepo(true)
    setDangerError(null)
    try {
      await deleteRepository(owner, repoName)
      setDeleteModalOpen(false)
      if (onRepoDeleted) {
        onRepoDeleted()
      }
    } catch (err: any) {
      setDangerError(err.message || 'Failed to delete repository')
    } finally {
      setDeletingRepo(false)
    }
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-6 items-start">
      {/* Left Navigation Sidebar */}
      <aside className="md:col-span-1 space-y-1">
        <div className="px-3 pb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          Repository Settings
        </div>
        <nav className="space-y-0.5">
          <button
            onClick={() => setActiveSection('general')}
            className={`w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer ${
              activeSection === 'general'
                ? 'bg-secondary text-foreground font-semibold border-l-2 border-primary'
                : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
            }`}
          >
            <Settings className="w-4 h-4 shrink-0 text-muted-foreground" />
            <span>General</span>
          </button>

          <button
            onClick={() => setActiveSection('collaborators')}
            className={`w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer ${
              activeSection === 'collaborators'
                ? 'bg-secondary text-foreground font-semibold border-l-2 border-primary'
                : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
            }`}
          >
            <Users className="w-4 h-4 shrink-0 text-muted-foreground" />
            <span>Collaborators & Access</span>
          </button>

          <button
            onClick={() => setActiveSection('danger')}
            className={`w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer ${
              activeSection === 'danger'
                ? 'bg-destructive/10 text-destructive font-semibold border-l-2 border-destructive'
                : 'text-muted-foreground hover:bg-secondary/50 hover:text-destructive'
            }`}
          >
            <AlertTriangle className="w-4 h-4 shrink-0 text-destructive" />
            <span>Danger Zone</span>
          </button>
        </nav>
      </aside>

      {/* Main Settings Panel */}
      <section className="md:col-span-3 space-y-6">
        {/* 1. General Section */}
        {activeSection === 'general' && (
          <form onSubmit={handleSaveGeneral} className="space-y-6">
            <Box>
              <BoxHeader className="bg-secondary/35 py-3">
                <div>
                  <h3 className="text-sm font-semibold text-foreground">General Repository Configuration</h3>
                  <p className="text-xs text-muted-foreground">Manage default branch, naming, and merge strategies.</p>
                </div>
              </BoxHeader>
              <BoxBody className="p-4 space-y-5">
                {generalFeedback && (
                  <div
                    className={`p-3 rounded-md text-xs flex items-center gap-2 ${
                      generalFeedback.includes('successfully')
                        ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-500'
                        : 'bg-destructive/10 border border-destructive/30 text-destructive'
                    }`}
                  >
                    {generalFeedback.includes('successfully') ? (
                      <Check className="w-4 h-4 shrink-0" />
                    ) : (
                      <AlertCircle className="w-4 h-4 shrink-0" />
                    )}
                    <span>{generalFeedback}</span>
                  </div>
                )}

                {/* Repository Name & Slug */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">Owner</label>
                    <input
                      type="text"
                      disabled
                      value={owner}
                      className="w-full h-8 px-3 text-xs bg-secondary/50 border border-border rounded-md text-muted-foreground font-mono outline-none"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">Repository Name</label>
                    <input
                      type="text"
                      disabled
                      value={repoName}
                      className="w-full h-8 px-3 text-xs bg-secondary/50 border border-border rounded-md text-muted-foreground font-mono outline-none"
                    />
                  </div>
                </div>

                {/* Default Branch */}
                <div className="space-y-1 pt-2 border-t border-border">
                  <label className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                    <GitBranch className="w-3.5 h-3.5 text-primary" />
                    <span>Default branch</span>
                  </label>
                  <p className="text-xs text-muted-foreground">
                    The default branch is considered the base branch in your repository against which all pull requests and code commits are automatically matched.
                  </p>
                  <div className="max-w-xs pt-1">
                    <select
                      value={defaultBranch}
                      onChange={(e) => setDefaultBranch(e.target.value)}
                      className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                    >
                      {branches.map((b) => (
                        <option key={b.name} value={b.name}>
                          {b.name}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                {/* Pull Request Merge Options */}
                <div className="space-y-3 pt-3 border-t border-border">
                  <label className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                    <GitMerge className="w-3.5 h-3.5 text-primary" />
                    <span>Pull Request Merge Buttons</span>
                  </label>

                  <div className="space-y-2 text-xs">
                    <label className="flex items-start gap-2.5 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        checked={allowSquash}
                        onChange={(e) => setAllowSquash(e.target.checked)}
                        className="mt-0.5 rounded border-border text-primary focus:ring-ring"
                      />
                      <div>
                        <span className="font-semibold text-foreground block">Allow squash merging</span>
                        <span className="text-muted-foreground">Combine all commits from the head branch into a single commit in the base branch.</span>
                      </div>
                    </label>

                    <label className="flex items-start gap-2.5 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        checked={allowMergeCommit}
                        onChange={(e) => setAllowMergeCommit(e.target.checked)}
                        className="mt-0.5 rounded border-border text-primary focus:ring-ring"
                      />
                      <div>
                        <span className="font-semibold text-foreground block">Allow merge commits</span>
                        <span className="text-muted-foreground">Add all commits from the head branch to the base branch with a 3-way merge commit.</span>
                      </div>
                    </label>

                    <label className="flex items-start gap-2.5 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        checked={deleteBranchOnMerge}
                        onChange={(e) => setDeleteBranchOnMerge(e.target.checked)}
                        className="mt-0.5 rounded border-border text-primary focus:ring-ring"
                      />
                      <div>
                        <span className="font-semibold text-foreground block">Automatically delete head branches</span>
                        <span className="text-muted-foreground">Pull requests that are merged will automatically clean up their head branch reference.</span>
                      </div>
                    </label>
                  </div>
                </div>
              </BoxBody>
              <BoxFooter className="bg-secondary/15 py-3 justify-end">
                <Button type="submit" variant="primary" size="sm" disabled={savingGeneral} className="gap-1.5 text-xs">
                  {savingGeneral && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                  <span>Save changes</span>
                </Button>
              </BoxFooter>
            </Box>
          </form>
        )}

        {/* 2. Collaborators Section */}
        {activeSection === 'collaborators' && (
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Manage Access & Collaborators</h3>
                <p className="text-xs text-muted-foreground">Control which team members have direct push or admin grants to this repository.</p>
              </div>
              <Button
                variant="primary"
                size="sm"
                onClick={() => setAddCollabOpen(true)}
                className="gap-1.5 text-xs shrink-0"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add collaborator</span>
              </Button>
            </div>

            <Box>
              <BoxHeader className="bg-secondary/35 py-2.5">
                <span className="text-xs font-semibold text-foreground flex items-center gap-2">
                  <span>Collaborators with access ({collaborators.length})</span>
                  {loadingCollabs && <Loader2 className="w-3 h-3 animate-spin text-muted-foreground" />}
                </span>
              </BoxHeader>
              <BoxBody className="p-0 divide-y divide-border">
                {collaborators.map((c) => (
                  <BoxRow key={c.id} className="p-3 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <LivingAvatar name={c.login} isAgent={false} size="sm" />
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-foreground">@{c.login}</span>
                          <Badge variant="outline" className="text-[10px] py-0 px-1.5">
                            {c.role}
                          </Badge>
                        </div>
                        <span className="text-[11px] text-muted-foreground font-mono">{c.name}</span>
                      </div>
                    </div>

                    {c.id !== 'owner' && (
                      <Button
                        variant="outline"
                        size="xs"
                        onClick={() => handleRemoveCollaborator(c.login)}
                        className="text-destructive hover:bg-destructive/10 border-destructive/30 text-xs"
                      >
                        Remove
                      </Button>
                    )}
                  </BoxRow>
                ))}
              </BoxBody>
            </Box>
          </div>
        )}

        {/* 3. Danger Zone Section */}
        {activeSection === 'danger' && (
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-foreground">Danger Zone</h3>
              <p className="text-xs text-muted-foreground">Irreversible actions that affect repository visibility or permanently destroy data.</p>
            </div>

            {dangerError && (
              <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{dangerError}</span>
              </div>
            )}

            <div className="border border-destructive/40 rounded-md divide-y divide-destructive/30 overflow-hidden bg-card">
              {/* Change Visibility Row */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="space-y-1">
                  <div className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                    {isPrivate ? <Lock className="w-3.5 h-3.5 text-amber-500" /> : <Globe className="w-3.5 h-3.5 text-emerald-500" />}
                    <span>Change repository visibility</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    This repository is currently <strong className="text-foreground">{isPrivate ? 'Private' : 'Public'}</strong>.
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleVisibility}
                  disabled={togglingVisibility}
                  className="shrink-0 text-xs text-foreground hover:bg-secondary border-border"
                >
                  {togglingVisibility ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    `Change to ${isPrivate ? 'Public' : 'Private'}`
                  )}
                </Button>
              </div>

              {/* Delete Repository Row */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-destructive/5">
                <div className="space-y-1">
                  <div className="text-xs font-semibold text-destructive flex items-center gap-1.5">
                    <Trash2 className="w-3.5 h-3.5" />
                    <span>Delete this repository</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Once you delete a repository, there is no going back. All Git commits, trees, refs, issues, and WAL database records will be permanently removed.
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setConfirmDeleteInput('')
                    setDangerError(null)
                    setDeleteModalOpen(true)
                  }}
                  className="shrink-0 text-xs text-destructive border-destructive/40 hover:bg-destructive hover:text-destructive-foreground transition-colors"
                >
                  Delete this repository
                </Button>
              </div>
            </div>
          </div>
        )}
      </section>

      {/* Add Collaborator Dialog */}
      <Dialog
        isOpen={addCollabOpen}
        onClose={() => setAddCollabOpen(false)}
        title="Add collaborator"
        description="Grant access to a member of your organization."
      >
        <form onSubmit={handleAddCollaborator} className="space-y-4 pt-1">
          {collabError && (
            <div className="p-2.5 bg-destructive/10 border border-destructive/30 rounded text-xs text-destructive flex items-center gap-2">
              <AlertCircle className="w-3.5 h-3.5 shrink-0" />
              <span>{collabError}</span>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">User Login Username</label>
            <input
              type="text"
              required
              value={newCollabLogin}
              onChange={(e) => setNewCollabLogin(e.target.value)}
              placeholder="e.g. alex"
              className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Permission Role</label>
            <select
              value={newCollabRole}
              onChange={(e) => setNewCollabRole(e.target.value as any)}
              className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
            >
              <option value="read">Read (Can view and clone)</option>
              <option value="write">Write (Can clone, commit, push, and open PRs)</option>
              <option value="admin">Admin (Full administrative permissions)</option>
            </select>
          </div>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" size="sm" onClick={() => setAddCollabOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="primary" size="sm" disabled={addingCollab || !newCollabLogin.trim()}>
              {addingCollab ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : 'Add collaborator'}
            </Button>
          </div>
        </form>
      </Dialog>

      {/* Delete Repository Confirmation Dialog */}
      <Dialog
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Delete repository confirmation"
      >
        <div className="space-y-4 pt-1">
          <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive space-y-1">
            <p className="font-semibold flex items-center gap-1.5">
              <AlertTriangle className="w-4 h-4 shrink-0" />
              <span>This action cannot be undone.</span>
            </p>
            <p>
              This will permanently delete the <strong>{owner}/{repoName}</strong> repository, Git bare directories on disk, all commit histories, issues, and branches.
            </p>
          </div>

          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">
              To confirm deletion, please type <strong className="text-foreground font-mono">{owner}/{repoName}</strong>:
            </label>
            <input
              type="text"
              value={confirmDeleteInput}
              onChange={(e) => setConfirmDeleteInput(e.target.value)}
              placeholder={`${owner}/${repoName}`}
              className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground font-mono outline-none focus:border-destructive focus:ring-1 focus:ring-destructive"
            />
          </div>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" size="sm" onClick={() => setDeleteModalOpen(false)}>
              Cancel
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleDeleteRepo}
              disabled={deletingRepo || confirmDeleteInput.trim() !== `${owner}/${repoName}`}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90 border-transparent text-xs"
            >
              {deletingRepo ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : 'I understand, delete this repository'}
            </Button>
          </div>
        </div>
      </Dialog>
    </div>
  )
}
