# AGENTS.md

## Cursor Cloud specific instructions

### Project overview

Go REST API (Gin framework) that validates Brazilian data formats (email, CPF, name, phone, RG, CEP, credit card). Single service, entry point is `main.go`, listens on port `8080`. See `README.md` for full endpoint documentation.

### Running the service

```bash
redis-server --daemonize yes   # optional: enables CPF caching
go run main.go                 # starts API on :8080
```

### Key commands

- **Tests:** `go test ./... -v`
- **Lint:** `go vet ./...`
- **Build:** `go build -o DataValidatorAPI .`
- **Run built binary:** `./DataValidatorAPI`

### Non-obvious notes

- **Redis is optional.** The API runs without Redis — CPF validation still works, but caching is skipped (logs a non-fatal error). All other validators never use Redis.
- **Go version:** `go.mod` specifies Go 1.22.5. The environment uses `GOFLAGS` managed by the Go toolchain.
- **Handler tests that exercise CPF paths** will log `"Error saving to cache: dial tcp ... connection refused"` when Redis is down. This is expected and does not cause test failures.
- **Swagger UI** is available at `http://localhost:8080/swagger/index.html` when the server is running.
- **Environment variable:** `REDIS_ADDR` overrides the Redis address (default `localhost:6379`).
