# Pingopher

A self-hosted uptime monitoring tool with **multi-DNS support** — ping a single host through multiple DNS resolvers simultaneously and compare results side by side.

## Features

- **Multi-DNS monitoring** — define custom DNS resolvers per host; Pingopher pings through all of them in parallel and records each result independently
- **Scheduled health checks** — cron-based HTTP(S) pings with configurable intervals per host
- **Dynamic backoff** — automatically reduces check frequency when a host is down, using exponential backoff up to a configurable max interval
- **Status change detection** — detects up→down and down→up transitions, re-sending notifications on every change
- **Repeated failure alerts** — re-notifies after N consecutive failures (configurable threshold)
- **Latency charting** — visualize latency over time with G2Plot line charts, broken down by DNS resolver
- **Downtime timeline** — view incident events with duration, DNS resolver, and error details
- **Discord notifications** — rich embed alerts via webhooks with per-DNS status codes, latency, and error messages; supports forum threads and thread posting
- **Dashboard UI** — React + Ant Design interface with dark/light mode
- **RESTful API** — full CRUD for hosts and notification channels
- **Redis caching** — fast in-memory caching with Valkey/Redis
- **Multiple database backends** — SQLite (local) or Cloudflare D1 (edge)

## Tech Stack

| Layer       | Technology                                                           |
| ----------- | -------------------------------------------------------------------- |
| Backend     | Go 1.26, Gin, GORM, Resty v3                                        |
| Frontend    | React 19, TypeScript, Ant Design 6, Vite, @antv/g2plot              |
| Database    | SQLite, Cloudflare D1                                                |
| Cache       | Valkey (Redis-compatible)                                            |
| Scheduler   | robfig/cron v3                                                       |
| IDs         | KSUID (segmentio/ksuid)                                              |
| Notify      | Discord webhooks (rich embeds)                                       |

## Why Multi-DNS?

Standard uptime monitors hit a single DNS resolver (usually the system default) and report whether the host is reachable. Pingopher lets you configure multiple DNS resolvers per host and pings through **all of them in parallel**. This helps you detect:

- **DNS propagation delays** — different resolvers returning different IPs after a record change
- **DNS hijacking or poisoning** — a resolver returning a rogue IP
- **Geographic routing differences** — CDNs returning region-specific IPs
- **Resolver outages** — one resolver failing while others succeed

Each ping result (status code, latency, DNS used, error) is recorded independently, so you can see exactly how each resolver performed at any point in time.

## Getting Started

> **Tip:** Pingopher is designed for internal use (e.g. homelab). It has **no built-in authentication**. If you want to expose it to the public internet, you must add your own auth layer, such as a reverse proxy (Caddy, nginx, Traefik) with basic auth or OAuth.

### Prerequisites

- Docker & Docker Compose

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/DarknessKiller/pingopher.git
   cd pingopher
   ```

2. Create a `.env` file from the example:
   ```sh
   cp .env.example .env
   ```

3. Start the services:
   ```sh
   docker compose up -d
   ```

4. Open the dashboard at `http://localhost:8080`

The compose file starts two services:
- **pingopher** — the Go backend serving the API and embedded React frontend
- **redis** — Valkey (Redis-compatible) for caching

### Configuration

Configure via environment variables in `.env`:

| Variable                        | Default              | Description                              |
| ------------------------------- | -------------------- | ---------------------------------------- |
| `PINGOPHER_HOST`                | `0.0.0.0`            | Server listen address                    |
| `PINGOPHER_PORT`                | `8080`               | Server port                              |
| `PINGOPHER_DB_TYPE`             | `sqlite`             | Database backend (`sqlite` or `cloudflare-d1`) |
| `PINGOPHER_DB_PATH`             | `pingopher.db`       | SQLite file path                         |
| `PINGOPHER_REDIS_HOST`          | `pingopher_redis`    | Redis host                               |
| `PINGOPHER_REDIS_PORT`          | `6379`               | Redis port                               |
| `PINGOPHER_REDIS_PASSWORD`      |                      | Redis password                           |
| `PINGOPHER_MAX_RETRY_INTERVAL`  | `900`                | Max backoff interval (seconds) when down  |

Cloudflare D1 requires additional `PINGOPHER_CF_D1_*` variables (account ID, auth token, database string).

### API Endpoints

| Method | Path                                        | Description                     |
| ------ | ------------------------------------------- | ------------------------------- |
| GET    | `/api/v1/health`                            | Health check                    |
| POST   | `/api/v1/migration`                         | Run database migrations         |
| POST   | `/api/v1/uptime/create`                     | Create a monitored host         |
| GET    | `/api/v1/uptime/all`                        | List all hosts                  |
| PUT    | `/api/v1/uptime/:hostId`                    | Update a host                   |
| DELETE | `/api/v1/uptime/:hostId`                    | Delete a host                   |
| GET    | `/api/v1/uptime/:hostId`                    | Ping a host on-demand           |
| GET    | `/api/v1/uptime/:hostId/history`            | Get ping history (query: `startAt`, `endAt` in RFC3339) |
| POST   | `/api/v1/uptime/:hostId/notification`       | Create notification for host    |
| GET    | `/api/v1/uptime/:hostId/notification`       | List notifications for host     |
| PUT    | `/api/v1/uptime/:hostId/notification/:notificationId` | Update notification |
| DELETE | `/api/v1/uptime/:hostId/notification/:notificationId` | Delete notification |

## License

MIT
