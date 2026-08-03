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
# ── React problems ────────────────────────────────────────────────────────────
# Usage:
#   ./run_tests.sh -f <path-to-answer.jsx> -c <npm-test-command...>
#
# Examples:
#   ./run_tests.sh \
#     -f react/practice_problems/problem_02_incident_dashboard.jsx \
#     -c npm run test:02
#
#   ./run_tests.sh \
#     -f react/practice_problems/problem_02_incident_dashboard.jsx \
#     -c npm run test:ui
#
# How it works (React):
#   The answer file is copied to react/src/App.jsx (which Vite serves), then
#   the Playwright command runs from the react/ directory. Vite is started
#   automatically by Playwright if it isn't already running.
#   react/src/App.jsx is restored to the placeholder after tests finish.

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

if [[ ! -f "$ANSWER" ]]; then
    echo "Error: file not found: $ANSWER"
    exit 1
fi

ANSWER_ABS="$(cd "$(dirname "$ANSWER")" && pwd)/$(basename "$ANSWER")"

# ── React mode: .jsx / .tsx answer files ─────────────────────────────────────
if [[ "$ANSWER_ABS" == *.jsx || "$ANSWER_ABS" == *.tsx ]]; then
    REACT_APP="$REPO_ROOT/react/src/App.jsx"
    PLACEHOLDER="$REPO_ROOT/react/src/App.jsx.bak"

    # Back up current App.jsx so we can restore it when done
    cp "$REACT_APP" "$PLACEHOLDER"

    cleanup() {
        cp "$PLACEHOLDER" "$REACT_APP"
        rm -f "$PLACEHOLDER"
    }
    trap cleanup EXIT

    cp "$ANSWER_ABS" "$REACT_APP"

    echo "Answer : $ANSWER → react/src/App.jsx"
    echo "Command: ${CMD[*]}"
    echo ""

    cd "$REPO_ROOT/react"
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
