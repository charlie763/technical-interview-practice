import { defineConfig } from 'vitest/config'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const answerName = process.env.PRACTICE_ANSWER // e.g. 'cw_answer_01_donation_processor'

/**
 * When PRACTICE_ANSWER is set, this plugin intercepts the import of the
 * corresponding practice_problems stub file and redirects it to the answer
 * file in practice_problem_answers/, so tests run against the implementation
 * without any changes to the test files themselves.
 *
 * Usage:
 *   PRACTICE_ANSWER=cw_answer_01_donation_processor npm run test:01
 */
function makeAnswerRedirectPlugin(answer: string) {
  const problemName = answer.replace(/^cw_answer_/, 'problem_')
  const problemAbs = path.resolve(__dirname, 'practice_problems', problemName)
  const answerAbs = path.resolve(__dirname, 'practice_problem_answers', answer + '.ts')

  return {
    name: 'answer-redirect',
    resolveId(id: string, importer?: string) {
      if (!importer || !id.startsWith('.')) return undefined
      const resolved = path.resolve(path.dirname(importer), id)
      if (resolved === problemAbs || resolved === problemAbs + '.ts') {
        return answerAbs
      }
      return undefined
    },
  }
}

export default defineConfig({
  plugins: answerName ? [makeAnswerRedirectPlugin(answerName)] : [],
  test: {
    include: ['tests/**/*.test.ts'],
  },
})
