# C2DP Go API

Bare-bones Instagram-style demo API built with Echo, GORM, and PostgreSQL.

## Features

- User signup and login
- JWT-protected post creation
- Image-required posts using multipart uploads
- Static serving for uploaded images
- Public feed and per-post fetch endpoints

## Setup

1. Start the API and Postgres with Docker:

```sh
docker compose up --build
```

The API runs on `http://localhost:8080`.

2. Or run only Postgres with Docker and run the API locally:

```sh
docker compose up -d postgres
```

Then configure environment:

```sh
cp .env.example .env
```

Install dependencies and run:

```sh
go mod tidy
go run ./cmd/api
```

You can also create and use a local Postgres database without Docker:

```sh
createdb c2dp_go_api
```

## Endpoints

### Health

```http
GET /health
```

### Signup

```http
POST /api/v1/auth/signup
Content-Type: application/json

{
  "username": "ada",
  "email": "ada@example.com",
  "password": "password123"
}
```

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "identifier": "ada",
  "password": "password123"
}
```

The `identifier` accepts an email address or username. Email matching ignores case;
usernames match the case used at signup. Surrounding whitespace is ignored.
Existing `email` requests and `username` requests also work when `identifier` is
absent. If an identifier matches two different accounts, login returns 401 rather
than choosing an account arbitrarily.

### Create Post

```http
POST /api/v1/posts
Authorization: Bearer <token>
Content-Type: multipart/form-data

caption=hello
image=@/path/to/image.jpg
```

### List Posts

```http
GET /api/v1/posts
```

### Get Post

```http
GET /api/v1/posts/:id
```

## Login integration tests

Set `TEST_DATABASE_URL` to a PostgreSQL test database and run `go test ./... -v`.
The login tests cover email/username authentication, legacy request fields,
normalization, rejected credentials, and ambiguous identifiers. Fixtures and
schema changes run inside a transaction that is rolled back. Without that
variable, the database integration test is skipped.

## Automatic C2PA signing (demo)

Every new image upload is signed with c2patool's **public built-in test certificate**
before its post is saved. Signatures and asset bindings can be verified, but the
signer is untrusted. This does not establish the uploader's identity or camera
capture. Existing credentials are retained as ingredient provenance.

For local development, install c2patool 0.27.22 and set `C2PATOOL_PATH` to its
executable, or place it at `go-api/bin/c2patool` (ignored by Git). A local macOS
binary has been placed there for this workspace. Run `go run ./cmd/api` from
`go-api`. Docker builds its own pinned c2patool and sets the path automatically;
the first Rust build can take several minutes.

POST `/api/v1/posts` is unchanged: bearer token plus multipart `image` and optional
`caption`. Supported formats: JPEG, PNG, WebP, GIF. Request limit: 20 MiB. The
response's `image_url` points to the signed image. Missing tooling or failed
signing returns an error without saving a post or publishing the unsigned image.
Signing times out after 45 seconds. Unsigned staging files are outside `/uploads`
and are removed when the request ends.

The signer uses isolated settings and strips C2PA configuration environment
variables so a developer's personal credentials are not used accidentally.
To inspect a downloaded image: `bin/c2patool downloaded.png`.

To test signing, serving, provenance preservation, and failure cleanup, set both
`TEST_DATABASE_URL` and an absolute `TEST_C2PATOOL_PATH`, then run `go test ./... -v`.
