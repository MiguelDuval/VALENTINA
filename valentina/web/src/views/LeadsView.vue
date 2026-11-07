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
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { apiFetch } from '../composables/useApi'

const filterStatus = ref('')
const loading = ref(false)
const error = ref(null)
const leads = ref([])

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
</style>

