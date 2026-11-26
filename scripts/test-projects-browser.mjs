// Optional local acceptance using the pinned web test dependency and installed Edge.
// Usage: node scripts/test-projects-browser.mjs <VALENTINA.exe> <prepared-web-root>
// No dependencies are installed; all records and the isolated profile are synthetic.
import { createRequire } from 'node:module'
import { spawn } from 'node:child_process'
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { resolve, join } from 'node:path'
import { createServer } from 'node:net'
import assert from 'node:assert/strict'

const [executable, webRoot] = process.argv.slice(2)
if (!executable || !webRoot) throw new Error('Expected executable and prepared web root')
const require = createRequire(join(resolve(webRoot), 'package.json'))
const { chromium } = require('@playwright/test')
const profile = await mkdtemp(join(tmpdir(), 'valentina-project-browser-'))
const portProbe = createServer()
await new Promise(resolve => portProbe.listen(0, '127.0.0.1', resolve))
const port = portProbe.address().port
await new Promise(resolve => portProbe.close(resolve))
const base = `http://127.0.0.1:${port}`
await mkdir(join(profile, '.huginn'))
await mkdir(join(profile, 'roaming'))
await mkdir(join(profile, 'local'))
await writeFile(join(profile, '.huginn/config.json'), JSON.stringify({version:13,web_ui:{enabled:true,port,auto_open:false,bind:'127.0.0.1'},scheduler_enabled:false,tools_enabled:false,git_stage_on_write:false,mcp_servers:[]}))
const server = spawn(resolve(executable), ['-no-tray','serve','--foreground'], {
  cwd:profile, windowsHide:true, stdio:'ignore',
  env:{...process.env,USERPROFILE:profile,HOME:profile,HUGINN_HOME:profile,APPDATA:join(profile,'roaming'),LOCALAPPDATA:join(profile,'local'),NVIDIA_NIM_API_KEY:'',VALENTINA_INTAKE_HMAC_SECRET:''},
})
let browser
try {
  let ready=false
  for(let attempt=0;attempt<40;attempt++) {
    try { ready=(await fetch(base+'/api/v1/valentina/health')).ok } catch {}
    if(ready) break
    if(server.exitCode !== null) throw new Error('Test runtime exited')
    await new Promise(resolve=>setTimeout(resolve,250))
  }
  assert(ready,'runtime ready')
  const {token}=await (await fetch(base+'/api/v1/token')).json()
  const api=async(path,body)=>{
    const response=await fetch(base+'/api/v1/valentina'+path,{method:body?'POST':'GET',headers:{Authorization:`Bearer ${token}`,'Content-Type':'application/json'},body:body?JSON.stringify(body):undefined})
    assert(response.ok,`synthetic API request failed: ${path}, ${response.status}`)
    return response.json()
  }
  const lead=await api('/leads',{request_summary:'Synthetic browser renovation',inbound_channel:'manual',privacy_notice:true})
  for(const status of ['triage','qualified']) await api('/review-decisions',{entity_type:'inbound_lead',entity_id:lead.lead_id,new_status:status,decision_reason:'Synthetic human browser acceptance'})
  browser=await chromium.launch({channel:'msedge',headless:true})
  const page=await browser.newPage({viewport:{width:1440,height:1000}})
  const pageErrors=[]
  page.on('pageerror',error=>pageErrors.push(error.message))
  await page.route('**/*',route=>new URL(route.request().url()).origin===base?route.continue():route.abort())
  await page.goto(base+'/#/valentina/leads?lead='+lead.lead_id)
  await page.getByRole('link',{name:'Создать проект из заявки'}).click()
  await page.locator('[data-test="project-title"]').fill('Synthetic browser project')
  await page.getByRole('button',{name:'Создать проект',exact:true}).click()
  await page.getByRole('heading',{name:'Synthetic browser project',exact:true}).waitFor()
  const projectID=new URL(page.url()).hash.split('/').pop()
  await page.getByRole('button',{name:'В работе',exact:true}).click()
  await page.getByRole('heading',{name:'Состояние: В работе'}).waitFor()
  await page.locator('[data-test="line-name"]').fill('Synthetic labor')
  await page.getByLabel('Количество',{exact:true}).fill('1,500')
  await page.locator('[data-test="line-price"]').fill('1,01')
  await page.getByLabel('IVA в процентах').fill('21')
  await page.locator('[data-test="add-line"]').click()
  await page.getByLabel('Вид позиции').nth(1).selectOption('material')
  await page.locator('[data-test="line-name"]').nth(1).fill('Synthetic material')
  await page.locator('[data-test="line-price"]').nth(1).fill('1')
  await page.getByLabel('IVA в процентах').nth(1).fill('10')
  await page.locator('[data-test="preview-estimate"]').click()
  await page.getByText('Итого:',{exact:false}).waitFor()
  assert((await page.locator('.totals').textContent()).includes('2,94'))
  await page.getByRole('button',{name:'Сохранить ревизию 1',exact:true}).click()
  await page.getByText('Ревизия 1 · Черновик · 2,94 €',{exact:true}).click()
  await page.getByRole('button',{name:'Утвердить эту ревизию'}).click()
  await page.getByText('Ревизия 1 · Утверждена · 2,94 €',{exact:true}).waitFor()
  await page.getByRole('button',{name:'Взять за основу'}).click()
  await page.locator('[data-test="line-price"]').first().fill('2')
  await page.locator('[data-test="preview-estimate"]').click()
  await page.getByRole('button',{name:'Сохранить ревизию 2',exact:true}).click()
  await page.getByText('Ревизия 2 · Черновик · 4,73 €',{exact:true}).waitFor()
  await page.locator('[data-test="document-name"]').fill('Synthetic estimate record')
  await page.getByLabel('Ревизия сметы',{exact:true}).selectOption({label:'Ревизия 1'})
  await page.getByRole('button',{name:'Добавить запись',exact:true}).click()
  await page.getByText('Synthetic estimate record',{exact:true}).waitFor()
  await page.reload()
  await page.getByText('Synthetic estimate record',{exact:true}).waitFor()
  const history=await api('/projects/'+projectID+'/estimates')
  assert.equal(history.estimates.length,2)
  assert.equal(history.estimates[1].total_cents,294)
  assert.equal(history.estimates[1].status,'approved')
  const original=await api('/leads/'+lead.lead_id)
  assert.equal((original.lead||original).status,'qualified')
  await page.getByRole('link',{name:lead.lead_id,exact:true}).click()
  await page.getByRole('link',{name:'Открыть проект',exact:true}).waitFor()
  await page.getByRole('button',{name:'О программе',exact:true}).click()
  await page.getByRole('heading',{name:'Лицензии',exact:true}).waitFor()
  await page.locator('pre').filter({hasText:'Business Source License'}).waitFor()
  const backupButton=page.getByRole('button',{name:'Скачать копию CRM',exact:true})
  assert.equal(await backupButton.isDisabled(),true)
  await page.getByRole('checkbox').check()
  const downloadPromise=page.waitForEvent('download')
  await backupButton.click()
  const download=await downloadPromise
  assert.equal(download.suggestedFilename(),'VALENTINA-CRM.sqlite')
  await page.getByRole('status').filter({hasText:'Копия подготовлена'}).waitFor()
  assert.deepEqual(pageErrors,[],'no browser JavaScript errors')
  console.log('VALENTINA_PROJECT_BROWSER=PASS (lead link, create, status, two-line cents/VAT, approve, revise, metadata, reload, original lead, no page errors)')
} finally {
  await browser?.close()
  if(server.exitCode===null) { server.kill(); await new Promise(resolve=>server.once('exit',resolve)) }
  // Only the unique synthetic profile created by mkdtemp above is removed.
  await rm(profile,{recursive:true,force:true,maxRetries:5,retryDelay:200})
}
