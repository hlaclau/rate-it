# Public User Lists & Community Search

**Date:** 2026-05-06  
**Status:** Approved

## Overview

Make every user's list publicly accessible via `/list/[username]`, allow the owner to edit their own list, and add a `/community` page with a user search bar.

## Backend

### New endpoints (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/users/search?q=query` | Search users by username substring |
| `GET` | `/api/users/{username}/list` | Fetch a user's list by username |

### `GET /api/users/search?q=query`
- Returns `[]` if `q` is under 2 characters (no error).
- Case-insensitive substring match on `username`.
- Capped at 20 results.
- Response shape: `[{ id, username, avatar_url }]`

### `GET /api/users/{username}/list`
- Resolves username → user ID via `GetByUsername` repo method.
- Delegates list retrieval to the existing `ListUseCase.GetList(ctx, userID)`.
- Returns `404 { "message": "user not found" }` for unknown usernames.
- Response shape: same as existing `GET /api/list`.

### Repository additions (`UserRepository`)
- `GetByUsername(ctx, username) (*domain.User, error)` — exact match lookup.
- `SearchByUsername(ctx, query string, limit int) ([]*domain.User, error)` — case-insensitive `ILIKE '%query%'`, ordered by username, capped at `limit`.

### New files
- `backend/internal/handler/user.go` — `UserHandler` with `Routes`, `Search`, `GetList`. Depends on a `UserUseCase` and the existing `ListUseCase`.
- `backend/internal/port/user.go` — extend with `UserUseCase` interface (`SearchUsers`, `GetByUsername`).
- `backend/internal/usecase/user.go` — thin `UserUseCase` implementation wrapping the new repo methods.

### Routing
Registered under `/api` in `main.go` alongside existing handlers:
```go
userHandler.Routes(r)
```

## Frontend

### Route changes

| Old | New | Notes |
|-----|-----|-------|
| `pages/list.vue` | `pages/list/[username].vue` | Dynamic route, public |
| — | `pages/list/index.vue` | Redirect only |
| — | `pages/community.vue` | New page |

### `pages/list/index.vue`
- No template.
- On setup: if logged in → `navigateTo('/list/' + user.username)`; else → `navigateTo('/')`.

### `pages/list/[username].vue`
- Fetches from `GET /api/users/{username}/list` (public, no auth middleware).
- `isOwn` computed: `route.params.username === loggedInUser?.username`.
- When `isOwn` is true: remove button and edit controls visible.
- When `isOwn` is false: read-only view, controls hidden.
- Page title: `"My List — Rate It"` when own, `"{username}'s List — Rate It"` otherwise.
- "User not found" empty state on 404.
- All existing filter/sort functionality preserved.

### `pages/community.vue`
- Single `UInput` bound to a `query` ref.
- Debounced watcher (300ms): calls `GET /api/users/search?q=query` when `query.length >= 2`; clears results otherwise.
- Results: list of user cards showing avatar + username, each linking to `/list/[username]`.
- Loading skeleton and "No users found" empty state.

### `composables/useUsers.ts` (new)
- `searchUsers(query)` — calls `GET /api/users/search?q=query`, returns `UserSummary[]`.
- `fetchUserList(username)` — calls `GET /api/users/{username}/list`, returns `ListEntry[]`.

## Error handling & security

- Unknown username on list page → 404 → "User not found" UI state (no crash).
- Search errors are swallowed silently (network blip shouldn't break the search bar).
- `isOwn` is UI-only. Mutating list routes (`POST/PUT/DELETE /api/list`) remain auth-gated on the backend — no server-side risk from bypassing the frontend.
- Redirect in `list/index.vue` happens before render (no flash).

## Out of scope

- Pagination for search results (capped at 20).
- User profile pages beyond the list view.
- Privacy settings / private lists.
