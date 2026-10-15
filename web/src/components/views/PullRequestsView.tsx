import React, { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { GitPullRequest, GitMerge, MessageSquare, CheckCircle2, ArrowRight, ShieldCheck, FileDiff, Sparkles, Send } from 'lucide-react'
import confetti from 'canvas-confetti'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { SlideToConfirm } from '@/components/ui/SlideToConfirm'
import { type PullRequest, samplePullRequests } from '@/lib/api'

export interface PullRequestsViewProps {
  onSelectAgentConsole: () => void
}

export const PullRequestsView: React.FC<PullRequestsViewProps> = ({
  onSelectAgentConsole: _onSelectAgentConsole,
}) => {
  const [prs, setPrs] = useState<PullRequest[]>(samplePullRequests)
  const [activeFilter, setActiveFilter] = useState<'all' | 'open' | 'merged' | 'closed'>('open')
  const [selectedPr, setSelectedPr] = useState<PullRequest | null>(samplePullRequests[0])
  const [activeSubTab, setActiveSubTab] = useState<'conversation' | 'diff' | 'checks'>('conversation')
  const [newCommentText, setNewCommentText] = useState('')
  const [mergeMethod, setMergeMethod] = useState<'squash' | 'merge'>('squash')

  const filteredPrs = prs.filter((pr) => {
    if (activeFilter === 'all') return true
    return pr.state === activeFilter
  })

  const handleMergePr = (prNumber: number) => {
    // Fire festive celebration
    confetti({
      particleCount: 80,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#38bdf8', '#818cf8', '#c084fc', '#34d399'],
    })

    // Update state to merged
    setPrs((prev) =>
      prev.map((p) => {
        if (p.number === prNumber) {
          return {
            ...p,
            state: 'merged',
            comments: [
              ...p.comments,
              {
                id: `c-merge-${Date.now()}`,
                author: 'Roshan Ravani',
                isAgent: false,
                createdAt: 'Just now',
                body: `Merged pull request #${prNumber} via ${mergeMethod.toUpperCase()} merge. Ruleset gates satisfied.`,
              },
            ],
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
              comments: [
                ...prev.comments,
                {
                  id: `c-merge-${Date.now()}`,
                  author: 'Roshan Ravani',
                  isAgent: false,
                  createdAt: 'Just now',
                  body: `Merged pull request #${prNumber} via ${mergeMethod.toUpperCase()} merge. Ruleset gates satisfied.`,
                },
              ],
            }
          : null
      )
    }
  }

  const handleAddComment = () => {
    if (!newCommentText.trim() || !selectedPr) return
    const newComment = {
      id: `c-${Date.now()}`,
      author: 'Roshan Ravani',
      isAgent: false,
      createdAt: 'Just now',
      body: newCommentText,
    }

    const updated = {
      ...selectedPr,
      comments: [...selectedPr.comments, newComment],
    }

    setSelectedPr(updated)
    setPrs((prev) => prev.map((p) => (p.number === selectedPr.number ? updated : p)))
    setNewCommentText('')
  }

  return (
    <div className="space-y-6">
      {/* Sub-navigation Filters */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-1.5 p-1 bg-slate-900/60 border border-slate-800 rounded-xl">
          <button
            onClick={() => { setActiveFilter('open'); setSelectedPr(null) }}
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
            onClick={() => { setActiveFilter('merged'); setSelectedPr(null) }}
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
            onClick={() => { setActiveFilter('all'); setSelectedPr(null) }}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors ${
              activeFilter === 'all' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-400 hover:text-white'
            }`}
          >
            All
          </button>
        </div>

        <Button variant="glow" size="sm" className="gap-1.5">
          <GitPullRequest className="w-3.5 h-3.5" />
          <span>New Pull Request</span>
        </Button>
      </div>

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
                      opened by <strong className="text-slate-200">{selectedPr.author}</strong> {selectedPr.createdAt}
                    </span>
                  </div>
                </div>

                <Button
                  variant="outline"
                  size="xs"
                  onClick={() => setSelectedPr(null)}
                  className="self-start md:self-center"
                >
                  Back to PR list
                </Button>
              </div>

              {/* Subtabs: Conversation / Diffs / Checks */}
              <div className="flex items-center gap-2 border-t border-slate-800/80 pt-4">
                <button
                  onClick={() => setActiveSubTab('conversation')}
                  className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
                    activeSubTab === 'conversation'
                      ? 'bg-sky-500/20 text-sky-300 font-semibold'
                      : 'text-slate-400 hover:text-white'
                  }`}
                >
                  <MessageSquare className="w-3.5 h-3.5" />
                  <span>Conversation</span>
                  <span className="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded-full">
                    {selectedPr.comments.length}
                  </span>
                </button>

                <button
                  onClick={() => setActiveSubTab('diff')}
                  className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
                    activeSubTab === 'diff'
                      ? 'bg-sky-500/20 text-sky-300 font-semibold'
                      : 'text-slate-400 hover:text-white'
                  }`}
                >
                  <FileDiff className="w-3.5 h-3.5" />
                  <span>Files Changed</span>
                  <span className="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded-full text-emerald-400">
                    +{selectedPr.diffStats.additions}
                  </span>
                  <span className="text-[10px] bg-slate-800 px-1.5 py-0.2 rounded-full text-rose-400">
                    -{selectedPr.diffStats.deletions}
                  </span>
                </button>

                <button
                  onClick={() => setActiveSubTab('checks')}
                  className={`px-3 py-1.5 text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5 ${
                    activeSubTab === 'checks'
                      ? 'bg-sky-500/20 text-sky-300 font-semibold'
                      : 'text-slate-400 hover:text-white'
                  }`}
                >
                  <ShieldCheck className="w-3.5 h-3.5 text-teal-400" />
                  <span>Checks</span>
                  <span className="text-[10px] bg-teal-950 text-teal-300 border border-teal-800 px-1.5 py-0.2 rounded-full">
                    {selectedPr.checks.length}/{selectedPr.checks.length} Pass
                  </span>
                </button>
              </div>
            </Card>

            {/* Subtab Contents */}
            {activeSubTab === 'conversation' && (
              <div className="space-y-6">
                {/* Description Body */}
                <Card className="border-slate-800 bg-slate-950/40 p-5 space-y-3">
                  <div className="flex items-center gap-3">
                    <LivingAvatar name={selectedPr.author} isAgent={selectedPr.isAgent} />
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-semibold text-slate-100">{selectedPr.author}</span>
                        {selectedPr.isAgent && (
                          <Badge variant="agent" className="py-0 text-[10px]">
                            <Sparkles className="w-3 h-3 mr-1" /> Autonomous Agent
                          </Badge>
                        )}
                        <span className="text-[11px] text-slate-500">{selectedPr.createdAt}</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-xs sm:text-sm text-slate-300 leading-relaxed pl-12">
                    {selectedPr.body}
                  </div>
                </Card>

                {/* Comment Thread Timeline */}
                <div className="space-y-4">
                  {selectedPr.comments.map((comment) => (
                    <Card key={comment.id} className="border-slate-800/80 bg-slate-950/50 p-5 space-y-3">
                      <div className="flex items-center gap-3">
                        <LivingAvatar name={comment.author} isAgent={comment.isAgent} size="sm" />
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-semibold text-slate-200">{comment.author}</span>
                          {comment.isAgent && (
                            <Badge variant="agent" className="text-[9px] py-0">
                              MCP Agent
                            </Badge>
                          )}
                          <span className="text-[11px] text-slate-500">{comment.createdAt}</span>
                        </div>
                      </div>
                      <div className="text-xs sm:text-sm text-slate-300 leading-relaxed pl-10">
                        {comment.body}
                      </div>
                    </Card>
                  ))}
                </div>

                {/* Ruleset & Merge Gate Box */}
                {selectedPr.state === 'open' && (
                  <Card className="border-emerald-500/30 bg-emerald-950/10 p-5 space-y-4">
                    <div className="flex items-start justify-between gap-4 flex-wrap">
                      <div className="space-y-1">
                        <h4 className="text-sm font-semibold text-emerald-300 flex items-center gap-2">
                          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                          <span>All Ruleset Gates Passed</span>
                        </h4>
                        <p className="text-xs text-slate-400">
                          3 passing checks, 1 approved review, zero merge conflicts against main.
                        </p>
                      </div>

                      {/* Merge Mode Toggle */}
                      <div className="flex items-center gap-2 bg-slate-900 border border-slate-800 rounded-xl p-1 text-xs">
                        <button
                          onClick={() => setMergeMethod('squash')}
                          className={`px-2.5 py-1 rounded-lg transition-colors ${
                            mergeMethod === 'squash' ? 'bg-sky-500/20 text-sky-300 font-semibold' : 'text-slate-400'
                          }`}
                        >
                          Squash and merge
                        </button>
                        <button
                          onClick={() => setMergeMethod('merge')}
                          className={`px-2.5 py-1 rounded-lg transition-colors ${
                            mergeMethod === 'merge' ? 'bg-purple-500/20 text-purple-300 font-semibold' : 'text-slate-400'
                          }`}
                        >
                          Create a merge commit
                        </button>
                      </div>
                    </div>

                    {/* Space UI Slide to Confirm living component */}
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

            {activeSubTab === 'diff' && (
              <div className="space-y-4">
                {selectedPr.diffHunks.map((diff, idx) => (
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
                ))}
              </div>
            )}

            {activeSubTab === 'checks' && (
              <Card className="border-slate-800 bg-slate-950/60 p-5 space-y-4">
                <h4 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
                  <ShieldCheck className="w-4 h-4 text-sky-400" />
                  <span>CI & Security Check Runs</span>
                </h4>
                <div className="divide-y divide-slate-800/80">
                  {selectedPr.checks.map((chk) => (
                    <div key={chk.id} className="flex items-center justify-between py-3">
                      <div className="flex items-center gap-2.5">
                        <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                        <span className="text-xs font-medium text-slate-200">{chk.name}</span>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-slate-400">
                        <span>duration: {chk.duration}</span>
                        <Badge variant="success" className="text-[10px] py-0">Passed</Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </motion.div>
        ) : (
          /* PR List Overview */
          <motion.div
            key="pr-list"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="space-y-4"
          >
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
                          {pr.labels.map((lbl) => (
                            <span
                              key={lbl.name}
                              className="text-[10px] px-2 py-0.5 rounded-full font-medium"
                              style={{ backgroundColor: `${lbl.color}20`, color: lbl.color, border: `1px solid ${lbl.color}40` }}
                            >
                              {lbl.name}
                            </span>
                          ))}
                        </div>

                        <div className="text-xs text-slate-500 flex items-center gap-2">
                          <span>#{pr.number}</span>
                          <span>•</span>
                          <span>opened by <strong className="text-slate-400">{pr.author}</strong></span>
                          <span>{pr.createdAt}</span>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-3 text-xs text-slate-500 shrink-0">
                      <div className="flex items-center gap-1">
                        <MessageSquare className="w-3.5 h-3.5" />
                        <span>{pr.comments.length}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
