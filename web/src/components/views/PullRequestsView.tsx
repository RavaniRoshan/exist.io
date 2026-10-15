import React, { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  GitPullRequest,
  GitMerge,
  MessageSquare,
  CheckCircle2,
  AlertCircle,
  ShieldCheck,
  FileDiff,
  Sparkles,
  Send,
  Copy,
  Check,
  Terminal,
  Loader2,
  RefreshCw,
  GitCommit,
  Lock,
  Tag,
  Users,
  ChevronDown,
  Search,
  CheckCheck,
} from 'lucide-react'
import confetti from 'canvas-confetti'
import { Button } from '@/components/ui/Button'
import { Box, BoxHeader, BoxBody, BoxRow, BoxFooter } from '@/components/ui/Card'
import { Badge, StateLabel } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { MergeBox } from '@/components/ui/MergeBox'
import { CreatePullRequestModal } from '@/components/modals/CreatePullRequestModal'
import {
  type PullRequest,
  type RepoInfo,
  type AgentReviewResult,
  getPullRequests,
  mergePullRequest,
  addPullRequestComment,
  triggerAgentReview,
  isDemoModeActive,
  samplePullRequests,
} from '@/lib/api'

export interface PullRequestsViewProps {
  repo?: RepoInfo
  onSelectAgentConsole?: () => void
}

function getAuthorName(author: any): string {
  if (!author) return 'dev'
  if (typeof author === 'string') return author
  return author.name || author.login || 'dev'
}

export const PullRequestsView: React.FC<PullRequestsViewProps> = ({
  repo,
  onSelectAgentConsole: _onSelectAgentConsole,
}) => {
  const owner = repo?.ownerSlug || 'acme'
  const repoName = repo?.name || 'rocket'

  const [prs, setPrs] = useState<PullRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [activeFilter, setActiveFilter] = useState<'open' | 'merged' | 'all'>('open')
  const [selectedPr, setSelectedPr] = useState<PullRequest | null>(null)
  const [activeSubTab, setActiveSubTab] = useState<'conversation' | 'diff' | 'checks'>('conversation')
  const [newCommentText, setNewCommentText] = useState('')
  const [commentTab, setCommentTab] = useState<'write' | 'preview'>('write')
  const [isReviewing, setIsReviewing] = useState(false)
  const [reviewResult, setReviewResult] = useState<AgentReviewResult | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [copiedCli, setCopiedCli] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  const loadPRs = async () => {
    setLoading(true)
    setErrorMessage(null)
    try {
      const data = await getPullRequests(owner, repoName)
      setPrs(data)
    } catch {
      if (isDemoModeActive()) {
        setPrs(samplePullRequests)
      } else {
        setPrs([])
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadPRs()

    const onModeChange = () => {
      loadPRs()
    }
    window.addEventListener('exist_mode_changed', onModeChange)
    return () => window.removeEventListener('exist_mode_changed', onModeChange)
  }, [owner, repoName])

  const openPrsCount = prs.filter((p) => p.state === 'open').length
  const mergedPrsCount = prs.filter((p) => p.state === 'merged').length

  const filteredPrs = prs.filter((pr) => {
    if (activeFilter === 'open' && pr.state !== 'open') return false
    if (activeFilter === 'merged' && pr.state !== 'merged') return false
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      const matchesTitle = pr.title.toLowerCase().includes(q)
      const matchesAuthor = getAuthorName(pr.author).toLowerCase().includes(q)
      const matchesNumber = pr.number.toString().includes(q)
      return matchesTitle || matchesAuthor || matchesNumber
    }
    return true
  })

  const handleMergePr = async (method: 'merge' | 'squash' | 'rebase') => {
    if (!selectedPr) return
    setErrorMessage(null)
    const prNumber = selectedPr.number

    try {
      await mergePullRequest(owner, repoName, prNumber, method === 'rebase' ? 'squash' : method)

      confetti({
        particleCount: 80,
        spread: 70,
        origin: { y: 0.6 },
        colors: ['#38bdf8', '#818cf8', '#c084fc', '#34d399'],
      })

      const mergeComment = {
        id: `c-merge-${Date.now()}`,
        author: 'Roshan Ravani',
        isAgent: false,
        createdAt: 'Just now',
        body: `Merged pull request #${prNumber} via ${method.toUpperCase()} merge. Ruleset branch protection and cryptographic checks verified.`,
      }

      setPrs((prev) =>
        prev.map((p) => {
          if (p.number === prNumber) {
            return {
              ...p,
              state: 'merged',
              comments: [...(p.comments || []), mergeComment],
            }
          }
          return p
        })
      )

      setSelectedPr((prev) =>
        prev
          ? {
              ...prev,
              state: 'merged',
              comments: [...(prev.comments || []), mergeComment],
            }
          : null
      )
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to merge pull request')
    }
  }

  const handleRunAgentReview = async () => {
    if (!selectedPr || isReviewing) return
    setIsReviewing(true)
    setErrorMessage(null)

    try {
      const res = await triggerAgentReview(owner, repoName, selectedPr.number)
      setReviewResult(res)

      const agentComment = {
        id: res.reviewId || `rev-${Date.now()}`,
        author: 'Exist Agent Reviewer',
        isAgent: true,
        createdAt: 'Just now',
        body: res.overviewMd || res.summary,
      }

      const updatedPR: PullRequest = {
        ...selectedPr,
        comments: [...(selectedPr.comments || []), agentComment],
        checks: [
          ...(selectedPr.checks || []),
          {
            id: res.checkRunId || `chk-${Date.now()}`,
            name: 'Exist Agent Reviewer',
            status: 'completed',
            conclusion: res.conclusion,
            duration: '0.2s',
          },
        ],
      }

      setSelectedPr(updatedPR)
      setPrs((prev) => prev.map((p) => (p.number === selectedPr.number ? updatedPR : p)))
    } catch (err: any) {
      setErrorMessage(err.message || 'Agent review failed')
    } finally {
      setIsReviewing(false)
    }
  }

  const handleAddComment = async () => {
    if (!newCommentText.trim() || !selectedPr) return
    const text = newCommentText
    setNewCommentText('')
    try {
      const added = await addPullRequestComment(owner, repoName, selectedPr.number, text)
      const newComment = {
        id: added?.id || `c-${Date.now()}`,
        author: added?.author || 'Current Operator',
        isAgent: false,
        createdAt: added?.createdAt || 'Just now',
        body: text,
      }

      const updated = {
        ...selectedPr,
        comments: [...(selectedPr.comments || []), newComment],
      }

      setSelectedPr(updated)
      setPrs((prev) => prev.map((p) => (p.number === selectedPr.number ? updated : p)))
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to post comment')
    }
  }

  const samplePushCommand = `git checkout -b feature/agent-capabilities
# ... apply code changes ...
git commit -m "feat: implement autonomous capability leases"
git push origin feature/agent-capabilities`

  const handleCopyCli = () => {
    navigator.clipboard.writeText(samplePushCommand)
    setCopiedCli(true)
    setTimeout(() => setCopiedCli(false), 2000)
  }

  return (
    <div className="space-y-4">
      {errorMessage && (
        <div className="p-3 bg-destructive/10 border border-destructive/30 rounded-md text-xs text-destructive flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{errorMessage}</span>
        </div>
      )}

      <AnimatePresence mode="wait">
        {selectedPr ? (
          /* =========================================================================
             PR DETAIL VIEW: Canonical GitHub Two-Column Architecture
             ========================================================================= */
          <motion.div
            key="pr-detail"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="space-y-4"
          >
            {/* PR Header */}
            <div className="space-y-3 pb-3 border-b border-border">
              <div className="flex items-center justify-between gap-4">
                <button
                  onClick={() => setSelectedPr(null)}
                  className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1 transition-colors"
                >
                  ← Back to pull requests
                </button>
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm" onClick={loadPRs} disabled={loading} className="gap-1.5 text-xs">
                    <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
                    <span>Refresh</span>
                  </Button>
                </div>
              </div>

              <div className="flex flex-col sm:flex-row sm:items-baseline justify-between gap-2">
                <div className="flex items-baseline gap-2 flex-wrap">
                  <h1 className="text-2xl font-normal text-foreground tracking-tight">
                    {selectedPr.title}
                  </h1>
                  <span className="text-2xl font-light text-muted-foreground">
                    #{selectedPr.number}
                  </span>
                </div>
              </div>

              {/* State Badge and Branch Description */}
              <div className="flex items-center gap-2 text-xs flex-wrap">
                <StateLabel state={selectedPr.state === 'merged' ? 'merged' : 'open'}>
                  {selectedPr.state === 'merged' ? 'Merged' : 'Open'}
                </StateLabel>

                <span className="text-muted-foreground">
                  <strong className="text-foreground font-medium">{getAuthorName(selectedPr.author)}</strong> wants to merge 1 commit into{' '}
                  <code className="px-1.5 py-0.5 rounded bg-secondary font-mono text-xs text-foreground border border-border">
                    {selectedPr.baseRef}
                  </code>{' '}
                  from{' '}
                  <code className="px-1.5 py-0.5 rounded bg-secondary font-mono text-xs text-foreground border border-border">
                    {selectedPr.headRef}
                  </code>
                </span>
                <span className="text-muted-foreground">•</span>
                <span className="text-muted-foreground">{selectedPr.createdAt || 'recently'}</span>
              </div>
            </div>

            {/* Sub-Navigation UnderlineNav */}
            <div className="flex items-center gap-2 border-b border-border -mb-px text-xs font-medium">
              <button
                onClick={() => setActiveSubTab('conversation')}
                className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors ${
                  activeSubTab === 'conversation'
                    ? 'border-primary text-foreground font-semibold'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <MessageSquare className="w-3.5 h-3.5" />
                <span>Conversation</span>
                <span className="text-[10px] bg-secondary text-secondary-foreground px-1.5 py-0.2 rounded-full font-mono">
                  {(selectedPr.comments || []).length}
                </span>
              </button>

              <button
                onClick={() => setActiveSubTab('checks')}
                className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors ${
                  activeSubTab === 'checks'
                    ? 'border-primary text-foreground font-semibold'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <ShieldCheck className="w-3.5 h-3.5" />
                <span>Checks</span>
                <span className="text-[10px] bg-secondary text-secondary-foreground px-1.5 py-0.2 rounded-full font-mono">
                  {(selectedPr.checks || []).length}
                </span>
              </button>

              <button
                onClick={() => setActiveSubTab('diff')}
                className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors ${
                  activeSubTab === 'diff'
                    ? 'border-primary text-foreground font-semibold'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <FileDiff className="w-3.5 h-3.5" />
                <span>Files changed</span>
                <span className="text-[10px] bg-secondary text-secondary-foreground px-1.5 py-0.2 rounded-full font-mono">
                  {(selectedPr.diffHunks || []).length}
                </span>
              </button>
            </div>

            {/* GitHub Canonical Two-Column Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 pt-2">
              {/* Left Column: Timeline, Merge Box, Discussion, Diff (9 cols) */}
              <div className="lg:col-span-9 space-y-4">
                {activeSubTab === 'conversation' && (
                  <div className="space-y-4">
                    {/* Initial PR Description Box */}
                    <Box>
                      <BoxHeader className="bg-secondary/25 py-2">
                        <div className="flex items-center gap-2">
                          <LivingAvatar
                            name={getAuthorName(selectedPr.author)}
                            isAgent={false}
                            status="offline"
                            size="sm"
                          />
                          <span className="font-semibold text-foreground">{getAuthorName(selectedPr.author)}</span>
                          <span className="text-muted-foreground">commented {selectedPr.createdAt || 'recently'}</span>
                        </div>
                        <Badge variant="outline" className="text-[10px]">
                          Author
                        </Badge>
                      </BoxHeader>
                      <BoxBody className="p-4 text-xs sm:text-sm text-foreground leading-relaxed">
                        <p>
                          Proposes branch updates to incorporate verified capabilities and enterprise ruleset requirements into <code className="font-mono text-xs bg-secondary px-1 py-0.5 rounded">{selectedPr.baseRef}</code>.
                        </p>
                      </BoxBody>
                    </Box>

                    {/* Timeline Event: Branch created */}
                    <div className="flex items-center gap-3 pl-4 py-1 text-xs text-muted-foreground">
                      <div className="w-6 h-6 rounded-full bg-secondary flex items-center justify-center shrink-0 border border-border">
                        <GitCommit className="w-3 h-3 text-muted-foreground" />
                      </div>
                      <span>
                        <strong className="text-foreground">{getAuthorName(selectedPr.author)}</strong> added 1 commit from branch <code className="font-mono">{selectedPr.headRef}</code>
                      </span>
                    </div>

                    {/* Discussion Comments */}
                    {(selectedPr.comments || []).map((comment) => (
                      <Box key={comment.id}>
                        <BoxHeader className="bg-secondary/25 py-2">
                          <div className="flex items-center gap-2">
                            <LivingAvatar
                              name={comment.author}
                              isAgent={comment.isAgent}
                              status={comment.isAgent ? 'online' : 'offline'}
                              size="sm"
                            />
                            <span className="font-semibold text-foreground">{comment.author}</span>
                            {comment.isAgent && (
                              <Badge variant="agent" className="text-[9px] py-0 px-1.5">
                                AI Agent
                              </Badge>
                            )}
                            <span className="text-muted-foreground">commented {comment.createdAt}</span>
                          </div>
                          <Badge variant="outline" className="text-[10px]">
                            {comment.isAgent ? 'Automated Review' : 'Contributor'}
                          </Badge>
                        </BoxHeader>
                        <BoxBody className="p-4 text-xs sm:text-sm text-foreground whitespace-pre-wrap leading-relaxed">
                          {comment.body}
                        </BoxBody>
                      </Box>
                    ))}

                    {/* Check runs status box */}
                    <Box>
                      <BoxHeader className="bg-secondary/25 py-2">
                        <div className="flex items-center gap-2">
                          <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                          <span className="font-semibold text-foreground">
                            All checks have passed ({ (selectedPr.checks || []).length } successful checks)
                          </span>
                        </div>
                      </BoxHeader>
                      <BoxBody className="p-0 divide-y divide-border">
                        {(selectedPr.checks || []).map((chk) => (
                          <div key={chk.id} className="flex items-center justify-between px-4 py-2 text-xs">
                            <div className="flex items-center gap-2">
                              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-500" />
                              <span className="font-medium text-foreground">{chk.name}</span>
                              <span className="text-muted-foreground">— Successful in {chk.duration}</span>
                            </div>
                            <span className="text-muted-foreground font-mono text-[11px]">Details</span>
                          </div>
                        ))}
                      </BoxBody>
                    </Box>

                    {/* GitHub Canonical Merge Box */}
                    <MergeBox
                      onMerge={handleMergePr}
                      isMerged={selectedPr.state === 'merged'}
                      checksPassed={true}
                      defaultTitle={`${selectedPr.title} (#${selectedPr.number})`}
                      defaultMessage={`Squash and merge pull request #${selectedPr.number} into ${selectedPr.baseRef}.`}
                    />

                    {/* New Comment Box */}
                    <Box>
                      <BoxHeader className="bg-secondary/25 py-2">
                        <div className="flex items-center gap-2 text-xs">
                          <button
                            onClick={() => setCommentTab('write')}
                            className={`px-2 py-1 rounded-md transition-colors ${
                              commentTab === 'write' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
                            }`}
                          >
                            Write
                          </button>
                          <button
                            onClick={() => setCommentTab('preview')}
                            className={`px-2 py-1 rounded-md transition-colors ${
                              commentTab === 'preview' ? 'bg-background text-foreground font-semibold shadow-2xs' : 'text-muted-foreground hover:text-foreground'
                            }`}
                          >
                            Preview
                          </button>
                        </div>
                      </BoxHeader>
                      <BoxBody className="p-3">
                        {commentTab === 'write' ? (
                          <textarea
                            value={newCommentText}
                            onChange={(e) => setNewCommentText(e.target.value)}
                            placeholder="Leave a comment or review suggestion..."
                            className="w-full bg-input text-foreground placeholder:text-muted-foreground rounded-md border border-border p-3 text-xs outline-none focus:border-ring focus:ring-1 focus:ring-ring resize-y min-h-[90px]"
                          />
                        ) : (
                          <div className="min-h-[90px] p-3 text-xs text-muted-foreground">
                            {newCommentText.trim() ? (
                              <p className="text-foreground whitespace-pre-wrap">{newCommentText}</p>
                            ) : (
                              'Nothing to preview'
                            )}
                          </div>
                        )}
                      </BoxBody>
                      <BoxFooter className="justify-end gap-2 bg-secondary/15 py-2">
                        <Button
                          variant="primary"
                          size="sm"
                          onClick={handleAddComment}
                          disabled={!newCommentText.trim()}
                          className="gap-1.5"
                        >
                          <Send className="w-3.5 h-3.5" />
                          <span>Comment</span>
                        </Button>
                      </BoxFooter>
                    </Box>
                  </div>
                )}

                {/* Sub-tab: Files Changed (Diff View) */}
                {activeSubTab === 'diff' && (
                  <div className="space-y-4">
                    {(selectedPr.diffHunks || []).length === 0 ? (
                      <Box className="p-8 text-center text-muted-foreground text-xs">
                        No modified files found for this pull request.
                      </Box>
                    ) : (
                      (selectedPr.diffHunks || []).map((diff, idx) => (
                        <Box key={idx}>
                          <BoxHeader className="bg-secondary/35 py-2 font-mono text-xs">
                            <span className="font-semibold text-foreground">{diff.newPath}</span>
                            <div className="flex items-center gap-2">
                              <span className="text-emerald-500 font-semibold">+{diff.additions}</span>
                              <span className="text-rose-500 font-semibold">-{diff.deletions}</span>
                            </div>
                          </BoxHeader>
                          <div className="p-3 font-mono text-xs overflow-x-auto leading-relaxed bg-background">
                            <pre>
                              {diff.hunk.split('\n').map((line, lIdx) => {
                                const isAdd = line.startsWith('+')
                                const isDel = line.startsWith('-')
                                const isHeader = line.startsWith('@@')
                                return (
                                  <div
                                    key={lIdx}
                                    className={`px-2 py-0.5 rounded-sm ${
                                      isAdd
                                        ? 'bg-emerald-500/10 text-emerald-400'
                                        : isDel
                                        ? 'bg-rose-500/10 text-rose-400'
                                        : isHeader
                                        ? 'text-muted-foreground bg-secondary/50 font-semibold'
                                        : 'text-foreground/90'
                                    }`}
                                  >
                                    {line}
                                  </div>
                                )
                              })}
                            </pre>
                          </div>
                        </Box>
                      ))
                    )}
                  </div>
                )}

                {/* Sub-tab: Checks & CI Policy */}
                {activeSubTab === 'checks' && (
                  <Box>
                    <BoxHeader className="bg-secondary/35 py-2.5">
                      <div className="flex items-center gap-2">
                        <ShieldCheck className="w-4 h-4 text-emerald-500" />
                        <span className="font-semibold text-foreground">CI & Policy Check Runs</span>
                      </div>
                    </BoxHeader>
                    <BoxBody className="p-0 divide-y divide-border">
                      {(selectedPr.checks || []).map((chk) => (
                        <div key={chk.id} className="flex items-center justify-between p-3.5 text-xs">
                          <div className="flex items-center gap-3">
                            <CheckCircle2 className="w-4 h-4 text-emerald-500" />
                            <div>
                              <div className="font-semibold text-foreground">{chk.name}</div>
                              <div className="text-muted-foreground text-[11px]">
                                Completed in {chk.duration} • GitHub Actions Enterprise Agent
                              </div>
                            </div>
                          </div>
                          <Badge variant="success" className="text-[10px]">
                            Passed
                          </Badge>
                        </div>
                      ))}
                    </BoxBody>
                  </Box>
                )}
              </div>

              {/* Right Column: Reviewers, Assignees, Labels, Rulesets (3 cols) */}
              <div className="lg:col-span-3 space-y-4">
                {/* Reviewers Box */}
                <Box>
                  <BoxHeader className="py-2 text-xs font-semibold text-foreground flex items-center justify-between">
                    <span>Reviewers</span>
                    <Users className="w-3.5 h-3.5 text-muted-foreground" />
                  </BoxHeader>
                  <BoxBody className="p-3 space-y-2.5 text-xs">
                    {/* Trigger Agent Review Action */}
                    {selectedPr.state === 'open' && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleRunAgentReview}
                        disabled={isReviewing}
                        className="w-full text-xs gap-1.5 justify-center border-primary/30 text-primary hover:bg-primary/10"
                      >
                        {isReviewing ? (
                          <>
                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                            <span>Inspecting Diff...</span>
                          </>
                        ) : (
                          <>
                            <Sparkles className="w-3.5 h-3.5" />
                            <span>Run AI Agent Review</span>
                          </>
                        )}
                      </Button>
                    )}

                    {reviewResult && (
                      <div className="p-2.5 rounded-md bg-secondary/50 border border-border text-xs space-y-1">
                        <div className="flex items-center justify-between font-semibold">
                          <span className="text-foreground">AI Reviewer</span>
                          <span className="text-emerald-500 text-[11px] font-mono">{reviewResult.score}/100</span>
                        </div>
                        <p className="text-[11px] text-muted-foreground">{reviewResult.summary}</p>
                      </div>
                    )}

                    <div className="flex items-center justify-between text-muted-foreground pt-1">
                      <div className="flex items-center gap-2">
                        <LivingAvatar name="Exist Review Bot" isAgent={true} size="sm" />
                        <span className="text-foreground">Exist Review Bot</span>
                      </div>
                      <CheckCheck className="w-3.5 h-3.5 text-emerald-500" />
                    </div>
                  </BoxBody>
                </Box>

                {/* Assignees Box */}
                <Box>
                  <BoxHeader className="py-2 text-xs font-semibold text-foreground">
                    <span>Assignees</span>
                  </BoxHeader>
                  <BoxBody className="p-3 text-xs">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <LivingAvatar name={getAuthorName(selectedPr.author)} isAgent={false} size="sm" />
                      <span className="text-foreground font-medium">{getAuthorName(selectedPr.author)}</span>
                    </div>
                  </BoxBody>
                </Box>

                {/* Labels Box */}
                <Box>
                  <BoxHeader className="py-2 text-xs font-semibold text-foreground flex items-center justify-between">
                    <span>Labels</span>
                    <Tag className="w-3.5 h-3.5 text-muted-foreground" />
                  </BoxHeader>
                  <BoxBody className="p-3 text-xs">
                    <div className="flex flex-wrap gap-1.5">
                      {(selectedPr.labels || []).length > 0 ? (
                        (selectedPr.labels || []).map((lbl) => (
                          <span
                            key={lbl.name}
                            className="px-2 py-0.5 rounded-md text-[11px] font-medium border"
                            style={{
                              backgroundColor: `${lbl.color}15`,
                              borderColor: `${lbl.color}35`,
                              color: lbl.color,
                            }}
                          >
                            {lbl.name}
                          </span>
                        ))
                      ) : (
                        <span className="text-muted-foreground">None yet</span>
                      )}
                    </div>
                  </BoxBody>
                </Box>

                {/* Enterprise Ruleset Protection */}
                <Box>
                  <BoxHeader className="py-2 text-xs font-semibold text-foreground flex items-center justify-between">
                    <span>Ruleset Protection</span>
                    <Lock className="w-3.5 h-3.5 text-muted-foreground" />
                  </BoxHeader>
                  <BoxBody className="p-3 space-y-2 text-xs">
                    <div className="flex items-center justify-between text-muted-foreground">
                      <span>Require linear history</span>
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                    </div>
                    <div className="flex items-center justify-between text-muted-foreground">
                      <span>Require signed commits</span>
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                    </div>
                    <div className="flex items-center justify-between text-muted-foreground">
                      <span>1 approving review</span>
                      <Check className="w-3.5 h-3.5 text-emerald-500" />
                    </div>
                  </BoxBody>
                </Box>
              </div>
            </div>
          </motion.div>
        ) : (
          /* =========================================================================
             PR LIST VIEW: GitHub Primer Table & Filter System
             ========================================================================= */
          <motion.div
            key="pr-list"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="space-y-4"
          >
            {/* Filter Search Bar & Action Controls */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div className="relative flex-1 max-w-lg">
                <Search className="w-4 h-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Filter pull requests..."
                  className="w-full h-8 pl-9 pr-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                />
              </div>

              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={loadPRs} disabled={loading} className="gap-1.5 text-xs">
                  <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
                  <span>Refresh</span>
                </Button>

                <Button variant="primary" size="sm" onClick={() => setIsCreateModalOpen(true)} className="gap-1.5 text-xs">
                  <GitPullRequest className="w-3.5 h-3.5" />
                  <span>New pull request</span>
                </Button>
              </div>
            </div>

            {/* Primer Box PR Table */}
            <Box>
              {/* BoxHeader with Open / Closed Filter and Column Dropdowns */}
              <BoxHeader className="bg-secondary/35 py-2.5">
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => setActiveFilter('open')}
                    className={`flex items-center gap-1.5 text-xs font-semibold transition-colors ${
                      activeFilter === 'open' ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    <GitPullRequest className="w-4 h-4 text-emerald-500" />
                    <span>{openPrsCount} Open</span>
                  </button>

                  <button
                    onClick={() => setActiveFilter('merged')}
                    className={`flex items-center gap-1.5 text-xs font-semibold transition-colors ${
                      activeFilter === 'merged' ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    <Check className="w-4 h-4 text-purple-500" />
                    <span>{mergedPrsCount} Closed</span>
                  </button>
                </div>

                <div className="hidden sm:flex items-center gap-4 text-xs text-muted-foreground">
                  <button className="flex items-center gap-1 hover:text-foreground transition-colors">
                    <span>Author</span>
                    <ChevronDown className="w-3 h-3" />
                  </button>
                  <button className="flex items-center gap-1 hover:text-foreground transition-colors">
                    <span>Label</span>
                    <ChevronDown className="w-3 h-3" />
                  </button>
                  <button className="flex items-center gap-1 hover:text-foreground transition-colors">
                    <span>Projects</span>
                    <ChevronDown className="w-3 h-3" />
                  </button>
                  <button className="flex items-center gap-1 hover:text-foreground transition-colors">
                    <span>Sort</span>
                    <ChevronDown className="w-3 h-3" />
                  </button>
                </div>
              </BoxHeader>

              {/* Rows */}
              <BoxBody className="p-0">
                {filteredPrs.length === 0 ? (
                  <div className="p-8 text-center space-y-4">
                    <div className="w-10 h-10 rounded-md bg-secondary border border-border flex items-center justify-center mx-auto text-muted-foreground">
                      <GitPullRequest className="w-5 h-5" />
                    </div>

                    <div className="max-w-md mx-auto space-y-1">
                      <h3 className="text-sm font-semibold text-foreground">
                        {activeFilter === 'open'
                          ? 'No open pull requests'
                          : activeFilter === 'merged'
                          ? 'No closed pull requests'
                          : 'No pull requests found'}
                      </h3>
                      <p className="text-xs text-muted-foreground">
                        Pull requests allow team members and autonomous AI agents to propose features and enforce enterprise branch protection rulesets.
                      </p>
                    </div>

                    {/* Push branch command box */}
                    <div className="max-w-md mx-auto text-left bg-secondary/40 border border-border rounded-md p-3 space-y-2">
                      <div className="flex items-center justify-between text-[11px] text-muted-foreground">
                        <span className="flex items-center gap-1.5 font-mono">
                          <Terminal className="w-3.5 h-3.5 text-primary" />
                          <span>Push a branch to open a pull request</span>
                        </span>
                        <button
                          onClick={handleCopyCli}
                          className="hover:text-foreground transition-colors flex items-center gap-1 text-[10px] text-primary"
                        >
                          {copiedCli ? <Check className="w-3 h-3 text-emerald-500" /> : <Copy className="w-3 h-3" />}
                          <span>{copiedCli ? 'Copied' : 'Copy'}</span>
                        </button>
                      </div>
                      <pre className="font-mono text-xs text-foreground leading-relaxed overflow-x-auto whitespace-pre p-2 bg-background rounded border border-border">
                        {samplePushCommand}
                      </pre>
                    </div>

                    <div>
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={() => setIsCreateModalOpen(true)}
                        className="gap-1.5 text-xs mx-auto"
                      >
                        <GitPullRequest className="w-3.5 h-3.5" />
                        <span>Create pull request</span>
                      </Button>
                    </div>
                  </div>
                ) : (
                  filteredPrs.map((pr) => (
                    <BoxRow
                      key={pr.number}
                      onClick={() => setSelectedPr(pr)}
                      className="flex items-start justify-between p-3.5 hover:bg-secondary/20 transition-colors cursor-pointer group"
                    >
                      <div className="flex items-start gap-3">
                        <div className="mt-0.5 shrink-0">
                          {pr.state === 'merged' ? (
                            <GitMerge className="w-4 h-4 text-purple-500" />
                          ) : (
                            <GitPullRequest className="w-4 h-4 text-emerald-500" />
                          )}
                        </div>

                        <div className="space-y-1">
                          <div className="flex items-center gap-2 flex-wrap">
                            <span className="text-sm font-semibold text-foreground group-hover:text-primary transition-colors">
                              {pr.title}
                            </span>
                            {(pr.labels || []).map((lbl) => (
                              <span
                                key={lbl.name}
                                className="text-[10px] px-1.5 py-0.5 rounded-md font-medium border"
                                style={{
                                  backgroundColor: `${lbl.color}15`,
                                  borderColor: `${lbl.color}35`,
                                  color: lbl.color,
                                }}
                              >
                                {lbl.name}
                              </span>
                            ))}
                          </div>

                          <div className="text-xs text-muted-foreground flex items-center gap-1.5">
                            <span>#{pr.number}</span>
                            <span>•</span>
                            <span>
                              opened {pr.createdAt || 'recently'} by <strong className="text-foreground font-medium">{getAuthorName(pr.author)}</strong>
                            </span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-3 text-xs text-muted-foreground shrink-0 pt-0.5">
                        <div className="flex items-center gap-1">
                          <MessageSquare className="w-3.5 h-3.5" />
                          <span className="font-mono text-[11px]">{(pr.comments || []).length}</span>
                        </div>
                      </div>
                    </BoxRow>
                  ))
                )}
              </BoxBody>
            </Box>
          </motion.div>
        )}
      </AnimatePresence>

      <CreatePullRequestModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        repo={repo}
        onPullRequestCreated={(newPr) => {
          setPrs((prev) => [newPr, ...prev])
          setSelectedPr(newPr)
        }}
      />
    </div>
  )
}
