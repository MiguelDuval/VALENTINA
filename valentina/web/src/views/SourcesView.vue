<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Источники</h1>
      <p class="valentina-page-subtitle">Утвержденные источники сбора данных и их состояние</p>
    </div>
    <div v-if="error" class="valentina-error">{{ error }}</div>
    <div class="valentina-sources-content">
      <div v-if="loading" class="valentina-loading">Загрузка...</div>
      <div v-else-if="sources.length === 0" class="valentina-empty">Нет утвержденных источников. Источники добавляются вручную через реестр источников.</div>
      <div v-else class="valentina-list">
        <div v-for="s in sources" :key="s.source_id" class="valentina-source-card">
          <div class="valentina-card-title">{{ s.name }}</div>
          <div class="valentina-card-meta">
            <span class="valentina-badge">{{ s.source_type }}</span>
            <span class="valentina-badge" :class="`valentina-approval-${s.approval_status}`">{{ s.approval_status }}</span>
            <span class="valentina-cell-url">{{ s.base_url }}</span>
          </div>
          <div class="valentina-card-publisher">{{ s.publisher }}</div>
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
const sources = ref([])

const load = async () => {
  loading.value = true
  error.value = null
  try {
    const res = await apiFetch('/api/v1/valentina/sources')
    sources.value = res.sources || []
  } catch (e) {
    error.value = 'Не удалось загрузить источники'
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
.valentina-sources-content { display: flex; flex-direction: column; gap: 16px; }
.valentina-loading { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-list { display: flex; flex-direction: column; gap: 12px; }
.valentina-source-card {
  padding: 16px;
  border: 1px solid var(--valentina-border);
  border-radius: 8px;
  background: var(--valentina-surface);
}
.valentina-card-title { font-size: 16px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-card-meta { display: flex; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.valentina-card-publisher { font-size: 13px; color: var(--valentina-muted); }
.valentina-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}
.valentina-approval-approved { background: var(--valentina-green); color: #001a14; }
.valentina-approval-pending_review { background: var(--valentina-amber); color: #001a14; }
.valentina-approval-blocked { background: var(--valentina-red); color: #ffffff; }
.valentina-approval-paused { background: var(--valentina-muted); color: #ffffff; }
.valentina-error { padding: 16px; color: var(--valentina-red); background: rgba(255,126,136,0.1); border-radius: 6px; margin-bottom: 16px; }
</style>
