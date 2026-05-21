<template>
  <Teleport to="body">
    <div class="modal-backdrop" :class="{ 'modal-backdrop-maximized': maximized }" @click.self="$emit('close')">
      <div
        ref="modalEl"
        class="modal"
        :class="{ 'modal-resizable': resizable, 'modal-maximized': maximized }"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
      >
        <div class="modal-header">
          <h3 :id="titleId">{{ title }}</h3>
          <div class="modal-header-actions">
            <button v-if="resizable" class="btn btn-ghost btn-sm" :aria-label="maximized ? $t('common.restore') : $t('common.maximize')" @click="toggleMaximize">
              <svg v-if="!maximized" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polyline points="4 14 10 14 10 20"/><polyline points="20 10 14 10 14 4"/><line x1="10" y1="14" x2="3" y2="21"/><line x1="21" y1="3" x2="14" y2="10"/></svg>
            </button>
            <button class="btn btn-ghost btn-sm" :aria-label="$t('common.close')" @click="$emit('close')">✕</button>
          </div>
        </div>
        <div class="modal-body">
          <slot />
        </div>
        <div class="modal-footer" v-if="$slots.footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const props = defineProps({ title: String, resizable: { type: Boolean, default: false } })
const emit = defineEmits(['close'])

const modalEl = ref(null)
const maximized = ref(false)
const titleId = 'modal-title-' + Math.random().toString(36).slice(2, 8)

const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

let previousFocus = null

function toggleMaximize() {
  maximized.value = !maximized.value
}

function onKeyDown(e) {
  if (e.key === 'Escape') { emit('close'); return }
  if (e.key !== 'Tab' || !modalEl.value) return

  const focusable = [...modalEl.value.querySelectorAll(FOCUSABLE)]
  if (!focusable.length) return

  const first = focusable[0]
  const last = focusable[focusable.length - 1]

  if (e.shiftKey) {
    if (document.activeElement === first) { e.preventDefault(); last.focus() }
  } else {
    if (document.activeElement === last) { e.preventDefault(); first.focus() }
  }
}

onMounted(() => {
  previousFocus = document.activeElement
  document.addEventListener('keydown', onKeyDown)
  const first = modalEl.value?.querySelector(FOCUSABLE)
  if (first) first.focus()
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeyDown)
  previousFocus?.focus()
})
</script>
