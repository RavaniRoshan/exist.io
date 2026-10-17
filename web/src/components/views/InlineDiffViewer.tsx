import React, { useState } from 'react'
import {
  Plus,
  Check,
  CheckCircle2,
  GitCommit,
  CornerDownRight,
  ChevronDown,
  Loader2,
  X,
  Send,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Box, BoxHeader } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import {
  type ReviewComment,
  applyCodeSuggestion,
  toggleThreadResolved,
  addPullRequestComment,
  submitPullRequestReview,
} from '@/lib/api'

export interface InlineDiffViewerProps {
  owner: string
  repoName: string
  prNumber: number
  diffHunks: {
    oldPath: string
    newPath: string
    status?: 'modified' | 'added' | 'deleted'
    additions: number
    deletions: number
    hunk: string
  }[]
  comments: ReviewComment[]
  onRefreshComments: () => void | Promise<void>
  isPrOpen?: boolean
}

/**
 * Extracts suggestion block from comment body if present.
 * Format: ```suggestion\n...lines...\n```
 */
export function extractSuggestion(body: string): string | null {
  const match = body.match(/```suggestion\r?\n([\s\S]*?)\r?\n```/)
  return match ? match[1] : null
}

export const InlineDiffViewer: React.FC<InlineDiffViewerProps> = ({
  owner,
  repoName,
  prNumber,
  diffHunks,
  comments,
  onRefreshComments,
  isPrOpen = true,
}) => {
  // State for composing line comments
  const [activeComposer, setActiveComposer] = useState<{
    path: string
    line: number
    side: 'RIGHT' | 'LEFT'
    currentLineText: string
  } | null>(null)
  const [composerText, setComposerText] = useState('')
  const [isSubmittingComment, setIsSubmittingComment] = useState(false)

  // State for review submission modal/popover
  const [showReviewModal, setShowReviewModal] = useState(false)
  const [reviewVerdict, setReviewVerdict] = useState<'comment' | 'approve' | 'request_changes'>('comment')
  const [reviewBody, setReviewBody] = useState('')
  const [isSubmittingReview, setIsSubmittingReview] = useState(false)

  // State for replying to threads
  const [replyingThreadId, setReplyingThreadId] = useState<string | null>(null)
  const [replyText, setReplyText] = useState('')
  const [isSubmittingReply, setIsSubmittingReply] = useState(false)

  // Applied suggestion tracking
  const [appliedSuggestions, setAppliedSuggestions] = useState<Record<string, boolean>>({})
  const [applyingSuggestionId, setApplyingSuggestionId] = useState<string | null>(null)
  const [actionMessage, setActionMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleApplySuggestion = async (commentId: string) => {
    setApplyingSuggestionId(commentId)
    setActionMessage(null)
    try {
      await applyCodeSuggestion(owner, repoName, prNumber, commentId)
      setAppliedSuggestions((prev) => ({ ...prev, [commentId]: true }))
      setActionMessage({ type: 'success', text: 'Suggestion successfully committed to branch!' })
      await onRefreshComments()
    } catch (err: any) {
      setActionMessage({ type: 'error', text: err.message || 'Failed to apply suggestion' })
    } finally {
      setApplyingSuggestionId(null)
    }
  }

  const handleToggleResolve = async (threadId: string, currentResolved: boolean) => {
    setActionMessage(null)
    try {
      await toggleThreadResolved(owner, repoName, prNumber, threadId, !currentResolved)
      await onRefreshComments()
    } catch (err: any) {
      setActionMessage({ type: 'error', text: err.message || 'Failed to update thread resolution' })
    }
  }

  const handleAddLineComment = async () => {
    if (!activeComposer || !composerText.trim()) return
    setIsSubmittingComment(true)
    setActionMessage(null)
    try {
      await addPullRequestComment(owner, repoName, prNumber, {
        body: composerText.trim(),
        path: activeComposer.path,
        line: activeComposer.line,
        side: activeComposer.side,
      })
      setComposerText('')
      setActiveComposer(null)
      await onRefreshComments()
    } catch (err: any) {
      setActionMessage({ type: 'error', text: err.message || 'Failed to post comment' })
    } finally {
      setIsSubmittingComment(false)
    }
  }

  const handleSendReply = async (threadId: string, path?: string, line?: number) => {
    if (!replyText.trim()) return
    setIsSubmittingReply(true)
    setActionMessage(null)
    try {
      await addPullRequestComment(owner, repoName, prNumber, {
        body: replyText.trim(),
        thread_id: threadId,
        path,
        line,
      })
      setReplyText('')
      setReplyingThreadId(null)
      await onRefreshComments()
    } catch (err: any) {
      setActionMessage({ type: 'error', text: err.message || 'Failed to post reply' })
    } finally {
      setIsSubmittingReply(false)
    }
  }

  const handleSubmitReview = async () => {
    setIsSubmittingReview(true)
    setActionMessage(null)
    try {
      await submitPullRequestReview(owner, repoName, prNumber, {
        verdict: reviewVerdict,
        body: reviewBody.trim(),
      })
      setShowReviewModal(false)
      setReviewBody('')
      setActionMessage({
        type: 'success',
        text: `Review submitted with verdict: ${reviewVerdict.replace('_', ' ').toUpperCase()}`,
      })
      await onRefreshComments()
    } catch (err: any) {
      setActionMessage({ type: 'error', text: err.message || 'Failed to submit review' })
    } finally {
      setIsSubmittingReview(false)
    }
  }

  if (diffHunks.length === 0) {
    return (
      <Box className="p-8 text-center text-muted-foreground text-xs">
        No modified files found for this pull request.
      </Box>
    )
  }

  return (
    <div className="space-y-4">
      {/* Top Review Changes Bar */}
      <div className="flex items-center justify-between pb-2 border-b border-border">
        <div className="text-xs text-muted-foreground">
          Showing <span className="font-semibold text-foreground">{diffHunks.length}</span> changed files with inline reviews
        </div>

        {isPrOpen && (
          <div className="relative">
            <Button
              variant="primary"
              size="sm"
              onClick={() => setShowReviewModal(!showReviewModal)}
              className="gap-1.5 text-xs font-semibold"
            >
              <CheckCircle2 className="w-3.5 h-3.5" />
              <span>Review changes</span>
              <ChevronDown className="w-3.5 h-3.5" />
            </Button>

            {/* Review Popover Dropdown */}
            {showReviewModal && (
              <div
                data-testid="review-changes-modal"
                className="absolute right-0 top-full mt-2 z-40 w-80 sm:w-96 rounded-md border border-border bg-card shadow-2xl p-4 space-y-3"
              >
                <div className="flex items-center justify-between border-b border-border pb-2">
                  <span className="font-semibold text-xs text-foreground">Finish your review</span>
                  <button
                    onClick={() => setShowReviewModal(false)}
                    className="text-muted-foreground hover:text-foreground text-xs"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>

                <div className="space-y-1">
                  <label className="text-[11px] font-semibold text-foreground">Summary Comment (Optional)</label>
                  <textarea
                    rows={3}
                    value={reviewBody}
                    onChange={(e) => setReviewBody(e.target.value)}
                    placeholder="Leave a comment summarizing your review..."
                    className="w-full bg-input text-foreground border border-border rounded-md p-2 text-xs font-sans outline-none focus:border-primary resize-y"
                  />
                </div>

                <div className="space-y-2 border-t border-border pt-2 text-xs">
                  <label className="flex items-start gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="verdict"
                      value="comment"
                      checked={reviewVerdict === 'comment'}
                      onChange={() => setReviewVerdict('comment')}
                      className="mt-0.5"
                    />
                    <div>
                      <span className="font-medium text-foreground block">Comment</span>
                      <span className="text-[11px] text-muted-foreground">
                        Submit general feedback without explicit approval.
                      </span>
                    </div>
                  </label>

                  <label className="flex items-start gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="verdict"
                      value="approve"
                      checked={reviewVerdict === 'approve'}
                      onChange={() => setReviewVerdict('approve')}
                      className="mt-0.5 text-emerald-500"
                    />
                    <div>
                      <span className="font-medium text-emerald-500 block">Approve</span>
                      <span className="text-[11px] text-muted-foreground">
                        Submit feedback and approve merging these changes.
                      </span>
                    </div>
                  </label>

                  <label className="flex items-start gap-2 cursor-pointer">
                    <input
                      type="radio"
                      name="verdict"
                      value="request_changes"
                      checked={reviewVerdict === 'request_changes'}
                      onChange={() => setReviewVerdict('request_changes')}
                      className="mt-0.5 text-rose-500"
                    />
                    <div>
                      <span className="font-medium text-rose-500 block">Request changes</span>
                      <span className="text-[11px] text-muted-foreground">
                        Submit feedback that must be addressed before merging.
                      </span>
                    </div>
                  </label>
                </div>

                <div className="pt-2 flex items-center justify-end gap-2 border-t border-border">
                  <Button variant="ghost" size="xs" onClick={() => setShowReviewModal(false)}>
                    Cancel
                  </Button>
                  <Button
                    variant="primary"
                    size="xs"
                    disabled={isSubmittingReview}
                    onClick={handleSubmitReview}
                    className="gap-1"
                  >
                    {isSubmittingReview && <Loader2 className="w-3 h-3 animate-spin" />}
                    <span>Submit review</span>
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {actionMessage && (
        <div
          className={`p-3 rounded-md text-xs border flex items-center justify-between gap-2 ${
            actionMessage.type === 'success'
              ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-500'
              : 'bg-destructive/10 border-destructive/30 text-destructive'
          }`}
        >
          <span>{actionMessage.text}</span>
          <button onClick={() => setActionMessage(null)}>
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* Diff Files List */}
      {diffHunks.map((diff, fIdx) => {
        let currentOldLine = 0
        let currentNewLine = 0

        const lines = diff.hunk.split('\n')

        return (
          <Box key={fIdx} className="overflow-hidden">
            <BoxHeader className="bg-secondary/40 py-2 font-mono text-xs flex items-center justify-between">
              <span className="font-semibold text-foreground">{diff.newPath}</span>
              <div className="flex items-center gap-2">
                <span className="text-emerald-500 font-semibold font-mono">+{diff.additions}</span>
                <span className="text-rose-500 font-semibold font-mono">-{diff.deletions}</span>
              </div>
            </BoxHeader>

            <div className="font-mono text-xs overflow-x-auto bg-background divide-y divide-border/20">
              {lines.map((line, lIdx) => {
                const isAdd = line.startsWith('+')
                const isDel = line.startsWith('-')
                const isHeader = line.startsWith('@@')

                if (isHeader) {
                  const match = line.match(/^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/)
                  if (match) {
                    currentOldLine = parseInt(match[1], 10) - 1
                    currentNewLine = parseInt(match[2], 10) - 1
                  }
                } else if (isAdd) {
                  currentNewLine++
                } else if (isDel) {
                  currentOldLine++
                } else {
                  currentOldLine++
                  currentNewLine++
                }

                const lineNumForComment = isDel ? currentOldLine : currentNewLine
                const sideForComment: 'RIGHT' | 'LEFT' = isDel ? 'LEFT' : 'RIGHT'

                // Filter comments for this path, line, and side
                const lineComments = comments.filter(
                  (c) =>
                    c.path === diff.newPath &&
                    c.line === lineNumForComment &&
                    (!c.side || c.side === sideForComment)
                )

                const isComposerOpen =
                  activeComposer?.path === diff.newPath &&
                  activeComposer?.line === lineNumForComment &&
                  activeComposer?.side === sideForComment

                return (
                  <React.Fragment key={lIdx}>
                    <div
                      className={`group flex items-stretch leading-relaxed ${
                        isAdd
                          ? 'bg-emerald-500/10 text-emerald-400'
                          : isDel
                          ? 'bg-rose-500/10 text-rose-400'
                          : isHeader
                          ? 'bg-secondary/60 text-muted-foreground font-semibold py-1'
                          : 'text-foreground/90 hover:bg-secondary/20'
                      }`}
                    >
                      {/* Left gutter: line numbers & hover plus icon */}
                      {!isHeader ? (
                        <div className="w-16 shrink-0 flex items-center justify-between px-2 bg-secondary/20 border-r border-border text-[11px] text-muted-foreground select-none">
                          <span className="w-6 text-right font-mono">{isDel ? currentOldLine : isAdd ? '' : currentOldLine}</span>
                          <span className="w-6 text-right font-mono">{isAdd ? currentNewLine : isDel ? '' : currentNewLine}</span>
                          {isPrOpen && (
                            <button
                              type="button"
                              title="Add inline comment"
                              onClick={() => {
                                if (isComposerOpen) {
                                  setActiveComposer(null)
                                } else {
                                  setActiveComposer({
                                    path: diff.newPath,
                                    line: lineNumForComment,
                                    side: sideForComment,
                                    currentLineText: line.substring(1),
                                  })
                                }
                              }}
                              className="opacity-0 group-hover:opacity-100 p-0.5 rounded bg-primary text-primary-foreground hover:scale-110 transition-all cursor-pointer ml-1"
                            >
                              <Plus className="w-2.5 h-2.5" />
                            </button>
                          )}
                        </div>
                      ) : (
                        <div className="w-16 shrink-0 bg-secondary/40 border-r border-border" />
                      )}

                      {/* Code line content */}
                      <div className="px-3 py-0.5 whitespace-pre flex-1 select-text">
                        {line}
                      </div>
                    </div>

                    {/* Inline Discussion Thread for this line */}
                    {lineComments.length > 0 && (
                      <div className="bg-secondary/15 p-3 space-y-3 border-y border-border">
                        {lineComments.map((comment) => {
                          const suggestion = extractSuggestion(comment.body)
                          const isApplied = appliedSuggestions[comment.id]

                          return (
                            <div
                              key={comment.id}
                              className="rounded-md border border-border bg-card p-3 space-y-2 text-xs font-sans shadow-2xs"
                            >
                              {/* Comment Header */}
                              <div className="flex items-center justify-between text-xs">
                                <div className="flex items-center gap-2">
                                  <LivingAvatar name={comment.author} isAgent={comment.isAgent} size="sm" />
                                  <span className="font-semibold text-foreground">{comment.author}</span>
                                  {comment.isAgent && (
                                    <Badge variant="agent" className="text-[9px] py-0 px-1">
                                      AI Agent
                                    </Badge>
                                  )}
                                  <span className="text-muted-foreground text-[11px]">{comment.createdAt}</span>
                                </div>

                                <div className="flex items-center gap-2">
                                  {comment.resolved ? (
                                    <Badge variant="success" className="text-[10px] gap-1">
                                      <Check className="w-3 h-3" />
                                      <span>Resolved</span>
                                    </Badge>
                                  ) : (
                                    <Button
                                      variant="ghost"
                                      size="xs"
                                      onClick={() => handleToggleResolve(comment.threadId || comment.id, false)}
                                      className="text-[11px] text-muted-foreground hover:text-foreground h-6 px-1.5"
                                    >
                                      Resolve conversation
                                    </Button>
                                  )}
                                </div>
                              </div>

                              {/* Comment Body */}
                              <div className="text-foreground leading-relaxed whitespace-pre-wrap">
                                {comment.body.replace(/```suggestion\r?\n[\s\S]*?\r?\n```/g, '').trim()}
                              </div>

                              {/* Suggestion Block Preview */}
                              {suggestion !== null && (
                                <div className="mt-2 rounded-md border border-border bg-secondary/30 overflow-hidden font-mono text-xs">
                                  <div className="px-3 py-1.5 bg-secondary/60 border-b border-border flex items-center justify-between">
                                    <div className="flex items-center gap-1.5 text-muted-foreground font-sans font-medium text-[11px]">
                                      <GitCommit className="w-3.5 h-3.5 text-primary" />
                                      <span>Suggested change</span>
                                    </div>

                                    {isPrOpen && (
                                      <Button
                                        variant={isApplied ? 'secondary' : 'primary'}
                                        size="xs"
                                        disabled={isApplied || applyingSuggestionId === comment.id}
                                        onClick={() => handleApplySuggestion(comment.id)}
                                        className="h-6 gap-1 text-[11px]"
                                      >
                                        {applyingSuggestionId === comment.id ? (
                                          <Loader2 className="w-3 h-3 animate-spin" />
                                        ) : isApplied ? (
                                          <Check className="w-3 h-3 text-emerald-500" />
                                        ) : (
                                          <Check className="w-3 h-3" />
                                        )}
                                        <span>{isApplied ? 'Applied' : 'Apply suggestion'}</span>
                                      </Button>
                                    )}
                                  </div>

                                  <div className="p-2 space-y-0.5">
                                    <div className="px-2 py-0.5 rounded bg-rose-500/15 text-rose-400">
                                      - {line.substring(1) || 'original'}
                                    </div>
                                    <div className="px-2 py-0.5 rounded bg-emerald-500/15 text-emerald-400">
                                      + {suggestion}
                                    </div>
                                  </div>
                                </div>
                              )}
                            </div>
                          )
                        })}

                        {/* Reply Composer inside thread */}
                        {replyingThreadId === (lineComments[0].threadId || lineComments[0].id) ? (
                          <div className="p-2.5 rounded-md border border-border bg-card space-y-2 font-sans">
                            <textarea
                              rows={2}
                              value={replyText}
                              onChange={(e) => setReplyText(e.target.value)}
                              placeholder="Reply..."
                              className="w-full bg-input text-foreground border border-border rounded-md p-2 text-xs outline-none focus:border-primary resize-y"
                            />
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                variant="ghost"
                                size="xs"
                                onClick={() => {
                                  setReplyingThreadId(null)
                                  setReplyText('')
                                }}
                              >
                                Cancel
                              </Button>
                              <Button
                                variant="primary"
                                size="xs"
                                disabled={isSubmittingReply || !replyText.trim()}
                                onClick={() =>
                                  handleSendReply(
                                    lineComments[0].threadId || lineComments[0].id,
                                    diff.newPath,
                                    lineNumForComment
                                  )
                                }
                                className="gap-1"
                              >
                                {isSubmittingReply && <Loader2 className="w-3 h-3 animate-spin" />}
                                <span>Reply</span>
                              </Button>
                            </div>
                          </div>
                        ) : (
                          <Button
                            variant="ghost"
                            size="xs"
                            onClick={() => setReplyingThreadId(lineComments[0].threadId || lineComments[0].id)}
                            className="text-[11px] text-muted-foreground hover:text-foreground gap-1"
                          >
                            <CornerDownRight className="w-3 h-3" />
                            <span>Reply to thread...</span>
                          </Button>
                        )}
                      </div>
                    )}

                    {/* Inline Comment Composer for this line */}
                    {isComposerOpen && (
                      <div
                        data-testid="inline-comment-composer"
                        className="bg-secondary/20 p-3 border-y border-border font-sans space-y-2.5"
                      >
                        <div className="flex items-center justify-between text-xs font-semibold text-foreground">
                          <span>
                            Comment on {diff.newPath}:{lineNumForComment}
                          </span>
                          <Button
                            type="button"
                            variant="outline"
                            size="xs"
                            onClick={() => {
                              const template = `\`\`\`suggestion\n${activeComposer.currentLineText}\n\`\`\``
                              setComposerText((prev) => (prev ? `${prev}\n\n${template}` : template))
                            }}
                            className="text-[11px] gap-1 h-6"
                          >
                            <GitCommit className="w-3 h-3 text-primary" />
                            <span>Add a suggestion</span>
                          </Button>
                        </div>

                        <textarea
                          rows={3}
                          value={composerText}
                          onChange={(e) => setComposerText(e.target.value)}
                          placeholder="Leave a comment or review suggestion..."
                          className="w-full bg-input text-foreground border border-border rounded-md p-2.5 text-xs outline-none focus:border-primary resize-y"
                        />

                        <div className="flex items-center justify-end gap-2 text-xs">
                          <Button
                            variant="ghost"
                            size="xs"
                            onClick={() => {
                              setActiveComposer(null)
                              setComposerText('')
                            }}
                          >
                            Cancel
                          </Button>
                          <Button
                            variant="primary"
                            size="xs"
                            disabled={isSubmittingComment || !composerText.trim()}
                            onClick={handleAddLineComment}
                            className="gap-1.5"
                          >
                            {isSubmittingComment ? (
                              <Loader2 className="w-3 h-3 animate-spin" />
                            ) : (
                              <Send className="w-3 h-3" />
                            )}
                            <span>Add single comment</span>
                          </Button>
                        </div>
                      </div>
                    )}
                  </React.Fragment>
                )
              })}
            </div>
          </Box>
        )
      })}
    </div>
  )
}
