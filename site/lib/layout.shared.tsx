import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { Logo } from '@/components/logo';

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: <Logo />,
      url: '/',
    },
    githubUrl: 'https://github.com/RavaniRoshan/exist.io',
    links: [
      {
        text: 'Marketing site',
        url: 'https://ravaniroshan.github.io/exist.os/',
        external: true,
      },
      {
        text: 'Releases',
        url: 'https://github.com/RavaniRoshan/exist.io/releases',
        external: true,
      },
    ],
  };
}
