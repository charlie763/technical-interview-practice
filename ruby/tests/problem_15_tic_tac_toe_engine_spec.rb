load_problem("15_tic_tac_toe_engine")

RSpec.describe TicTacToeEngine do
  let(:engine) { described_class.new }

  # 3x3 engine with moves played but no winner yet:
  #   X . O
  #   . X .
  #   . . .
  # X is at (0,0) and (1,1) — one move away from a diagonal win.
  # O is at (0,2).
  let(:mid_game) do
    e = described_class.new
    e.make_move(0, 0, "X")
    e.make_move(0, 2, "O")
    e.make_move(1, 1, "X")
    e
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Board Analysis
  # ---------------------------------------------------------------------------
  describe "#check_winner" do
    it "returns nil for an empty board" do
      board = [[nil, nil, nil], [nil, nil, nil], [nil, nil, nil]]
      expect(engine.check_winner(board)).to be_nil
    end

    it "detects a row win" do
      board = [%w[X X X], ["O", "O", nil], [nil, nil, nil]]
      expect(engine.check_winner(board)).to eq("X")
    end

    it "detects a win on the last row" do
      board = [[nil, nil, nil], ["X", "X", nil], %w[O O O]]
      expect(engine.check_winner(board)).to eq("O")
    end

    it "detects a column win" do
      board = [["O", "X", nil], ["O", "X", nil], ["O", nil, "X"]]
      expect(engine.check_winner(board)).to eq("O")
    end

    it "detects a middle column win" do
      board = [%w[X O X], [nil, "O", nil], ["X", "O", nil]]
      expect(engine.check_winner(board)).to eq("O")
    end

    it "detects a main diagonal win" do
      board = [%w[X O O], [nil, "X", "O"], [nil, nil, "X"]]
      expect(engine.check_winner(board)).to eq("X")
    end

    it "detects an anti-diagonal win" do
      board = [%w[O O X], ["O", "X", nil], ["X", nil, nil]]
      expect(engine.check_winner(board)).to eq("X")
    end

    it "returns nil for a partial board" do
      board = [%w[X O X], %w[O X O], ["O", "X", nil]]
      expect(engine.check_winner(board)).to be_nil
    end

    it "returns nil for a full-board draw" do
      # Every row, column, and diagonal has mixed symbols
      board = [%w[X O X], %w[O X O], %w[O X O]]
      expect(engine.check_winner(board)).to be_nil
    end

    it "supports arbitrary player symbols" do
      board = [%w[A A A], ["B", "B", nil], [nil, nil, nil]]
      expect(engine.check_winner(board)).to eq("A")
    end

    it "returns the correct winner among multiple symbols" do
      # B wins the middle column
      board = [%w[A B C], %w[C B A], %w[A B C]]
      expect(engine.check_winner(board)).to eq("B")
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Incremental Move Tracking
  # ---------------------------------------------------------------------------
  describe "#make_move returning nil" do
    it "returns nil mid-game with no winner" do
      # mid_game fixture has X at (0,0),(1,1) and O at (0,2)
      expect(mid_game.make_move(2, 0, "O")).to be_nil
    end

    it "returns nil for the first move" do
      expect(engine.make_move(0, 0, "X")).to be_nil
    end
  end

  describe "#make_move winning" do
    it "detects a row win" do
      engine.make_move(0, 0, "X")
      engine.make_move(1, 0, "O")
      engine.make_move(0, 1, "X")
      engine.make_move(1, 1, "O")
      expect(engine.make_move(0, 2, "X")).to eq("X")
    end

    it "detects a column win" do
      engine.make_move(0, 0, "X")
      engine.make_move(0, 1, "O")
      engine.make_move(1, 0, "X")
      engine.make_move(0, 2, "O")
      expect(engine.make_move(2, 0, "X")).to eq("X")
    end

    it "detects a main diagonal win" do
      engine.make_move(0, 0, "X")
      engine.make_move(0, 1, "O")
      engine.make_move(1, 1, "X")
      engine.make_move(0, 2, "O")
      expect(engine.make_move(2, 2, "X")).to eq("X")
    end

    it "detects an anti-diagonal win" do
      engine.make_move(0, 2, "X")
      engine.make_move(0, 0, "O")
      engine.make_move(1, 1, "X")
      engine.make_move(0, 1, "O")
      expect(engine.make_move(2, 0, "X")).to eq("X")
    end

    it "lets the second player win" do
      engine.make_move(0, 0, "X")
      engine.make_move(1, 0, "O")
      engine.make_move(0, 1, "X")
      engine.make_move(1, 1, "O")
      engine.make_move(2, 2, "X")
      expect(engine.make_move(1, 2, "O")).to eq("O")
    end
  end

  describe "#make_move errors" do
    it "raises ArgumentError for an occupied cell" do
      engine.make_move(0, 0, "X")
      expect { engine.make_move(0, 0, "O") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a row out of bounds" do
      expect { engine.make_move(3, 0, "X") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a column out of bounds" do
      expect { engine.make_move(0, 3, "X") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a negative row" do
      expect { engine.make_move(-1, 0, "X") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a negative column" do
      expect { engine.make_move(0, -1, "X") }.to raise_error(ArgumentError)
    end
  end

  describe "#get_board" do
    it "starts with every cell nil" do
      expect(engine.get_board.flatten).to all(be_nil)
    end

    it "reflects the moves made" do
      engine.make_move(0, 0, "X")
      engine.make_move(1, 1, "O")
      board = engine.get_board
      expect(board[0][0]).to eq("X")
      expect(board[1][1]).to eq("O")
      expect(board[0][1]).to be_nil
    end

    it "returns a copy, not a reference" do
      engine.make_move(0, 0, "X")
      board = engine.get_board
      board[0][0] = "TAMPERED"
      # Internal state must be unaffected
      expect(engine.get_board[0][0]).to eq("X")
    end

    it "matches the board size" do
      board = engine.get_board
      expect(board.size).to eq(3)
      expect(board).to all(satisfy { |row| row.size == 3 })
    end
  end

  describe "#reset" do
    it "clears the board" do
      mid_game.reset
      expect(mid_game.get_board.flatten).to all(be_nil)
    end

    it "allows replaying after reset" do
      engine.make_move(0, 0, "X")
      engine.make_move(0, 1, "X")
      engine.reset
      # After reset, (0,0) should be free again
      expect(engine.make_move(0, 0, "X")).to be_nil
    end

    it "resets win detection" do
      # Win a game, reset, then win the same game again
      engine.make_move(0, 0, "X")
      engine.make_move(1, 0, "O")
      engine.make_move(0, 1, "X")
      engine.make_move(1, 1, "O")
      engine.make_move(0, 2, "X") # X wins
      engine.reset
      engine.make_move(0, 0, "X")
      engine.make_move(1, 0, "O")
      engine.make_move(0, 1, "X")
      engine.make_move(1, 1, "O")
      expect(engine.make_move(0, 2, "X")).to eq("X") # X wins again
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Arbitrary Board Size and Win Length
  # ---------------------------------------------------------------------------
  describe "arbitrary board size" do
    it "requires the full row on a 4x4 board with win_length=4" do
      e = described_class.new(size: 4, win_length: 4)
      e.make_move(0, 0, "X")
      e.make_move(0, 1, "X")
      e.make_move(0, 2, "X")
      # Three in a row on a 4x4 with win_length=4 should not win
      expect(e.make_move(1, 0, "X")).to be_nil
      # Complete the full row
      expect(e.make_move(0, 3, "X")).to eq("X")
    end

    it "wins a 5x5 board when win_length equals size" do
      e = described_class.new(size: 5, win_length: 5)
      (0...4).each { |col| expect(e.make_move(0, col, "X")).to be_nil }
      expect(e.make_move(0, 4, "X")).to eq("X")
    end

    it "returns a board matching the configured size" do
      e = described_class.new(size: 5, win_length: 3)
      board = e.get_board
      expect(board.size).to eq(5)
      expect(board).to all(satisfy { |row| row.size == 5 })
    end
  end

  describe "win_length less than size" do
    it "wins a row with a shorter run" do
      e = described_class.new(size: 5, win_length: 3)
      e.make_move(2, 1, "X")
      e.make_move(2, 2, "X")
      expect(e.make_move(2, 3, "X")).to eq("X")
    end

    it "wins a column with a shorter run" do
      e = described_class.new(size: 5, win_length: 3)
      e.make_move(1, 4, "A")
      e.make_move(2, 4, "A")
      expect(e.make_move(3, 4, "A")).to eq("A")
    end

    it "wins a diagonal with a shorter run" do
      e = described_class.new(size: 5, win_length: 3)
      e.make_move(1, 1, "B")
      e.make_move(2, 2, "B")
      expect(e.make_move(3, 3, "B")).to eq("B")
    end

    it "wins an anti-diagonal with a shorter run" do
      e = described_class.new(size: 5, win_length: 3)
      e.make_move(1, 3, "O")
      e.make_move(2, 2, "O")
      expect(e.make_move(3, 1, "O")).to eq("O")
    end

    it "does not win before the run is complete" do
      e = described_class.new(size: 5, win_length: 3)
      expect(e.make_move(2, 1, "X")).to be_nil
      expect(e.make_move(2, 2, "X")).to be_nil # only 2 in a row
    end

    it "does not win with non-consecutive cells" do
      e = described_class.new(size: 5, win_length: 3)
      e.make_move(2, 0, "X")
      e.make_move(2, 2, "X") # gap at col 1
      expect(e.make_move(2, 4, "X")).to be_nil # gap at col 3
    end

    it "does not win when the opponent breaks the run" do
      e = described_class.new(size: 5, win_length: 3)
      e.make_move(0, 0, "X")
      e.make_move(0, 1, "O") # O breaks any X run here
      expect(e.make_move(0, 2, "X")).to be_nil
    end
  end

  describe "multiple players" do
    it "picks the correct winner in a three-player game" do
      # A, B, C take turns on a 4x4 board; B wins column 1.
      # Column 0 and column 2 deliberately alternate between A and C so
      # neither of them completes a full column of their own first.
      e = described_class.new(size: 4, win_length: 4)
      moves = [
        [0, 0, "A"], [0, 1, "B"], [0, 2, "C"],
        [1, 0, "C"], [1, 1, "B"], [1, 2, "A"],
        [2, 0, "A"], [2, 1, "B"], [2, 2, "C"],
        [3, 0, "C"],
      ]
      moves.each { |r, c, p| expect(e.make_move(r, c, p)).to be_nil }
      expect(e.make_move(3, 1, "B")).to eq("B")
    end

    it "does not false-positive on a partial column with three players" do
      e = described_class.new(size: 3, win_length: 3)
      e.make_move(0, 0, "A")
      e.make_move(1, 0, "B") # interrupts A's column
      expect(e.make_move(2, 0, "A")).to be_nil
    end

    it "lets a fourth player win a diagonal" do
      e = described_class.new(size: 4, win_length: 3)
      # D wins the main diagonal starting at (1,1)
      e.make_move(0, 0, "A")
      e.make_move(0, 1, "B")
      e.make_move(0, 2, "C")
      e.make_move(1, 1, "D")
      e.make_move(0, 3, "A")
      e.make_move(2, 2, "D")
      expect(e.make_move(3, 3, "D")).to eq("D")
    end
  end

  describe "#check_winner with a custom win_length" do
    it "detects a win_length run in a row" do
      e = described_class.new(size: 5, win_length: 3)
      board = Array.new(5) { Array.new(5) }
      board[1][1] = "X"
      board[1][2] = "X"
      board[1][3] = "X"
      expect(e.check_winner(board)).to eq("X")
    end

    it "returns nil when the run is too short" do
      e = described_class.new(size: 5, win_length: 3)
      board = Array.new(5) { Array.new(5) }
      board[1][1] = "X"
      board[1][2] = "X" # only 2 in a row, need 3
      expect(e.check_winner(board)).to be_nil
    end

    it "detects a win_length run in a column" do
      e = described_class.new(size: 5, win_length: 4)
      board = Array.new(5) { Array.new(5) }
      board[0][2] = "O"
      board[1][2] = "O"
      board[2][2] = "O"
      board[3][2] = "O"
      expect(e.check_winner(board)).to eq("O")
    end

    it "does not count a non-consecutive row as a win" do
      e = described_class.new(size: 5, win_length: 3)
      board = Array.new(5) { Array.new(5) }
      board[2][0] = "X"
      board[2][2] = "X" # gap at col 1
      board[2][4] = "X" # gap at col 3
      expect(e.check_winner(board)).to be_nil
    end
  end
end
