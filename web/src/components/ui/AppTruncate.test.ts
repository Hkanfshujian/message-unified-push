import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(__dirname, 'AppTruncate.vue'), 'utf-8')

describe('AppTruncate display contract', () => {
  it('renders truncated text as normal text by default', () => {
    expect(source).toContain('preview: false')
    expect(source).toContain('text-foreground')
    expect(source).not.toContain('cursor-pointer text-brand')
  })

  it('uses accessible button semantics only when preview is enabled', () => {
    expect(source).toContain('if (props.preview) visible.value = true')
    expect(source).toContain('v-if="preview"')
    expect(source).toContain('type="button"')
    expect(source).toContain(':aria-label="`预览${title}：${displayText}`"')
    expect(source).toContain('<span v-else')
  })
})
