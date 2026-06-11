# Repository Guidelines

## Project Structure & Module Organization

This is a Go 1.24 web clipboard service built on Gin with a pnpm/Vite React frontend. Backend code lives under `backend/`: `cmd/web-clipboard/main.go` is the application entry point and `backend/internal/` contains `handlers/`, `middleware/`, `models/`, `services/`, and `utils/`. Frontend source lives under `frontend/src/`, public assets under `frontend/public/`, and built files under ignored `frontend/dist/`. Docker and Compose support lives in `Dockerfile` and `docker-compose.yml`. Generated binaries, runtime data, dependencies, and local state belong in ignored paths such as `bin/`, `data/`, `node_modules/`, and `.omx/`.

## Build, Test, and Development Commands

- `make build`: builds `frontend/dist` and `bin/web-clipboard-go.exe`.
- `make run`: builds and starts the app at `http://localhost:5000`.
- `make test`: builds the frontend, then runs `go test -v ./...`.
- `pnpm --dir frontend install --frozen-lockfile`: installs frontend dependencies from `frontend/pnpm-lock.yaml`.
- `pnpm --dir frontend build`: builds Vite assets into `frontend/dist/`.
- `go build -o bin/web-clipboard-go.exe ./backend/cmd/web-clipboard`: manual build path.
- `make docker-build`, `make docker-run`, `make docker-logs`, `make docker-stop`: Docker lifecycle helpers.
- `docker-compose up -d`: starts the default compose deployment.

## Coding Style & Naming Conventions

Run `gofmt` on all changed Go files before submitting. Keep package names short and lowercase, exported Go identifiers in PascalCase, and unexported identifiers in camelCase. Follow the existing layered layout: handlers should stay thin, middleware should wrap requests, and reusable business/security logic should live in services or utilities. Frontend packages must be declared in `frontend/package.json` and imported from source modules; do not restore CDN or vendored browser package scripts in HTML. Do not add new dependencies unless they are clearly needed.

## Testing Guidelines

Place Go tests next to the code they cover using `*_test.go`. Prefer table-driven tests for handler, service, and validation logic. Run `make test` before opening a pull request. If a change affects authentication, rate limiting, file validation, cleanup, or concurrent shared state, include regression coverage or document why it was not practical.

## Commit & Pull Request Guidelines

Commit messages must use Chinese only. Keep commits focused and describe why the change exists. Follow the Lore commit protocol when possible: include useful trailers such as `Constraint:`, `Rejected:`, `Confidence:`, `Scope-risk:`, `Tested:`, and `Not-tested:` after a blank line. Trailer values must also use Chinese; command names, paths, API routes, environment variables, package names, configuration keys, and code symbols may remain unchanged.

Pull requests should include a short summary, verification commands run, configuration or security impact, and screenshots when UI behavior changes.

## Security & Configuration Tips

Do not commit secrets, certificates, local `.env` files, runtime `data/`, or generated binaries. Preserve existing file type checks, content scanning, rate limits, session expiry, and cleanup behavior unless the security tradeoff is explicit and tested.
