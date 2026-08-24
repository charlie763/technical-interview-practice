#!/usr/bin/env bash
# run_tests.sh — run a test command against a specific practice problem answer file.
#
# ── Python problems ───────────────────────────────────────────────────────────
# Usage:
#   ./run_tests.sh -f <path-to-answer.py> -c <pytest-command...>
#
# Examples:
#   ./run_tests.sh \
#     -f python/practice_problem_answers/cw_answer_03_permission_manager.py \
#     -c pytest python/tests/test_problem_03_permission_manager.py -v
#
#   ./run_tests.sh \
#     -f python/practice_problem_answers/cw_answer_03_permission_manager.py \
#     -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole
#
# How it works (Python):
#   --answer <abs-path> is appended to the test command, causing conftest.py to
#   inject your answer module in place of the problem stub before test collection.
#
# ── TypeScript problems ────────────────────────────────────────────────────────
# Usage:
#   ./run_tests.sh -f <path-to-answer.ts> -c <npm-test-command...>
#
# Examples:
#   ./run_tests.sh \
#     -f typescript/practice_problem_answers/cw_answer_01_donation_processor.ts \
#     -c npm run test:01
#
# How it works (TypeScript):
#   Extracts the stem from the answer filename (e.g. cw_answer_01_donation_processor),
#   sets PRACTICE_ANSWER to that stem, then runs the Jest command from typescript/.
#   jest.config.js uses moduleNameMapper to redirect the stub import to the answer file.
#
# ── Go problems ───────────────────────────────────────────────────────────────
# Usage:
#   ./run_tests.sh -f <path-to-answer.go> -c <go-test-command...>
#
# Examples:
#   ./run_tests.sh \
#     -f golang/practice_problem_answers/cw_answer_01_geofence_alert_engine.go \
#     -c go test -v .
#
#   ./run_tests.sh \
#     -f golang/practice_problem_answers/cw_answer_01_geofence_alert_engine.go \
#     -c go test -v -run TestIsInZone .
#
# How it works (Go):
#   Extracts the problem ID from the answer filename (e.g. 01_geofence_alert_engine),
#   copies *_test.go from the problem directory plus the answer file
#   (as solution.go) into a temp directory, writes a minimal go.mod, then runs
#   the test command inside that temp directory. No common files are modified.
#
# ── React problems ────────────────────────────────────────────────────────────
# Usage:
#   ./run_tests.sh -f <path-to-answer-dir> -c <npm-test-command...>
#   ./run_tests.sh -f <path-to-answer-dir/App.jsx> -c <npm-test-command...>
#
# Examples:
#   ./run_tests.sh \
#     -f react/practice_problem_answers/cw_answer_02_incident_dashboard \
#     -c npm run test:02
#
#   ./run_tests.sh \
#     -f react/practice_problem_answers/cw_answer_02_incident_dashboard/App.jsx \
#     -c npm run test:02
#
#   ./run_tests.sh \
#     -f react/practice_problem_answers/cw_answer_02_incident_dashboard \
#     -c npm run test:ui
#
# How it works (React):
#   Sets PRACTICE_ANSWER to the answer directory (absolute path), then runs the
#   Playwright command from react/. Vite's plugin redirects main.jsx's App import
#   to <PRACTICE_ANSWER>/App.jsx. react/src/App.jsx is never modified.
#   Playwright always spawns a fresh dev server when PRACTICE_ANSWER is set.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── parse arguments ──────────────────────────────────────────────────────────
ANSWER=""
CMD=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        -f)
            ANSWER="$2"
            shift 2
            ;;
        -c)
            shift
            CMD=("$@")
            break
            ;;
        *)
            echo "Error: unexpected argument '$1'"
            echo "Usage: $0 -f <answer-file> -c <test-command...>"
            exit 1
            ;;
    esac
done

if [[ -z "$ANSWER" ]]; then
    echo "Error: -f <answer-file> is required."
    echo "Usage: $0 -f <answer-file> -c <test-command...>"
    exit 1
fi

if [[ ${#CMD[@]} -eq 0 ]]; then
    echo "Error: -c <test-command> is required."
    echo "Usage: $0 -f <answer-file> -c <test-command...>"
    exit 1
fi

if [[ -d "$ANSWER" ]]; then
    ANSWER_ABS="$(cd "$ANSWER" && pwd)"
    IS_DIR=true
elif [[ -f "$ANSWER" ]]; then
    ANSWER_ABS="$(cd "$(dirname "$ANSWER")" && pwd)/$(basename "$ANSWER")"
    IS_DIR=false
else
    echo "Error: not found: $ANSWER"
    exit 1
fi

# ── TypeScript mode: .ts answer files (not .tsx) ─────────────────────────────
if [[ "$IS_DIR" == false && "$ANSWER_ABS" == *.ts && "$ANSWER_ABS" != *.tsx ]]; then
    ANSWER_STEM="$(basename "$ANSWER_ABS" .ts)"
    echo "Answer : $ANSWER → PRACTICE_ANSWER=$ANSWER_STEM"
    echo "Command: ${CMD[*]}"
    echo ""
    cd "$REPO_ROOT/typescript"
    PRACTICE_ANSWER="$ANSWER_STEM" "${CMD[@]}"
    exit $?
fi

# ── React mode: answer directory or .jsx / .tsx file ─────────────────────────
if [[ "$IS_DIR" == true || "$ANSWER_ABS" == *.jsx || "$ANSWER_ABS" == *.tsx ]]; then
    if [[ "$IS_DIR" == true ]]; then
        ANSWER_DIR="$ANSWER_ABS"
    else
        ANSWER_DIR="$(dirname "$ANSWER_ABS")"
    fi

    echo "Answer : $ANSWER → PRACTICE_ANSWER=$ANSWER_DIR"
    echo "Command: ${CMD[*]}"
    echo ""

    cd "$REPO_ROOT/react"
    PRACTICE_ANSWER="$ANSWER_DIR" "${CMD[@]}"
    exit $?
fi

# ── Go mode: .go answer files ────────────────────────────────────────────────
if [[ "$IS_DIR" == false && "$ANSWER_ABS" == *.go ]]; then
    FILENAME=$(basename "$ANSWER_ABS" .go)
    # Extract NN_name from filename (e.g. cw_answer_01_geofence_alert_engine → 01_geofence_alert_engine)
    PROBLEM_ID=$(echo "$FILENAME" | grep -oE '[0-9]{2}_[a-z_]+')

    if [[ -z "$PROBLEM_ID" ]]; then
        echo "Error: could not extract problem ID from filename: $FILENAME"
        echo "Expected format: *_NN_<name>.go (e.g. cw_answer_01_geofence_alert_engine.go)"
        exit 1
    fi

    PROBLEM_DIR="$REPO_ROOT/golang/practice_problems/problem_${PROBLEM_ID}"

    if [[ ! -d "$PROBLEM_DIR" ]]; then
        echo "Error: no problem directory found: $PROBLEM_DIR"
        exit 1
    fi

    # Create a temp working directory; clean it up on exit
    GO_TMPDIR=$(mktemp -d)
    trap "rm -rf '$GO_TMPDIR'" EXIT

    # Copy *_test.go from the problem directory
    for f in "$PROBLEM_DIR"/*_test.go; do
        cp "$f" "$GO_TMPDIR/"
    done

    # Copy the answer file as solution.go
    cp "$ANSWER_ABS" "$GO_TMPDIR/solution.go"

    # Write a minimal go.mod so 'go test' resolves correctly
    printf 'module practice\n\ngo 1.22.0\n' > "$GO_TMPDIR/go.mod"

    echo "Answer : $ANSWER → $GO_TMPDIR/solution.go"
    echo "Command: ${CMD[*]}"
    echo "Working: $GO_TMPDIR"
    echo ""

    cd "$GO_TMPDIR"
    "${CMD[@]}"
    exit $?
fi

# ── Python mode: .py answer files ────────────────────────────────────────────

# activate venv if present
if [[ -f "$REPO_ROOT/.venv/bin/activate" ]]; then
    source "$REPO_ROOT/.venv/bin/activate"
fi

echo "Answer : $ANSWER"
echo "Command: ${CMD[*]} --answer $ANSWER_ABS"
echo ""

cd "$REPO_ROOT"
"${CMD[@]}" --answer "$ANSWER_ABS"
