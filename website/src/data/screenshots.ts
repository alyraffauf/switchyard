import type { Screenshot } from "../types";

const SCREENSHOTS_BASE = `https://raw.githubusercontent.com/alyraffauf/switchyard/master/data/screenshots`;

export const screenshots: Screenshot[] = [
  {
    src: `${SCREENSHOTS_BASE}/launcher.png`,
    alt: "Browser launcher",
  },
  {
    src: `${SCREENSHOTS_BASE}/browser-rules.png`,
    alt: "Browser rules panel",
  },
  {
    src: `${SCREENSHOTS_BASE}/link-redirections.png`,
    alt: "Link redirections panel",
  },
  {
    src: `${SCREENSHOTS_BASE}/edit-rules.png`,
    alt: "Rule editor",
  },
];
