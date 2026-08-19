/**
 * Playwright global setup — runs before any tests.
 *
 * When the answer directory contains .ts or .tsx files this script runs
 * `tsc --noEmit` using the repo's tsconfig.json. A non-zero exit from tsc
 * aborts the entire test run so type errors are reported before Playwright
 * even opens a browser.
 *
 * For JSX-only answers this is a no-op.
 */
import { execSync } from 'child_process'
import { existsSync, readdirSync } from 'fs'
import path from 'path'

export default async function globalSetup() {
  const answerDir = process.env.PRACTICE_ANSWER
  if (!answerDir) return

  const fullPath = path.resolve(answerDir)
  if (!existsSync(fullPath)) return

  const files = readdirSync(fullPath)
  const hasTsFiles = files.some(f => f.endsWith('.ts') || f.endsWith('.tsx'))
  if (!hasTsFiles) return

  console.log('\n── TypeScript answer detected ─────────────────────────')
  console.log('Running tsc --noEmit …')
  // Throws on non-zero exit, which Playwright surfaces as a setup failure.
  execSync('npx tsc --noEmit', { stdio: 'inherit' })
  console.log('tsc passed ✓')
  console.log('───────────────────────────────────────────────────────\n')
}
