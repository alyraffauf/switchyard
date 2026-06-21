import type { Showcase } from "../types";
import { screenshots } from "./screenshots";

export const showcases: Showcase[] = [
  {
    screenshot: screenshots[1],
    title: "Point-and-Click Setup",
    description:
      "Create and manage rules with a clean, native interface. No config files required—just point and click.",
  },
  {
    screenshot: screenshots[2],
    title: "Your Links, Your Way",
    description:
      "Take control of every link you open. Strip tracking parameters and enforce privacy-friendly frontends automatically.",
  },
  {
    code: `[[rules]]
name = "Video Sites"
logic = "any"
browser = "brave-browser.desktop"

[[rules.conditions]]
type = "domain"
pattern = "youtube.com"

[[rules.conditions]]
type = "domain"
pattern = "twitch.tv"`,
    title: "Plain Text Configuration",
    description:
      "Prefer text editors? Define rules in a simple TOML file. Easy to version control, share, and back up.",
  },
];
