<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Холодные возможности</h1>
      <p class="valentina-page-subtitle">Общедоступные коммерческие возможности с проверенными источниками</p>
    </div>
    <div v-if="error" class="valentina-error">{{ error }}</div>
    <div v-else-if="loading" class="valentina-loading">Загрузка...</div>
    <div v-else-if="opps.length === 0" class="valentina-empty">Нет возможностей. Запустите исследование.</div>
    <div v-else class="valentina-opportunities-content">
      <div v-for="o in filtered" :key="o.opportunity_id" class="valentina-opportunity-card">
        <div class="valentina-card-title">{{ o.title }}</div>
        <div class="valentina-card-meta">
          <span class="valentina-badge">{{ demandLabel(o.demand_evidence_type) }}</span>
          <span class="valentina-badge">{{ statusLabel(o.status) }}</span>
          <span class="valentina-cell-location">{{ o.geography?.municipality || '—' }}</span>
        </div>
        <div class="valentina-card-summary">{{ o.summary }}</div>
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
const opps = ref([])

const demandLabels = {
  'direct_request': 'Прямой запрос',
  'public_tender': 'Публичный тендер',
  'subcontracting_request': 'Субподряд',
  'early_demand_signal': 'Ранний сигнал',
  'unknown': 'Неизвестно'
}

const statusLabels = {
  'candidate': 'Кандидат',
  'verified': 'Проверено',
  'needs_review': 'Требует проверки',
  'qualified': 'Квалифицировано',
  'rejected': 'Отклонено',
  'expired': 'Истекло',
  'archived': 'В архиве'
}

const filtered = computed(() => {
  return opps.value.filter(o => !filterStatus.value || o.status === filterStatus.value)
})

const demandLabel = (t) => demandLabels[t] || t
const statusLabel = (s) => statusLabels[s] || s

const load = async () => {
  loading.value = true
  error.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/opportunities')
    opps.value = res.opportunities || []
  } catch (e) {
    error.value = 'Не удалось загрузить возможности'
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
.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,0.1); border-radius: 6px; margin-bottom: 16px; }
.valentina-loading { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-opportunities-content { display: flex; flex-direction: column; gap: 16px; }
.valentina-opportunity-card {
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
.valentina-badge.early_demand_signal { background: var(--valentina-amber); color: #001a14; }
.valentina-badge.direct_request { background: var(--valentina-green); color: #001a14; }
.valentina-badge.public_tender { background: var(--valentina-blue); color: #ffffff; }
.valentina-badge.subcontracting_request { background: var(--valentina-violet); color: #ffffff; }
.valentina-badge.unknown { background: var(--valentina-muted); color: #ffffff; }
</style>

