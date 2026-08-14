# TypeScript — Agent Guidelines

## Setup

```bash
cd typescript
npm install
```

## Workflow per problem

1. The problem stub is in `practice_problems/problem_NN_<name>.ts` — **do not edit it**.
2. Copy the stub to `practice_problem_answers/` and implement it there:

   ```bash
   cp practice_problems/problem_01_donation_processor.ts \
      practice_problem_answers/cw_answer_01_donation_processor.ts
   ```

3. Run tests against your answer:

   ```bash
   PRACTICE_ANSWER=cw_answer_01_donation_processor npm run test:01
   ```

   Or watch mode:

   ```bash
   PRACTICE_ANSWER=cw_answer_01_donation_processor npm run test:watch -- tests/test_problem_01_donation_processor.test.ts
   ```

When `PRACTICE_ANSWER` is set, `vitest.config.ts` intercepts the import of the
stub file and transparently redirects it to your answer file — no changes to the
test files are ever needed.

Running `npm test` (without `PRACTICE_ANSWER`) runs all tests against the stubs,
which should all fail with "Not implemented". This is the expected baseline.

## Test commands

| Command | Problem |
|---|---|
| `npm run test:01` | Problem 01 — Donation Processor |
| `npm run test:02` | Problem 02 — Walkathon Pledge Tracker |
| `npm run test:03` | Problem 03 — Donor Suppressor |
| `npm run test:04` | Problem 04 — Platform Fee Calculator |
| `npm run test:05` | Problem 05 — Giving Day Challenge Engine |

Prefix with `PRACTICE_ANSWER=cw_answer_NN_<name>` to test your implementation.

---

## Problem design rules

### Parts must be self-contained
- Tests for Part N must only call methods defined in Parts 1–N.
- Never verify a Part 1 result using a Part 2 helper.

### Methods must compose
Design later-part methods to **call** earlier-part methods rather than duplicating
logic. If a Part 1 method returns a value a Part 3 method also needs, it should call
the Part 1 method — not re-implement the same check.

### Class-based problems
All problems use a single class with instance variables initialised in the
constructor. The candidate chooses internal data structures; the public interface
is what matters.

### Stubs use `throw new Error('Not implemented')`
Every stub method throws. Running tests against the stub → all tests fail.
Running tests against a correct implementation → all tests pass.

---

## Test writing rules

### Use `beforeEach` fixtures, never inline construction
```typescript
// BAD
it('does something', () => {
  const dp = new DonationProcessor()
  // ...
})

// GOOD
describe('Part 1', () => {
  let dp: DonationProcessor
  beforeEach(() => { dp = new DonationProcessor() })
  it('does something', () => { ... })
})
```

### Pre-seeded fixtures for Parts 2+
Create a second `beforeEach` block for Parts 2 and 3 that builds a realistic
pre-seeded instance, so candidates don't need to re-set-up data in every test.

### Cross-part independence
A Part 1 test must never fail because Part 2 is not implemented. If a Part 1
assertion requires a method from Part 2, move that method to Part 1.
