# Switchyard Website

Landing page for [Switchyard](https://github.com/alyraffauf/switchyard), a rules-based URL router for Linux.

Built with Astro. Published as a container image to `ghcr.io/alyraffauf/switchyard-website`.

## Development

```bash
bun install
bun run dev
```

## Build

```bash
bun run build
```

Outputs to `dist/`.

## Container

```bash
docker build -t switchyard-website .
docker run --rm -p 8080:80 switchyard-website
```
