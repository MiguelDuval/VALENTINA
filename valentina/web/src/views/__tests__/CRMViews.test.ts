import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
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

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/opportunities', component: OpportunitiesView },
    { path: '/leads', component: LeadsView },
    { path: '/companies', component: CompaniesView },
    { path: '/sources', component: SourcesView },
    { path: '/research', component: ResearchView },
  ],
})

describe('CRM Routes', () => {
  beforeEach(() => {
    router.push('/')
    vi.resetAllMocks()
  })

  it('OpportunitiesView route exists', () => {
    const route = router.getRoutes().find(r => r.path === '/opportunities')
    expect(route).toBeDefined()
    expect(route?.component).toBe(OpportunitiesView)
  })

  it('LeadsView route exists', () => {
    const route = router.getRoutes().find(r => r.path === '/leads')
    expect(route).toBeDefined()
    expect(route?.component).toBe(LeadsView)
  })

  it('CompaniesView route exists', () => {
    const route = router.getRoutes().find(r => r.path === '/companies')
    expect(route).toBeDefined()
    expect(route?.component).toBe(CompaniesView)
  })

  it('SourcesView route exists', () => {
    const route = router.getRoutes().find(r => r.path === '/sources')
    expect(route).toBeDefined()
    expect(route?.component).toBe(SourcesView)
  })

  it('ResearchView route exists', () => {
    const route = router.getRoutes().find(r => r.path === '/research')
    expect(route).toBeDefined()
    expect(route?.component).toBe(ResearchView)
  })
})

describe('OpportunitiesView', () => {
  it('renders opportunities title', () => {
    const wrapper = mount(OpportunitiesView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Холодные возможности')
  })

  it('shows empty state when no opportunities', () => {
    const wrapper = mount(OpportunitiesView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Нет возможностей')
  })
})

describe('LeadsView', () => {
  it('renders leads title', () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Входящие заявки')
  })

  it('shows empty state when no leads', () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Нет входящих заявок')
  })

  it('has manual lead creation form', () => {
    const wrapper = mount(LeadsView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Создать входящую заявку вручную')
    expect(wrapper.find('form').exists()).toBe(true)
  })
})

describe('CompaniesView', () => {
  it('renders companies title', () => {
    const wrapper = mount(CompaniesView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Компании')
  })
})

describe('SourcesView', () => {
  it('renders sources title', () => {
    const wrapper = mount(SourcesView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Источники')
  })

  it('shows pending review approval state', async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce({
      sources: [{ source_id: 's1', name: 'Test', source_type: 'public', base_url: 'https://example.com', publisher: 'Pub', approval_status: 'pending_review' }],
      count: 1
    })
    
    const wrapper = mount(SourcesView, {
      global: { plugins: [router] },
    })
    
    await new Promise(resolve => setTimeout(resolve, 0))
    await wrapper.vm.$nextTick()
    
    expect(wrapper.text()).toContain('pending_review')
  })
})

describe('ResearchView', () => {
  it('renders research title', () => {
    const wrapper = mount(ResearchView, {
      global: { plugins: [router] },
    })
    expect(wrapper.text()).toContain('Исследования')
  })
})
