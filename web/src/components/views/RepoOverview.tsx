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
  Search,
  Code2,
  Terminal,
  RefreshCw,
  Sparkles,
  BookOpen,
  Tag,
  Cpu,
  ChevronDown,
  Plus,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Box, BoxHeader, BoxBody, BoxRow } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { CreateFileModal } from '@/components/modals/CreateFileModal'
import {
  type RepoInfo,
  type FileItem,
  type SymbolItem,
  type ReleaseItem,
  getFiles,
  getFileContent,
  getReleases,
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
 * Dual-Mode:
 * 1. Empty Repository State: Canonical GitHub Quick Setup screen with HTTPS/SSH switcher and command blocks.
 * 2. Populated Repository State: File tree + real dynamic README + dynamic language distribution + metadata sidebar.
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
  const [currentBranch, setCurrentBranch] = useState(repo.defaultBranch || 'main')

  // Readme State
  const [readmeContent, setReadmeContent] = useState<string>('')
  const [loadingReadme, setLoadingReadme] = useState(false)

  // Quick Setup & Clone State
  const [cloneProtocol, setCloneProtocol] = useState<'http' | 'ssh'>('http')
  const [copiedBlock, setCopiedBlock] = useState<string | null>(null)
  const [cloneDropdownOpen, setCloneDropdownOpen] = useState(false)
  const [createFileOpen, setCreateFileOpen] = useState(false)

  // Symbol Search State
  const [searchQuery, setSearchQuery] = useState('')
  const [symbolResults, setSymbolResults] = useState<SymbolItem[]>([])
  const [isSearchingSymbols, setIsSearchingSymbols] = useState(false)

  // Releases Sidebar State
  const [latestRelease, setLatestRelease] = useState<ReleaseItem | null>(null)
  const [releaseCount, setReleaseCount] = useState<number>(0)

  const owner = repo.ownerSlug || 'acme'
  const name = repo.name || 'rocket'

  const cloneHttp = repo.cloneUrlHttp || `http://${window.location.host}/${owner}/${name}.git`
  const cloneSsh = repo.cloneUrlSsh || `git@${window.location.hostname}:2222/${owner}/${name}.git`
  const activeCloneUrl = cloneProtocol === 'http' ? cloneHttp : cloneSsh

  const loadTree = async () => {
    setLoading(true)
    try {
      const entries = await getFiles(owner, name, currentBranch)
      setFiles(entries)

      // Look for README
      const readmeEntry = entries.find((f) => f.name.toLowerCase() === 'readme.md')
      if (readmeEntry) {
        setLoadingReadme(true)
        try {
          const content = await getFileContent(owner, name, readmeEntry.path, currentBranch)
          setReadmeContent(content || (isDemoModeActive() ? sampleReadme : ''))
        } catch {
          setReadmeContent(isDemoModeActive() ? sampleReadme : '')
        } finally {
          setLoadingReadme(false)
        }
      } else {
        setReadmeContent(isDemoModeActive() ? sampleReadme : '')
      }
    } catch {
      if (isDemoModeActive()) {
        setFiles(sampleFiles)
        setReadmeContent(sampleReadme)
      } else {
        setFiles([])
        setReadmeContent('')
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

  useEffect(() => {
    getReleases(owner, name)
      .then((rels) => {
        if (rels && rels.length > 0) {
          setLatestRelease(rels[0])
          setReleaseCount(rels.length)
        } else {
          setLatestRelease(null)
          setReleaseCount(0)
        }
      })
      .catch(() => {
        setLatestRelease(null)
        setReleaseCount(0)
      })
  }, [owner, name])

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

  const handleCopy = (text: string, blockId: string) => {
    navigator.clipboard.writeText(text)
    setCopiedBlock(blockId)
    setTimeout(() => setCopiedBlock(null), 2000)
  }

  const handleCopyCode = (code: string) => {
    navigator.clipboard.writeText(code)
    setCopiedCode(true)
    setTimeout(() => setCopiedCode(false), 2000)
  }

  // Compute dynamic language distribution from file extensions
  const calculateLanguages = () => {
    if (files.length === 0) return []
    const counts: Record<string, { count: number; color: string }> = {}
    let total = 0

    files.forEach((f) => {
      if (f.type !== 'blob') return
      const ext = f.name.split('.').pop()?.toLowerCase() || ''
      let lang = 'Other'
      let color = '#8b949e'

      if (ext === 'go') {
        lang = 'Go'
        color = '#00ADD8'
      } else if (ext === 'ts' || ext === 'tsx') {
        lang = 'TypeScript'
        color = '#3178C6'
      } else if (ext === 'js' || ext === 'jsx') {
        lang = 'JavaScript'
        color = '#F7DF1E'
      } else if (ext === 'py') {
        lang = 'Python'
        color = '#3572A5'
      } else if (ext === 'rs') {
        lang = 'Rust'
        color = '#DEA584'
      } else if (ext === 'md') {
        lang = 'Markdown'
        color = '#083FA1'
      } else if (ext === 'json' || ext === 'yaml' || ext === 'yml') {
        lang = 'Config'
        color = '#E34C26'
      }

      counts[lang] = counts[lang] || { count: 0, color }
      counts[lang].count += 1
      total += 1
    })

    if (total === 0) return []
    return Object.entries(counts).map(([name, data]) => ({
      name,
      color: data.color,
      percent: Math.round((data.count / total) * 100),
    }))
  }

  const languages = calculateLanguages()
  const hasReadmeFile = files.some((f) => f.name.toLowerCase() === 'readme.md')

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
              {isDemoModeActive() && (
                <>
                  <option value="master">master</option>
                  <option value="feature/autonomous-agent">feature/autonomous-agent</option>
                </>
              )}
            </select>
            <GitBranch className="w-3.5 h-3.5 text-muted-foreground absolute left-2.5 pointer-events-none" />
            <ChevronDown className="w-3 h-3 text-muted-foreground absolute right-2.5 pointer-events-none" />
          </div>

          {/* Stats Badges */}
          <div className="flex items-center gap-3 text-xs text-muted-foreground font-medium">
            <span className="flex items-center gap-1 hover:text-foreground cursor-pointer">
              <GitBranch className="w-3.5 h-3.5" />
              <strong>{files.length > 0 ? 1 : 0}</strong> branch
            </span>
            <span>•</span>
            <span className="flex items-center gap-1 hover:text-foreground cursor-pointer">
              <Tag className="w-3.5 h-3.5" />
              <strong>{isDemoModeActive() ? 2 : 0}</strong> tags
            </span>
          </div>
        </div>

        {/* Top Right Action Tools */}
        <div className="flex items-center gap-2">
          {/* Add file button */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCreateFileOpen(true)}
            className="gap-1.5 text-xs font-semibold"
          >
            <Plus className="w-3.5 h-3.5 text-primary" />
            <span>Add file</span>
          </Button>

          {/* Code / Clone Dropdown */}
          <div className="relative">
            <Button
              variant="primary"
              size="sm"
              onClick={() => setCloneDropdownOpen(!cloneDropdownOpen)}
              className="gap-1.5 text-xs font-semibold"
            >
              <Code2 className="w-3.5 h-3.5" />
              <span>Code</span>
              <ChevronDown className="w-3 h-3" />
            </Button>

            {cloneDropdownOpen && (
              <div
                className="absolute right-0 mt-1.5 w-80 bg-card border border-border rounded-md shadow-xl p-3 z-30 space-y-3"
                onClick={(e) => e.stopPropagation()}
              >
                <div className="flex items-center justify-between text-xs font-semibold text-foreground border-b border-border pb-2">
                  <span>Clone</span>
                  <div className="flex items-center gap-1 bg-secondary rounded p-0.5">
                    <button
                      onClick={() => setCloneProtocol('http')}
                      className={`px-2 py-0.5 rounded text-[11px] font-semibold cursor-pointer transition-colors ${
                        cloneProtocol === 'http'
                          ? 'bg-primary text-primary-foreground'
                          : 'text-muted-foreground hover:text-foreground'
                      }`}
                    >
                      HTTPS
                    </button>
                    <button
                      onClick={() => setCloneProtocol('ssh')}
                      className={`px-2 py-0.5 rounded text-[11px] font-semibold cursor-pointer transition-colors ${
                        cloneProtocol === 'ssh'
                          ? 'bg-primary text-primary-foreground'
                          : 'text-muted-foreground hover:text-foreground'
                      }`}
                    >
                      SSH
                    </button>
                  </div>
                </div>

                <div className="flex items-center rounded-md border border-border bg-input px-2 py-1 gap-2">
                  <input
                    type="text"
                    readOnly
                    value={activeCloneUrl}
                    className="w-full bg-transparent border-none outline-none text-[11px] font-mono text-foreground select-all"
                  />
                  <button
                    onClick={() => handleCopy(activeCloneUrl, 'dropdown-clone')}
                    className="text-muted-foreground hover:text-foreground p-1 cursor-pointer shrink-0"
                    title="Copy URL"
                  >
                    {copiedBlock === 'dropdown-clone' ? (
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                    ) : (
                      <Copy className="w-3.5 h-3.5" />
                    )}
                  </button>
                </div>

                <div className="text-[11px] text-muted-foreground">
                  Use Git or checkout with SVN using the web URL.
                </div>
              </div>
            )}
          </div>

          {/* Semantic Symbol Search Input */}
          <div className="relative w-full sm:w-64">
            <div className="flex items-center gap-2 px-2.5 py-1 h-8 rounded-md bg-input border border-border text-xs focus-within:border-primary focus-within:ring-1 focus-within:ring-ring shadow-2xs">
              <Search className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search symbols..."
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
      </div>

      {/* 2. Main Content Body: File Viewer OR Empty Quick Setup OR 2-Column Overview */}
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
      ) : files.length === 0 && !loading ? (
        /* Canonical GitHub Quick Setup Screen (When Repo is Brand New with 0 Commits) */
        <div className="space-y-6">
          <Box className="overflow-hidden border border-border">
            {/* Quick Setup Blue/Secondary Banner */}
            <div className="p-4 sm:p-5 bg-secondary/50 border-b border-border space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="space-y-1">
                  <h3 className="text-sm font-semibold text-foreground">
                    Quick setup — if you've done this kind of thing before
                  </h3>
                  <p className="text-xs text-muted-foreground">
                    Get started by{' '}
                    <button
                      onClick={() => setCreateFileOpen(true)}
                      className="text-primary hover:underline font-medium cursor-pointer"
                    >
                      creating a new file
                    </button>{' '}
                    or pushing an existing repository.
                  </p>
                </div>

                {/* HTTPS / SSH switcher button */}
                <div className="flex items-center gap-1.5 bg-input border border-border rounded-md p-1">
                  <button
                    onClick={() => setCloneProtocol('http')}
                    className={`px-2.5 py-1 rounded text-xs font-semibold cursor-pointer transition-colors ${
                      cloneProtocol === 'http'
                        ? 'bg-primary text-primary-foreground shadow-2xs'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    HTTPS
                  </button>
                  <button
                    onClick={() => setCloneProtocol('ssh')}
                    className={`px-2.5 py-1 rounded text-xs font-semibold cursor-pointer transition-colors ${
                      cloneProtocol === 'ssh'
                        ? 'bg-primary text-primary-foreground shadow-2xs'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    SSH
                  </button>

                  <div className="flex items-center gap-1 pl-2">
                    <span className="font-mono text-xs text-foreground select-all px-1">
                      {activeCloneUrl}
                    </span>
                    <button
                      onClick={() => handleCopy(activeCloneUrl, 'quick-clone')}
                      className="p-1 rounded text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer"
                      title="Copy URL"
                    >
                      {copiedBlock === 'quick-clone' ? (
                        <Check className="w-3.5 h-3.5 text-emerald-400" />
                      ) : (
                        <Copy className="w-3.5 h-3.5" />
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            {/* Terminal Blocks for Git CLI Setup */}
            <div className="p-5 space-y-6 text-xs bg-card">
              {/* Block 1: Create a new repository on the command line */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="font-semibold text-foreground">
                    …or create a new repository on the command line
                  </h4>
                  <button
                    onClick={() =>
                      handleCopy(
                        `echo "# ${name}" >> README.md\ngit init\ngit add README.md\ngit commit -m "first commit"\ngit branch -M main\ngit remote add origin ${activeCloneUrl}\ngit push -u origin main`,
                        'block-create'
                      )
                    }
                    className="flex items-center gap-1 text-[11px] text-primary hover:underline cursor-pointer"
                  >
                    {copiedBlock === 'block-create' ? (
                      <>
                        <Check className="w-3 h-3 text-emerald-400" />
                        <span>Copied</span>
                      </>
                    ) : (
                      <>
                        <Copy className="w-3 h-3" />
                        <span>Copy commands</span>
                      </>
                    )}
                  </button>
                </div>
                <pre className="p-3 rounded-md bg-input border border-border font-mono text-[11px] text-foreground leading-relaxed overflow-x-auto whitespace-pre">
{`echo "# ${name}" >> README.md
git init
git add README.md
git commit -m "first commit"
git branch -M main
git remote add origin ${activeCloneUrl}
git push -u origin main`}
                </pre>
              </div>

              {/* Block 2: Push an existing repository from the command line */}
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="font-semibold text-foreground">
                    …or push an existing repository from the command line
                  </h4>
                  <button
                    onClick={() =>
                      handleCopy(
                        `git remote add origin ${activeCloneUrl}\ngit branch -M main\ngit push -u origin main`,
                        'block-push'
                      )
                    }
                    className="flex items-center gap-1 text-[11px] text-primary hover:underline cursor-pointer"
                  >
                    {copiedBlock === 'block-push' ? (
                      <>
                        <Check className="w-3 h-3 text-emerald-400" />
                        <span>Copied</span>
                      </>
                    ) : (
                      <>
                        <Copy className="w-3 h-3" />
                        <span>Copy commands</span>
                      </>
                    )}
                  </button>
                </div>
                <pre className="p-3 rounded-md bg-input border border-border font-mono text-[11px] text-foreground leading-relaxed overflow-x-auto whitespace-pre">
{`git remote add origin ${activeCloneUrl}
git branch -M main
git push -u origin main`}
                </pre>
              </div>

              {/* Block 3: Autonomous AI Agent Initialization */}
              <div className="space-y-2 pt-2 border-t border-border">
                <div className="flex items-center justify-between">
                  <h4 className="font-semibold text-foreground flex items-center gap-1.5">
                    <Sparkles className="w-3.5 h-3.5 text-primary" />
                    <span>…or initialize with an autonomous AI coding agent</span>
                  </h4>
                  <button
                    onClick={() =>
                      handleCopy(
                        `exist agent lease --repo ${owner}/${name} --role full-forge --ttl 1h`,
                        'block-agent'
                      )
                    }
                    className="flex items-center gap-1 text-[11px] text-primary hover:underline cursor-pointer"
                  >
                    {copiedBlock === 'block-agent' ? (
                      <>
                        <Check className="w-3 h-3 text-emerald-400" />
                        <span>Copied</span>
                      </>
                    ) : (
                      <>
                        <Copy className="w-3 h-3" />
                        <span>Copy command</span>
                      </>
                    )}
                  </button>
                </div>
                <pre className="p-3 rounded-md bg-input border border-border font-mono text-[11px] text-foreground leading-relaxed overflow-x-auto whitespace-pre">
{`exist agent lease --repo ${owner}/${name} --role full-forge --ttl 1h`}
                </pre>
              </div>
            </div>
          </Box>
        </div>
      ) : (
        /* Standard GitHub Two-Column Overview Layout (When files exist) */
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Left Column (9 cols / 75%): File Tree + Readme */}
          <div className="lg:col-span-8 xl:col-span-9 space-y-5">
            {/* File Tree Box (Primer Table) */}
            <Box>
              {/* Latest Commit Table Header */}
              <BoxHeader className="py-2">
                <div className="flex items-center gap-2 min-w-0">
                  <div className="w-5 h-5 rounded-full bg-primary text-primary-foreground text-[10px] font-bold flex items-center justify-center shrink-0">
                    {owner.slice(0, 1).toLowerCase()}
                  </div>
                  <span className="font-semibold text-foreground truncate">
                    {files[0]?.lastCommitMessage || 'Initial commit'}
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
                      {file.lastCommitMessage || 'Commit update'}
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

            {/* Readme Box OR Add a Readme Prompt */}
            {hasReadmeFile ? (
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
                  {loadingReadme ? (
                    <div className="p-8 text-center text-muted-foreground text-xs font-mono flex items-center justify-center gap-2">
                      <RefreshCw className="w-3.5 h-3.5 animate-spin text-primary" />
                      <span>Loading README...</span>
                    </div>
                  ) : readmeContent ? (
                    <div className="prose prose-invert max-w-none text-xs sm:text-sm text-foreground/90 font-sans leading-relaxed whitespace-pre-wrap">
                      {readmeContent}
                    </div>
                  ) : (
                    <div className="space-y-3">
                      <h2 className="text-xl font-bold text-foreground tracking-tight flex items-center gap-2 border-b border-border pb-2">
                        <span>{repo.name}</span>
                      </h2>
                      <p className="text-foreground/90 leading-relaxed text-xs sm:text-sm">
                        {repo.description || 'Enterprise repository on exist.io.'}
                      </p>
                    </div>
                  )}
                </BoxBody>
              </Box>
            ) : (
              <div className="p-4 rounded-md border border-dashed border-border bg-card/60 flex items-center justify-between gap-4">
                <div className="space-y-1">
                  <h4 className="text-xs font-semibold text-foreground">Add a README</h4>
                  <p className="text-[11px] text-muted-foreground">
                    Help people interested in this repository understand your project by adding a README.
                  </p>
                </div>
                <Button
                  variant="primary"
                  size="xs"
                  onClick={() => setCreateFileOpen(true)}
                  className="shrink-0 gap-1"
                >
                  <Plus className="w-3 h-3" />
                  <span>Add README</span>
                </Button>
              </div>
            )}
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
                {['git-forge', 'pure-go', 'mcp-server', 'autonomous-agents'].map((tag) => (
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
                <span className="text-[11px] text-muted-foreground font-mono">
                  {releaseCount}
                </span>
              </div>
              {latestRelease ? (
                <div className="flex items-center gap-2">
                  <Tag className="w-4 h-4 text-emerald-400" />
                  <div>
                    <span className="font-semibold text-foreground">{latestRelease.tagName}</span>
                    <span className="ml-1.5 text-[10px] px-1.5 py-0.2 rounded-full bg-emerald-500/15 text-emerald-400 font-semibold border border-emerald-500/30">
                      Latest
                    </span>
                  </div>
                </div>
              ) : (
                <div className="text-[11px] text-muted-foreground">
                  No releases published.
                </div>
              )}
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
              {languages.length > 0 ? (
                <>
                  {/* Progress Bar */}
                  <div className="h-2 w-full rounded-full bg-secondary overflow-hidden flex">
                    {languages.map((l) => (
                      <div
                        key={l.name}
                        className="h-full"
                        style={{ width: `${l.percent}%`, backgroundColor: l.color }}
                        title={`${l.name}: ${l.percent}%`}
                      />
                    ))}
                  </div>
                  {/* Legend */}
                  <div className="flex flex-wrap items-center gap-3 text-[11px] text-muted-foreground pt-1">
                    {languages.map((l) => (
                      <span key={l.name} className="flex items-center gap-1.5">
                        <span
                          className="w-2 h-2 rounded-full"
                          style={{ backgroundColor: l.color }}
                        />
                        <strong className="text-foreground">{l.name}</strong> {l.percent}%
                      </span>
                    ))}
                  </div>
                </>
              ) : (
                <div className="text-[11px] text-muted-foreground">
                  No language data available.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* In-Browser Create File Modal */}
      <CreateFileModal
        isOpen={createFileOpen}
        onClose={() => setCreateFileOpen(false)}
        owner={owner}
        name={name}
        branch={currentBranch}
        initialPath="README.md"
        onCommitSuccess={() => {
          loadTree()
        }}
      />
    </div>
  )
}
