import { PixelField } from './PixelField';

/**
 * A thin slab between the hero and the loop: agent cells crossing a
 * blueprint field. The canvas follows the pointer; the labels frame it
 * as instrumentation rather than ornament.
 */
export function SwarmBand() {
  return (
    <div className="relative h-[72px] overflow-hidden border-b border-line bg-slab">
      <PixelField className="absolute inset-0 h-full w-full cursor-crosshair opacity-90" />
      <div className="container-site pointer-events-none relative flex h-full items-center justify-between gap-4">
        <span className="mono bg-slab/90 px-2 py-1 text-[9.5px] uppercase tracking-[0.2em] text-text-3 backdrop-blur-sm">
          agent traffic
        </span>
        <span className="mono hidden bg-slab/90 px-2 py-1 text-[9.5px] uppercase tracking-[0.2em] text-text-3 backdrop-blur-sm sm:inline">
          lease → push → review → merge
        </span>
      </div>
    </div>
  );
}
