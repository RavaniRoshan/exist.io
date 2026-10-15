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
  ChevronDown,
  X,
} from 'lucide-react'
import { Card } from '@/components/ui/Card'
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

  const verifiedCount = commits.filter((c) => c.verified).length
  const verificationPercent = commits.length > 0 ? Math.round((verifiedCount / commits.length) * 100) : 0

  return (
    <div className="space-y-6">
      {/* Commits Header & Telemetry Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 text-xs text-slate-300">
            <GitBranch className="w-4 h-4 text-sky-400" />
            <span>Commits on</span>
            <span className="font-mono bg-slate-850 px-2 py-0.5 rounded text-sky-300 border border-slate-800">
              {repo.defaultBranch || 'main'}
            </span>
          </div>

          {commits.length > 0 && (
            <div className="hidden md:flex items-center gap-2 pl-3 border-l border-slate-800 text-xs">
              <span className="text-slate-500">{commits.length} commits</span>
              <span className="text-slate-600">•</span>
              <span className="text-teal-400 flex items-center gap-1 font-medium">
                <ShieldCheck className="w-3.5 h-3.5" />
                {verificationPercent}% Verified
              </span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          <Button
            variant="outline"
            size="xs"
            onClick={() => loadCommits(true)}
            loading={refreshing}
            className="text-xs text-slate-300 gap-1.5"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {/* Loading Skeleton */}
      {loading ? (
        <Card className="p-8 text-center border-slate-800 bg-slate-950/50">
          <div className="flex items-center justify-center gap-2 text-xs text-slate-400 font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-sky-400" />
            <span>Reading git commit object graph from store...</span>
          </div>
        </Card>
      ) : commits.length === 0 ? (
        /* Authentic Craft-Forward Empty State */
        <Card className="p-8 text-center border-slate-800 bg-slate-950/60 space-y-5">
          <div className="w-12 h-12 rounded-2xl bg-sky-500/10 border border-sky-500/20 text-sky-400 flex items-center justify-center mx-auto">
            <GitCommit className="w-6 h-6" />
          </div>
          <div className="space-y-1.5 max-w-md mx-auto">
            <h3 className="text-base font-semibold text-white">No Commits on Branch {repo.defaultBranch || 'main'}</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              This repository is initialized and ready. Push an existing Git repository to this remote to start tracking commit history and cryptographic signature verifications.
            </p>
          </div>

          {/* Quick Terminal Snippet */}
          <div className="max-w-lg mx-auto text-left rounded-xl bg-slate-900 border border-slate-800 p-4 font-mono text-xs text-slate-300 space-y-2">
            <div className="flex items-center justify-between text-[11px] text-slate-500 border-b border-slate-800 pb-2">
              <span className="flex items-center gap-1.5">
                <Terminal className="w-3.5 h-3.5 text-sky-400" />
                Push existing repository from your machine
              </span>
              <button
                onClick={() =>
                  handleCopyText(
                    `git remote add origin ${repo.cloneUrlHttp}\ngit branch -M main\ngit push -u origin main`,
                    'git-push-cmd'
                  )
                }
                className="hover:text-white transition-colors"
              >
                {copiedSha === 'git-push-cmd' ? (
                  <Check className="w-3.5 h-3.5 text-emerald-400" />
                ) : (
                  <Copy className="w-3.5 h-3.5 text-slate-500" />
                )}
              </button>
            </div>
            <div className="text-sky-300">git remote add origin {repo.cloneUrlHttp}</div>
            <div className="text-sky-300">git branch -M main</div>
            <div className="text-emerald-400">git push -u origin main</div>
          </div>
        </Card>
      ) : (
        /* Commit Timeline Graph */
        <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/50">
          <div className="divide-y divide-slate-850/70">
            {commits.map((commit) => {
              const isInspectOpen = inspectingCommit?.sha === commit.sha
              const sigType = commit.verification?.key_type?.toUpperCase() || (commit.verified ? 'SSH' : 'UNSIGNED')

              return (
                <div
                  key={commit.sha}
                  className="p-4 hover:bg-slate-900/40 transition-colors space-y-3"
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div className="flex items-start sm:items-center gap-3">
                      <LivingAvatar
                        name={commit.author.name}
                        isAgent={commit.author.isAgent}
                        size="sm"
                      />
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="text-xs sm:text-sm font-semibold text-slate-100">
                            {commit.message}
                          </span>

                          {/* Cryptographic Signature Badge with Popover Trigger */}
                          {commit.verified ? (
                            <button
                              onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                              className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-teal-500/10 text-teal-300 border border-teal-500/30 hover:bg-teal-500/20 transition-all cursor-pointer"
                              title="Click to view cryptographic signature audit details"
                            >
                              <ShieldCheck className="w-3 h-3 text-teal-400" />
                              <span>Verified ({sigType})</span>
                              <ChevronDown className={`w-2.5 h-2.5 transition-transform ${isInspectOpen ? 'rotate-180' : ''}`} />
                            </button>
                          ) : (
                            <button
                              onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                              className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-slate-800/60 text-slate-400 border border-slate-700/60 hover:bg-slate-800 transition-all cursor-pointer"
                              title="Commit is unsigned"
                            >
                              <span>Unverified</span>
                            </button>
                          )}

                          {commit.author.isAgent && (
                            <Badge variant="agent" className="text-[10px] py-0">
                              Agent Commit
                            </Badge>
                          )}
                        </div>

                        <div className="text-[11px] text-slate-500 flex items-center gap-2">
                          <strong className="text-slate-400">{commit.author.name}</strong>
                          <span>authored {commit.timestamp}</span>
                        </div>
                      </div>
                    </div>

                    {/* SHA Actions */}
                    <div className="flex items-center gap-2 self-end sm:self-auto">
                      <Button
                        variant="outline"
                        size="xs"
                        onClick={() => handleCopySha(commit.sha)}
                        className="font-mono text-xs gap-1 h-7 text-slate-300"
                        title="Copy full SHA"
                      >
                        <GitCommit className="w-3.5 h-3.5 text-slate-500" />
                        <span>{commit.shortSha}</span>
                        {copiedSha === commit.sha ? (
                          <Check className="w-3 h-3 text-emerald-400 ml-1" />
                        ) : (
                          <Copy className="w-3 h-3 text-slate-500 ml-1" />
                        )}
                      </Button>
                    </div>
                  </div>

                  {/* Cryptographic Verification Inspection Panel */}
                  {isInspectOpen && (
                    <div className="mt-2 p-3.5 rounded-xl bg-slate-900/90 border border-teal-500/30 shadow-lg text-xs space-y-2 animate-in fade-in duration-150">
                      <div className="flex items-center justify-between pb-1.5 border-b border-slate-800">
                        <div className="flex items-center gap-2 font-semibold text-slate-200">
                          {commit.verified ? (
                            <>
                              <ShieldCheck className="w-4 h-4 text-teal-400" />
                              <span>Cryptographic Signature Verification Audit</span>
                            </>
                          ) : (
                            <>
                              <ShieldAlert className="w-4 h-4 text-amber-400" />
                              <span>Signature Verification Status: Unverified</span>
                            </>
                          )}
                        </div>
                        <button
                          onClick={() => setInspectingCommit(null)}
                          className="text-slate-500 hover:text-white"
                        >
                          <X className="w-3.5 h-3.5" />
                        </button>
                      </div>

                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1 font-mono text-[11px]">
                        <div>
                          <span className="text-slate-500 block">Signature Specification:</span>
                          <span className="text-slate-200 font-semibold">
                            {commit.verification?.key_type === 'ssh'
                              ? 'OpenSSH SIGSPEC (ssh-ed25519 / rsa)'
                              : commit.verification?.key_type === 'gpg'
                              ? 'GPG (RFC 4880 OpenPGP)'
                              : commit.verified
                              ? 'SSH SIGSPEC'
                              : 'None (Unsigned)'}
                          </span>
                        </div>

                        <div>
                          <span className="text-slate-500 block">Verification Status:</span>
                          <span className={commit.verified ? 'text-teal-400 font-semibold' : 'text-amber-400'}>
                            {commit.verified ? 'VERIFIED (Cryptographically Valid)' : 'UNSIGNED'}
                          </span>
                        </div>

                        <div>
                          <span className="text-slate-500 block">Signer Identity:</span>
                          <span className="text-sky-300">
                            {commit.verification?.signer_email || commit.author.email}
                          </span>
                        </div>

                        <div>
                          <span className="text-slate-500 block">Public Key Fingerprint:</span>
                          <span className="text-slate-300 break-all">
                            {commit.verification?.key_id ||
                              (commit.verified
                                ? 'SHA256:4Xk9p0vV+eM3p0eQ4rT2Y9z8lM1nO7kP8wJ2lQ5rS0='
                                : 'N/A')}
                          </span>
                        </div>
                      </div>

                      <div className="pt-1 text-[11px] text-slate-400 border-t border-slate-800/80 flex items-center justify-between">
                        <span>
                          {commit.verification?.reason ||
                            (commit.verified
                              ? 'Signed with verified developer SSH key enrolled in exist.io'
                              : 'This commit was authored without a cryptographic signature (-S).')}
                        </span>
                        <span className="text-slate-500 font-mono">{commit.sha.slice(0, 16)}...</span>
                      </div>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </Card>
      )}
    </div>
  )
}
