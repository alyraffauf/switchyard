import type { InstallOption } from "../types";

const REPO_URL = "https://github.com/alyraffauf/switchyard";
const FLATHUB_URL = "https://flathub.org/apps/io.github.alyraffauf.Switchyard";

export const installOptions: InstallOption[] = [
  {
    icon: "flatpak",
    title: "Flatpak",
    description: "Recommended for most users",
    link: FLATHUB_URL,
    linkText: "View on Flathub",
  },
  {
    icon: "nixos",
    title: "Nix Flake",
    description: "Run directly or add to your config",
    command: "nix run github:alyraffauf/switchyard",
  },
  {
    icon: "github",
    title: "Build from Source",
    description: "Requires Go 1.24+ and GTK4",
    link: REPO_URL,
    linkText: "View on GitHub",
  },
];
