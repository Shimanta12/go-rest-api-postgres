# go-rest-api-postgres

RESTful API built from scratch in Go, backed by PostgreSQL. This project demonstrates clean architecture, idiomatic Go patterns, and real-world backend engineering practices — without relying on heavyweight frameworks.

---

## Tech Stack

- **Language:** Go 1.25.7
- **Database:** PostgreSQL
- **Driver:** `jackc/pgx` via `database/sql`
- **Config:** `joho/godotenv`
- **Standard library only** for HTTP (`net/http`)

---

## Project Structure

```
go-rest-api-postgres/
├── cmd/
│   └── api/
│       └── main.go           # Entry point — wires dependencies, starts server
├── internal/
│   ├── handler/
│   │   └── user.go           # HTTP handlers — request parsing, response writing
│   ├── store/
│   │   └── user.go           # Database layer — all SQL queries live here
│   ├── model/
│   │   └── user.go           # Data models — User, UserRequest, UpdateUserRequest
│   └── middleware/
│       └── logger.go         # Logger middleware — logs method, path, duration
├── .env.example              # Environment variable template
├── .gitignore
└── go.mod
```

The project follows a strict **layered architecture** where each package has a single responsibility:

- `handler` knows about HTTP. It knows nothing about SQL.
- `store` knows about the database. It knows nothing about HTTP.
- `model` is pure data. It has no logic, no dependencies.
- `middleware` wraps handlers to add cross-cutting behaviour like logging.

Dependencies flow in one direction only — `handler` → `store` → database, with `model` shared between them. This means the database can be swapped without touching a single handler, and the HTTP layer can change without touching a single query.

The `internal/` directory is a Go convention that marks these packages as private to this module — they cannot be imported by external code.

---

## API Endpoints

Base URL: `http://localhost:8080`

| Method | Endpoint | Description | Request Body | Success Response |
|--------|----------|-------------|--------------|-----------------|
| `GET` | `/users` | List all users | — | `200 OK` |
| `POST` | `/users` | Create a new user | `{ name, email }` | `201 Created` |
| `GET` | `/users/{id}` | Get a user by ID | — | `200 OK` |
| `PATCH` | `/users/{id}` | Update a user | `{ name?, email? }` | `200 OK` |
| `DELETE` | `/users/{id}` | Delete a user | — | `204 No Content` |

### Request & Response Shapes

**Create user** `POST /users`
```json
// Request
{ "name": "Alice", "email": "alice@example.com" }

// Response 201
{ "id": 1, "name": "Alice", "email": "alice@example.com", "created_at": "2026-01-01T00:00:00Z" }
```

**Update user** `PATCH /users/{id}`
```json
// Request — all fields optional
{ "name": "Alice Smith" }

// Response 200
{ "id": 1, "name": "Alice Smith", "email": "alice@example.com", "created_at": "2026-01-01T00:00:00Z" }
```

### Error Responses

All errors return a consistent JSON shape:
```json
{ "error": "description of what went wrong" }
```

| Status | Meaning |
|--------|---------|
| `400 Bad Request` | Invalid JSON or missing required field |
| `404 Not Found` | User with that ID does not exist |
| `409 Conflict` | Email address is already in use |
| `422 Unprocessable Entity` | Field was sent but value is invalid |
| `500 Internal Server Error` | Unexpected server error |

---

## Key Decisions

**`database/sql` over an ORM** — Raw SQL keeps queries explicit and predictable. There's no magic, no hidden N+1 queries, and no fighting an abstraction layer when something goes wrong.

**Dependency injection over globals** — The database store is passed into handlers at startup rather than accessed as a global variable. This makes the code easier to test and reason about.

**Graceful shutdown** — The server listens for `SIGINT`/`SIGTERM` and gives in-flight requests up to 10 seconds to complete before stopping. This prevents broken responses when the server is restarted.

**Pointer fields for PATCH** — `UpdateUserRequest` uses `*string` fields so the handler can distinguish between "field not sent" (`nil`) and "field sent as empty string" (`""`). Only non-nil fields are updated in the database.


---

## Getting Started

### Prerequisites
- Go 1.22+
- PostgreSQL

### 1. Clone the repository
```bash
git clone https://github.com/Shimanta12/go-rest-api-postgres.git
cd go-rest-api-postgres
```

### 2. Set up environment variables
```bash
cp .env.example .env
# Edit .env and fill in your DATABASE_URL
```

### 3. Create the database and table
```bash
sudo -u postgres psql
```
```sql
CREATE DATABASE myapi;
\c myapi
CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 4. Run the server
```bash
go run cmd/api/main.go
```

The server starts on `http://localhost:8080` by default.
