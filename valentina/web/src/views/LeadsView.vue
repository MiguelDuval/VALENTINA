<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Входящие заявки</h1>
      <p class="valentina-page-subtitle">Клиентские запросы, поступившие через каналы связи бизнеса</p>
    </div>
    <div class="valentina-leads-content">
      <div v-if="error" class="valentina-error">{{ error }}</div>
    <div v-if="loading" class="valentina-loading">Загрузка...</div>
      <div v-else-if="leads.length === 0" class="valentina-empty">Нет входящих заявок. Добавьте заявку вручную.</div>
      <div v-else class="valentina-list">
        <div v-for="l in filtered" :key="l.lead_id" class="valentina-lead-card">
          <div class="valentina-card-title">{{ l.request_summary }}</div>
          <div class="valentina-card-meta">
            <span class="valentina-badge">{{ l.inbound_channel }}</span>
            <span class="valentina-badge">{{ statusLabel(l.status) }}</span>
            <span class="valentina-cell-date">{{ formatDate(l.received_at) }}</span>
          </div>
          <div class="valentina-card-summary">{{ l.request_summary }}</div>
        </div>
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
        <button type="submit" class="valentina-btn" :disabled="creating">{{ creating ? 'Создаётся…' : 'Создать заявку' }}</button>
        <div v-if="createError" class="valentina-error">{{ createError }}</div>
        <div v-if="createSuccess" class="valentina-success">Заявка "{{ createdLeadSummary }}" создана.</div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { apiFetch } from '../composables/useApi'

const filterStatus = ref('')
const loading = ref(false)
const error = ref(null)
const leads = ref([])

const creating = ref(false)
const createError = ref(null)
const createSuccess = ref(false)
const createdLeadSummary = ref('')

const newLead = ref({
  request_summary: '',
  inbound_channel: ''
})

const statusLabels = {
  'new': 'Новое',
  'triage': 'Триаж',
  'needs_information': 'Требует информации',
  'qualified': 'Квалифицировано',
  'converted': 'Сконвертировано',
  'declined': 'Отклонено',
  'spam': 'Спам',
  'archived': 'В архиве'
}

const filtered = computed(() => {
  return leads.value.filter(l => !filterStatus.value || l.status === filterStatus.value)
})

const statusLabel = (s) => statusLabels[s] || s

const formatDate = (d) => d ? new Date(d).toLocaleDateString('ru-RU') : '—'

const load = async () => {
  loading.value = true
  error.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/leads')
    leads.value = res.leads || []
  } catch (e) {
    error.value = 'Не удалось загрузить заявки'
    console.error(e)
  } finally {
    loading.value = false
  }
}

const createLead = async () => {
  creating = true
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
    newLead.value = { request_summary: '', inbound_channel: '' }
    await load()
  } catch (e) {
    createError.value = 'Не удалось создать заявку'
    console.error(e)
  } finally {
    creating = false
  }
}

onMounted(load)
</script>

<style scoped>
.valentina-page { padding: 24px; max-width: 1000px; margin: 0 auto; }
.valentina-page-header { margin-bottom: 24px; }
.valentina-page-title { font-size: 24px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-page-subtitle { font-size: 14px; color: var(--valentina-muted); margin: 0; }
.valentina-leads-content { display: flex; flex-direction: column; gap: 16px; }
.valentina-loading { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-list { display: flex; flex-direction: column; gap: 12px; }
.valentina-lead-card {
  padding: 16px;
  border: 1px solid var(--valentina-border);
  border-radius: 8px;
  background: var(--valentina-surface);
}
.valentina-card-title { font-size: 16px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-card-meta { display: flex; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.valentina-card-summary { font-size: 13px; color: var(--valentina-text); line-height: 1.5; }
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
.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,0.1); border-radius: 6px; margin-bottom: 16px; }
.valentina-success { padding: 16px; color: var(--valentina-green); background: rgba(89,214,166,0.1); border-radius: 6px; margin-bottom: 16px; }
.valentina-create-section { margin-top: 24px; padding: 20px; border: 1px solid var(--valentina-border); border-radius: 8px; background: var(--valentina-surface); }
.valentina-create-title { font-size: 18px; font-weight: 600; margin: 0 0 16px 0; color: var(--valentina-text); }
.valentina-create-form { display: flex; flex-direction: column; gap: 12px; }
.valentina-form-field { display: flex; flex-direction: column; gap: 6px; }
.valentina-form-field label { font-size: 13px; color: var(--valentina-muted); font-weight: 500; }
.valentina-form-field input { padding: 8px 12px; border: 1px solid var(--valentina-border); border-radius: 6px; background: var(--valentina-bg); color: var(--valentina-text); font-size: 14px; }
.valentina-btn { padding: 10px 18px; border: 1px solid var(--valentina-border); border-radius: 6px; background: var(--valentina-accent); color: #001a14; font-weight: 600; cursor: pointer; font-size: 14px; }
.valentina-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>

