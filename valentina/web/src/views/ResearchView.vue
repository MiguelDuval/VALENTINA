<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Исследования</h1>
      <p class="valentina-page-subtitle">История исследований, прогресс запусков, источники сбоев и результаты проверки</p>
    </div>
    <div v-if="error" class="valentina-error">{{ error }}</div>
    <div class="valentina-research-content">
      <div v-if="loading" class="valentina-loading">Загрузка...</div>
      <div v-else-if="runs.length === 0" class="valentina-empty">Нет запусков исследований. Запустите ручное исследование.</div>
      <div v-else class="valentina-list">
        <div v-for="r in runs" :key="r.run_id" class="valentina-run-card">
          <div class="valentina-card-title">Исследование {{ r.run_id.slice(-8) }}</div>
          <div class="valentina-card-meta">
            <span class="valentina-badge" :class="`valentina-run-status-${r.status}`">{{ r.status }}</span>
            <span class="valentina-cell-date">{{ formatDate(r.started_at) }}</span>
          </div>
          <div class="valentina-run-stats">
            <span class="valentina-stat"><strong>{{ r.candidate_count }}</strong> кандидатов</span>
            <span class="valentina-stat"><strong>{{ r.verified_count }}</strong> проверено</span>
            <span class="valentina-stat"><strong>{{ r.duplicate_count }}</strong> дубликатов</span>
            <span class="valentina-stat"><strong>{{ r.accepted_count }}</strong> принято</span>
            <span class="valentina-stat"><strong>{{ r.rejected_count }}</strong> отклонено</span>
          </div>
          <div v-if="r.notes" class="valentina-card-notes">{{ r.notes }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../composables/useApi'

const loading = ref(false)
const error = ref(null)
const runs = ref([])

const load = async () => {
  loading.value = true
  error.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/research-runs')
    runs.value = res.runs || []
  } catch (e) {
    error.value = 'Не удалось загрузить исследования'
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)

const formatDate = (d) => d ? new Date(d).toLocaleDateString('ru-RU') : '—'
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
.valentina-run-status.running { background: var(--valentina-blue); color: #ffffff; }
.valentina-run-status.completed { background: var(--valentina-green); color: #001a14; }
.valentina-run-status.failed { background: var(--valentina-red); color: #ffffff; }
.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,0.1); border-radius: 6px; margin-bottom: 16px; }
</style>
