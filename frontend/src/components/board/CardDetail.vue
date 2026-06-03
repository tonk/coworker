<template>
  <BaseModal :title="isNew ? $t('board.add_card') : $t('board.edit_card')" @close="handleClose" :resizable="true" style="--modal-width: 700px">
    <div class="card-detail" :class="{ readonly }">
      <div v-if="!isNew" class="sections-menu-wrap" ref="sectionsMenuEl">
        <button class="sections-menu-btn" @click.stop="sectionsMenuOpen = !sectionsMenuOpen" :title="$t('board.toggle_sections')" :aria-label="$t('board.toggle_sections')">⋮</button>
        <div v-if="sectionsMenuOpen" class="sections-menu-dropdown">
          <div v-for="sec in sectionsConfig" :key="sec.key" class="sections-menu-item" :class="{ 'sections-menu-item--locked': !sectionEmpty[sec.key] }">
            <label class="sections-menu-label">
              <input type="checkbox" :checked="isSectionVisible(sec.key)" @change="toggleSection(sec.key)" :disabled="!sectionEmpty[sec.key]" />
              <span>{{ sec.label }}</span>
            </label>
          </div>
        </div>
      </div>
      <div v-if="parentCard" class="parent-card-nav">
        <button class="parent-card-back" @click="emit('back')">
          ← {{ parentCard.title }}
        </button>
      </div>
      <div v-if="cardRef && !isNew" class="card-ref-badge">{{ cardRef }}</div>
      <div class="form-group">
        <label class="form-label">{{ $t('board.card_title') }}</label>
        <input v-if="!locked" class="form-input" v-model="form.title" aria-label="Card title" spellcheck="true" :lang="auth.user?.locale || 'en'" :autofocus="isNew" />
        <div v-else class="description-text">{{ form.title }}</div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.description') }}</label>
        <div v-if="!locked" style="position:relative">
          <MentionDropdown v-if="descMentionUsers.length" :users="descMentionUsers" :active-index="descMentionIndex"
            @pick="descPickMention" @update:activeIndex="descMentionIndex = $event" />
          <InlineEmojiPicker
            v-if="descEmojiOpen"
            :initial-search="descEmojiQuery || ''"
            @pick="onDescEmojiPick"
            @escape="onDescEmojiEscape"
            @close="descEmojiOpen = false"
          />
          <textarea class="form-input description-textarea" v-model="form.description"
                    ref="descTextareaEl"
                    spellcheck="true" :lang="auth.user?.locale || 'en'"
                    :placeholder="$t('board.description')" rows="8"
                    @input="descOnInput" @keydown="descOnKeydown"></textarea>
        </div>
        <!-- nosemgrep: javascript.vue.security.audit.xss.templates.avoid-v-html.avoid-v-html -- renderMarkdown sanitizes with DOMPurify -->
        <div v-else class="description-text comment-text" v-html="renderMarkdown(form.description)"></div>
      </div>

      <div v-if="epicsStore.epics.length || form.epic_id" class="detail-row">
        <div class="form-group half">
          <label class="form-label" for="card-epic">{{ $t('epic.epic') }}</label>
          <select id="card-epic" class="form-input" v-model="form.epic_id" @change="saveEpic">
            <option :value="null">— {{ $t('epic.no_epic') }} —</option>
            <option v-for="e in epicsStore.epics" :key="e.id" :value="e.id">{{ e.name }}</option>
          </select>
        </div>
      </div>

      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label">{{ $t('board.priority') }}</label>
          <select class="form-input" v-model="form.priority">
            <option v-for="p in priorities" :key="p" :value="p">{{ $t(`board.priorities.${p}`) }}</option>
          </select>
        </div>
        <div class="form-group half">
          <label class="form-label">{{ $t('board.start_date') }}</label>
          <div class="date-input-row">
            <input
              class="form-input"
              type="text"
              v-model="displayStartDate"
              :placeholder="dateOnlyFormat()"
              aria-label="Start date"
              @blur="parseStartDate"
            />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs" aria-hidden="true">&#128197;</span>
              <input type="date" class="date-picker-overlay" :aria-label="$t('board.start_date')" :value="form.start_date" @change="onStartDatePickerChange" />
            </label>
            <button v-if="displayStartDate" class="btn-icon-xs" @click="displayStartDate = ''; form.start_date = ''" title="Clear" aria-label="Clear start date">×</button>
          </div>
        </div>
        <div class="form-group half">
          <label class="form-label">{{ $t('board.due_date') }}</label>
          <div class="date-input-row">
            <input
              class="form-input"
              type="text"
              v-model="displayDueDate"
              :placeholder="dateOnlyFormat()"
              aria-label="Due date"
              @blur="parseDueDate"
            />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs" aria-hidden="true">&#128197;</span>
              <input type="date" class="date-picker-overlay" :aria-label="$t('board.due_date')" :value="form.due_date" @change="onDatePickerChange" />
            </label>
            <button v-if="displayDueDate" class="btn-icon-xs" @click="displayDueDate = ''; form.due_date = ''" title="Clear" aria-label="Clear due date">×</button>
          </div>
        </div>
      </div>

      <div class="detail-row">
        <div v-if="systemStore.scrumStorypointsEnabled" class="form-group half">
          <label class="form-label">{{ $t('board.story_points') }}</label>
          <input class="form-input" type="number" min="0" step="1" v-model.number="form.story_points" style="width:90px" aria-label="Story points" />
        </div>
        <div v-if="!isNew && isSectionVisible('externalIssue')" class="form-group half">
          <label class="form-label">{{ $t('board.external_issue') }}</label>
          <div class="external-issue-row">
            <input
              class="form-input"
              type="url"
              v-model="form.external_issue_url"
              :placeholder="$t('board.external_issue_url_placeholder')"
              style="flex:1;min-width:0"
              aria-label="External issue URL"
            />
            <input
              class="form-input"
              type="text"
              v-model="form.external_issue_ref"
              :aria-label="$t('board.external_issue')"
              placeholder="#42"
              style="width:72px"
            />
            <a
              v-if="form.external_issue_url"
              :href="form.external_issue_url"
              target="_blank"
              rel="noopener noreferrer"
              class="btn-icon-xs"
              :title="$t('board.external_issue_open')"
              aria-label="Open external issue"
            >↗</a>
            <button
              v-if="form.external_issue_url"
              class="btn-icon-xs"
              type="button"
              aria-label="Clear external issue"
              @click="form.external_issue_url = ''; form.external_issue_ref = ''"
            >×</button>
          </div>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ $t('board.assignee') }}</label>
        <select class="form-input" v-model="form.assignee_id">
          <option :value="null">—</option>
          <option v-for="m in members" :key="m.user.id" :value="m.user.id">{{ m.user.display_name || m.user.username }}</option>
        </select>
      </div>

      <div v-if="!isNew" class="form-group">
        <label class="form-label">{{ $t('board.extra_assignees') }}</label>
        <div class="labels-picker">
          <span
            v-for="m in members.filter(m => m.user.id !== form.assignee_id)"
            :key="m.user.id"
            class="label-chip watcher-chip"
            :class="{ active: isAssigned(m.user.id) }"
            @click="toggleAssignee(m.user)"
          >{{ m.user.display_name || m.user.username }}</span>
        </div>
      </div>

      <div v-if="!isNew && isSectionVisible('labels')" class="form-group">
        <label class="form-label">{{ $t('board.labels') }}</label>
        <div class="labels-picker">
          <span
            v-for="label in labels"
            :key="label.id"
            class="label-chip"
            :class="{ active: hasLabel(label.id) }"
            :style="{ borderColor: label.color, color: hasLabel(label.id) ? '#fff' : label.color, background: hasLabel(label.id) ? label.color : 'transparent' }"
            @click="toggleLabel(label)"
          >{{ label.name }}</span>
        </div>
      </div>

      <div v-if="!isNew && isSectionVisible('tags')" class="form-group">
        <label class="form-label">{{ $t('board.tags') }}</label>
        <div class="tags-editor">
          <div class="tags-list" v-if="card.tags?.length">
            <span v-for="tag in card.tags" :key="tag.id" class="tag-chip">
              #{{ tag.name }}
              <button class="tag-remove" @click="removeTag(tag)" title="Remove tag" aria-label="Remove tag">×</button>
            </span>
          </div>
          <div class="tag-input-row">
            <input
              class="form-input tag-input"
              v-model="newTagName"
              :placeholder="$t('board.add_tag_placeholder')"
              @keydown.enter.prevent="addTag"
              @keydown.comma.prevent="addTag"
              aria-label="Add tag"
            />
            <button class="btn btn-secondary btn-sm" @click="addTag" :disabled="!newTagName.trim()">
              {{ $t('common.add') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="!isNew && isSectionVisible('watchers')" class="form-group">
        <label class="form-label">{{ $t('board.watchers') }}</label>
        <div class="labels-picker">
          <span
            v-for="m in members"
            :key="m.user.id"
            class="label-chip watcher-chip"
            :class="{ active: isWatching(m.user.id) }"
            @click="toggleWatcher(m.user)"
          >{{ m.user.display_name || m.user.username }}</span>
        </div>
      </div>

      <div v-if="!isNew && isSectionVisible('attachments')" class="form-group">
        <label class="form-label">Attachments</label>
        <AttachmentList :attachments="attachments" :can-delete="true" @remove="deleteAttachment" />
        <div
          class="upload-drop-zone"
          :class="{ dragging: isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave="isDragging = false"
          @drop.prevent="onDrop"
          @click="fileInput.click()"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          <span>Click or drop files to attach</span>
        </div>
        <input ref="fileInput" type="file" multiple style="display:none" @change="onFileSelected" />
        <div v-if="uploading" class="upload-progress">Uploading…</div>
      </div>

      <!-- Checklist -->
      <div v-if="!isNew && isSectionVisible('checklist')" class="checklist-section">
        <div class="checklist-header">
          <h4>{{ $t('checklist.title') }}</h4>
          <span v-if="checklist.length" class="checklist-progress">
            {{ checklist.filter(i => i.is_completed).length }}/{{ checklist.length }}
          </span>
        </div>
        <div v-if="checklist.length" class="checklist-progress-bar">
          <div class="checklist-progress-fill" :style="{ width: checklistPct + '%' }"></div>
        </div>
        <div class="checklist-items" ref="checklistListEl">
          <div v-for="item in checklist" :key="item.id" class="checklist-item" :data-id="item.id">
            <span class="checklist-drag-handle" title="Drag to reorder" aria-hidden="true">⠿</span>
            <input
              type="checkbox"
              class="checklist-checkbox"
              :checked="item.is_completed"
              :aria-label="item.body"
              @change="toggleChecklistItem(item)"
            />
            <span v-if="editingItemId !== item.id" class="checklist-body" :class="{ completed: item.is_completed }">{{ item.body }}</span>
            <input v-else class="form-input checklist-edit-input" v-model="editItemBody" aria-label="Edit checklist item" @blur="saveItemEdit(item)" @keydown.enter.prevent="saveItemEdit(item)" @keydown.esc="cancelItemEdit" />
            <button v-if="editingItemId !== item.id" class="btn-icon-xs" @click="startItemEdit(item)" title="Edit" aria-label="Edit checklist item">✏</button>
            <button class="btn-icon-xs btn-danger" @click="removeChecklistItem(item)" title="Delete" aria-label="Delete checklist item">×</button>
          </div>
        </div>
        <div class="checklist-add-row">
          <input
            class="form-input checklist-new-input"
            v-model="newChecklistItem"
            :placeholder="$t('checklist.add_item_placeholder')"
            @keydown.enter.prevent="addChecklistItem"
            aria-label="New checklist item"
          />
          <button class="btn btn-secondary btn-sm" @click="addChecklistItem" :disabled="!newChecklistItem.trim()">
            {{ $t('checklist.add_item') }}
          </button>
        </div>
      </div>

      <!-- Sub-cards section (hidden from board; shown only here) -->
      <div v-if="!isNew && !card.parent_card_id && isSectionVisible('subcards')" class="subcards-section">
        <div class="subcards-header">
          <h4>{{ $t('subcard.sub_cards') }}</h4>
          <span v-if="subCards.length" class="subcards-progress">
            {{ subCards.filter(s => s.closed).length }}/{{ subCards.length }}
          </span>
        </div>
        <div v-if="subCards.length" class="subcards-progress-bar">
          <div class="subcards-progress-fill" :style="{ width: subCardPct + '%' }"></div>
        </div>
        <div class="subcard-list">
          <div v-for="sc in subCards" :key="sc.id" class="subcard-row">
            <input type="checkbox" class="subcard-check" :checked="sc.closed" :aria-label="sc.title" @change="toggleSubCard(sc)" />
            <span class="subcard-title" :class="{ 'subcard-closed': sc.closed }">
              {{ sc.title }}
              <span class="subcard-ref">{{ cardRefString(sc) }}</span>
            </span>
            <button class="btn-icon-xs" @click="openSubCard(sc)" :title="$t('subcard.open')" aria-label="Open subcard">↗</button>
          </div>
        </div>
        <div class="subcard-add-row">
          <input
            class="form-input subcard-new-input"
            v-model="newSubCardTitle"
            :placeholder="$t('subcard.add_sub_card')"
            @keydown.enter.prevent="addSubCard"
            aria-label="Add subcard"
          />
          <button class="btn btn-secondary btn-sm" @click="addSubCard" :disabled="!newSubCardTitle.trim()">
            {{ $t('common.add') }}
          </button>
        </div>
      </div>

      <!-- Linked cards (cross-references) -->
      <div v-if="!isNew && isSectionVisible('linkedCards')" class="linked-cards-section">
        <div class="linked-cards-header">
          <h4>{{ $t('card_ref.linked_cards') }}</h4>
          <span v-if="linkedCards.length" class="linked-cards-count">{{ linkedCards.length }}</span>
        </div>
        <div class="linked-card-list">
          <div v-for="lc in linkedCards" :key="lc.ref_id" class="linked-card-row" @click="openLinkedCard(lc)">
            <span class="linked-card-ref" :class="{ 'ref-closed': lc.closed }">
              {{ lc.key_prefix }}-{{ lc.card_number }}
            </span>
            <span class="linked-card-title" :class="{ 'ref-closed': lc.closed }">{{ lc.title }}</span>
            <span class="linked-card-col">{{ lc.column_name }}</span>
            <span v-if="lc.project_slug !== projectSlug" class="linked-card-project">{{ lc.project_name }}</span>
            <button class="btn-icon-xs" @click.stop="removeLinkedCard(lc)" :title="$t('card_ref.remove_link')" aria-label="Remove linked card">✕</button>
          </div>
        </div>
        <div v-if="!locked" class="linked-card-add-row">
          <input
            class="form-input linked-card-new-input"
            v-model="newLinkedCardRef"
            :placeholder="$t('card_ref.add_link_placeholder')"
            @keydown.enter.prevent="addLinkedCard"
            aria-label="Link card"
          />
          <button class="btn btn-secondary btn-sm" @click="addLinkedCard" :disabled="!newLinkedCardRef.trim()">
            {{ $t('common.add') }}
          </button>
        </div>
      </div>

      <!-- Linked tickets -->
      <div v-if="!isNew && isSectionVisible('linkedTickets')" class="linked-cards-section">
        <div class="linked-cards-header">
          <h4>{{ $t('ticket.linked_tickets') }}</h4>
          <span v-if="linkedTickets.length" class="linked-cards-count">{{ linkedTickets.length }}</span>
        </div>
        <div class="linked-card-list">
          <div v-for="lt in linkedTickets" :key="lt.link_id" class="linked-card-row" @click="openLinkedTicket(lt)">
            <span class="linked-card-ref">{{ lt.status }}</span>
            <span class="linked-card-title">#{{ lt.ticket_id }} {{ lt.title }}</span>
            <button class="btn-icon-xs" @click.stop="removeLinkedTicket(lt)" :title="$t('ticket.remove_link')" aria-label="Remove linked ticket">✕</button>
          </div>
        </div>
      </div>

      <div v-if="!isNew && !readonly" class="comments-section">
        <div class="comments-header">
          <h4>{{ $t('board.comments') }}</h4>
          <span v-if="card.time_spent_minutes > 0" class="time-spent-total">
            ⏱ {{ $t('board.time_spent') }}: {{ fmtCardTime(card.time_spent_minutes) }}
          </span>
        </div>
        <div class="comment-list">
          <div v-for="comment in card.comments" :key="comment.id" class="comment" :class="{ 'comment-reply': comment.body.trimStart().startsWith('>') }">
            <div class="comment-avatar">
              <img v-if="avatarUrl(comment.user)" :src="avatarUrl(comment.user)" :alt="comment.user.display_name" class="comment-avatar-img" @error="e => e.target.style.display='none'" />
              <span v-else>{{ comment.user.display_name?.slice(0,2).toUpperCase() }}</span>
            </div>
            <div class="comment-body">
              <div class="comment-meta">
                <strong>{{ comment.user.display_name || comment.user.username }}</strong>
                <span class="comment-time">{{ formatDateTime(comment.created_at) }}</span>
                <span v-if="comment.is_edited" class="edited-badge" aria-hidden="true">✎</span>
                <span v-if="comment.time_spent_minutes > 0" class="comment-time-badge">
                  ⏱ {{ fmtCardTime(comment.time_spent_minutes) }}
                </span>
              </div>
              <!-- nosemgrep: javascript.vue.security.audit.xss.templates.avoid-v-html.avoid-v-html -- renderMarkdown sanitizes with DOMPurify -->
              <div class="comment-text" v-html="renderMarkdown(comment.body)"></div>
              <button class="btn btn-ghost btn-sm reply-btn" @click="replyTo(comment)">{{ $t('board.reply') }}</button>
            </div>
          </div>
        </div>

        <div class="add-comment">
          <div style="position:relative">
            <MentionDropdown v-if="commentMentionUsers.length" :users="commentMentionUsers" :active-index="commentMentionIndex"
              @pick="commentPickMention" @update:activeIndex="commentMentionIndex = $event" />
            <InlineEmojiPicker
              v-if="commentEmojiOpen"
              :initial-search="commentEmojiQuery || ''"
              @pick="onCommentEmojiPick"
              @escape="onCommentEmojiEscape"
              @close="commentEmojiOpen = false"
            />
            <textarea
              class="form-input comment-textarea"
              v-model="newComment"
              ref="commentTextareaEl"
              :placeholder="$t('board.add_comment')"
              spellcheck="true"
              :lang="auth.user?.locale || 'en'"
              rows="3"
              @input="commentOnInput"
              @keydown="commentOnKeydown"
            ></textarea>
          </div>
          <div v-if="auth.timeTrackingEnabled" class="log-time-row">
            <span class="log-time-label">{{ $t('board.log_time') }}</span>
            <input class="form-input time-input" type="number" min="0" v-model.number="newCommentHours" aria-label="Hours" />
            <span class="time-sep">{{ $t('board.time_hours') }}</span>
            <input class="form-input time-input" type="number" min="0" max="59" v-model.number="newCommentMinutes" aria-label="Minutes" />
            <span class="time-sep">{{ $t('board.time_minutes') }}</span>
          </div>
          <button class="btn btn-primary btn-sm" @click="submitComment" :disabled="!newComment.trim()">
            {{ $t('board.add_comment') }}
          </button>
        </div>
      </div>

      <div v-if="!isNew && gitLinks.length" class="git-links-section">
        <h4>Git Links</h4>
        <div class="git-links-list">
          <div
            v-for="link in gitLinks"
            :key="link.id"
            class="git-link-row"
            style="cursor: pointer"
            @click="openLink(link.url)"
          >
            <span class="git-link-icon" aria-hidden="true">
              <template v-if="link.link_type === 'commit'">⬡</template>
              <template v-else-if="link.link_type === 'pr'">⤵</template>
              <template v-else>◎</template>
            </span>
            <span class="git-link-meta">
              <span class="git-link-platform" :class="'platform-' + link.platform">{{ link.platform }}</span>
              <span class="git-link-type">{{ link.link_type === 'pr' ? 'Pull Request' : link.link_type === 'commit' ? 'Commit' : 'Issue' }}</span>
              <span v-if="link.reference" class="git-link-ref">
                {{ link.link_type === 'commit' ? link.reference.slice(0, 7) : '#' + link.reference }}
              </span>
            </span>
            <span class="git-link-title">{{ link.title }}</span>
            <span class="git-link-status" :class="'status-' + link.status">{{ link.status }}</span>
          </div>
        </div>
      </div>

      <!-- Transfer card section -->
      <div v-if="!isNew && showTransferPanel" class="transfer-section">
        <h4>{{ $t('board.transfer_card') }}</h4>
        <div class="detail-row">
          <div class="form-group half">
            <label class="form-label">{{ $t('board.transfer_project') }}</label>
            <select class="form-input" v-model="transferProjectSlug" @change="loadTransferColumns">
              <option value="">— {{ $t('board.select_project') }} —</option>
              <option v-for="p in transferProjects" :key="p.slug" :value="p.slug">{{ p.name }}</option>
            </select>
          </div>
          <div class="form-group half">
            <label class="form-label">{{ $t('board.transfer_column') }}</label>
            <select class="form-input" v-model="transferColumnId" :disabled="!transferProjectSlug">
              <option value="">— {{ $t('board.select_column') }} —</option>
              <option v-for="col in transferColumns" :key="col.id" :value="col.id">{{ col.name }}</option>
            </select>
          </div>
        </div>
        <div class="transfer-actions">
          <button class="btn btn-secondary btn-sm" @click="executeTransfer('copy')" :disabled="!transferColumnId || transferring">{{ $t('board.transfer_copy') }}</button>
          <button class="btn btn-secondary btn-sm" @click="executeTransfer('move')" :disabled="!transferColumnId || transferring">{{ $t('board.transfer_move') }}</button>
          <button class="btn btn-ghost btn-sm" @click="showTransferPanel = false">{{ $t('common.cancel') }}</button>
        </div>
      </div>

      <div v-if="!isNew && columnHistory.length" class="history-section">
        <h4>{{ $t('board.column_history') }}</h4>
        <div class="history-list">
          <div v-for="h in columnHistory" :key="h.id" class="history-entry">
            <span class="history-time">{{ formatDateTime(h.created_at) }}</span>
            <span class="history-who">{{ h.user.display_name || h.user.username }}</span>
            <span class="history-move">
              <span class="history-col">{{ h.from_column.name }}</span>
              →
              <span class="history-col">{{ h.to_column.name }}</span>
            </span>
          </div>
        </div>
      </div>

      <div v-if="!isNew && (showActivityPanel || readonly)" class="activity-section">
        <h4>{{ $t('board.card_history') }}</h4>
        <div v-if="!history.length" class="activity-empty">{{ $t('common.no_results') }}</div>
        <div v-else class="activity-list">
          <div v-for="h in history" :key="h.id" class="activity-entry">
            <span class="activity-icon" :class="`activity-type-${h.event_type}`" aria-hidden="true">{{ activityIcon(h.event_type) }}</span>
            <span class="activity-time">{{ formatDateTime(h.created_at) }}</span>
            <span class="activity-who">{{ h.user.display_name || h.user.username }}</span>
            <span class="activity-detail">{{ activityLabel(h) }}</span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <template v-if="showCancelConfirm">
        <span class="cancel-confirm-msg">{{ $t('board.unsaved_changes') }}</span>
        <button class="btn btn-primary btn-sm" @click="save">{{ $t('common.save') }}</button>
        <button class="btn btn-danger btn-sm" @click="$emit('close')">{{ $t('board.discard') }}</button>
        <button class="btn btn-ghost btn-sm" @click="showCancelConfirm = false">{{ $t('common.cancel') }}</button>
      </template>
      <template v-else-if="isNew">
        <button class="btn btn-secondary" @click="$emit('close')">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="save" :disabled="saving || !form.title.trim()">{{ saving ? $t('common.loading') : $t('common.create') }}</button>
      </template>
      <template v-else-if="readonly">
        <button class="btn btn-primary btn-sm" @click="restoreCard" :disabled="restoring">{{ restoring ? $t('common.loading') : $t('board.restore_card') }}</button>
        <button class="btn btn-secondary btn-sm" @click="handleClose">{{ $t('common.close') }}</button>
      </template>
      <template v-else>
        <button class="btn btn-danger btn-sm" @click="confirmDelete">{{ $t('board.delete_card') }}</button>
        <button class="btn btn-secondary btn-sm" @click="toggleClose">{{ isClosed ? $t('board.reopen_card') : $t('board.close_card') }}</button>
        <button class="btn btn-secondary btn-sm" @click="copyCard" :disabled="copying">{{ $t('board.copy_card') }}</button>
        <button class="btn btn-secondary btn-sm" @click="toggleTransferPanel">{{ $t('board.transfer_card') }}</button>
        <button class="btn btn-secondary btn-sm" :class="{ 'btn-active': showActivityPanel }" @click="showActivityPanel = !showActivityPanel">{{ $t('board.card_history') }}</button>
        <button class="btn btn-secondary btn-sm" @click="handleClose">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary btn-sm" @click="save" :disabled="saving">{{ saving ? $t('common.loading') : $t('common.save') }}</button>
      </template>
    </template>
  </BaseModal>

  <!-- Nested sub-card detail modal -->
  <CardDetail
    v-if="openSubCardRef"
    :card="openSubCardRef"
    :project-slug="props.projectSlug"
    :members="members"
    :labels="labels"
    :parent-card="props.card"
    @close="openSubCardRef = null; loadSubCards()"
    @back="openSubCardRef = null"
  />

  <!-- Nested linked-card detail modal -->
  <CardDetail
    v-if="openLinkedCardRef"
    :card="openLinkedCardRef"
    :project-slug="openLinkedCardSlug"
    :members="openLinkedCardSlug === props.projectSlug ? members : []"
    :labels="openLinkedCardSlug === props.projectSlug ? labels : []"
    :parent-card="props.card"
    @back="openLinkedCardRef = null; openLinkedCardSlug = null"
    @close="openLinkedCardRef = null; openLinkedCardSlug = null; loadLinkedCards()"
  />
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import Sortable from 'sortablejs'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import BaseModal from '@/components/common/BaseModal.vue'
import AttachmentList from '@/components/common/AttachmentList.vue'
import MentionDropdown from '@/components/common/MentionDropdown.vue'
import InlineEmojiPicker from '@/components/common/InlineEmojiPicker.vue'
import CardDetail from '@/components/board/CardDetail.vue'
import { useCompose } from '@/composables/useCompose'
import { useBoardStore } from '@/stores/board'
import { useProjectStore } from '@/stores/project'
import { useAuthStore } from '@/stores/auth'
import { projectsApi } from '@/api/projects'
import { ticketsApi } from '@/api/tickets'
import { attachmentsApi } from '@/api/attachments'
import { useUIStore } from '@/stores/ui'
import { useSystemStore } from '@/stores/system'
import { useEpicsStore } from '@/stores/epics'
import { useDateFormat } from '@/composables/useDateFormat'
import { avatarUrl } from '@/composables/useAvatar'

const props = defineProps({
  card: { type: Object, required: true },
  labels: { type: Array, default: () => [] },
  members: { type: Array, default: () => [] },
  projectSlug: { type: String, required: true },
  parentCard: { type: Object, default: null },
  readonly: { type: Boolean, default: false }
})
const emit = defineEmits(['close', 'deleted', 'back', 'restore'])

const { t } = useI18n()

// Flat user list for @mention in editors
const memberUsers = computed(() => props.members.map(m => m.user).filter(Boolean))

const boardStore = useBoardStore()
const systemStore = useSystemStore()
const projectStore = useProjectStore()
const epicsStore = useEpicsStore()
const ui = useUIStore()
const router = useRouter()

const cardRef = computed(() => {
  const prefix = projectStore.currentProject?.key_prefix
  return prefix && props.card.card_number ? `${prefix}-${props.card.card_number}` : null
})
const isNew = computed(() => !props.card.id)
const auth = useAuthStore()
const { formatDateTime, formatDate, dateOnlyFormat } = useDateFormat()

function parseDueDate() {
  const val = displayDueDate.value.trim()
  if (!val) {
    form.value.due_date = ''
    return
  }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY')
  const mPos = fmt.indexOf('MM')
  const dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayDueDate.value = form.value.due_date ? formatDate(form.value.due_date) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  form.value.due_date = iso
  displayDueDate.value = formatDate(iso)
}
function onDatePickerChange(e) {
  const iso = e.target.value  // always YYYY-MM-DD
  form.value.due_date = iso
  displayDueDate.value = iso ? formatDate(iso) : ''
}

const locked = ref(!!props.card.description)
const isClosed = ref(!!props.card.closed)
const newComment = ref('')
const history = ref([])
const showActivityPanel = ref(false)
const saving = ref(false)
const newTagName = ref('')
const attachments = ref([...(props.card.attachments || [])])
const uploading = ref(false)
const isDragging = ref(false)
const fileInput = ref(null)
const checklist = ref([])
const newChecklistItem = ref('')
const editingItemId = ref(null)
const editItemBody = ref('')
const checklistListEl = ref(null)
let sortableInstance = null
const assignees = ref([...(props.card.assignees || [])])

watch(() => form.value.assignee_id, async (newId) => {
  if (newId && isAssigned(newId)) {
    const member = props.members.find(m => m.user.id === newId)
    if (member) await toggleAssignee(member.user)
  }
})
const gitLinks = ref([])

watch(() => form.value.external_issue_url, (url) => {
  if (!url || form.value.external_issue_ref) return
  const m = url.match(/\/(?:issues|pull|merge_requests)\/(\d+)/)
  if (m) form.value.external_issue_ref = `#${m[1]}`
})
const copying = ref(false)
const restoring = ref(false)
const showTransferPanel = ref(false)
const showCancelConfirm = ref(false)

// Sections visibility menu — stores which sections are explicitly shown (empty = all hidden)
const sectionsMenuOpen = ref(false)
const sectionsMenuEl = ref(null)
const shownSections = ref(new Set(JSON.parse(localStorage.getItem('warmdesk-card-shown-sections') || '[]')))

const sectionsConfig = computed(() => [
  { key: 'labels', label: t('board.labels') },
  { key: 'tags', label: t('board.tags') },
  { key: 'attachments', label: t('board.attachments') },
  { key: 'checklist', label: t('checklist.title') },
  { key: 'subcards', label: t('subcard.sub_cards') },
  { key: 'linkedCards', label: t('card_ref.linked_cards') },
  { key: 'linkedTickets', label: t('ticket.linked_tickets') },
  { key: 'watchers', label: t('board.watchers') },
  { key: 'externalIssue', label: t('board.external_issue') },
].sort((a, b) => a.label.localeCompare(b.label)))

const sectionEmpty = computed(() => ({
  labels: !(props.card.labels?.length),
  tags: !(props.card.tags?.length),
  attachments: !attachments.value.length,
  checklist: !checklist.value.length,
  subcards: !subCards.value.length,
  linkedCards: !linkedCards.value.length,
  linkedTickets: !linkedTickets.value.length,
  watchers: !(props.card.watchers?.length),
  externalIssue: !form.value.external_issue_url,
}))

function isSectionVisible(key) {
  return shownSections.value.has(key) || !sectionEmpty.value[key]
}

function toggleSection(key) {
  if (!sectionEmpty.value[key]) return
  const next = new Set(shownSections.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  shownSections.value = next
  localStorage.setItem('warmdesk-card-shown-sections', JSON.stringify([...next]))
}

function onSectionsMenuDocClick(e) {
  if (sectionsMenuEl.value && !sectionsMenuEl.value.contains(e.target)) {
    sectionsMenuOpen.value = false
  }
}

watch(sectionsMenuOpen, (open) => {
  if (open) document.addEventListener('click', onSectionsMenuDocClick)
  else document.removeEventListener('click', onSectionsMenuDocClick)
})

// @mention support for description and comment textareas
const descTextareaEl = ref(null)
const commentTextareaEl = ref(null)
const transferProjectSlug = ref('')
const transferColumnId = ref('')
const transferColumns = ref([])
const transferProjects = ref([])
const transferring = ref(false)

const checklistPct = computed(() => {
  if (!checklist.value.length) return 0
  return Math.round(checklist.value.filter(i => i.is_completed).length / checklist.value.length * 100)
})

const columnHistory = computed(() =>
  history.value.filter(h => h.event_type === 'column_move' || !h.event_type)
)

function activityIcon(type) {
  const icons = {
    created: '✦', column_move: '→', closed: '✓', reopened: '↑',
    deleted: '✗', restored: '↩', title_changed: '✎', priority_changed: '⚑', assignee_changed: '◉',
    comment_added: '✉', start_date_changed: '▷', due_date_changed: '⏱',
    description_changed: '¶',
  }
  return icons[type] || '•'
}

function activityLabel(h) {
  switch (h.event_type) {
    case 'created': return t('board.history_created')
    case 'closed': return t('board.history_closed')
    case 'reopened': return t('board.history_reopened')
    case 'deleted': return t('board.history_deleted')
    case 'restored': return t('board.history_restored')
    case 'column_move': return `${h.from_column?.name ?? '?'} → ${h.to_column?.name ?? '?'}`
    case 'title_changed': return `${t('board.history_title_changed')}: "${h.detail}"`
    case 'priority_changed': return `${t('board.history_priority_changed')}: ${h.detail}`
    case 'assignee_changed': return `${t('board.history_assignee_changed')}: ${h.detail}`
    case 'comment_added': return t('board.history_comment_added')
    case 'start_date_changed': return `${t('board.history_start_date_changed')}: ${h.detail}`
    case 'due_date_changed': return `${t('board.history_due_date_changed')}: ${h.detail}`
    case 'description_changed': return t('board.history_description_changed')
    default: return h.detail || h.event_type
  }
}

// Sub-cards
const subCards = ref([])
const newSubCardTitle = ref('')
const openSubCardRef = ref(null) // card to show in nested modal

const subCardPct = computed(() => {
  if (!subCards.value.length) return 0
  return Math.round(subCards.value.filter(s => s.closed).length / subCards.value.length * 100)
})

function cardRefString(sc) {
  // sc may have a card_number; build a reference like the parent project's key
  return sc.card_number ? ` #${sc.card_number}` : ''
}

async function loadSubCards() {
  if (props.card.parent_card_id) return // don't load sub-cards for sub-cards
  try {
    const { data } = await projectsApi.listSubCards(props.projectSlug, props.card.id)
    subCards.value = data || []
  } catch {}
}

async function addSubCard() {
  if (props.readonly) return
  const title = newSubCardTitle.value.trim()
  if (!title) return
  try {
    const { data } = await projectsApi.createSubCard(props.projectSlug, props.card.id, { title })
    subCards.value.push(data)
    newSubCardTitle.value = ''
  } catch {}
}

async function toggleSubCard(sc) {
  if (props.readonly) return
  try {
    const { data } = await projectsApi.updateCard(props.projectSlug, sc.id, { closed: !sc.closed })
    const idx = subCards.value.findIndex(s => s.id === sc.id)
    if (idx !== -1) subCards.value[idx] = { ...subCards.value[idx], closed: data.closed }
  } catch {}
}

function openSubCard(sc) {
  openSubCardRef.value = sc
}

// Linked cards (cross-references)
const linkedCards = ref([])
const newLinkedCardRef = ref('')

// Linked tickets
const linkedTickets = ref([])

async function loadLinkedCards() {
  try {
    const { data } = await projectsApi.listCardRefs(props.projectSlug, props.card.id)
    linkedCards.value = data || []
  } catch {}
}

async function loadLinkedTickets() {
  try {
    const { data } = await projectsApi.listCardTickets(props.projectSlug, props.card.id)
    linkedTickets.value = data || []
  } catch {}
}

async function addLinkedCard() {
  if (props.readonly) return
  const ref = newLinkedCardRef.value.trim()
  if (!ref) return
  try {
    const { data } = await projectsApi.createCardRef(props.projectSlug, props.card.id, { ref })
    linkedCards.value.push(data)
    newLinkedCardRef.value = ''
  } catch (e) {
    const msg = e?.response?.data?.error || 'Failed to link card'
    ui.error(msg)
  }
}

async function removeLinkedCard(lc) {
  if (props.readonly) return
  try {
    await projectsApi.deleteCardRef(props.projectSlug, props.card.id, lc.ref_id)
    linkedCards.value = linkedCards.value.filter(c => c.ref_id !== lc.ref_id)
  } catch {}
}

const openLinkedCardRef = ref(null)
const openLinkedCardSlug = ref(null)

async function openLinkedCard(lc) {
  try {
    const { data } = await projectsApi.getCard(lc.project_slug, lc.id)
    openLinkedCardSlug.value = lc.project_slug
    openLinkedCardRef.value = data
  } catch {}
}

function openLinkedTicket(lt) {
  router.push({ name: 'ticket-detail', params: { id: lt.customer_id, ticketId: lt.ticket_id } })
}

async function removeLinkedTicket(lt) {
  if (props.readonly) return
  try {
    await ticketsApi.removeCardLink(lt.customer_id, lt.ticket_id, lt.link_id)
    await loadLinkedTickets()
  } catch {}
}

function isAssigned(userId) {
  return assignees.value.some(a => a.id === userId)
}

async function toggleAssignee(user) {
  if (props.readonly) return
  try {
    if (isAssigned(user.id)) {
      await projectsApi.removeAssignee(props.projectSlug, props.card.id, user.id)
      assignees.value = assignees.value.filter(a => a.id !== user.id)
    } else {
      const { data } = await projectsApi.addAssignee(props.projectSlug, props.card.id, user.id)
      assignees.value = data
    }
    boardStore.updateCard({ ...props.card, assignees: [...assignees.value] })
  } catch {
    ui.error('Failed to update assignees')
  }
}

async function addChecklistItem() {
  if (props.readonly) return
  const body = newChecklistItem.value.trim()
  if (!body) return
  try {
    const { data } = await projectsApi.createChecklistItem(props.projectSlug, props.card.id, body)
    checklist.value = [...checklist.value, data]
    newChecklistItem.value = ''
  } catch {
    ui.error('Failed to add checklist item')
  }
}

async function toggleChecklistItem(item) {
  if (props.readonly) return
  try {
    const { data } = await projectsApi.updateChecklistItem(props.projectSlug, props.card.id, item.id, { is_completed: !item.is_completed })
    const idx = checklist.value.findIndex(i => i.id === item.id)
    if (idx !== -1) checklist.value[idx] = data
  } catch {
    ui.error('Failed to update item')
  }
}

function startItemEdit(item) {
  editingItemId.value = item.id
  editItemBody.value = item.body
}

function cancelItemEdit() {
  editingItemId.value = null
  editItemBody.value = ''
}

async function saveItemEdit(item) {
  if (props.readonly) return
  if (!editItemBody.value.trim()) { cancelItemEdit(); return }
  try {
    const { data } = await projectsApi.updateChecklistItem(props.projectSlug, props.card.id, item.id, { body: editItemBody.value })
    const idx = checklist.value.findIndex(i => i.id === item.id)
    if (idx !== -1) checklist.value[idx] = data
    cancelItemEdit()
  } catch {
    ui.error('Failed to update item')
  }
}

async function removeChecklistItem(item) {
  if (props.readonly) return
  try {
    await projectsApi.deleteChecklistItem(props.projectSlug, props.card.id, item.id)
    checklist.value = checklist.value.filter(i => i.id !== item.id)
  } catch {
    ui.error('Failed to delete item')
  }
}

async function addTag() {
  if (props.readonly) return
  const name = newTagName.value.trim().replace(/^#/, '')
  if (!name) return
  try {
    const { data } = await projectsApi.addCardTag(props.projectSlug, props.card.id, name)
    if (!props.card.tags) props.card.tags = []
    if (!props.card.tags.some(t => t.id === data.id)) {
      props.card.tags = [...props.card.tags, data]
    }
    boardStore.updateCard({ ...props.card })
    newTagName.value = ''
  } catch (e) {
    ui.error('Failed to add tag')
  }
}

async function removeTag(tag) {
  if (props.readonly) return
  try {
    await projectsApi.removeCardTag(props.projectSlug, props.card.id, tag.id)
    props.card.tags = (props.card.tags || []).filter(t => t.id !== tag.id)
    boardStore.updateCard({ ...props.card })
  } catch (e) {
    ui.error('Failed to remove tag')
  }
}

async function uploadFiles(files) {
  if (props.readonly) return
  if (!files.length) return
  uploading.value = true
  for (const file of files) {
    try {
      const fd = new FormData()
      fd.append('file', file)
      fd.append('owner_type', 'card')
      fd.append('owner_id', String(props.card.id))
      const { data } = await attachmentsApi.upload(fd)
      attachments.value = [...attachments.value, data]
    } catch (e) {
      ui.error(`Failed to upload ${file.name}`)
    }
  }
  uploading.value = false
}

function onFileSelected(e) {
  uploadFiles([...e.target.files])
  e.target.value = ''
}

function onDrop(e) {
  isDragging.value = false
  uploadFiles([...e.dataTransfer.files])
}

async function deleteAttachment(a) {
  if (props.readonly) return
  try {
    await attachmentsApi.delete(a.id)
    attachments.value = attachments.value.filter(x => x.id !== a.id)
  } catch (e) {
    ui.error('Failed to delete attachment')
  }
}

function handleClose() {
  if (props.readonly) { emit('close'); return }
  if (!isDirty.value) { emit('close'); return }
  showCancelConfirm.value = true
}

function onKeyDown(e) {
  if (e.key === 'Escape' && !e.defaultPrevented) {
    if (props.readonly || !isDirty.value) emit('close')
  }
}

async function openLink(url) {
  if (window.__TAURI_INTERNALS__) {
    await window.__TAURI_INTERNALS__.invoke('plugin:opener|open_url', { url, with: null })
  } else {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}

function initSortable() {
  if (!checklistListEl.value || sortableInstance) return
  sortableInstance = new Sortable(checklistListEl.value, {
    animation: 150,
    handle: '.checklist-drag-handle',
    onEnd(evt) {
      if (evt.oldIndex === evt.newIndex) return
      const moved = checklist.value.splice(evt.oldIndex, 1)[0]
      checklist.value.splice(evt.newIndex, 0, moved)
      const items = checklist.value.map((item, i) => ({ id: item.id, position: (i + 1) * 1000 }))
      checklist.value.forEach((item, i) => { item.position = (i + 1) * 1000 })
      projectsApi.reorderChecklistItems(props.projectSlug, props.card.id, items).catch(() => {})
    },
  })
}

watch(checklist, async (val) => {
  if (val.length && !sortableInstance) {
    await nextTick()
    initSortable()
  }
}, { flush: 'post' })

watch(() => boardStore.checklistEvent, (event) => {
  if (!event) return
  const { type, payload } = event
  if (Number(payload.card_id) !== props.card.id) return
  if (type === 'board.card.checklist.created') {
    if (!checklist.value.find(i => i.id === payload.item.id)) {
      checklist.value.push(payload.item)
    }
  } else if (type === 'board.card.checklist.updated') {
    const idx = checklist.value.findIndex(i => i.id === payload.item.id)
    if (idx !== -1) checklist.value[idx] = payload.item
  } else if (type === 'board.card.checklist.deleted') {
    checklist.value = checklist.value.filter(i => i.id !== payload.item_id)
  } else if (type === 'board.card.checklist.reordered') {
    checklist.value = payload.items
  }
})

onMounted(async () => {
  ui.setHelpContext('cardDetail')
  document.addEventListener('keydown', onKeyDown)
  if (epicsStore.projectSlug !== props.projectSlug) {
    epicsStore.loadEpics(props.projectSlug)
  }
  if (isNew.value) return
  const [histRes, checkRes, linksRes] = await Promise.all([
    projectsApi.getCardHistory(props.projectSlug, props.card.id).catch(() => ({ data: [] })),
    projectsApi.listChecklist(props.projectSlug, props.card.id).catch(() => ({ data: [] })),
    projectsApi.getCardLinks(props.projectSlug, props.card.id).catch(() => ({ data: [] })),
  ])
  history.value = histRes.data
  checklist.value = checkRes.data || []
  gitLinks.value = linksRes.data || []
  await loadSubCards()
  await loadLinkedCards()
  await loadLinkedTickets()
})

onUnmounted(() => {
  ui.setHelpContext(null)
  document.removeEventListener('keydown', onKeyDown)
  document.removeEventListener('click', onSectionsMenuDocClick)
  sortableInstance?.destroy()
  sortableInstance = null
})

const priorities = ['none', 'low', 'medium', 'high', 'critical']

const _today = new Date()
const todayISO = `${_today.getFullYear()}-${String(_today.getMonth() + 1).padStart(2, '0')}-${String(_today.getDate()).padStart(2, '0')}`

const form = ref({
  title: props.card.title,
  description: props.card.description || '',
  priority: props.card.priority || 'none',
  start_date: props.card.start_date ? props.card.start_date.slice(0, 10) : '',
  due_date: props.card.due_date ? props.card.due_date.slice(0, 10) : '',
  assignee_id: props.card.assignee_id || null,
  epic_id: props.card.epic_id || null,
  story_points: props.card.story_points ?? null,
  external_issue_url: props.card.external_issue_url || '',
  external_issue_ref: props.card.external_issue_ref || '',
})

// Snapshot for dirty-check (plain values, not reactive)
const _init = {
  title: props.card.title,
  description: props.card.description || '',
  priority: props.card.priority || 'none',
  start_date: props.card.start_date ? props.card.start_date.slice(0, 10) : '',
  due_date: props.card.due_date ? props.card.due_date.slice(0, 10) : '',
  assignee_id: props.card.assignee_id || null,
  story_points: props.card.story_points ?? null,
  external_issue_url: props.card.external_issue_url || '',
  external_issue_ref: props.card.external_issue_ref || '',
}
const isDirty = computed(() => {
  if (props.readonly) return false
  return (
    newComment.value.trim() !== '' ||
    newCommentHours.value > 0 ||
    newCommentMinutes.value > 0 ||
    form.value.title !== _init.title ||
    form.value.description !== _init.description ||
    form.value.priority !== _init.priority ||
    form.value.start_date !== _init.start_date ||
    form.value.due_date !== _init.due_date ||
    form.value.assignee_id !== _init.assignee_id ||
    form.value.story_points !== _init.story_points ||
    form.value.external_issue_url !== _init.external_issue_url ||
    form.value.external_issue_ref !== _init.external_issue_ref
  )
})

const {
  mentionUsers: descMentionUsers, mentionIndex: descMentionIndex,
  onTextareaInput: descOnInput, onTextareaKeydown: descOnKeydown, pickMention: descPickMention, emojiQuery: descEmojiQuery, pickEmoji: pickDescEmoji
} = useCompose({
  textareaEl: descTextareaEl,
  getValue: () => form.value.description,
  setValue: (v) => { form.value.description = v },
  users: memberUsers
})

const {
  mentionUsers: commentMentionUsers, mentionIndex: commentMentionIndex,
  onTextareaInput: commentOnInput, onTextareaKeydown: commentOnKeydown, pickMention: commentPickMention, emojiQuery: commentEmojiQuery, pickEmoji: pickCommentEmoji
} = useCompose({
  textareaEl: commentTextareaEl,
  getValue: () => newComment.value,
  setValue: (v) => { newComment.value = v },
  users: memberUsers
})

const descEmojiOpen = ref(false)
const commentEmojiOpen = ref(false)

watch(descEmojiQuery, (q) => { descEmojiOpen.value = q !== null })
watch(commentEmojiQuery, (q) => { commentEmojiOpen.value = q !== null })

function onDescEmojiPick(emoji) {
  pickDescEmoji(emoji)
  descEmojiOpen.value = false
}

function onDescEmojiEscape() {
  descEmojiOpen.value = false
  descTextareaEl.value?.focus()
}

function onCommentEmojiPick(emoji) {
  pickCommentEmoji(emoji)
  commentEmojiOpen.value = false
}

function onCommentEmojiEscape() {
  commentEmojiOpen.value = false
  commentTextareaEl.value?.focus()
}

const displayStartDate = ref(form.value.start_date ? formatDate(form.value.start_date) : '')

function parseStartDate() {
  const val = displayStartDate.value.trim()
  if (!val) { form.value.start_date = ''; return }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY'), mPos = fmt.indexOf('MM'), dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayStartDate.value = form.value.start_date ? formatDate(form.value.start_date) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  form.value.start_date = iso
  displayStartDate.value = formatDate(iso)
}

function onStartDatePickerChange(e) {
  const iso = e.target.value
  form.value.start_date = iso
  displayStartDate.value = iso ? formatDate(iso) : ''
}

const displayDueDate = ref(form.value.due_date ? formatDate(form.value.due_date) : '')

// Time logging on new comments
const newCommentHours   = ref(0)
const newCommentMinutes = ref(0)

function fmtCardTime(minutes) {
  if (!minutes) return '0m'
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return h > 0 ? (m > 0 ? `${h}h ${m}m` : `${h}h`) : `${m}m`
}

function hasLabel(labelId) {
  return props.card.labels?.some(l => l.id === labelId)
}

function isWatching(userId) {
  return props.card.watchers?.some(w => w.id === userId)
}

async function toggleWatcher(user) {
  if (props.readonly) return
  try {
    if (isWatching(user.id)) {
      await projectsApi.removeWatcher(props.projectSlug, props.card.id, user.id)
      props.card.watchers = props.card.watchers.filter(w => w.id !== user.id)
    } else {
      await projectsApi.addWatcher(props.projectSlug, props.card.id, user.id)
      props.card.watchers = [...(props.card.watchers || []), user]
    }
  } catch (e) {
    ui.error('Failed to update watchers')
  }
}

async function toggleLabel(label) {
  if (props.readonly) return
  try {
    if (hasLabel(label.id)) {
      await projectsApi.removeLabel(props.projectSlug, props.card.id, label.id)
      props.card.labels = props.card.labels.filter(l => l.id !== label.id)
    } else {
      await projectsApi.assignLabel(props.projectSlug, props.card.id, label.id)
      props.card.labels = [...(props.card.labels || []), label]
    }
    boardStore.updateCard({ ...props.card })
  } catch (e) {
    ui.error('Failed to update label')
  }
}

async function saveEpic() {
  if (props.readonly || !props.card.id) return
  await boardStore.updateCardData(props.card.id, { epic_id: form.value.epic_id ?? null })
}

async function save() {
  if (props.readonly) return
  saving.value = true
  try {
    const payload = {
      title: form.value.title,
      description: form.value.description,
      priority: form.value.priority,
      start_date: form.value.start_date || null,
      due_date: form.value.due_date || null,
      assignee_id: form.value.assignee_id,
      epic_id: form.value.epic_id ?? null,
      story_points: form.value.story_points,
      external_issue_url: form.value.external_issue_url || '',
      external_issue_ref: form.value.external_issue_ref || '',
    }
    if (isNew.value) {
      await boardStore.createCard(props.card.column_id, payload)
      emit('close')
    } else {
      await boardStore.updateCardData(props.card.id, payload)
      locked.value = true
      if (newComment.value.trim()) await submitComment()
      ui.success('Saved')
      emit('close')
    }
  } catch (e) {
    ui.error(isNew.value ? 'Failed to create card' : 'Failed to save')
  } finally {
    saving.value = false
  }
}

async function submitComment() {
  if (!newComment.value.trim()) return
  const timeMinutes = (newCommentHours.value || 0) * 60 + (newCommentMinutes.value || 0)
  try {
    const { data } = await projectsApi.createComment(props.projectSlug, props.card.id, {
      body: newComment.value,
      time_spent_minutes: timeMinutes,
    })
    props.card.comments = [...(props.card.comments || []), data]
    if (timeMinutes > 0) {
      props.card.time_spent_minutes = (props.card.time_spent_minutes || 0) + timeMinutes
    }
    newComment.value = ''
    newCommentHours.value = 0
    newCommentMinutes.value = 0
  } catch (e) {
    ui.error('Failed to post comment')
  }
}

function replyTo(comment) {
  const author = comment.user.display_name || comment.user.username
  const quoted = comment.body.split('\n').map(l => `> ${l}`).join('\n')
  newComment.value = `> **${author}**\n${quoted}\n\n`
}

async function confirmDelete() {
  if (!await ui.confirm('Delete this card?')) return
  try {
    await boardStore.deleteCard(props.card.id, props.card.column_id)
    emit('deleted')
    emit('close')
  } catch (e) {
    ui.error('Failed to delete card')
  }
}

async function restoreCard() {
  restoring.value = true
  try {
    await projectsApi.restoreCard(props.projectSlug, props.card.id)
    ui.success(t('board.card_restored'))
    emit('restore')
    emit('close')
  } catch (e) {
    ui.error(t('board.card_restore_failed'))
  } finally {
    restoring.value = false
  }
}

async function toggleClose() {
  try {
    await boardStore.updateCardData(props.card.id, { closed: !isClosed.value })
    isClosed.value = !isClosed.value
    emit('close')
  } catch (e) {
    ui.error('Failed to update card')
  }
}

async function copyCard() {
  copying.value = true
  try {
    const { data } = await projectsApi.copyCard(props.projectSlug, props.card.id)
    boardStore.addCard(data)
    ui.success(t('board.copy_card_success'))
  } catch {
    ui.error('Failed to copy card')
  } finally {
    copying.value = false
  }
}

async function toggleTransferPanel() {
  showTransferPanel.value = !showTransferPanel.value
  if (showTransferPanel.value && !transferProjects.value.length) {
    try {
      const { data } = await projectsApi.list()
      transferProjects.value = (data || []).filter(p => !p.is_archived)
    } catch {}
  }
}

async function loadTransferColumns() {
  transferColumnId.value = ''
  transferColumns.value = []
  if (!transferProjectSlug.value) return
  try {
    const { data } = await projectsApi.listColumns(transferProjectSlug.value)
    transferColumns.value = data || []
  } catch {}
}

async function executeTransfer(action) {
  if (!transferProjectSlug.value || !transferColumnId.value) return
  transferring.value = true
  try {
    await projectsApi.transferCard(props.projectSlug, props.card.id, {
      target_project_slug: transferProjectSlug.value,
      column_id: parseInt(transferColumnId.value),
      action
    })
    if (action === 'move') {
      boardStore.removeCard({ card_id: props.card.id, column_id: props.card.column_id })
    }
    ui.success(action === 'copy' ? t('board.copy_card_success') : t('board.move_card_success'))
    emit('close')
  } catch {
    ui.error(`Failed to ${action} card`)
  } finally {
    transferring.value = false
  }
}

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}
</script>

<style scoped>
/* Sections visibility menu */
.sections-menu-wrap {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 20;
}
.sections-menu-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  font-size: 18px;
  line-height: 1;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  letter-spacing: 0.05em;
}
.sections-menu-btn:hover {
  background: var(--color-bg);
  color: var(--color-text);
}
.sections-menu-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 160px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: 0 4px 12px rgba(0,0,0,.12);
  padding: 4px 0;
}
.sections-menu-item {
  padding: 0;
}
.sections-menu-label {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  font-size: 13px;
  cursor: pointer;
  color: var(--color-text);
  user-select: none;
}
.sections-menu-label:hover {
  background: var(--color-bg);
}
.sections-menu-item--locked .sections-menu-label {
  color: var(--color-text-muted);
  cursor: default;
}
.sections-menu-item--locked .sections-menu-label:hover {
  background: none;
}

.card-ref-badge {
  display: inline-block;
  font-size: 11px;
  font-weight: 700;
  color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-primary) 25%, transparent);
  border-radius: 4px;
  padding: 2px 7px;
  margin-bottom: 12px;
  letter-spacing: 0.04em;
}
.card-detail { position: relative; padding-bottom: 8px; }

.form-hint { font-size: 11px; color: var(--color-text-muted); margin-top: 4px; display: block; }

.detail-row { display: flex; gap: 16px; }
.half { flex: 1; }
.time-input-row { display: flex; align-items: center; gap: 6px; }
.time-input { width: 70px; }
.time-sep { font-size: 13px; color: var(--color-text-muted); }

.tags-editor { display: flex; flex-direction: column; gap: 8px; }
.tags-list { display: flex; flex-wrap: wrap; gap: 6px; }
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  background: transparent;
}
.tag-remove {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  color: var(--color-text-muted);
  padding: 0 1px;
}
.tag-remove:hover { color: var(--color-danger); }
.tag-input-row { display: flex; gap: 8px; align-items: center; }
.tag-input { flex: 1; }

.labels-picker { display: flex; flex-wrap: wrap; gap: 6px; }
.label-chip {
  padding: 3px 10px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 600;
  border: 2px solid;
  cursor: pointer;
  transition: all .15s;
}

.upload-drop-zone {
  margin-top: 8px;
  border: 2px dashed var(--color-border);
  border-radius: var(--radius);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: border-color .15s, background .15s;
}
.upload-drop-zone:hover, .upload-drop-zone.dragging {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 5%, transparent);
  color: var(--color-primary);
}
.upload-progress { font-size: 12px; color: var(--color-text-muted); margin-top: 6px; }

.checklist-section { margin-top: 24px; margin-bottom: 32px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.checklist-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.checklist-header h4 { margin: 0; font-size: 14px; }
.checklist-progress { font-size: 12px; font-weight: 600; color: var(--color-text-muted); }
.checklist-progress-bar { height: 4px; background: var(--color-border); border-radius: 2px; margin-bottom: 12px; overflow: hidden; }
.checklist-progress-fill { height: 100%; background: var(--color-primary); border-radius: 2px; transition: width .3s; }

.checklist-items { display: flex; flex-direction: column; gap: 4px; margin-bottom: 12px; }
.checklist-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; }
.checklist-drag-handle { color: var(--color-text-muted); cursor: grab; font-size: 14px; flex-shrink: 0; user-select: none; }
.checklist-drag-handle:active { cursor: grabbing; }
.sortable-ghost { opacity: 0.4; }
.sortable-drag { opacity: 1; }
.checklist-checkbox { width: 15px; height: 15px; cursor: pointer; flex-shrink: 0; accent-color: var(--color-primary); }
.checklist-body { flex: 1; font-size: 13px; line-height: 1.4; }
.checklist-body.completed { text-decoration: line-through; color: var(--color-text-muted); }
.checklist-edit-input { flex: 1; padding: 2px 8px; font-size: 13px; }

.checklist-add-row { display: flex; gap: 8px; align-items: center; }
.checklist-new-input { flex: 1; }

.btn-icon-xs {
  background: none; border: none; cursor: pointer; color: var(--color-text-muted);
  padding: 2px 4px; font-size: 13px; line-height: 1; border-radius: 3px; flex-shrink: 0;
}
.btn-icon-xs:hover { background: var(--color-bg); color: var(--color-text); }
.btn-icon-xs.btn-danger:hover { color: var(--color-danger); }

.comments-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.comments-section h4 { font-size: 14px; }

.comment-list { display: flex; flex-direction: column; gap: 14px; margin-bottom: 20px; }
.comment { display: flex; gap: 10px; }
.comment-reply { margin-left: 28px; padding-left: 12px; border-left: 3px solid var(--color-border); }

.comment-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}
.comment-avatar-img { width: 100%; height: 100%; object-fit: cover; border-radius: 50%; }

.comment-body { flex: 1; }
.comments-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.comments-header h4 { margin: 0; }
.time-spent-total { font-size: 12px; color: var(--color-primary); font-weight: 600; }

.comment-meta { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; font-size: 12px; flex-wrap: wrap; }
.comment-time { color: var(--color-text-muted); }
.edited-badge { color: var(--color-text-muted); font-style: italic; font-size: 11px; }
.comment-time-badge { color: var(--color-primary); font-size: 11px; font-weight: 600; }

.log-time-row { display: flex; align-items: center; gap: 6px; font-size: 13px; }
.log-time-label { color: var(--color-text-muted); font-size: 12px; margin-right: 4px; }

.comment-text { font-size: 13px; line-height: 1.5; }
.comment-text :deep(p) { margin-bottom: 6px; }
.comment-text :deep(code) {
  background: var(--color-border);
  color: var(--color-text);
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 12px;
}
.comment-text :deep(pre) {
  background: var(--color-bg);
  color: var(--color-text);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  overflow-x: auto;
  margin: 6px 0;
  font-size: 12px;
  line-height: 1.5;
}
.comment-text :deep(pre code) {
  background: transparent;
  padding: 0;
  border-radius: 0;
  font-size: inherit;
}

.watcher-chip {
  border-color: var(--color-text-muted) !important;
  color: var(--color-text-muted) !important;
  background: transparent !important;
}
.watcher-chip.active {
  border-color: var(--color-primary) !important;
  color: #fff !important;
  background: var(--color-primary) !important;
}

.description-text {
  padding: 8px 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  min-height: 40px;
  font-size: 13px;
  line-height: 1.5;
}

.reply-btn { margin-top: 4px; font-size: 12px; color: var(--color-text-muted); padding: 2px 8px; }
.reply-btn:hover { color: var(--color-primary); }

.add-comment { display: flex; flex-direction: column; gap: 8px; }
.add-comment .btn { align-self: flex-end; }

.git-links-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.git-links-section h4 { margin-bottom: 10px; font-size: 14px; }
.git-links-list { display: flex; flex-direction: column; gap: 4px; }
.git-link-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  font-size: 12px;
  text-decoration: none;
  color: var(--color-text);
  background: var(--color-surface);
  transition: background .12s;
}
.git-link-row:hover { background: var(--color-bg); }
.git-link-icon { font-size: 14px; color: var(--color-text-muted); flex-shrink: 0; }
.git-link-meta { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }
.git-link-platform {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  padding: 1px 5px;
  border-radius: 3px;
  letter-spacing: .04em;
}
.platform-github { background: #24292e; color: #fff; }
.platform-gitlab { background: #e24329; color: #fff; }
.platform-gitea  { background: #609926; color: #fff; }
.platform-forgejo { background: #4a8ab5; color: #fff; }
.git-link-type { color: var(--color-text-muted); }
.git-link-ref { font-family: monospace; color: var(--color-primary); }
.git-link-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--color-text); }
.git-link-status {
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  padding: 1px 6px;
  border-radius: 9999px;
  letter-spacing: .03em;
}
.status-open   { background: color-mix(in srgb, #22c55e 15%, transparent); color: #16a34a; }
.status-closed { background: color-mix(in srgb, #ef4444 15%, transparent); color: #dc2626; }
.status-merged { background: color-mix(in srgb, #8b5cf6 15%, transparent); color: #7c3aed; }

.transfer-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.transfer-section h4 { margin-bottom: 12px; font-size: 14px; color: var(--color-text-muted); }
.transfer-actions { display: flex; gap: 8px; margin-top: 8px; }

.history-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.history-section h4 { margin-bottom: 12px; font-size: 14px; color: var(--color-text-muted); }
.history-list { display: flex; flex-direction: column; gap: 6px; }
.history-entry { display: flex; align-items: center; gap: 10px; font-size: 12px; }
.history-time { color: var(--color-text-muted); flex-shrink: 0; }
.history-who { font-weight: 600; flex-shrink: 0; }
.history-move { display: flex; align-items: center; gap: 6px; color: var(--color-text-muted); }
.history-col { background: var(--color-bg); border: 1px solid var(--color-border); border-radius: var(--radius-sm); padding: 1px 6px; color: var(--color-text); font-size: 11px; }

.activity-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.activity-section h4 { margin-bottom: 12px; font-size: 14px; color: var(--color-text-muted); }
.activity-empty { font-size: 12px; color: var(--color-text-muted); }
.activity-list { display: flex; flex-direction: column; gap: 6px; }
.activity-entry { display: flex; align-items: baseline; gap: 8px; font-size: 12px; }
.activity-icon { font-size: 11px; flex-shrink: 0; width: 14px; text-align: center; }
.activity-time { color: var(--color-text-muted); flex-shrink: 0; }
.activity-who { font-weight: 600; flex-shrink: 0; }
.activity-detail { color: var(--color-text-muted); }
.activity-type-created { color: #22c55e; }
.activity-type-closed { color: #ef4444; }
.activity-type-reopened { color: #22c55e; }
.activity-type-column_move { color: var(--color-primary); }
.activity-type-title_changed, .activity-type-priority_changed, .activity-type-assignee_changed,
.activity-type-comment_added, .activity-type-start_date_changed, .activity-type-due_date_changed,
.activity-type-description_changed { color: var(--color-text-muted); }

.btn-active { background: var(--color-primary) !important; color: #fff !important; border-color: var(--color-primary) !important; }

.date-input-row { display: flex; align-items: center; gap: 6px; }
.date-input-row .form-input { flex: 1; }

.external-issue-row { display: flex; align-items: center; gap: 6px; }

.picker-wrap {
  position: relative;
  display: inline-flex;
  cursor: pointer;
}
.date-picker-overlay {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

.comment-textarea { resize: vertical; min-height: 80px; font-family: inherit; }
.description-textarea { resize: vertical; min-height: 160px; font-family: monospace; font-size: 13px; }

.cancel-confirm-msg { font-size: 13px; color: var(--color-text-muted); flex: 1; }

/* Sub-cards */
.subcards-section { margin-bottom: 20px; }
.subcards-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.subcards-header h4 { margin: 0; font-size: 14px; font-weight: 600; }
.subcards-progress { font-size: 12px; color: var(--color-text-muted); }
.subcards-progress-bar { height: 4px; background: var(--color-border); border-radius: 2px; margin-bottom: 8px; overflow: hidden; }
.subcards-progress-fill { height: 100%; background: var(--color-primary); border-radius: 2px; transition: width .3s; }
.subcard-list { display: flex; flex-direction: column; gap: 4px; }
.subcard-row { display: flex; align-items: center; gap: 8px; padding: 4px 0; }
.subcard-check { flex-shrink: 0; }
.subcard-title { flex: 1; font-size: 13px; }
.subcard-closed { text-decoration: line-through; color: var(--color-text-muted); }
.subcard-ref { font-size: 11px; color: var(--color-text-muted); margin-left: 4px; }
.subcard-add-row { display: flex; gap: 8px; margin-top: 8px; }
.subcard-new-input { flex: 1; padding: 6px 8px; font-size: 13px; }
.parent-card-nav {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--color-border);
}
.parent-card-back {
  background: none; border: none; padding: 0; cursor: pointer;
  font-size: 12px; color: var(--color-text-muted);
  display: flex; align-items: center; gap: 4px;
}
.parent-card-back:hover { color: var(--color-primary); }


.linked-cards-section { margin-top: 24px; border-top: 1px solid var(--color-border); padding-top: 20px; }
.linked-cards-header { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.linked-cards-header h4 { margin: 0; font-size: 14px; font-weight: 600; }
.linked-cards-count { font-size: 12px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 10px; padding: 1px 7px; color: var(--color-text-muted); }
.linked-card-list { display: flex; flex-direction: column; gap: 4px; margin-bottom: 8px; }
.linked-card-row { display: flex; align-items: center; gap: 8px; padding: 5px 6px; border-radius: 6px; background: var(--color-bg); border: 1px solid var(--color-border); font-size: 13px; cursor: pointer; }
.linked-card-row:hover { border-color: var(--color-primary); }
.linked-card-ref { font-size: 11px; font-weight: 600; font-family: monospace; color: var(--color-primary); background: color-mix(in srgb, var(--color-primary) 12%, transparent); padding: 1px 5px; border-radius: 4px; white-space: nowrap; }
.linked-card-title { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.linked-card-col { font-size: 11px; color: var(--color-text-muted); white-space: nowrap; }
.linked-card-project { font-size: 11px; color: var(--color-text-muted); font-style: italic; white-space: nowrap; }
.ref-closed .linked-card-title, .ref-closed.linked-card-ref, .ref-closed.linked-card-title { text-decoration: line-through; opacity: .6; }
.linked-card-add-row { display: flex; gap: 8px; margin-top: 4px; }
.linked-card-new-input { flex: 1; font-size: 13px; }

.card-detail.readonly select,
.card-detail.readonly input,
.card-detail.readonly textarea {
  pointer-events: none;
  opacity: 0.6;
  background: var(--color-surface-alt, transparent);
  cursor: default;
}
.card-detail.readonly .btn-icon-xs,
.card-detail.readonly .label-chip,
.card-detail.readonly .watcher-chip,
.card-detail.readonly .tag-remove,
.card-detail.readonly .checklist-checkbox,
.card-detail.readonly .subcard-check,
.card-detail.readonly .checklist-drag-handle {
  pointer-events: none;
  opacity: 0.5;
}
</style>
