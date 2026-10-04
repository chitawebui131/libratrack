# Project conventions (read-only context for every task)

- Go module name is `libratrack`. NEVER rename it.
- Framework: github.com/gin-gonic/gin. Do not add other dependencies unless asked.
- Layout: cmd/api, internal/{model,handler,service,repository,middleware}.
- Handlers are thin; storage in internal/repository.
- Error envelope everywhere: {"error": {"code": "...", "message": "...", "details": "..."}}.
- Success: 200/201/204; validation 422; not found 404.
- Make minimal changes: do not touch files unrelated to the current task.
- Code must compile: `go build ./...` must pass.
- NEVER insert non-ASCII / Ukrainian characters into .go files.
- Output ONLY the file contents. No explanations.
