<template>
  <Teleport to="body">
    <ul
      ref="menuEl"
      class="ctx-menu"
      role="menu"
      tabindex="-1"
      :style="{ left: pos.left + 'px', top: pos.top + 'px', visibility: ready ? 'visible' : 'hidden' }"
      @keydown="onKeyDown"
    >
      <li v-for="(item, i) in items" :key="item.key" role="none">
        <button
          role="menuitem"
          type="button"
          class="ctx-menu-item"
          :class="{ 'ctx-menu-item-danger': item.danger }"
          :tabindex="i === activeIndex ? 0 : -1"
          :disabled="item.disabled"
          @click="select(item)"
          @mouseenter="activeIndex = i"
        >{{ item.label }}</button>
      </li>
    </ul>
  </Teleport>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'

const props = defineProps({
  x: { type: Number, required: true },
  y: { type: Number, required: true },
  items: { type: Array, required: true }, // [{ key, label, danger?, disabled? }]
})
const emit = defineEmits(['select', 'close'])

const menuEl = ref(null)
const ready = ref(false)
const pos = reactive({ left: props.x, top: props.y })
const activeIndex = ref(props.items.findIndex(i => !i.disabled))

function clampPosition() {
  const el = menuEl.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const margin = 8
  let left = props.x
  let top = props.y
  if (left + rect.width > window.innerWidth - margin) left = window.innerWidth - rect.width - margin
  if (top + rect.height > window.innerHeight - margin) top = window.innerHeight - rect.height - margin
  pos.left = Math.max(margin, left)
  pos.top = Math.max(margin, top)
}

function select(item) {
  if (item.disabled) return
  emit('select', item.key)
}

function focusActive() {
  const buttons = menuEl.value?.querySelectorAll('button[role="menuitem"]')
  buttons?.[activeIndex.value]?.focus()
}

function moveActive(delta) {
  const count = props.items.length
  if (!count) return
  let idx = activeIndex.value
  for (let step = 0; step < count; step++) {
    idx = (idx + delta + count) % count
    if (!props.items[idx].disabled) break
  }
  activeIndex.value = idx
  focusActive()
}

function onKeyDown(e) {
  if (e.key === 'Escape') { emit('close'); return }
  if (e.key === 'ArrowDown') { e.preventDefault(); moveActive(1); return }
  if (e.key === 'ArrowUp') { e.preventDefault(); moveActive(-1); return }
  if (e.key === 'Home') { e.preventDefault(); activeIndex.value = -1; moveActive(1); return }
  if (e.key === 'End') { e.preventDefault(); activeIndex.value = 0; moveActive(-1); return }
}

function onDocMouseDown(e) {
  if (menuEl.value && !menuEl.value.contains(e.target)) emit('close')
}

onMounted(async () => {
  await nextTick()
  clampPosition()
  ready.value = true
  focusActive()
  document.addEventListener('mousedown', onDocMouseDown)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDocMouseDown)
})
</script>

<style scoped>
.ctx-menu {
  position: fixed;
  z-index: 2000;
  min-width: 160px;
  margin: 0;
  padding: 4px;
  list-style: none;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-md);
}

.ctx-menu-item {
  display: block;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  border: none;
  background: none;
  color: var(--color-text);
  font-size: 13px;
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.ctx-menu-item:hover,
.ctx-menu-item:focus-visible {
  background: var(--color-bg);
  outline: none;
}

.ctx-menu-item:disabled {
  color: var(--color-text-muted);
  cursor: not-allowed;
}

.ctx-menu-item-danger {
  color: var(--color-danger);
}
</style>
