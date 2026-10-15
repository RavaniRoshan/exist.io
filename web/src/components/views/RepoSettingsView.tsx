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
  Archive,
  Edit3,
  KeyRound,
  Webhook,
  Send,
  RefreshCw,
  ChevronRight,
  ArrowRightLeft,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody, BoxRow, BoxFooter } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Dialog } from '@/components/ui/Dialog'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { ChangeVisibilityModal } from '@/components/modals/ChangeVisibilityModal'
import { ArchiveRepoModal } from '@/components/modals/ArchiveRepoModal'
import { RenameRepoModal } from '@/components/modals/RenameRepoModal'
import { TransferRepoModal } from '@/components/modals/TransferRepoModal'
import { DeleteRepoModal } from '@/components/modals/DeleteRepoModal'
import {
  type RepoInfo,
  type BranchItem,
  type CollaboratorItem,
  type DeployKeyItem,
  type RepoWebhookItem,
  type RepoWebhookDeliveryItem,
  getBranches,
  updateRepoSettings,
  getCollaborators,
  addCollaborator,
  removeCollaborator,
  defaultRepo,
  getDeployKeys,
  createDeployKey,
  deleteDeployKey,
  getRepoWebhooks,
  createRepoWebhook,
  deleteRepoWebhook,
  testRepoWebhook,
  getRepoWebhookDeliveries,
  redeliverRepoWebhook,
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

  const [activeSection, setActiveSection] = useState<'general' | 'collaborators' | 'deploy-keys' | 'webhooks' | 'danger'>('general')

  // General Settings State
  const [branches, setBranches] = useState<BranchItem[]>([])
  const [defaultBranch, setDefaultBranch] = useState(repo.defaultBranch || 'main')
  const [repoDescription, setRepoDescription] = useState(repo.description || '')
  const [savingDescription, setSavingDescription] = useState(false)
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

  // Deploy Keys State
  const [deployKeys, setDeployKeys] = useState<DeployKeyItem[]>([])
  const [loadingDeployKeys, setLoadingDeployKeys] = useState(false)
  const [addDeployKeyOpen, setAddDeployKeyOpen] = useState(false)
  const [newDeployKeyTitle, setNewDeployKeyTitle] = useState('')
  const [newDeployKeyContent, setNewDeployKeyContent] = useState('')
  const [newDeployKeyReadOnly, setNewDeployKeyReadOnly] = useState(true)
  const [addingDeployKey, setAddingDeployKey] = useState(false)
  const [deployKeyError, setDeployKeyError] = useState<string | null>(null)

  // Webhooks State
  const [webhooks, setWebhooks] = useState<RepoWebhookItem[]>([])
  const [loadingWebhooks, setLoadingWebhooks] = useState(false)
  const [addWebhookOpen, setAddWebhookOpen] = useState(false)
  const [newWebhookUrl, setNewWebhookUrl] = useState('')
  const [newWebhookContentType, setNewWebhookContentType] = useState('json')
  const [newWebhookSecret, setNewWebhookSecret] = useState('')
  const [newWebhookEventsMode, setNewWebhookEventsMode] = useState<'push' | 'all' | 'custom'>('push')
  const [newWebhookCustomEvents, setNewWebhookCustomEvents] = useState<string[]>(['push', 'pull_request'])
  const [newWebhookActive, setNewWebhookActive] = useState(true)
  const [addingWebhook, setAddingWebhook] = useState(false)
  const [webhookError, setWebhookError] = useState<string | null>(null)
  const [selectedHookDeliveries, setSelectedHookDeliveries] = useState<{ hook: RepoWebhookItem; deliveries: RepoWebhookDeliveryItem[] } | null>(null)
  const [loadingDeliveries, setLoadingDeliveries] = useState(false)
  const [testingHookId, setTestingHookId] = useState<string | null>(null)
  const [redeliveringDelId, setRedeliveringDelId] = useState<string | null>(null)
  const [webhookFeedback, setWebhookFeedback] = useState<string | null>(null)

  // Danger Zone / Visibility, Archival, Rename, and Deletion Modals State
  const [isPrivate, setIsPrivate] = useState(Boolean(repo.isPrivate))
  const [isArchived, setIsArchived] = useState(Boolean(repo.isArchived))
  const [changeVisModalOpen, setChangeVisModalOpen] = useState(false)
  const [archiveModalOpen, setArchiveModalOpen] = useState(false)
  const [renameModalOpen, setRenameModalOpen] = useState(false)
  const [transferModalOpen, setTransferModalOpen] = useState(false)
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

  useEffect(() => {
    setIsPrivate(Boolean(repo.isPrivate))
    setIsArchived(Boolean(repo.isArchived))
    setRepoDescription(repo.description || '')
    setDefaultBranch(repo.defaultBranch || 'main')
  }, [repo])

  useEffect(() => {
    getBranches(owner, repoName).then((data) => {
      setBranches(data)
      if (data.length > 0 && !defaultBranch) {
        setDefaultBranch(data[0].name)
      }
    }).catch(() => {})

    loadCollaborators()
  }, [owner, repoName])

  useEffect(() => {
    if (activeSection === 'deploy-keys') {
      loadDeployKeys()
    } else if (activeSection === 'webhooks') {
      loadWebhooks()
    }
  }, [activeSection])

  const loadCollaborators = async () => {
    setLoadingCollabs(true)
    try {
      const data = await getCollaborators(owner, repoName)
      setCollaborators(data)
    } finally {
      setLoadingCollabs(false)
    }
  }

  const loadDeployKeys = async () => {
    setLoadingDeployKeys(true)
    try {
      const data = await getDeployKeys(owner, repoName)
      setDeployKeys(data)
    } finally {
      setLoadingDeployKeys(false)
    }
  }

  const handleAddDeployKey = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newDeployKeyContent.trim()) return
    setAddingDeployKey(true)
    setDeployKeyError(null)
    try {
      const key = await createDeployKey(owner, repoName, {
        title: newDeployKeyTitle.trim() || 'Deploy Key',
        key: newDeployKeyContent.trim(),
        read_only: newDeployKeyReadOnly,
      })
      setDeployKeys([key, ...deployKeys])
      setNewDeployKeyTitle('')
      setNewDeployKeyContent('')
      setNewDeployKeyReadOnly(true)
      setAddDeployKeyOpen(false)
    } catch (err: any) {
      setDeployKeyError(err.message || 'Failed to add deploy key')
    } finally {
      setAddingDeployKey(false)
    }
  }

  const handleDeleteDeployKey = async (id: string) => {
    try {
      await deleteDeployKey(owner, repoName, id)
      setDeployKeys(deployKeys.filter((k) => k.id !== id))
    } catch (err: any) {
      alert(err.message || 'Failed to delete deploy key')
    }
  }

  const loadWebhooks = async () => {
    setLoadingWebhooks(true)
    try {
      const data = await getRepoWebhooks(owner, repoName)
      setWebhooks(data)
    } finally {
      setLoadingWebhooks(false)
    }
  }

  const handleAddWebhook = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newWebhookUrl.trim()) return
    setAddingWebhook(true)
    setWebhookError(null)
    try {
      let events = ['push']
      if (newWebhookEventsMode === 'all') {
        events = ['push', 'pull_request', 'issues', 'release']
      } else if (newWebhookEventsMode === 'custom') {
        events = newWebhookCustomEvents
      }
      const hook = await createRepoWebhook(owner, repoName, {
        url: newWebhookUrl.trim(),
        content_type: newWebhookContentType,
        secret: newWebhookSecret.trim(),
        events,
        active: newWebhookActive,
      })
      setWebhooks([hook, ...webhooks])
      setNewWebhookUrl('')
      setNewWebhookSecret('')
      setAddWebhookOpen(false)
    } catch (err: any) {
      setWebhookError(err.message || 'Failed to add webhook')
    } finally {
      setAddingWebhook(false)
    }
  }

  const handleDeleteWebhook = async (id: string) => {
    try {
      await deleteRepoWebhook(owner, repoName, id)
      setWebhooks(webhooks.filter((w) => w.id !== id))
      if (selectedHookDeliveries?.hook.id === id) {
        setSelectedHookDeliveries(null)
      }
    } catch (err: any) {
      alert(err.message || 'Failed to delete webhook')
    }
  }

  const handleTestWebhook = async (id: string) => {
    setTestingHookId(id)
    try {
      await testRepoWebhook(owner, repoName, id)
      setWebhookFeedback('Test ping dispatched successfully')
      setTimeout(() => setWebhookFeedback(null), 3000)
      if (selectedHookDeliveries?.hook.id === id) {
        const delivs = await getRepoWebhookDeliveries(owner, repoName, id)
        setSelectedHookDeliveries({ hook: selectedHookDeliveries.hook, deliveries: delivs })
      }
    } catch (err: any) {
      alert('Test ping failed: ' + err.message)
    } finally {
      setTestingHookId(null)
    }
  }

  const handleOpenDeliveries = async (hook: RepoWebhookItem) => {
    setLoadingDeliveries(true)
    try {
      const delivs = await getRepoWebhookDeliveries(owner, repoName, hook.id)
      setSelectedHookDeliveries({ hook, deliveries: delivs })
    } finally {
      setLoadingDeliveries(false)
    }
  }

  const handleRedeliver = async (hookId: string, delId: string) => {
    setRedeliveringDelId(delId)
    try {
      await redeliverRepoWebhook(owner, repoName, hookId, delId)
      setWebhookFeedback('Redelivery dispatched successfully')
      setTimeout(() => setWebhookFeedback(null), 3000)
      const delivs = await getRepoWebhookDeliveries(owner, repoName, hookId)
      if (selectedHookDeliveries) {
        setSelectedHookDeliveries({ hook: selectedHookDeliveries.hook, deliveries: delivs })
      }
    } catch (err: any) {
      alert('Redelivery failed: ' + err.message)
    } finally {
      setRedeliveringDelId(null)
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

  const handleSaveDescription = async () => {
    setSavingDescription(true)
    setGeneralFeedback(null)
    try {
      await updateRepoSettings(owner, repoName, {
        description: repoDescription.trim(),
      })
      setGeneralFeedback('Repository description saved successfully.')
      if (onRepoUpdated) onRepoUpdated()
      setTimeout(() => setGeneralFeedback(null), 3500)
    } catch (err: any) {
      setGeneralFeedback(err.message || 'Failed to update description')
    } finally {
      setSavingDescription(false)
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
            onClick={() => setActiveSection('deploy-keys')}
            className={`w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer ${
              activeSection === 'deploy-keys'
                ? 'bg-secondary text-foreground font-semibold border-l-2 border-primary'
                : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
            }`}
          >
            <KeyRound className="w-4 h-4 shrink-0 text-muted-foreground" />
            <span>Deploy Keys</span>
          </button>

          <button
            onClick={() => setActiveSection('webhooks')}
            className={`w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer ${
              activeSection === 'webhooks'
                ? 'bg-secondary text-foreground font-semibold border-l-2 border-primary'
                : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
            }`}
          >
            <Webhook className="w-4 h-4 shrink-0 text-muted-foreground" />
            <span>Webhooks</span>
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
                    <div className="flex items-center justify-between">
                      <label className="text-xs font-semibold text-foreground">Repository Name</label>
                      <button
                        type="button"
                        data-testid="repo-rename-open-btn"
                        onClick={() => setRenameModalOpen(true)}
                        className="text-[11px] text-primary hover:underline flex items-center gap-1"
                      >
                        <Edit3 className="w-3 h-3" />
                        <span>Rename</span>
                      </button>
                    </div>
                    <input
                      type="text"
                      disabled
                      value={repoName}
                      className="w-full h-8 px-3 text-xs bg-secondary/50 border border-border rounded-md text-foreground font-mono outline-none"
                    />
                  </div>
                </div>

                {/* Repository Description */}
                <div className="space-y-1.5 pt-2 border-t border-border">
                  <label className="text-xs font-semibold text-foreground">Description</label>
                  <p className="text-xs text-muted-foreground">Short summary of this repository displayed across forge lists and headers.</p>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={repoDescription}
                      onChange={(e) => setRepoDescription(e.target.value)}
                      placeholder="Enterprise repository on exist.io"
                      className="flex-1 h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={handleSaveDescription}
                      disabled={savingDescription || repoDescription === (repo.description || '')}
                    >
                      {savingDescription ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : 'Save description'}
                    </Button>
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

        {/* 3. Deploy Keys Section */}
        {activeSection === 'deploy-keys' && (
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Deploy Keys</h3>
                <p className="text-xs text-muted-foreground">
                  Deploy keys grant SSH access to fetch or push code to this repository without personal account credentials.
                </p>
              </div>
              <Button
                variant="primary"
                size="sm"
                data-testid="add-deploy-key-btn"
                onClick={() => {
                  setNewDeployKeyTitle('')
                  setNewDeployKeyContent('')
                  setNewDeployKeyReadOnly(true)
                  setDeployKeyError(null)
                  setAddDeployKeyOpen(true)
                }}
                className="gap-1.5 text-xs shrink-0"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add deploy key</span>
              </Button>
            </div>

            <Box>
              <BoxHeader className="bg-secondary/35 py-3">
                <div className="flex items-center justify-between w-full">
                  <span className="text-xs font-semibold text-foreground">Enrolled SSH Deploy Keys</span>
                  <Badge variant="outline" className="text-[11px] py-0 px-1.5">
                    {deployKeys.length} active
                  </Badge>
                </div>
              </BoxHeader>
              <BoxBody className="divide-y divide-border">
                {loadingDeployKeys ? (
                  <div className="p-8 text-center text-xs text-muted-foreground flex items-center justify-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin text-primary" />
                    <span>Loading deploy keys...</span>
                  </div>
                ) : deployKeys.length === 0 ? (
                  <div className="p-8 text-center text-xs text-muted-foreground space-y-2">
                    <div className="w-9 h-9 rounded-full bg-secondary/80 flex items-center justify-center mx-auto text-muted-foreground">
                      <KeyRound className="w-4 h-4" />
                    </div>
                    <p className="font-semibold text-foreground">There are no deploy keys for this repository.</p>
                    <p className="max-w-sm mx-auto text-muted-foreground">
                      Deploy keys allow external build servers and CI/CD pipelines to securely clone or push code.
                    </p>
                  </div>
                ) : (
                  deployKeys.map((k) => (
                    <BoxRow key={k.id} className="p-3.5 flex items-center justify-between gap-4">
                      <div className="flex items-start gap-3 min-w-0">
                        <div className="w-8 h-8 rounded-md bg-secondary flex items-center justify-center shrink-0 mt-0.5 text-muted-foreground">
                          <KeyRound className="w-4 h-4 text-primary" />
                        </div>
                        <div className="space-y-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-foreground truncate">{k.title}</span>
                            <Badge
                              variant="outline"
                              className={`text-[10px] py-0 px-1.5 ${
                                k.read_only
                                  ? 'border-border text-muted-foreground'
                                  : 'border-emerald-500/40 text-emerald-400 bg-emerald-500/10'
                              }`}
                            >
                              {k.read_only ? 'Read-only' : 'Read/write'}
                            </Badge>
                          </div>
                          <div className="text-[11px] font-mono text-muted-foreground truncate">{k.fingerprint}</div>
                          <div className="text-[10px] text-muted-foreground flex items-center gap-3">
                            <span>Added {k.created_at}</span>
                            <span>•</span>
                            <span>{k.last_used_at ? `Last used ${k.last_used_at}` : 'Never used'}</span>
                          </div>
                        </div>
                      </div>

                      <Button
                        variant="outline"
                        size="xs"
                        onClick={() => handleDeleteDeployKey(k.id)}
                        className="text-destructive hover:bg-destructive/10 border-destructive/30 text-xs shrink-0"
                      >
                        Delete
                      </Button>
                    </BoxRow>
                  ))
                )}
              </BoxBody>
            </Box>
          </div>
        )}

        {/* 4. Repository Webhooks Section */}
        {activeSection === 'webhooks' && (
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Webhooks</h3>
                <p className="text-xs text-muted-foreground">
                  Webhooks allow external services to be notified when events occur in this repository via HTTP POST payloads.
                </p>
              </div>
              <Button
                variant="primary"
                size="sm"
                data-testid="add-webhook-btn"
                onClick={() => {
                  setNewWebhookUrl('')
                  setNewWebhookSecret('')
                  setNewWebhookContentType('json')
                  setNewWebhookEventsMode('push')
                  setNewWebhookActive(true)
                  setWebhookError(null)
                  setAddWebhookOpen(true)
                }}
                className="gap-1.5 text-xs shrink-0"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add webhook</span>
              </Button>
            </div>

            {webhookFeedback && (
              <div className="p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-md text-xs text-emerald-400 flex items-center gap-2">
                <Check className="w-4 h-4 shrink-0" />
                <span>{webhookFeedback}</span>
              </div>
            )}

            <Box>
              <BoxHeader className="bg-secondary/35 py-3">
                <div className="flex items-center justify-between w-full">
                  <span className="text-xs font-semibold text-foreground">Configured Repository Webhooks</span>
                  <Badge variant="outline" className="text-[11px] py-0 px-1.5">
                    {webhooks.length} registered
                  </Badge>
                </div>
              </BoxHeader>
              <BoxBody className="divide-y divide-border">
                {loadingWebhooks ? (
                  <div className="p-8 text-center text-xs text-muted-foreground flex items-center justify-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin text-primary" />
                    <span>Loading webhooks...</span>
                  </div>
                ) : webhooks.length === 0 ? (
                  <div className="p-8 text-center text-xs text-muted-foreground space-y-2">
                    <div className="w-9 h-9 rounded-full bg-secondary/80 flex items-center justify-center mx-auto text-muted-foreground">
                      <Webhook className="w-4 h-4" />
                    </div>
                    <p className="font-semibold text-foreground">There are no webhooks for this repository.</p>
                    <p className="max-w-sm mx-auto text-muted-foreground">
                      Webhooks send JSON HTTP POST requests to configured URLs upon pushes, pull requests, issues, and releases.
                    </p>
                  </div>
                ) : (
                  webhooks.map((w) => (
                    <BoxRow key={w.id} className="p-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                      <div className="flex items-start gap-3 min-w-0">
                        <div className="mt-1">
                          <span
                            className={`inline-block w-2.5 h-2.5 rounded-full ${
                              w.active ? 'bg-emerald-400' : 'bg-muted-foreground'
                            }`}
                            title={w.active ? 'Webhook active' : 'Webhook disabled'}
                          />
                        </div>
                        <div className="space-y-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-foreground font-mono truncate">{w.url}</span>
                            <Badge variant="outline" className="text-[10px] py-0 px-1.5 uppercase font-mono">
                              {w.content_type}
                            </Badge>
                          </div>
                          <div className="text-[11px] text-muted-foreground flex items-center gap-2">
                            <span>Events:</span>
                            <code className="text-[10px] bg-secondary/80 px-1.5 py-0.5 rounded border border-border">
                              {w.events}
                            </code>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
                        <Button
                          variant="secondary"
                          size="xs"
                          disabled={testingHookId === w.id}
                          onClick={() => handleTestWebhook(w.id)}
                          className="gap-1 text-xs"
                          title="Dispatch test ping event"
                        >
                          {testingHookId === w.id ? (
                            <Loader2 className="w-3 h-3 animate-spin" />
                          ) : (
                            <Send className="w-3 h-3 text-primary" />
                          )}
                          <span>Ping</span>
                        </Button>

                        <Button
                          variant="secondary"
                          size="xs"
                          onClick={() => handleOpenDeliveries(w)}
                          className="gap-1 text-xs"
                        >
                          <span>Deliveries</span>
                          <ChevronRight className="w-3 h-3 text-muted-foreground" />
                        </Button>

                        <Button
                          variant="outline"
                          size="xs"
                          onClick={() => handleDeleteWebhook(w.id)}
                          className="text-destructive hover:bg-destructive/10 border-destructive/30 text-xs"
                        >
                          <Trash2 className="w-3 h-3" />
                        </Button>
                      </div>
                    </BoxRow>
                  ))
                )}
              </BoxBody>
            </Box>

            {/* Selected Webhook Deliveries Drawer */}
            {selectedHookDeliveries && (
              <Box className="border-primary/40 mt-4">
                <BoxHeader className="bg-primary/5 py-3 flex items-center justify-between">
                  <div className="space-y-0.5 min-w-0">
                    <h4 className="text-xs font-semibold text-foreground flex items-center gap-2">
                      <Webhook className="w-3.5 h-3.5 text-primary" />
                      <span>Deliveries for {selectedHookDeliveries.hook.url}</span>
                    </h4>
                    <p className="text-[11px] text-muted-foreground">Recent HTTP payload delivery history and status codes</p>
                  </div>
                  <Button
                    variant="outline"
                    size="xs"
                    onClick={() => setSelectedHookDeliveries(null)}
                  >
                    Close
                  </Button>
                </BoxHeader>
                <BoxBody className="divide-y divide-border">
                  {loadingDeliveries ? (
                    <div className="p-6 text-center text-xs text-muted-foreground flex items-center justify-center gap-2">
                      <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
                      <span>Loading deliveries...</span>
                    </div>
                  ) : selectedHookDeliveries.deliveries.length === 0 ? (
                    <div className="p-6 text-center text-xs text-muted-foreground">
                      No deliveries have been recorded for this webhook yet. Click &quot;Ping&quot; to dispatch a test payload.
                    </div>
                  ) : (
                    selectedHookDeliveries.deliveries.map((del) => (
                      <BoxRow key={del.id} className="p-3 flex items-center justify-between gap-4 text-xs">
                        <div className="flex items-center gap-2.5 min-w-0">
                          <Badge
                            variant="outline"
                            className={`text-[10px] font-mono py-0 px-1.5 ${
                              del.response_status_code >= 200 && del.response_status_code < 300
                                ? 'border-emerald-500/40 text-emerald-400 bg-emerald-500/10'
                                : 'border-destructive/40 text-destructive bg-destructive/10'
                            }`}
                          >
                            {del.response_status_code || 'ERR'}
                          </Badge>
                          <span className="font-semibold text-foreground">{del.event}</span>
                          <span className="font-mono text-[11px] text-muted-foreground truncate">{del.guid}</span>
                          <span className="text-[10px] text-muted-foreground">({del.duration_ms}ms)</span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          <span className="text-[10px] text-muted-foreground">{del.created_at}</span>
                          <Button
                            variant="secondary"
                            size="xs"
                            disabled={redeliveringDelId === del.id}
                            onClick={() => handleRedeliver(selectedHookDeliveries.hook.id, del.id)}
                            className="gap-1 text-[11px]"
                            title="Re-send this exact payload"
                          >
                            {redeliveringDelId === del.id ? (
                              <Loader2 className="w-3 h-3 animate-spin" />
                            ) : (
                              <RefreshCw className="w-3 h-3 text-muted-foreground" />
                            )}
                            <span>Redeliver</span>
                          </Button>
                        </div>
                      </BoxRow>
                    ))
                  )}
                </BoxBody>
              </Box>
            )}
          </div>
        )}

        {/* 5. Danger Zone Section */}
        {activeSection === 'danger' && (
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-foreground">Danger Zone</h3>
              <p className="text-xs text-muted-foreground">Irreversible actions that affect repository visibility or permanently destroy data.</p>
            </div>

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
                  data-testid="repo-change-visibility-btn"
                  onClick={() => setChangeVisModalOpen(true)}
                  className="shrink-0 text-xs text-foreground hover:bg-secondary border-border"
                >
                  Change visibility
                </Button>
              </div>

              {/* Archive Repository Row */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="space-y-1">
                  <div className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                    <Archive className="w-3.5 h-3.5 text-amber-500" />
                    <span>{isArchived ? 'Unarchive this repository' : 'Archive this repository'}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {isArchived
                      ? 'This repository is currently archived and read-only. Unarchiving restores push and issue operations.'
                      : 'Mark this repository as archived and read-only. Commits, issues, and pull requests will be locked.'}
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  data-testid="repo-archive-btn"
                  onClick={() => setArchiveModalOpen(true)}
                  className="shrink-0 text-xs text-amber-500 hover:bg-amber-500/10 border-amber-500/30"
                >
                  {isArchived ? 'Unarchive this repository' : 'Archive this repository'}
                </Button>
              </div>

              {/* Transfer Repository Row */}
              <div className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="space-y-1">
                  <div className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                    <ArrowRightLeft className="w-3.5 h-3.5 text-amber-500" />
                    <span>Transfer ownership</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Transfer this repository to another user or to an organization where you have admin rights.
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  data-testid="repo-transfer-btn"
                  onClick={() => setTransferModalOpen(true)}
                  className="shrink-0 text-xs text-amber-500 hover:bg-amber-500/10 border-amber-500/30"
                >
                  Transfer ownership
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
                  data-testid="repo-delete-btn"
                  onClick={() => setDeleteModalOpen(true)}
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
      <DeleteRepoModal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        owner={owner}
        repoName={repoName}
        onSuccess={() => {
          setDeleteModalOpen(false)
          if (onRepoDeleted) {
            onRepoDeleted()
          }
        }}
      />

      {/* Change Visibility Modal */}
      <ChangeVisibilityModal
        isOpen={changeVisModalOpen}
        onClose={() => setChangeVisModalOpen(false)}
        owner={owner}
        repoName={repoName}
        isCurrentlyPrivate={isPrivate}
        onSuccess={() => {
          setIsPrivate(!isPrivate)
          if (onRepoUpdated) onRepoUpdated()
        }}
      />

      {/* Archive Repository Modal */}
      <ArchiveRepoModal
        isOpen={archiveModalOpen}
        onClose={() => setArchiveModalOpen(false)}
        owner={owner}
        repoName={repoName}
        isCurrentlyArchived={isArchived}
        onSuccess={() => {
          setIsArchived(!isArchived)
          if (onRepoUpdated) onRepoUpdated()
        }}
      />

      {/* Rename Repository Modal */}
      <RenameRepoModal
        isOpen={renameModalOpen}
        onClose={() => setRenameModalOpen(false)}
        owner={owner}
        currentName={repoName}
        onSuccess={(newName) => {
          if (onRepoUpdated) onRepoUpdated()
          window.location.hash = `#/${owner}/${newName}/settings`
        }}
      />

      {/* Transfer Repository Modal */}
      <TransferRepoModal
        isOpen={transferModalOpen}
        onClose={() => setTransferModalOpen(false)}
        owner={owner}
        repoName={repoName}
        onSuccess={(newOwner, targetRepo) => {
          if (onRepoUpdated) onRepoUpdated()
          window.location.hash = `#/${newOwner}/${targetRepo}/settings`
        }}
      />
      {/* Add Deploy Key Dialog */}
      <Dialog
        isOpen={addDeployKeyOpen}
        onClose={() => setAddDeployKeyOpen(false)}
        title="Add deploy key"
        description="Deploy keys grant SSH access to fetch or push code to this repository directly."
      >
        <form onSubmit={handleAddDeployKey} className="space-y-4 pt-1">
          {deployKeyError && (
            <div className="p-2.5 bg-destructive/10 border border-destructive/30 rounded text-xs text-destructive flex items-center gap-2">
              <AlertCircle className="w-3.5 h-3.5 shrink-0" />
              <span>{deployKeyError}</span>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Title</label>
            <input
              type="text"
              required
              data-testid="deploy-key-title-input"
              value={newDeployKeyTitle}
              onChange={(e) => setNewDeployKeyTitle(e.target.value)}
              placeholder="e.g. Production CI Runner"
              className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Key</label>
            <textarea
              required
              rows={4}
              data-testid="deploy-key-content-input"
              value={newDeployKeyContent}
              onChange={(e) => setNewDeployKeyContent(e.target.value)}
              placeholder="Begins with 'ssh-rsa', 'ssh-ed25519', or 'ecdsa-sha2-nistp256'"
              className="w-full p-2.5 text-xs bg-input border border-border rounded-md text-foreground font-mono placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring resize-none"
            />
          </div>

          <label className="flex items-start gap-2.5 cursor-pointer text-xs text-foreground select-none">
            <input
              type="checkbox"
              data-testid="deploy-key-write-checkbox"
              checked={!newDeployKeyReadOnly}
              onChange={(e) => setNewDeployKeyReadOnly(!e.target.checked)}
              className="mt-0.5 rounded border-border text-primary focus:ring-ring"
            />
            <div className="space-y-0.5">
              <span className="font-semibold block text-foreground">Allow write access</span>
              <span className="text-muted-foreground text-[11px] leading-tight block">
                If not selected, this key will only be able to fetch and clone code. Select this if the deploy key needs to push commits.
              </span>
            </div>
          </label>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" size="sm" onClick={() => setAddDeployKeyOpen(false)}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              data-testid="deploy-key-submit-btn"
              disabled={addingDeployKey || !newDeployKeyContent.trim()}
            >
              {addingDeployKey ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : 'Add key'}
            </Button>
          </div>
        </form>
      </Dialog>

      {/* Add Webhook Dialog */}
      <Dialog
        isOpen={addWebhookOpen}
        onClose={() => setAddWebhookOpen(false)}
        title="Add webhook"
        description="We'll send a POST request to the URL below with details of any subscribed events."
      >
        <form onSubmit={handleAddWebhook} className="space-y-4 pt-1">
          {webhookError && (
            <div className="p-2.5 bg-destructive/10 border border-destructive/30 rounded text-xs text-destructive flex items-center gap-2">
              <AlertCircle className="w-3.5 h-3.5 shrink-0" />
              <span>{webhookError}</span>
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-semibold text-foreground">Payload URL</label>
            <input
              type="url"
              required
              data-testid="webhook-url-input"
              value={newWebhookUrl}
              onChange={(e) => setNewWebhookUrl(e.target.value)}
              placeholder="https://example.com/postreceive"
              className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Content type</label>
              <select
                value={newWebhookContentType}
                onChange={(e) => setNewWebhookContentType(e.target.value)}
                className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
              >
                <option value="json">application/json</option>
                <option value="form">application/x-www-form-urlencoded</option>
              </select>
            </div>

            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">Secret (HMAC-SHA256)</label>
              <input
                type="password"
                data-testid="webhook-secret-input"
                value={newWebhookSecret}
                onChange={(e) => setNewWebhookSecret(e.target.value)}
                placeholder="Optional webhook secret"
                className="w-full h-8 px-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring font-mono"
              />
            </div>
          </div>

          <div className="space-y-2 pt-1 border-t border-border">
            <label className="text-xs font-semibold text-foreground">Which events would you like to trigger this webhook?</label>
            <div className="space-y-2 text-xs">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="eventsMode"
                  checked={newWebhookEventsMode === 'push'}
                  onChange={() => setNewWebhookEventsMode('push')}
                  className="text-primary focus:ring-ring"
                />
                <span>Just the push event</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="eventsMode"
                  checked={newWebhookEventsMode === 'all'}
                  onChange={() => setNewWebhookEventsMode('all')}
                  className="text-primary focus:ring-ring"
                />
                <span>Send me everything (pushes, pull requests, issues, releases)</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="eventsMode"
                  checked={newWebhookEventsMode === 'custom'}
                  onChange={() => setNewWebhookEventsMode('custom')}
                  className="text-primary focus:ring-ring"
                />
                <span>Let me select individual events</span>
              </label>

              {newWebhookEventsMode === 'custom' && (
                <div className="pl-6 pt-1 space-y-2 border-l-2 border-border ml-2">
                  {[
                    { id: 'push', label: 'Pushes', desc: 'Git push to a repository.' },
                    { id: 'pull_request', label: 'Pull requests', desc: 'Pull request opened, closed, or merged.' },
                    { id: 'issues', label: 'Issues', desc: 'Issues opened, closed, or labeled.' },
                    { id: 'release', label: 'Releases', desc: 'Release created, published, or edited.' },
                  ].map((evt) => {
                    const checked = newWebhookCustomEvents.includes(evt.id)
                    return (
                      <label key={evt.id} className="flex items-start gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setNewWebhookCustomEvents([...newWebhookCustomEvents, evt.id])
                            } else {
                              setNewWebhookCustomEvents(newWebhookCustomEvents.filter((x) => x !== evt.id))
                            }
                          }}
                          className="mt-0.5 rounded border-border text-primary focus:ring-ring"
                        />
                        <div>
                          <span className="font-semibold text-foreground">{evt.label}</span>
                          <span className="text-[11px] text-muted-foreground block">{evt.desc}</span>
                        </div>
                      </label>
                    )
                  })}
                </div>
              )}
            </div>
          </div>

          <label className="flex items-center gap-2 cursor-pointer text-xs text-foreground select-none pt-1">
            <input
              type="checkbox"
              checked={newWebhookActive}
              onChange={(e) => setNewWebhookActive(e.target.checked)}
              className="rounded border-border text-primary focus:ring-ring"
            />
            <span>Active (Deliver event details when this hook is triggered)</span>
          </label>

          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button type="button" variant="outline" size="sm" onClick={() => setAddWebhookOpen(false)}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              data-testid="webhook-submit-btn"
              disabled={addingWebhook || !newWebhookUrl.trim()}
            >
              {addingWebhook ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : 'Add webhook'}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
