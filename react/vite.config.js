import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

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
            (id === './App.jsx' || id === './App')
          ) {
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
