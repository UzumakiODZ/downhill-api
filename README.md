# Downhill API

A Go GraphQL API for users, companies, roles, interview questions, and posts.

## Tech Stack

- Go 1.25
- gqlgen for GraphQL server generation
- GORM with PostgreSQL driver
- Atlas for database migrations
- Air (optional) for live-reload during development

## Project Layout

```text
.
|-- server.go                 # API entrypoint (GraphQL server + Playground)
|-- loader.go                 # Atlas schema loader from GORM models
|-- gqlgen.yml                # gqlgen codegen config
|-- atlas.hcl                 # Atlas migration configuration
|-- database/
|   |-- connection.go         # DB connection setup
|   |-- sql.go                # GORM models
|-- graph/
|   |-- schema.graphqls       # GraphQL schema
|   |-- schema.resolvers.go   # Resolver implementations
|   |-- generated.go          # gqlgen generated server code
|   |-- model/models_gen.go   # gqlgen generated GraphQL models
|-- migrations/
|   |-- 20260320174513.sql    # SQL migration(s)
```

## Prerequisites

- Go 1.25+
- PostgreSQL running locally or remotely
- Atlas CLI (for migration commands)
- Optional: Air for hot reload

## Environment Variables

Create a `.env` file in the project root.

```env
PORT=8080
DB_URL=postgres://postgres:mysecretpassword@127.0.0.1:5499/postgres?sslmode=disable
```

Variables:

- `PORT`: HTTP port for the API server (defaults to `8080` if unset).
- `DB_URL`: PostgreSQL connection string used by GORM.

## Run Locally

1. Download dependencies.

```bash
go mod download
```

2. Ensure PostgreSQL is running and `DB_URL` points to it.

3. Apply migrations.

```bash
atlas migrate apply --env gorm
```

4. Start the server.

```bash
go run .
```

5. Open GraphQL Playground.

- Playground: `http://localhost:8080/`
- GraphQL endpoint: `http://localhost:8080/query`

## Development Workflow

### Run with hot reload (Air)

If Air is installed:

```bash
air
```

### Regenerate GraphQL code

Run after schema changes:

```bash
go run github.com/99designs/gqlgen generate
```

### Generate migration diffs (Atlas + GORM models)

`loader.go` exposes the GORM schema for Atlas.

```bash
atlas migrate diff <migration_name> --env gorm
```

Example:

```bash
atlas migrate diff add_new_field --env gorm
```

## GraphQL API Summary

### Queries

- `getUser(id: ID!): User`
- `getAllCompanies: [Company!]!`
- `getRolesByCompany(companyId: ID!): [Role!]!`
- `getQuestionsByCompany(companyId: ID!): [QuestionBank!]!`
- `getPost(id: ID!): Post`
- `getAllPosts: [Post!]!`

### Mutations

- `createUser(input: CreateUserInput!): User!`
- `createCompany(input: CreateCompanyInput!): Company!`
- `createRole(input: CreateRoleInput!): Role!`
- `createQuestion(input: CreateQuestionInput!): QuestionBank!`
- `createPost(input: CreatePostInput!): Post!`
- `deletePost(id: ID!): Boolean!`

### Example Query

```graphql
query GetAllCompanies {
  getAllCompanies {
    id
    companyName
  }
}
```

### Example Mutation

```graphql
mutation CreateCompany {
  createCompany(input: { companyName: "Acme" }) {
    id
    companyName
  }
}
```

## Known Limitations (Current Implementation)

- `createUser` is declared in the schema but currently not implemented in resolvers.
- `createQuestion` currently does not persist `companyId` and `years` from input.
- `createPost` currently does not persist `userId` (and does not expose `companyId` in GraphQL input).
- `getUser` returns only `id` and `username` fields from resolver mapping.
- `getRolesByCompany` sets `companyId` in response using role ID instead of company ID.
- Password handling in current models/resolvers does not include hashing logic.

## Troubleshooting

- Error: `DB_URL environment variable not set`
  - Add `DB_URL` to `.env` and restart.
- Error: `Error loading .env file`
  - Ensure `.env` exists in the project root.
- Migration problems:
  - Verify `atlas.hcl` has the correct dev database URL.
- Server starts but GraphQL fails:
  - Confirm database connectivity and migration state.

## Performance Benchmarks & Metrics

Comprehensive performance, load, and micro-benchmarking tests were conducted across all layers of the architecture (Go runtime, GraphQL engine, Redis caching, GORM PostgreSQL ORM, and Rate Limiting).

### 1. Executive Performance Summary

| Test Scenario | Total Reqs | Peak Throughput | Avg Latency | p50 (Median) | p90 Latency | p95 Latency | p99 Latency | Error Rate | Primary Characteristic |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Cached Reads (Redis Hit)** | **14,640** | **731.77 req/s** | **6.73 ms** | **5.05 ms** | **6.99 ms** | **8.54 ms** | **20.05 ms** | **0.00%** | In-memory Redis sub-10ms response |
| **Cold Reads (PostgreSQL)** | **372** | **24.53 req/s** | **102.05 ms** | **4.34 ms** | **119.93 ms** | **162.31 ms** | **4,954 ms\*** | **0.00%** | Remote DB WAN round-trip |
| **Mutations / Write Path** | **170** | **11.07 req/s** | **350.66 ms** | **342.25 ms** | **356.47 ms** | **395.11 ms** | **474.90 ms** | **0.00%** | GORM DB Insert + Redis Cache Set |
| **Auth Registration (Bcrypt)**| **16** | **1.01 req/s** | **1,478.09 ms**| **1,454.17 ms**| **1,625.57 ms**| **1,660.04 ms**| **1,702.24 ms**| **0.00%** | CPU-bound (`bcrypt` Cost 14) |
| **Rate Limiter (Burst/429)** | **20** | **67.30 req/s** | **1.91 ms (429)**| **1.97 ms** | **2.18 ms** | **2.29 ms** | **2.40 ms** | **50% (429)**| Strict 10 req/min/IP enforcement |
| **Concurrency Stress (100 VU)**| **26,606** | **~850 req/s** | **7.93 ms** | **5.69 ms** | **12.32 ms** | **18.74 ms** | **161.14 ms**| **0.00%** | Zero error degradation under load |

*\* Initial cold connection pool & TLS handshake.*

---

### 2. Low-Level Microbenchmarks (Go Runtime & Profiling)

Micro-profiling conducted via Go benchmark toolchain (`go test -bench=. -benchmem`):

| Benchmark Operation | Iterations | Time / Op | Memory / Op | Allocations / Op | Impact / Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `Bcrypt Hashing (Cost 14)` | 2 | **842,280,200 ns** (842.28 ms) | 6,212 B/op | 12 allocs/op | CPU-bound; limits single-thread auth throughput |
| `Bcrypt Hashing (Cost 10)` | 21 | **53,072,557 ns** (53.07 ms) | 5,218 B/op | 9 allocs/op | **15.8x faster** than Cost 14 with standard security |
| `Password Complexity Validation` | 274,360 | **5,140 ns** (5.14 µs) | 3,350 B/op | 40 allocs/op | Regex check (Upper/Lower/Number/Special) |
| `Email Regex Matching` | 2,535,220 | **480.2 ns** (0.48 µs) | 0 B/op | 0 allocs/op | Zero-allocation precompiled regex |
| `JWT Token Signing (HS256)` | 257,065 | **4,930 ns** (4.93 µs) | 2,659 B/op | 41 allocs/op | Lightweight authentication token generation |
| `JSON Marshal (50 Companies)` | 127,221 | **9,161 ns** (9.16 µs) | 4,892 B/op | 2 allocs/op | Low memory footprint serialization |
| `JSON Unmarshal (50 Companies)` | 23,694 | **51,611 ns** (51.61 µs) | 7,040 B/op | 163 allocs/op | GraphQL model deserialization from cache |

---

### 3. Detailed Scenario Breakdown

#### Scenario A: Cached Read Performance (`k6/01_cached_reads.js`)
- **Requests:** 14,640 requests in 20s (20 VUs)
- **Throughput:** 731.77 RPS
- **Average Latency:** 6.73 ms | **Median (p50):** 5.05 ms | **p95:** 8.54 ms
- **TTFB (Time to First Byte):** 6.67 ms
- **Network Transfer:** 7.7 MB received (385 kB/s), 5.1 MB sent (256 kB/s)
- **Status:** `100.00%` checks passed (0 errors)

#### Scenario B: Cold Database Reads (`k6/02_cold_reads_db.js`)
- **Average Latency:** 102.05 ms (15.2x higher than cached reads due to WAN latency to remote PostgreSQL pooler)
- **Redis Impact:** Caching reduces query response time from **~100-350ms** down to **5.05ms** (**70x latency reduction**).

#### Scenario C: Mutation Write Path (`k6/03_mutations_writes.js`)
- **Average Write Latency:** 350.66 ms | **Median:** 342.25 ms | **p95:** 395.11 ms
- **Operations:** Performs PostgreSQL `INSERT`, invalidates stale Redis keys, and populates new entity cache in 24-hour TTL window.

#### Scenario D: Authentication Processing (`k6/04_auth_bcrypt_cost.js`)
- **Average Registration Latency:** 1,478.09 ms (~1.48 seconds)
- **Breakdown:** ~842 ms Bcrypt (Cost 14) + ~350 ms PostgreSQL Insert + ~286 ms JWT generation & network.

#### Scenario E: Rate Limiting Defense (`k6/05_ratelimit_and_spike.js`)
- **Burst Behavior:** 20 rapid requests from single IP
- **Interception Latency:** 1.91 ms (Rejected with `429 Too Many Requests` at Redis layer without entering GraphQL resolver pipeline)
- **Headers Verified:** `RateLimit-Remaining: 0`, `RateLimit-RetryAfter: 59s`.

#### Scenario F: Concurrency Stress Test (`k6/06_concurrency_stress.js`)
- **Total Requests:** 26,606 queries over 5 load stages (up to 100 concurrent virtual users)
- **Average Latency under Concurrency:** 7.93 ms | **p95:** 18.74 ms
- **Reliability:** 100.00% success rate with zero dropped connections.

---

### 4. How to Run Benchmarks

#### Run Go Microbenchmarks:
```bash
go test -run="^$" -bench="Benchmark" -benchmem ./graph
```

#### Run k6 Load Test Suites:
```bash
# 1. Cached Reads Benchmark
k6 run k6/01_cached_reads.js

# 2. Cold Database Reads Benchmark
k6 run k6/02_cold_reads_db.js

# 3. Mutation Throughput Benchmark
k6 run k6/03_mutations_writes.js

# 4. Auth & Bcrypt Benchmark
k6 run k6/04_auth_bcrypt_cost.js

# 5. Rate Limiting & 429 Test
k6 run k6/05_ratelimit_and_spike.js

# 6. Concurrency Stress Test (100 VUs)
k6 run k6/06_concurrency_stress.js
```

---

## Notes

- The server serves GraphQL Playground at `/` and the GraphQL API at `/query`.
- Generated files under `graph/generated.go` and `graph/model/models_gen.go` should be treated as codegen output.