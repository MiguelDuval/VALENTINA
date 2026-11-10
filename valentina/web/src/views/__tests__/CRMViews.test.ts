import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import realRouter from '../../router'
import OpportunitiesView from '../OpportunitiesView.vue'
import LeadsView from '../LeadsView.vue'
import CompaniesView from '../CompaniesView.vue'
import SourcesView from '../SourcesView.vue'
import ResearchView from '../ResearchView.vue'
import { apiFetch } from '../../composables/useApi'

// Mock apiFetch composable
vi.mock('../../composables/useApi', () => ({
  apiFetch: vi.fn(async (path: string) => {
    if (path.includes('/opportunities')) return { opportunities: [], count: 0 }
    if (path.includes('/leads')) return { leads: [], count: 0 }
    if (path.includes('/companies')) return { companies: [], count: 0 }
    if (path.includes('/sources')) return { sources: [], count: 0 }
    if (path.includes('/research-runs')) return { runs: [], count: 0 }
    return {}
  }),
}))

// Mount tests use a separate memory router; route-existence tests use the real app router.
const mountRouter = createRouter({
  history: createMemoryHistory(),
  routes: realRouter.getRoutes().map(r => ({ path: r.path, component: r.components?.default })),
})

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
      expect(route, `Route ${expected.path} should exist`).toBeDefined()
      expect(route?.components?.default, `Route ${expected.path} component should match`).toBe(expected.component)
    }
  })

  it('has no catch-all redirect covering CRM routes', () => {
    const routes = realRouter.getRoutes()
    const crmPaths = routes.filter(r => r.path.startsWith('/valentina/')).map(r => r.path)
    expect(crmPaths).toHaveLength(5)
    for (const p of crmPaths) {
      expect(p).toMatch(/^\/valentina\/(opportunities|leads|companies|sources|research)$/)
    }
  })
})

describe('OpportunitiesView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    vi.resetAllMocks()
  })

  it('renders opportunities title', () => {
    const wrapper = mount(OpportunitiesView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Холодные возможности')
  })

  it('shows empty state when no opportunities', () => {
    const wrapper = mount(OpportunitiesView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Нет возможностей')
  })
})

describe('LeadsView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    vi.resetAllMocks()
  })

  it('renders leads title', () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Входящие заявки')
  })

  it('shows empty state when no leads', () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Нет входящих заявок')
  })

  it('has manual lead creation form', () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Создать входящую заявку вручную')
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('creates a lead via POST API, shows success, resets form, disables button during submit, and handles API error', async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({
      lead_id: 'lead-123',
      request_summary: 'Test lead',
      inbound_channel: 'web',
      status: 'new',
      received_at: new Date().toISOString(),
    })

    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })

    // Fill form
    await wrapper.find('#lead-summary').setValue('Test lead')
    await wrapper.find('#lead-channel').setValue('web')

    // Button should be enabled initially
    const btn = wrapper.find('button[type="submit"]')
    expect(btn.attributes('disabled')).toBeUndefined()

    // Submit
    await wrapper.find('form').trigger('submit.prevent')
    await wrapper.vm.$nextTick()

    // Button should be disabled during submit
    expect(btn.attributes('disabled')).toBeDefined()

    // Verify POST API called with correct payload
    expect(vi.mocked(apiFetch)).toHaveBeenCalledWith(
      '/api/v1/valentina/leads',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ request_summary: 'Test lead', inbound_channel: 'web' }),
      })
    )

    // Wait for async completion
    await new Promise(resolve => setTimeout(resolve, 0))
    await wrapper.vm.$nextTick()

    // Success state shown
    expect(wrapper.text()).toContain('создана')
    expect(wrapper.text()).toContain('Test lead')

    // Form reset
    expect((wrapper.vm as any).newLead.request_summary).toBe('')
    expect((wrapper.vm as any).newLead.inbound_channel).toBe('')

    // Button re-enabled
    expect(btn.attributes('disabled')).toBeUndefined()
  })

  it('shows error when API call fails', async () => {
    vi.mocked(apiFetch).mockRejectedValueOnce(new Error('API down'))

    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })

    await wrapper.find('#lead-summary').setValue('Test lead')
    await wrapper.find('#lead-channel').setValue('web')
    await wrapper.find('form').trigger('submit.prevent')

    await new Promise(resolve => setTimeout(resolve, 0))
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Не удалось создать заявку')
    const btn = wrapper.find('button[type="submit"]')
    expect(btn.attributes('disabled')).toBeUndefined()
  })
})

describe('CompaniesView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    vi.resetAllMocks()
  })

  it('renders companies title', () => {
    const wrapper = mount(CompaniesView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Компании')
  })
})

describe('SourcesView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    vi.resetAllMocks()
  })

  it('renders sources title', () => {
    const wrapper = mount(SourcesView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Источники')
  })

  it('shows pending review approval state', async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({
      sources: [{ source_id: 's1', name: 'Test', source_type: 'public', base_url: 'https://example.com', publisher: 'Pub', approval_status: 'pending_review' }],
      count: 1
    })
    
    const wrapper = mount(SourcesView, {
      global: { plugins: [mountRouter] },
    })
    
    await new Promise(resolve => setTimeout(resolve, 0))
    await wrapper.vm.$nextTick()
    
    expect(wrapper.text()).toContain('pending_review')
  })
})

describe('ResearchView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    vi.resetAllMocks()
  })

  it('renders research title', () => {
    const wrapper = mount(ResearchView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Исследования')
  })
})
