<template>
  <Teleport to="body">
    <div class="modal-backdrop" @click.self="$emit('close')">
      <div class="about-modal" role="dialog" aria-label="About WarmDesk">
        <button class="about-close" @click="$emit('close')" aria-label="Close">✕</button>

        <div class="about-logo">
          <img src="/logo-full.svg" alt="WarmDesk" />
        </div>

        <div class="about-version">
          <span class="about-version-badge">v{{ appVersion }}</span>
          <span v-if="serverVersion" class="about-version-server">server&nbsp;{{ serverVersion }}</span>
          <span v-if="loading" class="about-version-server">…</span>
        </div>

        <p class="about-desc">
	  Self-hosted project management<br>
	  Kanban and Scrum boards,<br>
          team chat, discussions, and time reporting.
        </p>

        <table class="about-table">
          <tbody>
            <tr>
              <td>{{ $t('about.repository') }}</td>
              <td><a href="https://github.com/tonk/warmdesk" target="_blank" rel="noopener">github.com/tonk/warmdesk</a></td>
            </tr>
            <tr>
              <td>{{ $t('about.license') }}</td>
              <td><a href="https://www.gnu.org/licenses/gpl-3.0.html" target="_blank" rel="noopener">GNU General Public License v3</a></td>
            </tr>
            <tr>
              <td>{{ $t('about.built_with') }}</td>
              <td>Go · Vue 3 · SQLite / PostgreSQL / MySQL</td>
            </tr>
          </tbody>
        </table>

        <p class="about-copyright">© {{ year }} Ton Kersten — All rights reserved.</p>

        <div class="about-footer">
          <button class="btn btn-secondary" @click="$emit('close')">{{ $t('common.close') }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import client from '@/api/client'

const emit = defineEmits(['close'])

const appVersion = __APP_VERSION__
const serverVersion = ref('')
const loading = ref(true)
const year = new Date().getFullYear()

function onKeyDown(e) {
  if (e.key === 'Escape') emit('close')
}

onMounted(() => {
  document.addEventListener('keydown', onKeyDown)
  client.get('/version')
    .then(r => { serverVersion.value = r.data.version })
    .catch(() => {})
    .finally(() => { loading.value = false })
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeyDown)
})
</script>

<style scoped>
.about-modal {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: 0 20px 60px rgba(0,0,0,.25);
  width: 380px;
  max-width: calc(100vw - 32px);
  padding: 36px 32px 28px;
  position: relative;
  text-align: center;
}

.about-close {
  position: absolute;
  top: 12px;
  right: 12px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 16px;
  color: var(--color-text-muted);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  line-height: 1;
}
.about-close:hover { background: var(--color-bg); color: var(--color-text); }

.about-logo img {
  height: 32px;
  width: auto;
  display: block;
  margin: 0 auto 16px;
}

.about-version {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 16px;
}
.about-version-badge {
  background: var(--color-primary);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 999px;
}
.about-version-server {
  font-size: 12px;
  color: var(--color-text-muted);
}

.about-desc {
  font-size: 13px;
  color: var(--color-text-muted);
  line-height: 1.6;
  margin: 0 0 24px;
}

.about-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  margin-bottom: 24px;
  text-align: left;
}
.about-table td {
  padding: 6px 0;
  vertical-align: top;
}
.about-table td:first-child {
  color: var(--color-text-muted);
  white-space: nowrap;
  padding-right: 16px;
  width: 90px;
}
.about-table tr + tr td {
  border-top: 1px solid var(--color-border);
}
.about-table a {
  color: var(--color-primary);
  text-decoration: none;
}
.about-table a:hover { text-decoration: underline; }

.about-copyright {
  font-size: 11px;
  color: var(--color-text-muted);
  margin: 0 0 24px;
}

.about-footer {
  display: flex;
  justify-content: center;
}
</style>
