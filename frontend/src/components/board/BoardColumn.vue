<template>
  <div class="board-column" :data-column-id="column.id">
    <div class="column-header">
      <div class="column-header-left">
        <div class="column-dot" :style="{ background: column.color || '#94a3b8' }" aria-hidden="true"></div>
        <input
          v-if="editingName"
          class="column-name-input"
          v-model="editName"
          aria-label="Column name"
          @blur="saveName"
          @keydown.enter="saveName"
          @keydown.escape="cancelEdit"
          ref="nameInput"
        />
        <span
          v-else
          class="column-name"
          @dblclick="canManageColumns ? startEdit() : null"
          :title="canManageColumns ? $t('board.double_click_rename') : ''"
        >{{ column.name }}</span>
        <div
          class="header-stat-tip-host"
          :class="{ 'tip-pinned': countTipPinned }"
          tabindex="0"
          @click.stop="togglePinnedCount"
        >
          <span class="card-count">{{ cardCount }}</span>
          <div class="header-stat-popup" role="tooltip">{{ cardCountTooltip }}</div>
        </div>
        <div
          v-if="column.wip_limit"
          class="header-stat-tip-host wip-tip-host"
          :class="{ 'tip-pinned': wipTipPinned, 'wip-over': cardCount >= column.wip_limit }"
          tabindex="0"
          @click.stop="togglePinnedWip"
        >
          <span class="wip-badge">/ {{ column.wip_limit }}</span>
          <div class="header-stat-popup" role="tooltip">{{ wipBadgeTooltip }}</div>
        </div>
      </div>
      <div class="column-header-actions">
        <button class="btn btn-ghost btn-sm" @click="$emit('add-card', column.id)" :title="$t('board.add_card')" aria-label="Add card">+</button>
        <button
          v-if="canManageColumns"
          type="button"
          class="btn btn-ghost btn-sm col-edit-btn"
          @click.stop="$emit('edit-column', column)"
          :title="$t('board.edit_column')"
          aria-label="Edit column"
        >⚙</button>
        <button
          v-if="canManageColumns && column.cards.length === 0"
          type="button"
          class="btn btn-ghost btn-sm delete-col-btn"
          @click="$emit('delete-column', column.id)"
          :title="$t('board.delete_column')"
          aria-label="Delete column"
        >🗑</button>
      </div>
    </div>

    <div class="sort-bar">
      <select class="sort-select" v-model="sortField" aria-label="Sort cards by">
        <option value="">{{ $t('board.sort_none') }}</option>
        <option value="due_date">{{ $t('board.sort_date') }}</option>
        <option value="assignee">{{ $t('board.sort_assignee') }}</option>
        <option value="priority">{{ $t('board.sort_priority') }}</option>
      </select>
      <button v-if="sortField" class="sort-dir-btn" @click="sortDir = sortDir === 'asc' ? 'desc' : 'asc'" :title="sortDir === 'asc' ? $t('board.sort_asc') : $t('board.sort_desc')" :aria-label="sortDir === 'asc' ? 'Sort descending' : 'Sort ascending'">
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
import { useI18n } from 'vue-i18n'
import Sortable from 'sortablejs'
import BoardCard from './BoardCard.vue'

const { t } = useI18n()

const props = defineProps({
  column: { type: Object, required: true },
  canManageColumns: { type: Boolean, default: false }
})
const emit = defineEmits(['add-card', 'open-card', 'card-moved', 'rename-column', 'delete-column', 'edit-column'])

const listEl = ref(null)
const nameInput = ref(null)
const editingName = ref(false)
const editName = ref('')
let sortable = null

const sortField = ref('')
const sortDir = ref('asc')
const countTipPinned = ref(false)
const wipTipPinned = ref(false)

const PRIORITY_ORDER = { none: 0, low: 1, medium: 2, high: 3, critical: 4 }

const cardCount = computed(() => (props.column.cards || []).length)

const cardCountTooltip = computed(() =>
  t('board.column_cards_tooltip', { count: cardCount.value })
)

const wipBadgeTooltip = computed(() => {
  const limit = props.column.wip_limit
  if (limit == null) return ''
  const count = cardCount.value
  if (count >= limit) {
    return t('board.column_wip_tooltip_over', { count, limit })
  }
  return t('board.column_wip_tooltip', { count, limit })
})

function togglePinnedCount() {
  countTipPinned.value = !countTipPinned.value
  if (countTipPinned.value) wipTipPinned.value = false
}

function togglePinnedWip() {
  wipTipPinned.value = !wipTipPinned.value
  if (wipTipPinned.value) countTipPinned.value = false
}

const sortedCards = computed(() => {
  const cards = props.column.cards || []
  if (!sortField.value) return cards
  return [...cards].sort((a, b) => {
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
.column-header-left { display: flex; align-items: center; gap: 6px; flex: 1; min-width: 0; }
.column-header-actions { display: flex; align-items: center; gap: 2px; flex-shrink: 0; }
.delete-col-btn { color: var(--color-text-muted); font-size: 12px; opacity: 0.5; }
.col-edit-btn { font-size: 13px; opacity: 0.65; color: var(--color-text-muted); }
.col-edit-btn:hover { opacity: 1; color: var(--color-text); }
.delete-col-btn:hover { opacity: 1; color: var(--color-danger); }

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
  background: var(--color-primary);
  color: #fff;
  border-radius: 9999px;
  padding: 1px 7px;
  font-size: 11px;
  font-weight: 600;
}

.header-stat-tip-host {
  position: relative;
  display: inline-flex;
  align-items: center;
  vertical-align: middle;
  cursor: pointer;
  border-radius: var(--radius-sm);
}
.header-stat-tip-host:focus { outline: none; }
.header-stat-tip-host:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.header-stat-popup {
  position: absolute;
  top: calc(100% + 6px);
  left: 50%;
  transform: translateX(-50%);
  z-index: 50;
  min-width: max-content;
  max-width: 240px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.35;
  color: var(--color-text);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-md);
  text-align: center;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition: opacity 0.12s ease, visibility 0.12s;
}
/* Hover bridge so moving pointer into the popup doesn’t dismiss it */
.header-stat-popup::before {
  content: '';
  position: absolute;
  bottom: 100%;
  left: -18px;
  right: -18px;
  height: 10px;
}

.header-stat-tip-host:hover .header-stat-popup,
.header-stat-tip-host:focus-visible .header-stat-popup,
.header-stat-tip-host.tip-pinned .header-stat-popup {
  opacity: 1;
  visibility: visible;
  pointer-events: auto;
}

.wip-tip-host .wip-badge { font-size: 11px; color: var(--color-text-muted); }
.wip-tip-host.wip-over .wip-badge { color: var(--color-danger); font-weight: 700; }

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
