<template>
  <BaseModal :title="$t('help.title')" @close="$emit('close')" :resizable="true" style="--modal-width: 520px">
    <article class="help-panel">
      <h4 class="help-page-title">{{ title }}</h4>
      <p v-if="intro" class="help-intro">{{ intro }}</p>

      <section v-if="tasks.length" class="help-section">
        <h5 class="help-section-title">{{ $t('help.tasks_heading') }}</h5>
        <ul class="help-list">
          <li v-for="(task, idx) in tasks" :key="idx">{{ task }}</li>
        </ul>
      </section>

      <section v-if="shortcuts.length" class="help-section">
        <h5 class="help-section-title">{{ $t('help.shortcuts_heading') }}</h5>
        <ul class="help-list help-list--shortcuts">
          <li v-for="(shortcut, idx) in shortcuts" :key="idx">{{ shortcut }}</li>
        </ul>
      </section>
    </article>
  </BaseModal>
</template>

<script setup>
import BaseModal from '@/components/common/BaseModal.vue'
import { usePageHelp } from '@/composables/usePageHelp'

defineEmits(['close'])

const { title, intro, tasks, shortcuts } = usePageHelp()
</script>

<style scoped>
.help-panel {
  font-size: 14px;
  line-height: 1.55;
  color: var(--color-text);
}

.help-page-title {
  margin: 0 0 10px;
  font-size: 16px;
  font-weight: 600;
}

.help-intro {
  margin: 0 0 16px;
  color: var(--color-text-muted);
}

.help-section + .help-section {
  margin-top: 16px;
}

.help-section-title {
  margin: 0 0 8px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
}

.help-list {
  margin: 0;
  padding-left: 1.25rem;
}

.help-list li + li {
  margin-top: 6px;
}

.help-list--shortcuts {
  list-style: none;
  padding-left: 0;
}

.help-list--shortcuts li {
  padding: 6px 10px;
  background: var(--color-bg);
  border-radius: var(--radius, 6px);
  font-size: 13px;
}

.help-list--shortcuts li + li {
  margin-top: 4px;
}
</style>
