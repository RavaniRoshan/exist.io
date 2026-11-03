import { defineConfig } from "blume";

export default defineConfig({
  title: "exist.io",
  description:
    "The self-hosted Git forge built for coding agents. Vendor-neutral, single-binary, and designed for autonomous pull requests.",
  github: {
    owner: "RavaniRoshan",
    repo: "exist.io",
  },
  theme: {
    accent: { light: "#0f7a5a", dark: "#5ee0b0" },
    background: { light: "#fbfaf7", dark: "#0f1114" },
    radius: "md",
    mode: "system",
    fonts: {
      display: "inter-tight",
      body: "inter",
      mono: "jetbrains-mono",
    },
  },
  banner: {
    content:
      "v0.6.0-beta is out: Forge UI, verified installs, and consent-based activation.",
    link: {
      href: "https://github.com/RavaniRoshan/exist.io/releases/tag/v0.6.0-beta",
      text: "Release notes",
    },
    dismissible: true,
  },
  navigation: {
    cta: { label: "Install", href: "/getting-started" },
  },
  deployment: {
    site: "https://ravaniroshan.github.io",
    base: "/exist.io",
  },
});
