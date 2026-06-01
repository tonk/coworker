<template>
  <div class="board-card" :class="{ 'board-card--closed': card.closed, 'board-card--overdue': isOverdue }" role="button" tabindex="0" @click="$emit('open', card)" @keydown.enter="$emit('open', card)" @keydown.space.prevent="$emit('open', card)">
    <!-- Epic colour bar -->
    <div v-if="card.epic" class="card-epic-bar" :style="{ background: card.epic.color }" :title="card.epic.name"></div>
    <!-- Assignee avatars — top right (shows multi-assignees if present, else primary) -->
    <div v-if="allAssignees.length" class="card-avatars">
      <div
        v-for="(user, idx) in allAssignees.slice(0, 3)"
        :key="user.id"
        class="card-avatar"
        :style="{ right: (8 + idx * 18) + 'px', zIndex: 3 - idx }"
        :title="user.display_name || user.username"
      >
        <img
          v-if="avatarUrl(user)"
          :src="avatarUrl(user)"
          :alt="user.display_name || user.username"
          class="avatar-img"
        />
        <div v-else class="avatar-initials">
          {{ (user.display_name || user.username || '?').slice(0, 2).toUpperCase() }}
        </div>
      </div>
      <div v-if="allAssignees.length > 3" class="card-avatar card-avatar-more" :style="{ right: '62px', zIndex: 0 }">
        +{{ allAssignees.length - 3 }}
      </div>
    </div>

    <div v-if="card.epic" class="card-epic-badge" :style="{ background: card.epic.color + '22', color: card.epic.color }">{{ card.epic.name }}</div>
    <div class="card-ref" v-if="card.card_number">{{ cardRef }}</div>
    <div class="card-priority" v-if="card.priority !== 'none'">
      <span :class="`badge priority-${card.priority}`">{{ $t(`board.priorities.${card.priority}`) }}</span>
    </div>
    <div class="card-title">{{ card.title }}</div>
    <div class="card-labels" v-if="card.labels?.length">
      <span
        v-for="label in card.labels"
        :key="label.id"
        class="card-label"
        :style="{ background: label.color + '33', color: label.color, border: `1px solid ${label.color}66` }"
      >{{ label.name }}</span>
    </div>
    <div class="card-tags" v-if="card.tags?.length">
      <span v-for="tag in card.tags" :key="tag.id" class="card-tag">#{{ tag.name }}</span>
    </div>
    <div v-if="systemStore.scrumStorypointsEnabled && card.story_points != null" class="card-sp-badge" :title="$t('board.story_points')">
      {{ card.story_points }} SP
    </div>
    <div class="card-footer" v-if="card.due_date || card.sub_card_count > 0">
      <span v-if="card.due_date" class="card-due" :class="{ overdue: isOverdue }">
        📅 {{ formatDate(card.due_date.slice(0, 10)) }}
      </span>
      <span v-if="card.sub_card_count > 0" class="subcard-pill" :class="{ 'subcard-all-done': card.sub_cards_done === card.sub_card_count }">
        ☐ {{ card.sub_cards_done }}/{{ card.sub_card_count }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'
import { useProjectStore } from '@/stores/project'
import { useSystemStore } from '@/stores/system'

const props = defineProps({ card: { type: Object, required: true } })
defineEmits(['open'])

const { formatDate } = useDateFormat()
const projectStore = useProjectStore()
const systemStore = useSystemStore()

// Primary assignee first, then extra assignees (deduplicated)
const allAssignees = computed(() => {
  const primary = props.card.assignee ? [props.card.assignee] : []
  const extras = (props.card.assignees || []).filter(u => u.id !== props.card.assignee?.id)
  return [...primary, ...extras]
})

const cardRef = computed(() => {
  const prefix = projectStore.currentProject?.key_prefix
  return prefix && props.card.card_number ? `${prefix}-${props.card.card_number}` : null
})

const isOverdue = computed(() => {
  if (!props.card.due_date) return false
  const today = new Date()
  const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
  return props.card.due_date.slice(0, 10) < todayStr
})
</script>

<style scoped>
.card-epic-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
}
.card-epic-badge {
  font-size: 10px;
  font-weight: 600;
  border-radius: 3px;
  padding: 1px 6px;
  margin-bottom: 4px;
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.board-card {
  position: relative;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  padding-top: 14px;
  cursor: pointer;
  transition: box-shadow .15s;
  user-select: none;
}
.board-card:hover { box-shadow: var(--shadow-md); }

.card-avatars { position: absolute; top: 8px; right: 0; display: flex; }

.card-avatar {
  position: absolute;
  top: 0;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  border: 1.5px solid var(--color-surface);
}

.card-avatar-more {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-border);
  color: var(--color-text-muted);
  font-size: 9px;
  font-weight: 700;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-initials {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-ref {
  font-size: 10px;
  font-weight: 600;
  color: var(--color-text-muted);
  letter-spacing: 0.03em;
  margin-bottom: 4px;
}
.card-priority { margin-bottom: 6px; }
.card-sp-badge {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  background: var(--color-primary);
  color: #fff;
  border-radius: 10px;
  padding: 1px 7px;
  margin-bottom: 6px;
}
.card-title {
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
  margin-bottom: 8px;
  /* leave room for avatars on the right */
  padding-right: 34px;
}

.card-labels { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 8px; }
.card-label { font-size: 11px; font-weight: 600; padding: 2px 6px; border-radius: 9999px; }

.card-tags { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 8px; }
.card-tag {
  font-size: 11px;
  font-weight: 500;
  padding: 1px 6px;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  background: transparent;
}

.card-footer { display: flex; align-items: center; gap: 8px; }
.card-due { font-size: 11px; color: var(--color-text-muted); }
.card-due.overdue { color: var(--color-danger); }

.subcard-pill {
  font-size: 11px;
  color: var(--color-text-muted);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 9999px;
  padding: 0 6px;
  line-height: 18px;
}
.subcard-pill.subcard-all-done {
  color: var(--color-success);
  border-color: var(--color-success);
}

.board-card--closed { opacity: 0.6; }
.board-card--closed .card-title { text-decoration: line-through; color: var(--color-text-muted); }

.board-card--overdue {
  background: color-mix(in srgb, #ff8c00 8%, var(--color-surface));
  border-color: color-mix(in srgb, #ff8c00 35%, var(--color-border));
}
</style>
