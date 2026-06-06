<template>
  <span class="help-icon-wrap" ref="wrapRef">
    <button
      type="button"
      class="help-icon-btn"
      ref="btnRef"
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
    <Teleport to="body">
      <div
        v-if="open"
        :id="popoverId"
        class="help-icon-popover"
        :class="{ 'help-icon-popover--above': !placementBelow }"
        :style="popoverStyle"
        role="dialog"
        :aria-label="buttonLabel"
        @click.stop
        @keydown.esc.prevent="close"
      >
        {{ $t(i18nKey) }}
        <span class="help-icon-popover-arrow" :style="arrowStyle"></span>
      </div>
    </Teleport>
  </span>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
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
const btnRef = ref(null)
const popoverId = `help-icon-${Math.random().toString(36).slice(2, 9)}`
const popoverStyle = ref({})
const arrowStyle = ref({})
const placementBelow = ref(true)

const buttonLabel = computed(() => props.label || t('help.field_button'))

const POPOVER_WIDTH = 320
const GAP = 4
const MARGIN = 8

function getPageZoom() {
  const inline = parseFloat(document.documentElement.style.zoom)
  if (Number.isFinite(inline) && inline > 0) return inline
  const computedZoom = parseFloat(getComputedStyle(document.documentElement).zoom)
  if (Number.isFinite(computedZoom) && computedZoom > 0) return computedZoom
  return 1
}

function computePosition() {
  if (!btnRef.value || !open.value) return

  const rect = btnRef.value.getBoundingClientRect()
  const zoom = getPageZoom()
  const vw = window.innerWidth
  const vh = window.innerHeight
  const maxW = Math.min(POPOVER_WIDTH, vw * 0.7)

  let idealLeft
  if (props.align === 'end') {
    idealLeft = rect.right - maxW
  } else if (props.align === 'start') {
    idealLeft = rect.left
  } else {
    idealLeft = rect.left + rect.width / 2 - maxW / 2
  }

  const clampedLeft = Math.max(MARGIN, Math.min(idealLeft, vw - maxW - MARGIN))

  const popEl = document.getElementById(popoverId)
  const popHeight = popEl?.offsetHeight || 72

  const spaceBelow = vh - rect.bottom - GAP
  const spaceAbove = rect.top - GAP
  const showBelow = spaceBelow >= popHeight || spaceBelow >= spaceAbove
  placementBelow.value = showBelow

  let topViewport = showBelow
    ? rect.bottom + GAP
    : rect.top - GAP - popHeight
  topViewport = Math.max(MARGIN, Math.min(topViewport, vh - popHeight - MARGIN))

  const btnCenterX = rect.left + rect.width / 2
  const arrowLeft = Math.max(8, Math.min(btnCenterX - clampedLeft - 4, maxW - 16))

  popoverStyle.value = {
    position: 'fixed',
    top: `${topViewport / zoom}px`,
    left: `${clampedLeft / zoom}px`,
    width: `${maxW / zoom}px`,
    zIndex: 9999,
  }
  arrowStyle.value = { left: `${arrowLeft / zoom}px` }
}

function schedulePosition() {
  nextTick(() => {
    requestAnimationFrame(() => {
      computePosition()
      requestAnimationFrame(computePosition)
    })
  })
}

function toggle() {
  open.value = !open.value
}

function close() {
  open.value = false
}

function onDocClick(e) {
  if (open.value && wrapRef.value && !wrapRef.value.contains(e.target)) {
    const popEl = document.getElementById(popoverId)
    if (popEl && popEl.contains(e.target)) return
    close()
  }
}

function onReposition() {
  if (open.value) computePosition()
}

watch(open, (val) => {
  if (val) schedulePosition()
})

onMounted(() => {
  document.addEventListener('click', onDocClick)
  window.addEventListener('resize', onReposition)
  window.addEventListener('scroll', onReposition, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocClick)
  window.removeEventListener('resize', onReposition)
  window.removeEventListener('scroll', onReposition, true)
})
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
</style>

<style>
/* Global (not scoped) so Teleport'd element inherits the styles */
.help-icon-popover {
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius, 6px);
  background: var(--color-surface);
  color: var(--color-text);
  font-size: 12px;
  line-height: 1.45;
  box-shadow: var(--shadow-md, 0 4px 12px rgba(0, 0, 0, 0.12));
}

.help-icon-popover-arrow {
  position: absolute;
  top: -5px;
  width: 8px;
  height: 8px;
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
  border-top: 1px solid var(--color-border);
  transform: rotate(45deg);
}

.help-icon-popover--above .help-icon-popover-arrow {
  top: auto;
  bottom: -5px;
  transform: rotate(225deg);
}
</style>
