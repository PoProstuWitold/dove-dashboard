# The Dove Dashboard

## ***"Essential system stats, peacefully simple."***

A lightweight, simple and peaceful web-based system monitor - written in Go, with no dependencies and a minimal UI.

---

## Features

- **System overview** - OS, architecture, kernel, uptime, hostname.
- **CPU** - brand, model, cores, threads, frequency.
- **Memory & storage** - usage, free/used/total.
- **Sensors** - temperatures and voltages (where available).
- **Network** - main network interface.
- **Live updates** - refreshes every 10 seconds.
- **Minimal UI** - simple and elegant.
- **Self-contained** - single binary + static assets.

## Docker

You can download image from [DockerHub](https://hub.docker.com/repository/docker/poprostuwitold/dove-dashboard/general) and deploy using ``docker compose``:

```yaml
services:
  dove-dashboard:
    container_name: dove-dashboard
    image: poprostuwitold/dove-dashboard:latest
    restart: unless-stopped
    ports:
      - "2137:2137"
    volumes:
      - /:/mnt/host:ro
```