import { fileURLToPath } from "node:url";
import path from "node:path";
import fs from "node:fs";
import { defineConfig } from "vitest/config";

const root = path.dirname(fileURLToPath(import.meta.url));
const problemsDir = path.join(root, "practice_problems");
const answersDir = path.join(root, "practice_problem_answers");

/**
 * Mirrors python/conftest.py's --answer flag, adapted to Vitest's alias
 * resolution. Test files always import a fixed stub path, e.g.:
 *
 *   import { twoSum } from "@problems/problem_01_two_sum";
 *
 * When run_tests.sh sets the ANSWER_FILE env var, we redirect that exact
 * specifier to the candidate's answer file instead of the stub — no files
 * are copied or mutated, so an interrupted run can't leave the stub
 * clobbered.
 *
 * The answer filename must contain a segment matching NN_<name>
 * (e.g. cw_answer_01_two_sum.ts -> redirects "@problems/problem_01_two_sum").
 */
function answerFileAlias() {
  const answerPath = process.env.ANSWER_FILE;
  if (!answerPath) return [];

  const absAnswerPath = path.resolve(answerPath);
  if (!fs.existsSync(absAnswerPath)) {
    throw new Error(`ANSWER_FILE not found: ${absAnswerPath}`);
  }

  const filename = path.basename(absAnswerPath);
  const match = filename.match(/(\d{2}_[a-z_]+)\.ts$/);
  if (!match) {
    throw new Error(
      `ANSWER_FILE '${filename}' does not match the expected naming pattern.\n` +
        `Filename must contain a segment like '01_two_sum'.\n` +
        `Example: cw_answer_01_two_sum.ts`,
    );
  }

  return [
    {
      find: new RegExp(`^@problems/problem_${match[1]}$`),
      replacement: absAnswerPath,
    },
  ];
}

export default defineConfig({
  test: {
    environment: "node",
    globals: true,
    // Order matters: the answer-file override (exact match) must come
    // before the general @problems/ prefix alias so it wins for that one
    // specifier while every other import still resolves to the stub dir.
    alias: [...answerFileAlias(), { find: /^@problems\//, replacement: `${problemsDir}/` }, { find: /^@answers\//, replacement: `${answersDir}/` }],
  },
});
