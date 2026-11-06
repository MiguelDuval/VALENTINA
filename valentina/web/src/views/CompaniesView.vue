<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <h1 class="valentina-page-title">Компании</h1>
      <p class="valentina-page-subtitle">Организации, издатели и клиенты, связанные с возможностями и заявками</p>
    </div>
    <div class="valentina-companies-content">
      <div v-if="loading" class="valentina-loading">Загрузка...</div>
      <div v-else-if="companies.length === 0" class="valentina-empty">Нет компаний. Добавьте компанию вручную.</div>
      <div v-else class="valentina-list">
        <div v-for="c in companies" :key="c.company_id" class="valentina-company-card">
          <div class="valentina-card-title">{{ c.name }}</td>
          <div class="valentina-card-meta">
            <span class="valentina-badge">{{ c.domain || '—' }}</td>
            <span class="valentina-cell-location">{{ c.location?.municipality || '—' }}</td>
          </div>
          <div class="valentina-card-industries">
            <span v-for="(t, i) in c.industry_tags" :key="i" class="valentina-tag">{{ t }}</td>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const loading = ref(false)
const companies = ref([])

const load = async () => {
  loading.value = true
  try {
    const res = await fetch('/api/v1/valentina/companies')
    const data = await res.json()
    companies.value = data.companies || []
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
.valentina-companies-content { display: flex; flex-direction: column; gap: 16px; }
.valentina-loading { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-empty { display: flex; align-items: center; justify-content: center; padding: 48px; color: var(--valentina-muted); }
.valentina-list { display: flex; flex-direction: column; gap: 12px; }
.valentina-company-card {
  padding: 16px;
  border: 1px solid var(--valentina-border);
  border-radius: 8px;
  background: var(--valentina-surface);
}
.valentina-card-title { font-size: 16px; font-weight: 600; margin: 0 0 8px 0; color: var(--valentina-text); }
.valentina-card-meta { display: flex; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.valentina-card-industries { display: flex; gap: 6px; flex-wrap: wrap; }
.valentina-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--valentina-raised);
  color: var(--valentina-text);
}
.valentina-cell-municipality { font-size: 13px; color: var(--valentina-muted); }
</style>