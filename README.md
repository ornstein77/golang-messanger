# Messenger backend

A small messenger backend written in Go. Users register with a **nickname,
password and phone number**, log in to receive a session token, then browse
other users and exchange direct messages. This repository contains **only the
server side** (no frontend).

## Tech stack

- Go 1.25, standard-library `net/http` router (method-aware `ServeMux`)
- PostgreSQL (via `jackc/pgx/v5` + `pgxpool`)
- Versioned SQL migrations (`golang-migrate`)
- Config from env (`kelseyhightower/envconfig`)
- Structured logging (`go.uber.org/zap`)
- Request validation (`go-playground/validator/v10`)
- Swagger / OpenAPI docs (`swaggo/swag` + `swaggo/http-swagger`)
- Password hashing with bcrypt (`golang.org/x/crypto`)

## Architecture

```
cmd/app                     # entrypoint (main.go)
docs                        # generated swagger (docs.go)
internal/
  core/
    config                  # env-based configuration
    domain                  # core entities (User, Session, Message)
    errors                  # typed application errors + HTTP mapping
    logger                  # zap logger factory
    repository/postgres/
      pool                  # pgxpool builder
      pgx                   # shared Querier + pg error helpers
    transport/http/
      middleware            # recover, logging, CORS, auth
      request               # JSON decode + validation
      response              # JSON response/error envelope
      server                # router assembly + http.Server lifecycle
      types                 # response envelope + auth context helpers
  features/
    auth                    # register / login / logout / session auth
      repository | service | transport
    users                   # list users / current profile
      repository | service | transport
    messages                # send message / conversation history
      repository | service | transport
migrations                  # versioned SQL (golang-migrate)
out/
  logs                      # application log files
  pgdata                    # Postgres data (docker volume mount)
.env                        # local configuration
```

Each feature is a vertical slice with three layers that talk to each other only
through interfaces: `repository` (SQL) → `service` (business logic) →
`transport` (HTTP).

## Configuration

All configuration is read from environment variables (see `.env`). Key values:

| Variable            | Default     | Description                    |
| ------------------- | ----------- | ------------------------------ |
| `HTTP_PORT`         | `8080`      | HTTP listen port               |
| `POSTGRES_HOST`     | `localhost` | Postgres host                  |
| `POSTGRES_PORT`     | `5432`      | Postgres port                  |
| `POSTGRES_USER`     | `todoapp`   | Postgres user                  |
| `POSTGRES_PASSWORD` | `todoapp`   | Postgres password              |
| `POSTGRES_DB`       | `messenger` | Database name                  |
| `SESSION_TTL`       | `720h`      | Session token lifetime         |
| `LOG_LEVEL`         | `info`      | zap log level                  |

## Getting started

> First time only: fetch dependencies (requires network access).
>
> ```bash
> make tidy
> ```

### 1. Start Postgres, apply migrations, run the port-forwarder

```bash
make compose-up
```

This starts `todoapp-postgres`, runs the `todoapp-postgres-migrate` job (applies
`migrations/`), and exposes Postgres on `localhost:5432` through the
`alpine/socat` port-forwarder.

### 2. Run the server

```bash
make run
```

The API listens on `http://localhost:8080`.

- Health check: `GET /health`
- Swagger UI: `http://localhost:8080/swagger/index.html`

### Regenerating Swagger docs

A baseline `docs/docs.go` is committed. Regenerate it from code annotations with:

```bash
make swagger          # local swag binary
# or
make swagger-docker   # uses the swagger docker-compose service
```

## API overview

Base path: `/api/v1`. All responses use a `{ "data": ... }` / `{ "error": ... }`
envelope. Protected endpoints require an `Authorization: Bearer <token>` header.

| Method | Path             | Auth | Description                         |
| ------ | ---------------- | ---- | ----------------------------------- |
| POST   | `/auth/register` | no   | Register (nickname, password, phone)|
| POST   | `/auth/login`    | no   | Log in, returns a session token     |
| POST   | `/auth/logout`   | yes  | Revoke the current token            |
| GET    | `/users`         | yes  | List other users                    |
| GET    | `/users/me`      | yes  | Current user's profile              |
| POST   | `/messages`      | yes  | Send a message to a user            |
| GET    | `/messages?with=<uuid>` | yes | Conversation with a user       |

### Example flow

```bash
# Register
curl -s http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"alice","password":"s3cr3t-pass","phone":"+14155552671"}'

# Log in -> copy data.token
curl -s http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"alice","password":"s3cr3t-pass"}'

TOKEN=... # from the login response

# List users to find someone to chat with
curl -s http://localhost:8080/api/v1/users -H "Authorization: Bearer $TOKEN"

# Send a message
curl -s http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"recipient_id":"<other-user-uuid>","body":"Hello!"}'

# Read the conversation
curl -s "http://localhost:8080/api/v1/messages?with=<other-user-uuid>" \
  -H "Authorization: Bearer $TOKEN"
```

## Make targets

Run `make help` to see everything. The most common:

- `make tidy` – download/tidy dependencies
- `make run` – run the server
- `make build` – build `bin/app`
- `make compose-up` – start Postgres + migrations + port-forwarder
- `make compose-down` – stop the stack
- `make migrate-up` – apply migrations
- `make swagger` – regenerate Swagger docs

## Notes

- Authentication uses opaque session tokens (UUIDs) stored in the `sessions`
  table, validated on every protected request.
- Passwords are hashed with bcrypt and never returned by the API.
- Database schema changes are versioned under `migrations/` and applied with
  `golang-migrate`.
- The Postgres 18 image stores data in a version-specific layout, so the
  compose file mounts a single volume at `/var/lib/postgresql` (mapped to
  `out/pgdata`) rather than the legacy `/var/lib/postgresql/data`.
```
