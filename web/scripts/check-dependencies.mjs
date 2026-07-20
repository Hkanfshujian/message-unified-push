import { readFileSync } from 'node:fs'

const pkg = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'))
const dependencies = { ...pkg.dependencies, ...pkg.devDependencies }

const required = {
  vue: '3.5.30',
  typescript: '~5.8.3',
  vite: '~7.1.0',
  'element-plus': '~2.14.2',
  unocss: '66.7.2',
  sass: '~1.89.2',
  pinia: '~2.3.1',
  echarts: '~6.0.0',
  '@antv/g2': '~5.3.0',
  '@wangeditor/editor': '~5.1.23'
}

const forbidden = [
  'node-sass',
  'vue-sonner',
  'vue3-apexcharts',
  'apexcharts',
  '@tailwindcss/vite',
  '@tanstack/vue-table',
  'class-variance-authority',
  'reka-ui',
  'tailwind-merge',
  'tailwindcss',
  'tw-animate-css',
  'vaul-vue'
]
const failures = []

for (const [name, expected] of Object.entries(required)) {
  if (dependencies[name] !== expected) {
    failures.push(`${name}: expected ${expected}, got ${dependencies[name] || 'missing'}`)
  }
}

for (const name of forbidden) {
  if (dependencies[name]) {
    failures.push(`${name}: forbidden dependency is present`)
  }
}

if (failures.length > 0) {
  console.error(failures.join('\n'))
  process.exit(1)
}

console.log('Dependency contract satisfied.')
