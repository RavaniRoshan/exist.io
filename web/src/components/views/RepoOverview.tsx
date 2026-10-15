import React, { useState, useEffect } from 'react'
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
  BookOpen,
  Tag,
  Cpu,
  ChevronDown,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
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

/**
 * GitHub Primer Repository Overview Screen.
 * Branch selector + Commit summary table header + File tree + Right-hand metadata sidebar + README Box.
 */
export const RepoOverview: React.FC<RepoOverviewProps> = ({
  repo,
  onSelectCommits,
  onSelectAgentConsole,
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
    }, 200)

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
    <div className="space-y-4">
      {/* 1. Branch Selector & Top Controls Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {/* Branch Picker Dropdown Button */}
          <div className="relative inline-flex items-center">
            <select
              value={currentBranch}
              onChange={(e) => setCurrentBranch(e.target.value)}
              aria-label="Switch Git branch"
              className="h-8 pl-8 pr-7 bg-secondary text-foreground text-xs font-semibold rounded-md border border-border appearance-none cursor-pointer hover:bg-secondary/80 focus:border-primary outline-none shadow-2xs"
            >
              <option value="main">main</option>
              <option value="master">master</option>
              <option value="feature/autonomous-agent">feature/autonomous-agent</option>
            </select>
            <GitBranch className="w-3.5 h-3.5 text-muted-foreground absolute left-2.5 pointer-events-none" />
            <ChevronDown className="w-3 h-3 text-muted-foreground absolute right-2.5 pointer-events-none" />
          </div>

          {/* Stats Badges */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground font-medium">
            <span className="flex items-center gap-1 hover:text-foreground cursor-pointer">
              <GitBranch className="w-3.5 h-3.5" />
              <strong>1</strong> branch
            </span>
            <span>•</span>
            <span className="flex items-center gap-1 hover:text-foreground cursor-pointer">
              <Tag className="w-3.5 h-3.5" />
              <strong>2</strong> tags
            </span>
          </div>
        </div>

        {/* Semantic Symbol Search Input */}
        <div className="relative w-full sm:w-80">
          <div className="flex items-center gap-2 px-2.5 py-1 h-8 rounded-md bg-input border border-border text-xs focus-within:border-primary focus-within:ring-1 focus-within:ring-ring shadow-2xs">
            <Search className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search symbols (Go AST, TS, Python)..."
              className="w-full bg-transparent border-none outline-none text-foreground placeholder:text-muted-foreground text-xs font-mono"
            />
            {isSearchingSymbols && <RefreshCw className="w-3 h-3 animate-spin text-primary shrink-0" />}
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="text-[10px] text-muted-foreground hover:text-foreground bg-secondary px-1 py-0.2 rounded"
              >
                Clear
              </button>
            )}
          </div>

          {/* Semantic Symbol Search Results Popover */}
          {searchQuery && (
            <div className="absolute top-full left-0 right-0 mt-1.5 z-30 bg-card border border-border rounded-md shadow-xl overflow-hidden divide-y divide-border max-h-72 overflow-y-auto">
              <div className="px-3 py-1.5 bg-secondary/70 text-[11px] font-semibold text-muted-foreground flex items-center justify-between">
                <span className="flex items-center gap-1.5">
                  <Sparkles className="w-3 h-3 text-primary" />
                  <span>AST Symbol Search</span>
                </span>
                <span>{symbolResults.length} matches</span>
              </div>

              {symbolResults.length === 0 ? (
                <div className="p-3 text-center text-xs text-muted-foreground font-mono">
                  {isSearchingSymbols ? 'Parsing AST symbols...' : 'No symbols found'}
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
                    className="p-2.5 hover:bg-secondary/40 transition-colors cursor-pointer flex items-center justify-between gap-3 text-xs"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Code2 className="w-3.5 h-3.5 text-primary shrink-0" />
                      <div className="min-w-0">
                        <div className="flex items-center gap-1.5">
                          <span className="font-mono font-semibold text-foreground text-[11px]">{s.name}</span>
                          <span className="text-[9px] px-1 py-0.2 rounded bg-secondary text-muted-foreground font-mono">
                            {s.kind}
                          </span>
                        </div>
                        {s.signature && (
                          <div className="text-[10px] font-mono text-muted-foreground truncate">{s.signature}</div>
                        )}
                      </div>
                    </div>
                    <span className="font-mono text-[10px] text-muted-foreground shrink-0">
                      {s.file}:{s.line}
                    </span>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </div>

      {/* 2. Main Content Body: File Viewer OR 2-Column Overview */}
      {selectedFile ? (
        /* File Blob Viewer */
        <Box>
          <BoxHeader>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="xs"
                onClick={() => setSelectedFile(null)}
                className="p-1 h-6 w-6"
              >
                <ArrowLeft className="w-3.5 h-3.5" />
              </Button>
              <FileCode className="w-4 h-4 text-primary" />
              <span className="font-mono text-xs font-semibold text-foreground">{selectedFile.path}</span>
              {selectedFile.size !== undefined && selectedFile.size > 0 && (
                <span className="text-[11px] text-muted-foreground font-mono">({selectedFile.size} bytes)</span>
              )}
            </div>
            <Button
              variant="secondary"
              size="xs"
              onClick={() => handleCopyCode(fileContent)}
              className="gap-1 text-muted-foreground hover:text-foreground"
            >
              {copiedCode ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
              <span>{copiedCode ? 'Copied' : 'Raw copy'}</span>
            </Button>
          </BoxHeader>

          {/* Line Numbers and Code Canvas */}
          <div className="p-3 overflow-x-auto font-mono text-xs leading-relaxed text-foreground bg-input max-h-[600px] overflow-y-auto">
            {loadingContent ? (
              <div className="p-8 text-center text-muted-foreground text-xs font-mono flex items-center justify-center gap-2">
                <RefreshCw className="w-3.5 h-3.5 animate-spin text-primary" />
                <span>Loading git blob...</span>
              </div>
            ) : (
              <pre className="m-0">
                {fileContent.split('\n').map((line, idx) => (
                  <div key={idx} className="flex hover:bg-secondary/30 px-1 rounded">
                    <span className="w-10 select-none text-right pr-4 text-muted-foreground/60 font-mono text-[11px] shrink-0">
                      {idx + 1}
                    </span>
                    <span className="flex-1 whitespace-pre">{line}</span>
                  </div>
                ))}
              </pre>
            )}
          </div>
        </Box>
      ) : (
        /* Standard GitHub Two-Column Overview Layout */
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Left Column (9 cols / 75%): File Tree + Readme */}
          <div className="lg:col-span-8 xl:col-span-9 space-y-5">
            {files.length === 0 && !loading ? (
              /* Authentic Empty State Box */
              <Box className="p-6 text-center space-y-4">
                <div className="w-10 h-10 rounded-full bg-secondary flex items-center justify-center mx-auto text-muted-foreground">
                  <Folder className="w-5 h-5" />
                </div>
                <div className="max-w-md mx-auto space-y-1">
                  <h3 className="text-sm font-semibold text-foreground">This repository is empty</h3>
                  <p className="text-xs text-muted-foreground leading-normal">
                    Get started by pushing files from your terminal or issuing an autonomous agent lease.
                  </p>
                </div>

                <div className="max-w-md mx-auto text-left bg-input border border-border rounded-md p-3 space-y-2">
                  <div className="flex items-center justify-between text-[11px] text-muted-foreground font-mono">
                    <span className="flex items-center gap-1.5">
                      <Terminal className="w-3 h-3 text-primary" />
                      <span>Push an existing repository from command line</span>
                    </span>
                    <button
                      onClick={handleCopyClone}
                      className="hover:text-foreground text-primary cursor-pointer"
                    >
                      {copiedClone ? 'Copied' : 'Copy'}
                    </button>
                  </div>
                  <pre className="font-mono text-[11px] text-foreground overflow-x-auto whitespace-pre p-2 bg-secondary/30 rounded">
                    {`git remote add origin ${cloneHttp}
git branch -M main
git push -u origin main`}
                  </pre>
                </div>
              </Box>
            ) : (
              /* File Tree Box (Primer Table) */
              <Box>
                {/* Latest Commit Table Header */}
                <BoxHeader className="py-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <div className="w-5 h-5 rounded-full bg-primary text-primary-foreground text-[10px] font-bold flex items-center justify-center shrink-0">
                      e
                    </div>
                    <span className="font-semibold text-foreground truncate">
                      {files[0]?.lastCommitMessage || 'feat(core): initialize forge workspace'}
                    </span>
                  </div>

                  <div className="flex items-center gap-3 text-xs text-muted-foreground shrink-0 font-mono">
                    <button
                      onClick={onSelectCommits}
                      className="hover:text-primary transition-colors flex items-center gap-1 cursor-pointer"
                      title="View full commit history"
                    >
                      <Clock className="w-3.5 h-3.5" />
                      <span className="hover:underline">History</span>
                    </button>
                    <span>•</span>
                    <Button variant="ghost" size="xs" onClick={loadTree} className="p-1 h-6 w-6">
                      <RefreshCw className={`w-3 h-3 ${loading ? 'animate-spin' : ''}`} />
                    </Button>
                  </div>
                </BoxHeader>

                {/* File Rows */}
                <div>
                  {files.map((file) => (
                    <BoxRow
                      key={file.path}
                      onClick={() => handleSelectFile(file)}
                      className="cursor-pointer group"
                    >
                      {/* Name & Icon */}
                      <div className="flex items-center gap-2.5 min-w-0 flex-1 pr-4">
                        {file.type === 'tree' ? (
                          <Folder className="w-4 h-4 text-primary shrink-0" />
                        ) : (
                          <FileText className="w-4 h-4 text-muted-foreground shrink-0 group-hover:text-foreground" />
                        )}
                        <span className="font-mono text-foreground group-hover:text-primary group-hover:underline truncate font-medium">
                          {file.name}
                        </span>
                      </div>

                      {/* Last Commit Message */}
                      <div className="hidden sm:block text-muted-foreground text-xs truncate max-w-xs md:max-w-md flex-1 text-left px-2">
                        {file.lastCommitMessage || 'Update file contents'}
                      </div>

                      {/* Age / Time ago */}
                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground font-mono shrink-0">
                        <span>{file.lastCommitDate || 'recently'}</span>
                        <ChevronRight className="w-3.5 h-3.5 text-muted-foreground/50 group-hover:text-foreground" />
                      </div>
                    </BoxRow>
                  ))}
                </div>
              </Box>
            )}

            {/* Readme Box */}
            <Box>
              <BoxHeader>
                <div className="flex items-center gap-2 text-xs font-semibold text-foreground">
                  <BookOpen className="w-4 h-4 text-primary" />
                  <span>README.md</span>
                </div>
                <Badge variant="outline" className="text-[10px] py-0 border-border">
                  Markdown
                </Badge>
              </BoxHeader>

              <BoxBody className="p-6 space-y-4">
                <div className="space-y-3">
                  <h2 className="text-xl font-bold text-foreground tracking-tight flex items-center gap-2 border-b border-border pb-2">
                    <span>{repo.name}</span>
                    <span className="text-xs font-normal text-muted-foreground bg-secondary px-2 py-0.5 rounded-full border border-border">
                      The Autonomous Agent Git Forge
                    </span>
                  </h2>
                  <p className="text-foreground/90 leading-relaxed text-xs sm:text-sm">
                    {repo.description ||
                      'exist.io is an enterprise-grade Git forge engine built in pure Go. It delivers seamless Smart-HTTP Git and SSH transport alongside first-class Model Context Protocol (MCP) server endpoints, automated AI code review, and ephemeral capability leases.'}
                  </p>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                    <div className="p-3 rounded-md bg-secondary/30 border border-border space-y-1">
                      <h4 className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                        <GitBranch className="w-3.5 h-3.5 text-primary" /> Pure-Go Smart-HTTP & SSH
                      </h4>
                      <p className="text-[11px] text-muted-foreground leading-normal">
                        Zero cgo runtime with high-throughput Git transport supporting LFS v1, cryptographic signature verification,
                        and SQLite WAL persistence.
                      </p>
                    </div>

                    <div className="p-3 rounded-md bg-secondary/30 border border-border space-y-1">
                      <h4 className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                        <Shield className="w-3.5 h-3.5 text-primary" /> Policy & Rulesets
                      </h4>
                      <p className="text-[11px] text-muted-foreground leading-normal">
                        Enforce automated agent code reviews, required status checks, linear history, and CIDR IP allowlists.
                      </p>
                    </div>
                  </div>
                </div>
              </BoxBody>
            </Box>
          </div>

          {/* Right Column (4 cols / 25%): GitHub Repo Metadata Sidebar */}
          <div className="lg:col-span-4 xl:col-span-3 space-y-5 text-xs">
            {/* About Section */}
            <div className="space-y-2.5 border-b border-border pb-5">
              <h3 className="font-semibold text-sm text-foreground">About</h3>
              <p className="text-muted-foreground leading-normal">
                {repo.description || 'Enterprise self-hosted Git forge built in pure Go for humans and autonomous coding agents.'}
              </p>

              {/* Topic Tags */}
              <div className="flex flex-wrap gap-1.5 pt-1">
                {['git-forge', 'pure-go', 'mcp-server', 'autonomous-agents', 'iac'].map((tag) => (
                  <span
                    key={tag}
                    className="px-2 py-0.5 rounded-full bg-secondary text-foreground text-[11px] font-mono border border-border hover:border-primary cursor-pointer transition-colors"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>

            {/* Releases Section */}
            <div className="space-y-2 border-b border-border pb-5">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-sm text-foreground flex items-center gap-1.5">
                  <Tag className="w-3.5 h-3.5 text-primary" />
                  <span>Releases</span>
                </h3>
                <span className="text-[11px] text-muted-foreground font-mono">1</span>
              </div>
              <div className="flex items-center gap-2">
                <Tag className="w-4 h-4 text-emerald-400" />
                <div>
                  <a
                    href="https://github.com/RavaniRoshan/exist.io/releases/tag/v0.3.0"
                    target="_blank"
                    rel="noreferrer"
                    className="font-semibold text-foreground hover:text-primary hover:underline"
                  >
                    v0.3.0
                  </a>
                  <span className="ml-1.5 text-[10px] px-1.5 py-0.2 rounded-full bg-emerald-500/15 text-emerald-400 font-semibold border border-emerald-500/30">
                    Latest
                  </span>
                </div>
              </div>
            </div>

            {/* MCP Agent Capabilities Section */}
            <div className="space-y-2 border-b border-border pb-5">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-sm text-foreground flex items-center gap-1.5">
                  <Cpu className="w-3.5 h-3.5 text-primary" />
                  <span>Autonomous Agents</span>
                </h3>
              </div>
              <div className="space-y-1.5 text-[11px] text-muted-foreground">
                <div className="flex items-center justify-between">
                  <span>Mounted MCP tools:</span>
                  <strong className="text-foreground font-mono">35 tools</strong>
                </div>
                <div className="flex items-center justify-between">
                  <span>Capability leases:</span>
                  <span className="text-emerald-400 font-semibold">Enabled</span>
                </div>
                <div className="flex items-center justify-between">
                  <span>AST Symbol Search:</span>
                  <span className="text-emerald-400 font-semibold">Active</span>
                </div>
              </div>
              <Button
                variant="secondary"
                size="xs"
                onClick={onSelectAgentConsole}
                className="w-full mt-2 gap-1.5"
              >
                <Terminal className="w-3 h-3 text-primary" />
                <span>Open Agent Console</span>
              </Button>
            </div>

            {/* Languages Section */}
            <div className="space-y-2">
              <h3 className="font-semibold text-sm text-foreground">Languages</h3>
              {/* Progress Bar */}
              <div className="h-2 w-full rounded-full bg-secondary overflow-hidden flex">
                <div className="bg-primary h-full" style={{ width: '92%' }} title="Go: 92%" />
                <div className="bg-amber-400 h-full" style={{ width: '8%' }} title="TypeScript: 8%" />
              </div>
              {/* Legend */}
              <div className="flex items-center gap-3 text-[11px] text-muted-foreground pt-1">
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-primary" />
                  <strong className="text-foreground">Go</strong> 92.0%
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-amber-400" />
                  <strong className="text-foreground">TypeScript</strong> 8.0%
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
