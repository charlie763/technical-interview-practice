import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import fs from 'fs'

export default defineConfig(() => {
  const answerDir = process.env.PRACTICE_ANSWER

  return {
    plugins: [
      react(),
      answerDir && {
        name: 'practice-answer-redirect',
        enforce: 'pre',
        resolveId(id, importer) {
          if (
            importer?.endsWith('/src/main.jsx') &&
            (id === './App.jsx' || id === './App.tsx' || id === './App')
          ) {
            // Prefer App.tsx for TypeScript problems; fall back to App.jsx
            const tsxPath = path.resolve(process.cwd(), answerDir, 'App.tsx')
            if (fs.existsSync(tsxPath)) return tsxPath
            return path.resolve(process.cwd(), answerDir, 'App.jsx')
          }
        },
      },
    ].filter(Boolean),
    server: {
      port: 5173,
    },
  }
})
