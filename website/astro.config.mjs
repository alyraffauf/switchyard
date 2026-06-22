import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import sitemap from "@astrojs/sitemap";
import icon from "astro-icon";

export default defineConfig({
  site: "https://switchyard.aly.codes",
  integrations: [
    sitemap(),
    icon(),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
  build: {
    assets: "assets",
  },
});
