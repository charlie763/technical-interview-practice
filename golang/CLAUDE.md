# Go — Agent Guidelines

## Prerequisites

- Go 1.22 or newer must be installed (`go version`)
- No external dependencies — all problems use only the standard library

## Workflow per problem

### 1. Read the problem

Each problem lives in `golang/practice_problems/problem_NN_<name>/`:

| File | Purpose |
|---|---|
| `problem.go` | Problem description + type definitions + function stubs (all `panic("not implemented")`) — copy this to start |
| `problem_test.go` | Test suite — read-only during practice |

### 2. Create your answer file

Copy `problem.go` to `practice_problem_answers/` and rename it:

```bash
cp golang/practice_problems/problem_01_geofence_alert_engine/problem.go \
   golang/practice_problem_answers/cw_answer_01_geofence_alert_engine.go
```

Open the answer file and implement every function (replace each `panic("not implemented")`).

Your answer file **must**:
- Declare the same package name as the problem (e.g. `package geofence`)
- Only import standard-library packages
- Include the type definitions copied from `problem.go` (keep them as-is, just implement the panicking functions)

For class-based problems (Problem 03+), define your own struct and implement the interface —
the problem file tells you what interface to satisfy.

### 3. Run the tests against your answer

Use `run_tests.sh` from the repo root:

```bash
./run_tests.sh \
  -f golang/practice_problem_answers/cw_answer_01_geofence_alert_engine.go \
  -c go test -v .
```

Run a specific test group:

```bash
./run_tests.sh \
  -f golang/practice_problem_answers/cw_answer_01_geofence_alert_engine.go \
  -c go test -v -run TestIsInZone .
```

#### How it works

`run_tests.sh` detects the `.go` extension and:
1. Extracts the problem ID from the answer filename (e.g. `01_geofence_alert_engine`).
2. Locates `golang/practice_problems/problem_01_geofence_alert_engine/`.
3. Creates a temporary directory containing:
   - `problem_test.go` from the problem directory (the test suite)
   - Your answer file, copied as `solution.go`
   - A minimal `go.mod` (`module practice`)
4. Runs your `-c` command (e.g. `go test -v .`) inside that temp directory.
5. Cleans up the temp directory on exit.

No common files are ever modified.

### 4. (Optional) Browse the stub test output

Run the stub directly to see all test names (they will all panic):

```bash
cd golang
go test ./practice_problems/problem_01_geofence_alert_engine/ -v 2>&1 | head -50
```

---

## Problem design rules

### Package naming

Each problem uses a short, descriptive package name:

| Problem | Package |
|---|---|
| Problem 01 — Geofence Alert Engine | `geofence` |
| Problem 02 — API Rate Limiter | `ratelimiter` |
| Problem 03 — Permission Manager | `rbac` |
| Problem 04 — Biomarker Alert Monitor | `biomarker` |
| Problem 05 — Medication Titration Tracker | `titration` |
| Problem 06 — Lab Cadence Compliance Monitor | `labcadence` |
| Problem 07 — Care Team Assignment Manager | `careteam` |
| Problem 08 — Patient Enrollment Pipeline | `enrollment` |
| Problem 09 — Multi-Source Incident Aggregator | `incidents` |
| Problem 10 — Responder Dispatch Manager | `dispatch` |
| Problem 11 — Sensor Coverage Tracker | `coverage` |
| Problem 12 — Contract Expiration Alert Scheduler | `alertscheduler` |
| Problem 13 — Contract Lifecycle State Machine | `lifecycle` |
| Problem 14 — Contract Amendment Manager | `amendments` |
| Problem 15 — Tic-Tac-Toe Engine | `tictactoe` |
| Problem 16 — Policy Premium Rating Engine | `premiumrating` |
| Problem 17 — Claims Processing Pipeline | `claims` |
| Problem 18 — Donation Processor | `donations` |
| Problem 19 — Walkathon Pledge Tracker | `pledges` |
| Problem 20 — Donor Communication Suppressor | `suppressor` |
| Problem 21 — Platform Fee Calculator | `fees` |
| Problem 22 — Giving Day Challenge Engine | `challenges` |

New problems should follow the same pattern. Avoid `main` and generic names like `solution`.

### File layout per problem

```
problem_NN_<name>/
  problem.go      # Problem description + type definitions + function stubs (copy this to start)
  problem_test.go # White-box tests (same package, read-only during practice)
```

### Writing problem files

- All function bodies must be `panic("not implemented")`.
- Include a comprehensive comment block at the top (context, data model, example).
- Declare error sentinel variables and all types in `problem.go`:
  ```go
  var ErrAlreadyExists = errors.New("already exists")
  var ErrNotFound      = errors.New("not found")
  ```
- For class-based problems, define the interface and a `New<Name>()` constructor stub in `problem.go`.

### Error handling

Use sentinel errors and `errors.Is` for test assertions:

| Python | Go |
|---|---|
| `raise ValueError` | `return ErrAlreadyExists` |
| `raise KeyError` | `return ErrNotFound` |
| `raise ValueError` (bad plan) | `return ErrInvalidPlan` |

Return `(value, error)` pairs for fallible operations. Predicates and void operations use `bool` or just `error`.

### Test writing rules

- Tests are **white-box**: the test file declares the same package as the implementation.
- Use `t.Run("description", func(t *testing.T) {...})` for subtests — one `TestXxx` function per logical group (e.g. `TestIsInZone`, `TestProcessLocationUpdate`).
- Use helper functions (`newState(t)`, `seededGW(t)`) instead of ad-hoc inline construction.
- Order test functions to match Part order — finishing Part 1 should show only Part 1 tests passing.
- Only import standard library in test files.
- Tests for Part N must only call functions defined in Parts 1–N.

### Standard library only

Problems and their answers must only import standard-library packages. The temp-dir test runner creates a fresh `go.mod` with no dependencies. Common useful packages:

- `"errors"` — sentinel errors + `errors.Is`
- `"fmt"` — `fmt.Errorf`
- `"sort"` — sorting slices
- `"strings"` — string manipulation

---

## Answer file naming convention

```
[prefix]_answer_NN_<name>.go
```

Examples:
- `cw_answer_01_geofence_alert_engine.go` — Charlie's answer
- `en_answer_02_api_rate_limiter.go` — another developer's answer

The `run_tests.sh` script extracts the problem ID via the regex `[0-9]{2}_[a-z_]+`, so any prefix works as long as the `NN_name` segment is present.
