/**
 * =============================================================================
 * INTERVIEW PROBLEM 15: Tic-Tac-Toe Engine
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You are building a configurable game engine for Tic-Tac-Toe and its
 * generalizations. The engine starts with a classic 3×3 board analysis tool,
 * evolves to efficient incremental win detection, and finally generalizes to
 * arbitrary board sizes and win conditions.
 *
 * For this problem you are building a TicTacToeEngine class.
 * Store all state in instance properties initialized in the constructor.
 * Class-level (static) fields will bleed between tests and between
 * TicTacToeEngine instances — avoid them.
 * You choose the internal data structures; the public interface is what matters.
 *
 * ═════════════════════════════════════════════════════════════════════════
 * Part 1 — Board Analysis  (~10 min)
 * ═════════════════════════════════════════════════════════════════════════
 * Implement:
 *     checkWinner(board: (string | undefined)[][]): string | undefined
 *
 * Given a square 2D board (array of arrays), return the symbol of the winning
 * player, or undefined if there is no winner. A player wins by filling an
 * entire row, column, main diagonal, or anti-diagonal with their symbol.
 * Empty cells are represented by undefined. The board may contain any number
 * of distinct player symbols (not just 'X' and 'O').
 *
 * # Example
 * const engine = new TicTacToeEngine();
 * engine.checkWinner([
 *   ['X', 'X', 'X'],
 *   ['O', 'O', undefined],
 *   [undefined, undefined, undefined],
 * ]);                            // -> 'X'
 *
 * engine.checkWinner([
 *   ['X', 'O', 'X'],
 *   ['O', 'X', 'O'],
 *   ['O', 'X', undefined],
 * ]);                            // -> undefined  (no winner)
 *
 * ═════════════════════════════════════════════════════════════════════════
 * Part 2 — Incremental Move Tracking  (~20 min)
 * ═════════════════════════════════════════════════════════════════════════
 * Implement:
 *     makeMove(row: number, col: number, player: string): string | undefined
 *     getBoard(): (string | undefined)[][]
 *     reset(): void
 *
 * `makeMove` records `player`'s move at (row, col) on the internal board and
 * returns the winning player's symbol if this move wins the game, otherwise
 * undefined.
 *
 * Optimization goal: do NOT rescan the entire board on each move. Instead,
 * maintain per-player counters for each row, column, and diagonal so that
 * a win can be detected in O(1) after every move.
 *
 * Throw an Error if the cell is already occupied or the position is out
 * of bounds.
 *
 * `getBoard` returns a deep copy of the current board as a 2D array.
 * `reset` clears the board and all counters for a new game.
 *
 * # Example
 * const engine = new TicTacToeEngine();
 * engine.makeMove(0, 0, 'X');  // -> undefined
 * engine.makeMove(1, 0, 'O');  // -> undefined
 * engine.makeMove(0, 1, 'X');  // -> undefined
 * engine.makeMove(1, 1, 'O');  // -> undefined
 * engine.makeMove(0, 2, 'X');  // -> 'X'   (top row complete)
 * engine.getBoard();
 * // -> [['X', 'X', 'X'], ['O', 'O', undefined], [undefined, undefined, undefined]]
 *
 * ═════════════════════════════════════════════════════════════════════════
 * Part 3 — Arbitrary Board Size and Win Length  (~15 min)
 * ═════════════════════════════════════════════════════════════════════════
 * Extend the constructor to accept:
 *     size: number = 3            — side length of the square board
 *     winLength?: number          — consecutive same-symbol cells required to
 *                                    win; defaults to `size` (standard rules)
 *
 * When winLength < size, a player wins by placing `winLength` consecutive
 * symbols in any row, column, or diagonal — they do NOT need to fill the
 * entire row/column.
 *
 * Update `makeMove` to handle the general case. The O(1) counter approach
 * from Part 2 works when winLength === size (each counter can only reach one
 * target). When winLength < size, instead scan outward from the newly placed
 * cell in each of the 4 axis directions (horizontal, vertical, main-diagonal,
 * anti-diagonal), counting consecutive same-symbol cells. If the combined run
 * length in any axis reaches winLength, the moving player wins. This is
 * O(winLength) per move.
 *
 * Also update `checkWinner` so it detects `winLength` consecutive
 * same-symbol cells anywhere on the board, rather than requiring a full
 * row/column/diagonal to be filled.
 *
 * Multiple players (more than 2) are naturally supported — any string is a
 * valid player symbol.
 *
 * # Example
 * const engine = new TicTacToeEngine(5, 3);
 * engine.makeMove(2, 1, 'X');  // -> undefined
 * engine.makeMove(2, 2, 'X');  // -> undefined
 * engine.makeMove(2, 3, 'X');  // -> 'X'  (3 consecutive in row 2)
 *
 * // Three-player game on a 4×4 board
 * const engine2 = new TicTacToeEngine(4, 4);
 * engine2.makeMove(0, 0, 'A');  // -> undefined
 * engine2.makeMove(0, 1, 'B');  // -> undefined
 * engine2.makeMove(0, 0, 'C');  // -> throws (cell occupied)
 * =============================================================================
 */

export class TicTacToeEngine {
  constructor(size: number = 3, winLength?: number) {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — Board Analysis  (~10 min)
  // ---------------------------------------------------------------------------

  /**
   * Given a square 2D board, return the winning player's symbol or undefined.
   *
   * Part 1: A player wins by filling an entire row, column, or diagonal.
   * Part 3: Update to detect `winLength` consecutive same-symbol cells
   *         anywhere on the board (not necessarily a full row/column).
   *
   * Empty cells are undefined. Any non-undefined string is a valid player symbol.
   */
  checkWinner(board: (string | undefined)[][]): string | undefined {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — Incremental Move Tracking  (~20 min)
  // ---------------------------------------------------------------------------

  /**
   * Record player's move at (row, col). Return the winning player's symbol
   * if this move wins the game, otherwise undefined.
   *
   * Throw an Error if the cell is already occupied or out of bounds.
   *
   * Part 2: Maintain per-player row/column/diagonal counters for O(1) win
   *         detection (valid when winLength === size).
   * Part 3: When winLength < size, scan outward from (row, col) in each
   *         of the 4 axis directions to count consecutive same-symbol cells.
   *         Return player if any axis reaches winLength. O(winLength) per move.
   */
  makeMove(row: number, col: number, player: string): string | undefined {
    throw new Error("Not implemented");
  }

  /** Return a deep copy of the current board state. */
  getBoard(): (string | undefined)[][] {
    throw new Error("Not implemented");
  }

  /** Reset the board and all counters for a new game. */
  reset(): void {
    throw new Error("Not implemented");
  }
}
