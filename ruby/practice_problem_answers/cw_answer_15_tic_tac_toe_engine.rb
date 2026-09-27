class TicTacToeEngine
  DIRECTIONS = [[0, 1], [1, 0], [1, 1], [1, -1]].freeze

  def initialize(size: 3, win_length: nil)
    @size = size
    @win_length = win_length || size
    reset
  end

  def check_winner(board)
    size = board.length
    (0...size).each do |r|
      (0...size).each do |c|
        symbol = board[r][c]
        next if symbol.nil?

        DIRECTIONS.each do |dr, dc|
          count = 1
          rr, cc = r + dr, c + dc
          while rr.between?(0, size - 1) && cc.between?(0, size - 1) && board[rr][cc] == symbol
            count += 1
            break if count >= @win_length

            rr += dr
            cc += dc
          end
          return symbol if count >= @win_length
        end
      end
    end
    nil
  end

  def make_move(row, col, player)
    unless row.between?(0, @size - 1) && col.between?(0, @size - 1)
      raise ArgumentError, "position out of bounds: (#{row}, #{col})"
    end
    raise ArgumentError, "cell already occupied: (#{row}, #{col})" unless @board[row][col].nil?

    @board[row][col] = player

    won = DIRECTIONS.any? do |dr, dc|
      run_length(row, col, dr, dc, player) + run_length(row, col, -dr, -dc, player) + 1 >= @win_length
    end
    won ? player : nil
  end

  def get_board
    @board.map(&:dup)
  end

  def reset
    @board = Array.new(@size) { Array.new(@size) }
  end

  private

  def run_length(row, col, dr, dc, player)
    count = 0
    rr, cc = row + dr, col + dc
    while rr.between?(0, @size - 1) && cc.between?(0, @size - 1) && @board[rr][cc] == player
      count += 1
      rr += dr
      cc += dc
    end
    count
  end
end
