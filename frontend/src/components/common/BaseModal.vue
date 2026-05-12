<template>
  <Teleport to="body">
    <div class="modal-backdrop" @click.self="$emit('close')">
      <div
        ref="modalEl"
        class="modal"
        :class="{ 'modal-resizable': resizable }"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
      >
        <div class="modal-header">
          <h3 :id="titleId">{{ title }}</h3>
          <button class="btn btn-ghost btn-sm" :aria-label="$t('common.close')" @click="$emit('close')">✕</button>
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
const titleId = 'modal-title-' + Math.random().toString(36).slice(2, 8)

const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

let previousFocus = null

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
