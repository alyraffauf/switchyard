import type { HeroButton } from "../types";
import { DONATE_URL, FLATHUB_URL, REPO_URL } from "./urls";

export const heroButtons: HeroButton[] = [
  {
    href: FLATHUB_URL,
    label: "Flathub",
    icon: "flatpak",
    primary: true,
  },
  {
    href: REPO_URL,
    label: "GitHub",
    icon: "github",
    primary: true,
  },
  {
    href: DONATE_URL,
    label: "Donate",
    icon: "heart",
  },
];
