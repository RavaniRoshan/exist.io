import React, { useState, useEffect } from 'react'
import {
  Tag,
  Package,
  FileArchive,
  Download,
  Copy,
  Check,
  GitCommit,
  Plus,
  ShieldCheck,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Box, BoxHeader, BoxTitle, BoxBody, BoxRow } from '@/components/ui/Card'
import { Dialog } from '@/components/ui/Dialog'
import { getReleases, type RepoInfo, type ReleaseItem } from '@/lib/api'
import { cn } from '@/lib/utils'

export interface ReleasesViewProps {
  repo?: RepoInfo
}

export const ReleasesView: React.FC<ReleasesViewProps> = ({ repo }) => {
  const owner = repo?.ownerSlug || 'shiva'
  const repoName = repo?.name || 'exist.io'

  const [activeSubTab, setActiveSubTab] = useState<'releases' | 'tags'>('releases')
  const [releases, setReleases] = useState<ReleaseItem[]>([])
  const [loading, setLoading] = useState(true)
  const [copiedSha, setCopiedSha] = useState<string | null>(null)

  // Draft Release Modal
  const [isDraftModalOpen, setIsDraftModalOpen] = useState(false)
  const [draftTag, setDraftTag] = useState('')
  const [draftTitle, setDraftTitle] = useState('')
  const [draftBody, setDraftBody] = useState('')
  const [isLatest, setIsLatest] = useState(true)
  const [isPrerelease, setIsPrerelease] = useState(false)
  const [isPublishing, setIsPublishing] = useState(false)

  useEffect(() => {
    loadReleases()
  }, [owner, repoName])

  const loadReleases = async () => {
    setLoading(true)
    const data = await getReleases(owner, repoName)
    setReleases(data)
    setLoading(false)
  }

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSha(id)
    setTimeout(() => setCopiedSha(null), 2000)
  }

  const handlePublishRelease = (e: React.FormEvent) => {
    e.preventDefault()
    if (!draftTag.trim() || !draftTitle.trim()) return

    setIsPublishing(true)
    const newRel: ReleaseItem = {
      id: `rel_${Date.now()}`,
      tagName: draftTag.trim(),
      name: draftTitle.trim(),
      body: draftBody.trim(),
      author: 'shiva',
      isDraft: false,
      isPrerelease,
      createdAt: new Date().toISOString(),
      publishedAt: new Date().toISOString(),
      commitSha: '5c92f55',
      assets: [
        { name: `existd-${draftTag.trim()}-linux-amd64.tar.gz`, size: 34500000, downloadCount: 0 },
        { name: `existd-${draftTag.trim()}-darwin-arm64.tar.gz`, size: 33100000, downloadCount: 0 },
      ],
    }

    setReleases((prev) => [newRel, ...prev])
    setIsPublishing(false)
    setIsDraftModalOpen(false)
    setDraftTag('')
    setDraftTitle('')
    setDraftBody('')
  }

  const formatSize = (bytes: number) => {
    if (bytes >= 1000000) return `${(bytes / 1000000).toFixed(1)} MB`
    if (bytes >= 1000) return `${(bytes / 1000).toFixed(1)} KB`
    return `${bytes} B`
  }

  return (
    <div className="space-y-6">
      {/* Top Header & Sub-Navigation */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        {/* Toggle Pills: Releases vs Tags */}
        <div className="flex items-center rounded-md bg-secondary/50 p-1 border border-border text-xs font-semibold w-fit">
          <button
            onClick={() => setActiveSubTab('releases')}
            className={cn(
              'px-3 py-1 rounded transition-colors cursor-pointer',
              activeSubTab === 'releases' ? 'bg-card text-foreground shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            Releases ({releases.length})
          </button>
          <button
            onClick={() => setActiveSubTab('tags')}
            className={cn(
              'px-3 py-1 rounded transition-colors cursor-pointer',
              activeSubTab === 'tags' ? 'bg-card text-foreground shadow-2xs' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            Tags ({releases.length})
          </button>
        </div>

        <Button
          onClick={() => setIsDraftModalOpen(true)}
          size="sm"
          className="gap-1.5 shrink-0"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Draft a new release</span>
        </Button>
      </div>

      {/* 1. RELEASES TAB */}
      {activeSubTab === 'releases' && (
        <div className="space-y-8">
          {loading ? (
            <div className="py-12 text-center text-xs text-muted-foreground">
              Loading releases...
            </div>
          ) : (
            releases.map((rel, idx) => (
              <div
                key={rel.id}
                className="grid grid-cols-1 md:grid-cols-12 gap-6 items-start pb-8 border-b border-border last:border-b-0"
              >
                {/* Left Meta Column (3 cols) */}
                <div className="md:col-span-3 space-y-2 text-xs">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-mono font-bold text-sm text-foreground">
                      {rel.tagName}
                    </span>
                    {idx === 0 && (
                      <Badge variant="outline" className="border-emerald-500/40 text-emerald-400 bg-emerald-500/10 text-[10px] font-semibold">
                        Latest
                      </Badge>
                    )}
                    {rel.isPrerelease && (
                      <Badge variant="outline" className="border-amber-500/40 text-amber-400 bg-amber-500/10 text-[10px]">
                        Pre-release
                      </Badge>
                    )}
                  </div>

                  <div className="flex items-center gap-1.5 text-muted-foreground font-mono">
                    <GitCommit className="w-3.5 h-3.5" />
                    <span className="hover:text-primary cursor-pointer hover:underline">
                      {rel.commitSha}
                    </span>
                  </div>

                  <div className="text-muted-foreground text-[11px] leading-relaxed">
                    Released by <strong className="text-foreground">{rel.author}</strong> on {new Date(rel.publishedAt).toLocaleDateString()}
                  </div>
                </div>

                {/* Right Content Column (9 cols) */}
                <div className="md:col-span-9 space-y-4">
                  <h2 className="text-lg font-bold text-foreground tracking-tight">
                    {rel.name}
                  </h2>

                  {/* Markdown Body */}
                  <div className="text-xs sm:text-sm text-foreground/90 whitespace-pre-wrap leading-relaxed bg-card rounded-md border border-border p-4 shadow-2xs font-sans">
                    {rel.body}
                  </div>

                  {/* Assets Box */}
                  <Box flush>
                    <BoxHeader className="py-2.5 bg-secondary/30 flex items-center justify-between">
                      <BoxTitle className="text-xs flex items-center gap-1.5">
                        <Package className="w-4 h-4 text-primary" />
                        <span>Assets ({rel.assets.length + 2})</span>
                      </BoxTitle>
                      <span className="text-[11px] text-muted-foreground">All binaries cryptographically signed</span>
                    </BoxHeader>
                    <BoxBody>
                      {rel.assets.map((asset) => (
                        <BoxRow key={asset.name} className="flex items-center justify-between py-2.5 text-xs">
                          <div className="flex items-center gap-2.5">
                            <FileArchive className="w-4 h-4 text-primary shrink-0" />
                            <a
                              href={`#download-${asset.name}`}
                              className="font-mono text-primary font-semibold hover:underline"
                            >
                              {asset.name}
                            </a>
                            {asset.sha256 && (
                              <button
                                onClick={() => handleCopy(asset.sha256!, asset.name)}
                                className="px-1.5 py-0.5 rounded bg-secondary border border-border text-[10px] font-mono text-muted-foreground hover:text-foreground flex items-center gap-1 cursor-pointer"
                                title={`SHA256: ${asset.sha256}`}
                              >
                                {copiedSha === asset.name ? <Check className="w-2.5 h-2.5 text-emerald-400" /> : <Copy className="w-2.5 h-2.5" />}
                                <span>checksum</span>
                              </button>
                            )}
                          </div>

                          <div className="flex items-center gap-4 text-muted-foreground font-mono text-[11px]">
                            <span>{formatSize(asset.size)}</span>
                            <span>{asset.downloadCount} downloads</span>
                            <Download className="w-3.5 h-3.5 hover:text-foreground cursor-pointer" />
                          </div>
                        </BoxRow>
                      ))}

                      {/* Source Code Archives */}
                      <BoxRow className="flex items-center justify-between py-2.5 text-xs">
                        <div className="flex items-center gap-2.5 font-mono text-primary">
                          <FileArchive className="w-4 h-4 text-muted-foreground shrink-0" />
                          <span className="hover:underline cursor-pointer">Source code (zip)</span>
                        </div>
                        <Download className="w-3.5 h-3.5 text-muted-foreground hover:text-foreground cursor-pointer" />
                      </BoxRow>
                      <BoxRow className="flex items-center justify-between py-2.5 text-xs">
                        <div className="flex items-center gap-2.5 font-mono text-primary">
                          <FileArchive className="w-4 h-4 text-muted-foreground shrink-0" />
                          <span className="hover:underline cursor-pointer">Source code (tar.gz)</span>
                        </div>
                        <Download className="w-3.5 h-3.5 text-muted-foreground hover:text-foreground cursor-pointer" />
                      </BoxRow>
                    </BoxBody>
                  </Box>
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* 2. TAGS TAB */}
      {activeSubTab === 'tags' && (
        <Box flush>
          <BoxHeader className="py-2.5 text-xs font-semibold text-foreground">
            <span>All Git Tags</span>
          </BoxHeader>
          <BoxBody>
            {releases.map((rel) => (
              <BoxRow key={rel.id} className="flex items-center justify-between py-3">
                <div className="flex items-center gap-3">
                  <Tag className="w-4 h-4 text-primary shrink-0" />
                  <div className="space-y-0.5 font-mono text-xs">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-foreground">{rel.tagName}</span>
                      <Badge variant="outline" className="border-emerald-500/40 text-emerald-400 bg-emerald-500/10 text-[10px] font-sans flex items-center gap-1">
                        <ShieldCheck className="w-3 h-3" />
                        <span>Verified</span>
                      </Badge>
                    </div>
                    <div className="text-muted-foreground text-[11px]">
                      Commit <span className="text-primary hover:underline cursor-pointer">{rel.commitSha}</span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <Button variant="secondary" size="xs" className="gap-1 font-mono text-xs">
                    <Download className="w-3 h-3" />
                    <span>zip</span>
                  </Button>
                  <Button variant="secondary" size="xs" className="gap-1 font-mono text-xs">
                    <Download className="w-3 h-3" />
                    <span>tar.gz</span>
                  </Button>
                </div>
              </BoxRow>
            ))}
          </BoxBody>
        </Box>
      )}

      {/* DRAFT NEW RELEASE MODAL */}
      <Dialog
        isOpen={isDraftModalOpen}
        onClose={() => setIsDraftModalOpen(false)}
        title="Draft a new release"
        description="Publish release notes and attach compiled artifacts for distribution."
      >
        <form onSubmit={handlePublishRelease} className="space-y-4 pt-1">
          <div className="space-y-3 text-xs">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block font-medium text-foreground mb-1">Tag version</label>
                <input
                  type="text"
                  placeholder="v0.5.0"
                  value={draftTag}
                  onChange={(e) => setDraftTag(e.target.value)}
                  className="w-full px-3 py-1.5 rounded-md bg-input border border-border text-foreground font-mono focus:ring-1 focus:ring-primary focus:outline-none"
                  required
                />
              </div>

              <div>
                <label className="block font-medium text-foreground mb-1">Target branch</label>
                <select className="w-full px-3 py-1.5 rounded-md bg-secondary border border-border text-foreground focus:outline-none font-mono">
                  <option value="main">main (5c92f55)</option>
                  <option value="develop">develop</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Release title</label>
              <input
                type="text"
                placeholder="Release title"
                value={draftTitle}
                onChange={(e) => setDraftTitle(e.target.value)}
                className="w-full px-3 py-1.5 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none"
                required
              />
            </div>

            <div>
              <label className="block font-medium text-foreground mb-1">Describe this release</label>
              <textarea
                placeholder="Markdown release notes, feature highlights, and breaking changes..."
                value={draftBody}
                onChange={(e) => setDraftBody(e.target.value)}
                rows={5}
                className="w-full p-3 rounded-md bg-input border border-border text-foreground focus:ring-1 focus:ring-primary focus:outline-none font-sans"
              />
            </div>

            <div className="space-y-2 pt-1 border-t border-border">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isLatest}
                  onChange={(e) => setIsLatest(e.target.checked)}
                  className="rounded border-border text-primary focus:ring-primary"
                />
                <span className="text-foreground font-medium">Set as the latest release</span>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isPrerelease}
                  onChange={(e) => setIsPrerelease(e.target.checked)}
                  className="rounded border-border text-primary focus:ring-primary"
                />
                <span className="text-foreground font-medium">Set as a pre-release</span>
              </label>
            </div>
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsDraftModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={isPublishing || !draftTag.trim() || !draftTitle.trim()}>
              {isPublishing ? 'Publishing...' : 'Publish release'}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
