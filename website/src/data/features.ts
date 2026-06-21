import type { Feature } from "../types";

export const features: Feature[] = [
  {
    icon: "checkCircle",
    title: "Automatic Routing",
    description:
      "Create rules by domain or regex. Links open in the right browser automatically.",
  },
  {
    icon: "sliders",
    title: "Browser Launcher",
    description:
      "No matching rule? Launch a browser with a quick keystroke or mouse click.",
  },
  {
    icon: "shuffle",
    title: "Link Redirections",
    description:
      "Rewrite URLs before they open. Strip trackers, swap domains, and more.",
  },
  {
    icon: "link",
    title: "Custom URLs",
    description:
      "Encode rules directly with switchyard:// links. Perfect for notes and to-dos.",
  },
  {
    icon: "star",
    title: "Favorite Browser",
    description:
      "Pin your favorite browser to appear first and act as fallback.",
  },
  {
    icon: "zap",
    title: "Fast & Native",
    description:
      "A native GNOME app that launches instantly. No background services.",
  },
];
