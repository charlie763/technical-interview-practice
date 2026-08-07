"""Tests for Problem 15: Tic-Tac-Toe Engine

Run from the python/ directory:
    pytest tests/test_problem_15_tic_tac_toe_engine.py -v
"""

import copy
import pytest
from practice_problems.problem_15_tic_tac_toe_engine import TicTacToeEngine


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def engine():
    """Fresh 3×3 engine (default size and win_length)."""
    return TicTacToeEngine()


@pytest.fixture
def mid_game():
    """
    3×3 engine with moves played but no winner yet:

        X . O
        . X .
        . . .

    X is at (0,0) and (1,1) — one move away from a diagonal win.
    O is at (0,2).
    """
    e = TicTacToeEngine()
    e.make_move(0, 0, 'X')
    e.make_move(0, 2, 'O')
    e.make_move(1, 1, 'X')
    return e


# ---------------------------------------------------------------------------
# PART 1 — Board Analysis
# ---------------------------------------------------------------------------

class TestCheckWinner:
    def test_empty_board_returns_none(self, engine):
        board = [[None, None, None],
                 [None, None, None],
                 [None, None, None]]
        assert engine.check_winner(board) is None

    def test_row_win(self, engine):
        board = [['X', 'X', 'X'],
                 ['O', 'O', None],
                 [None, None, None]]
        assert engine.check_winner(board) == 'X'

    def test_last_row_win(self, engine):
        board = [[None, None, None],
                 ['X', 'X', None],
                 ['O', 'O', 'O']]
        assert engine.check_winner(board) == 'O'

    def test_column_win(self, engine):
        board = [['O', 'X', None],
                 ['O', 'X', None],
                 ['O', None, 'X']]
        assert engine.check_winner(board) == 'O'

    def test_middle_column_win(self, engine):
        board = [['X', 'O', 'X'],
                 [None, 'O', None],
                 ['X', 'O', None]]
        assert engine.check_winner(board) == 'O'

    def test_main_diagonal_win(self, engine):
        board = [['X', 'O', 'O'],
                 [None, 'X', 'O'],
                 [None, None, 'X']]
        assert engine.check_winner(board) == 'X'

    def test_anti_diagonal_win(self, engine):
        board = [['O', 'O', 'X'],
                 ['O', 'X', None],
                 ['X', None, None]]
        assert engine.check_winner(board) == 'X'

    def test_no_winner_partial_board(self, engine):
        board = [['X', 'O', 'X'],
                 ['O', 'X', 'O'],
                 ['O', 'X', None]]
        assert engine.check_winner(board) is None

    def test_no_winner_full_board_draw(self, engine):
        # Every row, column, and diagonal has mixed symbols
        board = [['X', 'O', 'X'],
                 ['O', 'X', 'O'],
                 ['O', 'X', 'O']]
        assert engine.check_winner(board) is None

    def test_arbitrary_player_symbols(self, engine):
        board = [['A', 'A', 'A'],
                 ['B', 'B', None],
                 [None, None, None]]
        assert engine.check_winner(board) == 'A'

    def test_returns_correct_winner_multiple_symbols(self, engine):
        # B wins the middle column
        board = [['A', 'B', 'C'],
                 ['C', 'B', 'A'],
                 ['A', 'B', 'C']]
        assert engine.check_winner(board) == 'B'


# ---------------------------------------------------------------------------
# PART 2 — Incremental Move Tracking
# ---------------------------------------------------------------------------

class TestMakeMoveReturnsNone:
    def test_no_winner_mid_game(self, mid_game):
        # No winner yet — mid_game fixture has X at (0,0),(1,1) and O at (0,2)
        result = mid_game.make_move(2, 0, 'O')
        assert result is None

    def test_first_move_returns_none(self, engine):
        assert engine.make_move(0, 0, 'X') is None


class TestMakeMoveWin:
    def test_row_win(self, engine):
        engine.make_move(0, 0, 'X')
        engine.make_move(1, 0, 'O')
        engine.make_move(0, 1, 'X')
        engine.make_move(1, 1, 'O')
        result = engine.make_move(0, 2, 'X')
        assert result == 'X'

    def test_column_win(self, engine):
        engine.make_move(0, 0, 'X')
        engine.make_move(0, 1, 'O')
        engine.make_move(1, 0, 'X')
        engine.make_move(0, 2, 'O')
        result = engine.make_move(2, 0, 'X')
        assert result == 'X'

    def test_main_diagonal_win(self, engine):
        engine.make_move(0, 0, 'X')
        engine.make_move(0, 1, 'O')
        engine.make_move(1, 1, 'X')
        engine.make_move(0, 2, 'O')
        result = engine.make_move(2, 2, 'X')
        assert result == 'X'

    def test_anti_diagonal_win(self, engine):
        engine.make_move(0, 2, 'X')
        engine.make_move(0, 0, 'O')
        engine.make_move(1, 1, 'X')
        engine.make_move(0, 1, 'O')
        result = engine.make_move(2, 0, 'X')
        assert result == 'X'

    def test_second_player_wins(self, engine):
        engine.make_move(0, 0, 'X')
        engine.make_move(1, 0, 'O')
        engine.make_move(0, 1, 'X')
        engine.make_move(1, 1, 'O')
        engine.make_move(2, 2, 'X')
        result = engine.make_move(1, 2, 'O')
        assert result == 'O'


class TestMakeMoveErrors:
    def test_occupied_cell_raises(self, engine):
        engine.make_move(0, 0, 'X')
        with pytest.raises(ValueError):
            engine.make_move(0, 0, 'O')

    def test_row_out_of_bounds_raises(self, engine):
        with pytest.raises(ValueError):
            engine.make_move(3, 0, 'X')

    def test_col_out_of_bounds_raises(self, engine):
        with pytest.raises(ValueError):
            engine.make_move(0, 3, 'X')

    def test_negative_row_raises(self, engine):
        with pytest.raises(ValueError):
            engine.make_move(-1, 0, 'X')

    def test_negative_col_raises(self, engine):
        with pytest.raises(ValueError):
            engine.make_move(0, -1, 'X')


class TestGetBoard:
    def test_initial_board_all_none(self, engine):
        board = engine.get_board()
        assert all(cell is None for row in board for cell in row)

    def test_reflects_moves(self, engine):
        engine.make_move(0, 0, 'X')
        engine.make_move(1, 1, 'O')
        board = engine.get_board()
        assert board[0][0] == 'X'
        assert board[1][1] == 'O'
        assert board[0][1] is None

    def test_returns_copy_not_reference(self, engine):
        engine.make_move(0, 0, 'X')
        board = engine.get_board()
        board[0][0] = 'TAMPERED'
        # Internal state must be unaffected
        assert engine.get_board()[0][0] == 'X'

    def test_board_size_is_correct(self, engine):
        board = engine.get_board()
        assert len(board) == 3
        assert all(len(row) == 3 for row in board)


class TestReset:
    def test_clears_board(self, mid_game):
        mid_game.reset()
        board = mid_game.get_board()
        assert all(cell is None for row in board for cell in row)

    def test_can_replay_after_reset(self, engine):
        engine.make_move(0, 0, 'X')
        engine.make_move(0, 1, 'X')
        engine.reset()
        # After reset, (0,0) should be free again
        result = engine.make_move(0, 0, 'X')
        assert result is None

    def test_win_detection_resets(self, engine):
        # Win a game, reset, then win the same game again
        engine.make_move(0, 0, 'X')
        engine.make_move(1, 0, 'O')
        engine.make_move(0, 1, 'X')
        engine.make_move(1, 1, 'O')
        engine.make_move(0, 2, 'X')   # X wins
        engine.reset()
        engine.make_move(0, 0, 'X')
        engine.make_move(1, 0, 'O')
        engine.make_move(0, 1, 'X')
        engine.make_move(1, 1, 'O')
        result = engine.make_move(0, 2, 'X')  # X wins again
        assert result == 'X'


# ---------------------------------------------------------------------------
# PART 3 — Arbitrary Board Size and Win Length
# ---------------------------------------------------------------------------

class TestArbitrarySize:
    def test_4x4_standard_rules(self):
        """4×4 board, win_length=4: must fill entire row."""
        e = TicTacToeEngine(size=4, win_length=4)
        e.make_move(0, 0, 'X')
        e.make_move(0, 1, 'X')
        e.make_move(0, 2, 'X')
        # Three in row on a 4×4 with win_length=4 should not win
        assert e.make_move(1, 0, 'X') is None
        # Complete the full row
        result = e.make_move(0, 3, 'X')
        assert result == 'X'

    def test_5x5_board_win_length_equals_size(self):
        e = TicTacToeEngine(size=5, win_length=5)
        for col in range(4):
            assert e.make_move(0, col, 'X') is None
        result = e.make_move(0, 4, 'X')
        assert result == 'X'

    def test_get_board_reflects_size(self):
        e = TicTacToeEngine(size=5, win_length=3)
        board = e.get_board()
        assert len(board) == 5
        assert all(len(row) == 5 for row in board)


class TestWinLengthLessThanSize:
    def test_row_win_with_win_length_3_on_5x5(self):
        e = TicTacToeEngine(size=5, win_length=3)
        e.make_move(2, 1, 'X')
        e.make_move(2, 2, 'X')
        result = e.make_move(2, 3, 'X')
        assert result == 'X'

    def test_column_win_with_win_length_3_on_5x5(self):
        e = TicTacToeEngine(size=5, win_length=3)
        e.make_move(1, 4, 'A')
        e.make_move(2, 4, 'A')
        result = e.make_move(3, 4, 'A')
        assert result == 'A'

    def test_diagonal_win_with_win_length_3_on_5x5(self):
        e = TicTacToeEngine(size=5, win_length=3)
        e.make_move(1, 1, 'B')
        e.make_move(2, 2, 'B')
        result = e.make_move(3, 3, 'B')
        assert result == 'B'

    def test_anti_diagonal_win_with_win_length_3_on_5x5(self):
        e = TicTacToeEngine(size=5, win_length=3)
        e.make_move(1, 3, 'O')
        e.make_move(2, 2, 'O')
        result = e.make_move(3, 1, 'O')
        assert result == 'O'

    def test_no_win_before_run_complete(self):
        e = TicTacToeEngine(size=5, win_length=3)
        assert e.make_move(2, 1, 'X') is None
        assert e.make_move(2, 2, 'X') is None  # only 2 in a row

    def test_non_consecutive_cells_do_not_win(self):
        """Symbols in the same row but not adjacent must not trigger a win."""
        e = TicTacToeEngine(size=5, win_length=3)
        e.make_move(2, 0, 'X')
        e.make_move(2, 2, 'X')  # gap at col 1
        result = e.make_move(2, 4, 'X')  # gap at col 3
        assert result is None

    def test_partial_run_filled_by_opponent_does_not_win(self):
        """A run of win_length interrupted by another player's symbol must not win."""
        e = TicTacToeEngine(size=5, win_length=3)
        e.make_move(0, 0, 'X')
        e.make_move(0, 1, 'O')  # O breaks any X run here
        result = e.make_move(0, 2, 'X')
        assert result is None


class TestMultiplePlayers:
    def test_three_player_game_correct_winner(self):
        """A, B, C take turns on a 4×4 board; B wins column 1."""
        e = TicTacToeEngine(size=4, win_length=4)
        moves = [
            (0, 0, 'A'), (0, 1, 'B'), (0, 2, 'C'),
            (1, 0, 'A'), (1, 1, 'B'), (1, 2, 'C'),
            (2, 0, 'A'), (2, 1, 'B'), (2, 2, 'C'),
            (3, 0, 'A'),
        ]
        for r, c, p in moves:
            assert e.make_move(r, c, p) is None
        result = e.make_move(3, 1, 'B')
        assert result == 'B'

    def test_three_player_game_no_false_positive(self):
        """With 3 players, a partial column for any single player must not win."""
        e = TicTacToeEngine(size=3, win_length=3)
        e.make_move(0, 0, 'A')
        e.make_move(1, 0, 'B')  # interrupts A's column
        result = e.make_move(2, 0, 'A')
        assert result is None

    def test_fourth_player_wins_diagonal(self):
        e = TicTacToeEngine(size=4, win_length=3)
        # D wins the main diagonal starting at (1,1)
        e.make_move(0, 0, 'A')
        e.make_move(0, 1, 'B')
        e.make_move(0, 2, 'C')
        e.make_move(1, 1, 'D')
        e.make_move(0, 3, 'A')
        e.make_move(2, 2, 'D')
        result = e.make_move(3, 3, 'D')
        assert result == 'D'


class TestCheckWinnerPart3:
    def test_detects_win_length_run_in_row(self):
        e = TicTacToeEngine(size=5, win_length=3)
        board = [[None] * 5 for _ in range(5)]
        board[1][1] = 'X'
        board[1][2] = 'X'
        board[1][3] = 'X'
        assert e.check_winner(board) == 'X'

    def test_no_winner_when_run_too_short(self):
        e = TicTacToeEngine(size=5, win_length=3)
        board = [[None] * 5 for _ in range(5)]
        board[1][1] = 'X'
        board[1][2] = 'X'   # only 2 in a row, need 3
        assert e.check_winner(board) is None

    def test_detects_win_length_run_in_column(self):
        e = TicTacToeEngine(size=5, win_length=4)
        board = [[None] * 5 for _ in range(5)]
        board[0][2] = 'O'
        board[1][2] = 'O'
        board[2][2] = 'O'
        board[3][2] = 'O'
        assert e.check_winner(board) == 'O'

    def test_non_consecutive_row_is_not_a_win(self):
        e = TicTacToeEngine(size=5, win_length=3)
        board = [[None] * 5 for _ in range(5)]
        board[2][0] = 'X'
        board[2][2] = 'X'   # gap at col 1
        board[2][4] = 'X'   # gap at col 3
        assert e.check_winner(board) is None
