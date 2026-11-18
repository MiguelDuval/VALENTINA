import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import OpportunityDetailView from '../OpportunityDetailView.vue'
import { apiFetch } from '../../composables/useApi'

vi.mock('../../composables/useApi', () => ({ apiFetch: vi.fn() }))

const opportunity = {
  opportunity_id: 'opp-1',
  title: 'Restauración de baño',
  summary: 'Solicitud pública de reparación',
  opportunity_type: 'direct_request',
  status: 'needs_review',
  demand_evidence_type: 'direct_request',
  source_evidence_ids: ['evidence-1'],
  geography: { municipality: 'Valencia', province: 'Valencia', country_code: 'ES' },
  value: null,
  fit_assessment: null,
}

const evidence = {
  evidence_id: 'evidence-1',
  page_title: 'Aviso público de reparación',
  publisher: 'Example publisher',
  canonical_url: 'https://example.test/notice',
  observed_at: '2026-11-18T10:00:00Z',
  published_at: null,
  retrieval_status: 'success',
  extractor_id: 'manual',
  extractor_version: '1',
  facts: [{
    field_path: 'title',
    value: 'Reparación de baño',
    evidence_excerpt: 'Se solicita reparar el baño.',
    evidence_locator: 'https://example.test/notice#description',
    extraction_method: 'manual',
    confidence: 0.9,
  }],
}

const decisions = [{
  decision_id: 'decision-1',
  entity_type: 'opportunity',
  entity_id: 'opp-1',
  previous_status: 'candidate',
  new_status: 'needs_review',
  decision_time: '2026-11-18T10:05:00Z',
  decision_reason: 'Requires source verification',
}]

describe('OpportunityDetailView', () => {
  let router

  const mountView = async () => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/valentina/opportunities', component: { template: '<div />' } },
        { path: '/valentina/opportunities/:id', component: OpportunityDetailView },
      ],
    })
    await router.push('/valentina/opportunities/opp-1')
    await router.isReady()
    const wrapper = mount(OpportunityDetailView, { global: { plugins: [router] } })
    await flushPromises()
    return wrapper
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiFetch).mockImplementation(async (path, options = {}) => {
      if (path === '/api/v1/valentina/opportunities/opp-1') return opportunity
      if (String(path).startsWith('/api/v1/valentina/source-evidence?')) return { evidence: [evidence], count: 1 }
      if (String(path).startsWith('/api/v1/valentina/review-decisions?')) return { decisions, count: decisions.length }
      if (path === '/api/v1/valentina/review-decisions' && options.method === 'POST') return { decision_id: 'decision-2' }
      throw new Error('Unexpected API request: ' + path)
    })
  })

  it('loads the opportunity, linked evidence and scoped review history', async () => {
    const wrapper = await mountView()
    expect(wrapper.text()).toContain('Restauración de baño')
    expect(wrapper.text()).toContain('Aviso público de reparación')
    expect(wrapper.text()).toContain('Requires source verification')
    expect(wrapper.text()).toContain('Reparación de baño')
    const sourceLink = wrapper.find('a[href="https://example.test/notice"]')
    expect(sourceLink.exists()).toBe(true)
    expect(sourceLink.attributes('target')).toBe('_blank')
    expect(sourceLink.attributes('rel')).toContain('noopener')
    expect(apiFetch).toHaveBeenCalledWith(expect.stringContaining('source-evidence?id=evidence-1'))
    expect(apiFetch).toHaveBeenCalledWith(expect.stringContaining('review-decisions?entity_type=opportunity&entity_id=opp-1'))
    wrapper.unmount()
  })

  it('requires an operator reason and posts the selected valid transition', async () => {
    const wrapper = await mountView()
    await wrapper.find('select').setValue('qualified')
    await wrapper.find('textarea').setValue('The source confirms a repair request.')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(apiFetch).toHaveBeenCalledWith('/api/v1/valentina/review-decisions', {
      method: 'POST',
      body: JSON.stringify({
        entity_type: 'opportunity',
        entity_id: 'opp-1',
        new_status: 'qualified',
        decision_reason: 'The source confirms a repair request.',
      }),
    })
    wrapper.unmount()
  })
})
