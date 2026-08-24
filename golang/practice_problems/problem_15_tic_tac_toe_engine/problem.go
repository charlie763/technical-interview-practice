// =============================================================================
// INTERVIEW PROBLEM 15: Tic-Tac-Toe Engine
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// You are building a configurable game engine for Tic-Tac-Toe and its
// generalizations. The engine starts with a classic 3×3 board analysis tool,
// evolves to efficient incremental win detection, and finally generalizes to
// arbitrary board sizes and win conditions.
//
// Store all state in struct fields. You choose the internal data structures;
// the public interface (TicTacToeEngine) is what matters.
//
// DATA MODEL
// ----------
// Board is a [][]string where "" (empty string) represents an empty cell.
// Player symbols are arbitrary non-empty strings (e.g. "X", "O", "A", "B").
//
// PART 1 — Board Analysis
// -----------------------
// check_winner(board [][]string) string
//   Given a square 2D board, return the winning player's symbol, or "" if
//   there is no winner. A player wins by filling an entire row, column, main
//   diagonal, or anti-diagonal with their symbol.
//
// PART 2 — Incremental Move Tracking
// -----------------------------------
// Maintain an internal board and per-player counters (row, column, diagonal)
// for O(1) win detection after each move.
//
// PART 3 — Arbitrary Board Size and Win Length
// ---------------------------------------------
// Constructor accepts size (default 3) and winLength (default = size).
// When winLength < size, scan outward from the placed cell to count
// consecutive same-symbol cells in each of the 4 axis directions.
//
// EXAMPLE
// -------
//   engine := NewTicTacToeEngine(3, 0)  // 0 means default win_length = size
//   engine.MakeMove(0, 0, "X")  // -> ""
//   engine.MakeMove(1, 0, "O")  // -> ""
//   engine.MakeMove(0, 1, "X")  // -> ""
//   engine.MakeMove(1, 1, "O")  // -> ""
//   engine.MakeMove(0, 2, "X")  // -> "X"   (top row complete)
//
//   // Generalized: 5×5 board, win_length=3
//   e2 := NewTicTacToeEngine(5, 3)
//   e2.MakeMove(2, 1, "X")  // -> ""
//   e2.MakeMove(2, 2, "X")  // -> ""
//   e2.MakeMove(2, 3, "X")  // -> "X"  (3 consecutive in row 2)
// =============================================================================

package tictactoe

import "errors"

// ErrOccupied is returned when a move targets an already-occupied cell.
var ErrOccupied = errors.New("cell already occupied")

// ErrOutOfBounds is returned when a move is outside the board boundaries.
var ErrOutOfBounds = errors.New("position out of bounds")

// TicTacToeEngine is the interface candidates must implement.
//
// Implement it by defining your own struct and a NewTicTacToeEngine constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myEngine struct {
//	    size      int
//	    winLength int
//	    board     [][]string
//	    // per-player counters for O(1) detection (rows, cols, diags)
//	    rows      map[string][]int
//	    cols      map[string][]int
//	    diag      map[string]int
//	    antiDiag  map[string]int
//	}
//
//	func NewTicTacToeEngine(size, winLength int) TicTacToeEngine {
//	    if winLength == 0 {
//	        winLength = size
//	    }
//	    // initialize board and counters ...
//	    return &myEngine{...}
//	}
type TicTacToeEngine interface {
	// -------------------------------------------------------------------------
	// PART 1 — Board Analysis  (~10 min)
	// -------------------------------------------------------------------------

	// CheckWinner examines the given board and returns the winning player's
	// symbol, or "" if there is no winner.
	//
	// Part 1: A player wins by filling an entire row, column, or diagonal.
	// Part 3: Update to detect self.winLength consecutive same-symbol cells
	//         anywhere on the board (not necessarily a full row/column).
	//
	// Empty cells are represented by "" (empty string).
	// Any non-empty string is a valid player symbol.
	CheckWinner(board [][]string) string

	// -------------------------------------------------------------------------
	// PART 2 — Incremental Move Tracking  (~20 min)
	// -------------------------------------------------------------------------

	// MakeMove records player's move at (row, col) on the internal board.
	// Returns the winning player's symbol if this move wins the game, or "".
	//
	// Returns ErrOccupied if the cell is already taken.
	// Returns ErrOutOfBounds if row or col is outside [0, size).
	//
	// Part 2: Maintain per-player row/column/diagonal counters for O(1) win
	//         detection (valid when winLength == size).
	// Part 3: When winLength < size, scan outward from (row, col) in each
	//         of the 4 axis directions to count consecutive same-symbol cells.
	//         Return player if any axis reaches winLength. O(winLength) per move.
	MakeMove(row, col int, player string) (string, error)

	// GetBoard returns a deep copy of the current board state.
	// Empty cells are represented by "".
	GetBoard() [][]string

	// Reset clears the board and all counters for a new game.
	Reset()
}

// NewTicTacToeEngine returns a fresh TicTacToeEngine implementation.
//
// size is the side length of the square board (default 3 when 0 is passed).
// winLength is the number of consecutive same-symbol cells required to win
// (defaults to size when 0 is passed).
//
// Candidates implement their own struct type in the answer file.
func NewTicTacToeEngine(size, winLength int) TicTacToeEngine {
	panic("not implemented")
}
