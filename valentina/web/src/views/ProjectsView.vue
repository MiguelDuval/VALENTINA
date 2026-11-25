<template>
  <main class="projects-page">
    <header class="page-heading">
      <div><p class="eyebrow">VALENTINA · Рабочие проекты</p><h1>{{ project ? project.title : 'Проекты' }}</h1>
        <p>Работы, сметы и документы. Все изменения подтверждает оператор.</p></div>
      <RouterLink v-if="project" to="/valentina/projects">← Все проекты</RouterLink>
      <button data-test="reload" :disabled="busy || loading" @click="load">Обновить</button>
    </header>
    <p v-if="error" role="alert" class="error">{{ error }}</p>
    <p v-if="notice" role="status">{{ notice }}</p>
    <p v-if="loading" role="status">Загрузка…</p>
    <template v-else-if="ready">
      <section v-if="!project" class="panel">
        <h2>Список проектов</h2><p v-if="projects.length === 0">Проектов пока нет. Создайте проект ниже или откройте квалифицированную заявку.</p>
        <ul class="project-list"><li v-for="item in projects" :key="item.project_id">
          <RouterLink :to="'/valentina/projects/' + item.project_id">{{ item.title }}</RouterLink>
          <span>{{ projectStatuses[item.status] }} · {{ item.location_text || 'Место не указано' }}</span>
        </li></ul>
        <div class="actions"><button :disabled="offset === 0 || busy" @click="offset -= 100; load()">Назад</button><span>Страница {{ offset / 100 + 1 }}</span><button :disabled="projects.length < 100 || busy" @click="offset += 100; load()">Далее</button></div>
      </section>
      <section v-if="project" class="panel">
        <h2>Состояние: {{ projectStatuses[project.status] }}</h2>
        <p v-if="project.source_lead_id">Исходная заявка: <RouterLink :to="{ path: '/valentina/leads', query: { lead: project.source_lead_id } }">{{ project.source_lead_id }}</RouterLink>. Оригинал и история сохранены.</p>
        <p v-if="!project.service_scope">Объём работ пока не указан.</p>
        <div class="actions"><button v-for="status in projectTransitions[project.status] || []" :key="status" :disabled="busy" @click="setStatus(status)">{{ projectStatuses[status] }}</button></div>
        <p v-if="!(projectTransitions[project.status]?.length)">Конечный статус. История смет и документы доступны.</p>
      </section>
      <section class="panel">
        <h2>{{ project ? 'Карточка проекта' : 'Новый проект' }}</h2>
        <p v-if="form.source_lead_id && !project">Создание из квалифицированной заявки {{ form.source_lead_id }}. Проверьте сведения перед сохранением.</p>
        <form data-test="project-form" @submit.prevent="saveProject">
          <fieldset :disabled="busy"><div class="form-grid">
            <label>Название<input v-model="form.title" data-test="project-title" required maxlength="240" /></label>
            <label>Компания (необязательно)<select v-model="form.company_id"><option value="">Не указана</option><option v-for="company in companies" :key="company.company_id" :value="company.company_id">{{ company.name }}</option></select></label>
            <label class="wide">Описание<textarea v-model="form.description" maxlength="10000" /></label>
            <label class="wide">Объём и виды работ<textarea v-model="form.service_scope" maxlength="4000" /></label>
            <label>Место работ<input v-model="form.location_text" maxlength="500" placeholder="Не указано" /></label>
            <label>Начало<input v-model="form.start_date" type="date" /></label>
            <label>Окончание<input v-model="form.end_date" type="date" :min="form.start_date" /></label>
          </div><button type="submit">{{ project ? 'Сохранить карточку' : 'Создать проект' }}</button></fieldset>
        </form>
      </section>
      <template v-if="project">
        <section class="panel">
          <h2>Сметы · EUR</h2><p>Каждое сохранение создаёт неизменяемую ревизию в статусе «Черновик». Утверждение не отправляет смету клиенту. Это не счёт-фактура и не налоговая консультация.</p>
          <p>Количество — до 3 знаков; цена — до 2. IVA вводится оператором, без автоматической ставки.</p>
          <form @submit.prevent="saveEstimate"><fieldset :disabled="busy">
            <div class="table-wrap"><table><thead><tr><th>Вид</th><th>Наименование</th><th>Количество</th><th>Цена, €</th><th>IVA, %</th><th></th></tr></thead>
              <tbody><tr v-for="(line, index) in lines" :key="line.key">
                <td><select v-model="line.kind" aria-label="Вид позиции"><option value="labor">Работа</option><option value="material">Материал</option></select></td>
                <td><input v-model="line.name" data-test="line-name" aria-label="Наименование позиции" maxlength="240" required /></td>
                <td><input v-model="line.quantity" aria-label="Количество" inputmode="decimal" required /></td>
                <td><input v-model="line.price" data-test="line-price" aria-label="Цена в евро" inputmode="decimal" required /></td>
                <td><input v-model="line.vat" aria-label="IVA в процентах" inputmode="decimal" required /></td>
                <td><button type="button" :disabled="lines.length === 1" @click="lines.splice(index, 1)">Удалить</button></td>
              </tr></tbody></table></div>
            <div class="actions"><button data-test="add-line" type="button" :disabled="lines.length >= 200" @click="lines.push(blankLine())">Добавить позицию</button><button data-test="preview-estimate" type="button" @click="previewEstimate">Рассчитать</button><button type="submit" :disabled="!preview">Сохранить ревизию {{ latestRevision + 1 }}</button></div>
            <p v-if="revisionSource">Основа: ревизия {{ revisionSource }}. Будет создана новая ревизия.</p>
            <p v-if="preview" class="totals">Без IVA: {{ euros(preview.net_cents) }} · IVA: {{ euros(preview.vat_cents) }} · Итого: <strong>{{ euros(preview.total_cents) }}</strong></p>
          </fieldset></form>
          <h3>История ревизий</h3><p v-if="estimates.length === 0">Сохранённых смет пока нет.</p>
          <details v-for="estimate in estimates" :key="estimate.estimate_id" class="revision">
            <summary>Ревизия {{ estimate.revision }} · {{ estimate.status === 'approved' ? 'Утверждена' : 'Черновик' }} · {{ euros(estimate.total_cents) }}</summary>
            <p>Создана: {{ date(estimate.created_at) }}<span v-if="estimate.approved_at"> · Утверждена: {{ date(estimate.approved_at) }}</span></p>
            <ul><li v-for="(line, index) in estimate.lines" :key="index">{{ line.kind === 'labor' ? 'Работа' : 'Материал' }}: {{ line.name }} · {{ integerToDecimal(line.quantity_milli, 3) }} × {{ euros(line.unit_cents) }} · IVA {{ integerToDecimal(line.vat_bps, 2) }}%</li></ul>
            <p>Без IVA {{ euros(estimate.net_cents) }} · IVA {{ euros(estimate.vat_cents) }}</p>
            <div class="actions"><button :disabled="busy" @click="revise(estimate)">Взять за основу</button><button v-if="estimate.status === 'draft'" :disabled="busy" @click="approve(estimate)">Утвердить эту ревизию</button></div>
          </details>
        </section>
        <section class="panel">
          <h2>Реестр документов</h2><p>Только метаданные. Файлы не загружаются и не скачиваются; пути к файлам не сохраняются.</p>
          <form data-test="document-form" @submit.prevent="saveDocument"><fieldset :disabled="busy"><div class="form-grid">
            <label>Тип<select v-model="document.type"><option v-for="(label, key) in documentTypes" :key="key" :value="key">{{ label }}</option></select></label>
            <label>Название<input v-model="document.name" data-test="document-name" required maxlength="240" /></label>
            <label>Ревизия сметы<select v-model="document.estimate_id" aria-label="Ревизия сметы"><option value="">Без связи со сметой</option><option v-for="estimate in estimates" :key="estimate.estimate_id" :value="estimate.estimate_id">Ревизия {{ estimate.revision }}</option></select></label>
            <label class="wide">Аннотация<textarea v-model="document.annotation" maxlength="4000" /></label>
          </div><button type="submit">Добавить запись</button></fieldset></form>
          <p v-if="documents.length === 0">Записей пока нет.</p>
          <ul class="document-list"><li v-for="item in documents" :key="item.document_id"><strong>{{ item.name }}</strong> · {{ documentTypes[item.type] }} · {{ date(item.created_at) }}<span v-if="item.estimate_id"> · Смета, ревизия {{ estimates.find(e => e.estimate_id === item.estimate_id)?.revision }}</span><p>{{ item.annotation || 'Без аннотации' }}</p></li></ul>
        </section>
      </template>
    </template>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { apiFetch } from '../composables/useApi'
import { decimalToInteger, integerToDecimal, euros, projectStatuses, projectTransitions } from './projectForms'
import type { Project, ProjectInput, Estimate, Totals, DocumentRecord } from './projectForms'

const route = useRoute(), router = useRouter(), base = '/api/v1/valentina'
const loading = ref(false), busy = ref(false), ready = ref(false), error = ref(''), notice = ref(''), offset = ref(0)
const project = ref<Project | null>(null), projects = ref<Project[]>([]), companies = ref<{company_id:string; name:string}[]>([])
const estimates = ref<Estimate[]>([]), documents = ref<DocumentRecord[]>([])
const emptyForm = (): ProjectInput => ({ title:'', description:'', service_scope:'', company_id:'', source_lead_id:'', location_text:'', start_date:'', end_date:'' })
const form = ref(emptyForm())
let lineKey = 0, loadSequence = 0
const blankLine = () => ({ key: ++lineKey, kind:'labor', name:'', quantity:'1', price:'', vat:'' })
const lines = ref([blankLine()]), preview = ref<Totals | null>(null), revisionSource = ref<number | null>(null)
const latestRevision = computed(() => estimates.value[0]?.revision || 0)
const document = ref({type:'note', name:'', annotation:'', estimate_id:''})
const documentTypes: Record<string,string> = {estimate:'Смета',contract:'Договор',plan:'План',note:'Заметка',other:'Другое'}
const date = (value: string) => new Date(value).toLocaleString('ru-RU')
const send = <T,>(path:string, body:unknown, method='POST') => apiFetch<T>(base+path,{method,headers:{'Content-Type':'application/json'},body:JSON.stringify(body)})
function failure(e:unknown) {
  const message = e instanceof Error ? e.message : ''
  if (message.includes('project or revision changed')) return 'Данные уже изменены. Обновите карточку перед повтором.'
  if (message.includes('invalid estimate line')) return 'Проверьте позиции: название обязательно, количество больше нуля, цена неотрицательна, IVA от 0 до 100%.'
  if (message.includes('Введите') || message.includes('Слишком большое')) return 'Проверьте числа: количество до 3 знаков после запятой; цена и IVA до 2; без минуса и экспоненты.'
  if (message.includes('end date precedes')) return 'Окончание не может предшествовать началу.'
  if (message.includes('document metadata')) return 'Укажите название документа без пути, слешей и двоеточия.'
  if (message.includes('qualified')) return 'Для нового проекта нужна действующая заявка, квалифицированная человеком.'
  return 'Не удалось выполнить действие. Проверьте поля, связь и актуальность данных; при конфликте обновите карточку.'
}
async function load() {
  const sequence = ++loadSequence; loading.value=true; ready.value=false; error.value=''
  try {
    const id=typeof route.params.id === 'string' ? route.params.id : ''
    const companyData=await apiFetch<{companies:{company_id:string;name:string}[]}>(base+'/companies?limit=100')
    if (sequence !== loadSequence) return
    companies.value=companyData.companies || []
    if(id) {
      const [p,e,d]=await Promise.all([apiFetch<Project>(base+'/projects/'+encodeURIComponent(id)),apiFetch<{estimates:Estimate[]}>(base+'/projects/'+encodeURIComponent(id)+'/estimates'),apiFetch<{documents:DocumentRecord[]}>(base+'/projects/'+encodeURIComponent(id)+'/documents')])
      if (sequence !== loadSequence) return
      project.value=p; form.value=Object.fromEntries(Object.keys(emptyForm()).map(key=>[key,p[key as keyof ProjectInput] || ''])) as unknown as ProjectInput
      estimates.value=e.estimates; documents.value=d.documents
    } else {
      const list=await apiFetch<{projects:Project[]}>(base+'/projects?offset='+offset.value)
      if (sequence !== loadSequence) return
      projects.value=list.projects || []; project.value=null; form.value=emptyForm()
      if(typeof route.query.lead === 'string') {
        const linked=await apiFetch<{project:Project|null}>(base+'/leads/'+encodeURIComponent(route.query.lead)+'/project')
        if (sequence !== loadSequence) return
        if(linked.project) { await router.replace('/valentina/projects/'+linked.project.project_id); return }
        const data=await apiFetch<{lead?:any}&Record<string,any>>(base+'/leads/'+encodeURIComponent(route.query.lead)); const lead=data.lead || data
        if (sequence !== loadSequence) return
        if(lead.status!=='qualified') throw new Error('lead not qualified')
        form.value={...emptyForm(),title:lead.request_summary || '',source_lead_id:route.query.lead,company_id:lead.company_id || '',service_scope:(lead.requested_services || []).join(', '),location_text:lead.project_location?.location_text || ''}
      }
    }
    lines.value=[blankLine()]; preview.value=null; revisionSource.value=null; document.value={type:'note',name:'',annotation:'',estimate_id:''}; ready.value=true
  } catch(e) { if(sequence===loadSequence) error.value=failure(e) }
  finally { if(sequence===loadSequence) loading.value=false }
}
async function act(action:()=>Promise<void>) {
  if(busy.value) return
  busy.value=true; error.value=''; notice.value=''
  try { await action() } catch(e) { error.value=failure(e) } finally { busy.value=false }
}
async function saveProject() { await act(async()=>{
  const p=project.value
  const saved=await send<Project>(p?'/projects/'+p.project_id:'/projects',p?{...form.value,version:p.version}:form.value,p?'PUT':'POST')
  if(p) { project.value=saved; notice.value='Карточка сохранена.' } else await router.push('/valentina/projects/'+saved.project_id)
}) }
async function setStatus(status:string) { await act(async()=>{ const p=project.value!; project.value=await send<Project>('/projects/'+p.project_id+'/status',{status,version:p.version}); notice.value='Статус сохранён.' }) }
function estimateInput() {
  return {base_revision:latestRevision.value,lines:lines.value.map(line=>({kind:line.kind,name:line.name,quantity_milli:decimalToInteger(line.quantity,3),unit_cents:decimalToInteger(line.price,2),vat_bps:decimalToInteger(line.vat,2)}))}
}
async function previewEstimate() { await act(async()=>{ preview.value=await send<Totals>('/projects/'+project.value!.project_id+'/estimates/preview',estimateInput()) }) }
async function refreshEstimates() { estimates.value=(await apiFetch<{estimates:Estimate[]}>(base+'/projects/'+project.value!.project_id+'/estimates')).estimates }
async function saveEstimate() { if(!preview.value) return; await act(async()=>{
  await send('/projects/'+project.value!.project_id+'/estimates',estimateInput()); preview.value=null; await refreshEstimates(); notice.value='Новая ревизия сохранена как черновик.'
}) }
function revise(estimate:Estimate) { lines.value=estimate.lines.map(line=>({key:++lineKey,kind:line.kind,name:line.name,quantity:integerToDecimal(line.quantity_milli,3),price:integerToDecimal(line.unit_cents,2),vat:integerToDecimal(line.vat_bps,2)})); revisionSource.value=estimate.revision; preview.value=null }
async function approve(estimate:Estimate) { await act(async()=>{ await send('/projects/'+project.value!.project_id+'/estimates/'+estimate.estimate_id+'/approve',{}); await refreshEstimates(); notice.value='Ревизия утверждена. Ничего не отправлено.' }) }
async function saveDocument() { await act(async()=>{
  await send('/projects/'+project.value!.project_id+'/documents',document.value)
  document.value={type:'note',name:'',annotation:'',estimate_id:''}
  documents.value=(await apiFetch<{documents:DocumentRecord[]}>(base+'/projects/'+project.value!.project_id+'/documents')).documents; notice.value='Запись добавлена в реестр.'
}) }
watch(lines,()=>{preview.value=null},{deep:true,flush:'sync'})
watch(()=>[route.params.id,route.query.lead],()=>{notice.value=''; void load()},{immediate:true})
</script>

<style scoped>
.projects-page { @apply text-valentina-text; height: 100%; overflow: auto; padding: 28px; max-width: 1440px; margin: auto; }
.page-heading, .actions { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.page-heading { justify-content: space-between; margin-bottom: 24px; }
.eyebrow { @apply text-valentina-accent; font-size: 12px; letter-spacing: .12em; }
.panel { @apply border border-valentina-border bg-valentina-surface; border-radius: 12px; padding: 24px; margin: 20px 0; }
h1 { font-size: 28px; font-weight: 650; }
h2 { font-size: 21px; font-weight: 600; margin-bottom: 12px; }
h3 { font-size: 18px; margin: 24px 0 12px; }
p { margin: 10px 0; line-height: 1.6; }
a { @apply text-valentina-accent; text-decoration: underline; }
.form-grid { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 16px; margin-bottom: 20px; }
.wide { grid-column: 1/-1; }
label { display: flex; flex-direction: column; gap: 6px; }
input, textarea, select { @apply border border-valentina-border bg-valentina-bg; width: 100%; border-radius: 6px; padding: 9px; color: inherit; }
textarea { min-height: 74px; }
button { @apply border border-valentina-border bg-valentina-panel; border-radius: 6px; padding: 8px 14px; cursor: pointer; }
button:hover:not(:disabled) { @apply border-valentina-accent; }
button:disabled { opacity: .45; cursor: default; }
fieldset { min-width: 0; }
.error { @apply border-valentina-red; border-left-width: 4px; padding: 12px; }
.table-wrap { overflow: auto; }
table { width: 100%; border-collapse: collapse; margin: 16px 0; }
th, td { text-align: left; padding: 7px; min-width: 100px; }
th:nth-child(2) { min-width: 220px; }
.project-list li { @apply border-b border-valentina-border; display: flex; justify-content: space-between; gap: 16px; padding: 14px 0; }
.revision, .document-list li { @apply border-b border-valentina-border; padding: 16px 0; }
summary { cursor: pointer; }
.totals { font-size: 18px; }
@media(max-width:720px) { .projects-page { padding: 14px; } .panel { padding: 16px; } .form-grid { grid-template-columns: 1fr; } .project-list li { flex-direction: column; } }
</style>
