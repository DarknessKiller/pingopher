<!-- CODEGRAPH_START -->
## CodeGraph

In repositories indexed by CodeGraph (a `.codegraph/` directory exists at the repo root), reach for it BEFORE grep/find or reading files when you need to understand or locate code:

- **MCP tool** (when available): `codegraph_explore` answers most code questions in one call — the relevant symbols' verbatim source plus the call paths between them, including dynamic-dispatch hops grep can't follow. Name a file or symbol in the query to read its current line-numbered source. If it's listed but deferred, load it by name via tool search.
- **Shell** (always works): `codegraph explore "<symbol names or question>"` prints the same output.

If there is no `.codegraph/` directory, run `codegraph init` at the repo root. Only `.codegraph/.gitignore` is committed, so a fresh clone builds its own index.
<!-- CODEGRAPH_END -->

## Default routine

1. Read `CODING_STANDARDS.md` before changing a file.
2. Before editing a `.go` file, load `.agents/skills/use-modern-go/SKILL.md` and apply the idiom list it prints for that file.
3. Locate code with CodeGraph, and run `codegraph sync` after a structural change.
4. Run the verify block in `CODING_STANDARDS.md` for the changed files: `go test ./...` and `go tool golangci-lint run` for Go, `npx tsc -b` and `npx eslint .` for the frontend.
5. Check each changed line against its section; where the code is right and the rule is wrong, fix the doc in the same change.

Rules always true live in this file. Rules consulted while writing code live in `CODING_STANDARDS.md`.
