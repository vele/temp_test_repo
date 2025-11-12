## API Guide

This document provides a concise walkthrough of every endpoint exposed by the User Management API. All `api/v1` routes require a JWT token retrieved from `/auth/login`.

### Authentication

```
POST /auth/login
{
  "username": "admin",
  "password": "changeme"
}
```

Response:

```
200 OK
{ "token": "<jwt>" }
```

Set `Authorization: Bearer <jwt>` for all requests below.

### Users

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/api/v1/users` | List all users |
| `GET` | `/api/v1/users/{id}` | Fetch a single user |
| `POST` | `/api/v1/users` | Create user (`name`, `email`, `age`) |
| `PUT` | `/api/v1/users/{id}` | Update user (any subset of `name/email/age`) |
| `DELETE` | `/api/v1/users/{id}` | Delete user |

Constraints:

- `email` must be unique.
- `age` must be greater than 18.

### Files

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/api/v1/users/{id}/files` | List files for user |
| `POST` | `/api/v1/users/{id}/files` | Attach a file (`name`, `path`) |
| `DELETE` | `/api/v1/users/{id}/files` | Delete all files for user |

### Events

Create/Update/Delete actions publish `UserCreated`, `UserUpdated`, and `UserDeleted` events to RabbitMQ. The payload includes the user ID plus current state. See `cmd/consumer` for an example subscriber.

### Local Testing (Postman)

1. Import `docs/postman_collection.json`.
2. Set collection variables:
   - `base_url`: e.g., `http://localhost:8080`
   - `username`, `password` (defaults match `.env` / docker compose).
3. Run the collection in sequence. The login request stores `token` as a collection variable that later requests reuse.

### Sequence Example

1. `Auth / Login`
2. `Users / Create`
3. `Users / List`
4. `Users / Get`
5. `Users / Update`
6. `Files / Add`
7. `Files / List`
8. `Files / Delete All`
9. `Users / Delete`

Each request uses the shared JWT and IDs saved via Postman `tests` scripts (see collection). Feel free to extend the collection with additional flows or environments (e.g., staging).
