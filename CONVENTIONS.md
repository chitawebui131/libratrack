# CONVENTIONS.md — libratrack

## Critical rules for the AI

1. Output ONLY the requested file contents in Aider format.
2. Never write explanations, comments outside code, or Ukrainian text inside .go files.
3. Never invent extra files or change the directory structure.
4. Always use pure UTF-8. Never insert non-ASCII characters into source files.
5. Prefer simple, correct, readable Go code over clever solutions.
6. Make minimal changes only.

## Import rules (VERY IMPORTANT)

- Module name is exactly: `libratrack`
- All internal imports MUST be written as:
  - "libratrack/internal/model"
  - "libratrack/internal/repository"
  - "libratrack/internal/handler"
  - "libratrack/internal/middleware"
  - "libratrack/internal/service"

- NEVER write any of these (forbidden):
  - "github.com/libratrack/..."
  - "github.com/..." for internal packages
  - Any other prefix for internal packages

Wrong (FORBIDDEN):
import "github.com/libratrack/internal/model"

Correct:
import "libratrack/internal/model"

## Project structure (do not change)

libratrack/
├── cmd/api/main.go
├── internal/
│   ├── model/
│   ├── handler/
│   ├── repository/
│   ├── middleware/
│   └── service/
├── go.mod
└── go.sum

## Coding conventions

- Module name: `libratrack`
- Go version: 1.22+
- Framework: Gin (github.com/gin-gonic/gin)
- All book routes live under /api/v1
- Health check: GET /health → {"status":"ok"}
- Error envelope (mandatory):
  {
    "error": {
      "code": "some_code",
      "message": "human readable",
      "details": "..."
    }
  }
- Status codes:
  - 201 Created
  - 204 No Content
  - 404 Not Found
  - 422 Unprocessable Entity
  - 500 Internal Server Error

## Repository rules

- Use sync.RWMutex
- Write operations (Create / Update / Delete) must take Lock()
- Read operations take RLock()
- Package-level error: var ErrNotFound = errors.New("not found")

## Handler rules

- Constructor: NewBookHandler(repo repository.BookRepository)
- Use ShouldBindJSON for validation
- Never panic — return proper JSON errors

## Middleware order (mandatory)

1. RecoveryMiddleware
2. LoggingMiddleware
3. CORSMiddleware

## What you must NOT do

- Do not add authentication
- Do not add database (Postgres, SQLite, etc.)
- Do not change field names of the Book model
- Do not rename packages or directories
- Do not add extra endpoints beyond the required ones
- Do not write long comments or documentation inside code unless explicitly asked
- Do not use github.com/libratrack/... in any import

## Preferred response style

When asked to create or edit a file — reply with the complete file content only.
When asked to fix something — show only the files that actually changed.
