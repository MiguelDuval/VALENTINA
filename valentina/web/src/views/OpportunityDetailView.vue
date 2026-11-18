<template>
  <div class="valentina-page">
    <div class="valentina-page-header">
      <RouterLink class="valentina-back-link" to="/valentina/opportunities">← Все возможности</RouterLink>
      <h1 class="valentina-page-title">{{ opportunity?.title || 'Карточка возможности' }}</h1>
      <p class="valentina-page-subtitle">Проверка первичного источника и решение оператора</p>
    </div>
    <div v-if="loading" class="valentina-state" role="status">Загрузка карточки...</div>
    <div v-else-if="error" class="valentina-error" role="alert">
      <strong>{{ notFound ? 'Возможность не найдена' : 'Не удалось загрузить карточку' }}</strong>
      <p>{{ error }}</p>
      <button class="valentina-button" type="button" @click="load">Повторить</button>
    </div>
    <template v-else-if="opportunity">
      <section class="valentina-panel">
        <div class="valentina-status-row">
          <span class="valentina-badge">{{ statusLabel(opportunity.status) }}</span>
          <span class="valentina-badge valentina-badge-muted">{{ demandLabel(opportunity.demand_evidence_type) }}</span>
          <span class="valentina-muted">ID: {{ opportunity.opportunity_id }}</span>
        </div>
        <p class="valentina-summary">{{ opportunity.summary || 'Описание не сохранено.' }}</p>
        <p v-if="opportunity.demand_evidence_note" class="valentina-note">{{ opportunity.demand_evidence_note }}</p>
        <p v-if="opportunity.review_note" class="valentina-note">Заметка проверки: {{ opportunity.review_note }}</p>
        <div class="valentina-facts-grid">
          <div><span>Тип</span><strong>{{ opportunity.opportunity_type || '—' }}</strong></div>
          <div><span>Муниципалитет</span><strong>{{ opportunity.geography?.municipality || '—' }}</strong></div>
          <div><span>Провинция</span><strong>{{ opportunity.geography?.province || '—' }}</strong></div>
          <div><span>Страна</span><strong>{{ opportunity.geography?.country_code || '—' }}</strong></div>
          <div><span>Опубликовано</span><strong>{{ formatDate(opportunity.published_at) }}</strong></div>
          <div><span>Срок подачи</span><strong>{{ formatDate(opportunity.deadline_at) }}</strong></div>
          <div><span>Действует с</span><strong>{{ formatDate(opportunity.valid_from) }}</strong></div>
          <div><span>Действует до</span><strong>{{ formatDate(opportunity.valid_until) }}</strong></div>
          <div><span>Стоимость / бюджет</span><strong>{{ formatValue(opportunity.value) }}</strong></div>
          <div><span>Последнее наблюдение</span><strong>{{ formatDate(opportunity.last_seen_at) }}</strong></div>
        </div>
      </section>

      <section class="valentina-panel">
        <div class="valentina-section-heading">
          <h2>Доказательства и первоисточники</h2>
          <span class="valentina-muted">{{ evidence.length }} записей</span>
        </div>
        <div v-if="evidence.length === 0" class="valentina-empty">
          Для этой возможности пока нет сохранённых записей доказательств. Не принимайте её как подтверждённую без первоисточника.
        </div>
        <article v-for="item in evidence" :key="item.evidence_id" class="valentina-evidence">
          <div class="valentina-evidence-heading">
            <div>
              <h3>{{ item.page_title || item.publisher || 'Источник без заголовка' }}</h3>
              <p>{{ item.publisher || 'Издатель не указан' }}</p>
            </div>
            <span class="valentina-badge valentina-badge-muted">{{ item.retrieval_status || 'статус неизвестен' }}</span>
          </div>
          <a v-if="safeURL(item.canonical_url)" :href="safeURL(item.canonical_url)" target="_blank" rel="noopener noreferrer" class="valentina-source-link">Открыть первоисточник ↗</a>
          <p class="valentina-source-url">{{ item.canonical_url }}</p>
          <div class="valentina-evidence-meta">
            <span>Наблюдение: {{ formatDate(item.observed_at) }}</span>
            <span>Публикация: {{ formatDate(item.published_at) }}</span>
            <span>Извлекатель: {{ item.extractor_id || '—' }} {{ item.extractor_version || '' }}</span>
          </div>
          <div v-if="item.facts?.length" class="valentina-evidence-facts">
            <div v-for="(fact, index) in item.facts" :key="index" class="valentina-fact">
              <strong>{{ fact.field_path || 'Факт' }}</strong>
              <span>{{ displayFact(fact.value) }}</span>
              <blockquote v-if="fact.evidence_excerpt">{{ fact.evidence_excerpt }}</blockquote>
              <a v-if="fact.evidence_locator && safeURL(fact.evidence_locator)" :href="safeURL(fact.evidence_locator)" target="_blank" rel="noopener noreferrer">Локатор фрагмента ↗</a>
              <small>{{ fact.extraction_method || 'Метод не указан' }}<template v-if="fact.confidence != null"> · уверенность {{ Math.round(fact.confidence * 100) }}%</template></small>
            </div>
          </div>
        </article>
      </section>

      <section v-if="opportunity.fit_assessment" class="valentina-panel">
        <div class="valentina-section-heading"><h2>Соответствие профилю</h2><span class="valentina-badge valentina-badge-muted">{{ opportunity.fit_assessment.fit_status || 'не оценено' }}</span></div>
        <p class="valentina-muted">Версия профиля: {{ opportunity.fit_assessment.profile_version || '—' }} · {{ formatDate(opportunity.fit_assessment.evaluated_at) }}</p>
        <ul v-if="opportunity.fit_assessment.criteria?.length" class="valentina-criteria">
          <li v-for="criterion in opportunity.fit_assessment.criteria" :key="criterion.criterion_id">
            <strong>{{ criterion.criterion_id }} — {{ criterion.outcome }}</strong>
            <span>{{ criterion.reason }}</span>
          </li>
        </ul>
        <div v-if="opportunity.fit_assessment.missing_information?.length">
          <h3>Недостающая информация</h3>
          <ul><li v-for="item in opportunity.fit_assessment.missing_information" :key="item">{{ item }}</li></ul>
        </div>
      </section>

      <section class="valentina-panel">
        <div class="valentina-section-heading"><h2>История решений</h2><span class="valentina-muted">{{ decisions.length }} записей</span></div>
        <div v-if="decisions.length === 0" class="valentina-empty">Решений по этой возможности ещё нет.</div>
        <ol v-else class="valentina-history">
          <li v-for="decision in decisions" :key="decision.decision_id">
            <div class="valentina-history-heading">
              <strong>{{ statusLabel(decision.previous_status) }} → {{ statusLabel(decision.new_status) }}</strong>
              <time>{{ formatDate(decision.decision_time) }}</time>
            </div>
            <p>{{ decision.decision_reason }}</p>
            <small v-if="decision.profile_version">Профиль: {{ decision.profile_version }}</small>
          </li>
        </ol>
      </section>

      <section class="valentina-panel">
        <h2>Решение оператора</h2>
        <p v-if="actionError" class="valentina-error" role="alert">{{ actionError }}</p>
        <p v-if="allowedStatuses.length === 0" class="valentina-muted">Для текущего статуса нет доступных переходов.</p>
        <form v-else class="valentina-review-form" @submit.prevent="submitDecision">
          <label for="next-status">Новый статус</label>
          <select id="next-status" v-model="nextStatus" required>
            <option value="" disabled>Выберите решение</option>
            <option v-for="status in allowedStatuses" :key="status" :value="status">{{ statusLabel(status) }}</option>
          </select>
          <label for="decision-reason">Обоснование решения</label>
          <textarea id="decision-reason" v-model="reason" rows="3" required maxlength="2000" placeholder="Укажите, что именно подтверждено или почему возможность отклоняется." />
          <button class="valentina-button" type="submit" :disabled="saving || !nextStatus || !reason.trim()">{{ saving ? 'Сохранение...' : 'Сохранить решение' }}</button>
        </form>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { apiFetch } from '../composables/useApi'

const route = useRoute()
const opportunity = ref(null)
const evidence = ref([])
const decisions = ref([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const notFound = ref(false)
const actionError = ref('')
const nextStatus = ref('')
const reason = ref('')

const id = computed(() => String(route.params.id || ''))
const transitions = {
  candidate: ['verified', 'needs_review', 'rejected', 'archived'],
  verified: ['needs_review', 'qualified', 'rejected', 'expired', 'archived'],
  needs_review: ['qualified', 'rejected', 'archived'],
  qualified: ['expired', 'archived'],
}
const allowedStatuses = computed(() => transitions[opportunity.value?.status] || ['archived'])
const statusLabels = {
  candidate: 'Кандидат', verified: 'Проверено', needs_review: 'Требует проверки',
  qualified: 'Квалифицировано', rejected: 'Отклонено', expired: 'Истекло', archived: 'В архиве',
}
const demandLabels = {
  direct_request: 'Прямой запрос', public_tender: 'Публичный тендер',
  subcontracting_request: 'Субподряд', early_demand_signal: 'Ранний сигнал', unknown: 'Неизвестно',
}
const statusLabel = value => statusLabels[value] || value || '—'
const demandLabel = value => demandLabels[value] || value || '—'
const formatDate = value => {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }).format(date)
}
const formatValue = value => {
  if (!value || value.amount == null) return '—'
  return [new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 }).format(value.amount), value.currency || '', value.basis || ''].filter(Boolean).join(' ')
}
const displayFact = value => value == null ? '—' : typeof value === 'object' ? JSON.stringify(value) : String(value)
const safeURL = value => {
  try {
    const url = new URL(value)
    return ['https:', 'http:'].includes(url.protocol) ? url.href : ''
  } catch {
    return ''
  }
}

const load = async () => {
  if (!id.value) return
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    const item = await apiFetch('/api/v1/valentina/opportunities/' + encodeURIComponent(id.value))
    opportunity.value = item
    const evidenceIDs = Array.isArray(item.source_evidence_ids) ? item.source_evidence_ids.filter(Boolean).slice(0, 100) : []
    const params = new URLSearchParams()
    evidenceIDs.forEach(evidenceID => params.append('id', evidenceID))
    const evidenceRequest = evidenceIDs.length
      ? apiFetch('/api/v1/valentina/source-evidence?' + params.toString())
      : Promise.resolve({ evidence: [] })
    const reviewRequest = apiFetch('/api/v1/valentina/review-decisions?entity_type=opportunity&entity_id=' + encodeURIComponent(id.value))
    const [evidenceResponse, reviewResponse] = await Promise.all([evidenceRequest, reviewRequest])
    evidence.value = evidenceResponse.evidence || []
    decisions.value = reviewResponse.decisions || []
  } catch (err) {
    opportunity.value = null
    evidence.value = []
    decisions.value = []
    error.value = err?.message || 'Неизвестная ошибка'
    notFound.value = /404|not found|resource not found/i.test(error.value)
  } finally {
    loading.value = false
  }
}

const submitDecision = async () => {
  if (!nextStatus.value || !reason.value.trim()) {
    actionError.value = 'Выберите статус и укажите обоснование.'
    return
  }
  saving.value = true
  actionError.value = ''
  try {
    await apiFetch('/api/v1/valentina/review-decisions', {
      method: 'POST',
      body: JSON.stringify({
        entity_type: 'opportunity',
        entity_id: id.value,
        new_status: nextStatus.value,
        decision_reason: reason.value.trim(),
      }),
    })
    nextStatus.value = ''
    reason.value = ''
    await load()
  } catch (err) {
    actionError.value = err?.message || 'Не удалось сохранить решение'
  } finally {
    saving.value = false
  }
}

watch(id, load, { immediate: true })
</script>

<style scoped>
.valentina-page { max-width: 1040px; margin: 0 auto; padding: 24px; color: var(--valentina-text); }
.valentina-page-header { margin-bottom: 20px; }
.valentina-back-link { display: inline-block; margin-bottom: 14px; color: var(--valentina-blue); text-decoration: none; }
.valentina-back-link:hover { text-decoration: underline; }
.valentina-page-title { margin: 0 0 6px; font-size: 24px; font-weight: 650; overflow-wrap: anywhere; }
.valentina-page-subtitle,.valentina-muted { color: var(--valentina-muted); font-size: 13px; }
.valentina-panel { margin-bottom: 16px; padding: 18px; border: 1px solid var(--valentina-border); border-radius: 10px; background: var(--valentina-surface); }
.valentina-panel h2 { margin: 0; font-size: 17px; }
.valentina-status-row,.valentina-section-heading,.valentina-evidence-heading,.valentina-history-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
.valentina-status-row { justify-content: flex-start; margin-bottom: 12px; }
.valentina-badge { padding: 4px 9px; border-radius: 5px; background: var(--valentina-green); color: #071b14; font-size: 11px; font-weight: 650; }
.valentina-badge-muted { background: rgba(130,145,160,.18); color: var(--valentina-text); }
.valentina-summary { margin: 12px 0; line-height: 1.6; white-space: pre-wrap; }
.valentina-note { color: var(--valentina-muted); font-size: 13px; }
.valentina-facts-grid { display: grid; grid-template-columns: repeat(auto-fit,minmax(180px,1fr)); gap: 14px; margin-top: 18px; }
.valentina-facts-grid div { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.valentina-facts-grid span,.valentina-evidence-meta,.valentina-fact small { color: var(--valentina-muted); font-size: 12px; }
.valentina-facts-grid strong { overflow-wrap: anywhere; font-size: 13px; font-weight: 550; }
.valentina-section-heading { margin-bottom: 14px; }
.valentina-state,.valentina-empty { padding: 24px; color: var(--valentina-muted); text-align: center; line-height: 1.5; }
.valentina-error { padding: 14px; border-radius: 8px; color: var(--valentina-red); background: rgba(255,126,136,.1); }
.valentina-error p { margin: 6px 0 0; }
.valentina-evidence { padding: 16px 0; border-top: 1px solid var(--valentina-border); }
.valentina-evidence:first-of-type { border-top: 0; }
.valentina-evidence h3 { margin: 0 0 4px; font-size: 15px; }
.valentina-evidence-heading p { margin: 0; color: var(--valentina-muted); font-size: 12px; }
.valentina-source-link { display: inline-block; margin-top: 10px; color: var(--valentina-blue); overflow-wrap: anywhere; }
.valentina-source-url { color: var(--valentina-muted); font-size: 11px; overflow-wrap: anywhere; }
.valentina-evidence-meta { display: flex; gap: 12px; flex-wrap: wrap; }
.valentina-evidence-facts { display: grid; grid-template-columns: repeat(auto-fit,minmax(220px,1fr)); gap: 10px; margin-top: 12px; }
.valentina-fact { display: flex; flex-direction: column; gap: 5px; padding: 10px; border-radius: 7px; background: rgba(130,145,160,.08); font-size: 12px; overflow-wrap: anywhere; }
.valentina-fact blockquote { margin: 0; padding-left: 9px; border-left: 2px solid var(--valentina-border); color: var(--valentina-muted); }
.valentina-fact a { color: var(--valentina-blue); }
.valentina-criteria { padding-left: 20px; }
.valentina-criteria li { margin: 8px 0; }
.valentina-criteria li strong,.valentina-criteria li span { display: block; }
.valentina-criteria li span { color: var(--valentina-muted); font-size: 13px; }
.valentina-history { margin: 0; padding-left: 20px; }
.valentina-history li { margin: 12px 0; padding-left: 4px; }
.valentina-history-heading time,.valentina-history small { color: var(--valentina-muted); font-size: 12px; }
.valentina-history p { margin: 5px 0; }
.valentina-review-form { display: grid; gap: 9px; max-width: 640px; margin-top: 14px; }
.valentina-review-form label { font-size: 13px; color: var(--valentina-muted); }
.valentina-review-form select,.valentina-review-form textarea { width: 100%; padding: 10px; border: 1px solid var(--valentina-border); border-radius: 7px; color: var(--valentina-text); background: var(--valentina-bg); font: inherit; }
.valentina-button { width: fit-content; padding: 9px 14px; border: 0; border-radius: 7px; color: #061b14; background: var(--valentina-green); font-weight: 650; cursor: pointer; }
.valentina-button:disabled { opacity: .5; cursor: not-allowed; }
@media(max-width:640px){.valentina-page{padding:16px}.valentina-panel{padding:14px}}
</style>
