import React, { useState } from 'react'
import { GitCommit, Copy, Check, ShieldCheck, GitBranch } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { sampleCommits, type CommitItem } from '@/lib/api'

export const CommitsView: React.FC = () => {
  const [commits] = useState<CommitItem[]>(sampleCommits)
  const [copiedSha, setCopiedSha] = useState<string | null>(null)

  const handleCopySha = (sha: string) => {
    navigator.clipboard.writeText(sha)
    setCopiedSha(sha)
    setTimeout(() => setCopiedSha(null), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Commits Header Bar */}
      <div className="flex items-center justify-between p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-2 text-xs text-slate-300">
          <GitBranch className="w-4 h-4 text-sky-400" />
          <span>Commits on</span>
          <span className="font-mono bg-slate-850 px-2 py-0.5 rounded text-sky-300 border border-slate-800">main</span>
        </div>
        <span className="text-xs text-slate-500">{commits.length} commits in history</span>
      </div>

      {/* Commit Timeline Graph */}
      <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/50">
        <div className="divide-y divide-slate-850/70">
          {commits.map((commit) => (
            <div
              key={commit.sha}
              className="flex flex-col sm:flex-row sm:items-center justify-between p-4 hover:bg-slate-900/50 transition-colors gap-3"
            >
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
                    {commit.verified && (
                      <Badge variant="outline" className="text-[10px] text-teal-400 border-teal-500/40 gap-1 py-0">
                        <ShieldCheck className="w-3 h-3 text-teal-400" />
                        <span>Verified</span>
                      </Badge>
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
          ))}
        </div>
      </Card>
    </div>
  )
}
