# Ticket System

A backend service where a user can register, log in, create tickets, view only their own tickets, and update the status of their own tickets.

## Features

- Email/password registration and login
- JWT-based authentication (`Authorization: Bearer <token>`)
- Passwords hashed with PBKDF2-HMAC-SHA256, unique salt per user
- Ownership-based authorization on every ticket endpoint (no IDOR)
- Centralized ticket status state machine: `open -> in_progress -> closed`
- Closed tickets cannot be reopened or moved to any other state

## Tech Stack

- Go 1.22, standard library only (`net/http` with Go 1.22 method+path routing)
- No third-party dependencies

## Assumptions

- **Storage:** the assignment explicitly allows in-memory storage as a valid option ("Use in-memory storage, SQLite, PostgreSQL, or any simple persistent store"). This implementation uses a thread-safe in-memory store (`internal/repository`). Data does not persist across restarts. Swapping in SQLite/Postgres later only requires implementing the same store interface used by the handlers.
- **JWT/password hashing:** implemented with the Go standard library only (`crypto/hmac`, `crypto/sha256`) instead of a third-party JWT or bcrypt library, so the project has zero external dependencies and builds reliably anywhere without network access to a module proxy. The JWT implementation follows the standard HS256 JWT structure (header.payload.signature). Password hashing uses PBKDF2 with 100,000 iterations and a random 16-byte salt per user.
- Email is case-normalized (lowercased) before storage/lookup to avoid duplicate accounts differing only by case.
- Password minimum length of 8 characters is enforced at registration.

## Project Structure

```
ticket-system/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/            # JWT + password hashing
│   ├── handlers/        # HTTP handlers + router
│   ├── middleware/       # JWT auth middleware
│   ├── models/           # User / Ticket types
│   ├── repository/       # In-memory data store
│   └── service/          # Status transition rules
├── Dockerfile
├── .dockerignore
├── .env.example
├── go.mod
└── README.md
```

## Environment Variables

| Variable     | Required | Default | Description                         |
|--------------|----------|---------|--------------------------------------|
| `PORT`       | No       | `8080`  | Port the server listens on           |
| `JWT_SECRET` | Yes      | —       | Secret used to sign/verify JWTs      |

See `.env.example`.

## Local Setup

```bash
export JWT_SECRET=some-long-random-secret
go run ./cmd/server
```

The server listens on `0.0.0.0:8080`.

## Running Tests

```bash
go test ./...
```

To run with verbose output:

```bash
go test ./... -v
```

## Docker

Build:

```bash
docker build -t ticket-system .
```

Run:

```bash
docker run -p 8080:8080 -e JWT_SECRET=some-long-random-secret ticket-system
```

Verify:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## API Documentation

All request/response bodies are JSON. All timestamps are RFC3339.

### `GET /health`

- **Auth:** none
- **Response:** `200 OK`
```json
{"status": "ok"}
```

### `POST /auth/register`

- **Auth:** none
- **Request body:**
```json
{"email": "alice@example.com", "password": "password123"}
```
- **Success response:** `201 Created`
```json
{"id": 1, "email": "alice@example.com"}
```
- **Errors:**
  - `400 Bad Request` — missing/invalid email or password under 8 characters
  - `409 Conflict` — email already registered

### `POST /auth/login`

- **Auth:** none
- **Request body:**
```json
{"email": "alice@example.com", "password": "password123"}
```
- **Success response:** `200 OK`
```json
{"token": "<jwt>"}
```
- **Errors:**
  - `400 Bad Request` — malformed request body
  - `401 Unauthorized` — unknown user or wrong password

### `POST /tickets`

- **Auth:** required (`Authorization: Bearer <token>`)
- **Request body:**
```json
{"title": "Printer not working", "description": "Office printer on 2nd floor"}
```
- **Success response:** `201 Created`
```json
{
  "id": 1,
  "user_id": 1,
  "title": "Printer not working",
  "description": "Office printer on 2nd floor",
  "status": "open",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```
- **Errors:**
  - `400 Bad Request` — missing title or malformed body
  - `401 Unauthorized` — missing/invalid/expired token

### `GET /tickets`

- **Auth:** required
- **Response:** `200 OK` — array of the authenticated user's own tickets only
```json
[{"id": 1, "user_id": 1, "title": "...", "status": "open", "...": "..."}]
```

### `GET /tickets/{id}`

- **Auth:** required
- **Response:** `200 OK` — the ticket, only if owned by the authenticated user
- **Errors:**
  - `400 Bad Request` — malformed ID
  - `401 Unauthorized` — missing/invalid/expired token
  - `404 Not Found` — ticket does not exist, or belongs to another user

### `PATCH /tickets/{id}/status`

- **Auth:** required
- **Request body:**
```json
{"status": "in_progress"}
```
- **Success response:** `200 OK` — updated ticket
- **Errors:**
  - `400 Bad Request` — invalid status value or invalid transition
  - `401 Unauthorized` — missing/invalid/expired token
  - `404 Not Found` — ticket does not exist, or belongs to another user

## Authentication

Login returns a JWT (HS256) containing the user's ID and email, valid for 24 hours. Every protected endpoint requires:

```
Authorization: Bearer <token>
```

Requests with a missing, malformed, invalid, or expired token receive `401 Unauthorized`.

## Ticket Status Flow

```
open -> in_progress -> closed
```

Only forward, single-step transitions are allowed. Any other transition (including `closed -> open`, `closed -> in_progress`, `open -> closed`, or a reverse move) is rejected with `400 Bad Request`. A closed ticket is terminal.

## Ownership Behavior

The authenticated user's ID always comes from the verified JWT, never from client input. `GET /tickets` only returns the caller's own tickets. `GET /tickets/{id}` and `PATCH /tickets/{id}/status` return `404 Not Found` (not `403`) when the ticket exists but belongs to another user, to avoid confirming the existence of tickets the caller does not own.

## Deployment

**Status: NOT VERIFIED.** No deployment has been performed in this environment — the assignment requires deploying to a live host, and this sandbox has no access to any deployment provider or credentials. Below are the exact remaining steps to deploy on a free platform such as Render.

1. Push this repository to GitHub.
2. On Render, create a new **Web Service** from the repository.
3. Set the runtime to **Docker** (Render will detect the `Dockerfile`).
4. Set environment variable `JWT_SECRET` to a long random value in the Render dashboard.
5. Deploy. Render will build the Docker image and run the container, exposing it on Render's assigned public URL (Render maps the container's `8080` automatically).
6. Once live, verify:
   ```bash
   curl https://<your-render-url>/health
   ```
   should return `{"status":"ok"}`.
7. Update this README with the actual deployed URL and public health URL once confirmed.

**Deployed application URL:** _not yet available — see steps above_
**Public health URL:** _not yet available — see steps above_
