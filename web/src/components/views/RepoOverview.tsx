import React, { useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import { Folder, FileText, GitBranch, Clock, Copy, Check, ChevronRight, FileCode, ArrowLeft, Shield } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { type RepoInfo, type FileItem, sampleFiles, sampleReadme } from '@/lib/api'

export interface RepoOverviewProps {
  repo: RepoInfo
  onSelectCommits: () => void
  onSelectAgentConsole: () => void
}

export const RepoOverview: React.FC<RepoOverviewProps> = ({
  repo,
  onSelectCommits,
  onSelectAgentConsole: _onSelectAgentConsole,
}) => {
  const [selectedFile, setSelectedFile] = useState<FileItem | null>(null)
  const [copiedCode, setCopiedCode] = useState(false)
  const [currentBranch, setCurrentBranch] = useState(repo.defaultBranch)

  const handleCopyCode = (code: string) => {
    navigator.clipboard.writeText(code)
    setCopiedCode(true)
    setTimeout(() => setCopiedCode(false), 2000)
  }

  const getFileContent = (file: FileItem) => {
    if (file.name === 'README.md') return sampleReadme
    if (file.name === 'go.mod') {
      return `module github.com/exist-io/exist

go 1.24

require (
\tgithub.com/go-git/go-git/v6 v6.15.0
\tmodernc.org/sqlite v1.34.4
)
`
    }
    if (file.name === '.gitignore') {
      return `/bin/
/dist/
/data/
*.db
*.db-wal
*.db-shm
.oxlintrc.json
`
    }
    return `// Sample contents of ${file.path}
package main

import (
\t"fmt"
\t"net/http"
)

func main() {
\tfmt.Println("exist.io autonomous agent Git forge active.")
}
`
  }

  return (
    <div className="space-y-6">
      {/* Repo Meta Hero Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3">
          {/* Branch Switcher Pill */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-950 border border-slate-800 text-xs font-medium text-slate-200">
            <GitBranch className="w-3.5 h-3.5 text-sky-400" />
            <select
              value={currentBranch}
              onChange={(e) => setCurrentBranch(e.target.value)}
              className="bg-transparent border-none outline-none cursor-pointer pr-1"
            >
              <option value="main" className="bg-slate-900 text-slate-100">main</option>
              <option value="feature/agent-checks" className="bg-slate-900 text-slate-100">feature/agent-checks</option>
              <option value="release/v0.1.0" className="bg-slate-900 text-slate-100">release/v0.1.0</option>
            </select>
          </div>

          <span className="text-xs text-slate-500">|</span>

          {/* Quick stats */}
          <div className="flex items-center gap-3 text-xs text-slate-400">
            <button
              onClick={onSelectCommits}
              className="hover:text-sky-400 flex items-center gap-1 transition-colors"
            >
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              <strong className="text-slate-200 font-semibold">148</strong> commits
            </button>
            <span className="hidden sm:inline text-slate-500">•</span>
            <span className="hidden sm:inline">
              <strong className="text-slate-200 font-semibold">4</strong> branches
            </span>
            <span className="hidden sm:inline text-slate-500">•</span>
            <span className="hidden sm:inline">
              <strong className="text-slate-200 font-semibold">2</strong> tags
            </span>
          </div>
        </div>

        {/* Security & Ruleset Active Indicator */}
        <div className="flex items-center gap-2">
          <Badge variant="success" pulse className="gap-1.5 py-1">
            <Shield className="w-3.5 h-3.5 text-teal-300" />
            <span>Branch Ruleset Enforced</span>
          </Badge>
        </div>
      </div>

      <AnimatePresence mode="wait">
        {selectedFile ? (
          /* File Blob Viewer */
          <motion.div
            key="file-viewer"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
          >
            <Card className="overflow-hidden p-0 border-slate-800 bg-slate-950/70">
              <div className="flex items-center justify-between px-4 py-3 border-b border-slate-800 bg-slate-900/80">
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="xs"
                    onClick={() => setSelectedFile(null)}
                    className="p-1 h-7 w-7 text-slate-400 hover:text-white"
                  >
                    <ArrowLeft className="w-4 h-4" />
                  </Button>
                  <FileCode className="w-4 h-4 text-sky-400" />
                  <span className="text-xs font-mono text-slate-200 font-medium">{selectedFile.path}</span>
                  {selectedFile.size && (
                    <span className="text-[10px] text-slate-500 font-mono">({selectedFile.size} bytes)</span>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="xs"
                    onClick={() => handleCopyCode(getFileContent(selectedFile))}
                    className="gap-1 text-slate-300"
                  >
                    {copiedCode ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                    <span>{copiedCode ? 'Copied' : 'Copy'}</span>
                  </Button>
                </div>
              </div>

              {/* Code Line Canvas */}
              <div className="p-4 overflow-x-auto font-mono text-xs leading-relaxed text-slate-200 bg-slate-950/90 max-h-[600px] overflow-y-auto">
                <pre>
                  {getFileContent(selectedFile).split('\n').map((line, idx) => (
                    <div key={idx} className="flex hover:bg-slate-900/60 px-2 rounded">
                      <span className="w-10 select-none text-right pr-4 text-slate-600 font-mono text-[11px] shrink-0">
                        {idx + 1}
                      </span>
                      <span className="flex-1 whitespace-pre">{line}</span>
                    </div>
                  ))}
                </pre>
              </div>
            </Card>
          </motion.div>
        ) : (
          /* Main Directory Tree & README */
          <motion.div
            key="directory-tree"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="space-y-6"
          >
            {/* File List Card */}
            <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/40">
              {/* Directory Table Header */}
              <div className="flex items-center justify-between px-4 py-2.5 bg-slate-900/80 border-b border-slate-800 text-xs text-slate-400 font-medium">
                <div className="flex items-center gap-2">
                  <span>commit</span>
                  <span className="font-mono text-sky-400 font-semibold">7aa8446</span>
                  <span className="text-slate-500 truncate max-w-sm sm:max-w-md">
                    docs: expand Diátaxis site to 15 pages and install global agent skills
                  </span>
                </div>
                <span className="text-[11px] text-slate-500 shrink-0">2 hours ago</span>
              </div>

              {/* File Rows */}
              <div className="divide-y divide-slate-850/60">
                {sampleFiles.map((file) => (
                  <div
                    key={file.path}
                    onClick={() => {
                      if (file.type === 'blob') setSelectedFile(file)
                    }}
                    className="flex items-center justify-between px-4 py-2.5 hover:bg-slate-900/60 transition-colors cursor-pointer text-xs group"
                  >
                    <div className="flex items-center gap-2.5 min-w-0 flex-1 pr-4">
                      {file.type === 'tree' ? (
                        <Folder className="w-4 h-4 text-sky-400 shrink-0 fill-sky-500/20" />
                      ) : (
                        <FileText className="w-4 h-4 text-slate-400 shrink-0 group-hover:text-sky-300 transition-colors" />
                      )}
                      <span className="font-mono text-slate-200 group-hover:text-sky-400 transition-colors truncate">
                        {file.name}
                      </span>
                    </div>
                    <div className="hidden sm:block text-slate-500 text-xs truncate max-w-xs md:max-w-md flex-1 text-left px-2">
                      {file.lastCommitMessage}
                    </div>
                    <div className="flex items-center gap-2 text-[11px] text-slate-500 shrink-0">
                      <span>{file.lastCommitDate}</span>
                      <ChevronRight className="w-3.5 h-3.5 text-slate-600 group-hover:text-slate-400 transition-colors" />
                    </div>
                  </div>
                ))}
              </div>
            </Card>

            {/* Readme Card */}
            <Card className="border-slate-800 bg-slate-950/60 p-6 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-slate-300">
                  <FileText className="w-4 h-4 text-sky-400" />
                  <span>README.md</span>
                </div>
                <Badge variant="outline" className="text-[10px] py-0 border-slate-700">
                  Markdown
                </Badge>
              </div>

              {/* Formatted Markdown Content */}
              <div className="prose prose-invert prose-slate max-w-none text-xs sm:text-sm space-y-3">
                <h2 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
                  <span className="text-sky-400">exist.io</span>
                  <span className="text-xs font-normal text-slate-400 bg-slate-800/80 px-2 py-0.5 rounded-full border border-slate-700">
                    The Autonomous Agent Git Forge
                  </span>
                </h2>
                <p className="text-slate-300 leading-relaxed">
                  exist.io is a high-performance, hermetic Git forge engine built in pure Go. It delivers seamless 
                  Smart-HTTP Git compatibility alongside first-class Model Context Protocol (MCP) server endpoints, 
                  allowing both human developers and autonomous AI coding agents to collaborate with zero frictions.
                </p>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                  <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 space-y-1">
                    <h4 className="text-xs font-semibold text-sky-300 flex items-center gap-1.5">
                      <GitBranch className="w-3.5 h-3.5" /> Smart-HTTP Engine
                    </h4>
                    <p className="text-[11px] text-slate-400">
                      Pure-Go Git transport supporting push, fetch, and clone without relying on external system binaries.
                    </p>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 space-y-1">
                    <h4 className="text-xs font-semibold text-purple-300 flex items-center gap-1.5">
                      <Shield className="w-3.5 h-3.5" /> Policy & Rulesets
                    </h4>
                    <p className="text-[11px] text-slate-400">
                      Enforce passing CI checks, review approvals, linear history, and CIDR IP allowlists.
                    </p>
                  </div>
                </div>
              </div>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
