import { describe, it, expect, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ResearchView from '../ResearchView.vue'
import SourcesView from '../SourcesView.vue'
import { apiFetch } from '../../composables/useApi'
vi.mock('../../composables/useApi', () => ({ apiFetch: vi.fn() }))
const api = vi.mocked(apiFetch)
const sources = [
  { source_id: 'approved-rss', name: 'Fixture RSS', approval_status: 'approved', manual_run_allowed: true, manual_run_reason: '' },
  { source_id: 'pending', name: 'Pending feed', approval_status: 'approved', manual_run_allowed: false, manual_run_reason: 'Условия источника не проверены' },
]
describe('Manual research contract', () => {
  it('runs only server-eligible feeds, blocks double submit and displays persisted counts', async () => {
    let finish: (value: unknown) => void = () => {}
    let completed = false
    api.mockImplementation(async (path, options) => {
      if (path === '/api/v1/valentina/sources') return { sources }
      if (options?.method === 'POST') {
        expect(JSON.parse(options.body as string).source_ids).toEqual(['approved-rss'])
        return new Promise(resolve => { finish = value => { completed = true; resolve(value) } })
      }
      return { runs: completed ? [{ run_id: 'fixture-run', status: 'completed', candidate_count: 1, verified_count: 0, duplicate_count: 2 }] : [] }
    })
    const wrapper = mount(ResearchView)
    await flushPromises()
    const start = wrapper.get('button')
    expect(start.attributes('disabled')).toBeUndefined()
    await start.trigger('click')
    expect(start.attributes('disabled')).toBeDefined()
    finish({ status: 'completed' })
    await flushPromises()
    expect(wrapper.text()).toContain('1 кандидатов')
    expect(wrapper.text()).toContain('2 дубликатов')
    expect(wrapper.text()).toContain('ручной проверки')
    wrapper.unmount()
  })
  it('fails closed on missing eligibility and load errors', async () => {
    api.mockResolvedValue({ sources: [{ source_id: 'legacy', approval_status: 'approved' }], runs: [] })
    const wrapper = mount(ResearchView)
    await flushPromises()
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    wrapper.unmount()
    api.mockRejectedValue(new Error('synthetic failure'))
    const failed = mount(ResearchView)
    await flushPromises()
    expect(failed.get('button').attributes('disabled')).toBeDefined()
    expect(failed.text()).toContain('Не удалось')
    failed.unmount()
  })
  it('shows the server reason for a blocked source', async () => {
    api.mockResolvedValue({ sources })
    const wrapper = mount(SourcesView)
    await flushPromises()
    expect(wrapper.text()).toContain('Условия источника не проверены')
    wrapper.unmount()
  })
  it('shows a rejected run and permits a later manual retry', async () => {
    api.mockImplementation(async (path, options) => {
      if (options?.method === 'POST') throw new Error('synthetic rejection')
      if (path === '/api/v1/valentina/sources') return { sources }
      return { runs: [] }
    })
    const wrapper = mount(ResearchView)
    await flushPromises()
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Не удалось запустить исследование')
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })
})
