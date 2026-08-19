/** @type {import('jest').Config} */

const answerName = process.env.PRACTICE_ANSWER // e.g. 'cw_answer_01_donation_processor'

/**
 * When PRACTICE_ANSWER is set, Jest redirects imports of the corresponding
 * practice_problems stub to the answer file in practice_problem_answers/,
 * so tests run against the implementation without any changes to the test files.
 *
 * Usage (from typescript/):
 *   PRACTICE_ANSWER=cw_answer_01_donation_processor npm run test:01
 *
 * Or via run_tests.sh from the repo root:
 *   ./run_tests.sh \
 *     -f typescript/practice_problem_answers/cw_answer_01_donation_processor.ts \
 *     -c npm run test:01
 */
function buildModuleNameMapper() {
  if (!answerName) return {}
  const problemName = answerName.replace(/^cw_answer_/, 'problem_')
  return {
    [`.*practice_problems/${problemName}(\\.ts)?$`]:
      `<rootDir>/practice_problem_answers/${answerName}.ts`,
  }
}

module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  testMatch: ['**/tests/**/*.test.ts'],
  moduleNameMapper: buildModuleNameMapper(),
}
