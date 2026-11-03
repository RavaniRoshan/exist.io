import defaultMdxComponents from 'fumadocs-ui/mdx';
import type { MDXComponents } from 'mdx/types';
import Link from 'next/link';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';
import { Kbd, KbdGroup } from '@/components/ui/kbd';

const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '/exist.io';

function Badge({ children }: { children: ReactNode }) {
  return (
    <span className="mr-1 inline-flex items-center rounded-full border border-fd-border bg-fd-muted px-2 py-0.5 font-mono text-[11px] tracking-tight text-fd-muted-foreground">
      {children}
    </span>
  );
}

function Frame({
  caption,
  src,
  alt,
  children,
}: {
  caption?: string;
  src?: string;
  alt?: string;
  children?: ReactNode;
}) {
  const resolved =
    typeof src === 'string' && src.startsWith('/') && !src.startsWith('//')
      ? `${basePath}${src}`
      : src;

  return (
    <figure className="my-6 overflow-hidden rounded-xl border border-fd-border bg-fd-card">
      {resolved ? (
        <img src={resolved} alt={alt ?? ''} className="block w-full" />
      ) : (
        <div className="[&_img]:block [&_img]:w-full">{children}</div>
      )}
      {caption ? (
        <figcaption className="border-t border-fd-border px-4 py-3 text-xs text-fd-muted-foreground">
          {caption}
        </figcaption>
      ) : null}
    </figure>
  );
}

function CardGroup({ cols = 2, children }: { cols?: 1 | 2 | 3; children: ReactNode }) {
  const layout =
    cols === 3 ? 'md:grid-cols-3' : cols === 1 ? 'md:grid-cols-1' : 'md:grid-cols-2';

  return <div className={`my-6 grid grid-cols-1 gap-3 ${layout}`}>{children}</div>;
}

function Card({ title, href, children }: { title: ReactNode; href: string; children: ReactNode }) {
  return (
    <Link
      href={href}
      className="group flex flex-col rounded-xl border border-fd-border bg-fd-card p-5 transition-colors hover:border-fd-primary/50 hover:bg-fd-accent/40"
    >
      <span className="text-sm font-medium text-fd-foreground">{title}</span>
      <span className="mt-2 text-sm leading-relaxed text-fd-muted-foreground [&_code]:text-[0.8em]">
        {children}
      </span>
      <span className="mt-4 font-mono text-[11px] text-fd-primary opacity-0 transition-opacity group-hover:opacity-100">
        read →
      </span>
    </Link>
  );
}

function Img({ src, alt, ...props }: ComponentPropsWithoutRef<'img'>) {
  const resolved =
    typeof src === 'string' && src.startsWith('/') && !src.startsWith('//')
      ? `${basePath}${src}`
      : src;

  return <img src={resolved} alt={alt} {...props} />;
}

function A({ href, ...props }: ComponentPropsWithoutRef<'a'>) {
  if (typeof href === 'string' && href.startsWith('/') && !href.startsWith('//')) {
    return <Link href={href} {...props} />;
  }

  return <a href={href} {...props} />;
}

export function getMDXComponents(components?: MDXComponents) {
  return {
    ...defaultMdxComponents,
    Badge,
    Frame,
    CardGroup,
    Card,
    Kbd,
    KbdGroup,
    img: Img,
    a: A,
    ...components,
  } satisfies MDXComponents;
}

export const useMDXComponents = getMDXComponents;

declare global {
  type MDXProvidedComponents = ReturnType<typeof getMDXComponents>;
}
