/**
 * Tests for Problem 15: Tic-Tac-Toe Engine
 *
 * Run from the typescript/ directory:
 *   npm run test:15
 */

import { describe, expect, it } from "vitest";
import { TicTacToeEngine } from "@problems/problem_15_tic_tac_toe_engine";

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Fresh 3×3 engine (default size and winLength). */
function freshEngine(): TicTacToeEngine {
  return new TicTacToeEngine();
}

/**
 * 3×3 engine with moves played but no winner yet:
 *
 *     X . O
 *     . X .
 *     . . .
 *
 * X is at (0,0) and (1,1) — one move away from a diagonal win.
 * O is at (0,2).
 */
function midGame(): TicTacToeEngine {
  const e = new TicTacToeEngine();
  e.makeMove(0, 0, "X");
  e.makeMove(0, 2, "O");
  e.makeMove(1, 1, "X");
  return e;
}

// ---------------------------------------------------------------------------
// PART 1 — Board Analysis
// ---------------------------------------------------------------------------

describe("checkWinner", () => {
  it("empty board returns undefined", () => {
    const engine = freshEngine();
    const board = [
      [undefined, undefined, undefined],
      [undefined, undefined, undefined],
      [undefined, undefined, undefined],
    ];
    expect(engine.checkWinner(board)).toBeUndefined();
  });

  it("row win", () => {
    const engine = freshEngine();
    const board = [
      ["X", "X", "X"],
      ["O", "O", undefined],
      [undefined, undefined, undefined],
    ];
    expect(engine.checkWinner(board)).toBe("X");
  });

  it("last row win", () => {
    const engine = freshEngine();
    const board = [
      [undefined, undefined, undefined],
      ["X", "X", undefined],
      ["O", "O", "O"],
    ];
    expect(engine.checkWinner(board)).toBe("O");
  });

  it("column win", () => {
    const engine = freshEngine();
    const board = [
      ["O", "X", undefined],
      ["O", "X", undefined],
      ["O", undefined, "X"],
    ];
    expect(engine.checkWinner(board)).toBe("O");
  });

  it("middle column win", () => {
    const engine = freshEngine();
    const board = [
      ["X", "O", "X"],
      [undefined, "O", undefined],
      ["X", "O", undefined],
    ];
    expect(engine.checkWinner(board)).toBe("O");
  });

  it("main diagonal win", () => {
    const engine = freshEngine();
    const board = [
      ["X", "O", "O"],
      [undefined, "X", "O"],
      [undefined, undefined, "X"],
    ];
    expect(engine.checkWinner(board)).toBe("X");
  });

  it("anti-diagonal win", () => {
    const engine = freshEngine();
    const board = [
      ["O", "O", "X"],
      ["O", "X", undefined],
      ["X", undefined, undefined],
    ];
    expect(engine.checkWinner(board)).toBe("X");
  });

  it("no winner partial board", () => {
    const engine = freshEngine();
    const board = [
      ["X", "O", "X"],
      ["O", "X", "O"],
      ["O", "X", undefined],
    ];
    expect(engine.checkWinner(board)).toBeUndefined();
  });

  it("no winner full board draw", () => {
    // Every row, column, and diagonal has mixed symbols
    const engine = freshEngine();
    const board = [
      ["X", "O", "X"],
      ["O", "X", "O"],
      ["O", "X", "O"],
    ];
    expect(engine.checkWinner(board)).toBeUndefined();
  });

  it("arbitrary player symbols", () => {
    const engine = freshEngine();
    const board = [
      ["A", "A", "A"],
      ["B", "B", undefined],
      [undefined, undefined, undefined],
    ];
    expect(engine.checkWinner(board)).toBe("A");
  });

  it("returns correct winner with multiple symbols", () => {
    // B wins the middle column
    const engine = freshEngine();
    const board = [
      ["A", "B", "C"],
      ["C", "B", "A"],
      ["A", "B", "C"],
    ];
    expect(engine.checkWinner(board)).toBe("B");
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Incremental Move Tracking
// ---------------------------------------------------------------------------

describe("makeMove returns undefined when no winner", () => {
  it("no winner mid-game", () => {
    // No winner yet — midGame fixture has X at (0,0),(1,1) and O at (0,2)
    const result = midGame().makeMove(2, 0, "O");
    expect(result).toBeUndefined();
  });

  it("first move returns undefined", () => {
    expect(freshEngine().makeMove(0, 0, "X")).toBeUndefined();
  });
});

describe("makeMove win detection", () => {
  it("row win", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(1, 0, "O");
    engine.makeMove(0, 1, "X");
    engine.makeMove(1, 1, "O");
    const result = engine.makeMove(0, 2, "X");
    expect(result).toBe("X");
  });

  it("column win", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(0, 1, "O");
    engine.makeMove(1, 0, "X");
    engine.makeMove(0, 2, "O");
    const result = engine.makeMove(2, 0, "X");
    expect(result).toBe("X");
  });

  it("main diagonal win", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(0, 1, "O");
    engine.makeMove(1, 1, "X");
    engine.makeMove(0, 2, "O");
    const result = engine.makeMove(2, 2, "X");
    expect(result).toBe("X");
  });

  it("anti-diagonal win", () => {
    const engine = freshEngine();
    engine.makeMove(0, 2, "X");
    engine.makeMove(0, 0, "O");
    engine.makeMove(1, 1, "X");
    engine.makeMove(0, 1, "O");
    const result = engine.makeMove(2, 0, "X");
    expect(result).toBe("X");
  });

  it("second player wins", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(1, 0, "O");
    engine.makeMove(0, 1, "X");
    engine.makeMove(1, 1, "O");
    engine.makeMove(2, 2, "X");
    const result = engine.makeMove(1, 2, "O");
    expect(result).toBe("O");
  });
});

describe("makeMove errors", () => {
  it("occupied cell throws", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    expect(() => engine.makeMove(0, 0, "O")).toThrow();
  });

  it("row out of bounds throws", () => {
    const engine = freshEngine();
    expect(() => engine.makeMove(3, 0, "X")).toThrow();
  });

  it("col out of bounds throws", () => {
    const engine = freshEngine();
    expect(() => engine.makeMove(0, 3, "X")).toThrow();
  });

  it("negative row throws", () => {
    const engine = freshEngine();
    expect(() => engine.makeMove(-1, 0, "X")).toThrow();
  });

  it("negative col throws", () => {
    const engine = freshEngine();
    expect(() => engine.makeMove(0, -1, "X")).toThrow();
  });
});

describe("getBoard", () => {
  it("initial board all undefined", () => {
    const engine = freshEngine();
    const board = engine.getBoard();
    expect(board.every((row) => row.every((cell) => cell === undefined))).toBe(true);
  });

  it("reflects moves", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(1, 1, "O");
    const board = engine.getBoard();
    expect(board[0][0]).toBe("X");
    expect(board[1][1]).toBe("O");
    expect(board[0][1]).toBeUndefined();
  });

  it("returns copy not reference", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    const board = engine.getBoard();
    board[0][0] = "TAMPERED";
    // Internal state must be unaffected
    expect(engine.getBoard()[0][0]).toBe("X");
  });

  it("board size is correct", () => {
    const engine = freshEngine();
    const board = engine.getBoard();
    expect(board).toHaveLength(3);
    expect(board.every((row) => row.length === 3)).toBe(true);
  });
});

describe("reset", () => {
  it("clears board", () => {
    const engine = midGame();
    engine.reset();
    const board = engine.getBoard();
    expect(board.every((row) => row.every((cell) => cell === undefined))).toBe(true);
  });

  it("can replay after reset", () => {
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(0, 1, "X");
    engine.reset();
    // After reset, (0,0) should be free again
    const result = engine.makeMove(0, 0, "X");
    expect(result).toBeUndefined();
  });

  it("win detection resets", () => {
    // Win a game, reset, then win the same game again
    const engine = freshEngine();
    engine.makeMove(0, 0, "X");
    engine.makeMove(1, 0, "O");
    engine.makeMove(0, 1, "X");
    engine.makeMove(1, 1, "O");
    engine.makeMove(0, 2, "X"); // X wins
    engine.reset();
    engine.makeMove(0, 0, "X");
    engine.makeMove(1, 0, "O");
    engine.makeMove(0, 1, "X");
    engine.makeMove(1, 1, "O");
    const result = engine.makeMove(0, 2, "X"); // X wins again
    expect(result).toBe("X");
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Arbitrary Board Size and Win Length
// ---------------------------------------------------------------------------

describe("arbitrary size", () => {
  it("4x4 standard rules (must fill entire row)", () => {
    const e = new TicTacToeEngine(4, 4);
    e.makeMove(0, 0, "X");
    e.makeMove(0, 1, "X");
    e.makeMove(0, 2, "X");
    // Three in row on a 4×4 with winLength=4 should not win
    expect(e.makeMove(1, 0, "X")).toBeUndefined();
    // Complete the full row
    const result = e.makeMove(0, 3, "X");
    expect(result).toBe("X");
  });

  it("5x5 board winLength equals size", () => {
    const e = new TicTacToeEngine(5, 5);
    for (let col = 0; col < 4; col++) {
      expect(e.makeMove(0, col, "X")).toBeUndefined();
    }
    const result = e.makeMove(0, 4, "X");
    expect(result).toBe("X");
  });

  it("getBoard reflects size", () => {
    const e = new TicTacToeEngine(5, 3);
    const board = e.getBoard();
    expect(board).toHaveLength(5);
    expect(board.every((row) => row.length === 5)).toBe(true);
  });
});

describe("winLength less than size", () => {
  it("row win with winLength 3 on 5x5", () => {
    const e = new TicTacToeEngine(5, 3);
    e.makeMove(2, 1, "X");
    e.makeMove(2, 2, "X");
    const result = e.makeMove(2, 3, "X");
    expect(result).toBe("X");
  });

  it("column win with winLength 3 on 5x5", () => {
    const e = new TicTacToeEngine(5, 3);
    e.makeMove(1, 4, "A");
    e.makeMove(2, 4, "A");
    const result = e.makeMove(3, 4, "A");
    expect(result).toBe("A");
  });

  it("diagonal win with winLength 3 on 5x5", () => {
    const e = new TicTacToeEngine(5, 3);
    e.makeMove(1, 1, "B");
    e.makeMove(2, 2, "B");
    const result = e.makeMove(3, 3, "B");
    expect(result).toBe("B");
  });

  it("anti-diagonal win with winLength 3 on 5x5", () => {
    const e = new TicTacToeEngine(5, 3);
    e.makeMove(1, 3, "O");
    e.makeMove(2, 2, "O");
    const result = e.makeMove(3, 1, "O");
    expect(result).toBe("O");
  });

  it("no win before run complete", () => {
    const e = new TicTacToeEngine(5, 3);
    expect(e.makeMove(2, 1, "X")).toBeUndefined();
    expect(e.makeMove(2, 2, "X")).toBeUndefined(); // only 2 in a row
  });

  it("non-consecutive cells do not win", () => {
    // Symbols in the same row but not adjacent must not trigger a win.
    const e = new TicTacToeEngine(5, 3);
    e.makeMove(2, 0, "X");
    e.makeMove(2, 2, "X"); // gap at col 1
    const result = e.makeMove(2, 4, "X"); // gap at col 3
    expect(result).toBeUndefined();
  });

  it("partial run filled by opponent does not win", () => {
    // A run of winLength interrupted by another player's symbol must not win.
    const e = new TicTacToeEngine(5, 3);
    e.makeMove(0, 0, "X");
    e.makeMove(0, 1, "O"); // O breaks any X run here
    const result = e.makeMove(0, 2, "X");
    expect(result).toBeUndefined();
  });
});

describe("multiple players", () => {
  it("three-player game correct winner", () => {
    // A, B, C take turns on a 4×4 board; B wins column 1.
    const e = new TicTacToeEngine(4, 4);
    const moves: [number, number, string][] = [
      [0, 0, "A"],
      [0, 1, "B"],
      [0, 2, "C"],
      [1, 0, "A"],
      [1, 1, "B"],
      [1, 2, "C"],
      [2, 0, "A"],
      [2, 1, "B"],
      [2, 2, "C"],
      [3, 2, "A"],
    ];
    for (const [r, c, p] of moves) {
      expect(e.makeMove(r, c, p)).toBeUndefined();
    }
    const result = e.makeMove(3, 1, "B");
    expect(result).toBe("B");
  });

  it("three-player game no false positive", () => {
    // With 3 players, a partial column for any single player must not win.
    const e = new TicTacToeEngine(3, 3);
    e.makeMove(0, 0, "A");
    e.makeMove(1, 0, "B"); // interrupts A's column
    const result = e.makeMove(2, 0, "A");
    expect(result).toBeUndefined();
  });

  it("fourth player wins diagonal", () => {
    const e = new TicTacToeEngine(4, 3);
    // D wins the main diagonal starting at (1,1)
    e.makeMove(0, 0, "A");
    e.makeMove(0, 1, "B");
    e.makeMove(0, 2, "C");
    e.makeMove(1, 1, "D");
    e.makeMove(0, 3, "A");
    e.makeMove(2, 2, "D");
    const result = e.makeMove(3, 3, "D");
    expect(result).toBe("D");
  });
});

describe("checkWinner (Part 3)", () => {
  it("detects winLength run in row", () => {
    const e = new TicTacToeEngine(5, 3);
    const board: (string | undefined)[][] = Array.from({ length: 5 }, () => Array(5).fill(undefined));
    board[1][1] = "X";
    board[1][2] = "X";
    board[1][3] = "X";
    expect(e.checkWinner(board)).toBe("X");
  });

  it("no winner when run too short", () => {
    const e = new TicTacToeEngine(5, 3);
    const board: (string | undefined)[][] = Array.from({ length: 5 }, () => Array(5).fill(undefined));
    board[1][1] = "X";
    board[1][2] = "X"; // only 2 in a row, need 3
    expect(e.checkWinner(board)).toBeUndefined();
  });

  it("detects winLength run in column", () => {
    const e = new TicTacToeEngine(5, 4);
    const board: (string | undefined)[][] = Array.from({ length: 5 }, () => Array(5).fill(undefined));
    board[0][2] = "O";
    board[1][2] = "O";
    board[2][2] = "O";
    board[3][2] = "O";
    expect(e.checkWinner(board)).toBe("O");
  });

  it("non-consecutive row is not a win", () => {
    const e = new TicTacToeEngine(5, 3);
    const board: (string | undefined)[][] = Array.from({ length: 5 }, () => Array(5).fill(undefined));
    board[2][0] = "X";
    board[2][2] = "X"; // gap at col 1
    board[2][4] = "X"; // gap at col 3
    expect(e.checkWinner(board)).toBeUndefined();
  });
});
