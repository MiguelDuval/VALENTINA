<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Входящие заявки</h1>
      <p class="valentina-page-subtitle">Клиентские запросы, поступившие через каналы связи бизнеса</p>
    </div>
    <div class="valentina-leads-content">
      <div v-if="loading" class="valentina-loading">Загрузка...</div>
      <div v-else-if="leads.length === 0" class="valentina-empty">Нет входящих заявок. Добавьте заявку вручную.</div>
      <div v-else class="valentina-list">
        <div v-for="l in filtered" :key="l.lead_id" class="valentina-lead-card">
          <div class="valentina-card-title">{{ l.request_summary }}</div>
          <div class="valentina-card-meta">
            <span class="valentina-badge">{{ l.inbound_channel }}</div>
            <span class="valentina-badge">{{ statusLabel(l.status) }}</div>
            <span class="valentina-cell-date">{{ formatDate(l.received_at) }}</div>
          </div>
          <div class="valentina-card-summary">{{ l.request_summary }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

const filterStatus = ref('')
const loading = ref(false)
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
  try {
    const res = await fetch('/api/v1/valentina/leads')
    const data = await res.json()
    leads.value = data.leads || []
  } catch (e) {
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
</style>

