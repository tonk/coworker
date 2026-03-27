<template>
  <div class="board-column">
    <div class="column-header">
      <div class="column-header-left">
        <div class="column-dot" :style="{ background: column.color || '#94a3b8' }"></div>
        <input
          v-if="editingName"
          class="column-name-input"
          v-model="editName"
          @blur="saveName"
          @keydown.enter="saveName"
          @keydown.escape="cancelEdit"
          ref="nameInput"
        />
        <span
          v-else
          class="column-name"
          @dblclick="startEdit"
          :title="$t('board.double_click_rename')"
        >{{ column.name }}</span>
        <span class="card-count">{{ column.cards.length }}</span>
        <span v-if="column.wip_limit" :class="['wip-badge', { 'wip-over': column.cards.length >= column.wip_limit }]">
          / {{ column.wip_limit }}
        </span>
      </div>
      <button class="btn btn-ghost btn-sm" @click="$emit('add-card', column.id)" title="Add card">+</button>
    </div>

    <div class="sort-bar">
      <select class="sort-select" v-model="sortField">
        <option value="">{{ $t('board.sort_none') }}</option>
        <option value="due_date">{{ $t('board.sort_date') }}</option>
        <option value="assignee">{{ $t('board.sort_assignee') }}</option>
        <option value="priority">{{ $t('board.sort_priority') }}</option>
      </select>
      <button v-if="sortField" class="sort-dir-btn" @click="sortDir = sortDir === 'asc' ? 'desc' : 'asc'" :title="sortDir === 'asc' ? $t('board.sort_asc') : $t('board.sort_desc')">
        {{ sortDir === 'asc' ? '↑' : '↓' }}
      </button>
    </div>

    <div class="cards-list" ref="listEl">
      <BoardCard
        v-for="card in sortedCards"
        :key="card.id"
        :card="card"
        @open="$emit('open-card', $event)"
        class="sortable-card"
        :data-id="card.id"
      />
    </div>

    <button class="add-card-btn" @click="$emit('add-card', column.id)">
      + {{ $t('board.add_card') }}
    </button>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import Sortable from 'sortablejs'
import BoardCard from './BoardCard.vue'

const props = defineProps({ column: { type: Object, required: true } })
const emit = defineEmits(['add-card', 'open-card', 'card-moved', 'rename-column'])

const listEl = ref(null)
const nameInput = ref(null)
const editingName = ref(false)
const editName = ref('')
let sortable = null

const sortField = ref('')
const sortDir = ref('asc')

const PRIORITY_ORDER = { none: 0, low: 1, medium: 2, high: 3, critical: 4 }

const sortedCards = computed(() => {
  if (!sortField.value) return props.column.cards
  return [...props.column.cards].sort((a, b) => {
    let av, bv
    if (sortField.value === 'due_date') {
      av = a.due_date ? new Date(a.due_date).getTime() : Infinity
      bv = b.due_date ? new Date(b.due_date).getTime() : Infinity
    } else if (sortField.value === 'assignee') {
      av = (a.assignee?.display_name || a.assignee?.username || '').toLowerCase()
      bv = (b.assignee?.display_name || b.assignee?.username || '').toLowerCase()
      if (av === '' && bv !== '') return 1
      if (bv === '' && av !== '') return -1
    } else if (sortField.value === 'priority') {
      av = PRIORITY_ORDER[a.priority] ?? 0
      bv = PRIORITY_ORDER[b.priority] ?? 0
    }
    if (av < bv) return sortDir.value === 'asc' ? -1 : 1
    if (av > bv) return sortDir.value === 'asc' ? 1 : -1
    return 0
  })
})

function startEdit() {
  editName.value = props.column.name
  editingName.value = true
  nextTick(() => nameInput.value?.select())
}

function cancelEdit() {
  editingName.value = false
}

function saveName() {
  const name = editName.value.trim()
  editingName.value = false
  if (name && name !== props.column.name) {
    emit('rename-column', { columnId: props.column.id, name })
  }
}

onMounted(() => {
  sortable = Sortable.create(listEl.value, {
    group: 'cards',
    animation: 150,
    ghostClass: 'card-ghost',
    dragClass: 'card-drag',
    dataIdAttr: 'data-id',
    onEnd(evt) {
      emit('card-moved', {
        cardId: parseInt(evt.item.dataset.id),
        fromColumnId: parseInt(evt.from.closest('.board-column').dataset.columnId || props.column.id),
        toColumnId: parseInt(evt.to.closest('.board-column').dataset.columnId || props.column.id),
        newIndex: evt.newIndex,
        oldIndex: evt.oldIndex
      })
    }
  })
})

onBeforeUnmount(() => sortable?.destroy())
</script>

<style scoped>
.board-column {
  flex-shrink: 0;
  width: 280px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  display: flex;
  flex-direction: column;
  max-height: 100%;
}

.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 12px 8px;
  border-bottom: 1px solid var(--color-border);
}
.column-header-left { display: flex; align-items: center; gap: 6px; }

.column-dot { width: 10px; height: 10px; border-radius: 50%; }
.column-name { font-weight: 600; font-size: 13px; cursor: default; }
.column-name-input {
  font-weight: 600;
  font-size: 13px;
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-sm);
  padding: 2px 6px;
  background: var(--color-surface);
  color: var(--color-text);
  outline: none;
  width: 120px;
}
.card-count {
  background: var(--color-border);
  color: var(--color-text-muted);
  border-radius: 9999px;
  padding: 1px 7px;
  font-size: 11px;
  font-weight: 600;
}

.wip-badge { font-size: 11px; color: var(--color-text-muted); }
.wip-over { color: var(--color-danger); font-weight: 700; }

.sort-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--color-border);
}

.sort-select {
  flex: 1;
  font-size: 11px;
  padding: 3px 6px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
  cursor: pointer;
}

.sort-dir-btn {
  flex-shrink: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 3px 7px;
  font-size: 13px;
  cursor: pointer;
  color: var(--color-text);
  line-height: 1;
}
.sort-dir-btn:hover { background: var(--color-bg); }

.cards-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 40px;
}

.add-card-btn {
  display: block;
  width: 100%;
  padding: 10px;
  background: transparent;
  border: none;
  color: var(--color-text-muted);
  font-size: 13px;
  cursor: pointer;
  border-top: 1px solid var(--color-border);
  text-align: left;
  transition: background .15s;
}
.add-card-btn:hover { background: var(--color-border); color: var(--color-text); }

:global(.card-ghost) { opacity: 0.4; background: var(--color-primary) !important; }
:global(.card-drag) { transform: rotate(1deg); box-shadow: var(--shadow-md); }
</style>
