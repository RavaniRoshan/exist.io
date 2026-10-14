import { defineConfig } from "blume";

export default defineConfig({
  title: "exist.io",
  description: "A self-hosted Git forge for people and coding agents.",
  github: {
    owner: "RavaniRoshan",
    repo: "exist.io",
  },
  theme: {
    accent: {
      light: "#2D6A57",
      dark: "#76B49A",
    },
  },
  deployment: {
    site: "https://ravaniroshan.github.io",
    base: "/exist.io",
  },
});
