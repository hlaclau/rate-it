# Rate It

## Development Tasks

This project uses [mise](https://mise.jdx.dev/) as a task runner and environment manager. All custom tasks are defined inside the `.tasks` directory and can be executed using `mise run <task>`.

To view the list of available commands at any time in your terminal, run:

```bash
mise tasks ls
```

### Available Commands

Here are the commands you can run for local development:

| Command | Description |
|---|---|
| `mise run install` | Install all dependencies |
| `mise run dev` | Start everything (api + frontend + databases) |
| `mise run dev-api` | Start the Go API (port 8080) |
| `mise run dev-ui` | Start the Nuxt frontend (port 3000) |
| `mise run stop` | Stop api, frontend and databases |
| `mise run db-up` | Start PostgreSQL and Redis |
| `mise run db-down` | Stop PostgreSQL and Redis |
| `mise run db-reset` | Wipe and restart databases (destroys data) |
| `mise run lint` | Lint frontend and backend |
| `mise run format` | Format frontend and backend |
| `mise run test` | Run all backend tests (compact dots output) |
| `mise run test-pretty` | Run all backend tests (verbose colored output) |
| `mise run test-handlers` | Run handler tests only |
| `mise run test-ci` | Run all tests without cache (for CI) |
