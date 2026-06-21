import type { IconName } from "./components/Icon";

export interface Screenshot {
  src: string;
  alt: string;
}

export interface Feature {
  icon: IconName;
  title: string;
  description: string;
}

export interface HeroButton {
  href: string;
  label: string;
  icon?: IconName;
  primary?: boolean;
}

interface InstallOptionCommand {
  icon: IconName;
  title: string;
  description: string;
  command: string;
}

interface InstallOptionLink {
  icon: IconName;
  title: string;
  description: string;
  link: string;
  linkText: string;
}

export type InstallOption = InstallOptionCommand | InstallOptionLink;

export interface Showcase {
  screenshot?: Screenshot;
  code?: string;
  title: string;
  description: string;
}

export interface FooterLink {
  href: string;
  label: string;
}
