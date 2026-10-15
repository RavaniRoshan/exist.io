import React, { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  Folder,
  FileText,
  GitBranch,
  Clock,
  Copy,
  Check,
  ChevronRight,
  FileCode,
  ArrowLeft,
  Shield,
  Search,
  Code2,
  Terminal,
  RefreshCw,
  Sparkles,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import {
  type RepoInfo,
  type FileItem,
  type SymbolItem,
  getFiles,
  getFileContent,
  searchRepoSymbols,
  isDemoModeActive,
  sampleFiles,
  sampleReadme,
} from '@/lib/api'

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
  const [files, setFiles] = useState<FileItem[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedFile, setSelectedFile] = useState<FileItem | null>(null)
  const [fileContent, setFileContent] = useState<string>('')
  const [loadingContent, setLoadingContent] = useState(false)
  const [copiedCode, setCopiedCode] = useState(false)
  const [copiedClone, setCopiedClone] = useState(false)
  const [currentBranch, setCurrentBranch] = useState(repo.defaultBranch || 'main')

  // Symbol Search State
  const [searchQuery, setSearchQuery] = useState('')
  const [symbolResults, setSymbolResults] = useState<SymbolItem[]>([])
  const [isSearchingSymbols, setIsSearchingSymbols] = useState(false)

  const owner = repo.ownerSlug || 'acme'
  const name = repo.name || 'rocket'

  const loadTree = async () => {
    setLoading(true)
    try {
      const entries = await getFiles(owner, name, currentBranch)
      setFiles(entries)
    } catch {
      if (isDemoModeActive()) {
        setFiles(sampleFiles)
      } else {
        setFiles([])
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadTree()
    const onModeChange = () => loadTree()
    window.addEventListener('exist_mode_changed', onModeChange)
    return () => window.removeEventListener('exist_mode_changed', onModeChange)
  }, [owner, name, currentBranch])

  const handleSelectFile = async (file: FileItem) => {
    if (file.type !== 'blob') return
    setSelectedFile(file)
    setLoadingContent(true)
    try {
      const content = await getFileContent(owner, name, file.path, currentBranch)
      setFileContent(content || (isDemoModeActive() && file.name === 'README.md' ? sampleReadme : ''))
    } catch {
      setFileContent('// Failed to read file content from git repository.')
    } finally {
      setLoadingContent(false)
    }
  }

  // Symbol Search debounced trigger
  useEffect(() => {
    if (!searchQuery.trim()) {
      setSymbolResults([])
      setIsSearchingSymbols(false)
      return
    }

    const timer = setTimeout(async () => {
      setIsSearchingSymbols(true)
      try {
        const syms = await searchRepoSymbols(owner, name, searchQuery.trim(), currentBranch)
        setSymbolResults(syms)
      } catch {
        setSymbolResults([])
      } finally {
        setIsSearchingSymbols(false)
      }
    }, 250)

    return () => clearTimeout(timer)
  }, [searchQuery, owner, name, currentBranch])

  const handleCopyCode = (code: string) => {
    navigator.clipboard.writeText(code)
    setCopiedCode(true)
    setTimeout(() => setCopiedCode(false), 2000)
  }

  const cloneHttp = `http://127.0.0.1:8080/${owner}/${name}.git`
  const handleCopyClone = () => {
    navigator.clipboard.writeText(`git clone ${cloneHttp}`)
    setCopiedClone(true)
    setTimeout(() => setCopiedClone(false), 2000)
  }

  return (
    <div className="space-y-6">
      {/* Repo Meta Hero Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-2xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md">
        <div className="flex items-center gap-3 flex-wrap">
          {/* Branch Switcher Pill */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-950 border border-slate-800 text-xs font-medium text-slate-200">
            <GitBranch className="w-3.5 h-3.5 text-sky-400" />
            <select
              value={currentBranch}
              onChange={(e) => setCurrentBranch(e.target.value)}
              className="bg-transparent border-none outline-none cursor-pointer pr-1"
            >
              <option value="main" className="bg-slate-900 text-slate-100">
                main
              </option>
              <option value="feature/agent-checks" className="bg-slate-900 text-slate-100">
                feature/agent-checks
              </option>
              <option value="release/v0.2.0" className="bg-slate-900 text-slate-100">
                release/v0.2.0
              </option>
            </select>
          </div>

          <span className="text-xs text-slate-500">|</span>

          {/* Quick stats */}
          <div className="flex items-center gap-3 text-xs text-slate-400">
            <button onClick={onSelectCommits} className="hover:text-sky-400 flex items-center gap-1 transition-colors">
              <Clock className="w-3.5 h-3.5 text-slate-500" />
              <strong className="text-slate-200 font-semibold">{repo.commitsCount || 148}</strong> commits
            </button>
            <span className="hidden sm:inline text-slate-500">•</span>
            <span className="hidden sm:inline">
              <strong className="text-slate-200 font-semibold">{repo.branchesCount || 3}</strong> branches
            </span>
          </div>
        </div>

        {/* Clone button & Ruleset Badge */}
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="xs"
            onClick={handleCopyClone}
            className="text-xs font-mono text-slate-300 border border-slate-800 bg-slate-950/80 gap-1.5"
          >
            {copiedClone ? <Check className="w-3 h-3 text-emerald-400" /> : <Terminal className="w-3 h-3 text-sky-400" />}
            <span>{copiedClone ? 'Copied Clone URL' : 'Clone Repo'}</span>
          </Button>

          <Badge variant="success" pulse className="gap-1.5 py-1">
            <Shield className="w-3.5 h-3.5 text-teal-300" />
            <span>Branch Ruleset Enforced</span>
          </Badge>
        </div>
      </div>

      {/* Semantic Symbol Search & File Filter Bar */}
      <div className="relative">
        <div className="flex items-center gap-2 px-3 py-2 rounded-xl bg-slate-900/70 border border-slate-800 text-xs">
          <Search className="w-4 h-4 text-slate-400 shrink-0" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search code symbols (Go AST functions, structs, TS classes, Python defs)..."
            className="w-full bg-transparent border-none outline-none text-slate-200 placeholder:text-slate-500 font-mono text-xs"
          />
          {isSearchingSymbols && <RefreshCw className="w-3.5 h-3.5 animate-spin text-sky-400 shrink-0" />}
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="text-[10px] text-slate-400 hover:text-white bg-slate-800 px-1.5 py-0.5 rounded"
            >
              Clear
            </button>
          )}
        </div>

        {/* Semantic Symbol Search Dropdown */}
        {searchQuery && (
          <div className="absolute top-full left-0 right-0 mt-2 z-20 bg-slate-950 border border-slate-800 rounded-xl shadow-2xl overflow-hidden divide-y divide-slate-850/80 max-h-80 overflow-y-auto">
            <div className="px-3 py-2 bg-slate-900/80 text-[11px] font-semibold text-slate-400 flex items-center justify-between">
              <span className="flex items-center gap-1.5">
                <Sparkles className="w-3.5 h-3.5 text-sky-400" />
                <span>Semantic Symbol Results</span>
              </span>
              <span>{symbolResults.length} matches</span>
            </div>

            {symbolResults.length === 0 ? (
              <div className="p-4 text-center text-xs text-slate-500 font-mono">
                {isSearchingSymbols ? 'Indexing AST symbols...' : 'No symbols found matching query'}
              </div>
            ) : (
              symbolResults.map((s, idx) => (
                <div
                  key={idx}
                  onClick={() => {
                    handleSelectFile({
                      name: s.file.split('/').pop() || s.file,
                      path: s.file,
                      type: 'blob',
                      lastCommitMessage: `Symbol ${s.name}`,
                      lastCommitDate: 'Recently',
                    })
                    setSearchQuery('')
                  }}
                  className="p-3 hover:bg-slate-900/60 transition-colors cursor-pointer flex items-center justify-between gap-3 text-xs"
                >
                  <div className="flex items-center gap-2.5 min-w-0">
                    <Code2 className="w-4 h-4 text-sky-400 shrink-0" />
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-mono font-semibold text-slate-100">{s.name}</span>
                        <Badge variant="outline" className="text-[9px] py-0 px-1.5 border-slate-700 font-mono">
                          {s.kind}
                        </Badge>
                      </div>
                      {s.signature && (
                        <div className="text-[11px] font-mono text-slate-400 truncate">{s.signature}</div>
                      )}
                    </div>
                  </div>

                  <span className="font-mono text-[11px] text-slate-500 shrink-0">
                    {s.file}:{s.line}
                  </span>
                </div>
              ))
            )}
          </div>
        )}
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
                  {selectedFile.size !== undefined && selectedFile.size > 0 && (
                    <span className="text-[10px] text-slate-500 font-mono">({selectedFile.size} bytes)</span>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="xs"
                    onClick={() => handleCopyCode(fileContent)}
                    className="gap-1 text-slate-300"
                  >
                    {copiedCode ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                    <span>{copiedCode ? 'Copied' : 'Copy'}</span>
                  </Button>
                </div>
              </div>

              {/* Code Line Canvas */}
              <div className="p-4 overflow-x-auto font-mono text-xs leading-relaxed text-slate-200 bg-slate-950/90 max-h-[600px] overflow-y-auto">
                {loadingContent ? (
                  <div className="p-8 text-center text-slate-500 text-xs font-mono flex items-center justify-center gap-2">
                    <RefreshCw className="w-4 h-4 animate-spin text-sky-400" />
                    <span>Loading blob contents from gitstore...</span>
                  </div>
                ) : (
                  <pre>
                    {fileContent.split('\n').map((line, idx) => (
                      <div key={idx} className="flex hover:bg-slate-900/60 px-2 rounded">
                        <span className="w-10 select-none text-right pr-4 text-slate-600 font-mono text-[11px] shrink-0">
                          {idx + 1}
                        </span>
                        <span className="flex-1 whitespace-pre">{line}</span>
                      </div>
                    ))}
                  </pre>
                )}
              </div>
            </Card>
          </motion.div>
        ) : (
          /* Main Directory Tree & README */
          <motion.div key="directory-tree" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-6">
            {files.length === 0 ? (
              /* Authentic High-Craft Empty State */
              <Card className="border-slate-800 bg-slate-950/60 p-8 text-center space-y-5">
                <div className="w-12 h-12 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center mx-auto text-sky-400">
                  <Folder className="w-6 h-6" />
                </div>

                <div className="max-w-md mx-auto space-y-1.5">
                  <h3 className="text-base font-semibold text-slate-100">This repository is empty</h3>
                  <p className="text-xs text-slate-400 leading-relaxed">
                    Get started by pushing files from your local environment or granting an autonomous AI agent capability
                    lease to initialize your codebase.
                  </p>
                </div>

                {/* Copyable Quickstart Instructions */}
                <div className="max-w-lg mx-auto text-left bg-slate-900/90 border border-slate-800 rounded-xl p-3 space-y-2">
                  <div className="flex items-center justify-between text-[11px] text-slate-400">
                    <span className="flex items-center gap-1.5 font-mono">
                      <Terminal className="w-3.5 h-3.5 text-sky-400" />
                      <span>Push an existing repository from the command line</span>
                    </span>
                  </div>
                  <pre className="font-mono text-xs text-slate-300 leading-relaxed overflow-x-auto whitespace-pre p-2 bg-slate-950 rounded-lg">
                    {`git remote add origin ${cloneHttp}
git branch -M main
git push -u origin main`}
                  </pre>
                </div>
              </Card>
            ) : (
              /* File List Card */
              <Card className="p-0 overflow-hidden border-slate-800 bg-slate-950/40">
                {/* Directory Table Header */}
                <div className="flex items-center justify-between px-4 py-2.5 bg-slate-900/80 border-b border-slate-800 text-xs text-slate-400 font-medium">
                  <div className="flex items-center gap-2">
                    <span>branch</span>
                    <span className="font-mono text-sky-400 font-semibold">{currentBranch}</span>
                    <span className="text-slate-500">•</span>
                    <span>{files.length} items</span>
                  </div>
                  <Button variant="ghost" size="xs" onClick={loadTree} className="text-slate-400 hover:text-white p-1">
                    <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
                  </Button>
                </div>

                {/* File Rows */}
                <div className="divide-y divide-slate-850/60">
                  {files.map((file) => (
                    <div
                      key={file.path}
                      onClick={() => handleSelectFile(file)}
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
                        {file.lastCommitMessage || 'Commit update'}
                      </div>
                      <div className="flex items-center gap-2 text-[11px] text-slate-500 shrink-0">
                        <span>{file.lastCommitDate || 'Recently'}</span>
                        <ChevronRight className="w-3.5 h-3.5 text-slate-600 group-hover:text-slate-400 transition-colors" />
                      </div>
                    </div>
                  ))}
                </div>
              </Card>
            )}

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
                  exist.io is an enterprise-grade Git forge engine built in pure Go. It delivers seamless Smart-HTTP Git
                  and SSH transport alongside first-class Model Context Protocol (MCP) server endpoints, automated AI code review,
                  and ephemeral capability leases.
                </p>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                  <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 space-y-1">
                    <h4 className="text-xs font-semibold text-sky-300 flex items-center gap-1.5">
                      <GitBranch className="w-3.5 h-3.5" /> Pure-Go Smart-HTTP & SSH
                    </h4>
                    <p className="text-[11px] text-slate-400">
                      Zero cgo runtime with high-throughput Git transport supporting LFS v1, cryptographic signature verification,
                      and SQLite WAL persistence.
                    </p>
                  </div>
                  <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 space-y-1">
                    <h4 className="text-xs font-semibold text-purple-300 flex items-center gap-1.5">
                      <Shield className="w-3.5 h-3.5" /> Policy & Rulesets
                    </h4>
                    <p className="text-[11px] text-slate-400">
                      Enforce automated agent code reviews, required status checks, linear history, and CIDR IP allowlists.
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
