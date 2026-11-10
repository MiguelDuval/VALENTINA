import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import realRouter from '../../router'
import OpportunitiesView from '../OpportunitiesView.vue'
import LeadsView from '../LeadsView.vue'
import CompaniesView from '../CompaniesView.vue'
import SourcesView from '../SourcesView.vue'
import ResearchView from '../ResearchView.vue'
import { apiFetch } from '../../composables/useApi'

// Vitest hoists vi.mock() above normal module declarations; keep its factory dependencies in vi.hoisted().
const { apiFetchMock, baseMock } = vi.hoisted(() => {
  const baseMock = async (path: string) => {
    if (path.includes('/opportunities')) return { opportunities: [], count: 0 }
    if (path.includes('/leads')) return { leads: [], count: 0 }
    if (path.includes('/companies')) return { companies: [], count: 0 }
    if (path.includes('/sources')) return { sources: [], count: 0 }
    if (path.includes('/research-runs')) return { runs: [], count: 0 }
    return {}
  }
  return { apiFetchMock: vi.fn(baseMock), baseMock }
})

vi.mock('../../composables/useApi', () => ({
  apiFetch: apiFetchMock,
}))

const mockedApiFetch = vi.mocked(apiFetch)

const expectedRoutes: { path: string; component: object }[] = [
  { path: '/valentina/opportunities', component: OpportunitiesView },
  { path: '/valentina/leads', component: LeadsView },
  { path: '/valentina/companies', component: CompaniesView },
  { path: '/valentina/sources', component: SourcesView },
  { path: '/valentina/research', component: ResearchView },
]

describe('CRM Routes (real application router)', () => {
  it('registers all five CRM routes under /valentina/', () => {
    const routes = realRouter.getRoutes()
    for (const expected of expectedRoutes) {
      const route = routes.find(r => r.path === expected.path)
      expect(route).toBeDefined()
      expect(route?.components?.default).toBe(expected.component)
    }
  })

  it('preserves all Huginn routes', () => {
    const routes = realRouter.getRoutes()
    const huginnPaths = [
      '/chat/:sessionId?', '/space/:spaceId', '/agents/:agentName?', '/models/:provider?',
      '/connections', '/settings', '/logs', '/stats', '/inbox', '/memory',
      '/routines', '/workflows/:id?', '/cloud', '/skills/:tab?',
    ]
    for (const p of huginnPaths) {
      expect(routes.find(r => r.path === p)).toBeDefined()
    }
  })

  it('has root redirect to /overview', () => {
    const routes = realRouter.getRoutes()
    const root = routes.find(r => r.path === '/')
    expect(root).toBeDefined()
    expect(root?.redirect).toBe('/overview')
  })

  it('has catch-all at the end', () => {
    const routes = realRouter.getRoutes()
    const last = routes[routes.length - 1]
    expect(last.path).toBe('/:pathMatch(.*)*')
  })

  it('does not catch CRM routes with catch-all', () => {
    const routes = realRouter.getRoutes()
    const crmPaths = routes.filter(r => r.path.startsWith('/valentina/')).map(r => r.path)
    expect(crmPaths).toHaveLength(5)
    for (const p of crmPaths) {
      expect(p).toMatch(/\/valentina\/(opportunities|leads|companies|sources|research)$/)
    }
    const catchAll = routes.find(r => r.path === '/:pathMatch(.*)*')
    expect(catchAll).toBeDefined()
    const catchAllIndex = routes.indexOf(catchAll!)
    for (const p of crmPaths) {
      const idx = routes.indexOf(routes.find(r => r.path === p)!)
      expect(idx).toBeLessThan(catchAllIndex)
    }
  })
})

// Mount tests use the real application router.
const mountRouter = realRouter
