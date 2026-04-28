    <template>
  <div class="dm-layout">

    <!-- ── Sidebar ──────────────────────────────────────── -->
    <aside class="dm-sidebar">

      <div class="dm-sidebar-header">
        <h2>{{ $t('dm.title') }}</h2>
        <button class="new-chat-btn" @click="toggleNewConv" :class="{ active: showNewConv }" title="New conversation">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        </button>
      </div>

      <!-- New conversation panel: multi-select users -->
      <div v-if="showNewConv" class="new-conv-panel">

        <!-- Tab bar: People / Teams -->
        <div class="new-conv-tabs">
          <button :class="['new-conv-tab', { active: newConvTab === 'people' }]" @click="newConvTab = 'people'">
            {{ $t('dm.tab_people') }}
          </button>
          <button :class="['new-conv-tab', { active: newConvTab === 'teams' }]" @click="newConvTab = 'teams'">
            {{ $t('dm.tab_teams') }}
          </button>
        </div>

        <!-- ── People tab ── -->
        <template v-if="newConvTab === 'people'">
          <div class="search-wrap">
            <svg class="search-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
            <input
              class="search-input"
              v-model="userSearch"
              :placeholder="$t('dm.search_users')"
              @input="filterUsers"
              autofocus
            />
          </div>

          <!-- Selected user chips -->
          <div v-if="selectedUsers.length" class="selected-chips">
            <span v-for="u in selectedUsers" :key="u.id" class="chip">
              {{ u.display_name || u.username }}
              <button class="chip-remove" @click="toggleUser(u)">×</button>
            </span>
          </div>

          <!-- Group name field (only when 2+ users selected) -->
          <input
            v-if="selectedUsers.length > 1"
            class="group-name-input"
            v-model="newGroupName"
            :placeholder="$t('dm.group_name_placeholder')"
          />

          <div class="user-search-results">
            <div
              v-for="u in filteredUsers"
              :key="u.id"
              :class="['user-result', { selected: isSelected(u) }]"
              @click="toggleUser(u)"
            >
              <div class="conv-avatar" :style="avatarBg(u)">
                <img v-if="getAvatar(u)" :src="getAvatar(u)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(u) }}</span>
              </div>
              <div class="conv-info">
                <div class="conv-name">{{ u.display_name || u.username }}</div>
                <div class="conv-handle">@{{ u.username }}</div>
              </div>
              <svg v-if="isSelected(u)" class="check-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
            </div>
            <div v-if="!filteredUsers.length" class="search-empty">{{ $t('dm.no_users_found') }}</div>
          </div>

          <button
            class="start-conv-btn"
            :disabled="!selectedUsers.length"
            @click="startConversation"
          >
            {{ selectedUsers.length > 1 ? $t('dm.start_group_chat') : $t('dm.open_chat') }}
          </button>
        </template>

        <!-- ── Teams tab ── -->
        <template v-else>
          <div v-if="loadingTeams" class="search-empty" style="padding:16px 12px">{{ $t('common.loading') }}</div>
          <div v-else class="user-search-results">
            <div
              v-for="p in allProjects"
              :key="p.id"
              class="user-result team-result"
              @click="selectProjectTeam(p)"
            >
              <div class="team-dot" :style="{ background: p.color || '#94a3b8' }"></div>
              <div class="conv-info">
                <div class="conv-name">{{ p.name }}</div>
                <div class="conv-handle">{{ $t('dm.team_members_count', { count: p.member_count ?? '…' }) }}</div>
              </div>
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="team-arrow"><polyline points="9 18 15 12 9 6"/></svg>
            </div>
            <div v-if="!allProjects.length" class="search-empty">{{ $t('dm.no_teams') }}</div>
          </div>
        </template>

      </div>

      <!-- Conversation list -->
      <div class="conv-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          :class="['conv-item', { active: activeConvId === conv.id }]"
          @click="openConversation(conv)"
        >
          <!-- Avatar: stacked for groups, single for 1-on-1 -->
          <div class="conv-avatar-wrap">
            <template v-if="conv.is_group">
              <div class="group-avatar">
                <img v-if="conv.avatar" :src="conv.avatar" class="avatar-img" @error="e => e.target.style.display='none'" />
                <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
              </div>
            </template>
            <template v-else>
              <div class="conv-avatar" :style="avatarBg(otherMember(conv))">
                <img v-if="getAvatar(otherMember(conv))" :src="getAvatar(otherMember(conv))" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(otherMember(conv)) }}</span>
              </div>
            </template>
          </div>

          <div class="conv-info">
            <div class="conv-name">{{ convDisplayName(conv) }}</div>
            <div class="conv-handle">
              {{ conv.is_group ? memberList(conv) : ('@' + (otherMember(conv)?.username || '')) }}
            </div>
          </div>
          <div v-if="activeConvId === conv.id" class="conv-active-dot"></div>
        </div>

        <div v-if="!conversations.length && !showNewConv" class="conv-empty">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          <p>{{ $t('dm.no_conversations') }}</p>
          <button class="btn btn-primary btn-sm" @click="showNewConv = true">Start a chat</button>
        </div>
      </div>

    </aside>

    <!-- ── Chat main area ────────────────────────────────── -->
    <main class="dm-main">

      <!-- Empty state -->
      <div v-if="!activeConv" class="dm-empty">
        <div class="dm-empty-icon">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        </div>
        <h3>{{ $t('dm.select_conversation') }}</h3>
        <p>Choose a conversation from the list or start a new one.</p>
      </div>

      <template v-else>

        <!-- Chat header -->
        <div class="dm-chat-header">
          <div class="conv-avatar-wrap">
            <template v-if="activeConv.is_group">
              <div class="group-avatar group-avatar-md group-avatar-upload" @click="triggerAvatarUpload" title="Change group avatar">
                <img v-if="activeConv.avatar" :src="activeConv.avatar" class="avatar-img" @error="e => e.target.style.display='none'" />
                <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                <div class="avatar-upload-overlay">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                </div>
              </div>
              <input ref="avatarInputEl" type="file" accept="image/*" class="hidden-input" @change="onAvatarSelected" />
            </template>
            <template v-else>
              <div class="conv-avatar conv-avatar-md" :style="avatarBg(otherMember(activeConv))">
                <img v-if="getAvatar(otherMember(activeConv))" :src="getAvatar(otherMember(activeConv))" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(otherMember(activeConv)) }}</span>
              </div>
            </template>
          </div>
          <div class="dm-header-info">
            <div class="dm-header-name">{{ convDisplayName(activeConv) }}</div>
            <div class="dm-header-handle">
              <template v-if="activeConv.is_group">
                <span v-for="m in activeConv.members" :key="m.user_id" class="member-chip">
                  {{ m.user?.display_name || m.user?.username }}
                  <button
                    v-if="m.user_id !== auth.user?.id"
                    class="chip-remove chip-remove-sm"
                    @click.stop="removeMember(m)"
                    :title="$t('dm.remove_member')"
                  >×</button>
                </span>
              </template>
              <template v-else>{{ memberList(activeConv) }}</template>
            </div>
          </div>
          <!-- Layout picker -->
          <div class="layout-picker">
            <button v-for="l in ['bubble','comfortable','compact','cozy','grouped']" :key="l"
              :class="['layout-btn', { active: layout === l }]"
              @click="setLayout(l)" :title="l.charAt(0).toUpperCase() + l.slice(1)">
              <!-- bubble -->
              <svg v-if="l === 'bubble'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><path d="M3 9a2 2 0 0 1 2-2h14"/></svg>
              <!-- comfortable -->
              <svg v-else-if="l === 'comfortable'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="4" cy="6" r="1.5" fill="currentColor" stroke="none"/><line x1="8" y1="6" x2="21" y2="6"/><circle cx="4" cy="12" r="1.5" fill="currentColor" stroke="none"/><line x1="8" y1="12" x2="21" y2="12"/><circle cx="4" cy="18" r="1.5" fill="currentColor" stroke="none"/><line x1="8" y1="18" x2="21" y2="18"/></svg>
              <!-- compact -->
              <svg v-else-if="l === 'compact'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="5" x2="21" y2="5"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="3" y1="13" x2="21" y2="13"/><line x1="3" y1="17" x2="21" y2="17"/><line x1="3" y1="21" x2="21" y2="21"/></svg>
              <!-- cozy -->
              <svg v-else-if="l === 'cozy'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="4" rx="1"/><rect x="3" y="10" width="18" height="4" rx="1"/><rect x="3" y="16" width="18" height="4" rx="1"/></svg>
              <!-- grouped -->
              <svg v-else-if="l === 'grouped'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="3" cy="5" r="2.5" fill="currentColor" stroke="none"/><line x1="8" y1="4" x2="21" y2="4"/><line x1="8" y1="8" x2="17" y2="8"/><line x1="8" y1="12" x2="19" y2="12"/><circle cx="3" cy="18" r="2.5" fill="currentColor" stroke="none"/><line x1="8" y1="17" x2="21" y2="17"/><line x1="8" y1="21" x2="15" y2="21"/></svg>
            </button>
            <!-- Bell: toggle new-message notifications -->
            <button :class="['layout-btn', { active: notifyEnabled }]"
              @click="toggleNotify"
              :title="notifyEnabled ? 'Mute notifications' : 'Enable notifications'">
              <svg v-if="notifyEnabled" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13.73 21a2 2 0 0 1-3.46 0"/><path d="M18.63 13A17.89 17.89 0 0 1 18 8"/><path d="M6.26 6.26A5.86 5.86 0 0 0 6 8c0 7-3 9-3 9h14"/><path d="M18 8a6 6 0 0 0-9.33-5"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
            </button>
          </div>
          <!-- Add member button for group chats -->
          <button v-if="activeConv.is_group" class="add-member-btn" @click="showAddMember = !showAddMember" title="Add member">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/></svg>
          </button>
        </div>

        <!-- Add member dropdown -->
        <div v-if="showAddMember" class="add-member-panel">
          <div class="search-wrap">
            <input class="search-input" v-model="addMemberSearch" placeholder="Search users…" @input="filterAddMembers" autofocus />
          </div>
          <div class="user-search-results">
            <div v-for="u in filteredAddMembers" :key="u.id" class="user-result" @click="addMember(u)">
              <div class="conv-avatar" :style="avatarBg(u)">
                <img v-if="getAvatar(u)" :src="getAvatar(u)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(u) }}</span>
              </div>
              <div class="conv-info">
                <div class="conv-name">{{ u.display_name || u.username }}</div>
              </div>
            </div>
            <div v-if="!filteredAddMembers.length" class="search-empty">No users to add</div>
          </div>
        </div>

        <!-- Messages -->
        <div class="dm-messages" :class="'layout-' + layout" ref="messagesEl" @click="handleCardRefClick" @auxclick="handleCardRefClick">

          <template v-for="(msg, i) in messages" :key="msg.id">

            <div v-if="isDifferentDay(messages, i)" class="date-sep">
              <span class="date-sep-label">{{ dayLabel(msg.created_at) }}</span>
            </div>

            <div :class="['msg-row', {
              'msg-own': msg.sender_id === auth.user?.id,
              'group-start': layout === 'grouped' && !isSameGroup(messages, i),
              'group-continue': layout === 'grouped' && isSameGroup(messages, i),
            }]"
              @mouseenter="hoveredReactionMessageId = msg.id"
              @mouseleave="onMessageHoverLeave(msg.id)"
            >

              <div class="msg-avatar" :style="avatarBg(msg.sender)">
                <img v-if="getAvatar(msg.sender)" :src="getAvatar(msg.sender)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials">{{ initials(msg.sender) }}</span>
              </div>

              <div class="msg-content">
                <div class="msg-sender" v-if="layout === 'grouped' ? !isSameGroup(messages, i) : (layout !== 'bubble' || msg.sender_id !== auth.user?.id)">
                  {{ msg.sender?.display_name || msg.sender?.username }}
                  <span v-if="layout === 'grouped'" class="msg-time-grouped">{{ formatTime(msg.created_at) }}</span>
                </div>
                <!-- Edit mode -->
                <template v-if="editingMsgId === msg.id">
                  <div class="edit-textarea-wrap" style="position:relative; width: 100%;">
                    <InlineEmojiPicker
                      v-if="editEmojiOpen"
                      :initial-search="editEmojiQuery || ''"
                      @pick="onEditEmojiPick"
                      @escape="onEditEmojiEscape"
                      @close="editEmojiOpen = false"
                    />
                    <MentionDropdown
                      v-if="editMentionUsers.length"
                      :users="editMentionUsers"
                      :active-index="editMentionIndex"
                      @pick="pickEditMention"
                      @update:activeIndex="editMentionIndex = $event"
                    />
                    <textarea class="edit-textarea" v-model="editBody" rows="2" ref="editTextareaEl" spellcheck="true" :lang="auth.user?.locale || 'en'" @keydown.enter.exact="onEditEnter($event, msg)" @keydown="onEditKeydown" @input="onEditInput"></textarea>
                  </div>
                  <div class="edit-actions">
                    <button class="btn btn-primary btn-sm" @click="saveEdit(msg)">Save</button>
                    <button class="btn btn-ghost btn-sm" @click="editingMsgId = null">Cancel</button>
                  </div>
                </template>

                <template v-else>
                  <div
                    v-if="canUseHoverReactions(msg)"
                    class="msg-hover-actions"
                    :class="{ visible: isHoverReactionsVisible(msg.id) }"
                  >
                    <button
                      v-for="emoji in quickReactionEmojis"
                      :key="`${msg.id}-${emoji}`"
                      class="msg-hover-emoji-btn"
                      type="button"
                      @click.stop="toggleConvReaction(msg, emoji)"
                    >
                      {{ emoji }}
                    </button>
                    <button
                      class="msg-hover-more-btn"
                      type="button"
                      title="More reactions"
                      @click.stop="toggleReactionPicker(msg.id)"
                    >
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <circle cx="12" cy="12" r="10"/>
                        <line x1="12" y1="8" x2="12" y2="16"/>
                        <line x1="8" y1="12" x2="16" y2="12"/>
                      </svg>
                    </button>
                    <InlineEmojiPicker
                      v-if="reactionPickerMessageId === msg.id"
                      :initial-search="''"
                      @pick="(emoji) => onHoverReactionPick(msg, emoji)"
                      @escape="reactionPickerMessageId = null"
                      @close="reactionPickerMessageId = null"
                    />
                  </div>
                  <div :class="['msg-bubble', msg.sender_id === auth.user?.id ? 'bubble-own' : 'bubble-other']">
                    <span v-if="msg.is_deleted" class="msg-deleted">{{ $t('chat.deleted') }}</span>
                    <!-- nosemgrep: javascript.vue.security.audit.xss.templates.avoid-v-html.avoid-v-html -- renderMarkdown sanitizes with DOMPurify -->
                    <div v-else class="msg-body" v-html="renderMarkdown(msg.body)"></div>
                    <span v-if="msg.is_edited && !msg.is_deleted" class="msg-edited"> · {{ $t('chat.edited') }}</span>
                  </div>
                  <AttachmentList v-if="!msg.is_deleted" :attachments="msg.attachments" />
                  <LinkPreviewCard v-if="!msg.is_deleted && firstUrl(msg.body)" :url="firstUrl(msg.body)" />
                  <MessageReactions
                    v-if="!msg.is_deleted"
                    :reactions="msg.reactions"
                    :users="allUsers"
                    @toggle="(emoji) => toggleConvReaction(msg, emoji)"
                  />
                  <div class="msg-meta">
                    <span class="msg-time">{{ formatTime(msg.created_at) }}</span>
                    <button
                      v-if="msg.sender_id === auth.user?.id && !msg.is_deleted"
                      class="msg-action-btn"
                      @click="startEdit(msg)"
                      title="Edit"
                    >
                      <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                    </button>
                    <button
                      v-if="msg.sender_id === auth.user?.id && !msg.is_deleted"
                      class="msg-action-btn"
                      @click="deleteMsg(msg)"
                      title="Delete"
                    >
                      <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    </button>
                  </div>
                </template>
              </div>

            </div>

          </template>

          <div v-if="!messages.length" class="messages-empty">
            <p>No messages yet. Say hello! 👋</p>
          </div>

        </div>

        <!-- New-message toast -->
        <Transition name="chat-toast">
          <div v-if="chatToast" class="chat-toast-popup" @click="chatToast = null">
            <strong>{{ chatToast.sender }}</strong>: {{ chatToast.body }}
          </div>
        </Transition>

        <!-- Compose -->
        <div class="dm-compose">
          <AttachmentList v-if="pendingFiles.length" :attachments="pendingFiles" :can-delete="true" @remove="removePending" />
          <div class="compose-outer" style="position:relative">
            <InlineEmojiPicker
              v-if="emojiOpen"
              :initial-search="emojiQuery || ''"
              @pick="onEmojiPick"
              @escape="onEmojiEscape"
              @close="emojiOpen = false"
            />
            <MentionDropdown
              v-if="mentionUsers.length"
              :users="mentionUsers"
              :active-index="mentionIndex"
              @pick="pickMention"
              @update:activeIndex="mentionIndex = $event"
            />
            <div class="compose-body">
              <div class="compose-avatar" :style="avatarBg(auth.user)">
                <img v-if="getAvatar(auth.user)" :src="getAvatar(auth.user)" class="avatar-img" @error="e => e.target.style.display='none'" />
                <span v-else class="avatar-initials avatar-initials-sm">{{ initials(auth.user) }}</span>
              </div>
              <FileUploadButton @files-selected="onFilesSelected" />
              <button class="emoji-trigger-btn" @click="emojiOpen = !emojiOpen" title="Emoji" type="button">😊</button>
              <textarea
                class="compose-textarea"
                v-model="newMessage"
                :placeholder="$t('chat.placeholder')"
                rows="1"
                :disabled="sending"
                ref="textareaEl"
                spellcheck="true"
                :lang="auth.user?.locale || 'en'"
                @keydown.enter.exact="onEnter"
                @keydown="onKeydown"
                @input="onInput"
                @paste="onPaste"
              ></textarea>
              <button class="compose-send-btn" @mousedown.prevent @click="send" :disabled="(!newMessage.trim() && !pendingFiles.length) || sending" :title="$t('chat.send')">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
              </button>
            </div>
          </div>
          <div class="compose-hint">Enter to send · Markdown · @mention</div>
        </div>

      </template>
    </main>

  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { messagesApi } from '@/api/messages'
import { projectsApi } from '@/api/projects'
import { attachmentsApi } from '@/api/attachments'
import { useNotificationsStore } from '@/stores/notifications'
import { useDateFormat } from '@/composables/useDateFormat'
import { useChatLayout } from '@/composables/useChatLayout'
import { useChatNotify } from '@/composables/useChatNotify'
import { avatarUrl } from '@/composables/useAvatar'
import { useCompose } from '@/composables/useCompose'
import { QUICK_REACTION_EMOJIS } from '@/utils/emoticons'
import { renderMarkdown, firstUrl, useCardRef } from '@/composables/useCardRef'
import InlineEmojiPicker from '@/components/common/InlineEmojiPicker.vue'
import MentionDropdown from '@/components/common/MentionDropdown.vue'
import AttachmentList from '@/components/common/AttachmentList.vue'
import FileUploadButton from '@/components/common/FileUploadButton.vue'
import MessageReactions from '@/components/common/MessageReactions.vue'
import LinkPreviewCard from '@/components/chat/LinkPreviewCard.vue'

const route = useRoute()
const { t } = useI18n()
const auth = useAuthStore()
const notificationsStore = useNotificationsStore()
const { formatTime } = useDateFormat()
const { layout, setLayout } = useChatLayout()
const { notifyEnabled, toggleNotify, desktopNotify, shouldNotifyNow } = useChatNotify()
const ui = useUIStore()

// ── New-message toast ────────────────────────────────────────────────────────
const chatToast = ref(null)
let toastTimer = null
const quickReactionEmojis = QUICK_REACTION_EMOJIS
const hoveredReactionMessageId = ref(null)
const reactionPickerMessageId = ref(null)

function showDMToast(msg) {
  if (!shouldNotifyNow()) return
  clearTimeout(toastTimer)
  const sender = msg.sender?.display_name || msg.sender?.username || 'Someone'
  const body = (msg.body || '').replace(/```[\s\S]*?```|`[^`]+`/g, '[code]').slice(0, 90)
  chatToast.value = { sender, body }
  toastTimer = setTimeout(() => { chatToast.value = null }, 4500)
  desktopNotify(sender, body)
}

let lastSeenMsgId = 0

const conversations = ref([])
const allUsers = ref([])
const filteredUsers = ref([])
const userSearch = ref('')
const showNewConv = ref(false)
const newConvTab = ref('people')
const selectedUsers = ref([])
const newGroupName = ref('')
const allProjects = ref([])
const loadingTeams = ref(false)

const activeConv = ref(null)
const activeConvId = ref(null)
const messages = ref([])
const newMessage = ref('')
const sending = ref(false)
const messagesEl = ref(null)
const textareaEl = ref(null)
let pollTimer = null

// Emoji + mention
const emojiOpen = ref(false)

const { mentionUsers, mentionIndex, onTextareaInput, onTextareaKeydown, pickMention, emojiQuery, pickEmoji } = useCompose({
  textareaEl,
  getValue: () => newMessage.value,
  setValue: (v) => { newMessage.value = v },
  users: allUsers,
})

function onEmojiPick(emoji) {
  pickEmoji(emoji)
  emojiOpen.value = false
}

function onEmojiEscape() {
  emojiOpen.value = false
  nextTick(() => textareaEl.value?.focus())
}

watch(emojiQuery, (q) => {
  emojiOpen.value = q !== null
})

function onInput(e) {
  autoResize(e)
  onTextareaInput()
}

function onEnter(e) {
  if (mentionUsers.value.length || emojiOpen.value) {
    onTextareaKeydown(e)
  } else {
    e.preventDefault()
    send()
  }
}

function onKeydown(e) {
  if (e.key === 'Escape' && (mentionUsers.value.length || emojiOpen.value)) {
    onTextareaKeydown(e)
    return
  }
  if (e.key !== 'Enter') onTextareaKeydown(e)
}

// Edit state
const editingMsgId = ref(null)
const editBody = ref('')
const editTextareaEl = ref(null)
const editEmojiOpen = ref(false)
const {
  mentionUsers: editMentionUsers,
  mentionIndex: editMentionIndex,
  onTextareaInput: onEditTextareaInput,
  onTextareaKeydown: onEditTextareaKeydown,
  pickMention: pickEditMention,
  emojiQuery: editEmojiQuery,
  pickEmoji: pickEditEmoji,
} = useCompose({
  textareaEl: editTextareaEl,
  getValue: () => editBody.value,
  setValue: (v) => { editBody.value = v },
  users: allUsers,
})

watch(editEmojiQuery, (q) => {
  editEmojiOpen.value = q !== null
})

function onEditEmojiPick(emoji) {
  pickEditEmoji(emoji)
  editEmojiOpen.value = false
}

function onEditEmojiEscape() {
  editEmojiOpen.value = false
  nextTick(() => editTextareaEl.value?.focus())
}

function onEditInput() {
  onEditTextareaInput()
}

function onEditEnter(e, msg) {
  if (editMentionUsers.value.length || editEmojiOpen.value) {
    onEditTextareaKeydown(e)
  } else {
    e.preventDefault()
    saveEdit(msg)
  }
}

function onEditKeydown(e) {
  if (e.key === 'Escape') {
    if (editMentionUsers.value.length || editEmojiOpen.value) {
      onEditTextareaKeydown(e)
    } else {
      editingMsgId.value = null
    }
    return
  }
  if (e.key !== 'Enter') onEditTextareaKeydown(e)
}

// Pending file attachments
const pendingFiles = ref([])

// Add member panel
const showAddMember = ref(false)
const addMemberSearch = ref('')
const filteredAddMembers = ref([])

// Group avatar upload
const avatarInputEl = ref(null)
function triggerAvatarUpload() {
  avatarInputEl.value?.click()
}
async function onAvatarSelected(e) {
  const file = e.target.files?.[0]
  if (!file || !activeConv.value) return
  e.target.value = ''
  try {
    const fd = new FormData()
    fd.append('avatar', file)
    const { data } = await messagesApi.uploadAvatar(activeConv.value.id, fd)
    activeConv.value = { ...activeConv.value, avatar: data.avatar }
    const idx = conversations.value.findIndex(c => c.id === activeConv.value.id)
    if (idx !== -1) conversations.value[idx] = { ...conversations.value[idx], avatar: data.avatar }
  } catch {
    ui.error('Failed to upload avatar')
  }
}

onMounted(async () => {
  try {
    const [convRes, userRes, projRes] = await Promise.all([
      messagesApi.getConversations(),
      messagesApi.listUsers(),
      projectsApi.list()
    ])
    conversations.value = convRes.data || []
    allUsers.value = (userRes.data || []).filter(u => u.id !== auth.user?.id)
    filteredUsers.value = allUsers.value
    allProjects.value = projRes.data || []
  } catch {}

  const targetUserId = route.query.user ? Number(route.query.user) : null
  if (targetUserId) await openOrCreateDM(targetUserId)
  const targetConvId = route.query.conv ? Number(route.query.conv) : null
  if (targetConvId) {
    const conv = conversations.value.find(c => c.id === targetConvId)
    if (conv) openConversation(conv)
  }
})

// Sidebar click navigation
watch(() => route.query.user, async (userId) => {
  if (!userId) return
  const loads = []
  if (!allUsers.value.length) {
    loads.push(messagesApi.listUsers().then(({ data }) => {
      allUsers.value = (data || []).filter(u => u.id !== auth.user?.id)
      filteredUsers.value = allUsers.value
    }).catch(() => {}))
  }
  if (!conversations.value.length) {
    loads.push(messagesApi.getConversations().then(({ data }) => {
      conversations.value = data || []
    }).catch(() => {}))
  }
  if (loads.length) await Promise.all(loads)
  await openOrCreateDM(Number(userId))
})

watch(() => route.query.conv, async (convId) => {
  if (!convId) return
  if (!conversations.value.length) {
    const { data } = await messagesApi.getConversations().catch(() => ({ data: [] }))
    conversations.value = data || []
  }
  const conv = conversations.value.find(c => c.id === Number(convId))
  if (conv) openConversation(conv)
})

// Open an existing 1-on-1 with this user, or create one
async function openOrCreateDM(userId) {
  if (!userId) return
  const existing = conversations.value.find(c =>
    !c.is_group &&
    c.members?.some(m => m.user_id === userId)
  )
  if (existing) {
    await openConversation(existing)
    return
  }
  // Create a new 1-on-1 conversation
  try {
    const { data } = await messagesApi.createConversation({ user_ids: [userId] })
    if (!conversations.value.find(c => c.id === data.id)) {
      conversations.value.unshift(data)
    }
    await openConversation(data)
  } catch {
    ui.error('Could not open conversation')
  }
}

function toggleNewConv() {
  showNewConv.value = !showNewConv.value
  if (!showNewConv.value) {
    selectedUsers.value = []
    newGroupName.value = ''
    userSearch.value = ''
    newConvTab.value = 'people'
    filteredUsers.value = allUsers.value
  }
}

async function selectProjectTeam(project) {
  loadingTeams.value = true
  try {
    const { data } = await projectsApi.listMembers(project.slug)
    const members = (data || [])
      .filter(m => m.user_id !== auth.user?.id)
      .map(m => m.user)
      .filter(Boolean)
    selectedUsers.value = members
    newGroupName.value = project.name
    newConvTab.value = 'people'
    userSearch.value = ''
    filteredUsers.value = allUsers.value
  } catch {
    ui.error('Failed to load project members')
  } finally {
    loadingTeams.value = false
  }
}

function filterUsers() {
  const q = userSearch.value.toLowerCase()
  filteredUsers.value = allUsers.value.filter(u =>
    u.username.toLowerCase().includes(q) ||
    (u.display_name || '').toLowerCase().includes(q)
  )
}

function isSelected(u) {
  return selectedUsers.value.some(s => s.id === u.id)
}

function toggleUser(u) {
  if (isSelected(u)) {
    selectedUsers.value = selectedUsers.value.filter(s => s.id !== u.id)
  } else {
    selectedUsers.value.push(u)
  }
}

async function startConversation() {
  if (!selectedUsers.value.length) return
  const userIds = selectedUsers.value.map(u => u.id)
  try {
    const { data } = await messagesApi.createConversation({
      user_ids: userIds,
      name: newGroupName.value.trim() || ''
    })
    if (!conversations.value.find(c => c.id === data.id)) {
      conversations.value.unshift(data)
    }
    showNewConv.value = false
    selectedUsers.value = []
    newGroupName.value = ''
    userSearch.value = ''
    filteredUsers.value = allUsers.value
    await openConversation(data)
  } catch {
    ui.error('Could not create conversation')
  }
}

async function openConversation(conv) {
  activeConv.value = conv
  activeConvId.value = conv.id
  showAddMember.value = false
  lastSeenMsgId = 0
  notificationsStore.markConvSeen(conv.id)
  // Stop polling the previous conversation
  clearInterval(pollTimer)
  await fetchMessages(true)
  // Poll for new messages from other participants every 5 s
  pollTimer = setInterval(fetchMessages, 5_000)
}

async function fetchMessages(initial = false) {
  if (!activeConvId.value) return
  try {
    const { data } = await messagesApi.getMessages(activeConvId.value)
    const incoming = data || []
    const atBottom = initial || isAtBottom()
    if (incoming.length > 0) {
      const maxId = Math.max(...incoming.map(m => m.id))
      if (!initial && lastSeenMsgId > 0 && notifyEnabled.value && shouldNotifyNow()) {
        const newFromOthers = incoming.filter(m => m.id > lastSeenMsgId && m.sender_id !== auth.user?.id)
        if (newFromOthers.length > 0) showDMToast(newFromOthers[newFromOthers.length - 1])
      }
      lastSeenMsgId = maxId
    }
    messages.value = incoming
    if (atBottom) {
      await nextTick()
      scrollToBottom()
      notificationsStore.markConvSeen(activeConvId.value)
    }
  } catch {
    if (initial) ui.error('Failed to load messages')
  }
}

function isAtBottom() {
  if (!messagesEl.value) return true
  const el = messagesEl.value
  return el.scrollHeight - el.scrollTop - el.clientHeight < 60
}

onUnmounted(() => clearInterval(pollTimer))

async function send() {
  const body = newMessage.value.trim()
  if (!body && !pendingFiles.value.length || !activeConv.value) return
  sending.value = true
  try {
    const sendBody = body || '📎'
    const { data } = await messagesApi.sendConvMessage(activeConv.value.id, { body: sendBody })
    const newMsg = { ...data, attachments: [], reactions: [] }

    // Upload any pending files linked to this message
    if (pendingFiles.value.length) {
      const filesToUpload = [...pendingFiles.value]
      pendingFiles.value = []
      filesToUpload.forEach(pf => { if (pf._previewUrl) URL.revokeObjectURL(pf._previewUrl) })
      for (const pf of filesToUpload) {
        const fd = new FormData()
        fd.append('file', pf._file)
        fd.append('owner_type', 'conv_message')
        fd.append('owner_id', String(data.id))
        try {
          const { data: att } = await attachmentsApi.upload(fd)
          newMsg.attachments.push(att)
        } catch {}
      }
    }

    messages.value.push(newMsg)
    newMessage.value = ''
    if (textareaEl.value) textareaEl.value.style.height = 'auto'
    // Bump this conversation to the top
    const idx = conversations.value.findIndex(c => c.id === activeConv.value.id)
    if (idx > 0) {
      const [c] = conversations.value.splice(idx, 1)
      conversations.value.unshift(c)
    }
    // Sending counts as reading — don't show your own message as unread
    notificationsStore.markConvSeen(activeConv.value.id)
    await nextTick()
    scrollToBottom()
    // Refresh to pick up any concurrent messages from others
    await fetchMessages()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to send message')
  } finally {
    sending.value = false
    await nextTick()
    if (textareaEl.value) textareaEl.value.focus()
  }
}

async function deleteMsg(msg) {
  try {
    await messagesApi.deleteConvMessage(activeConv.value.id, msg.id)
    msg.is_deleted = true
  } catch {
    ui.error('Failed to delete message')
  }
}

function startEdit(msg) {
  editingMsgId.value = msg.id
  editBody.value = msg.body
}

async function saveEdit(msg) {
  if (!editBody.value.trim()) return
  try {
    await messagesApi.editConvMessage(activeConv.value.id, msg.id, editBody.value)
    msg.body = editBody.value
    msg.is_edited = true
    editingMsgId.value = null
  } catch {
    ui.error('Failed to edit message')
  }
}

async function toggleConvReaction(msg, emoji) {
  if (!activeConv.value) return
  try {
    const { data } = await messagesApi.toggleConvReaction(activeConv.value.id, msg.id, emoji)
    msg.reactions = data.reactions
  } catch {}
}

function canUseHoverReactions(msg) {
  return msg.sender_id !== auth.user?.id && !msg.is_deleted
}

function isHoverReactionsVisible(messageId) {
  return hoveredReactionMessageId.value === messageId || reactionPickerMessageId.value === messageId
}

function toggleReactionPicker(messageId) {
  hoveredReactionMessageId.value = messageId
  reactionPickerMessageId.value = reactionPickerMessageId.value === messageId ? null : messageId
}

async function onHoverReactionPick(msg, emoji) {
  await toggleConvReaction(msg, emoji)
  reactionPickerMessageId.value = null
}

function onMessageHoverLeave(messageId) {
  if (reactionPickerMessageId.value !== messageId) hoveredReactionMessageId.value = null
}

function onFilesSelected(files) {
  for (const f of files) {
    pendingFiles.value.push({
      id: Math.random(),
      filename: f.name,
      size_bytes: f.size,
      mime_type: f.type || 'application/octet-stream',
      _file: f,
      _previewUrl: f.type?.startsWith('image/') ? URL.createObjectURL(f) : null,
    })
  }
}

async function onPaste(e) {
  const items = Array.from(e.clipboardData?.items || [])
  const images = items.filter(it => it.kind === 'file' && it.type.startsWith('image/'))
  if (images.length) {
    e.preventDefault()
    onFilesSelected(images.map(it => it.getAsFile()).filter(Boolean))
    return
  }
  if (window.__TAURI_INTERNALS__ && navigator.clipboard?.read) {
    try {
      const clipItems = await navigator.clipboard.read()
      const files = []
      for (const item of clipItems) {
        for (const type of item.types) {
          if (type.startsWith('image/')) {
            const blob = await item.getType(type)
            const ext = type.split('/')[1]?.split('+')[0] || 'png'
            files.push(new File([blob], `paste.${ext}`, { type }))
          }
        }
      }
      if (files.length) { e.preventDefault(); onFilesSelected(files) }
    } catch {}
  }
}

function removePending(a) {
  if (a._previewUrl) URL.revokeObjectURL(a._previewUrl)
  pendingFiles.value = pendingFiles.value.filter(p => p.id !== a.id)
}

// Add member to active group conversation
function filterAddMembers() {
  const q = addMemberSearch.value.toLowerCase()
  const memberIds = new Set(activeConv.value?.members?.map(m => m.user_id) || [])
  filteredAddMembers.value = allUsers.value.filter(u =>
    !memberIds.has(u.id) &&
    (u.username.toLowerCase().includes(q) || (u.display_name || '').toLowerCase().includes(q))
  )
}

watch(showAddMember, (v) => {
  if (v) {
    addMemberSearch.value = ''
    filterAddMembers()
  }
})

async function addMember(user) {
  try {
    await messagesApi.addMember(activeConv.value.id, { user_id: user.id })
    // Refresh conversation to get updated member list
    const { data } = await messagesApi.getConversations()
    conversations.value = data || []
    activeConv.value = conversations.value.find(c => c.id === activeConvId.value) || activeConv.value
    showAddMember.value = false
  } catch {
    ui.error('Failed to add member')
  }
}

async function removeMember(member) {
  if (!confirm(t('dm.remove_member_confirm'))) return
  try {
    const { data } = await messagesApi.removeMember(activeConv.value.id, member.user_id)
    if (data?.conversation_deleted) {
      // Conversation was auto-deleted (only creator left, no messages)
      conversations.value = conversations.value.filter(c => c.id !== activeConvId.value)
      activeConv.value = null
      activeConvId.value = null
      clearInterval(pollTimer)
      messages.value = []
    } else {
      // Refresh conversation to get updated member list
      const { data: convs } = await messagesApi.getConversations()
      conversations.value = convs || []
      activeConv.value = conversations.value.find(c => c.id === activeConvId.value) || activeConv.value
    }
  } catch {
    ui.error('Failed to remove member')
  }
}

function scrollToBottom() {
  if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
}


function autoResize(e) {
  const el = e.target
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
}

// ── Helpers ──────────────────────────────────────────────

function getAvatar(user) {
  return avatarUrl(user)
}

function initials(u) {
  if (!u) return '?'
  const name = u.display_name || u.username || '?'
  return name.slice(0, 2).toUpperCase()
}

const AVATAR_COLORS = ['#6366f1','#8b5cf6','#ec4899','#f59e0b','#10b981','#3b82f6','#ef4444']
function avatarBg(u) {
  const idx = (u?.username?.charCodeAt(0) || 0) % AVATAR_COLORS.length
  return { background: AVATAR_COLORS[idx] }
}

// Returns the other participant in a 1-on-1 conversation
function otherMember(conv) {
  if (!conv?.members) return null
  const m = conv.members.find(m => m.user_id !== auth.user?.id)
  return m?.user || null
}

// Human-readable conversation name
function convDisplayName(conv) {
  if (!conv) return ''
  if (conv.name) return conv.name
  if (conv.is_group) {
    return conv.members
      ?.filter(m => m.user_id !== auth.user?.id)
      .map(m => m.user?.display_name || m.user?.username)
      .join(', ') || 'Group'
  }
  const other = otherMember(conv)
  return other?.display_name || other?.username || 'Unknown'
}

// Short member list for subtitle
function memberList(conv) {
  if (!conv?.members) return ''
  return conv.members
    .filter(m => m.user_id !== auth.user?.id)
    .map(m => m.user?.display_name || m.user?.username)
    .join(', ')
}

// Date grouping
function isDifferentDay(msgs, index) {
  if (index === 0) return true
  const curr = new Date(msgs[index].created_at)
  const prev = new Date(msgs[index - 1].created_at)
  return curr.getFullYear() !== prev.getFullYear() ||
    curr.getMonth() !== prev.getMonth() ||
    curr.getDate() !== prev.getDate()
}

function isSameGroup(msgs, i) {
  if (i === 0) return false
  const curr = msgs[i]
  const prev = msgs[i - 1]
  if (curr.sender_id !== prev.sender_id) return false
  if (isDifferentDay(msgs, i)) return false
  return new Date(curr.created_at) - new Date(prev.created_at) < 5 * 60 * 1000
}

const { handleCardRefClick } = useCardRef()

function dayLabel(dateStr) {
  const d = new Date(dateStr)
  const now = new Date()
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  const sameDay = (a, b) =>
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  if (sameDay(d, now)) return 'Today'
  if (sameDay(d, yesterday)) return 'Yesterday'
  return d.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' })
}
</script>

<style scoped>
/* ── Layout ──────────────────────────────────────────────── */
.dm-layout { flex: 1; display: flex; overflow: hidden; height: 100%; }

/* ── Sidebar ─────────────────────────────────────────────── */
.dm-sidebar {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border);
  background: var(--color-surface);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dm-sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 54px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.dm-sidebar-header h2 { font-size: 15px; font-weight: 700; }

.new-chat-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  transition: all .15s;
}
.new-chat-btn:hover, .new-chat-btn.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
}

/* New conversation panel */
.new-conv-panel {
  border-bottom: 1px solid var(--color-border);
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.new-conv-tabs {
  display: flex;
  gap: 2px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 2px;
}
.new-conv-tab {
  flex: 1;
  font-size: 11px;
  font-weight: 600;
  padding: 4px 8px;
  border: none;
  border-radius: calc(var(--radius-sm) - 2px);
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all .15s;
}
.new-conv-tab.active {
  background: var(--color-surface);
  color: var(--color-text);
  box-shadow: 0 1px 2px rgba(0,0,0,.08);
}

.team-result { cursor: pointer; gap: 10px; }
.team-dot { width: 28px; height: 28px; border-radius: 50%; flex-shrink: 0; }
.team-arrow { color: var(--color-text-muted); flex-shrink: 0; }
.search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 9px;
  color: var(--color-text-muted);
  pointer-events: none;
}
.search-input {
  width: 100%;
  padding: 7px 10px 7px 28px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  color: var(--color-text);
}
.search-input:focus { border-color: var(--color-primary); }

/* Selected user chips */
.selected-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--color-primary);
  color: #fff;
  border-radius: 9999px;
  padding: 2px 8px 2px 10px;
  font-size: 12px;
  font-weight: 500;
}
.chip-remove {
  background: none;
  border: none;
  color: rgba(255,255,255,.8);
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
}
.chip-remove:hover { color: #fff; }

.member-chip {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: var(--color-surface-raised);
  border-radius: 10px;
  padding: 1px 6px;
  font-size: 11px;
  margin-right: 3px;
}
.chip-remove-sm {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  font-size: 13px;
  line-height: 1;
  padding: 0;
}
.chip-remove-sm:hover { color: var(--color-danger); }

.group-name-input {
  width: 100%;
  padding: 6px 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  color: var(--color-text);
  box-sizing: border-box;
}
.group-name-input:focus { border-color: var(--color-primary); }

.user-search-results { max-height: 180px; overflow-y: auto; }
.user-result {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 6px;
  border-radius: 8px;
  cursor: pointer;
  transition: background .1s;
}
.user-result:hover { background: var(--color-bg); }
.user-result.selected { background: color-mix(in srgb, var(--color-primary) 10%, transparent); }
.check-icon { color: var(--color-primary); flex-shrink: 0; margin-left: auto; }
.search-empty { padding: 12px 6px; font-size: 13px; color: var(--color-text-muted); text-align: center; }

.start-conv-btn {
  width: 100%;
  padding: 8px;
  background: var(--color-primary);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity .15s;
}
.start-conv-btn:disabled { opacity: .35; cursor: default; }
.start-conv-btn:not(:disabled):hover { opacity: .88; }

/* Conversation list */
.conv-list { flex: 1; overflow-y: auto; padding: 6px; }

.conv-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 10px;
  cursor: pointer;
  position: relative;
  transition: background .1s;
}
.conv-item:hover { background: var(--color-bg); }
.conv-item.active { background: color-mix(in srgb, var(--color-primary) 12%, transparent); }

.conv-active-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--color-primary);
  margin-left: auto;
  flex-shrink: 0;
}

/* Avatars */
.conv-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}
.conv-avatar-md { width: 42px; height: 42px; }
.conv-avatar-wrap { flex-shrink: 0; }

.group-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--color-text-muted);
}
.group-avatar-md { width: 42px; height: 42px; }
.group-avatar-upload { cursor: pointer; position: relative; }
.group-avatar-upload:hover .avatar-upload-overlay { opacity: 1; }
.avatar-upload-overlay {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: rgba(0,0,0,.45);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity .15s;
  color: #fff;
}
.hidden-input { display: none; }

.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-initials { color: #fff; font-size: 13px; font-weight: 700; }

.conv-info { flex: 1; min-width: 0; }
.conv-name { font-size: 14px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.conv-handle { font-size: 11px; color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.conv-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 10px;
  color: var(--color-text-muted);
  text-align: center;
}
.conv-empty p { font-size: 13px; margin: 0; }

/* ── Main chat area ───────────────────────────────────────── */
.dm-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
  position: relative;
}

.dm-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  gap: 10px;
  text-align: center;
  padding: 40px;
}
.dm-empty-icon {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}
.dm-empty h3 { font-size: 16px; font-weight: 600; margin: 0; color: var(--color-text); }
.dm-empty p { font-size: 13px; margin: 0; }

/* Chat header */
.dm-chat-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  height: 54px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-surface);
  flex-shrink: 0;
}
.dm-header-info { flex: 1; min-width: 0; }
.dm-header-name { font-size: 15px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dm-header-handle { font-size: 12px; color: var(--color-text-muted); display: flex; flex-wrap: wrap; gap: 3px; align-items: center; }

.add-member-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  flex-shrink: 0;
  transition: all .15s;
}
.add-member-btn:hover { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }

/* Add member panel */
.add-member-panel {
  border-bottom: 1px solid var(--color-border);
  padding: 10px 20px;
  background: var(--color-surface);
  flex-shrink: 0;
}
.add-member-panel .search-input { padding-left: 10px; }

/* Messages */
.dm-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.messages-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  font-size: 14px;
}

/* Date separator */
.date-sep {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 14px 0 10px;
}
.date-sep::before,
.date-sep::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--color-border);
}
.date-sep-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: .06em;
  white-space: nowrap;
  padding: 0 4px;
}

/* Message rows */
.msg-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 4px;
}
.msg-row.msg-own { flex-direction: row-reverse; }

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}
.msg-avatar .avatar-initials { color: #fff; font-size: 11px; font-weight: 700; }

.msg-content {
  display: flex;
  flex-direction: column;
  max-width: calc(100% - 48px);
  position: relative;
}
.msg-row.msg-own .msg-content { align-items: flex-end; }

.msg-sender {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  margin-bottom: 3px;
  padding: 0 4px;
}

.msg-bubble {
  padding: 9px 14px;
  border-radius: 18px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
  max-width: 100%;
  font-variant-emoji: emoji;
}
.bubble-other {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-bottom-left-radius: 4px;
  color: var(--color-text);
}
.bubble-own {
  background: var(--color-primary);
  color: #fff;
  border-bottom-right-radius: 4px;
}
.msg-deleted { font-style: italic; opacity: .6; }
.msg-edited { font-size: 11px; opacity: .7; }

/* Markdown body inside bubbles */
.msg-body { line-height: 1.5; }
.msg-body :deep(.card-ref-link) {
  display: inline-block;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  background: var(--color-primary);
  border-radius: 4px;
  padding: 1px 6px;
  cursor: pointer;
  text-decoration: none;
  white-space: nowrap;
  vertical-align: baseline;
}
.msg-body :deep(.card-ref-link:hover) { opacity: 0.8; }
.bubble-own .msg-body :deep(.card-ref-link) {
  color: var(--color-primary);
  background: #fff;
}
.msg-body :deep(p) { margin: 0 0 4px; }
.msg-body :deep(p:last-child) { margin-bottom: 0; }
.msg-body :deep(strong) { font-weight: 700; }
.msg-body :deep(em) { font-style: italic; }
.msg-body :deep(code) {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  background: rgba(0,0,0,.12);
  border-radius: 4px;
  padding: 1px 5px;
}
.bubble-own .msg-body :deep(code) {
  background: rgba(255,255,255,0.2);
  color: #ffffff;
  border: 1px solid rgba(255,255,255,0.28);
}
.msg-body :deep(pre) {
  background: rgba(0,0,0,.15);
  border-radius: 8px;
  padding: 10px 12px;
  overflow-x: auto;
  margin: 6px 0;
}
.bubble-own .msg-body :deep(pre) {
  background: #ffffff !important;
  color: #0b1220 !important;
  border: 1px solid rgba(0,0,0,0.12) !important;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08) !important;
  padding: 12px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, 'Roboto Mono', 'Courier New', monospace;
  overflow-x: auto;
}
.bubble-own .msg-body :deep(pre) .hljs,
.bubble-own .msg-body :deep(pre) .hljs * {
  background: transparent !important;
  box-shadow: none !important;
  opacity: 1 !important;
  text-shadow: none !important;
}
.bubble-own .msg-body :deep(pre) .hljs { color: #0f172a !important; }
.bubble-own .msg-body :deep(pre) .hljs-comment,
.bubble-own .msg-body :deep(pre) .hljs-quote { color: #64748b !important; font-style: italic; }
.bubble-own .msg-body :deep(pre) .hljs-keyword,
.bubble-own .msg-body :deep(pre) .hljs-selector-tag,
.bubble-own .msg-body :deep(pre) .hljs-literal,
.bubble-own .msg-body :deep(pre) .hljs-type { color: #7c3aed !important; }
.bubble-own .msg-body :deep(pre) .hljs-string,
.bubble-own .msg-body :deep(pre) .hljs-doctag,
.bubble-own .msg-body :deep(pre) .hljs-regexp { color: #15803d !important; }
.bubble-own .msg-body :deep(pre) .hljs-number,
.bubble-own .msg-body :deep(pre) .hljs-symbol,
.bubble-own .msg-body :deep(pre) .hljs-bullet { color: #1d4ed8 !important; }
.bubble-own .msg-body :deep(pre) .hljs-title,
.bubble-own .msg-body :deep(pre) .hljs-section,
.bubble-own .msg-body :deep(pre) .hljs-function .hljs-title { color: #0f766e !important; }
.msg-body :deep(pre code) { background: none; padding: 0; font-size: 12px; }
/* Prevent per-token/line backgrounds inside code blocks */
.msg-body :deep(pre) :deep(*) { background: transparent !important; box-shadow: none !important; border-radius: 0 !important; }
.msg-body :deep(pre .hljs-addition), .msg-body :deep(pre .hljs-deletion) { background: transparent !important; }
.msg-body :deep(ul), .msg-body :deep(ol) { margin: 4px 0; padding-left: 18px; }
.msg-body :deep(li) { margin: 2px 0; }
.msg-body :deep(blockquote) {
  border-left: 3px solid rgba(0,0,0,.2);
  margin: 4px 0;
  padding: 2px 10px;
  opacity: .85;
}
.bubble-own .msg-body :deep(blockquote) { border-left-color: rgba(255,255,255,.4); }
.msg-body :deep(a) { color: inherit; text-decoration: underline; }
.msg-body :deep(h1), .msg-body :deep(h2), .msg-body :deep(h3) { font-size: 1em; font-weight: 700; margin: 4px 0; }
.msg-body :deep(hr) { border: none; border-top: 1px solid rgba(0,0,0,.15); margin: 6px 0; }
.bubble-own .msg-body :deep(hr) { border-top-color: rgba(255,255,255,.3); }

.msg-hover-actions {
  position: absolute;
  top: -18px;
  right: 2px;
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 2px 5px;
  border: 1px solid var(--color-border);
  border-radius: 999px;
  background: var(--color-surface);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.12);
  opacity: 0;
  transform: translateY(4px);
  pointer-events: none;
  transition: opacity .15s, transform .15s;
  z-index: 30;
}
.msg-hover-actions.visible {
  opacity: 1;
  transform: translateY(0);
  pointer-events: auto;
}
.msg-hover-emoji-btn,
.msg-hover-more-btn {
  border: none;
  background: none;
  cursor: pointer;
  border-radius: 999px;
  width: 24px;
  height: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text);
}
.msg-hover-emoji-btn:hover,
.msg-hover-more-btn:hover {
  background: var(--color-bg);
}
.msg-hover-emoji-btn { font-size: 14px; }

.msg-time {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 3px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.msg-meta {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 3px;
}
.msg-action-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 1px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity .15s;
}
.msg-meta:hover .msg-action-btn { opacity: 1; }
.msg-action-btn:hover { color: var(--color-danger); background: var(--color-bg); }

/* Edit inline */
.edit-textarea {
  width: 100%;
  border: 1px solid var(--color-primary);
  border-radius: 8px;
  padding: 6px 10px;
  font-size: 13px;
  background: var(--color-bg);
  color: var(--color-text);
  resize: none;
  outline: none;
  font-family: inherit;
}
.edit-actions {
  display: flex;
  gap: 6px;
  margin-top: 4px;
}

/* Compose */
.dm-compose {
  border-top: 1px solid var(--color-border);
  padding: 10px 20px 8px;
  flex-shrink: 0;
  background: var(--color-surface);
}
.compose-body {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 14px;
  padding: 7px 10px 7px 12px;
  transition: border-color .15s;
}
.compose-body:focus-within { border-color: var(--color-primary); }

.compose-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2px;
}
.avatar-initials-sm { color: #fff; font-size: 9px; font-weight: 700; }

.compose-textarea {
  flex: 1;
  border: none;
  background: transparent;
  resize: none;
  outline: none;
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text);
  font-family: inherit;
  padding: 2px 0;
  min-height: 24px;
  max-height: 120px;
  overflow-y: auto;
}
.compose-textarea::placeholder { color: var(--color-text-muted); }

.compose-send-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: var(--color-primary);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: opacity .15s;
}
.compose-send-btn:disabled { opacity: .35; cursor: default; }
.compose-send-btn:not(:disabled):hover { opacity: .85; }

.compose-hint {
  font-size: 10px;
  color: var(--color-text-muted);
  margin-top: 5px;
  text-align: right;
  padding-right: 2px;
}

.emoji-trigger-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 2px 3px;
  border-radius: 5px;
  line-height: 1;
  flex-shrink: 0;
  opacity: .55;
  transition: opacity .1s;
  margin-bottom: 2px;
}
.emoji-trigger-btn:hover { opacity: 1; }

.compose-outer { position: relative; }

/* ── New-message toast ───────────────────────────────────── */
.chat-toast-popup {
  position: absolute;
  bottom: 80px;
  left: 20px;
  right: 20px;
  max-width: 480px;
  margin: 0 auto;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 8px 16px;
  font-size: 13px;
  box-shadow: 0 4px 16px rgba(0,0,0,.15);
  cursor: pointer;
  z-index: 20;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.chat-toast-enter-from, .chat-toast-leave-to { opacity: 0; transform: translateY(8px); }
.chat-toast-enter-active, .chat-toast-leave-active { transition: opacity .25s, transform .25s; }

/* ── Layout picker ───────────────────────────────────────── */
.layout-picker {
  display: flex;
  gap: 2px;
  margin-left: auto;
  margin-right: 8px;
}
.layout-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 4px 6px;
  border-radius: 5px;
  display: flex;
  align-items: center;
  transition: background .15s, color .15s;
}
.layout-btn:hover { background: var(--color-bg); color: var(--color-text); }
.layout-btn.active { background: var(--color-primary); color: #fff; }

/* ── Comfortable: Slack-style, all left, same bubble ──────── */
.layout-comfortable .msg-row.msg-own { flex-direction: row; }
.layout-comfortable .msg-row.msg-own .msg-content { align-items: flex-start; }
.layout-comfortable .bubble-own {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 18px;
  border-bottom-left-radius: 4px;
  color: var(--color-text);
}
.layout-comfortable .bubble-own .msg-body :deep(code) { background: rgba(0,0,0,.08); }
.layout-comfortable .bubble-own .msg-body :deep(pre) { background: rgba(0,0,0,.08); }

/* ── Compact: IRC/terminal, no avatar, inline ────────────── */
.layout-compact .msg-avatar { display: none; }
.layout-compact .msg-row { margin-bottom: 1px; align-items: baseline; }
.layout-compact .msg-row.msg-own { flex-direction: row; }
.layout-compact .msg-content {
  flex-direction: row;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0 5px;
  max-width: 100%;
}
.layout-compact .msg-row.msg-own .msg-content { align-items: baseline; }
.layout-compact .msg-sender {
  order: 1;
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text);
  margin: 0;
  padding: 0;
  flex-shrink: 0;
}
.layout-compact .msg-sender::after { content: ':'; }
.layout-compact .msg-bubble {
  order: 2;
  background: transparent !important;
  border: none !important;
  border-radius: 0 !important;
  padding: 0 !important;
  color: var(--color-text) !important;
  font-size: 13px;
  max-width: none;
}
.layout-compact .msg-meta {
  order: 0;
  margin-top: 0;
  padding: 0;
  flex-shrink: 0;
  min-width: 36px;
}
.layout-compact .msg-time { font-size: 11px; }
.layout-compact .msg-action-btn { display: none; }

/* ── Cozy: document-style, left-border accent for own ────── */
.layout-cozy .msg-row.msg-own {
  flex-direction: row;
  border-left: 3px solid var(--color-primary);
  padding-left: 10px;
  margin-left: -13px;
  border-radius: 0 4px 4px 0;
}
.layout-cozy .msg-row.msg-own .msg-content { align-items: flex-start; }
.layout-cozy .bubble-own {
  background: transparent;
  border: none;
  border-radius: 0;
  padding: 2px 0;
  color: var(--color-text);
}
.layout-cozy .bubble-other {
  background: transparent;
  border: none;
  border-radius: 0;
  padding: 2px 0;
}
.layout-cozy .bubble-own .msg-body :deep(code),
.layout-cozy .bubble-other .msg-body :deep(code) { background: rgba(0,0,0,.08); }

/* ── Grouped: Discord/Mattermost-style ───────────────────── */
.layout-grouped .msg-row {
  flex-direction: row !important;
  align-items: flex-start;
  margin-bottom: 0;
  padding: 1px 0;
}
.layout-grouped .msg-row.msg-own .msg-content { align-items: flex-start; }
.layout-grouped .msg-row.group-start { margin-top: 14px; }
.layout-grouped .msg-row.group-continue .msg-avatar { visibility: hidden; }
.layout-grouped .msg-avatar { align-self: flex-start; }
.layout-grouped .msg-sender {
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 2px;
  padding: 0;
}
.msg-time-grouped {
  font-size: 11px;
  font-weight: 400;
  color: var(--color-text-muted);
}
.layout-grouped .msg-bubble {
  background: transparent !important;
  border: none !important;
  border-radius: 0 !important;
  padding: 2px 0 !important;
  color: var(--color-text) !important;
  max-width: 100%;
}
.layout-grouped .msg-body :deep(pre) {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
}
.layout-grouped .msg-body :deep(code) { background: var(--color-bg); }
.layout-grouped .msg-time { display: none; }
.layout-grouped .msg-meta { margin-top: 0; }
</style>
