#!/usr/bin/env bash
# run_tests.sh — run a test command against a specific practice problem answer file.
#
# Usage:
#   ./run_tests.sh -f <path-to-answer-file> -c <test-command...>
#
# Examples:
#   ./run_tests.sh -f python/practice_problem_answers/cw_answer_03_permission_manager.py -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole
#
#   ./run_tests.sh -f python/practice_problem_answers/cw_answer_03_permission_manager.py -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole::test_empty_permissions_by_default
#
# How it works:
#   --answer <abs-path> is appended to the test command, causing conftest.py to
#   inject your answer module in place of the problem stub before test collection.

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

# ── activate venv if present ─────────────────────────────────────────────────
if [[ -f "$REPO_ROOT/.venv/bin/activate" ]]; then
    source "$REPO_ROOT/.venv/bin/activate"
fi

# ── run ──────────────────────────────────────────────────────────────────────
echo "Answer : $ANSWER"
echo "Command: ${CMD[*]} --answer $ANSWER_ABS"
echo ""

cd "$REPO_ROOT"
"${CMD[@]}" --answer "$ANSWER_ABS"
