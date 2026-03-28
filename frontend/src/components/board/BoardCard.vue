<template>
  <div class="board-card" @click="$emit('open', card)">
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
    <div class="card-footer" v-if="card.due_date">
      <span class="card-due" :class="{ overdue: isOverdue }">
        📅 {{ formatDate(card.due_date) }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'
import { useProjectStore } from '@/stores/project'

const props = defineProps({ card: { type: Object, required: true } })
defineEmits(['open'])

const { formatDate } = useDateFormat()
const projectStore = useProjectStore()

// Show multi-assignees if present, fall back to primary assignee_id field
const allAssignees = computed(() => {
  if (props.card.assignees?.length) return props.card.assignees
  if (props.card.assignee) return [props.card.assignee]
  return []
})

const cardRef = computed(() => {
  const prefix = projectStore.currentProject?.key_prefix
  return prefix && props.card.card_number ? `${prefix}-${props.card.card_number}` : null
})

const isOverdue = computed(() => {
  if (!props.card.due_date) return false
  return new Date(props.card.due_date) < new Date()
})
</script>

<style scoped>
.board-card {
  position: relative;
  background: #fff;
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

.card-footer { display: flex; align-items: center; }
.card-due { font-size: 11px; color: var(--color-text-muted); }
.card-due.overdue { color: var(--color-danger); }
</style>
