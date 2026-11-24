<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Исследования</h1>
      <p class="valentina-page-subtitle">История исследований, прогресс запусков, источники сбоев и результаты проверки</p>
    </div>

    <div class="valentina-toolbar">
      <button type="button" class="valentina-btn" @click="startRun" :disabled="loading || running || !hasApprovedSources">{{ running ? 'Исследование выполняется…' : 'Запустить исследование' }}</button>
      <button type="button" class="valentina-btn" @click="refresh" :disabled="loading">Обновить</button>
      <span v-if="!hasApprovedSources" class="valentina-toolbar-count" style="color: var(--valentina-red)">Нет утверждённых источников — настройте их на странице Источники</span>
      <span v-else class="valentina-toolbar-count">Ручной запуск утверждённых RSS/Atom. Результаты требуют ручной проверки.</span>
    </div>
    <p v-for="s in sources.filter(s => !s.manual_run_allowed)" :key="s.source_id" class="valentina-toolbar-count">{{ s.name }}: {{ s.manual_run_reason || 'Запуск не разрешён сервером' }}</p>
    <p class="valentina-toolbar-count">Наблюдение в ленте — не подтверждённый заказ. Оригинальный URL, время наблюдения и выдержки доступны в карточке кандидата: <a href="#/valentina/opportunities">Открыть кандидатов</a>.</p>

    <p v-if="notice" role="status">{{ notice }}</p>
    <div v-if="error" role="alert" class="valentina-error">{{ error }}</div>
    <div class="valentina-research-content">
      <div v-if="loading" class="valentina-loading">Загрузка...</div>
      <div v-else-if="runs.length === 0" class="valentina-empty">Запусков пока нет. Выберите утверждённые источники на странице Источники.</div>
      <div v-else class="valentina-list">
        <div v-for="r in runs" :key="r.run_id" class="valentina-run-card">
          <div class="valentina-card-title">Исследование {{ r.run_id.slice(-8) }}</div>
          <div class="valentina-card-meta">
            <span class="valentina-badge" :class="`valentina-run-status-${r.status}`">{{ statusLabel(r.status) }}</span>
            <span class="valentina-cell-date">{{ formatDate(r.started_at) }}</span>
            <span v-if="r.completed_at">Завершено: {{ formatDate(r.completed_at) }}</span>
          </div>
          <div class="valentina-run-stats">
            <span class="valentina-stat"><strong>{{ r.candidate_count }}</strong> кандидатов</span>
            <span class="valentina-stat"><strong>{{ r.verified_count }}</strong> проверено</span>
            <span class="valentina-stat"><strong>{{ r.duplicate_count }}</strong> дубликатов</span>
            <span class="valentina-stat"><strong>{{ r.accepted_count }}</strong> принято</span>
            <span class="valentina-stat"><strong>{{ r.rejected_count }}</strong> отклонено</span>
          </div>
          <div v-if="r.status === 'running'" class="valentina-run-actions">
            <button type="button" class="valentina-btn valentina-btn-danger" @click="cancelRun(r.run_id)">Отмена</button>
          </div>
          <div v-else-if="r.status === 'completed' || r.status === 'failed'" class="valentina-run-actions">
            <button type="button" class="valentina-btn" @click="startRun" :disabled="loading || running || !hasApprovedSources">Новый запуск</button>
          </div>
          <div v-if="r.notes" class="valentina-card-notes">{{ r.notes }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { apiFetch } from '../composables/useApi'

const loading = ref(false)
const error = ref(null)
const runs = ref([])
const sources = ref([])
const running = ref(false)
const notice = ref(null)
const hasApprovedSources = computed(() => sources.value.some(s => s.manual_run_allowed === true))

const load = async (preserveError = false) => {
  loading.value = true
  if (!preserveError) error.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/research-runs')
    runs.value = res.runs || []
  } catch (e) {
    error.value = 'Не удалось загрузить исследования'
  } finally {
    loading.value = false
  }
}

const loadSources = async () => {
  try {
    const res = await apiFetch('/api/v1/valentina/sources')
    sources.value = res.sources || []
  } catch (e) {
    sources.value = []
    error.value = 'Не удалось загрузить разрешения источников'
  }
}

onMounted(async () => {
  await load()
  await loadSources()
})

const refresh = async () => { await load(); await loadSources() }
const statusLabel = status => ({ running: 'Выполняется', completed: 'Завершено', failed: 'Ошибка / прервано', cancelled: 'Отменено' }[status] || status)
const runError = e => ({
  'research source busy': 'Источник уже используется. Дождитесь завершения и обновите историю.',
  'research quota unavailable': 'Достигнут дневной лимит или ещё не прошёл минимальный интервал. Повторите вручную позже.',
  'research run is not active': 'Запуск уже завершён или не выполняется. Обновите историю.',
  'research policy invalid': 'Проверьте утверждение и ограничения источников.',
}[e?.message] || 'Не удалось запустить исследование')

const startRun = async () => {
  if (loading.value || running.value || !hasApprovedSources.value) return
  running.value = true
  error.value = null
  notice.value = null
  try {
    const approvedSourceIds = sources.value
      .filter(s => s.manual_run_allowed === true)
      .map(s => s.source_id)
    const res = await apiFetch('/api/v1/valentina/research-runs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ source_ids: approvedSourceIds, notes: 'Manual research run' })
    })
    if (res.error && res.status === 'blocked') {
      error.value = 'Исследование заблокировано: проверьте разрешения источников'
      return
    }
    await load()
  } catch (e) {
    error.value = runError(e)
    await load(true)
  } finally {
    running.value = false
    await loadSources()
  }
}

const cancelRun = async (runID) => {
  loading.value = true
  error.value = null
  try {
    await apiFetch('/api/v1/valentina/research-runs/' + encodeURIComponent(runID) + '/cancel', {
      method: 'POST'
    })
    notice.value = 'Отмена запрошена. Источник остаётся занят до остановки; обновите историю.'
    await load()
  } catch (e) {
    error.value = e?.message === 'research run is not active' ? runError(e) : 'Не удалось отменить исследование'
    await load(true)
  } finally {
    loading.value = false
  }
}

const formatDate = (d) => d ? new Date(d).toLocaleString('ru-RU') : '—'
</script>

<style scoped>
.valentina-page { padding: 24px; max-width: 1000px; margin: 0 auto; }
.valentina-page-header { margin-bottom: 24px; }
.valentina-page-title { font-size: 24px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-page-subtitle { font-size: 14px; color: var(--valentina-muted); margin: 0; }
.valentina-research-content { display: flex; flex-direction: column; gap: 16px; }
.valentina-loading { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-list { display: flex; flex-direction: column; gap: 12px; }
.valentina-run-card {
  padding: 16px;
  border: 1px solid var(--valentina-border);
  border-radius: 8px;
  background: var(--valentina-surface);
}
.valentina-card-title { font-size: 16px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-card-meta { display: flex; gap: 8px; margin-bottom: 12px; flex-wrap: wrap; }
.valentina-card-notes { font-size: 13px; color: var(--valentina-text); margin-top: 12px; }
.valentina-run-stats { display: flex; gap: 16px; margin-top: 12px; }
.valentina-stat { font-size: 13px; color: var(--valentina-text); }
.valentina-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}
.valentina-run-status-running { background: var(--valentina-blue); color: #ffffff; }
.valentina-run-status-completed { background: var(--valentina-green); color: #001a14; }
.valentina-run-status-failed { background: var(--valentina-red); color: #ffffff; }
.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,0.1); border-radius: 6px; margin-bottom: 16px; }
.valentina-toolbar { display: flex; align-items: center; gap: 10px; margin-bottom: 18px; flex-wrap: wrap; }
.valentina-toolbar-count { color: var(--valentina-muted); font-size: 13px; }
.valentina-run-actions { display: flex; gap: 8px; margin-top: 12px; }
.valentina-btn-danger { background: var(--valentina-red); color: #ffffff; }
</style>
