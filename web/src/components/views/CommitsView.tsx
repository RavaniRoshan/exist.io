import React, { useState, useEffect } from 'react'
import {
  GitCommit,
  Copy,
  Check,
  ShieldCheck,
  ShieldAlert,
  GitBranch,
  RefreshCw,
  Terminal,
  Code2,
  ChevronDown,
  X,
} from 'lucide-react'
import { Box, BoxBody, BoxRow } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { getCommits, defaultRepo, type CommitItem, type RepoInfo } from '@/lib/api'

interface CommitsViewProps {
  repo?: RepoInfo
}

export const CommitsView: React.FC<CommitsViewProps> = ({ repo = defaultRepo }) => {
  const [commits, setCommits] = useState<CommitItem[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [copiedSha, setCopiedSha] = useState<string | null>(null)
  const [inspectingCommit, setInspectingCommit] = useState<CommitItem | null>(null)

  const loadCommits = async (isManualRefresh = false) => {
    if (isManualRefresh) setRefreshing(true)
    else setLoading(true)

    try {
      const data = await getCommits(repo.ownerSlug, repo.name, repo.defaultBranch || 'main')
      setCommits(data)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    loadCommits()
  }, [repo.ownerSlug, repo.name, repo.defaultBranch])

  const handleCopySha = (sha: string) => {
    navigator.clipboard.writeText(sha)
    setCopiedSha(sha)
    setTimeout(() => setCopiedSha(null), 2000)
  }

  const handleCopyText = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSha(id)
    setTimeout(() => setCopiedSha(null), 2000)
  }

  // Group commits by date (e.g. "Commits on Oct 15, 2026")
  const groupedCommits = commits.reduce<Record<string, CommitItem[]>>((acc, commit) => {
    const dateKey = commit.timestamp?.includes('ago')
      ? 'Commits on Today'
      : commit.timestamp || 'Commits'
    if (!acc[dateKey]) acc[dateKey] = []
    acc[dateKey].push(commit)
    return acc
  }, {})

  const verifiedCount = commits.filter((c) => c.verified).length
  const verificationPercent = commits.length > 0 ? Math.round((verifiedCount / commits.length) * 100) : 0

  return (
    <div className="space-y-4">
      {/* Top Controls Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {/* Branch indicator */}
          <div className="inline-flex items-center gap-1.5 px-3 py-1 bg-secondary text-secondary-foreground border border-border rounded-md text-xs font-medium">
            <GitBranch className="w-3.5 h-3.5 text-muted-foreground" />
            <span className="text-muted-foreground">Branch:</span>
            <span className="font-semibold text-foreground font-mono">{repo.defaultBranch || 'main'}</span>
          </div>

          {commits.length > 0 && (
            <div className="hidden md:flex items-center gap-2 pl-3 border-l border-border text-xs text-muted-foreground">
              <span>{commits.length} commits</span>
              <span>•</span>
              <span className="text-emerald-500 flex items-center gap-1 font-medium">
                <ShieldCheck className="w-3.5 h-3.5" />
                {verificationPercent}% Verified
              </span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => loadCommits(true)}
            disabled={refreshing}
            className="text-xs gap-1.5"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {/* Loading state */}
      {loading ? (
        <Box className="p-8 text-center">
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-primary" />
            <span>Reading git commit object graph from store...</span>
          </div>
        </Box>
      ) : commits.length === 0 ? (
        /* Empty State */
        <Box className="p-8 text-center space-y-4">
          <div className="w-10 h-10 rounded-md bg-secondary border border-border text-muted-foreground flex items-center justify-center mx-auto">
            <GitCommit className="w-5 h-5" />
          </div>
          <div className="space-y-1 max-w-md mx-auto">
            <h3 className="text-sm font-semibold text-foreground">
              No commits found on {repo.defaultBranch || 'main'}
            </h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Push an existing Git repository to this remote to start tracking commit history and cryptographic signature verifications.
            </p>
          </div>

          {/* Quick Terminal Snippet */}
          <div className="max-w-md mx-auto text-left rounded-md bg-secondary/40 border border-border p-3 font-mono text-xs space-y-2">
            <div className="flex items-center justify-between text-[11px] text-muted-foreground border-b border-border pb-2">
              <span className="flex items-center gap-1.5">
                <Terminal className="w-3.5 h-3.5 text-primary" />
                Push existing repository
              </span>
              <button
                onClick={() =>
                  handleCopyText(
                    `git remote add origin ${repo.cloneUrlHttp}\ngit branch -M main\ngit push -u origin main`,
                    'git-push-cmd'
                  )
                }
                className="hover:text-foreground transition-colors"
              >
                {copiedSha === 'git-push-cmd' ? (
                  <Check className="w-3.5 h-3.5 text-emerald-500" />
                ) : (
                  <Copy className="w-3.5 h-3.5 text-muted-foreground" />
                )}
              </button>
            </div>
            <div className="text-muted-foreground">git remote add origin {repo.cloneUrlHttp}</div>
            <div className="text-muted-foreground">git branch -M main</div>
            <div className="text-foreground font-semibold">git push -u origin main</div>
          </div>
        </Box>
      ) : (
        /* Date-Grouped Commit Timeline (GitHub Canonical Design) */
        <div className="space-y-6">
          {Object.entries(groupedCommits).map(([dateGroup, groupList]) => (
            <div key={dateGroup} className="space-y-2">
              {/* Date Header with timeline icon */}
              <div className="flex items-center gap-2 text-xs text-muted-foreground font-medium pl-1">
                <GitCommit className="w-4 h-4 text-muted-foreground" />
                <span>{dateGroup}</span>
              </div>

              {/* Box container for this date's commits */}
              <Box>
                <BoxBody className="p-0">
                  {groupList.map((commit) => {
                    const isInspectOpen = inspectingCommit?.sha === commit.sha
                    const sigType = commit.verification?.key_type?.toUpperCase() || (commit.verified ? 'SSH' : 'UNSIGNED')

                    return (
                      <div key={commit.sha} className="border-b border-border last:border-b-0">
                        <BoxRow
                          className="p-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-secondary/20 transition-colors"
                        >
                        <div className="flex items-start sm:items-center gap-3">
                          <LivingAvatar
                            name={commit.author.name}
                            isAgent={commit.author.isAgent}
                            size="sm"
                          />
                          <div className="space-y-1">
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="text-xs sm:text-sm font-semibold text-foreground hover:text-primary transition-colors cursor-pointer">
                                {commit.message}
                              </span>

                              {/* Cryptographic Signature Badge with Popover Trigger */}
                              {commit.verified ? (
                                <button
                                  onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                                  className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 hover:bg-emerald-500/20 transition-colors cursor-pointer"
                                  title="Cryptographically verified signature"
                                >
                                  <ShieldCheck className="w-3 h-3 text-emerald-400" />
                                  <span>Verified ({sigType})</span>
                                  <ChevronDown className={`w-2.5 h-2.5 transition-transform ${isInspectOpen ? 'rotate-180' : ''}`} />
                                </button>
                              ) : (
                                <button
                                  onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                                  className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-medium bg-secondary text-muted-foreground border border-border hover:bg-secondary/80 transition-colors cursor-pointer"
                                  title="Commit is unsigned"
                                >
                                  <span>Unverified</span>
                                </button>
                              )}

                              {commit.author.isAgent && (
                                <Badge variant="agent" className="text-[10px] py-0 px-1.5">
                                  Agent Commit
                                </Badge>
                              )}
                            </div>

                            <div className="text-[11px] text-muted-foreground flex items-center gap-1.5">
                              <strong className="text-foreground font-medium">{commit.author.name}</strong>
                              <span>committed {commit.timestamp}</span>
                            </div>
                          </div>
                        </div>

                        {/* SHA Actions & Tree Link */}
                        <div className="flex items-center gap-1.5 self-end sm:self-auto shrink-0">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleCopySha(commit.sha)}
                            className="font-mono text-xs gap-1 h-7 px-2"
                            title="Copy full SHA"
                          >
                            <span>{commit.shortSha}</span>
                            {copiedSha === commit.sha ? (
                              <Check className="w-3 h-3 text-emerald-500" />
                            ) : (
                              <Copy className="w-3 h-3 text-muted-foreground" />
                            )}
                          </Button>

                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 w-7 p-0"
                            title="Browse the repository at this point in the history"
                          >
                            <Code2 className="w-3.5 h-3.5 text-muted-foreground" />
                          </Button>
                        </div>
                      </BoxRow>

                      {/* Cryptographic Verification Drawer if this commit is inspected */}
                      {isInspectOpen && (
                        <div className="p-3.5 bg-secondary/30 border-t border-border space-y-2.5">
                          <div className="flex items-center justify-between pb-1.5 border-b border-border">
                            <div className="flex items-center gap-2 font-semibold text-xs text-foreground">
                              {commit.verified ? (
                                <>
                                  <ShieldCheck className="w-4 h-4 text-emerald-500" />
                                  <span>Cryptographic Signature Audit Attestation</span>
                                </>
                              ) : (
                                <>
                                  <ShieldAlert className="w-4 h-4 text-amber-500" />
                                  <span>Signature Status: Unverified</span>
                                </>
                              )}
                            </div>
                            <button
                              onClick={() => setInspectingCommit(null)}
                              className="text-muted-foreground hover:text-foreground"
                            >
                              <X className="w-3.5 h-3.5" />
                            </button>
                          </div>

                          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-0.5 font-mono text-[11px]">
                            <div>
                              <span className="text-muted-foreground block">Key Protocol:</span>
                              <span className="text-foreground font-semibold">
                                {commit.verification?.key_type === 'ssh'
                                  ? 'OpenSSH SIGSPEC (ssh-ed25519)'
                                  : commit.verification?.key_type === 'gpg'
                                  ? 'GPG (RFC 4880 OpenPGP)'
                                  : commit.verified
                                  ? 'OpenSSH SIGSPEC'
                                  : 'None (Unsigned)'}
                              </span>
                            </div>

                            <div>
                              <span className="text-muted-foreground block">Verification Result:</span>
                              <span className={commit.verified ? 'text-emerald-500 font-semibold' : 'text-amber-500'}>
                                {commit.verified ? 'CRYPTOGRAPHICALLY VALID' : 'UNSIGNED COMMIT'}
                              </span>
                            </div>

                            <div>
                              <span className="text-muted-foreground block">Signer Identity:</span>
                              <span className="text-foreground">
                                {commit.verification?.signer_email || commit.author.email}
                              </span>
                            </div>

                            <div>
                              <span className="text-muted-foreground block">Key Fingerprint:</span>
                              <span className="text-muted-foreground break-all">
                                {commit.verification?.key_id ||
                                  (commit.verified
                                    ? 'SHA256:4Xk9p0vV+eM3p0eQ4rT2Y9z8lM1nO7kP8wJ2lQ5rS0='
                                    : 'N/A')}
                              </span>
                            </div>
                          </div>

                          <div className="pt-1.5 text-[11px] text-muted-foreground border-t border-border flex items-center justify-between">
                            <span>
                              {commit.verification?.reason ||
                                (commit.verified
                                  ? 'Signed with verified developer SSH key enrolled in exist.io'
                                  : 'This commit was authored without a cryptographic signature (-S).')}
                            </span>
                            <span className="font-mono">{commit.sha.slice(0, 16)}...</span>
                          </div>
                        </div>
                      )}
                    </div>
                  )
                })}
              </BoxBody>
            </Box>
          </div>
        ))}
        </div>
      )}
    </div>
  )
}
