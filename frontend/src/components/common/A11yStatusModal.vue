<template>
  <div class="modal-overlay" role="dialog" aria-modal="true" :aria-label="$t('dashboard.a11y_title')" @click.self="$emit('close')" @keydown.esc="$emit('close')">
    <div class="modal-box" tabindex="-1" ref="boxRef">
      <button class="modal-close" :aria-label="$t('common.close')" @click="$emit('close')">×</button>
      <div class="a11y-tag">{{ $t('dashboard.a11y_title') }}</div>
      <div class="a11y-badge">WCAG 2.1 AA</div>
      <ul class="a11y-list">
        <li>✓ Skip-to-content &amp; focus management</li>
        <li>✓ ARIA roles, labels &amp; live regions</li>
        <li>✓ Keyboard shortcuts &amp; focus trap</li>
        <li>✓ Heading hierarchy (h1 on every page)</li>
        <li>✓ Form error announcements</li>
        <li>✓ Alt text on all images</li>
        <li class="a11y-open">◌ Color contrast (muted text pending)</li>
      </ul>
      <button class="link-btn" @click="openShortcuts">{{ $t('dashboard.view_shortcuts') }} →</button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'

defineEmits(['close'])

const boxRef = ref(null)

onMounted(() => {
  boxRef.value?.focus()
})

function openShortcuts() {
  window.dispatchEvent(new CustomEvent('open-keyboard-shortcuts'))
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-box {
  position: relative;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 2rem;
  min-width: 280px;
  max-width: 380px;
  outline: none;
}

.modal-close {
  position: absolute;
  top: 0.75rem;
  right: 0.75rem;
  background: none;
  border: none;
  font-size: 1.25rem;
  color: var(--color-text-muted);
  cursor: pointer;
  line-height: 1;
  padding: 0.25rem 0.5rem;
}
.modal-close:hover { color: var(--color-text); }

.a11y-tag {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-success, #22c55e);
  margin-bottom: 0.75rem;
}

.a11y-badge {
  display: inline-block;
  background: var(--color-success, #22c55e);
  color: #fff;
  font-weight: 800;
  font-size: 1rem;
  border-radius: 6px;
  padding: 0.2rem 0.7rem;
  margin-bottom: 1rem;
}

.a11y-list {
  list-style: none;
  padding: 0;
  margin: 0 0 1.25rem;
  font-size: 0.875rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.a11y-list li { color: var(--color-text); }
.a11y-open { opacity: 0.65; }

.link-btn {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 0.875rem;
  cursor: pointer;
  padding: 0;
}
.link-btn:hover { text-decoration: underline; }
</style>
