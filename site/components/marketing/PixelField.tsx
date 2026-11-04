'use client';

import { useEffect, useRef } from 'react';

type Agent = {
  x: number;
  y: number;
  vx: number;
  vy: number;
  trail: Array<[number, number]>;
  tone: 'patina' | 'brass';
};

/**
 * A sparse pixel field: a quiet blueprint grid, a handful of agent cells
 * crossing it, and a cursor that wakes up nearby pixels. Decorative only —
 * hidden from assistive tech, frozen under reduced-motion.
 */
export function PixelField({ className = '' }: { className?: string }) {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const pointer = { x: -999, y: -999, active: false };
    let width = 0;
    let height = 0;
    let cell = 14;
    let cols = 0;
    let rows = 0;
    let raf = 0;
    let last = 0;
    let visible = true;

    let colors = { ink: '#1f2328', patina: '#2e6f63', brass: '#a67c2e' };
    const readColors = () => {
      const styles = getComputedStyle(document.documentElement);
      colors = {
        ink: styles.getPropertyValue('--text-1').trim() || colors.ink,
        patina: styles.getPropertyValue('--patina').trim() || colors.patina,
        brass: styles.getPropertyValue('--brass').trim() || colors.brass,
      };
    };
    readColors();

    let agents: Agent[] = [];
    const seed = () => {
      const seeds: Array<[number, number, number, number, Agent['tone']]> = [
        [0.08, 0.22, 1, 0, 'patina'],
        [0.3, 0.68, 0, -1, 'brass'],
        [0.55, 0.3, -1, 0, 'patina'],
        [0.72, 0.78, 1, 0, 'patina'],
        [0.9, 0.44, 0, 1, 'brass'],
      ];
      agents = seeds.map(([x, y, vx, vy, tone]) => ({
        x: Math.floor(cols * x),
        y: Math.floor(rows * y),
        vx,
        vy,
        trail: [],
        tone,
      }));
    };

    const resize = () => {
      const rect = canvas.getBoundingClientRect();
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      width = rect.width;
      height = rect.height;
      cell = width < 520 ? 10 : 14;
      cols = Math.ceil(width / cell);
      rows = Math.ceil(height / cell);
      canvas.width = Math.round(width * dpr);
      canvas.height = Math.round(height * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      seed();
      draw();
    };

    const draw = () => {
      ctx.clearRect(0, 0, width, height);

      // Blueprint dots.
      ctx.globalAlpha = 0.14;
      ctx.fillStyle = colors.ink;
      for (let y = 1; y < rows; y += 2) {
        for (let x = 1; x < cols; x += 2) {
          ctx.fillRect(x * cell, y * cell, 1.5, 1.5);
        }
      }

      // Cursor wake.
      if (pointer.active) {
        ctx.globalAlpha = 0.32;
        ctx.fillStyle = colors.patina;
        const px = pointer.x / cell;
        const py = pointer.y / cell;
        for (let y = Math.floor(py - 5); y <= py + 5; y++) {
          for (let x = Math.floor(px - 5); x <= px + 5; x++) {
            const distance = Math.hypot(x - px, y - py);
            if (distance > 5) continue;
            ctx.globalAlpha = 0.3 * (1 - distance / 5);
            ctx.fillRect(x * cell, y * cell, 2, 2);
          }
        }
      }

      // Agent trails.
      for (const agent of agents) {
        for (let i = 0; i < agent.trail.length; i++) {
          const [tx, ty] = agent.trail[i];
          const fade = (i + 1) / (agent.trail.length + 1);
          ctx.globalAlpha = 0.34 * fade;
          ctx.fillStyle = agent.tone === 'patina' ? colors.patina : colors.brass;
          const inset = (1 - fade) * 4;
          ctx.fillRect(tx * cell + inset, ty * cell + inset, cell - inset * 2, cell - inset * 2);
        }
      }

      // Agents.
      for (const agent of agents) {
        ctx.globalAlpha = 0.9;
        ctx.fillStyle = agent.tone === 'patina' ? colors.patina : colors.brass;
        ctx.fillRect(agent.x * cell + 2, agent.y * cell + 2, cell - 4, cell - 4);
        ctx.globalAlpha = 1;
        ctx.fillStyle = colors.ink;
        ctx.fillRect(agent.x * cell + cell / 2 - 1.5, agent.y * cell + cell / 2 - 1.5, 3, 3);
      }
      ctx.globalAlpha = 1;
    };

    const step = () => {
      for (const agent of agents) {
        agent.trail.push([agent.x, agent.y]);
        if (agent.trail.length > 5) agent.trail.shift();

        const nx = agent.x + agent.vx;
        const ny = agent.y + agent.vy;
        if (nx < 1 || nx > cols - 2) agent.vx *= -1;
        if (ny < 1 || ny > rows - 2) agent.vy *= -1;
        agent.x = Math.max(1, Math.min(cols - 2, nx));
        agent.y = Math.max(1, Math.min(rows - 2, ny));
      }
    };

    const loop = (time: number) => {
      raf = requestAnimationFrame(loop);
      if (!visible || reduce) return;
      if (time - last < 96) return;
      last = time;
      step();
      draw();
    };

    const onPointer = (event: PointerEvent) => {
      const rect = canvas.getBoundingClientRect();
      pointer.x = event.clientX - rect.left;
      pointer.y = event.clientY - rect.top;
      pointer.active = true;
      if (reduce) draw();
    };
    const onLeave = () => {
      pointer.active = false;
      if (reduce) draw();
    };

    const observer = new ResizeObserver(() => resize());
    observer.observe(canvas);
    const intersection = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
    });
    intersection.observe(canvas);
    const theme = new MutationObserver(() => {
      readColors();
      draw();
    });
    theme.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

    resize();
    if (!reduce) {
      canvas.addEventListener('pointermove', onPointer);
      canvas.addEventListener('pointerleave', onLeave);
      raf = requestAnimationFrame(loop);
    }

    return () => {
      cancelAnimationFrame(raf);
      observer.disconnect();
      intersection.disconnect();
      theme.disconnect();
      canvas.removeEventListener('pointermove', onPointer);
      canvas.removeEventListener('pointerleave', onLeave);
    };
  }, []);

  return <canvas ref={ref} aria-hidden="true" className={className} />;
}
