# Public User Lists & Community Search — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every user's list publicly accessible at `/list/[username]`, let owners edit their own list, and add a `/community` page with a debounced user search bar.

**Architecture:** Two new backend endpoints (`GET /api/users/search` and `GET /api/users/{username}/list`) served by a new `UserHandler` backed by a thin `UserUseCase`. On the frontend, `list.vue` becomes `list/[username].vue` (public, edit-controls gated by `isOwn`), a redirect page handles `/list`, and a new `community.vue` page hosts the search UI.

**Tech Stack:** Go 1.23 + chi v5 (backend), Nuxt 3 + Nuxt UI (frontend), bun (frontend package manager)

---

## File Map

**Create:**
- `backend/internal/usecase/user.go`
- `backend/internal/usecase/user_test.go`
- `backend/internal/handler/user.go`
- `backend/internal/handler/user_test.go`
- `frontend/app/composables/useUsers.ts`
- `frontend/app/pages/list/index.vue`
- `frontend/app/pages/list/[username].vue`
- `frontend/app/pages/community.vue`

**Modify:**
- `backend/internal/port/user.go` — add `GetByUsername`/`SearchByUsername` to `UserRepository`; add `UserUseCase` interface
- `backend/internal/repository/user.go` — implement `GetByUsername` and `SearchByUsername`
- `backend/internal/usecase/auth_test.go` — add stubs to `mockUserRepository` for new interface methods
- `backend/cmd/api/main.go` — wire up `UserHandler`
- `frontend/app/app.vue` — add Community nav link

**Delete:**
- `frontend/app/pages/list.vue`

---

## Task 1: Extend port/user.go — UserRepository interface + UserUseCase interface

**Files:**
- Modify: `backend/internal/port/user.go`

- [ ] **Step 1: Replace the file content**

```go
package port

import (
	"context"
	"errors"

	"github.com/hlaclau/rate-it-api/internal/domain"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrEmailAlreadyExists    = errors.New("email already in use")
	ErrUsernameAlreadyExists = errors.New("username already taken")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUnauthorized          = errors.New("unauthorized")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	SearchByUsername(ctx context.Context, query string, limit int) ([]*domain.User, error)
}

type AuthUseCase interface {
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user *domain.User, err error)
	Refresh(ctx context.Context, refreshToken string) (accessToken string, userID string, err error)
	ValidateAccessToken(tokenString string) (userID string, err error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	Logout(ctx context.Context, userID string) error
}

type UserUseCase interface {
	SearchUsers(ctx context.Context, query string) ([]*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}
```

- [ ] **Step 2: Update `mockUserRepository` in auth_test.go to satisfy the extended interface**

Open `backend/internal/usecase/auth_test.go`. The existing `mockUserRepository` struct has three fields: `createFn`, `getByEmailFn`, `getByIDFn`. Add two new nil-safe stubs after the existing three:

```go
type mockUserRepository struct {
	createFn            func(ctx context.Context, user *domain.User) error
	getByEmailFn        func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn           func(ctx context.Context, id string) (*domain.User, error)
	getByUsernameFn     func(ctx context.Context, username string) (*domain.User, error)
	searchByUsernameFn  func(ctx context.Context, query string, limit int) ([]*domain.User, error)
}

// ... keep existing Create, GetByEmail, GetByID methods unchanged ...

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, port.ErrUserNotFound
}

func (m *mockUserRepository) SearchByUsername(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	if m.searchByUsernameFn != nil {
		return m.searchByUsernameFn(ctx, query, limit)
	}
	return nil, nil
}
```

- [ ] **Step 3: Verify existing tests still compile and pass**

```bash
cd backend && go test ./internal/usecase/... -v
```

Expected: all existing tests PASS, no compile errors.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/port/user.go backend/internal/usecase/auth_test.go
git commit -m "feat: extend UserRepository interface with GetByUsername and SearchByUsername"
```

---

## Task 2: Implement new repository methods

**Files:**
- Modify: `backend/internal/repository/user.go`

- [ ] **Step 1: Add `GetByUsername` method to `UserRepository`**

Append after the existing `GetByID` method in `backend/internal/repository/user.go`:

```go
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	const q = `SELECT id, username, email, password_hash, bio, avatar_url, created_at FROM users WHERE username = $1`

	var u domain.User
	err := r.db.GetContext(ctx, &u, q, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, port.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) SearchByUsername(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	const q = `
		SELECT id, username, email, password_hash, bio, avatar_url, created_at
		FROM users
		WHERE username ILIKE '%' || $1 || '%'
		ORDER BY username
		LIMIT $2`

	var users []*domain.User
	err := r.db.SelectContext(ctx, &users, q, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search users by username: %w", err)
	}

	return users, nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/user.go
git commit -m "feat: implement GetByUsername and SearchByUsername in UserRepository"
```

---

## Task 3: Create UserUseCase

**Files:**
- Create: `backend/internal/usecase/user.go`
- Create: `backend/internal/usecase/user_test.go`

- [ ] **Step 1: Write the failing tests first**

Create `backend/internal/usecase/user_test.go`:

```go
package usecase_test

import (
	"context"
	"testing"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
	"github.com/hlaclau/rate-it-api/internal/usecase"
)

func TestUserUseCase_SearchUsers_ReturnsResults(t *testing.T) {
	repo := &mockUserRepository{
		searchByUsernameFn: func(_ context.Context, query string, limit int) ([]*domain.User, error) {
			if query != "ali" {
				t.Errorf("unexpected query: %q", query)
			}
			if limit != 20 {
				t.Errorf("unexpected limit: %d", limit)
			}
			return []*domain.User{{ID: "u1", Username: "alice"}}, nil
		},
	}

	uc := usecase.NewUserUseCase(repo)
	users, err := uc.SearchUsers(context.Background(), "ali")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 || users[0].Username != "alice" {
		t.Errorf("unexpected users: %+v", users)
	}
}

func TestUserUseCase_GetByUsername_Found(t *testing.T) {
	repo := &mockUserRepository{
		getByUsernameFn: func(_ context.Context, username string) (*domain.User, error) {
			if username != "alice" {
				t.Errorf("unexpected username: %q", username)
			}
			return &domain.User{ID: "u1", Username: "alice"}, nil
		},
	}

	uc := usecase.NewUserUseCase(repo)
	user, err := uc.GetByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("unexpected user ID: %s", user.ID)
	}
}

func TestUserUseCase_GetByUsername_NotFound(t *testing.T) {
	repo := &mockUserRepository{
		getByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, port.ErrUserNotFound
		},
	}

	uc := usecase.NewUserUseCase(repo)
	_, err := uc.GetByUsername(context.Background(), "ghost")
	if err != port.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
```

- [ ] **Step 2: Run to confirm it fails (UserUseCase not yet defined)**

```bash
cd backend && go test ./internal/usecase/... -run TestUserUseCase -v
```

Expected: compile error — `usecase.NewUserUseCase undefined`.

- [ ] **Step 3: Create `backend/internal/usecase/user.go`**

```go
package usecase

import (
	"context"

	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/port"
)

type UserUseCase struct {
	userRepo port.UserRepository
}

func NewUserUseCase(userRepo port.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (uc *UserUseCase) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	return uc.userRepo.SearchByUsername(ctx, query, 20)
}

func (uc *UserUseCase) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return uc.userRepo.GetByUsername(ctx, username)
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd backend && go test ./internal/usecase/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/usecase/user.go backend/internal/usecase/user_test.go
git commit -m "feat: add UserUseCase with SearchUsers and GetByUsername"
```

---

## Task 4: Create UserHandler

**Files:**
- Create: `backend/internal/handler/user.go`
- Create: `backend/internal/handler/user_test.go`

- [ ] **Step 1: Write the failing tests first**

Create `backend/internal/handler/user_test.go`:

```go
package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/domain"
	"github.com/hlaclau/rate-it-api/internal/handler"
	"github.com/hlaclau/rate-it-api/internal/port"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockUserUseCase struct {
	searchUsersFn  func(ctx context.Context, query string) ([]*domain.User, error)
	getByUsernameFn func(ctx context.Context, username string) (*domain.User, error)
}

func (m *mockUserUseCase) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	return m.searchUsersFn(ctx, query)
}
func (m *mockUserUseCase) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return m.getByUsernameFn(ctx, username)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newUserRouter(userUC port.UserUseCase, listUC port.ListUseCase) http.Handler {
	r := chi.NewRouter()
	h := handler.NewUserHandler(userUC, listUC)
	h.Routes(r)
	return r
}

// ── GET /users/search ─────────────────────────────────────────────────────────

func TestUserSearch_ShortQuery_ReturnsEmpty(t *testing.T) {
	userUC := &mockUserUseCase{}
	router := newUserRouter(userUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=a", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty array, got %d items", len(resp))
	}
}

func TestUserSearch_ValidQuery_ReturnsUsers(t *testing.T) {
	userUC := &mockUserUseCase{
		searchUsersFn: func(_ context.Context, query string) ([]*domain.User, error) {
			return []*domain.User{{ID: "u1", Username: "alice"}}, nil
		},
	}
	router := newUserRouter(userUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=ali", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp))
	}
	if resp[0]["username"] != "alice" {
		t.Errorf("unexpected username: %v", resp[0]["username"])
	}
	// email must not be present in response
	if _, ok := resp[0]["email"]; ok {
		t.Error("email should not be in search response")
	}
}

func TestUserSearch_MissingQuery_ReturnsEmpty(t *testing.T) {
	router := newUserRouter(&mockUserUseCase{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty array, got %d items", len(resp))
	}
}

// ── GET /users/{username}/list ────────────────────────────────────────────────

func TestGetUserList_UserNotFound(t *testing.T) {
	userUC := &mockUserUseCase{
		getByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, port.ErrUserNotFound
		},
	}
	router := newUserRouter(userUC, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/ghost/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetUserList_Success(t *testing.T) {
	userUC := &mockUserUseCase{
		getByUsernameFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: "u1", Username: "alice"}, nil
		},
	}
	listUC := &mockListUseCase{
		getListFn: func(_ context.Context, userID string) ([]port.ListEntry, error) {
			if userID != "u1" {
				return nil, nil
			}
			return []port.ListEntry{sampleListEntry()}, nil
		},
	}
	router := newUserRouter(userUC, listUC)

	req := httptest.NewRequest(http.MethodGet, "/users/alice/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp []port.ListEntry
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp))
	}
}
```

- [ ] **Step 2: Run to confirm it fails**

```bash
cd backend && go test ./internal/handler/... -run TestUser -v
```

Expected: compile error — `handler.NewUserHandler undefined`.

- [ ] **Step 3: Create `backend/internal/handler/user.go`**

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hlaclau/rate-it-api/internal/port"
)

type UserHandler struct {
	uc     port.UserUseCase
	listUC port.ListUseCase
}

func NewUserHandler(uc port.UserUseCase, listUC port.ListUseCase) *UserHandler {
	return &UserHandler{uc: uc, listUC: listUC}
}

func (h *UserHandler) Routes(r chi.Router) {
	r.Get("/users/search", h.Search)
	r.Get("/users/{username}/list", h.GetList)
}

type userSummary struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		h.writeJSON(w, http.StatusOK, []userSummary{})
		return
	}

	users, err := h.uc.SearchUsers(r.Context(), q)
	if err != nil {
		slog.Error("search users", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	summaries := make([]userSummary, len(users))
	for i, u := range users {
		summaries[i] = userSummary{ID: u.ID, Username: u.Username, AvatarURL: u.AvatarURL}
	}
	h.writeJSON(w, http.StatusOK, summaries)
}

func (h *UserHandler) GetList(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.uc.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, port.ErrUserNotFound) {
			h.writeJSON(w, http.StatusNotFound, map[string]string{"message": "user not found"})
			return
		}
		slog.Error("get user by username", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	entries, err := h.listUC.GetList(r.Context(), user.ID)
	if err != nil {
		slog.Error("get user list", "error", err)
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		return
	}

	if entries == nil {
		entries = []port.ListEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd backend && go test ./internal/handler/... -v
```

Expected: all tests including new ones PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/user.go backend/internal/handler/user_test.go
git commit -m "feat: add UserHandler with search and public list endpoints"
```

---

## Task 5: Wire UserHandler in main.go

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Add UserUseCase and UserHandler construction**

In `main.go`, after the `listHandler` block, add:

```go
userHandler := handler.NewUserHandler(
    usecase.NewUserUseCase(userRepo),
    usecase.NewListUseCase(
        repository.NewUserMediaRepository(db),
        mediaRepo,
        tmdbFetcher,
    ),
)
```

And inside `r.Route("/api", ...)`, add:

```go
userHandler.Routes(r)
```

The final `r.Route` block should look like:

```go
r.Route("/api", func(r chi.Router) {
    authHandler.Routes(r)
    mediaHandler.Routes(r)
    listHandler.Routes(r)
    userHandler.Routes(r)
})
```

- [ ] **Step 2: Build to verify**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Run all tests**

```bash
cd backend && go test ./... -v
```

Expected: all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/api/main.go
git commit -m "feat: wire UserHandler into API router"
```

---

## Task 6: Create useUsers composable (frontend)

**Files:**
- Create: `frontend/app/composables/useUsers.ts`

- [ ] **Step 1: Create the composable**

```typescript
import type { ListEntry } from '~/composables/useList'

export interface UserSummary {
  id: string
  username: string
  avatar_url: string | null
}

export const useUsers = () => {
  const config = useRuntimeConfig()

  const publicFetch = $fetch.create({
    baseURL: `${config.public.apiBase}/api/`,
  })

  const searchUsers = async (query: string): Promise<UserSummary[]> => {
    if (query.length < 2) return []
    try {
      return await publicFetch<UserSummary[]>(
        `users/search?q=${encodeURIComponent(query)}`
      )
    } catch {
      return []
    }
  }

  const fetchUserList = async (
    username: string
  ): Promise<{ entries: ListEntry[]; notFound: boolean }> => {
    try {
      const entries = await publicFetch<ListEntry[]>(
        `users/${encodeURIComponent(username)}/list`
      )
      return { entries: entries ?? [], notFound: false }
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status
      if (status === 404) {
        return { entries: [], notFound: true }
      }
      throw err
    }
  }

  return { searchUsers, fetchUserList }
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/composables/useUsers.ts
git commit -m "feat: add useUsers composable for search and public list fetch"
```

---

## Task 7: Create list/index.vue redirect and list/[username].vue

**Files:**
- Create: `frontend/app/pages/list/index.vue`
- Create: `frontend/app/pages/list/[username].vue`
- Delete: `frontend/app/pages/list.vue`

- [ ] **Step 1: Delete the old list.vue**

```bash
rm frontend/app/pages/list.vue
```

- [ ] **Step 2: Create `frontend/app/pages/list/index.vue`**

```vue
<script setup lang="ts">
const { user, authInitialized } = useAuth()

watchEffect(() => {
  if (!authInitialized.value) return
  if (user.value) {
    navigateTo(`/list/${user.value.username}`, { replace: true })
  } else {
    navigateTo('/', { replace: true })
  }
})
</script>

<template>
  <div />
</template>
```

- [ ] **Step 3: Create `frontend/app/pages/list/[username].vue`**

```vue
<script setup lang="ts">
import type { ListEntry } from '~/composables/useList'

const route = useRoute()
const username = route.params.username as string

const { user } = useAuth()
const { fetchUserList } = useUsers()
const { remove: removeFromList } = useList()

const isOwn = computed(() => !!user.value && user.value.username === username)

const entries = ref<ListEntry[]>([])
const loading = ref(false)
const notFound = ref(false)

const activeFilter = ref<'all' | 'watched' | 'plan_to_watch'>('all')

type SortValue =
  | 'added_at_desc'
  | 'added_at_asc'
  | 'title_asc'
  | 'title_desc'
  | 'rating_desc'
  | 'rating_asc'
  | 'release_year_desc'
  | 'release_year_asc'
const sortValue = ref<SortValue>('added_at_desc')

const sortOptions: { label: string; value: SortValue }[] = [
  { label: 'Date added (newest)', value: 'added_at_desc' },
  { label: 'Date added (oldest)', value: 'added_at_asc' },
  { label: 'Title A → Z', value: 'title_asc' },
  { label: 'Title Z → A', value: 'title_desc' },
  { label: 'Rating (highest)', value: 'rating_desc' },
  { label: 'Rating (lowest)', value: 'rating_asc' },
  { label: 'Year (newest)', value: 'release_year_desc' },
  { label: 'Year (oldest)', value: 'release_year_asc' },
]

const filteredList = computed<ListEntry[]>(() => {
  const items =
    activeFilter.value === 'all'
      ? entries.value
      : entries.value.filter((e) => e.status === activeFilter.value)

  const lastUnderscore = sortValue.value.lastIndexOf('_')
  const key = sortValue.value.slice(0, lastUnderscore)
  const dir = sortValue.value.slice(lastUnderscore + 1)

  return [...items].sort((a, b) => {
    if (key === 'title') {
      return dir === 'asc'
        ? a.title.localeCompare(b.title)
        : b.title.localeCompare(a.title)
    }
    if (key === 'rating') {
      const ra = a.rating ?? (dir === 'asc' ? Infinity : -Infinity)
      const rb = b.rating ?? (dir === 'asc' ? Infinity : -Infinity)
      return dir === 'asc' ? ra - rb : rb - ra
    }
    if (key === 'release_year') {
      const ya = a.release_year ?? (dir === 'asc' ? Infinity : -Infinity)
      const yb = b.release_year ?? (dir === 'asc' ? Infinity : -Infinity)
      return dir === 'asc' ? ya - yb : yb - ya
    }
    const da = new Date(a.added_at).getTime()
    const db = new Date(b.added_at).getTime()
    return dir === 'asc' ? da - db : db - da
  })
})

onMounted(async () => {
  loading.value = true
  try {
    const result = await fetchUserList(username)
    notFound.value = result.notFound
    entries.value = result.entries
  } finally {
    loading.value = false
  }
})

const posterUrl = (path: string | null) =>
  path ? `https://image.tmdb.org/t/p/w300${path}` : null

const pendingRemove = ref<ListEntry | null>(null)
const removing = ref(false)

const confirmRemove = async () => {
  if (!pendingRemove.value) return
  removing.value = true
  try {
    await removeFromList(pendingRemove.value.media_id)
    entries.value = entries.value.filter(
      (e) => e.media_id !== pendingRemove.value!.media_id
    )
    pendingRemove.value = null
  } finally {
    removing.value = false
  }
}

useSeoMeta({
  title: computed(() =>
    isOwn.value ? 'My List — Rate It' : `${username}'s List — Rate It`
  ),
})
</script>

<template>
  <UContainer class="py-10">
    <!-- User not found -->
    <div
      v-if="notFound"
      class="flex flex-col items-center gap-4 py-24 text-center"
    >
      <UIcon name="i-lucide-user-x" class="size-16 text-muted" />
      <div>
        <p class="font-semibold text-lg">User not found</p>
        <p class="text-muted text-sm mt-1">
          No account with username "{{ username }}".
        </p>
      </div>
      <UButton to="/community" variant="soft" leading-icon="i-lucide-users">
        Find users
      </UButton>
    </div>

    <template v-else>
      <div class="flex flex-wrap items-center justify-between gap-4 mb-8">
        <h1 class="text-3xl font-bold">
          {{ isOwn ? 'My List' : `${username}'s List` }}
        </h1>
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex gap-2">
            <UButton
              v-for="f in [
                { label: 'All', value: 'all' },
                { label: 'Watched', value: 'watched' },
                { label: 'Plan to watch', value: 'plan_to_watch' },
              ]"
              :key="f.value"
              :variant="activeFilter === f.value ? 'solid' : 'ghost'"
              color="neutral"
              size="sm"
              @click="
                activeFilter = f.value as 'all' | 'watched' | 'plan_to_watch'
              "
            >
              {{ f.label }}
            </UButton>
          </div>
          <USelect
            v-model="sortValue"
            :items="sortOptions"
            value-key="value"
            label-key="label"
            size="sm"
            class="w-52"
          />
        </div>
      </div>

      <!-- Loading -->
      <div
        v-if="loading"
        class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
      >
        <USkeleton v-for="i in 10" :key="i" class="aspect-[2/3] rounded-xl" />
      </div>

      <!-- Empty state -->
      <div
        v-else-if="filteredList.length === 0"
        class="flex flex-col items-center gap-4 py-24 text-center"
      >
        <UIcon name="i-lucide-bookmark" class="size-16 text-muted" />
        <div>
          <p class="font-semibold text-lg">Nothing here yet</p>
          <p class="text-muted text-sm mt-1">
            {{
              activeFilter === 'all'
                ? isOwn
                  ? 'Add movies or series to your list from their detail page.'
                  : `${username} hasn't added anything yet.`
                : 'No entries with this status.'
            }}
          </p>
        </div>
        <UButton
          v-if="isOwn"
          to="/discover"
          variant="soft"
          leading-icon="i-lucide-compass"
        >
          Browse media
        </UButton>
      </div>

      <!-- List grid -->
      <div
        v-else
        class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
      >
        <div
          v-for="entry in filteredList"
          :key="entry.media_id"
          class="group relative rounded-xl overflow-hidden bg-elevated ring-1 ring-default hover:ring-primary transition-all"
        >
          <NuxtLink
            :to="
              entry.type === 'series'
                ? `/series/${entry.external_id}`
                : `/movie/${entry.external_id}`
            "
          >
            <img
              v-if="posterUrl(entry.poster_path)"
              :src="posterUrl(entry.poster_path)!"
              :alt="entry.title"
              class="w-full aspect-[2/3] object-cover"
            />
            <div
              v-else
              class="w-full aspect-[2/3] bg-muted flex items-center justify-center"
            >
              <UIcon name="i-lucide-film" class="size-10 text-muted" />
            </div>
          </NuxtLink>

          <div class="p-3 space-y-1.5">
            <p class="font-semibold text-sm leading-tight line-clamp-2">
              {{ entry.title }}
            </p>
            <div class="flex items-center justify-between gap-2">
              <UBadge
                :color="entry.status === 'watched' ? 'success' : 'info'"
                size="xs"
              >
                {{ entry.status === 'watched' ? 'Watched' : 'Plan to watch' }}
              </UBadge>
              <span
                v-if="entry.rating"
                class="flex items-center gap-0.5 text-xs font-medium shrink-0"
              >
                <UIcon
                  name="i-lucide-star"
                  class="size-3 text-yellow-400 fill-yellow-400"
                />
                {{ entry.rating }}
              </span>
            </div>
            <p
              v-if="entry.review"
              class="text-xs text-muted line-clamp-2 italic"
            >
              "{{ entry.review }}"
            </p>
          </div>

          <!-- Remove button (own list only) -->
          <UButton
            v-if="isOwn"
            icon="i-lucide-x"
            size="xs"
            color="neutral"
            variant="solid"
            class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
            @click.prevent="pendingRemove = entry"
          />
        </div>
      </div>

      <!-- Remove confirmation modal (own list only) -->
      <UModal
        v-if="isOwn"
        :open="!!pendingRemove"
        title="Remove from list"
        :description="
          pendingRemove
            ? `Remove &quot;${pendingRemove.title}&quot; from your list?`
            : ''
        "
        @update:open="
          (v) => {
            if (!v) pendingRemove = null
          }
        "
      >
        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              @click="pendingRemove = null"
              >Cancel</UButton
            >
            <UButton
              color="error"
              :loading="removing"
              leading-icon="i-lucide-trash-2"
              @click="confirmRemove"
            >
              Remove
            </UButton>
          </div>
        </template>
      </UModal>
    </template>
  </UContainer>
</template>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/app/pages/list/ && git rm frontend/app/pages/list.vue
git commit -m "feat: replace list.vue with dynamic list/[username].vue and redirect index"
```

---

## Task 8: Create community.vue

**Files:**
- Create: `frontend/app/pages/community.vue`

- [ ] **Step 1: Create the page**

```vue
<script setup lang="ts">
import type { UserSummary } from '~/composables/useUsers'

const { searchUsers } = useUsers()

const query = ref('')
const results = ref<UserSummary[]>([])
const loading = ref(false)
const searched = ref(false)

const avatarUrl = (url: string | null) =>
  url ?? null

let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(query, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  if (val.length < 2) {
    results.value = []
    searched.value = false
    return
  }
  debounceTimer = setTimeout(async () => {
    loading.value = true
    searched.value = true
    try {
      results.value = await searchUsers(val)
    } finally {
      loading.value = false
    }
  }, 300)
})

useSeoMeta({ title: 'Community — Rate It' })
</script>

<template>
  <UContainer class="py-10 max-w-2xl">
    <h1 class="text-3xl font-bold mb-2">Community</h1>
    <p class="text-muted mb-8">Find users and explore their lists.</p>

    <UInput
      v-model="query"
      placeholder="Search by username…"
      leading-icon="i-lucide-search"
      size="lg"
      class="mb-8"
    />

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <USkeleton v-for="i in 4" :key="i" class="h-14 rounded-xl" />
    </div>

    <!-- Results -->
    <div v-else-if="results.length > 0" class="space-y-2">
      <NuxtLink
        v-for="u in results"
        :key="u.id"
        :to="`/list/${u.username}`"
        class="flex items-center gap-4 p-4 rounded-xl bg-elevated ring-1 ring-default hover:ring-primary transition-all"
      >
        <UAvatar
          :src="avatarUrl(u.avatar_url) ?? undefined"
          :alt="u.username"
          size="md"
        />
        <span class="font-medium">{{ u.username }}</span>
        <UIcon
          name="i-lucide-chevron-right"
          class="ml-auto size-4 text-muted"
        />
      </NuxtLink>
    </div>

    <!-- No results -->
    <div
      v-else-if="searched && query.length >= 2"
      class="flex flex-col items-center gap-3 py-16 text-center"
    >
      <UIcon name="i-lucide-user-search" class="size-12 text-muted" />
      <p class="font-semibold">No users found</p>
      <p class="text-muted text-sm">Try a different username.</p>
    </div>
  </UContainer>
</template>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/pages/community.vue
git commit -m "feat: add community page with user search"
```

---

## Task 9: Update app.vue — add Community nav link

**Files:**
- Modify: `frontend/app/app.vue`

- [ ] **Step 1: Add Community to the nav and update footer**

In `app.vue`, update two places:

**Nav bar** — add Community link after the Discover dropdown (inside the `<nav>` tag):

```html
<nav class="hidden md:flex items-center gap-1">
  <UDropdownMenu :items="discoverItems">
    <!-- ... existing dropdown ... -->
  </UDropdownMenu>
  <NuxtLink
    to="/community"
    class="flex items-center gap-1 px-3 py-1.5 text-sm font-medium rounded-md transition-colors"
    :class="
      $route.path === '/community'
        ? 'text-primary bg-primary/10'
        : 'text-muted hover:text-default hover:bg-elevated'
    "
  >
    Community
  </NuxtLink>
</nav>
```

**Footer** — add Community link next to Discover:

```html
<NuxtLink
  to="/community"
  class="text-sm font-medium text-muted hover:text-primary transition-colors"
>
  Community
</NuxtLink>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/app.vue
git commit -m "feat: add Community link to nav and footer"
```

---

## Task 10: Smoke test

- [ ] **Step 1: Start backend**

```bash
cd backend && go run ./cmd/api
```

- [ ] **Step 2: Start frontend**

```bash
cd frontend && bun run dev
```

- [ ] **Step 3: Verify these flows work**

1. Visit `/list` while logged in → redirects to `/list/[your-username]`, edit controls visible.
2. Visit `/list` while logged out → redirects to `/`.
3. Visit `/list/[other-username]` → read-only view, no remove buttons.
4. Visit `/list/nonexistent` → "User not found" state.
5. Visit `/community`, type 1 char → no results. Type 2+ chars → user results appear, linking to `/list/[username]`.
6. `GET /api/users/search?q=a` → `[]`. `GET /api/users/search?q=al` → results without `email` field.
7. `GET /api/users/[username]/list` → returns list entries publicly (no cookie needed).

- [ ] **Step 4: Final commit if any fixes were needed**

```bash
git add -p
git commit -m "fix: smoke test corrections"
```
