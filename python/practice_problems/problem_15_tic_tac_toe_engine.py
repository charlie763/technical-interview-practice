"""
=============================================================================
INTERVIEW PROBLEM 15: Tic-Tac-Toe Engine
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You are building a configurable game engine for Tic-Tac-Toe and its
generalizations. The engine starts with a classic 3×3 board analysis tool,
evolves to efficient incremental win detection, and finally generalizes to
arbitrary board sizes and win conditions.

For this problem you are building a TicTacToeEngine class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between TicTacToeEngine
instances — avoid them.
You choose the internal data structures; the public interface is what matters.

═════════════════════════════════════════════════════════════════════════════
Part 1 — Board Analysis  (~10 min)
═════════════════════════════════════════════════════════════════════════════
Implement:
    check_winner(board: list[list[str | None]]) -> str | None

Given a square 2D board (list of lists), return the symbol of the winning
player, or None if there is no winner. A player wins by filling an entire
row, column, main diagonal, or anti-diagonal with their symbol. Empty cells
are represented by None. The board may contain any number of distinct player
symbols (not just 'X' and 'O').

# Example
# engine = TicTacToeEngine()
# engine.check_winner([
#     ['X', 'X', 'X'],
#     ['O', 'O', None],
#     [None, None, None],
# ])                            # -> 'X'
#
# engine.check_winner([
#     ['X', 'O', 'X'],
#     ['O', 'X', 'O'],
#     ['O', 'X', None],
# ])                            # -> None  (no winner)

═════════════════════════════════════════════════════════════════════════════
Part 2 — Incremental Move Tracking  (~20 min)
═════════════════════════════════════════════════════════════════════════════
Implement:
    make_move(row: int, col: int, player: str) -> str | None
    get_board() -> list[list[str | None]]
    reset() -> None

`make_move` records `player`'s move at (row, col) on the internal board and
returns the winning player's symbol if this move wins the game, otherwise None.

Optimization goal: do NOT rescan the entire board on each move. Instead,
maintain per-player counters for each row, column, and diagonal so that
a win can be detected in O(1) after every move.

Raise ValueError if the cell is already occupied or the position is out
of bounds.

`get_board` returns a deep copy of the current board as a 2D list.
`reset`  clears the board and all counters for a new game.

# Example
# engine = TicTacToeEngine()
# engine.make_move(0, 0, 'X')  # -> None
# engine.make_move(1, 0, 'O')  # -> None
# engine.make_move(0, 1, 'X')  # -> None
# engine.make_move(1, 1, 'O')  # -> None
# engine.make_move(0, 2, 'X')  # -> 'X'   (top row complete)
# engine.get_board()
# # -> [['X', 'X', 'X'], ['O', 'O', None], [None, None, None]]

═════════════════════════════════════════════════════════════════════════════
Part 3 — Arbitrary Board Size and Win Length  (~15 min)
═════════════════════════════════════════════════════════════════════════════
Extend `__init__` to accept:
    size: int = 3          — side length of the square board
    win_length: int = None — consecutive same-symbol cells required to win;
                             defaults to `size` (standard Tic-Tac-Toe rules)

When win_length < size, a player wins by placing `win_length` consecutive
symbols in any row, column, or diagonal — they do NOT need to fill the
entire row/column.

Update `make_move` to handle the general case. The O(1) counter approach
from Part 2 works when win_length == size (each counter can only reach one
target). When win_length < size, instead scan outward from the newly placed
cell in each of the 4 axis directions (horizontal, vertical, main-diagonal,
anti-diagonal), counting consecutive same-symbol cells. If the combined run
length in any axis reaches win_length, the moving player wins. This is
O(win_length) per move.

Also update `check_winner` so it detects `self._win_length` consecutive
same-symbol cells anywhere on the board, rather than requiring a full
row/column/diagonal to be filled.

Multiple players (more than 2) are naturally supported — any string is a
valid player symbol.

# Example
# engine = TicTacToeEngine(size=5, win_length=3)
# engine.make_move(2, 1, 'X')  # -> None
# engine.make_move(2, 2, 'X')  # -> None
# engine.make_move(2, 3, 'X')  # -> 'X'  (3 consecutive in row 2)
#
# # Three-player game on a 4×4 board
# engine2 = TicTacToeEngine(size=4, win_length=4)
# engine2.make_move(0, 0, 'A')  # -> None
# engine2.make_move(0, 1, 'B')  # -> None
# engine2.make_move(0, 0, 'C')  # -> ValueError (cell occupied)
=============================================================================
"""


class TicTacToeEngine:
    def __init__(self, size: int = 3, win_length: int = None):
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 1 — Board Analysis  (~10 min)
    # -------------------------------------------------------------------------

    def check_winner(self, board: list[list[str | None]]) -> str | None:
        """
        Given a square 2D board, return the winning player's symbol or None.

        Part 1: A player wins by filling an entire row, column, or diagonal.
        Part 3: Update to detect self._win_length consecutive same-symbol cells
                anywhere on the board (not necessarily a full row/column).

        Empty cells are None. Any non-None string is a valid player symbol.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 2 — Incremental Move Tracking  (~20 min)
    # -------------------------------------------------------------------------

    def make_move(self, row: int, col: int, player: str) -> str | None:
        """
        Record player's move at (row, col). Return the winning player's symbol
        if this move wins the game, otherwise None.

        Raise ValueError if the cell is already occupied or out of bounds.

        Part 2: Maintain per-player row/column/diagonal counters for O(1) win
                detection (valid when win_length == size).
        Part 3: When win_length < size, scan outward from (row, col) in each
                of the 4 axis directions to count consecutive same-symbol cells.
                Return player if any axis reaches win_length. O(win_length) per move.
        """
        raise NotImplementedError

    def get_board(self) -> list[list[str | None]]:
        """Return a deep copy of the current board state."""
        raise NotImplementedError

    def reset(self) -> None:
        """Reset the board and all counters for a new game."""
        raise NotImplementedError
