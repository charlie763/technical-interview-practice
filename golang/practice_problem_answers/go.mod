// Answer files are standalone .go snippets used only via run_tests.sh.
// They are NOT a standalone Go module — they require types.go from the
// corresponding problem directory to compile (the run_tests.sh temp-dir
// mechanism handles this automatically).
//
// This go.mod exists solely to prevent `go vet ./...` / `go test ./...`
// from the parent golang/ module from recursing into this directory.
module practice_problem_answers

go 1.22.0
