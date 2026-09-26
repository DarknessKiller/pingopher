<!-- CODEGRAPH_START -->
## CodeGraph

In repositories indexed by CodeGraph (a `.codegraph/` directory exists at the repo root), reach for it BEFORE grep/find or reading files when you need to understand or locate code:

- **MCP tool** (when available): `codegraph_explore` answers most code questions in one call — the relevant symbols' verbatim source plus the call paths between them, including dynamic-dispatch hops grep can't follow. Name a file or symbol in the query to read its current line-numbered source. If it's listed but deferred, load it by name via tool search.
- **Shell** (always works): `codegraph explore "<symbol names or question>"` prints the same output.

If there is no `.codegraph/` directory, skip CodeGraph entirely — indexing is the user's decision.
<!-- CODEGRAPH_END -->

## Comments

- Default to no comment. Add one only for the non-obvious why: a constraint, a workaround, a trap.
- One line, above the line it explains. A comment that restates the code or the name goes.
- Rationale longer than a line belongs in the commit message or PR description. Exported doc comments stay one sentence.
- Tool-owned comments stay: the Vite template URL in `frontend/vite.config.ts`, `eslint-disable` directives, `//go:` pragmas. Comment cleanup never deletes them.

## Testing

- Test files use the external test package: `package <name>_test` beside `package <name>`, importing the package under test (`package uptime_test` in `internal/service/uptime`). Drive the exported API; mocks implement the package's exported interfaces.
- Drive the exported API before reaching for a seam: the alert-cooldown test calls `SendNotification`, not the unexported gate inside it.
- `export_test.go` in the package under test is the fallback for a symbol the test cannot otherwise construct or observe (`dnsCache`, `queryServer` in `internal/service/uptime`) — an alias or one-line wrapper, compiled into the test binary only.
- Test knobs come from the production constructor (`config.Config` fields), not a `Set...ForTest` hook for a value the constructor already takes.
- Runtime behaviour CI cannot execute (the ICMP prober) swaps through exported `Set...ForTest` / `Reset...ForTest` hooks in the production file.
- Gate: `go test ./...` and `go vet ./...` pass.
