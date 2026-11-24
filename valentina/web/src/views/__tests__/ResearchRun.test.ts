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


describe('Research admission and recovery feedback', () => {
  it.each([
    ['research source busy', 'Источник уже используется'],
    ['research quota unavailable', 'Достигнут дневной лимит'],
  ])('shows safe admission error %s and reloads stored history', async (message, text) => {
    api.mockImplementation(async (path, options) => {
      if (options?.method === 'POST') throw new Error(message)
      if (path === '/api/v1/valentina/sources') return { sources }
      return { runs: [{ run_id: 'recovered-run', status: 'failed', completed_at: '2026-11-24T10:00:00Z', candidate_count: 3, verified_count: 0, duplicate_count: 1, notes: 'Interrupted by application shutdown; manual retry required' }] }
    })
    const wrapper = mount(ResearchView)
    await flushPromises()
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain(text)
    expect(wrapper.text()).toContain('Ошибка / прервано')
    expect(wrapper.text()).toContain('3 кандидатов')
    expect(wrapper.text()).toContain('Завершено:')
    expect(wrapper.text()).toContain('Interrupted by application shutdown')
    expect(wrapper.get('button').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })
  it('does not display arbitrary server error payloads', async () => {
    api.mockImplementation(async (path, options) => {
      if (options?.method === 'POST') throw new Error('private-source-diagnostics')
      if (path === '/api/v1/valentina/sources') return { sources }
      return { runs: [] }
    })
    const wrapper = mount(ResearchView)
    await flushPromises()
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).not.toContain('private-source-diagnostics')
    expect(wrapper.text()).toContain('Не удалось запустить исследование')
    wrapper.unmount()
  })
  it('acknowledges cancellation without claiming execution has stopped', async () => {
    api.mockImplementation(async (path, options) => {
      if (options?.method === 'POST') return { status: 'cancellation_requested' }
      if (path === '/api/v1/valentina/sources') return { sources }
      return { runs: [{ run_id: 'active-run', status: 'running', candidate_count: 0 }] }
    })
    const wrapper = mount(ResearchView)
    await flushPromises()
    await wrapper.get('.valentina-btn-danger').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Отмена запрошена')
    expect(wrapper.text()).toContain('Выполняется')
    wrapper.unmount()
  })
  it('preserves approval reasons and shows nested rate limits', async () => {
    api.mockResolvedValue({ sources: [{ ...sources[0], approval_reason: 'Synthetic reviewed permission', rate_limits: { minimum_interval_seconds: 60, daily_request_limit: 2, timeout_seconds: 10 } }] })
    const wrapper = mount(SourcesView)
    await flushPromises()
    expect(wrapper.text()).toContain('Synthetic reviewed permission')
    expect(wrapper.text()).toContain('Минимальный интервал: 60с')
    expect(wrapper.text()).toContain('Лимит попыток в день (UTC): 2')
    wrapper.unmount()
  })
})
