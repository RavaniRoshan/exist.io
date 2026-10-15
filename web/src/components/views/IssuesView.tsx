import React, { useState, useEffect } from 'react'
import {
  CircleDot,
  CheckCircle2,
  MessageSquare,
  Search,
  Plus,
  ArrowLeft,
  Bot,
  Sparkles,
  Send,
  GitPullRequest,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge, StateLabel } from '@/components/ui/Badge'
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
import { Dialog } from '@/components/ui/Dialog'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import {
  getIssues,
  createIssue,
  addIssueComment,
  toggleIssueState,
  delegateIssueToAgent,
  type RepoInfo,
  type Issue,
} from '@/lib/api'
import { cn } from '@/lib/utils'

export interface IssuesViewProps {
  repo?: RepoInfo
  onSelectPullRequest?: (prNumber: number) => void
}

export const IssuesView: React.FC<IssuesViewProps> = ({ repo, onSelectPullRequest }) => {
  const owner = repo?.ownerSlug || 'shiva'
  const repoName = repo?.name || 'exist.io'

  const [issues, setIssues] = useState<Issue[]>([])
  const [loading, setLoading] = useState(true)
  const [activeFilter, setActiveFilter] = useState<'open' | 'closed' | 'all'>('open')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedIssue, setSelectedIssue] = useState<Issue | null>(null)

  // New Issue Modal
  const [isNewIssueModalOpen, setIsNewIssueModalOpen] = useState(false)
  const [newTitle, setNewTitle] = useState('')
  const [newBody, setNewBody] = useState('')
  const [newLabels, setNewLabels] = useState<string[]>(['enhancement'])
  const [isSubmittingNew, setIsSubmittingNew] = useState(false)

  // Comment Thread
  const [newCommentBody, setNewCommentBody] = useState('')
  const [isSubmittingComment, setIsSubmittingComment] = useState(false)
  const [isDelegating, setIsDelegating] = useState(false)

  useEffect(() => {
    loadIssues()
  }, [owner, repoName])

  const loadIssues = async () => {
    setLoading(true)
    const data = await getIssues(owner, repoName)
    setIssues(data)
    setLoading(false)
  }

  const openIssuesCount = issues.filter((i) => i.state === 'open').length
  const closedIssuesCount = issues.filter((i) => i.state === 'closed').length

  const filteredIssues = issues.filter((issue) => {
    if (activeFilter === 'open' && issue.state !== 'open') return false
    if (activeFilter === 'closed' && issue.state !== 'closed') return false
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      const matchesTitle = issue.title.toLowerCase().includes(q)
      const matchesAuthor = issue.author.toLowerCase().includes(q)
      const matchesNum = issue.number.toString().includes(q)
      return matchesTitle || matchesAuthor || matchesNum
    }
    return true
  })

  const handleCreateNewIssue = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTitle.trim()) return

    setIsSubmittingNew(true)
    const created = await createIssue(owner, repoName, newTitle.trim(), newBody.trim(), newLabels)
    setIsSubmittingNew(false)

    setIssues((prev) => [created, ...prev])
    setIsNewIssueModalOpen(false)
    setNewTitle('')
    setNewBody('')
    setSelectedIssue(created)
  }

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedIssue || !newCommentBody.trim()) return

    setIsSubmittingComment(true)
    const comment = await addIssueComment(owner, repoName, selectedIssue.number, newCommentBody.trim())
    setIsSubmittingComment(false)

    const updatedIssue = {
      ...selectedIssue,
      comments: [...(selectedIssue.comments || []), comment],
    }
    setSelectedIssue(updatedIssue)
    setIssues((prev) => prev.map((i) => (i.id === updatedIssue.id ? updatedIssue : i)))
    setNewCommentBody('')
  }

  const handleToggleState = async () => {
    if (!selectedIssue) return
    const nextState = await toggleIssueState(owner, repoName, selectedIssue.number, selectedIssue.state)
    const updated = {
      ...selectedIssue,
      state: nextState,
      closedAt: nextState === 'closed' ? 'Just now' : undefined,
    }
    setSelectedIssue(updated)
    setIssues((prev) => prev.map((i) => (i.id === updated.id ? updated : i)))
  }

  const handleDelegateToAgent = async () => {
    if (!selectedIssue) return
    setIsDelegating(true)
    const res = await delegateIssueToAgent(owner, repoName, selectedIssue.number)
    setIsDelegating(false)

    if (res.ok && res.comment) {
      const updatedIssue = {
        ...selectedIssue,
        comments: [...(selectedIssue.comments || []), res.comment],
      }
      setSelectedIssue(updatedIssue)
      setIssues((prev) => prev.map((i) => (i.id === updatedIssue.id ? updatedIssue : i)))
    }
  }

  return (
    <div className="space-y-4">
      {/* 1. ISSUE DETAIL VIEW */}
      {selectedIssue ? (
        <div className="space-y-6">
          {/* Header & Back Button */}
          <div className="space-y-3 pb-4 border-b border-border">
            <button
              onClick={() => setSelectedIssue(null)}
              className="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground font-medium cursor-pointer"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              <span>Back to all issues</span>
            </button>

            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <h1 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight leading-snug">
                {selectedIssue.title}{' '}
                <span className="text-muted-foreground font-mono font-normal">
                  #{selectedIssue.number}
                </span>
              </h1>

              <div className="flex items-center gap-2">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={handleToggleState}
                  className="gap-1.5"
                >
                  {selectedIssue.state === 'open' ? (
                    <>
                      <CheckCircle2 className="w-3.5 h-3.5 text-purple-400" />
                      <span>Close issue</span>
                    </>
                  ) : (
                    <>
                      <CircleDot className="w-3.5 h-3.5 text-emerald-400" />
                      <span>Reopen issue</span>
                    </>
                  )}
                </Button>
                <Button
                  size="sm"
                  onClick={() => setIsNewIssueModalOpen(true)}
                  className="gap-1.5"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>New issue</span>
                </Button>
              </div>
            </div>

            {/* State Pill & Meta */}
            <div className="flex items-center gap-2.5 text-xs text-muted-foreground flex-wrap">
              <StateLabel state={selectedIssue.state}>
                {selectedIssue.state === 'open' ? 'Open' : 'Closed'}
              </StateLabel>
              <span>
                <strong className="text-foreground">{selectedIssue.author}</strong> opened this issue {selectedIssue.createdAt} • {(selectedIssue.comments || []).length} comments
              </span>
            </div>
          </div>

          {/* Two-Column Detail Layout */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
            {/* Left Thread (9 cols) */}
            <div className="lg:col-span-9 space-y-4">
              {/* Initial Issue Description Box */}
              <Box>
                <BoxHeader className="bg-secondary/25 py-2">
                  <div className="flex items-center gap-2">
                    <LivingAvatar name={selectedIssue.author} isAgent={false} size="sm" />
                    <span className="font-semibold text-foreground text-xs">{selectedIssue.author}</span>
                    <span className="text-muted-foreground text-xs">opened {selectedIssue.createdAt}</span>
                  </div>
                  <Badge variant="outline" className="text-[10px]">
                    Author
                  </Badge>
                </BoxHeader>
                <BoxBody className="p-4 text-xs sm:text-sm text-foreground leading-relaxed whitespace-pre-wrap">
                  {selectedIssue.body}
                </BoxBody>
              </Box>

              {/* Discussion Thread */}
              {(selectedIssue.comments || []).map((comment) => (
                <Box key={comment.id}>
                  <BoxHeader className={cn('py-2', comment.isAgent ? 'bg-primary/10 border-b border-primary/20' : 'bg-secondary/25')}>
                    <div className="flex items-center gap-2">
                      <LivingAvatar
                        name={comment.author}
                        isAgent={comment.isAgent}
                        status={comment.isAgent ? 'online' : 'offline'}
                        size="sm"
                      />
                      <span className="font-semibold text-foreground text-xs">{comment.author}</span>
                      <span className="text-muted-foreground text-xs">commented {comment.createdAt}</span>
                    </div>
                    {comment.isAgent ? (
                      <Badge variant="outline" className="border-primary/40 text-primary bg-primary/10 text-[10px] font-semibold flex items-center gap-1">
                        <Bot className="w-3 h-3" />
                        <span>Autonomous Agent</span>
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="text-[10px]">
                        Contributor
                      </Badge>
                    )}
                  </BoxHeader>
                  <BoxBody className="p-4 text-xs sm:text-sm text-foreground leading-relaxed whitespace-pre-wrap font-sans">
                    {comment.body}
                  </BoxBody>
                </Box>
              ))}

              {/* Autonomous Agent Dispatch Action Box */}
              {selectedIssue.state === 'open' && (
                <div className="rounded-md border border-primary/30 bg-primary/5 p-4 flex flex-col sm:flex-row items-center justify-between gap-3 shadow-2xs">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-md bg-primary text-primary-foreground flex items-center justify-center shrink-0 shadow-2xs">
                      <Bot className="w-4 h-4" />
                    </div>
                    <div>
                      <div className="text-xs font-semibold text-foreground">
                        Autonomous Issue Triage with Exist Agent
                      </div>
                      <div className="text-[11px] text-muted-foreground">
                        Delegate reproduction, test synthesis, and draft pull request generation to the agent.
                      </div>
                    </div>
                  </div>
                  <Button
                    size="sm"
                    onClick={handleDelegateToAgent}
                    disabled={isDelegating}
                    className="gap-1.5 shrink-0 w-full sm:w-auto"
                  >
                    <Sparkles className="w-3.5 h-3.5" />
                    <span>{isDelegating ? 'Triaging...' : 'Delegate to Agent'}</span>
                  </Button>
                </div>
              )}

              {/* New Comment Form */}
              <Box>
                <form onSubmit={handleAddComment}>
                  <BoxHeader className="py-2 text-xs font-semibold text-foreground">
                    <span>Add a comment</span>
                  </BoxHeader>
                  <BoxBody className="p-4 space-y-3">
                    <textarea
                      placeholder="Leave a comment or markdown notes..."
                      value={newCommentBody}
                      onChange={(e) => setNewCommentBody(e.target.value)}
                      rows={3}
                      className="w-full text-xs p-3 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none font-sans"
                    />
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        type="submit"
                        size="sm"
                        disabled={isSubmittingComment || !newCommentBody.trim()}
                        className="gap-1.5"
                      >
                        <Send className="w-3.5 h-3.5" />
                        <span>{isSubmittingComment ? 'Posting...' : 'Comment'}</span>
                      </Button>
                    </div>
                  </BoxBody>
                </form>
              </Box>
            </div>

            {/* Right Meta Sidebar (3 cols) */}
            <div className="lg:col-span-3 space-y-4">
              {/* Assignees Box */}
              <Box>
                <BoxHeader className="py-2 text-xs font-semibold text-foreground">
                  <span>Assignees</span>
                </BoxHeader>
                <BoxBody className="p-3 text-xs space-y-2">
                  {(selectedIssue.assignees || []).length === 0 ? (
                    <span className="text-muted-foreground text-xs">No one assigned</span>
                  ) : (
                    (selectedIssue.assignees || []).map((assignee) => (
                      <div key={assignee} className="flex items-center gap-2 text-foreground font-medium">
                        <LivingAvatar name={assignee} isAgent={assignee.toLowerCase().includes('agent')} size="sm" />
                        <span>{assignee}</span>
                      </div>
                    ))
                  )}
                </BoxBody>
              </Box>

              {/* Labels Box */}
              <Box>
                <BoxHeader className="py-2 text-xs font-semibold text-foreground">
                  <span>Labels</span>
                </BoxHeader>
                <BoxBody className="p-3 flex flex-wrap gap-1.5">
                  {(selectedIssue.labels || []).map((lbl) => (
                    <span
                      key={lbl.name}
                      className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium border"
                      style={{
                        backgroundColor: `${lbl.color}20`,
                        color: lbl.color,
                        borderColor: `${lbl.color}40`,
                      }}
                    >
                      {lbl.name}
                    </span>
                  ))}
                </BoxBody>
              </Box>

              {/* Development Reference Box */}
              <Box>
                <BoxHeader className="py-2 text-xs font-semibold text-foreground">
                  <span>Development</span>
                </BoxHeader>
                <BoxBody className="p-3 text-xs text-muted-foreground space-y-2">
                  <div className="flex items-center gap-2 text-primary font-medium cursor-pointer hover:underline" onClick={() => onSelectPullRequest?.(4)}>
                    <GitPullRequest className="w-3.5 h-3.5" />
                    <span>Pull Request #4</span>
                  </div>
                  <div className="text-[11px] text-muted-foreground">
                    Linked to branch <code className="font-mono text-foreground">feat/enterprise-auth</code>
                  </div>
                </BoxBody>
              </Box>
            </div>
          </div>
        </div>
      ) : (
        /* 2. ISSUES LIST VIEW */
        <div className="space-y-4">
          {/* Controls Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
            <div className="relative w-full sm:w-96">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
              <input
                type="text"
                placeholder="is:issue is:open "
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 text-xs rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
              />
            </div>

            <Button
              onClick={() => setIsNewIssueModalOpen(true)}
              size="sm"
              className="gap-1.5 shrink-0 w-full sm:w-auto"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>New issue</span>
            </Button>
          </div>

          {/* Issues Box Container */}
          <Box flush>
            <BoxHeader className="flex flex-wrap items-center justify-between gap-3 py-2.5">
              {/* Filter Tabs: Open vs Closed */}
              <div className="flex items-center gap-4 text-xs">
                <button
                  onClick={() => setActiveFilter('open')}
                  className={cn(
                    'flex items-center gap-1.5 cursor-pointer transition-colors',
                    activeFilter === 'open'
                      ? 'font-bold text-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  <CircleDot className="w-4 h-4 text-emerald-400" />
                  <span>{openIssuesCount} Open</span>
                </button>
                <button
                  onClick={() => setActiveFilter('closed')}
                  className={cn(
                    'flex items-center gap-1.5 cursor-pointer transition-colors',
                    activeFilter === 'closed'
                      ? 'font-bold text-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  <CheckCircle2 className="w-4 h-4 text-purple-400" />
                  <span>{closedIssuesCount} Closed</span>
                </button>
              </div>

              <div className="flex items-center gap-3 text-xs text-muted-foreground">
                <span className="hover:text-foreground cursor-pointer">Author ▾</span>
                <span className="hover:text-foreground cursor-pointer">Label ▾</span>
                <span className="hover:text-foreground cursor-pointer">Assignee ▾</span>
                <span className="hover:text-foreground cursor-pointer">Sort ▾</span>
              </div>
            </BoxHeader>

            <BoxBody>
              {loading ? (
                <div className="py-12 text-center text-xs text-muted-foreground">
                  Loading issues...
                </div>
              ) : filteredIssues.length === 0 ? (
                <div className="py-12 text-center text-xs text-muted-foreground space-y-2">
                  <CircleDot className="w-8 h-8 text-muted-foreground mx-auto stroke-1" />
                  <div>No issues match your current query.</div>
                </div>
              ) : (
                filteredIssues.map((issue) => (
                  <BoxRow
                    key={issue.id}
                    onClick={() => setSelectedIssue(issue)}
                    className="flex items-start justify-between gap-4 py-3 cursor-pointer group"
                  >
                    <div className="flex items-start gap-2.5 flex-1 min-w-0">
                      <div className="mt-0.5 shrink-0">
                        {issue.state === 'open' ? (
                          <CircleDot className="w-4 h-4 text-emerald-400" />
                        ) : (
                          <CheckCircle2 className="w-4 h-4 text-purple-400" />
                        )}
                      </div>

                      <div className="space-y-1 flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-semibold text-sm text-foreground group-hover:text-primary transition-colors">
                            {issue.title}
                          </span>
                          {(issue.labels || []).map((lbl) => (
                            <span
                              key={lbl.name}
                              className="inline-flex items-center px-2 py-0.2 rounded-full text-[10px] font-medium border"
                              style={{
                                backgroundColor: `${lbl.color}20`,
                                color: lbl.color,
                                borderColor: `${lbl.color}40`,
                              }}
                            >
                              {lbl.name}
                            </span>
                          ))}
                        </div>

                        <div className="text-xs text-muted-foreground font-sans">
                          #{issue.number} opened {issue.createdAt} by {issue.author}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-4 shrink-0 text-xs text-muted-foreground">
                      {(issue.assignees || []).length > 0 && (
                        <div className="hidden sm:flex -space-x-1.5 items-center">
                          {issue.assignees?.map((a) => (
                            <div
                              key={a}
                              className="w-5 h-5 rounded-full bg-secondary border border-border flex items-center justify-center font-bold text-[9px] text-foreground"
                              title={a}
                            >
                              {a.charAt(0).toUpperCase()}
                            </div>
                          ))}
                        </div>
                      )}

                      {(issue.comments || []).length > 0 && (
                        <div className="flex items-center gap-1">
                          <MessageSquare className="w-3.5 h-3.5" />
                          <span className="font-mono text-[11px]">{(issue.comments || []).length}</span>
                        </div>
                      )}
                    </div>
                  </BoxRow>
                ))
              )}
            </BoxBody>
          </Box>
        </div>
      )}

      {/* CREATE NEW ISSUE MODAL */}
      <Dialog
        isOpen={isNewIssueModalOpen}
        onClose={() => setIsNewIssueModalOpen(false)}
        title="Create new issue"
        description="File a bug report, request a feature, or assign a task to an autonomous agent."
      >
        <form onSubmit={handleCreateNewIssue} className="space-y-4 pt-1">
          <div className="space-y-3 text-xs">
            <div>
              <label className="block font-medium text-foreground mb-1">Title</label>
              <input
                type="text"
                placeholder="Title"
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                required
              />
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Description</label>
              <textarea
                placeholder="Leave a comment or describe the steps to reproduce..."
                value={newBody}
                onChange={(e) => setNewBody(e.target.value)}
                rows={5}
                className="w-full p-3 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none font-sans"
                required
              />
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Labels</label>
              <div className="flex flex-wrap gap-2">
                {['bug', 'enhancement', 'security', 'performance', 'agent-task'].map((lbl) => {
                  const isSelected = newLabels.includes(lbl)
                  return (
                    <button
                      type="button"
                      key={lbl}
                      onClick={() => {
                        if (isSelected) {
                          setNewLabels(newLabels.filter((l) => l !== lbl))
                        } else {
                          setNewLabels([...newLabels, lbl])
                        }
                      }}
                      className={cn(
                        'px-2.5 py-0.5 rounded-full text-xs font-medium border transition-colors cursor-pointer',
                        isSelected
                          ? 'bg-primary text-primary-foreground border-primary'
                          : 'bg-secondary text-muted-foreground border-border hover:text-foreground'
                      )}
                    >
                      {lbl}
                    </button>
                  )
                })}
              </div>
            </div>
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsNewIssueModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={isSubmittingNew || !newTitle.trim()}>
              {isSubmittingNew ? 'Submitting...' : 'Submit new issue'}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
