# Agent Notes

- Use the `$Ponytail` skill for all implementation work in this repo.
- In this app, integration tests means Spring-style in-app integration tests: app routes, validation, lifespan, dependency wiring, and mocked external Ark boundary.

## Validation

```bash
go test ./...
go test -race ./...
go vet ./...
make staticcheck
CLOUD_DRIVER_INTEGRATION=1 go test ./internal/... -run Integration
```
