import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ProjectsView from '../ProjectsView.vue'
import { decimalToInteger, integerToDecimal } from '../projectForms'
import { sectionFromPath, NAV_ITEMS } from '../../utils/navLayout'

const { api } = vi.hoisted(() => ({ api: vi.fn() }))
vi.mock('../../composables/useApi', () => ({ apiFetch: api }))
const project = { project_id: 'p1', title: 'Synthetic project', status: 'planned', version: 1, source_lead_id: '', company_id: '', description: '', service_scope: '', location_text: '', start_date: '', end_date: '' }
beforeEach(() => {
 api.mockReset().mockImplementation(async (path: string) => {
  if (path.includes('/companies')) return { companies: [] }
  if (path.endsWith('/estimates')) return { estimates: [] }
  if (path.endsWith('/documents')) return { documents: [] }
  if (path.endsWith('/projects/p1')) return project
  return { projects: [] }
 })
})
async function render(path='/valentina/projects') {
 const router=createRouter({history:createMemoryHistory(),routes:[{path:'/valentina/projects/:id?',component:ProjectsView}]})
 await router.push(path); await router.isReady()
 const wrapper=mount(ProjectsView,{global:{plugins:[router]}}); await flushPromises(); return wrapper
}
describe('Projects operations',()=>{
 it('parses decimal strings exactly and rejects ambiguous/overflow money',()=>{
  expect(decimalToInteger('1,005',3)).toBe(1005)
  expect(decimalToInteger('0.29',2)).toBe(29)
  expect(integerToDecimal(1005,3)).toBe('1.005')
  for(const value of ['-1','1e3','1.001','9007199254740992','']) expect(()=>decimalToInteger(value,2)).toThrow()
 })
 it('registers project navigation',()=>{
  expect(sectionFromPath('/valentina/projects/p1')).toBe('crm-projects')
  expect(NAV_ITEMS.some(item=>item.path==='/valentina/projects')).toBe(true)
 })
 it('creates a manual project from the form',async()=>{
  const wrapper=await render()
  await wrapper.get('[data-test="project-title"]').setValue('Synthetic new')
  api.mockImplementationOnce(async()=>({...project,title:'Synthetic new'}))
  await wrapper.get('[data-test="project-form"]').trigger('submit'); await flushPromises()
  const call=api.mock.calls.find(([path,opts])=>path==='/api/v1/valentina/projects' && opts?.method==='POST')
  expect(JSON.parse(call![1].body).title).toBe('Synthetic new')
 })
 it('previews two lines with exact cents and saves only after a human click',async()=>{
  const wrapper=await render('/valentina/projects/p1')
  await wrapper.get('[data-test="add-line"]').trigger('click')
  const names=wrapper.findAll('[data-test="line-name"]')
  await names[0]!.setValue('Labor'); await names[1]!.setValue('Material')
  for(const input of wrapper.findAll('[data-test="line-price"]')) await input.setValue('0.29')
  for(const input of wrapper.findAll('[aria-label="IVA в процентах"]')) await input.setValue('0')
  api.mockImplementationOnce(async()=>({net_cents:58,vat_cents:0,total_cents:58}))
  await wrapper.get('[data-test="preview-estimate"]').trigger('click'); await flushPromises()
  const call=api.mock.calls.find(([path])=>path.endsWith('/preview'))
  expect(JSON.parse(call![1].body).lines).toHaveLength(2)
  expect(JSON.parse(call![1].body).lines[0].unit_cents).toBe(29)
  expect(api.mock.calls.some(([path,opts])=>path.endsWith('/estimates')&&opts?.method==='POST')).toBe(false)
  expect(wrapper.text()).toContain('0,58')
 })
 it('reports loading failure and allows retry',async()=>{
  api.mockRejectedValueOnce(new Error('offline'))
  const wrapper=await render()
  expect(wrapper.get('[role="alert"]').text()).toContain('Не удалось')
  await wrapper.get('[data-test="reload"]').trigger('click'); await flushPromises()
  expect(wrapper.find('[role="alert"]').exists()).toBe(false)
 })
 it('labels metadata-only documents and submits no file path',async()=>{
  const wrapper=await render('/valentina/projects/p1')
  expect(wrapper.text()).toContain('Файлы не загружаются')
  await wrapper.get('[data-test="document-name"]').setValue('Synthetic note')
  await wrapper.get('[data-test="document-form"]').trigger('submit'); await flushPromises()
  const call=api.mock.calls.find(([path,opts])=>path.endsWith('/documents')&&opts?.method==='POST')
  expect(JSON.parse(call![1].body).name).toBe('Synthetic note')
 })
 it('invalidates a preview when a line changes',async()=>{
  const wrapper=await render('/valentina/projects/p1')
  await wrapper.get('[data-test="line-name"]').setValue('Synthetic')
  await wrapper.get('[data-test="line-price"]').setValue('1')
  await wrapper.get('[aria-label="IVA в процентах"]').setValue('0')
  api.mockImplementationOnce(async()=>({net_cents:100,vat_cents:0,total_cents:100}))
  await wrapper.get('[data-test="preview-estimate"]').trigger('click'); await flushPromises()
  expect(wrapper.find('.totals').exists()).toBe(true)
  await wrapper.get('[data-test="line-price"]').setValue('2')
  expect(wrapper.find('.totals').exists()).toBe(false)
 })
 it('revises a saved snapshot without mutating history and explicitly approves',async()=>{
  const estimate={estimate_id:'e1',revision:1,status:'draft',created_at:'2026-11-24T00:00:00Z',net_cents:100,vat_cents:21,total_cents:121,lines:[{kind:'labor',name:'Original',quantity_milli:1000,unit_cents:100,vat_bps:2100}]}
  const fallback=api.getMockImplementation()!
  api.mockImplementation(async(path:string,...args:any[])=>path.endsWith('/estimates')?{estimates:[estimate]}:fallback(path,...args))
  const wrapper=await render('/valentina/projects/p1')
  await wrapper.findAll('button').find(b=>b.text()==='Взять за основу')!.trigger('click')
  await wrapper.get('[data-test="line-name"]').setValue('Revised')
  expect(estimate.lines[0]!.name).toBe('Original')
  await wrapper.findAll('button').find(b=>b.text()==='Утвердить эту ревизию')!.trigger('click'); await flushPromises()
  expect(api.mock.calls.some(([path,opts])=>path.endsWith('/e1/approve')&&opts.method==='POST')).toBe(true)
 })
 it('resolves an existing lead project instead of posting a duplicate',async()=>{
  const fallback=api.getMockImplementation()!
  api.mockImplementation(async(path:string,...args:any[])=>path.endsWith('/leads/l1/project')?{project}:fallback(path,...args))
  const wrapper=await render('/valentina/projects?lead=l1')
  expect(wrapper.text()).toContain('Synthetic project')
  expect(api.mock.calls.some(([,opts])=>opts?.method==='POST')).toBe(false)
 })
 it('shows conflicts without overwriting the form',async()=>{
  const wrapper=await render('/valentina/projects/p1')
  await wrapper.get('[data-test="project-title"]').setValue('Unsaved')
  api.mockRejectedValueOnce(new Error('project or revision changed; reload before retrying'))
  await wrapper.get('[data-test="project-form"]').trigger('submit'); await flushPromises()
  expect(wrapper.get('[role="alert"]').text()).toContain('Обновите')
  expect((wrapper.get('[data-test="project-title"]').element as HTMLInputElement).value).toBe('Unsaved')
 })
})
