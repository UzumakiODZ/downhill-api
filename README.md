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

## Notes

- The server serves GraphQL Playground at `/` and the GraphQL API at `/query`.
- Generated files under `graph/generated.go` and `graph/model/models_gen.go` should be treated as codegen output.