#!/usr/bin/env bash
# run_tests.sh — run the test suite for a practice problem answer file.
#
# Usage:
#   ./run_tests.sh <path-to-answer-file> [extra pytest/jest args]
#
# Examples:
#   ./run_tests.sh python/practice_problem_answers/cw_answer_03_permission_manager.py
#   ./run_tests.sh python/practice_problem_answers/cw_answer_01_geofence_alert_engine.py -x
#
# How it works:
#   The answer filename must contain a segment matching NN_<name>
#   (e.g. "03_permission_manager"). The script:
#     1. Locates the corresponding test file (python/tests/test_problem_NN_<name>.py)
#     2. For Python files, runs pytest with --answer <abs-path>, which causes
#        conftest.py to inject your answer module in place of the problem stub
#        before any test imports happen.
#
# Language support:
#   .py         → pytest   (python/ directory)
#   .jsx/.tsx   → coming soon (jest/vitest, react/ directory)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── argument check ──────────────────────────────────────────────────────────
if [[ $# -eq 0 ]]; then
    echo "Usage: $0 <path-to-answer-file> [extra args]"
    echo ""
    echo "Examples:"
    echo "  $0 python/practice_problem_answers/cw_answer_03_permission_manager.py"
    echo "  $0 python/practice_problem_answers/cw_answer_01_geofence_alert_engine.py -x"
    exit 1
fi

ANSWER="$1"
shift  # remaining args forwarded to the test runner

if [[ ! -f "$ANSWER" ]]; then
    echo "Error: file not found: $ANSWER"
    exit 1
fi

# Absolute path — safe to use after any directory changes
ANSWER_ABS="$(cd "$(dirname "$ANSWER")" && pwd)/$(basename "$ANSWER")"
BASENAME="$(basename "$ANSWER")"
EXT="${BASENAME##*.}"

# ── dispatch by language ────────────────────────────────────────────────────
case "$EXT" in
    py)
        # Extract NN_name from filename (first match of two-digit number + snake_case word)
        PROBLEM="$(echo "$BASENAME" | grep -oE '[0-9]{2}_[a-z_]+' | head -1 || true)"
        if [[ -z "$PROBLEM" ]]; then
            echo "Error: cannot extract a problem ID from '$BASENAME'."
            echo "The filename must contain a segment like '03_permission_manager'."
            echo "Example: cw_answer_03_permission_manager.py"
            exit 1
        fi

        TEST_FILE="$REPO_ROOT/python/tests/test_problem_${PROBLEM}.py"
        if [[ ! -f "$TEST_FILE" ]]; then
            echo "Error: no test file found at:"
            echo "  $TEST_FILE"
            exit 1
        fi

        echo "Answer : $ANSWER"
        echo "Tests  : python/tests/test_problem_${PROBLEM}.py"
        echo ""
        cd "$REPO_ROOT/python"
        # Activate the repo-local virtual environment if it exists
        if [[ -f "$REPO_ROOT/.venv/bin/activate" ]]; then
            source "$REPO_ROOT/.venv/bin/activate"
        fi
        python3 -m pytest "tests/test_problem_${PROBLEM}.py" --answer "$ANSWER_ABS" -v "$@"
        ;;

    jsx|tsx|js|ts)
        echo "JavaScript/TypeScript testing is not yet implemented."
        echo "Coming soon: jest/vitest integration for react/ problems."
        exit 1
        ;;

    *)
        echo "Error: unsupported file type '.$EXT'."
        exit 1
        ;;
esac
