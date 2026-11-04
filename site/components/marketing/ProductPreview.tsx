'use client';

import Image from 'next/image';
import { useRef, useState, type KeyboardEvent } from 'react';
import { asset } from '@/lib/asset';

const TABS = [
  {
    id: 'home',
    label: 'forge home',
    src: '/product/forge-home.png',
    alt: 'exist.io web UI — repository overview with agent activity feed and cluster telemetry',
    caption: 'fig. 2 — web ui · repository activity feed',
    meta: '1440 × 900 · real ui',
  },
  {
    id: 'agents',
    label: 'agent console',
    src: '/product/forge-agents.png',
    alt: 'exist.io agent console — live MCP invocations, capability leases and tool catalog',
    caption: 'fig. 3 — agent console · MCP invocations and leases',
    meta: '1440 × 900 · real ui',
  },
  {
    id: 'pr',
    label: 'pull request',
    src: '/product/forge-pr.png',
    alt: 'exist.io pull request view — checks passing, capability lease verified, merge gate ready',
    caption: 'fig. 4 — pull request · checks passed · merge gate',
    meta: '1440 × 900 · real ui',
  },
] as const;

export function ProductPreview() {
  const [active, setActive] = useState(0);
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const tab = TABS[active];

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    const last = TABS.length - 1;
    let next = active;
    if (event.key === 'ArrowRight') next = active === last ? 0 : active + 1;
    else if (event.key === 'ArrowLeft') next = active === 0 ? last : active - 1;
    else if (event.key === 'Home') next = 0;
    else if (event.key === 'End') next = last;
    else return;
    event.preventDefault();
    setActive(next);
    tabRefs.current[next]?.focus();
  };

  return (
    <div className="window">
      <div className="window-bar">
        <span className="window-dots" aria-hidden="true">
          <span />
          <span />
          <span />
        </span>
        <span className="window-title">acme/demo — existd :8080</span>
        <span className="mono ml-auto hidden text-[9.5px] uppercase tracking-[0.16em] text-text-3 sm:inline">
          live ui
        </span>
      </div>

      <div
        role="tablist"
        aria-label="Product screenshots"
        className="flex overflow-x-auto border-b border-line bg-raised"
        onKeyDown={onKeyDown}
      >
        {TABS.map((item, index) => (
          <button
            key={item.id}
            ref={(node) => {
              tabRefs.current[index] = node;
            }}
            type="button"
            role="tab"
            id={`product-tab-${item.id}`}
            aria-selected={active === index}
            aria-controls="product-panel"
            tabIndex={active === index ? 0 : -1}
            className="tab min-w-max"
            onClick={() => setActive(index)}
          >
            <span className={`h-1.5 w-1.5 rounded-full ${active === index ? 'bg-patina' : 'bg-line-strong'}`} aria-hidden="true" />
            {item.label}
          </button>
        ))}
        <span className="mono ml-auto hidden items-center px-4 text-[9.5px] uppercase tracking-[0.14em] text-text-3 md:flex">
          select to inspect
        </span>
      </div>

      <div id="product-panel" role="tabpanel" aria-labelledby={`product-tab-${tab.id}`}>
        <div key={tab.id} className="scanline relative aspect-[16/10] overflow-hidden bg-canvas">
          <Image
            src={asset(tab.src)}
            alt={tab.alt}
            width={1440}
            height={900}
            sizes="(min-width: 1400px) 1336px, 100vw"
            className="h-full w-full object-cover object-top"
          />
        </div>
        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-line px-4 py-3">
          <span className="mono text-[10.5px] text-text-3">{tab.caption}</span>
          <span className="mono rounded-sm border border-patina/40 bg-patina-soft px-2 py-1 text-[9px] font-semibold uppercase tracking-[0.16em] text-patina">
            {tab.meta}
          </span>
        </div>
      </div>
    </div>
  );
}
