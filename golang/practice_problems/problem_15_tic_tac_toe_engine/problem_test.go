// Tests for Problem 15: Tic-Tac-Toe Engine
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_15_tic_tac_toe_engine/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_15_tic_tac_toe_engine.go \
//	  -c go test -v .
package tictactoe

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newEngine returns a default 3×3 engine with win_length=3.
func newEngine(t *testing.T) TicTacToeEngine {
	t.Helper()
	return NewTicTacToeEngine(3, 0)
}

// midGame returns a 3×3 engine with moves played but no winner yet:
//
//	X . O
//	. X .
//	. . .
//
// X is at (0,0) and (1,1) — one move away from a diagonal win.
// O is at (0,2).
func midGame(t *testing.T) TicTacToeEngine {
	t.Helper()
	e := NewTicTacToeEngine(3, 0)
	mustMove(t, e, 0, 0, "X")
	mustMove(t, e, 0, 2, "O")
	mustMove(t, e, 1, 1, "X")
	return e
}

func mustMove(t *testing.T, e TicTacToeEngine, row, col int, player string) string {
	t.Helper()
	result, err := e.MakeMove(row, col, player)
	if err != nil {
		t.Fatalf("MakeMove(%d,%d,%q): unexpected error: %v", row, col, player, err)
	}
	return result
}

// board3 builds a 3×3 board from 9 strings (left-to-right, top-to-bottom).
// Use "" for empty cells.
func board3(cells ...string) [][]string {
	b := [][]string{
		{cells[0], cells[1], cells[2]},
		{cells[3], cells[4], cells[5]},
		{cells[6], cells[7], cells[8]},
	}
	return b
}

// ---------------------------------------------------------------------------
// PART 1 — Board Analysis
// ---------------------------------------------------------------------------

func TestCheckWinner(t *testing.T) {
	tests := []struct {
		name  string
		board [][]string
		want  string
	}{
		{
			"empty_board",
			board3("", "", "", "", "", "", "", "", ""),
			"",
		},
		{
			"row_win_X",
			board3("X", "X", "X", "O", "O", "", "", "", ""),
			"X",
		},
		{
			"last_row_win_O",
			board3("", "", "", "X", "X", "", "O", "O", "O"),
			"O",
		},
		{
			"column_win_O",
			board3("O", "X", "", "O", "X", "", "O", "", "X"),
			"O",
		},
		{
			"middle_column_win_O",
			board3("X", "O", "X", "", "O", "", "X", "O", ""),
			"O",
		},
		{
			"main_diagonal_win_X",
			board3("X", "O", "O", "", "X", "O", "", "", "X"),
			"X",
		},
		{
			"anti_diagonal_win_X",
			board3("O", "O", "X", "O", "X", "", "X", "", ""),
			"X",
		},
		{
			"no_winner_partial",
			board3("X", "O", "X", "O", "X", "O", "O", "X", ""),
			"",
		},
		{
			"no_winner_full_draw",
			board3("X", "O", "X", "O", "X", "O", "O", "X", "O"),
			"",
		},
		{
			"arbitrary_player_symbols",
			board3("A", "A", "A", "B", "B", "", "", "", ""),
			"A",
		},
		{
			"multi_symbol_B_wins_middle_column",
			board3("A", "B", "C", "C", "B", "A", "A", "B", "C"),
			"B",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEngine(t)
			got := e.CheckWinner(tc.board)
			if got != tc.want {
				t.Errorf("CheckWinner = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// PART 2 — Incremental Move Tracking
// ---------------------------------------------------------------------------

func TestMakeMoveReturnsEmpty(t *testing.T) {
	t.Run("no_winner_mid_game", func(t *testing.T) {
		e := midGame(t)
		result := mustMove(t, e, 2, 0, "O")
		if result != "" {
			t.Errorf("MakeMove = %q, want ''", result)
		}
	})

	t.Run("first_move_returns_empty", func(t *testing.T) {
		e := newEngine(t)
		result := mustMove(t, e, 0, 0, "X")
		if result != "" {
			t.Errorf("MakeMove = %q, want ''", result)
		}
	})
}

func TestMakeMoveWin(t *testing.T) {
	t.Run("row_win", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 1, 0, "O")
		mustMove(t, e, 0, 1, "X")
		mustMove(t, e, 1, 1, "O")
		result := mustMove(t, e, 0, 2, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("column_win", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 0, 1, "O")
		mustMove(t, e, 1, 0, "X")
		mustMove(t, e, 0, 2, "O")
		result := mustMove(t, e, 2, 0, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("main_diagonal_win", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 0, 1, "O")
		mustMove(t, e, 1, 1, "X")
		mustMove(t, e, 0, 2, "O")
		result := mustMove(t, e, 2, 2, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("anti_diagonal_win", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 2, "X")
		mustMove(t, e, 0, 0, "O")
		mustMove(t, e, 1, 1, "X")
		mustMove(t, e, 0, 1, "O")
		result := mustMove(t, e, 2, 0, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("second_player_wins", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 1, 0, "O")
		mustMove(t, e, 0, 1, "X")
		mustMove(t, e, 1, 1, "O")
		mustMove(t, e, 2, 2, "X")
		result := mustMove(t, e, 1, 2, "O")
		if result != "O" {
			t.Errorf("MakeMove = %q, want 'O'", result)
		}
	})
}

func TestMakeMoveErrors(t *testing.T) {
	t.Run("occupied_cell_returns_error", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		_, err := e.MakeMove(0, 0, "O")
		if !errors.Is(err, ErrOccupied) {
			t.Errorf("expected ErrOccupied, got %v", err)
		}
	})

	t.Run("row_out_of_bounds_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.MakeMove(3, 0, "X")
		if !errors.Is(err, ErrOutOfBounds) {
			t.Errorf("expected ErrOutOfBounds, got %v", err)
		}
	})

	t.Run("col_out_of_bounds_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.MakeMove(0, 3, "X")
		if !errors.Is(err, ErrOutOfBounds) {
			t.Errorf("expected ErrOutOfBounds, got %v", err)
		}
	})

	t.Run("negative_row_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.MakeMove(-1, 0, "X")
		if !errors.Is(err, ErrOutOfBounds) {
			t.Errorf("expected ErrOutOfBounds, got %v", err)
		}
	})

	t.Run("negative_col_returns_error", func(t *testing.T) {
		e := newEngine(t)
		_, err := e.MakeMove(0, -1, "X")
		if !errors.Is(err, ErrOutOfBounds) {
			t.Errorf("expected ErrOutOfBounds, got %v", err)
		}
	})
}

func TestGetBoard(t *testing.T) {
	t.Run("initial_board_all_empty", func(t *testing.T) {
		e := newEngine(t)
		board := e.GetBoard()
		for r, row := range board {
			for c, cell := range row {
				if cell != "" {
					t.Errorf("board[%d][%d] = %q, want ''", r, c, cell)
				}
			}
		}
	})

	t.Run("reflects_moves", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 1, 1, "O")
		board := e.GetBoard()
		if board[0][0] != "X" {
			t.Errorf("board[0][0] = %q, want 'X'", board[0][0])
		}
		if board[1][1] != "O" {
			t.Errorf("board[1][1] = %q, want 'O'", board[1][1])
		}
		if board[0][1] != "" {
			t.Errorf("board[0][1] = %q, want ''", board[0][1])
		}
	})

	t.Run("returns_copy_not_reference", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		board := e.GetBoard()
		board[0][0] = "TAMPERED"
		// Internal state must be unaffected
		if e.GetBoard()[0][0] != "X" {
			t.Error("internal board was mutated through GetBoard return value")
		}
	})

	t.Run("board_size_is_correct", func(t *testing.T) {
		e := newEngine(t)
		board := e.GetBoard()
		if len(board) != 3 {
			t.Errorf("len(board) = %d, want 3", len(board))
		}
		for i, row := range board {
			if len(row) != 3 {
				t.Errorf("len(board[%d]) = %d, want 3", i, len(row))
			}
		}
	})
}

func TestReset(t *testing.T) {
	t.Run("clears_board", func(t *testing.T) {
		e := midGame(t)
		e.Reset()
		board := e.GetBoard()
		for r, row := range board {
			for c, cell := range row {
				if cell != "" {
					t.Errorf("board[%d][%d] = %q after reset, want ''", r, c, cell)
				}
			}
		}
	})

	t.Run("can_replay_after_reset", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 0, 1, "X")
		e.Reset()
		// After reset, (0,0) should be free again
		result := mustMove(t, e, 0, 0, "X")
		if result != "" {
			t.Errorf("MakeMove after reset = %q, want ''", result)
		}
	})

	t.Run("win_detection_resets", func(t *testing.T) {
		e := newEngine(t)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 1, 0, "O")
		mustMove(t, e, 0, 1, "X")
		mustMove(t, e, 1, 1, "O")
		mustMove(t, e, 0, 2, "X") // X wins
		e.Reset()
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 1, 0, "O")
		mustMove(t, e, 0, 1, "X")
		mustMove(t, e, 1, 1, "O")
		result := mustMove(t, e, 0, 2, "X") // X wins again
		if result != "X" {
			t.Errorf("MakeMove after reset = %q, want 'X'", result)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Arbitrary Board Size and Win Length
// ---------------------------------------------------------------------------

func TestArbitrarySize(t *testing.T) {
	t.Run("4x4_standard_rules", func(t *testing.T) {
		// 4×4 board, win_length=4: must fill entire row
		e := NewTicTacToeEngine(4, 4)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 0, 1, "X")
		mustMove(t, e, 0, 2, "X")
		// Three in a row on 4×4 with win_length=4 should not win
		if result := mustMove(t, e, 1, 0, "X"); result != "" {
			t.Errorf("three in a row should not win on 4×4 with win_length=4, got %q", result)
		}
		// Complete the full row
		result := mustMove(t, e, 0, 3, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("5x5_board_win_length_equals_size", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 5)
		for col := 0; col < 4; col++ {
			if result := mustMove(t, e, 0, col, "X"); result != "" {
				t.Errorf("partial row should not win, got %q at col %d", result, col)
			}
		}
		result := mustMove(t, e, 0, 4, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("get_board_reflects_size", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		board := e.GetBoard()
		if len(board) != 5 {
			t.Errorf("len(board) = %d, want 5", len(board))
		}
		for i, row := range board {
			if len(row) != 5 {
				t.Errorf("len(board[%d]) = %d, want 5", i, len(row))
			}
		}
	})
}

func TestWinLengthLessThanSize(t *testing.T) {
	t.Run("row_win_with_win_length_3_on_5x5", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		mustMove(t, e, 2, 1, "X")
		mustMove(t, e, 2, 2, "X")
		result := mustMove(t, e, 2, 3, "X")
		if result != "X" {
			t.Errorf("MakeMove = %q, want 'X'", result)
		}
	})

	t.Run("column_win_with_win_length_3_on_5x5", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		mustMove(t, e, 1, 4, "A")
		mustMove(t, e, 2, 4, "A")
		result := mustMove(t, e, 3, 4, "A")
		if result != "A" {
			t.Errorf("MakeMove = %q, want 'A'", result)
		}
	})

	t.Run("diagonal_win_with_win_length_3_on_5x5", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		mustMove(t, e, 1, 1, "B")
		mustMove(t, e, 2, 2, "B")
		result := mustMove(t, e, 3, 3, "B")
		if result != "B" {
			t.Errorf("MakeMove = %q, want 'B'", result)
		}
	})

	t.Run("anti_diagonal_win_with_win_length_3_on_5x5", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		mustMove(t, e, 1, 3, "O")
		mustMove(t, e, 2, 2, "O")
		result := mustMove(t, e, 3, 1, "O")
		if result != "O" {
			t.Errorf("MakeMove = %q, want 'O'", result)
		}
	})

	t.Run("no_win_before_run_complete", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		if result := mustMove(t, e, 2, 1, "X"); result != "" {
			t.Errorf("one cell should not win, got %q", result)
		}
		if result := mustMove(t, e, 2, 2, "X"); result != "" {
			t.Errorf("two in a row should not win with win_length=3, got %q", result)
		}
	})

	t.Run("non_consecutive_cells_do_not_win", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		mustMove(t, e, 2, 0, "X")
		mustMove(t, e, 2, 2, "X") // gap at col 1
		result := mustMove(t, e, 2, 4, "X") // gap at col 3
		if result != "" {
			t.Errorf("non-consecutive symbols should not win, got %q", result)
		}
	})

	t.Run("run_interrupted_by_opponent_does_not_win", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		mustMove(t, e, 0, 0, "X")
		mustMove(t, e, 0, 1, "O") // O breaks any X run here
		result := mustMove(t, e, 0, 2, "X")
		if result != "" {
			t.Errorf("interrupted run should not win, got %q", result)
		}
	})
}

func TestMultiplePlayers(t *testing.T) {
	t.Run("three_player_game_correct_winner", func(t *testing.T) {
		// A, B, C take turns on a 4×4 board; B wins column 1
		e := NewTicTacToeEngine(4, 4)
		moves := [][3]interface{}{
			{0, 0, "A"}, {0, 1, "B"}, {0, 2, "C"},
			{1, 0, "A"}, {1, 1, "B"}, {1, 2, "C"},
			{2, 0, "A"}, {2, 1, "B"}, {2, 2, "C"},
			{3, 0, "A"},
		}
		for _, m := range moves {
			result := mustMove(t, e, m[0].(int), m[1].(int), m[2].(string))
			if result != "" {
				t.Errorf("expected no winner yet, got %q", result)
			}
		}
		result := mustMove(t, e, 3, 1, "B")
		if result != "B" {
			t.Errorf("MakeMove = %q, want 'B'", result)
		}
	})

	t.Run("three_player_no_false_positive", func(t *testing.T) {
		e := NewTicTacToeEngine(3, 3)
		mustMove(t, e, 0, 0, "A")
		mustMove(t, e, 1, 0, "B") // interrupts A's column
		result := mustMove(t, e, 2, 0, "A")
		if result != "" {
			t.Errorf("partial column for A should not win when interrupted, got %q", result)
		}
	})

	t.Run("fourth_player_wins_diagonal", func(t *testing.T) {
		e := NewTicTacToeEngine(4, 3)
		// D wins the main diagonal starting at (1,1)
		mustMove(t, e, 0, 0, "A")
		mustMove(t, e, 0, 1, "B")
		mustMove(t, e, 0, 2, "C")
		mustMove(t, e, 1, 1, "D")
		mustMove(t, e, 0, 3, "A")
		mustMove(t, e, 2, 2, "D")
		result := mustMove(t, e, 3, 3, "D")
		if result != "D" {
			t.Errorf("MakeMove = %q, want 'D'", result)
		}
	})
}

func TestCheckWinnerPart3(t *testing.T) {
	t.Run("detects_win_length_run_in_row", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		board := make([][]string, 5)
		for i := range board {
			board[i] = make([]string, 5)
		}
		board[1][1] = "X"
		board[1][2] = "X"
		board[1][3] = "X"
		if got := e.CheckWinner(board); got != "X" {
			t.Errorf("CheckWinner = %q, want 'X'", got)
		}
	})

	t.Run("no_winner_when_run_too_short", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		board := make([][]string, 5)
		for i := range board {
			board[i] = make([]string, 5)
		}
		board[1][1] = "X"
		board[1][2] = "X" // only 2 in a row, need 3
		if got := e.CheckWinner(board); got != "" {
			t.Errorf("CheckWinner = %q, want '' (run too short)", got)
		}
	})

	t.Run("detects_win_length_run_in_column", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 4)
		board := make([][]string, 5)
		for i := range board {
			board[i] = make([]string, 5)
		}
		board[0][2] = "O"
		board[1][2] = "O"
		board[2][2] = "O"
		board[3][2] = "O"
		if got := e.CheckWinner(board); got != "O" {
			t.Errorf("CheckWinner = %q, want 'O'", got)
		}
	})

	t.Run("non_consecutive_row_is_not_a_win", func(t *testing.T) {
		e := NewTicTacToeEngine(5, 3)
		board := make([][]string, 5)
		for i := range board {
			board[i] = make([]string, 5)
		}
		board[2][0] = "X"
		board[2][2] = "X" // gap at col 1
		board[2][4] = "X" // gap at col 3
		if got := e.CheckWinner(board); got != "" {
			t.Errorf("CheckWinner = %q, want '' (non-consecutive)", got)
		}
	})
}
