<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Входящие заявки</h1>
      <p class="valentina-page-subtitle">Клиентские запросы, поступившие через каналы связи бизнеса</p>
    </div>
    <div v-if="selectedLead" class="valentina-lead-detail">
      <div class="valentina-detail-header">
        <button type="button" class="valentina-btn-back" @click="selectedLead = null">← Назад к списку</button>
        <h2 class="valentina-detail-title">Заявка {{ selectedLead.lead_id.slice(-8) }}</h2>
      </div>
      <div v-if="detailError" class="valentina-error">{{ detailError }}</div>
      <div v-else-if="detailLoading" class="valentina-loading">Загрузка…</div>
      <div v-else class="valentina-lead-detail-content">
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Описание</label>
          <div class="valentina-detail-value">{{ selectedLead.request_summary }}</div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Канал</label>
          <div class="valentina-detail-value">{{ selectedLead.inbound_channel }}</div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Получено</label>
          <div class="valentina-detail-value">{{ formatDate(selectedLead.received_at) }}</div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Статус</label>
          <div class="valentina-detail-value">
            <span class="valentina-badge" :class="`valentina-badge-${selectedLead.status}`">{{ statusLabel(selectedLead.status) }}</span>
          </div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Согласие на обработку данных</label>
          <div class="valentina-detail-value">{{ selectedLead.privacy_notice ? 'Да' : 'Нет' }}</div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Сервисы</label>
          <div class="valentina-detail-value">{{ (selectedLead.requested_services || []).join(', ') || '—' }}</div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Местоположение</label>
          <div class="valentina-detail-value">{{ selectedLead.location_text || '—' }}</div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Бюджет</label>
          <div class="valentina-detail-value">
            <span v-if="selectedLead.budget !== undefined">{{ selectedLead.budget }} {{ selectedLead.budget_currency || '' }}</span>
            <span v-else>—</span>
          </div>
        </div>
        <div class="valentina-detail-field">
          <label class="valentina-detail-label">Контактное лицо</label>
          <div class="valentina-detail-value">
            <span v-if="selectedLead.contact_name">{{ selectedLead.contact_name }}</span>
            <span v-if="selectedLead.contact_email"> ({{ selectedLead.contact_email }})</span>
            <span v-if="selectedLead.contact_phone"> · {{ selectedLead.contact_phone }}</span>
            <span v-if="selectedLead.contact_purpose"> — {{ selectedLead.contact_purpose }}</span>
            <span v-if="!selectedLead.contact_name && !selectedLead.contact_email">—</span>
          </div>
        </div>
        <div class="valentina-detail-actions">
          <RouterLink v-if="selectedLead.status === 'qualified' || linkedProject" :to="linkedProject ? '/valentina/projects/' + linkedProject.project_id : { path: '/valentina/projects', query: { lead: selectedLead.lead_id } }">{{ linkedProject ? 'Открыть проект' : 'Создать проект из заявки' }}</RouterLink>
          <button type="button" class="valentina-btn" :class="`valentina-btn-${actionClass('triage')}`" @click="changeStatus('triage')" :disabled="!canTransition('triage')">В триаж</button>
          <button type="button" class="valentina-btn" :class="`valentina-btn-${actionClass('qualified')}`" @click="changeStatus('qualified')" :disabled="!canTransition('qualified')">Квалифицировать</button>
          <button type="button" class="valentina-btn" :class="`valentina-btn-${actionClass('declined')}`" @click="changeStatus('declined')" :disabled="!canTransition('declined')">Отклонить</button>
          <button type="button" class="valentina-btn" :class="`valentina-btn-${actionClass('spam')}`" @click="changeStatus('spam')" :disabled="!canTransition('spam')">Пометить спамом</button>
          <button type="button" class="valentina-btn" @click="convertToOpportunity" :disabled="selectedLead.status !== 'qualified'" v-if="selectedLead.status === 'qualified'">Создать возможность</button>
        </div>
      </div>
    </div>

    <div v-else class="valentina-leads-list">
      <div class="valentina-toolbar">
        <label for="lead-status">Статус</label>
        <select id="lead-status" v-model="filterStatus" @change="changeFilter">
          <option value="">Все статусы</option>
          <option v-for="status in statuses" :key="status" :value="status">{{ statusLabel(status) }}</option>
        </select>
        <span class="valentina-toolbar-count">{{ leads.length }} заявок</span>
      </div>

      <div v-if="error" class="valentina-error" role="alert">
        {{ error }}
        <button type="button" @click="load">Повторить</button>
      </div>
      <div v-else-if="loading" class="valentina-loading" role="status">Загрузка…</div>
      <div v-else-if="leads.length === 0" class="valentina-empty">
        Нет входящих заявок. Добавьте заявку вручную.
      </div>
      <div v-else class="valentina-list">
        <div v-for="l in filtered" :key="l.lead_id" class="valentina-lead-card" @click="selectLead(l.lead_id)">
          <div class="valentina-card-title">{{ l.request_summary }}</div>
          <div class="valentina-card-meta">
            <span class="valentina-badge">{{ l.inbound_channel }}</span>
            <span class="valentina-badge" :class="`valentina-badge-${l.status}`">{{ statusLabel(l.status) }}</span>
            <span class="valentina-cell-date">{{ formatDate(l.received_at) }}</span>
          </div>
          <div class="valentina-card-summary">{{ l.request_summary }}</div>
          <span class="valentina-card-open">Просмотреть →</span>
        </div>
      </div>

      <div class="valentina-pagination">
        <button type="button" :disabled="loading || offset === 0" @click="previousPage">Назад</button>
        <span>Страница {{ pageNumber }}</span>
        <button type="button" :disabled="loading || !hasMore" @click="nextPage">Далее</button>
      </div>
    </div>

    <div class="valentina-create-section">
      <h3 class="valentina-create-title">Создать входящую заявку вручную</h3>
      <form @submit.prevent="createLead" class="valentina-create-form">
        <div class="valentina-form-field">
          <label for="lead-summary">Описание заявки</label>
          <input id="lead-summary" v-model="newLead.request_summary" type="text" required placeholder="Что нужно клиенту" />
        </div>
        <div class="valentina-form-field">
          <label for="lead-channel">Канал</label>
          <input id="lead-channel" v-model="newLead.inbound_channel" type="text" required placeholder="например: веб-сайт, телефон" />
        </div>
        <div class="valentina-form-field">
          <label for="lead-consent">Согласие на обработку персональных данных</label>
          <input id="lead-consent" type="checkbox" v-model="newLead.privacy_notice" required />
        </div>
        <button type="submit" class="valentina-btn" :disabled="creating">{{ creating ? 'Создаётся…' : 'Создать заявку' }}</button>
        <div v-if="createError" class="valentina-error">{{ createError }}</div>
        <div v-if="createSuccess" class="valentina-success">Заявка "{{ createdLeadSummary }}" создана.</div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { apiFetch } from '../composables/useApi'

const PAGE_SIZE = 25
const filterStatus = ref('')
const offset = ref(0)
const hasMore = ref(false)
const loading = ref(false)
const error = ref(null)
const leads = ref([])

const selectedLead = ref(null)
const linkedProject = ref(null)
const route = useRoute()
const detailLoading = ref(false)
const detailError = ref(null)

const creating = ref(false)
const createError = ref(null)
const createSuccess = ref(false)
const createdLeadSummary = ref('')

const newLead = ref({
  request_summary: '',
  inbound_channel: '',
  privacy_notice: false
})

const statuses = ['new', 'triage', 'needs_information', 'qualified', 'converted', 'declined', 'spam', 'archived']

const statusLabels = {
  'new': 'Новое',
  'triage': 'Триаж',
  'needs_information': 'Требует информации',
  'qualified': 'Квалифицировано',
  'converted': 'Сконвертировано',
  'declined': 'Отклонить',
  'spam': 'Спам',
  'archived': 'В архиве'
}

const actionClasses = {
  'triage': 'primary',
  'qualified': 'success',
  'declined': 'danger',
  'spam': 'neutral'
}

const actionClass = (status) => actionClasses[status] || 'neutral'

const validTransitions = {
  'new': ['triage', 'declined', 'spam'],
  'triage': ['needs_information', 'qualified', 'declined', 'spam'],
  'needs_information': ['triage', 'qualified', 'declined', 'spam'],
  'qualified': ['converted', 'declined', 'spam', 'archived'],
  'converted': ['archived'],
  'declined': ['archived'],
  'spam': ['archived'],
  'archived': []
}

const canTransition = (newStatus) => {
  if (!selectedLead.value) return false
  return validTransitions[selectedLead.value.status]?.includes(newStatus) || false
}

const filtered = computed(() => {
  return leads.value.filter(l => !filterStatus.value || l.status === filterStatus.value)
})

const statusLabel = (s) => statusLabels[s] || s

const formatDate = (d) => d ? new Date(d).toLocaleDateString('ru-RU') : '—'

const pageNumber = computed(() => Math.floor(offset.value / PAGE_SIZE) + 1)

const load = async () => {
  loading.value = true
  error.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/leads?limit=' + PAGE_SIZE + '&offset=' + offset.value)
    leads.value = res.leads || []
    hasMore.value = Boolean(res.has_more)
  } catch (e) {
    error.value = 'Не удалось загрузить заявки'
    console.error(e)
  } finally {
    loading.value = false
  }
}

const changeFilter = () => {
  offset.value = 0
  load()
}

const nextPage = () => {
  if (!loading.value && hasMore.value) {
    offset.value += PAGE_SIZE
    load()
  }
}

const previousPage = () => {
  if (!loading.value && offset.value > 0) {
    offset.value = Math.max(0, offset.value - PAGE_SIZE)
    load()
  }
}

const selectLead = async (leadID) => {
  detailLoading.value = true
  detailError.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/leads/' + encodeURIComponent(leadID))
    selectedLead.value = res.lead || res
    linkedProject.value = null
    const linked = await apiFetch('/api/v1/valentina/leads/' + encodeURIComponent(leadID) + '/project')
    linkedProject.value = linked.project || null
  } catch (e) {
    detailError.value = 'Не удалось загрузить детали заявки'
    console.error(e)
  } finally {
    detailLoading.value = false
  }
}

watch(() => route?.query?.lead, (id) => { if (typeof id === 'string') selectLead(id) }, { immediate: true })

const changeStatus = async (newStatus) => {
  if (!selectedLead.value || !canTransition(newStatus)) return
  detailLoading.value = true
  detailError.value = null
  try {
    await apiFetch('/api/v1/valentina/review-decisions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        entity_type: 'inbound_lead',
        entity_id: selectedLead.value.lead_id,
        new_status: newStatus,
        decision_reason: 'Статус изменён через интерфейс заявок'
      })
    })
    await load()
    selectedLead.value = null
  } catch (e) {
    detailError.value = 'Не удалось изменить статус'
    console.error(e)
  } finally {
    detailLoading.value = false
  }
}

const convertToOpportunity = async () => {
  if (!selectedLead.value || selectedLead.value.status !== 'qualified') return
  detailLoading.value = true
  detailError.value = null
  try {
    await apiFetch('/api/v1/valentina/opportunities', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: 'Из входящей заявки: ' + selectedLead.value.request_summary,
        summary: 'Создано из заявки ' + selectedLead.value.lead_id,
        opportunity_type: 'direct_request',
        status: 'candidate',
        demand_evidence_type: 'direct_request',
        source_evidence_ids: [],
        company_id: selectedLead.value.company_id || undefined
      })
    })
    await load()
    selectedLead.value = null
  } catch (e) {
    detailError.value = 'Не удалось создать возможность'
    console.error(e)
  } finally {
    detailLoading.value = false
  }
}

const createLead = async () => {
  creating.value = true
  createError.value = null
  createSuccess.value = false
  createdLeadSummary.value = ''
  try {
    const res = await apiFetch('/api/v1/valentina/leads', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newLead.value)
    })
    createdLeadSummary.value = newLead.value.request_summary
    createSuccess.value = true
    newLead.value = { request_summary: '', inbound_channel: '', privacy_notice: false }
    await load()
  } catch (e) {
    createError.value = 'Не удалось создать заявку'
    console.error(e)
  } finally {
    creating.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.valentina-page { padding: 24px; max-width: 1000px; margin: 0 auto; }
.valentina-page-header { margin-bottom: 24px; }
.valentina-page-title { font-size: 24px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-page-subtitle { font-size: 14px; color: var(--valentina-muted); margin: 0; }

.valentina-leads-list, .valentina-lead-detail { display: flex; flex-direction: column; gap: 16px; }

.valentina-toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 18px; flex-wrap: wrap; }
.valentina-toolbar label,.valentina-toolbar-count { color: var(--valentina-muted); font-size: 13px; }
.valentina-toolbar select { min-width: 190px; padding: 8px 10px; border: 1px solid var(--valentina-border); border-radius: 7px; color: var(--valentina-text); background: var(--valentina-surface); }

.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,.1); border-radius: 6px; margin-bottom: 16px; }
.valentina-error button { margin-left: 12px; }
.valentina-loading,.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); text-align: center; line-height: 1.5; }

.valentina-list { display: flex; flex-direction: column; gap: 12px; }
.valentina-lead-card {
  padding: 16px;
  border: 1px solid var(--valentina-border);
  border-radius: 8px;
  background: var(--valentina-surface);
  cursor: pointer;
  transition: border-color .15s;
}
.valentina-lead-card:hover { border-color: var(--valentina-blue); }
.valentina-card-title { font-size: 16px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-card-meta { display: flex; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; align-items: center; }
.valentina-card-summary { font-size: 13px; color: var(--valentina-text); line-height: 1.5; }
.valentina-card-open { display: inline-block; margin-top: 12px; color: var(--valentina-blue); font-size: 12px; }

.valentina-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}
.valentina-badge.new { background: var(--valentina-blue); color: #ffffff; }
.valentina-badge.qualified { background: var(--valentina-green); color: #001a14; }
.valentina-badge.declined { background: var(--valentina-red); color: #ffffff; }
.valentina-badge.spam { background: var(--valentina-muted); color: #ffffff; }
.valentina-badge.converted { background: var(--valentina-violet); color: #ffffff; }
.valentina-badge.triage { background: var(--valentina-amber); color: #001a14; }
.valentina-badge.needs_information { background: var(--valentina-amber); color: #001a14; }
.valentina-badge.archived { background: var(--valentina-muted); color: #ffffff; }

.valentina-pagination { display: flex; align-items: center; justify-content: center; gap: 18px; margin: 20px 0 8px; color: var(--valentina-muted); font-size: 13px; }
.valentina-pagination button { padding: 8px 13px; border: 1px solid var(--valentina-border); border-radius: 6px; color: var(--valentina-text); background: var(--valentina-surface); cursor: pointer; }
.valentina-pagination button:disabled { opacity: .4; cursor: not-allowed; }

.valentina-create-section { margin-top: 24px; padding: 20px; border: 1px solid var(--valentina-border); border-radius: 8px; background: var(--valentina-surface); }
.valentina-create-title { font-size: 18px; font-weight: 600; margin: 0 0 16px 0; color: var(--valentina-text); }
.valentina-create-form { display: flex; flex-direction: column; gap: 12px; }
.valentina-form-field { display: flex; flex-direction: column; gap: 6px; }
.valentina-form-field label { font-size: 13px; color: var(--valentina-muted); font-weight: 500; }
.valentina-form-field input[type="text"] { padding: 8px 12px; border: 1px solid var(--valentina-border); border-radius: 6px; background: var(--valentina-bg); color: var(--valentina-text); font-size: 14px; }
.valentina-form-field input[type="checkbox"] { margin: 4px 0; width: 16px; height: 16px; }
.valentina-btn { padding: 10px 18px; border: 1px solid var(--valentina-border); border-radius: 6px; background: var(--valentina-accent); color: #001a14; font-weight: 600; cursor: pointer; font-size: 14px; }
.valentina-btn:disabled { opacity: .6; cursor: not-allowed; }
.valentina-btn-primary { background: var(--valentina-blue); color: #ffffff; }
.valentina-btn-success { background: var(--valentina-green); color: #001a14; }
.valentina-btn-danger { background: var(--valentina-red); color: #ffffff; }
.valentina-btn-neutral { background: var(--valentina-muted); color: #ffffff; }

.valentina-detail-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.valentina-btn-back { padding: 6px 14px; font-size: 13px; }
.valentina-detail-title { margin: 0; font-size: 20px; color: var(--valentina-text); }
.valentina-lead-detail-content { display: flex; flex-direction: column; gap: 14px; }
.valentina-detail-field { display: flex; gap: 16px; }
.valentina-detail-label { width: 200px; color: var(--valentina-muted); font-size: 13px; font-weight: 500; }
.valentina-detail-value { color: var(--valentina-text); font-size: 14px; }
.valentina-detail-actions { display: flex; gap: 10px; margin-top: 24px; flex-wrap: wrap; }
</style>
