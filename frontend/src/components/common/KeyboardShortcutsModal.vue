<template>
  <BaseModal :title="$t('shortcuts.title')" @close="$emit('close')" style="--modal-width: 680px">
    <div class="shortcuts-body">
      <div v-for="section in sections" :key="section.key" class="shortcut-section">
        <h4 class="shortcut-section-title">{{ $t(`shortcuts.section_${section.key}`) }}</h4>
        <table class="shortcut-table">
          <tbody>
            <tr v-for="row in section.rows" :key="row.action">
              <td class="shortcut-keys">
                <span class="shortcut-chord">
                  <template v-for="(k, ki) in row.chord" :key="ki">
                    <span v-if="ki > 0" class="shortcut-sep">+</span>
                    <kbd>{{ k }}</kbd>
                  </template>
                </span>
                <template v-if="row.alt">
                  <span class="shortcut-or">{{ $t('common.or') }}</span>
                  <span class="shortcut-chord">
                    <template v-for="(k, ki) in row.alt" :key="ki">
                      <span v-if="ki > 0" class="shortcut-sep">+</span>
                      <kbd>{{ k }}</kbd>
                    </template>
                  </span>
                </template>
              </td>
              <td class="shortcut-desc">{{ $t(row.action) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </BaseModal>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseModal from '@/components/common/BaseModal.vue'

defineEmits(['close'])
const { t } = useI18n()

const isMac = navigator.platform.toUpperCase().includes('MAC')
const mod = isMac ? '⌘' : 'Ctrl'

const sections = computed(() => [
  {
    key: 'global',
    rows: [
      { chord: [mod, '+'], action: 'shortcuts.zoom_in' },
      { chord: [mod, '−'], action: 'shortcuts.zoom_out' },
      { chord: [mod, '0'], action: 'shortcuts.zoom_reset' },
      { chord: [mod, 'K'], action: 'shortcuts.focus_search' },
      { chord: ['?'], action: 'shortcuts.show_shortcuts' },
      { chord: ['Alt', 'A'], action: 'shortcuts.show_a11y' },
      { chord: ['F5'], action: 'shortcuts.reload_desktop' },
    ],
  },
  {
    key: 'navigation',
    rows: [
      { chord: ['Tab'], action: 'shortcuts.skip_to_main' },
      { chord: ['↑ / ↓'], action: 'shortcuts.navigate_results' },
      { chord: ['Enter'], action: 'shortcuts.select_result' },
      { chord: ['Esc'], action: 'shortcuts.close' },
    ],
  },
  {
    key: 'modals',
    rows: [
      { chord: ['Esc'], action: 'shortcuts.close' },
      { chord: ['Tab'], action: 'shortcuts.next_focus' },
      { chord: ['Shift', 'Tab'], action: 'shortcuts.prev_focus' },
    ],
  },
  {
    key: 'board',
    rows: [
      { chord: ['Esc'], action: 'shortcuts.close_card' },
    ],
  },
  {
    key: 'chat',
    rows: [
      { chord: ['Enter'], action: 'shortcuts.send_message' },
      { chord: ['Esc'], action: 'shortcuts.cancel_edit' },
      { chord: ['@'], action: 'shortcuts.mention' },
      { chord: [':name:'], action: 'shortcuts.emoji_shortcode' },
      { chord: ['↑ / ↓'], action: 'shortcuts.navigate_results' },
      { chord: ['Enter / Tab'], action: 'shortcuts.pick_suggestion' },
    ],
  },
])
</script>

<style scoped>
.shortcuts-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px 32px;
  padding: 4px 0 8px;
}
@media (max-width: 520px) {
  .shortcuts-body { grid-template-columns: 1fr; }
}

.shortcut-section-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .06em;
  color: var(--color-text-muted);
  margin-bottom: 8px;
}

.shortcut-table {
  width: 100%;
  border-collapse: collapse;
}
.shortcut-table tr + tr td {
  padding-top: 6px;
}
.shortcut-keys {
  white-space: nowrap;
  padding-right: 12px;
  width: 1%;
}
.shortcut-chord {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
.shortcut-sep {
  font-size: 10px;
  color: var(--color-text-muted);
  margin: 0 1px;
}
.shortcut-or {
  font-size: 11px;
  color: var(--color-text-muted);
  margin: 0 4px;
}
.shortcut-desc {
  font-size: 13px;
  color: var(--color-text);
}
</style>
