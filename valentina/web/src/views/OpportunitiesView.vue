<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Холодные возможности</h1>
      <p class="valentina-page-subtitle">Общедоступные коммерческие возможности с проверяемыми первоисточниками</p>
    </div>

    <div class="valentina-toolbar">
      <label for="opportunity-status">Статус</label>
      <select id="opportunity-status" v-model="filterStatus" @change="changeFilter">
        <option value="">Все статусы</option>
        <option v-for="status in statuses" :key="status" :value="status">{{ statusLabel(status) }}</option>
      </select>
      <span class="valentina-toolbar-count">{{ opps.length }} на этой странице</span>
    </div>

    <div v-if="error" class="valentina-error" role="alert">
      {{ error }}
      <button type="button" @click="load">Повторить</button>
    </div>
    <div v-else-if="loading" class="valentina-loading" role="status">Загрузка...</div>
    <div v-else-if="opps.length === 0" class="valentina-empty">
      Нет сохранённых возможностей для выбранного фильтра. Возможности появятся после сохранения проверяемых данных из разрешённых источников.
    </div>
    <div v-else class="valentina-opportunities-content">
      <RouterLink
        v-for="o in opps"
        :key="o.opportunity_id"
        :to="'/valentina/opportunities/' + encodeURIComponent(o.opportunity_id)"
        class="valentina-opportunity-card"
        :aria-label="'Открыть возможность: ' + o.title"
      >
        <div class="valentina-card-title">{{ o.title }}</div>
        <div class="valentina-card-meta">
          <span class="valentina-badge">{{ demandLabel(o.demand_evidence_type) }}</span>
          <span class="valentina-badge valentina-badge-muted">{{ statusLabel(o.status) }}</span>
          <span class="valentina-cell-location">{{ o.geography?.municipality || 'Местоположение не указано' }}</span>
          <span class="valentina-cell-location">{{ (o.source_evidence_ids || []).length }} доказательств</span>
        </div>
        <div class="valentina-card-summary">{{ o.summary || 'Описание не сохранено.' }}</div>
        <span class="valentina-card-open">Открыть карточку →</span>
      </RouterLink>
    </div>

    <div class="valentina-pagination">
      <button type="button" :disabled="loading || offset === 0" @click="previousPage">Назад</button>
      <span>Страница {{ pageNumber }}</span>
      <button type="button" :disabled="loading || !hasMore" @click="nextPage">Далее</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { apiFetch } from '../composables/useApi'

const PAGE_SIZE = 25
const statuses = ['candidate', 'verified', 'needs_review', 'qualified', 'rejected', 'expired', 'archived']
const filterStatus = ref('')
const offset = ref(0)
const hasMore = ref(false)
const loading = ref(false)
const error = ref('')
const opps = ref([])
const pageNumber = computed(() => Math.floor(offset.value / PAGE_SIZE) + 1)

const demandLabels = {
  direct_request: 'Прямой запрос',
  public_tender: 'Публичный тендер',
  subcontracting_request: 'Субподряд',
  early_demand_signal: 'Ранний сигнал',
  unknown: 'Неизвестно',
}
const statusLabels = {
  candidate: 'Кандидат',
  verified: 'Проверено',
  needs_review: 'Требует проверки',
  qualified: 'Квалифицировано',
  rejected: 'Отклонено',
  expired: 'Истекло',
  archived: 'В архиве',
}
const demandLabel = value => demandLabels[value] || value || '—'
const statusLabel = value => statusLabels[value] || value || '—'

const load = async () => {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ limit: String(PAGE_SIZE), offset: String(offset.value) })
    if (filterStatus.value) params.set('status', filterStatus.value)
    const response = await apiFetch('/api/v1/valentina/opportunities?' + params.toString())
    opps.value = response.opportunities || []
    hasMore.value = Boolean(response.has_more)
  } catch (err) {
    error.value = 'Не удалось загрузить возможности: ' + (err?.message || 'ошибка сервера')
    opps.value = []
    hasMore.value = false
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

onMounted(load)
</script>

<style scoped>
.valentina-page { padding: 24px; max-width: 1000px; margin: 0 auto; }
.valentina-page-header { margin-bottom: 24px; }
.valentina-page-title { font-size: 24px; font-weight: 600; margin: 0 0 8px; color: var(--valentina-text); }
.valentina-page-subtitle { font-size: 14px; color: var(--valentina-muted); margin: 0; }
.valentina-toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 18px; flex-wrap: wrap; }
.valentina-toolbar label,.valentina-toolbar-count { color: var(--valentina-muted); font-size: 13px; }
.valentina-toolbar select { min-width: 190px; padding: 8px 10px; border: 1px solid var(--valentina-border); border-radius: 7px; color: var(--valentina-text); background: var(--valentina-surface); }
.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,.1); border-radius: 6px; margin-bottom: 16px; }
.valentina-error button { margin-left: 12px; }
.valentina-loading,.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); text-align: center; line-height: 1.5; }
.valentina-opportunities-content { display: flex; flex-direction: column; gap: 12px; }
.valentina-opportunity-card { display: block; padding: 16px; border: 1px solid var(--valentina-border); border-radius: 8px; background: var(--valentina-surface); color: inherit; text-decoration: none; transition: border-color .15s, transform .15s; }
.valentina-opportunity-card:hover,.valentina-opportunity-card:focus-visible { border-color: var(--valentina-blue); outline: none; transform: translateY(-1px); }
.valentina-card-title { font-size: 16px; font-weight: 600; margin: 0 0 8px; color: var(--valentina-text); }
.valentina-card-meta { display: flex; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; align-items: center; }
.valentina-card-summary { font-size: 13px; color: var(--valentina-text); line-height: 1.5; }
.valentina-card-open { display: inline-block; margin-top: 12px; color: var(--valentina-blue); font-size: 12px; }
.valentina-badge { font-size: 11px; padding: 3px 8px; border-radius: 4px; background: var(--valentina-green); color: #001a14; }
.valentina-badge-muted { background: rgba(130,145,160,.18); color: var(--valentina-text); }
.valentina-cell-location { font-size: 12px; color: var(--valentina-muted); }
.valentina-pagination { display: flex; align-items: center; justify-content: center; gap: 18px; margin: 20px 0 8px; color: var(--valentina-muted); font-size: 13px; }
.valentina-pagination button { padding: 8px 13px; border: 1px solid var(--valentina-border); border-radius: 6px; color: var(--valentina-text); background: var(--valentina-surface); cursor: pointer; }
.valentina-pagination button:disabled { opacity: .4; cursor: not-allowed; }
</style>
