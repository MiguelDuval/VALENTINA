<template>
  <main class="about-page">
    <header><p>VALENTINA · Локальное рабочее пространство</p><h1>О программе</h1></header>
    <section aria-labelledby="backup-title">
      <h2 id="backup-title">Резервная копия CRM</h2>
      <p>Копия содержит частные данные клиентов, заявки, проекты, сметы и историю. Храните файл в защищённом месте и не публикуйте его.</p>
      <p>Это не полная копия приложения: настройки, ключи, агенты и файлы вне базы CRM не включены. Восстановление через интерфейс пока недоступно. Для восстановления нужна отдельная проверка копии при остановленном приложении.</p>
      <label><input v-model="acknowledged" type="checkbox" /> Я понимаю, что файл содержит частные данные</label>
      <button :disabled="!acknowledged || busy" @click="download">{{ busy ? 'Подготовка копии…' : 'Скачать копию CRM' }}</button>
      <p v-if="backupError" role="alert">{{ backupError }}</p><p v-if="saved" role="status">Копия подготовлена. Сохраните её в защищённом месте.</p>
    </section>
    <section aria-labelledby="licenses-title">
      <h2 id="licenses-title">Лицензии</h2>
      <p>VALENTINA использует изменённый интерфейс и локальный runtime scrypster/Huginn. Исходная лицензия — Business Source License 1.1, не OSI open source. Условия использования, Change Date и Change License приведены ниже полностью. Распространение и предоставление сервиса третьим лицам требуют отдельной проверки условий.</p>
      <p>Инвентаризация лицензий frontend и транзитивных зависимостей ещё не завершена. Доступные сторонние уведомления включены в установленный пакет: LICENSES/THIRD-PARTY-NOTICES.txt.</p>
      <p v-if="loading" role="status">Загрузка лицензии…</p>
      <p v-if="licenseError" role="alert">{{ licenseError }} <button @click="load">Повторить</button></p>
      <template v-if="license"><p>Коммит: {{ license.upstream_commit }}</p><pre>{{ license.license }}</pre></template>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiFetch, ensureToken } from '../composables/useApi'
const license = ref<{ license: string; upstream_commit: string } | null>(null)
const loading = ref(false), busy = ref(false), acknowledged = ref(false), saved = ref(false)
const licenseError = ref(''), backupError = ref('')
async function load() {
  loading.value = true; licenseError.value = ''; license.value = null
  try { license.value = await apiFetch('/api/v1/valentina/licenses') }
  catch { licenseError.value = 'Не удалось загрузить лицензию.' }
  finally { loading.value = false }
}
async function download() {
  if (!acknowledged.value || busy.value) return
  busy.value = true; backupError.value = ''; saved.value = false
  const controller = new AbortController(), timeout = setTimeout(() => controller.abort(), 65000)
  try {
    const token = await ensureToken()
    const response = await fetch('/api/v1/valentina/backup', { headers: { Authorization: `Bearer ${token}` }, cache: 'no-store', redirect: 'error', signal: controller.signal })
    if (!response.ok || response.headers.get('Content-Type') !== 'application/vnd.sqlite3') throw new Error('backup failed')
    const data = await response.arrayBuffer()
    if (!data.byteLength || data.byteLength > 128 * 1024 * 1024) throw new Error('backup size')
    const digest = await crypto.subtle.digest('SHA-256', data)
    const hash = Array.from(new Uint8Array(digest), b => b.toString(16).padStart(2, '0')).join('')
    if (hash !== response.headers.get('X-Backup-SHA256')) throw new Error('backup integrity')
    const url = URL.createObjectURL(new Blob([data], { type: 'application/vnd.sqlite3' }))
    const link = document.createElement('a'); link.href = url; link.download = 'VALENTINA-CRM.sqlite'
    document.body.appendChild(link); link.click(); link.remove()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    saved.value = true
  } catch { backupError.value = 'Не удалось создать или проверить копию CRM. Повторите попытку.' }
  finally { clearTimeout(timeout); busy.value = false }
}
onMounted(load)
</script>

<style scoped>
.about-page { max-width: 960px; margin: 0 auto; padding: 32px; }
section { margin-top: 28px; padding: 24px; border: 1px solid var(--border); border-radius: 12px; }
p { line-height: 1.6; margin: 12px 0; }
pre { white-space: pre-wrap; overflow-wrap: anywhere; font-size: 13px; line-height: 1.6; }
button { margin: 12px; padding: 10px 16px; cursor: pointer; }
button:disabled { opacity: .5; cursor: default; }
[role=alert] { color: var(--danger, #b42318); }
</style>
