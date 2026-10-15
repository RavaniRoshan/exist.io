import React, { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  GitPullRequest,
  GitMerge,
  MessageSquare,
  CheckCircle2,
  AlertCircle,
  ArrowRight,
  ShieldCheck,
  FileDiff,
  Sparkles,
  Send,
  Copy,
  Check,
  Terminal,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import confetti from 'canvas-confetti'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { SlideToConfirm } from '@/components/ui/SlideToConfirm'
import {
  type PullRequest,
  type RepoInfo,
  type AgentReviewResult,
  getPullRequests,
  mergePullRequest,
  triggerAgentReview,
  isDemoModeActive,
  samplePullRequests,
} from '@/lib/api'

export interface PullRequestsViewProps {
  repo?: RepoInfo
  onSelectAgentConsole?: () => void
}

export const PullRequestsView: React.FC<PullRequestsViewProps> = ({
  repo,
  onSelectAgentConsole: _onSelectAgentConsole,
}) => {
  const owner = repo?.ownerSlug || 'acme'
  const repoName = repo?.name || 'rocket'

  const [prs, setPrs] = useState<PullRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [activeFilter, setActiveFilter] = useState<'all' | 'open' | 'merged' | 'closed'>('open')
  const [selectedPr, setSelectedPr] = useState<PullRequest | null>(null)
  const [activeSubTab, setActiveSubTab] = useState<'conversation' | 'diff' | 'checks'>('conversation')
  const [newCommentText, setNewCommentText] = useState('')
  const [mergeMethod, setMergeMethod] = useState<'squash' | 'merge'>('squash')
  const [isReviewing, setIsReviewing] = useState(false)
  const [reviewResult, setReviewResult] = useState<AgentReviewResult | null>(null)
  const [copiedCli, setCopiedCli] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const loadPRs = async () => {
    setLoading(true)
    setErrorMessage(null)
    try {
      const data = await getPullRequests(owner, repoName)
      setPrs(data)
      if (data.length > 0 && !selectedPr) {
        setSelectedPr(data[0])
      }
    } catch (err: any) {
      if (isDemoModeActive()) {
        setPrs(samplePullRequests)
        setSelectedPr(samplePullRequests[0])
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

  const filteredPrs = prs.filter((pr) => {
    if (activeFilter === 'all') return true
    return pr.state === activeFilter
  })

  const handleMergePr = async (prNumber: number) => {
    setErrorMessage(null)
    try {
      await mergePullRequest(owner, repoName, prNumber, mergeMethod)

      // Fire festive celebration
      confetti({
        particleCount: 80,
        spread: 70,
        origin: { y: 0.6 },
        colors: ['#38bdf8', '#818cf8', '#c084fc', '#34d399'],
      })

      // Update state to merged
      const mergeComment = {
        id: `c-merge-${Date.now()}`,
        author: 'Roshan Ravani',
        isAgent: false,
        createdAt: 'Just now',
        body: `Merged pull request #${prNumber} via ${mergeMethod.toUpperCase()} merge. Ruleset gates and security checks satisfied.`,
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

      if (selectedPr && selectedPr.number === prNumber) {
        setSelectedPr((prev) =>
          prev
            ? {
                ...prev,
                state: 'merged',
                comments: [...(prev.comments || []), mergeComment],
              }
            : null
        )
      }
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

      // Append agent review finding to conversation
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

  const handleAddComment = () => {
    if (!newCommentText.trim() || !selectedPr) return
    const newComment = {
      id: `c-${Date.now()}`,
      author: 'Current Operator',
      isAgent: false,
      createdAt: 'Just now',
      body: newCommentText,
    }

    const updated = {
      ...selectedPr,
      comments: [...(selectedPr.comments || []), newComment],
    }

    setSelectedPr(updated)
    setPrs((prev) => prev.map((p) => (p.number === selectedPr.number ? updated : p)))
    setNewCommentText('')
  }

  const samplePushCommand = `git checkout -b feature/agent-capabilities
# ... apply code enhancements ...
git commit -m "feat: implement autonomous capability leases"
git push origin feature/agent-capabilities`

  const handleCopyCli = () => {
    navigator.clipboard.writeText(samplePushCommand)
    setCopiedCli(true)
    setTimeout(() => setCopiedCli(false), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Sub-navigation Filters & Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-1.5 p-1 bg-slate-900/60 border border-slate-800 rounded-xl">
          <button
            onClick={() => {
              setActiveFilter('open')
              setSelectedPr(null)
            }}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
              activeFilter === 'open' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            <GitPullRequest className="w-3.5 h-3.5 text-emerald-400" />
            <span>Open</span>
            <span className="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded-full">
              {prs.filter((p) => p.state === 'open').length}
            </span>
          </button>
          <button
            onClick={() => {
              setActiveFilter('merged')
              setSelectedPr(null)
            }}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
              activeFilter === 'merged' ? 'bg-purple-500/20 text-purple-300 font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            <GitMerge className="w-3.5 h-3.5 text-purple-400" />
            <span>Merged</span>
            <span className="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded-full">
              {prs.filter((p) => p.state === 'merged').length}
            </span>
          </button>
          <button
            onClick={() => {
              setActiveFilter('all')
              setSelectedPr(null)
            }}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors ${
              activeFilter === 'all' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            All
          </button>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={loadPRs} disabled={loading} className="gap-1.5 text-slate-400">
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>

          <Button variant="glow" size="sm" className="gap-1.5">
            <GitPullRequest className="w-3.5 h-3.5" />
            <span>New Pull Request</span>
          </Button>
        </div>
      </div>

      {errorMessage && (
        <div className="p-3 bg-rose-950/40 border border-rose-800/80 rounded-xl text-xs text-rose-300 flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
          <span>{errorMessage}</span>
        </div>
      )}

      <AnimatePresence mode="wait">
        {selectedPr ? (
          /* PR Detail View */
          <motion.div
            key="pr-detail"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="space-y-6"
          >
            {/* PR Header Card */}
            <Card className="border-slate-800 bg-slate-950/60 p-6 space-y-4">
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="space-y-2">
                  <div className="flex items-center gap-2.5 flex-wrap">
                    <button
                      onClick={() => setSelectedPr(null)}
                      className="text-xs text-sky-400 hover:underline mr-1 flex items-center gap-1"
                    >
                      ← Back to list
                    </button>
                    <span className="text-xl sm:text-2xl font-bold text-white tracking-tight">
                      {selectedPr.title}
                    </span>
                    <span className="text-xl font-light text-slate-500">#{selectedPr.number}</span>
                  </div>

                  {/* Branch and author meta */}
                  <div className="flex items-center gap-2 text-xs text-slate-400 flex-wrap">
                    <Badge variant={selectedPr.state === 'merged' ? 'merged' : 'open'} pulse={selectedPr.state === 'open'}>
                      {selectedPr.state === 'merged' ? (
                        <>
                          <GitMerge className="w-3 h-3 mr-1" /> Merged
                        </>
                      ) : (
                        <>
                          <GitPullRequest className="w-3 h-3 mr-1" /> Open
                        </>
                      )}
                    </Badge>

                    <div className="flex items-center gap-1.5 font-mono text-[11px]">
                      <span className="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-sky-300">
                        {selectedPr.headRef}
                      </span>
                      <ArrowRight className="w-3 h-3 text-slate-600" />
                      <span className="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-slate-300">
                        {selectedPr.baseRef}
                      </span>
                    </div>

                    <span>•</span>
                    <span className="flex items-center gap-1">
                      opened by <strong className="text-foreground">{selectedPr.author || 'dev'}</strong> {selectedPr.createdAt || 'recently'}
                    </span>
                  </div>
                </div>

                {/* Agent Review Action Button */}
                {selectedPr.state === 'open' && (
                  <div className="flex items-center gap-2 shrink-0">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleRunAgentReview}
                      disabled={isReviewing}
                      className="border-primary/40 text-primary hover:bg-primary/10 gap-1.5"
                    >
                      {isReviewing ? (
                        <>
                          <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
                          <span>Reviewing Diff & Security...</span>
                        </>
                      ) : (
                        <>
                          <Sparkles className="w-3.5 h-3.5 text-primary" />
                          <span>Run AI Agent Review</span>
                        </>
                      )}
                    </Button>
                  </div>
                )}
              </div>

              {/* Review result summary banner if just executed */}
              {reviewResult && (
                <div
                  className={`p-3 rounded-xl border text-xs flex items-center justify-between gap-3 ${
                    reviewResult.conclusion === 'success'
                      ? 'bg-emerald-950/30 border-emerald-800/80 text-emerald-300'
                      : reviewResult.conclusion === 'failure'
                      ? 'bg-destructive/15 border-destructive/30 text-destructive'
                      : 'bg-amber-950/30 border-amber-800/80 text-amber-300'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Sparkles className="w-4 h-4 shrink-0" />
                    <span>{reviewResult.summary}</span>
                  </div>
                  <Badge variant={reviewResult.conclusion === 'success' ? 'success' : 'warning'}>
                    Score: {reviewResult.score}/100
                  </Badge>
                </div>
              )}

              {/* Sub-tabs inside PR */}
              <div className="flex items-center gap-4 pt-3 border-t border-border text-xs font-medium">
                <button
                  onClick={() => setActiveSubTab('conversation')}
                  className={`flex items-center gap-1.5 pb-2 border-b-2 transition-colors ${
                    activeSubTab === 'conversation'
                      ? 'border-primary text-primary font-semibold'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <MessageSquare className="w-3.5 h-3.5" />
                  <span>Conversation</span>
                  <span className="text-[10px] bg-secondary text-muted-foreground px-1.5 py-0.2 rounded-full">
                    {(selectedPr.comments || []).length}
                  </span>
                </button>

                <button
                  onClick={() => setActiveSubTab('diff')}
                  className={`flex items-center gap-1.5 pb-2 border-b-2 transition-colors ${
                    activeSubTab === 'diff'
                      ? 'border-primary text-primary font-semibold'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <FileDiff className="w-3.5 h-3.5" />
                  <span>Files Changed</span>
                  <span className="text-[10px] bg-secondary text-muted-foreground px-1.5 py-0.2 rounded-full">
                    {(selectedPr.diffHunks || []).length}
                  </span>
                </button>

                <button
                  onClick={() => setActiveSubTab('checks')}
                  className={`flex items-center gap-1.5 pb-2 border-b-2 transition-colors ${
                    activeSubTab === 'checks'
                      ? 'border-primary text-primary font-semibold'
                      : 'border-transparent text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <ShieldCheck className="w-3.5 h-3.5" />
                  <span>Checks & Policy</span>
                  <span className="text-[10px] bg-secondary text-muted-foreground px-1.5 py-0.2 rounded-full">
                    {(selectedPr.checks || []).length}
                  </span>
                </button>
              </div>
            </Card>

            {/* Sub-tab 1: Conversation & Discussion */}
            {activeSubTab === 'conversation' && (
              <div className="space-y-4">
                {(selectedPr.comments || []).map((comment) => (
                  <Card key={comment.id} className="border-slate-800 bg-slate-950/70 p-5 space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2.5">
                        <LivingAvatar
                          name={comment.author}
                          isAgent={comment.isAgent}
                          status={comment.isAgent ? 'online' : 'offline'}
                          size="sm"
                        />
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-slate-200">{comment.author}</span>
                          {comment.isAgent && (
                            <Badge variant="agent" className="text-[9px] py-0 px-1.5">
                              Agent Reviewer
                            </Badge>
                          )}
                          <span className="text-[11px] text-slate-500">{comment.createdAt}</span>
                        </div>
                      </div>
                    </div>

                    <div className="text-xs sm:text-sm text-slate-300 whitespace-pre-wrap leading-relaxed font-sans pl-8 border-l border-slate-800/60 ml-3">
                      {comment.body}
                    </div>
                  </Card>
                ))}

                {/* Merge Action Card */}
                {selectedPr.state === 'open' && (
                  <Card className="border-sky-500/20 bg-slate-950/80 p-5 space-y-4">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                      <div className="space-y-1">
                        <h4 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
                          <GitMerge className="w-4 h-4 text-sky-400" />
                          <span>Merge Pull Request</span>
                        </h4>
                        <p className="text-xs text-slate-400">
                          Merges branches safely into <code className="text-sky-300 font-mono">{selectedPr.baseRef}</code>.
                        </p>
                      </div>

                      <div className="flex items-center gap-1.5 bg-slate-900 border border-slate-800 p-1 rounded-xl text-xs">
                        <button
                          onClick={() => setMergeMethod('squash')}
                          className={`px-2.5 py-1 rounded-lg transition-colors ${
                            mergeMethod === 'squash' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400'
                          }`}
                        >
                          Squash
                        </button>
                        <button
                          onClick={() => setMergeMethod('merge')}
                          className={`px-2.5 py-1 rounded-lg transition-colors ${
                            mergeMethod === 'merge' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400'
                          }`}
                        >
                          Create Commit
                        </button>
                      </div>
                    </div>

                    <div className="pt-2">
                      <SlideToConfirm
                        onConfirm={() => handleMergePr(selectedPr.number)}
                        label={`Slide to ${mergeMethod === 'squash' ? 'Squash & Merge' : 'Merge'} Pull Request`}
                      />
                    </div>
                  </Card>
                )}

                {/* New Comment Box */}
                <Card className="border-slate-800 bg-slate-950/60 p-4 space-y-3">
                  <textarea
                    value={newCommentText}
                    onChange={(e) => setNewCommentText(e.target.value)}
                    placeholder="Leave a comment or review suggestion..."
                    className="w-full bg-slate-900/80 text-xs sm:text-sm text-slate-200 placeholder:text-slate-500 rounded-xl border border-slate-800 p-3 outline-none focus:border-sky-500/70 focus:ring-1 focus:ring-sky-500/50 resize-y min-h-[90px]"
                  />
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="glow"
                      size="sm"
                      onClick={handleAddComment}
                      disabled={!newCommentText.trim()}
                      className="gap-1.5"
                    >
                      <Send className="w-3.5 h-3.5" />
                      <span>Comment</span>
                    </Button>
                  </div>
                </Card>
              </div>
            )}

            {/* Sub-tab 2: Diff View */}
            {activeSubTab === 'diff' && (
              <div className="space-y-4">
                {(selectedPr.diffHunks || []).length === 0 ? (
                  <Card className="border-slate-800 bg-slate-950/60 p-8 text-center text-slate-400 text-xs">
                    No individual diff hunks recorded for this pull request.
                  </Card>
                ) : (
                  (selectedPr.diffHunks || []).map((diff, idx) => (
                    <Card key={idx} className="p-0 overflow-hidden border-slate-800 bg-slate-950/70">
                      <div className="flex items-center justify-between px-4 py-2.5 bg-slate-900/80 border-b border-slate-800 text-xs font-mono">
                        <span className="text-slate-200 font-medium">{diff.newPath}</span>
                        <div className="flex items-center gap-2">
                          <span className="text-emerald-400 font-semibold">+{diff.additions}</span>
                          <span className="text-rose-400 font-semibold">-{diff.deletions}</span>
                        </div>
                      </div>

                      <div className="p-3 font-mono text-xs overflow-x-auto leading-relaxed bg-slate-950">
                        <pre>
                          {diff.hunk.split('\n').map((line, lIdx) => {
                            const isAdd = line.startsWith('+')
                            const isDel = line.startsWith('-')
                            const isHeader = line.startsWith('@@')
                            return (
                              <div
                                key={lIdx}
                                className={`px-2 py-0.5 rounded ${
                                  isAdd
                                    ? 'bg-emerald-950/40 text-emerald-300'
                                    : isDel
                                    ? 'bg-rose-950/40 text-rose-300'
                                    : isHeader
                                    ? 'text-sky-400/80 bg-sky-950/20'
                                    : 'text-slate-300'
                                }`}
                              >
                                {line}
                              </div>
                            )
                          })}
                        </pre>
                      </div>
                    </Card>
                  ))
                )}
              </div>
            )}

            {/* Sub-tab 3: Checks View */}
            {activeSubTab === 'checks' && (
              <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
                <h4 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
                  <ShieldCheck className="w-4 h-4 text-sky-400" />
                  <span>CI & Security Check Runs</span>
                </h4>
                <div className="divide-y divide-slate-800/80">
                  {(selectedPr.checks || []).map((chk) => (
                    <div key={chk.id} className="flex items-center justify-between py-3">
                      <div className="flex items-center gap-2.5">
                        <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                        <span className="text-xs font-medium text-slate-200">{chk.name}</span>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-slate-400">
                        <span>duration: {chk.duration}</span>
                        <Badge variant="success" className="text-[10px] py-0">
                          Passed
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </motion.div>
        ) : (
          /* PR List Overview */
          <motion.div key="pr-list" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-4">
            {filteredPrs.length === 0 ? (
              /* Authentic High-Craft Empty State */
              <Card className="border-slate-800 bg-slate-950/60 p-8 text-center space-y-5">
                <div className="w-12 h-12 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center mx-auto text-sky-400">
                  <GitPullRequest className="w-6 h-6" />
                </div>

                <div className="max-w-md mx-auto space-y-1.5">
                  <h3 className="text-base font-semibold text-slate-100">
                    {activeFilter === 'open'
                      ? 'No open pull requests'
                      : activeFilter === 'merged'
                      ? 'No merged pull requests yet'
                      : 'No pull requests found'}
                  </h3>
                  <p className="text-xs text-slate-400 leading-relaxed">
                    Pull requests allow team members and autonomous AI agents to propose features, review cryptographic diffs,
                    and enforce enterprise branch protection rulesets before merging.
                  </p>
                </div>

                {/* Copyable CLI Instructions */}
                <div className="max-w-lg mx-auto text-left bg-slate-900/90 border border-slate-800 rounded-xl p-3 space-y-2">
                  <div className="flex items-center justify-between text-[11px] text-slate-400">
                    <span className="flex items-center gap-1.5 font-mono">
                      <Terminal className="w-3.5 h-3.5 text-sky-400" />
                      <span>Push a branch to open a pull request</span>
                    </span>
                    <button
                      onClick={handleCopyCli}
                      className="hover:text-white transition-colors flex items-center gap-1 text-[10px] text-sky-400"
                    >
                      {copiedCli ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                      <span>{copiedCli ? 'Copied' : 'Copy'}</span>
                    </button>
                  </div>
                  <pre className="font-mono text-xs text-slate-300 leading-relaxed overflow-x-auto whitespace-pre p-2 bg-slate-950 rounded-lg">
                    {samplePushCommand}
                  </pre>
                </div>
              </Card>
            ) : (
              <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/50">
                <div className="divide-y divide-slate-850/70">
                  {filteredPrs.map((pr) => (
                    <div
                      key={pr.number}
                      onClick={() => setSelectedPr(pr)}
                      className="flex items-start justify-between p-4 hover:bg-slate-900/50 transition-colors cursor-pointer group"
                    >
                      <div className="flex items-start gap-3">
                        <div className="mt-1">
                          {pr.state === 'merged' ? (
                            <GitMerge className="w-4 h-4 text-purple-400" />
                          ) : (
                            <GitPullRequest className="w-4 h-4 text-emerald-400" />
                          )}
                        </div>
                        <div className="space-y-1">
                          <div className="flex items-center gap-2 flex-wrap">
                            <span className="text-sm font-semibold text-slate-100 group-hover:text-sky-400 transition-colors">
                              {pr.title}
                            </span>
                            {(pr.labels || []).map((lbl) => (
                              <span
                                key={lbl.name}
                                className="text-[10px] px-2 py-0.5 rounded-full font-medium"
                                style={{
                                  backgroundColor: `${lbl.color}20`,
                                  color: lbl.color,
                                  border: `1px solid ${lbl.color}40`,
                                }}
                              >
                                {lbl.name}
                              </span>
                            ))}
                          </div>

                          <div className="text-xs text-slate-500 flex items-center gap-2">
                            <span>#{pr.number}</span>
                            <span>•</span>
                            <span>
                              opened by <strong className="text-slate-400">{pr.author || 'dev'}</strong>
                            </span>
                            <span>{pr.createdAt || 'recently'}</span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-3 text-xs text-slate-500 shrink-0">
                        <div className="flex items-center gap-1">
                          <MessageSquare className="w-3.5 h-3.5" />
                          <span>{(pr.comments || []).length}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
