# Rate It

A full-stack web application for rating and discovering movies and TV series.

## Projects

### Backend — Go API

A REST API built with Go, following a clean architecture (handler → usecase → repository). It handles authentication, user lists, and media metadata fetched from the TMDB API. Data is persisted in PostgreSQL, with Redis used for session caching.

**Libraries**

| Library | Role |
|---|---|
| [chi](https://github.com/go-chi/chi) | HTTP router and middleware |
| [chi/cors](https://github.com/go-chi/cors) | CORS middleware |
| [sqlx](https://github.com/jmoiron/sqlx) | SQL query builder on top of `database/sql` |
| [pgx](https://github.com/jackc/pgx) | PostgreSQL driver |
| [go-redis](https://github.com/redis/go-redis) | Redis client |
| [golang-jwt](https://github.com/golang-jwt/jwt) | JWT authentication |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Database migrations |
| [godotenv](https://github.com/joho/godotenv) | `.env` file loading |
| [google/uuid](https://github.com/google/uuid) | UUID generation |

### Frontend — Nuxt 4

A Nuxt 4 app built on Vue.js. Users can discover movies and TV series, view details, rate them, and manage personal lists. Communicates with the Go API.

**Libraries**

| Library | Role |
|---|---|
| [Vue.js](https://vuejs.org) | Underlying component framework |
| [Nuxt UI](https://ui.nuxt.com) | Component library (buttons, modals, forms, etc.) |
| [Tailwind CSS v4](https://tailwindcss.com) | Utility-first styling |
| [Iconify / Lucide](https://lucide.dev) | Icon set |
| [TypeScript](https://www.typescriptlang.org) | Type safety across the frontend |
| [ESLint](https://eslint.org) + [Prettier](https://prettier.io) | Linting and formatting |

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | [Nuxt 4](https://nuxt.com) + [Nuxt UI](https://ui.nuxt.com) |
| Backend | Go ([chi](https://github.com/go-chi/chi) router) |
| Database | PostgreSQL 17 (Docker) |
| Cache / Sessions | Redis 7 (Docker) |
| Media Metadata | [TMDB API](https://www.themoviedb.org/documentation/api) |
| Task Runner | [mise](https://mise.jdx.dev/) |

## Prerequisites

- [mise](https://mise.jdx.dev/getting-started.html) — manages Go, Bun, and all dev tasks
- [Docker](https://docs.docker.com/get-docker/) — **required** to run PostgreSQL and Redis locally

## Getting Started

**1. Clone and install dependencies**

```bash
git clone https://github.com/hlaclau/rate-it.git
cd rate-it
mise run install
```

**2. Configure the backend**

Copy the example env file and fill in your values:

```bash
cp backend/.env.example backend/.env
```

The only value you need to set manually is your TMDB API key:

```
TMDB_API_KEY=your_key_here
```

Get a free key at [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api). All other defaults work out of the box with Docker.

**3. Start everything**

```bash
mise run dev
```

This spins up PostgreSQL and Redis via Docker, then starts the Go API on port `8080` and the Nuxt frontend on port `3000`.

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Development Commands

| Command | Description |
|---|---|
| `mise run install` | Install all dependencies |
| `mise run dev` | Start everything (API + frontend + databases) |
| `mise run dev-api` | Start the Go API only (port 8080) |
| `mise run dev-ui` | Start the Nuxt frontend only (port 3000) |
| `mise run stop` | Stop API, frontend and databases |
| `mise run db-up` | Start PostgreSQL and Redis (Docker) |
| `mise run db-down` | Stop PostgreSQL and Redis |
| `mise run db-reset` | Wipe and restart databases (destroys all data) |
| `mise run lint` | Lint frontend and backend |
| `mise run format` | Format frontend and backend |
| `mise run test` | Run all backend tests |
| `mise run test-handlers` | Run handler tests only |
| `mise run test-ci` | Run all tests without cache (for CI) |

To see all available commands at any time:

```bash
mise tasks ls
```

## Project Structure

```
rate-it/
├── backend/              # Go API
│   ├── cmd/              # Entry point
│   ├── internal/
│   │   ├── handler/      # HTTP handlers
│   │   ├── usecase/      # Business logic
│   │   ├── repository/   # Database access
│   │   ├── domain/       # Core types
│   │   ├── adapter/      # External integrations (TMDB)
│   │   └── port/         # Interface definitions
│   ├── migrations/       # SQL migrations
│   └── .env.example      # Environment variable reference
├── frontend/             # Nuxt 4 app
│   └── app/
│       ├── pages/        # File-based routing
│       ├── components/   # Vue components
│       ├── composables/  # Shared logic
│       └── middleware/   # Route guards
└── docker-compose.dev.yml  # PostgreSQL + Redis for local dev
```

## API Routes

The Go API is available at `http://localhost:8080`. All `/api/*` routes are prefixed automatically.

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/register` | — | Register a new user |
| `POST` | `/api/login` | — | Log in and start a session |
| `POST` | `/api/refresh` | Cookie | Refresh the session token |
| `POST` | `/api/logout` | Cookie | Invalidate the session |
| `GET` | `/api/me` | Cookie | Get the authenticated user |

### Media

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/media/movie/{id}` | — | Get movie details by TMDB ID |
| `GET` | `/api/media/series/{id}` | — | Get series details by TMDB ID |
| `GET` | `/api/media/search` | — | Search or discover media |

`/api/media/search` query parameters: `q`, `type`, `page`, `sort_by`, `year_from`, `year_to`, `vote_average_min`, `vote_average_max`, `vote_count_min`, `with_genres`, `watch_providers`, `watch_region`, `language`, `include_adult`.

### List

All list routes require a valid session cookie.

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/list` | Get the authenticated user's list |
| `POST` | `/api/list` | Add or update a media entry |
| `PUT` | `/api/list/{mediaID}` | Update an existing entry |
| `DELETE` | `/api/list/{mediaID}` | Remove an entry |
| `GET` | `/api/list/status/{externalID}` | Get the list status of a media item |

### Users

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/users/search` | — | Search users by username (`q` param, min 2 chars) |
| `GET` | `/api/users/{username}/list` | — | Get a user's public list |

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |

Bruno API collection is available in `docs/api/`.

## Environment Variables

Backend configuration lives in `backend/.env`. Copy `backend/.env.example` to get started — all defaults are pre-configured for the local Docker setup.

| Variable | Description | Default |
|---|---|---|
| `PORT` | API server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://rateit:rateit@localhost:5432/rateit` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `TMDB_API_KEY` | TMDB API key (**required**) | — |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |
