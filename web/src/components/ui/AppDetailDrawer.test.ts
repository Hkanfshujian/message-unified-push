import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(__dirname, 'AppDetailDrawer.vue'), 'utf-8')
const paginationSource = readFileSync(resolve(__dirname, 'AppPagination.vue'), 'utf-8')
const instanceConfigSource = readFileSync(resolve(__dirname, 'InstanceConfig.vue'), 'utf-8')

describe('AppDetailDrawer contract', () => {
  it('keeps close behavior delegated to drawer model state', () => {
    expect(source).toContain('defineModel<boolean>')
    expect(source).toContain('v-model="visible"')
    expect(source).toContain('destroy-on-close')
  })

  it('provides a single-line header, body, and footer regions', () => {
    expect(source).toContain('app-detail-drawer-header')
    expect(source).not.toContain('description')
    expect(source).toContain('app-detail-drawer-body')
    expect(source).toContain('app-detail-drawer-footer')
    expect(source).toContain('append-to-body')
  })

  it('uses the same flat header structure as form drawers', () => {
    expect(source).toContain('app-detail-drawer dora-material-overlay')
    expect(source).toContain('class="app-detail-drawer-header"')
    expect(source).not.toContain('app-detail-drawer-header dora-material-panel')
    expect(source).toContain('app-detail-drawer-body dora-material-inset')
    expect(source).toContain('app-detail-drawer-footer dora-material-panel')
  })

  it('supports shared body mode and density contracts', () => {
    expect(source).toContain("bodyMode?: 'scroll' | 'managed'")
    expect(source).toContain("density?: 'default' | 'compact'")
    expect(source).toContain("bodyMode: 'scroll'")
    expect(source).toContain("density: 'default'")
    expect(source).toContain('app-drawer-body-${bodyMode}')
    expect(source).toContain('app-drawer-density-${density}')
    expect(source).not.toContain('app-drawer-scroll-body')
  })

  it('renders compact pagination and keeps its change payload contract', () => {
    expect(paginationSource).toContain('compact?: boolean')
    expect(paginationSource).toContain('v-if="compact"')
    expect(paginationSource).toContain('{{ currentPage }} / {{ totalPages }}')
    expect(paginationSource).toContain("change: [payload: { page: number; pageSize: number }]")
    expect(instanceConfigSource).toContain('compact')
    expect(instanceConfigSource).toContain('@change="handleChannelPaginationChange"')
    expect(instanceConfigSource).toContain('v-model:current-page="channelState.currPage"')
    expect(instanceConfigSource).not.toContain('@page-change')
  })
})
