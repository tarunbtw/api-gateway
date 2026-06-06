# api-gateway

A minimal API gateway written in Go. Handles JWT authentication, token bucket rate limiting, and reverse proxying to multiple backend services. Runs as a single binary in front of any number of upstream services.

Similar in concept to Kong or AWS API Gateway, without the complexity.

---

## Architecture

```
client
  |
  v
gateway :8080
  |
  +-- validate JWT
  |
  +-- check rate limit (token bucket, per IP per route)
  |
  +-- route by path prefix
        |
        +-- /api/users  -->  users service :3001
        |
        +-- /api/orders -->  orders service :3002
```

Auth and rate limiting happen at the gateway. Backend services receive only authenticated, rate-limited traffic and trust whatever comes through. They do not implement auth themselves.

---

## Features

- **Reverse proxy** — forwards full request (method, headers, body) to upstream, strips route prefix before forwarding
- **JWT auth** — validates HS256 signed tokens on every request, returns 401 on missing or invalid tokens
- **Token bucket rate limiting** — per IP per route, configurable max tokens and refill rate, returns 429 on exhaustion
- **Environment-based config** — upstream URLs and JWT secret are injected via env vars, no hardcoded values in production

---

## Running locally

```bash
# start backend services
go run services/users/main.go
go run services/orders/main.go

# start gateway
go run .
```

Get a token:

```bash
curl http://localhost:8080/token
```

Hit a route:

```bash
curl http://localhost:8080/api/users/123 \
  -H "Authorization: Bearer <token>"

curl http://localhost:8080/api/orders/456 \
  -H "Authorization: Bearer <token>"
```

Without a token:

```bash
curl http://localhost:8080/api/users/123
# 401 missing authorization header
```

---

## Docker

All three services boot with a single command:

```bash
docker-compose up
```

This starts the gateway on `:8080` and both backend services on an isolated Docker network. The gateway reaches backends by service name (`http://users:3001`, `http://orders:3002`) — no exposed ports on the backend containers.

---

## Configuration

| Environment variable | Default                  | Description                        |
|---------------------|--------------------------|------------------------------------|
| `JWT_SECRET`        | `supersecretkey`         | HMAC secret for JWT signing        |
| `USERS_SERVICE`     | `http://localhost:3001`  | Upstream URL for users service     |
| `ORDERS_SERVICE`    | `http://localhost:3002`  | Upstream URL for orders service    |

---

## Project structure

```
.
├── main.go          entry point, route config, middleware chain
├── proxy.go         reverse proxy, prefix stripping, request forwarding
├── auth.go          JWT validation middleware, token generator
├── ratelimit.go     token bucket implementation, per-IP per-route tracking
├── services/
│   ├── users/       dummy upstream service on :3001
│   └── orders/      dummy upstream service on :3002
├── Dockerfile.gateway
├── Dockerfile.users
├── Dockerfile.orders
└── docker-compose.yml
```

---

## Stack

- Go 1.26
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) for JWT parsing and signing
- Standard library only for HTTP — no frameworks