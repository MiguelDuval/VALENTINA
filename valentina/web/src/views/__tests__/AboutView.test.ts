import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AboutView from '../AboutView.vue'
import { NAV_ITEMS, sectionFromPath, showChatSidebar } from '../../utils/navLayout'
const { api, token } = vi.hoisted(() => ({ api: vi.fn(), token: vi.fn() }))
vi.mock('../../composables/useApi', () => ({ apiFetch: api, ensureToken: token }))
beforeEach(() => { vi.restoreAllMocks(); api.mockReset().mockResolvedValue({ license: '<complete license>', upstream_commit: 'pinned' }); token.mockReset().mockResolvedValue('synthetic-token') })
afterEach(() => { vi.unstubAllGlobals(); vi.useRealTimers() })
describe('About and CRM backup', () => {
 it('is reachable and renders license as text with honest limitations', async () => {
  expect(NAV_ITEMS.find(i => i.path === '/about')?.label).toBe('О программе')
  expect(sectionFromPath('/about')).toBe('about'); expect(showChatSidebar('/about')).toBe(false)
  const w=mount(AboutView); await flushPromises()
  expect(w.get('pre').text()).toBe('<complete license>')
  expect(w.text()).toContain('не завершена'); expect(w.text()).toContain('не полная копия')
  expect(w.get('button').attributes('disabled')).toBeDefined()
 })
 it('shows sanitized license error and retries', async () => {
  api.mockRejectedValueOnce(new Error('private internal detail'))
  const w=mount(AboutView);await flushPromises();expect(w.text()).toContain('Не удалось загрузить')
  expect(w.text()).not.toContain('private internal detail')
  await w.get('[role=alert] button').trigger('click');await flushPromises();expect(w.find('pre').exists()).toBe(true)
 })
 it('requires acknowledgement and reports download failure without leaking details', async () => {
  const fetchMock=vi.spyOn(globalThis,'fetch').mockResolvedValue(new Response('private failure',{status:503}))
  const w=mount(AboutView);await flushPromises();expect(fetchMock).not.toHaveBeenCalled()
  await w.get('input').setValue(true);await w.get('button').trigger('click');await flushPromises()
  expect(fetchMock).toHaveBeenCalledWith('/api/v1/valentina/backup',expect.objectContaining({cache:'no-store',redirect:'error',headers:{Authorization:'Bearer synthetic-token'}}))
  expect(w.text()).toContain('Не удалось создать');expect(w.text()).not.toContain('private failure')
 })
 it('downloads only bytes matching the advertised digest and cleans the object URL', async () => {
  vi.useFakeTimers()
  vi.stubGlobal('crypto', { subtle: { digest: vi.fn().mockResolvedValue(new Uint8Array(32).buffer) } })
  const create=vi.fn().mockReturnValue('blob:synthetic'), revoke=vi.fn()
  vi.stubGlobal('URL', { createObjectURL:create, revokeObjectURL:revoke })
  const click=vi.spyOn(HTMLAnchorElement.prototype,'click').mockImplementation(() => {})
  vi.spyOn(globalThis,'fetch').mockResolvedValue(new Response('SQLite fixture',{headers:{'Content-Type':'application/vnd.sqlite3','X-Backup-SHA256':'0'.repeat(64)}}))
  const w=mount(AboutView);await flushPromises();await w.get('input').setValue(true);await w.get('button').trigger('click');await flushPromises()
  expect(click).toHaveBeenCalledOnce();expect(create).toHaveBeenCalledOnce();expect(w.text()).toContain('Копия подготовлена')
  await vi.advanceTimersByTimeAsync(1000);expect(revoke).toHaveBeenCalledWith('blob:synthetic')
 })
 it('rejects a mismatched digest before creating a browser download', async () => {
  vi.stubGlobal('crypto', { subtle: { digest: vi.fn().mockResolvedValue(new Uint8Array(32).buffer) } })
  const click=vi.spyOn(HTMLAnchorElement.prototype,'click').mockImplementation(() => {})
  vi.spyOn(globalThis,'fetch').mockResolvedValue(new Response('SQLite fixture',{headers:{'Content-Type':'application/vnd.sqlite3','X-Backup-SHA256':'1'.repeat(64)}}))
  const w=mount(AboutView);await flushPromises();await w.get('input').setValue(true);await w.get('button').trigger('click');await flushPromises()
  expect(click).not.toHaveBeenCalled();expect(w.text()).toContain('Не удалось создать')
 })
})
