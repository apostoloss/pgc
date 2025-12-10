# Implementation Plan

## 1. Define Asset Models
- Create Go structs for Chart, Insight, Audience (with description and unique fields).
- Use Go interfaces/embedding for common attributes.

## 2. In-Memory Store
- Struct with mutex and maps for users and their favourites.
- Methods for add, remove, edit, and list.

## 3. REST API Endpoints
- Use `net/http` package.
- Routes:
    - `GET /users/{id}/favorites`      -- List all favourites for a user
    - `POST /users/{id}/favorites`     -- Add asset to favourites
    - `DELETE /users/{id}/favorites/{asset_id}` -- Remove asset
    - `PATCH /users/{id}/favorites/{asset_id}` -- Edit description
- JSON request/response types.


## 4. Add GitHub Actions Workflow for Automated Testing
- Create a `.github/workflows/go.yml` file for CI pipeline.
- Configure it to run `go test ./...` and lint on every push/PR.


## 5. Basic Server Setup
- Main entry point in `cmd/server/main.go`.
- Route handling, graceful shutdown.

## 6. Health/Status Endpoint
- Add a `GET /healthz` endpoint.
- Returns service status and implementation version in a JSON response, e.g. `{ "status": "ok", "version": "1.0.0" }`.
- Ensure this endpoint is documented in Swagger/OpenAPI.


## 6. Tests
- Unit tests for models and store logic.
- Handler (integration) tests for API, using `net/http/httptest`.


## 7. Metrics and Monitoring Endpoint
- Add a `/metrics` endpoint using a library like Prometheus client_golang.
- Expose basic metrics: uptime, request counts, error rates, and resource usage.
- Ensure metrics are available for internal monitoring and can be secured as needed.
- Document this endpoint in Swagger/OpenAPI.

## 8. Linting, Formatting, and Makefile Tasks
- Ensure `gofmt`/`golangci-lint` pass locally and in CI.

## 9. Swagger/OpenAPI Documentation
- Integrate Swagger or a similar tool for API documentation.
- Maintain up-to-date docs describing all endpoints and usage examples.

## 10. (Optional) Dockerfile
- Add Dockerfile for containerization.
