# Coding standards

Where `.agents/skills/use-modern-go/SKILL.md` and this file disagree, this file wins.

## Layering

**Files:** `internal/router/`, `internal/handler/`, `internal/service/`, `internal/repository/`, `internal/model/`, `internal/dto/`

One direction: `app → router → handler → service → repository → model`. A layer calls the one below it and never reaches back up.

- `internal/router/route.go` is the only place that knows a URL. A new endpoint adds a line there and a method on a handler.
- Handlers bind a DTO, call one service method, map the result, return. SQL, branching on business state, and cache reach-ins sit below.
- Services take `context.Context` first and return `(*model.X, error)` or `error`.
- Wiring lives in `internal/bootstrap/bootstrap.go` as `Bootstrap` fields, passed down through `New(...)`. A package-level service variable or `init()` cuts the seam tests mock through.

**Canonical:** `internal/handler/uptime/handler.go` (`CreateHost`), `internal/bootstrap/bootstrap.go`.

## HTTP boundary and errors

**Files:** `internal/util/errors.go`, `internal/handler/**`

Every failure leaves through `util.HandleError(ctx, err)` followed by `return`. It is the single mapper: `dto.ValidationError` → 400 with its message, `io.EOF` → no response, anything else → logged plus a generic 500. A second mapping drifts; `err.Error()` in a 500 body leaks internals.

Statuses are `http.Status*` constants. `internal/handler/uptime/handler.go` and `internal/handler/health.go` still carry bare `200`/`201`; new and touched lines use the constants, as `internal/handler/notification/handler.go:34` does.

A request-scoped goroutine detaches with `context.WithoutCancel(ctx)` and sets its own deadline, as `pingHostAndNotify` does.

**Canonical:** `internal/util/errors.go`.

## DTOs

**Files:** `internal/dto/`

Every request and response crosses the boundary as a DTO; `model` types never appear in JSON.

- `<Verb><Noun>Request` carries both `json:` and `binding:` tags, plus `Validate() error` when a rule spans fields. `Validate` returns `ValidationError{MissingFields: []string{"Field"}}`, the type `util.HandleError` turns into a 400.
- The request reaches `model` through `ToModel()`; the way back is a free function per shape: `ToHost`, `ToAllHosts`, `ToHistories`. Allocation is `make([]T, len(in))` with an index loop, matching the others.
- Entry is `dto.BindAndMap[dto.SomeRequest](ctx)`, which binds, collects missing fields, and runs `Validate` in one step. A handler that calls `ctx.ShouldBindJSON` skips the validation the DTO already declared.
- Nested request types stay unexported (`tlsRequest`, `dnsRequest`) behind unexported `mapTLS` / `mapDNS`, so the wire shape never escapes `dto`.

**Canonical:** `internal/dto/host.go`.

## Repository interfaces

**Files:** `internal/repository/`, `internal/service/*/repository.go`

- `internal/repository` holds the GORM implementation: generic `BaseRepository[T]` plus typed interfaces (`HostRepository`, `HistoryRepository`, `NotificationRepository`).
- Each service redeclares the interface it needs in `internal/service/<name>/repository.go`, next to a `newRepository` struct forwarding to the concrete repos. Tests mock that interface; a service importing `*repository.BaseRepository` directly closes the seam.
- Cache invalidation sits in the service that writes, keyed `pingopher_<entity>:<id>` and `pingopher_hosts:all`.
- Cache writes and invalidations go through `cache.Remember` / `cache.Forget`, which log the failure: a dropped cache error is a lint error here, and a stale cache is worth a log line, not a failed request.

**Canonical:** `internal/service/uptime/repository.go`.

## Models

**Files:** `internal/model/`

- Embed `BaseModel` (ksuid ID) and `Timestamps`; soft-delete with `gorm.DeletedAt`. Tags stay inline and lowercase.
- A closed set is `type X string` with a `const` block of `XValue X = "value"` (`HostStatus`).
- A JSON column is a named type with `driver.Value` on the value receiver and `sql.Scanner` on the pointer receiver, accepting `[]byte` and `string` and returning a typed error otherwise (`TLS`, `DNSs`).
- A URL built from protocol, host, and port is a method on the model, and the frontend builds it in exactly one place (`frontend/src/components/HostList.tsx:15`). Change both.

**Canonical:** `internal/model/host.go`.

## Go tests

**Files:** `internal/**/*_test.go`

- Tests are `package <name>_test` beside `package <name>`, importing the package under test. Intrapackage tests let a test pass on internals a caller cannot reach.
- Drive the exported API. A mock is a struct of `XxxFunc` fields implementing the package's exported interface, defaulting to `errors.New("mock XxxFunc not implemented")` — `MockHostRepository` is the shape.
- Test knobs come from the production constructor (`config.Config` fields), not a `Set...ForTest` hook for a value the constructor already takes.
- `export_test.go` is the fallback for a symbol a test cannot otherwise construct or observe: an alias or one-line wrapper over an unexported symbol (`dnsCache`, `queryServer`), compiled into the test binary only.
- Runtime behaviour CI cannot execute (the ICMP prober) swaps through exported `Set...ForTest` / `Reset...ForTest` hooks in the production file.
- A failure message carries both sides: what was got, what was wanted, named in the format string.

**Canonical:** `internal/service/uptime/service_test.go`, `internal/dto/history_test.go`.

## Frontend

**Files:** `frontend/src/`

- `api.ts` owns every wire type and every axios call. A component imports from it and builds no URL of its own.
- Components are `const Name: React.FC<Props> = (props) => {...}` with `export default Name` at the bottom. `HostDetail` and `NotificationManager` are `lazy()`-loaded in `HostList.tsx`, so they keep a default export.
- State stays local: `useState` plus `useEffect` polling (`POLL_INTERVAL` in `HostList.tsx`), with cleanup in the effect's return. Context is for `ToastProvider` / `useToast` only.
- Pure logic moves to `frontend/src/utils/*.ts` with a `node:test` file beside it. Components are not tested.
- Style through the custom properties in `tokens.css`; a literal colour or radius drifts from the theme.

**Canonical:** `frontend/src/api.ts`, `frontend/src/components/HostList.tsx`.

## Comments

**Files:** repo-wide

- A comment earns its line only as a constraint, a workaround, or a trap the code cannot show: `// PingHost applies its own deadline, so only detach from the request.` Uncommented code is the default.
- One or two lines, above the line it explains. An exported doc comment is one sentence.
- Tool-owned comments stay: the Vite template URL in `frontend/vite.config.ts`, `eslint-disable` directives, `//go:` pragmas.

## Verify

Run the block for the changed files; all green before handoff.

```sh
# Go (CI: .github/workflows/ci.yml)
go test ./...
go tool golangci-lint run # lint + gofmt/goimports, test files included

# Go idioms for a touched file (prints the guideline list for Go 1.27)
sh .agents/skills/use-modern-go/scripts/run-tool.sh list --file-path internal/service/uptime/service.go

# Frontend
cd frontend && npx tsc -b && npx eslint .
cd frontend && node --test src/utils/historyParser.test.ts
```
