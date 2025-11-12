## User Management API

Golang REST API for managing company users, their files, and lifecycle events. The implementation follows a clean, interface-driven architecture that decouples HTTP transport, business logic, storage, and event delivery.

### Requirements Coverage

- **Endpoints (8 total)**: list/get/create/update/delete users plus list/add/delete user files at `/api/v1/users/**`.
- **Business rules**: `email` uniqueness enforced at the service + DB layer, `age > 18` validation on create/update.
- **File management**: files persist alongside users via dedicated repository methods.
- **Event publishing**: every create/update/delete emits `UserCreated/UserUpdated/UserDeleted` to RabbitMQ; events include IDs + payloads.
- **Consumer**: `cmd/consumer` is a simple, pluggable RabbitMQ listener that prints received events.
- **Auth & logging (bonus)**: JWT login at `/auth/login`, Gin middleware enforces tokens and logs every request with Logrus.
- **Testing (bonus)**: business-rule tests plus full end-to-end API tests run against PostgreSQL via Testcontainers—no in-memory stores.
- **Dockerization (bonus)**: Dockerfile and Compose stand up the API, PostgreSQL, and RabbitMQ.

### Architecture Overview

```
cmd/
 ├─ api        # HTTP server bootstrap
 └─ consumer   # RabbitMQ event subscriber (prints events)
internal/
 ├─ config     # env loading
 ├─ domain     # entities + errors
 ├─ event      # publisher/consumer interfaces + RabbitMQ impl
 ├─ repository # storage contracts
 ├─ service    # business logic (validation, events)
 ├─ storage    # Postgres GORM repository
 ├─ transport  # Gin router, handlers, middleware
 ├─ e2e        # end-to-end HTTP tests
 └─ testutil   # Postgres test helpers (Testcontainers)
```

The service layer depends on repository interfaces and an event publisher, so swapping storage (e.g., Mongo) or messaging (e.g., Kafka) requires only new adapters.

### Configuration

Environment variables (defaults in parentheses):

| Variable | Description |
|----------|-------------|
| `HTTP_PORT` (`8080`) | Port for Gin server |
| `POSTGRES_DSN` (`postgres://postgres:postgres@localhost:5432/users?sslmode=disable`) | GORM DSN |
| `RABBITMQ_DSN` (`amqp://guest:guest@localhost:5672/`) | Rabbit connection |
| `JWT_SECRET` (`supersecret`) | JWT signing secret |
| `TOKEN_TTL_MINUTES` (`60`) | Auth token TTL |
| `ADMIN_USERNAME` (`admin`) / `ADMIN_PASSWORD` (`changeme`) | Credentials for `/auth/login` |

### Running locally

```bash
go run ./cmd/api
```

or via Docker Compose (brings up API + PostgreSQL + RabbitMQ):

```bash
docker compose up --build
```

### API Overview

1. `POST /auth/login` – obtain JWT
2. `GET /api/v1/users` – list users
3. `GET /api/v1/users/:id` – fetch user
4. `POST /api/v1/users` – create user
5. `PUT /api/v1/users/:id` – update user
6. `DELETE /api/v1/users/:id` – delete user
7. `GET /api/v1/users/:id/files` – list files
8. `POST /api/v1/users/:id/files` – attach file
9. `DELETE /api/v1/users/:id/files` – remove all files

All `/api/v1/**` routes require `Authorization: Bearer <token>` obtained from `/auth/login`.

### Tests

```bash
go test ./...
```

Tests spin up ephemeral PostgreSQL containers via Testcontainers. If Docker is unavailable (e.g., rootless Windows), point the suite at an existing database by setting:

```bash
export TEST_POSTGRES_DSN="postgres://user:pass@host:port/db?sslmode=disable"
go test ./...
```

### Documentation & Postman

- Detailed endpoint walkthrough: `docs/API.md`
- Ready-to-run Postman collection: import `docs/postman_collection.json`, update `base_url`, `username`, and `password` variables, then execute the requests in order to validate the full workflow end-to-end.

### GitHub Actions

The repo includes `.github/workflows/ci.yml` which installs Go 1.24, runs `go test ./...`, and therefore covers both service-level and e2e suites on every push / PR.

### RabbitMQ Consumer

Run `go run ./cmd/consumer` to start the console-based RabbitMQ subscriber (requires `RABBITMQ_DSN`). It logs every event type + user ID, demonstrating a pluggable consumer that works with the same event contracts.
