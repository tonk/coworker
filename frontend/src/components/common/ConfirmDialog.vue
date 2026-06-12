<template>
  <Teleport to="body">
    <Transition name="confirm-fade">
      <div v-if="ui.confirmState" class="confirm-backdrop" @click.self="cancel" role="dialog" aria-modal="true" :aria-labelledby="titleId">
        <div ref="dialogEl" class="confirm-dialog">
          <h3 :id="titleId" class="confirm-title">{{ $t('common.confirm') }}</h3>
          <p class="confirm-message">{{ ui.confirmState.message }}</p>
          <div class="confirm-actions">
            <button class="btn btn-secondary btn-sm" @click="cancel">{{ $t('common.cancel') }}</button>
            <button
              class="btn btn-sm"
              :class="ui.confirmState.destructive ? 'btn-danger' : 'btn-primary'"
              @click="ok"
              ref="okBtn"
            >{{ confirmLabel }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, nextTick, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUIStore } from '@/stores/ui'

const ui = useUIStore()
const { t } = useI18n()
const dialogEl = ref(null)
const okBtn = ref(null)
const titleId = 'confirm-dialog-title'

const confirmLabel = computed(() => {
  const state = ui.confirmState
  if (!state) return t('common.yes')
  if (state.confirmLabel) return state.confirmLabel
  return state.destructive ? t('common.delete') : t('common.yes')
})

let previousFocus = null

watch(() => ui.confirmState, async (state) => {
  if (state) {
    previousFocus = document.activeElement
    await nextTick()
    okBtn.value?.focus()
    document.addEventListener('keydown', onKey)
  } else {
    document.removeEventListener('keydown', onKey)
    previousFocus?.focus()
  }
})

function onKey(e) {
  if (e.key === 'Escape') { cancel(); return }
  if (e.key !== 'Tab' || !dialogEl.value) return
  const focusable = [...dialogEl.value.querySelectorAll('button:not([disabled])')]
  if (!focusable.length) return
  const first = focusable[0], last = focusable[focusable.length - 1]
  if (e.shiftKey) { if (document.activeElement === first) { e.preventDefault(); last.focus() } }
  else { if (document.activeElement === last) { e.preventDefault(); first.focus() } }
}

function ok() { ui._confirmResolve(true) }
function cancel() { ui._confirmResolve(false) }
</script>

<style scoped>
.confirm-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.confirm-dialog {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 24px;
  max-width: 400px;
  width: 90%;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
}

.confirm-title {
  font-size: 16px;
  font-weight: 700;
  margin: 0 0 12px;
  color: var(--color-text);
}

.confirm-message {
  font-size: 14px;
  color: var(--color-text-muted);
  margin: 0 0 20px;
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.confirm-fade-enter-active,
.confirm-fade-leave-active { transition: opacity 0.15s; }
.confirm-fade-enter-from,
.confirm-fade-leave-to { opacity: 0; }
</style>
