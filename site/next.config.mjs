import { createMDX } from 'fumadocs-mdx/next';

const basePath = '/exist.io';

// Make the base path available while the app renders during `next build`,
// not only in the client bundle.
process.env.NEXT_PUBLIC_BASE_PATH ??= basePath;

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  output: 'export',
  basePath,
  trailingSlash: true,
  reactStrictMode: true,
  images: {
    unoptimized: true,
  },
  env: {
    NEXT_PUBLIC_BASE_PATH: basePath,
  },
};

export default withMDX(config);
