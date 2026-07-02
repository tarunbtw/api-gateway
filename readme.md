# api-gateway

A minimal API gateway written in Go. Handles JWT authentication, token bucket rate limiting, and reverse proxying to multiple backend services. Runs as a single binary in front of any number of upstream services.

Similar in concept to Kong or AWS API Gateway, without the complexity.

---

## Live Demo

---

## Architecture

```
client
  |
  v
gateway :8080
  |
  +-- validate JWT (HS256)
  |
  +-- check rate limit (token bucket, per IP per route)
  |
  +-- route by path prefix
        |
        +-- /api/users  -->  users service :3001 (in-process for demo)
        |
        +-- /api/orders -->  orders service :3002 (in-process for demo)
```

> **Note:** For this demo deployment the users and orders services run as goroutines inside the gateway process (single binary, single port). In a production setup they would be separate deployable services the gateway reverse-proxies to over the network.

Auth and rate limiting happen at the gateway. Backend services receive only authenticated, rate-limited traffic and trust whatever comes through.

---

## Features

- **Reverse proxy** — forwards full request (method, headers, body) to upstream, strips route prefix before forwarding
- **JWT auth** — validates HS256 signed tokens on every request, returns 401 on missing or invalid tokens
- **Token bucket rate limiting** — per IP per route, configurable max tokens and refill rate, returns 429 on exhaustion; stale buckets are evicted every 5 minutes to prevent unbounded memory growth
- **Environment-based config** — upstream URLs and JWT secret are injected via env vars, no hardcoded values in production
- **Health endpoint** — `GET /health` returns gateway liveness status, configured routes, and rate-limit params
- **Demo frontend** — static dashboard served at `/` with interactive panels for every gateway feature

---

## Running locally

```bash
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

Health check:

```bash
curl http://localhost:8080/health
# {"rateLimit":{"maxTokens":10,"refillPerSecond":2},"routes":["/api/users","/api/orders"],"status":"ok"}
```

---

## Docker

Build and run with a single container:

```bash
docker build -t api-gateway .
docker run -p 8080:8080 -e JWT_SECRET=yourown api-gateway
```

Then open `http://localhost:8080` in your browser.

---

## Configuration

| Environment variable | Default                  | Description                        |
|---------------------|--------------------------|-------------------------------------|
| `JWT_SECRET`        | `supersecretkey`         | HMAC secret for JWT signing (set this!) |
| `USERS_SERVICE`     | `http://localhost:3001`  | Upstream URL for users service      |
| `ORDERS_SERVICE`    | `http://localhost:3002`  | Upstream URL for orders service     |

---

## Project structure

```
.
├── main.go          entry point, route config, middleware chain, /health, /token
├── proxy.go         reverse proxy, prefix stripping, request forwarding
├── auth.go          JWT validation middleware, token generator
├── ratelimit.go     token bucket implementation, per-IP per-route tracking, stale eviction
├── services.go      in-process users (:3001) and orders (:3002) service goroutines
├── static/
│   └── index.html   single-file demo dashboard (no build step)
├── Dockerfile       single-stage build → single container, single port
└── go.mod / go.sum
```

---

## Note

- **`/token` issues tokens without authentication.** Anyone can call `GET /token` and receive a valid JWT for `test-user`. This is intentional for the interactive demo. In a real deployment replace this endpoint with proper credential validation (password/bcrypt, OAuth, etc.).
- **JWT secret has an insecure default fallback (`supersecretkey`).** The gateway works with zero config so the demo runs on Render without any setup, but you _must_ set `JWT_SECRET` to a strong random value in any non-demo deployment.
- **Users and orders are dummy in-process stand-ins.** They return a static JSON blob. In production these would be separate services (different repos, different containers, different teams) that the gateway proxies to over the network.

---

## Stack

- Go 1.26
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) for JWT parsing and signing
- Standard library only for HTTP — no frameworks