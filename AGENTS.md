# Repository Guidelines

## Project Structure & Module Organization

`cmd/app` contains the bot entrypoint and CGO bootstrap for `ntgcalls`. Core application code lives under `internal/`:
`config` loads env vars, `core` manages runtime state, `modules` contains Telegram command handlers, `database` wraps
MongoDB persistence, `platforms` resolves media sources, and `locales` stores YAML translations such as
`internal/locales/en.yml` and `ptbr.yml`. Low-level call bindings live in `ntgcalls/`, while `ubot/` contains Telegram
group call helpers.

## Build, Test, and Development Commands

Use Go `1.25.7`, matching `go.mod` and `install.sh`.

```bash
cp sample.env .env          # create local config
go run ./cmd/app            # run locally
go build ./cmd/app          # compile the worker binary
go test ./...               # compile-check all packages and run tests
./install.sh --all          # install Go, FFmpeg, yt-dlp, Deno, and ntgcalls deps
docker build -t flora-bot . # build the production image
```

In sandboxed environments, set `GOCACHE=/tmp/go-build`. Validation here was blocked by a local Go toolchain mismatch (
`go1.25.7` module vs `go1.26.2` toolchain), so contributors should verify commands with the pinned Go version.

## Coding Style & Naming Conventions

Format all Go code with `gofmt -w` before opening a PR. Follow standard Go style: tabs for indentation, lowercase
package names, exported identifiers in `CamelCase`, unexported helpers in `camelCase`, and one command/feature per file
where practical (`internal/modules/play.go`, `pause.go`, `skip.go`). Keep locale keys stable and descriptive; add new
user-facing strings through `internal/locales/*.yml` rather than hardcoding text.

## Testing Guidelines

The repository currently has no `*_test.go` files, so add tests with every non-trivial change. Place tests beside the
package they cover and name them `*_test.go`; prefer table-driven tests for parsers, config loading, and platform
selection logic. Run `go test ./...` before submitting, and include manual verification notes for Telegram voice/chat
flows that cannot be covered automatically.

## Commit & Pull Request Guidelines

Recent history follows short Conventional Commit prefixes such as `fix:`, `feat:`, `style:`, and `chore(deps):`. Keep
commits focused and imperative, for example `fix: handle empty queue state`. PRs should include a brief summary, config
or env changes, linked issues, and screenshots or logs when behavior changes affect chat output, playback, or
localization.

## Security & Configuration Tips

Never commit real `.env` values, session strings, Mongo URIs, or bot tokens. Treat `sample.env` as the source template,
keep secrets in environment variables, and document any new required keys in both `sample.env` and the PR description.
