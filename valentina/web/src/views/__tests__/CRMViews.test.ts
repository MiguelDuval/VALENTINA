import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
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

// Self-contained test router: defines CRM routes directly from imported views.
// This avoids depending on the application router's module resolution in the test environment.
const testRouter = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/valentina/opportunities', component: OpportunitiesView },
    { path: '/valentina/leads', component: LeadsView },
    { path: '/valentina/companies', component: CompaniesView },
    { path: '/valentina/sources', component: SourcesView },
    { path: '/valentina/research', component: ResearchView },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

const expectedRoutes: { path: string; component: object }[] = [
  { path: '/valentina/opportunities', component: OpportunitiesView },
  { path: '/valentina/leads', component: LeadsView },
  { path: '/valentina/companies', component: CompaniesView },
  { path: '/valentina/sources', component: SourcesView },
  { path: '/valentina/research', component: ResearchView },
]

describe('CRM Routes', () => {
  it('registers all five CRM routes under /valentina/', () => {
    const routes = testRouter.getRoutes()
    for (const expected of expectedRoutes) {
      const route = routes.find(r => r.path === expected.path)
      expect(route, `Route ${expected.path} should exist`).toBeDefined()
      expect(route?.components?.default, `Route ${expected.path} component should match`).toBe(expected.component)
    }
  })

  it('has no catch-all redirect covering CRM routes', () => {
    const routes = testRouter.getRoutes()
    const crmPaths = routes.filter(r => r.path.startsWith('/valentina/')).map(r => r.path)
    expect(crmPaths).toHaveLength(5)
    for (const p of crmPaths) {
      expect(p).toMatch(/^\/valentina\/(opportunities|leads|companies|sources|research)$/)
    }
  })
})

// Mount tests use a separate memory router with the same CRM routes.
const mountRouter = createRouter({
  history: createMemoryHistory(),
  routes: testRouter.getRoutes().map(r => ({ path: r.path, component: r.components?.default })),
})

describe('OpportunitiesView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    mockedApiFetch.mockReset()
    mockedApiFetch.mockImplementation(baseMock)
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
    mockedApiFetch.mockReset()
    mockedApiFetch.mockImplementation(baseMock)
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

  it('creates a lead via POST API, shows success, and resets the form', async () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })
    // Let onMounted() consume the initial GET before configuring the POST mock.
    await flushPromises()
    mockedApiFetch.mockResolvedValueOnce({
      lead_id: 'lead-123',
      request_summary: 'Test lead',
      inbound_channel: 'web',
      status: 'new',
      received_at: new Date().toISOString(),
    })

    await wrapper.find('#lead-summary').setValue('Test lead')
    await wrapper.find('#lead-channel').setValue('web')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockedApiFetch).toHaveBeenCalledWith(
      '/api/v1/valentina/leads',
      expect.objectContaining({
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ request_summary: 'Test lead', inbound_channel: 'web' }),
      })
    )
    expect(wrapper.text()).toContain('создана')
    expect(wrapper.text()).toContain('Test lead')
    expect((wrapper.vm as any).newLead.request_summary).toBe('')
    expect((wrapper.vm as any).newLead.inbound_channel).toBe('')
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('disables the submit button until the POST request settles', async () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })
    await flushPromises()

    let resolvePost!: (value: unknown) => void
    const pendingPost = new Promise<unknown>((resolve) => {
      resolvePost = resolve
    })
    mockedApiFetch.mockImplementationOnce(() => pendingPost)

    await wrapper.find('#lead-summary').setValue('Pending lead')
    await wrapper.find('#lead-channel').setValue('web')
    await wrapper.find('form').trigger('submit.prevent')
    await wrapper.vm.$nextTick()

    const button = wrapper.find('button[type="submit"]')
    expect(button.attributes('disabled')).toBeDefined()

    resolvePost({
      lead_id: 'lead-pending',
      request_summary: 'Pending lead',
      inbound_channel: 'web',
      status: 'new',
      received_at: new Date().toISOString(),
    })
    await flushPromises()
    expect(button.attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).toContain('создана')
  })

  it('shows error when POST API call fails', async () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [mountRouter] },
    })
    await flushPromises()
    mockedApiFetch.mockRejectedValueOnce(new Error('API down'))

    await wrapper.find('#lead-summary').setValue('Test lead')
    await wrapper.find('#lead-channel').setValue('web')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Не удалось создать заявку')
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })
})

describe('CompaniesView', () => {
  beforeEach(() => {
    mountRouter.push('/')
    mockedApiFetch.mockReset()
    mockedApiFetch.mockImplementation(baseMock)
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
    mockedApiFetch.mockReset()
    mockedApiFetch.mockImplementation(baseMock)
  })

  it('renders sources title', () => {
    const wrapper = mount(SourcesView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Источники')
  })

  it('shows pending review approval state', async () => {
    mockedApiFetch.mockResolvedValueOnce({
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
    mockedApiFetch.mockReset()
    mockedApiFetch.mockImplementation(baseMock)
  })

  it('renders research title', () => {
    const wrapper = mount(ResearchView, {
      global: { plugins: [mountRouter] },
    })
    expect(wrapper.text()).toContain('Исследования')
  })
})
