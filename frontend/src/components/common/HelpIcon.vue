<template>
  <span class="help-icon-wrap" ref="wrapRef">
    <button
      type="button"
      class="help-icon-btn"
      :aria-label="buttonLabel"
      :aria-expanded="open"
      aria-haspopup="dialog"
      @click.stop="toggle"
    >
      <svg aria-hidden="true" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/>
        <line x1="12" y1="17" x2="12.01" y2="17"/>
      </svg>
    </button>
    <div
      v-if="open"
      :id="popoverId"
      :class="['help-icon-popover', align !== 'center' && `help-icon-popover--${align}`]"
      role="dialog"
      :aria-label="buttonLabel"
      @keydown.esc.prevent="close"
    >
      {{ $t(i18nKey) }}
    </div>
  </span>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  /** Full vue-i18n key, e.g. admin.allowed_ips_hint */
  i18nKey: { type: String, required: true },
  /** Optional override for the icon aria-label */
  label: { type: String, default: '' },
  /** Popover alignment relative to the icon: 'center' (default), 'start' (extends right), 'end' (extends left) */
  align: { type: String, default: 'center' },
})

const { t } = useI18n()
const open = ref(false)
const wrapRef = ref(null)
const popoverId = `help-icon-${Math.random().toString(36).slice(2, 9)}`

const buttonLabel = computed(() => props.label || t('help.field_button'))

function toggle() {
  open.value = !open.value
}

function close() {
  open.value = false
}

function onDocClick(e) {
  if (open.value && wrapRef.value && !wrapRef.value.contains(e.target)) {
    close()
  }
}

onMounted(() => document.addEventListener('click', onDocClick))
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))
</script>

<style scoped>
.help-icon-wrap {
  display: inline-flex;
  vertical-align: middle;
  position: relative;
  margin-left: 4px;
}

.help-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
}

.help-icon-btn:hover,
.help-icon-btn:focus-visible {
  background: var(--color-bg);
  color: var(--color-primary);
}

.help-icon-popover {
  position: absolute;
  z-index: 200;
  top: calc(100% + 6px);
  left: 50%;
  transform: translateX(-50%);
  width: max-content;
  max-width: min(320px, 70vw);
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius, 6px);
  background: var(--color-surface);
  color: var(--color-text);
  font-size: 12px;
  line-height: 1.45;
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0, 0, 0, 0.12));
}

.help-icon-popover::before {
  content: '';
  position: absolute;
  top: -5px;
  left: 50%;
  width: 8px;
  height: 8px;
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
  border-top: 1px solid var(--color-border);
  transform: translateX(-50%) rotate(45deg);
}

/* Popover extends to the left — right edge anchors to the icon */
.help-icon-popover--end {
  left: auto;
  right: 0;
  transform: none;
}
.help-icon-popover--end::before {
  left: auto;
  right: 7px;
  transform: rotate(45deg);
}

/* Popover extends to the right — left edge anchors to the icon */
.help-icon-popover--start {
  left: 0;
  transform: none;
}
.help-icon-popover--start::before {
  left: 7px;
  transform: rotate(45deg);
}
</style>
