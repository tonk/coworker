<template>
  <Transition name="call-ui">
    <div v-if="showWebRTC || showLiveKit">

      <!-- ── LiveKit: connecting bar ───────────────────────────────────────── -->
      <CallChatSidebar
        v-if="showCallChat && activeCallConvId"
        :conversation-id="activeCallConvId"
        fixed
        @close="setCallChat(false)"
      />
      <div v-if="lkState.phase === 'connecting'" class="active-call-bar lk-connecting-bar" :class="{ 'bar-with-chat': showCallChat }" :style="showCallChat ? { right: callChatWidth + 'px' } : null">
        <div class="call-bar-info">
          <div class="call-bar-name">{{ lkState.title || $t('call.group_call') }}</div>
          <div class="call-bar-status"><span class="status-calling">{{ $t('call.joining_group') }}</span></div>
        </div>
      </div>

      <!-- ── LiveKit: active (grid) ───────────────────────────────────────── -->
      <div v-else-if="lkState.phase === 'active'" class="call-video-chat-row" :class="{ 'is-windowed': isWindowed }" :style="windowStyle">
        <div class="call-stage lk-grid-overlay" @click.self="closeInvitePicker">
        <div
          v-if="isWindowed"
          class="call-window-drag-handle"
          role="button"
          tabindex="0"
          :aria-label="$t('call.move_window')"
          @pointerdown="onWindowDragStart"
          @keydown="onWindowDragHandleKeydown"
        ></div>
        <div class="lk-grid-top">
          <div class="remote-name">{{ lkState.title || $t('call.group_call') }}</div>
          <div class="call-duration">{{ formattedLkDuration }}</div>
        </div>
        <div class="lk-grid">
          <LkVideoCell
            v-for="t in lkTiles"
            :key="t.sid"
            :track="t.videoTrack"
            :label="tileLabel(t)"
            :avatar="t.avatar"
            :is-local="t.isLocal"
            :camera-off="t.cameraOff"
            :is-screen-share="t.isScreenShare || false"
          />
        </div>

        <!-- ── Invite picker panel ───────────────────────────────────────── -->
        <div v-if="showInvitePicker" class="invite-picker" role="dialog" aria-modal="true" aria-labelledby="invite-picker-title" @click.stop @keydown.escape="closeInvitePicker">
          <div id="invite-picker-title" class="invite-picker-header">{{ $t('call.invite_to_call') }}</div>
          <input
            v-model="inviteSearch"
            class="invite-picker-search"
            :placeholder="$t('call.search_users')"
            aria-label="Search participants to invite"
            autofocus
          />
          <div class="invite-picker-list">
            <label v-for="u in filteredInviteUsers" :key="u.id" class="invite-picker-user">
              <input type="checkbox" :value="u.id" v-model="selectedInviteIds" class="invite-check" />
              <div class="invite-avatar" :style="inviteAvatarBg(u)">
                <img v-if="u.avatar_url" :src="u.avatar_url" class="invite-avatar-img" alt="" aria-hidden="true" @error="e => e.target.style.display='none'" />
                <span v-else class="invite-avatar-initials">{{ inviteInitials(u) }}</span>
              </div>
              <span class="invite-user-name">{{ u.display_name || u.username }}</span>
            </label>
            <div v-if="!filteredInviteUsers.length" class="invite-empty">{{ $t('call.no_users_to_invite') }}</div>
          </div>
          <button class="invite-send-btn" :disabled="!selectedInviteIds.length" @click="sendInvites">
            {{ $t('call.invite_users') }}
          </button>
        </div>

        <div class="video-controls">
          <button
            class="vc-btn"
            :aria-label="isWindowed ? $t('call.restore_fullscreen') : $t('call.pop_out_window')"
            @click.stop="toggleWindowed"
          >
            <svg v-if="isWindowed" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="4 14 10 14 10 20"/><polyline points="20 10 14 10 14 4"/><line x1="10" y1="14" x2="3" y2="21"/><line x1="21" y1="3" x2="14" y2="10"/></svg>
          </button>
          <button
            :class="['vc-btn', { active: showCallChat }]"
            :aria-label="showCallChat ? $t('call.hide_chat') : $t('call.show_chat')"
            :aria-pressed="showCallChat"
            @click.stop="toggleCallChat"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </button>
          <button class="vc-btn invite-btn" :aria-label="$t('call.invite_to_call')" @click.stop="openInvitePicker">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="8.5" cy="7" r="4"/>
              <line x1="20" y1="8" x2="20" y2="14"/>
              <line x1="23" y1="11" x2="17" y2="11"/>
            </svg>
          </button>
          <button
            :class="['vc-btn', { active: lkState.isScreenSharing }]"
            :aria-label="lkState.isScreenSharing ? $t('call.stop_share') : $t('call.share_screen')"
            :aria-pressed="lkState.isScreenSharing"
            @click.stop="lkToggleScreenShare"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
          </button>
          <button :class="['vc-btn', { active: lkState.isMuted }]" :aria-label="lkState.isMuted ? $t('call.unmute') : $t('call.mute')" :aria-pressed="lkState.isMuted" @click="lkToggleMute">
            <svg v-if="!lkState.isMuted" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
              <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
              <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
          </button>
          <button :class="['vc-btn', { active: lkState.isCameraOff }]" :aria-label="lkState.isCameraOff ? $t('call.camera_on') : $t('call.camera_off')" :aria-pressed="lkState.isCameraOff" @click="lkToggleCamera">
            <svg v-if="!lkState.isCameraOff" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <polygon points="23 7 16 12 23 17 23 7"/>
              <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
            </svg>
          </button>
          <button class="vc-btn end-btn" :aria-label="$t('call.hangup')" @click="lkEnd">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
            </svg>
          </button>
        </div>
        <div
          v-if="isWindowed"
          class="call-window-resize-handle"
          role="button"
          tabindex="0"
          :aria-label="$t('call.resize_window')"
          @pointerdown.stop="onWindowResizeStart"
          @keydown.stop="onWindowResizeHandleKeydown"
        ></div>
        </div>
        <CallChatSidebar
          v-if="showCallChat && activeCallConvId"
          :conversation-id="activeCallConvId"
          @close="setCallChat(false)"
        />
      </div>

      <!-- ── VIDEO OVERLAY (active + video) — WebRTC 1:1 ─────────────────────── -->
      <div v-else-if="state.phase === 'active' && state.hasVideo" class="call-video-chat-row" :class="{ 'is-windowed': isWindowed }" :style="windowStyle">
        <div class="call-stage video-overlay">
        <div
          v-if="isWindowed"
          class="call-window-drag-handle"
          role="button"
          tabindex="0"
          :aria-label="$t('call.move_window')"
          @pointerdown="onWindowDragStart"
          @keydown="onWindowDragHandleKeydown"
        ></div>
        <!-- Remote video fills the background; screen shares are fit (not cropped) since
             their aspect ratio rarely matches the viewport, unlike a camera feed -->
        <video ref="remoteVideo" autoplay playsinline class="remote-video" :class="{ 'remote-video--contain': state.remoteScreenSharing }"></video>

        <!-- Local self-preview (PiP, bottom-right) -->
        <div class="local-pip">
          <video ref="localVideo" autoplay playsinline muted class="local-video"></video>
          <div v-if="state.isCameraOff" class="pip-off">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
            </svg>
          </div>
        </div>

        <!-- Remote name (top-left) -->
        <div class="remote-name">
          {{ state.remoteName }}
          <svg v-if="state.remoteMuted" class="remote-muted-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" role="img" :aria-label="$t('call.remote_muted', { name: state.remoteName || '…' })">
            <line x1="1" y1="1" x2="23" y2="23"/>
            <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
            <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
          </svg>
        </div>
        <div class="call-duration">{{ formattedDuration }}</div>

        <!-- Invite picker for 1:1 → group upgrade -->
        <div v-if="showInvitePicker && upgradeMode" class="invite-picker" role="dialog" aria-modal="true" aria-labelledby="invite-picker-title" @click.stop @keydown.escape="closeInvitePicker">
          <div id="invite-picker-title" class="invite-picker-header">{{ $t('call.invite_to_call') }}</div>
          <input v-model="inviteSearch" class="invite-picker-search" :placeholder="$t('call.search_users')" aria-label="Search participants to invite" autofocus />
          <div class="invite-picker-list">
            <label v-for="u in filteredInviteUsers" :key="u.id" class="invite-picker-user">
              <input type="checkbox" :value="u.id" v-model="selectedInviteIds" class="invite-check" />
              <div class="invite-avatar" :style="inviteAvatarBg(u)">
                <img v-if="u.avatar_url" :src="u.avatar_url" class="invite-avatar-img" alt="" aria-hidden="true" @error="e => e.target.style.display='none'" />
                <span v-else class="invite-avatar-initials">{{ inviteInitials(u) }}</span>
              </div>
              <span class="invite-user-name">{{ u.display_name || u.username }}</span>
            </label>
            <div v-if="!filteredInviteUsers.length" class="invite-empty">{{ $t('call.no_users_to_invite') }}</div>
          </div>
          <button class="invite-send-btn" :disabled="!selectedInviteIds.length" @click="sendInvites">
            {{ $t('call.invite_users') }}
          </button>
        </div>

        <!-- Controls bar -->
        <div class="video-controls">
          <button
            class="vc-btn"
            :aria-label="isWindowed ? $t('call.restore_fullscreen') : $t('call.pop_out_window')"
            @click.stop="toggleWindowed"
          >
            <svg v-if="isWindowed" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="4 14 10 14 10 20"/><polyline points="20 10 14 10 14 4"/><line x1="10" y1="14" x2="3" y2="21"/><line x1="21" y1="3" x2="14" y2="10"/></svg>
          </button>
          <button
            :class="['vc-btn', { active: showCallChat }]"
            :aria-label="showCallChat ? $t('call.hide_chat') : $t('call.show_chat')"
            :aria-pressed="showCallChat"
            @click.stop="toggleCallChat"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </button>
          <button class="vc-btn invite-btn" :aria-label="$t('call.invite_to_call')" @click.stop="openInvitePickerWebRTC">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="8.5" cy="7" r="4"/>
              <line x1="20" y1="8" x2="20" y2="14"/>
              <line x1="23" y1="11" x2="17" y2="11"/>
            </svg>
          </button>

          <button
            :class="['vc-btn', { active: state.isScreenSharing }]"
            :aria-label="state.isScreenSharing ? $t('call.stop_share') : $t('call.share_screen')"
            :aria-pressed="state.isScreenSharing"
            @click.stop="toggleScreenShare"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
          </button>
          <button :class="['vc-btn', { active: state.isMuted }]" :aria-label="state.isMuted ? $t('call.unmute') : $t('call.mute')" :aria-pressed="state.isMuted" @click="toggleMute">
            <svg v-if="!state.isMuted" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
              <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
              <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
          </button>

          <button :class="['vc-btn', { active: state.isCameraOff }]" :aria-label="state.isCameraOff ? $t('call.camera_on') : $t('call.camera_off')" :aria-pressed="state.isCameraOff" @click="toggleCamera">
            <svg v-if="!state.isCameraOff" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <polygon points="23 7 16 12 23 17 23 7"/>
              <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
            </svg>
          </button>

          <button class="vc-btn end-btn" :aria-label="$t('call.hangup')" @click="endCall()">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
            </svg>
          </button>
        </div>
        <div
          v-if="isWindowed"
          class="call-window-resize-handle"
          role="button"
          tabindex="0"
          :aria-label="$t('call.resize_window')"
          @pointerdown.stop="onWindowResizeStart"
          @keydown.stop="onWindowResizeHandleKeydown"
        ></div>
        </div>
        <CallChatSidebar
          v-if="showCallChat && activeCallConvId"
          :conversation-id="activeCallConvId"
          @close="setCallChat(false)"
        />
      </div>

      <!-- ── AUDIO BAR (calling phase or active without video) — WebRTC ─────── -->
      <CallChatSidebar
        v-if="showCallChat && activeCallConvId"
        :conversation-id="activeCallConvId"
        fixed
        @close="setCallChat(false)"
      />
      <div v-else-if="showWebRTC" class="active-call-bar" :class="{ 'bar-with-chat': showCallChat }" :style="showCallChat ? { right: callChatWidth + 'px' } : null">
        <!-- Hidden audio element for remote stream in audio-only mode -->
        <audio ref="remoteAudio" autoplay playsinline></audio>

        <!-- Invite picker anchored above the audio bar -->
        <div v-if="showInvitePicker && upgradeMode" class="invite-picker invite-picker-above-bar" role="dialog" aria-modal="true" aria-labelledby="invite-picker-title" @click.stop @keydown.escape="closeInvitePicker">
          <div id="invite-picker-title" class="invite-picker-header">{{ $t('call.invite_to_call') }}</div>
          <input v-model="inviteSearch" class="invite-picker-search" :placeholder="$t('call.search_users')" aria-label="Search participants to invite" autofocus />
          <div class="invite-picker-list">
            <label v-for="u in filteredInviteUsers" :key="u.id" class="invite-picker-user">
              <input type="checkbox" :value="u.id" v-model="selectedInviteIds" class="invite-check" />
              <div class="invite-avatar" :style="inviteAvatarBg(u)">
                <img v-if="u.avatar_url" :src="u.avatar_url" class="invite-avatar-img" alt="" aria-hidden="true" @error="e => e.target.style.display='none'" />
                <span v-else class="invite-avatar-initials">{{ inviteInitials(u) }}</span>
              </div>
              <span class="invite-user-name">{{ u.display_name || u.username }}</span>
            </label>
            <div v-if="!filteredInviteUsers.length" class="invite-empty">{{ $t('call.no_users_to_invite') }}</div>
          </div>
          <button class="invite-send-btn" :disabled="!selectedInviteIds.length" @click="sendInvites">
            {{ $t('call.invite_users') }}
          </button>
        </div>

        <div class="call-bar-info">
          <div class="call-bar-name">
            {{ state.remoteName || '…' }}
            <svg v-if="state.remoteMuted" class="remote-muted-icon remote-muted-icon--bar" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" role="img" :aria-label="$t('call.remote_muted', { name: state.remoteName || '…' })">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
              <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
            </svg>
          </div>
          <div class="call-bar-status">
            <span v-if="state.phase === 'calling'" class="status-calling">{{ $t('call.calling') }}</span>
            <span v-else class="status-duration">{{ formattedDuration }}</span>
          </div>
        </div>

        <div class="call-bar-actions">
          <button
            :class="['call-bar-btn', { active: showCallChat }]"
            :aria-label="showCallChat ? $t('call.hide_chat') : $t('call.show_chat')"
            :aria-pressed="showCallChat"
            @click.stop="toggleCallChat"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </button>
          <button v-if="state.phase === 'active'" class="call-bar-btn" :aria-label="$t('call.invite_to_call')" @click.stop="openInvitePickerWebRTC">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="8.5" cy="7" r="4"/>
              <line x1="20" y1="8" x2="20" y2="14"/>
              <line x1="23" y1="11" x2="17" y2="11"/>
            </svg>
          </button>
          <button
            v-if="state.hasVideo && state.phase === 'active'"
            :class="['call-bar-btn', { active: state.isScreenSharing }]"
            :aria-label="state.isScreenSharing ? $t('call.stop_share') : $t('call.share_screen')"
            :aria-pressed="state.isScreenSharing"
            @click.stop="toggleScreenShare"
          >
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
          </button>
          <button :class="['call-bar-btn', { active: state.isMuted }]" :aria-label="state.isMuted ? $t('call.unmute') : $t('call.mute')" :aria-pressed="state.isMuted" @click="toggleMute">
            <svg v-if="!state.isMuted" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
              <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
            <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
              <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
              <line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/>
            </svg>
          </button>

          <button v-if="state.hasVideo" :class="['call-bar-btn', { active: state.isCameraOff }]" :aria-label="state.isCameraOff ? $t('call.camera_on') : $t('call.camera_off')" :aria-pressed="state.isCameraOff" @click="toggleCamera">
            <svg v-if="!state.isCameraOff" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <polygon points="23 7 16 12 23 17 23 7"/>
              <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
            </svg>
            <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <line x1="1" y1="1" x2="23" y2="23"/>
              <path d="M21 21H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h3m3-3h6l2 3h4a2 2 0 0 1 2 2v9.34"/>
            </svg>
          </button>

          <button class="call-bar-btn end-btn" :aria-label="$t('call.hangup')" @click="endCall()">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M10.68 13.31a16 16 0 0 0 3.41 2.6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7 2 2 0 0 1 1.72 2v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.42 19.42 0 0 1-3.33-2.67m-2.67-3.34a19.79 19.79 0 0 1-3.07-8.63A2 2 0 0 1 3.6 1.27h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L7.91 9.91M1 1l22 22"/>
            </svg>
          </button>
        </div>
      </div>

    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUIStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'
import { useWebRTCCall } from '@/composables/useWebRTCCall'
import { useLiveKitGroupCall } from '@/composables/useLiveKitGroupCall'
import { useCallSettings } from '@/composables/useCallSettings'
import { useCallChatPanel } from '@/composables/useCallChatPanel'
import { messagesApi } from '@/api/messages'
import LkVideoCell from '@/components/call/LkVideoCell.vue'
import CallChatSidebar from '@/components/call/CallChatSidebar.vue'

const { t } = useI18n()
const ui = useUIStore()
const auth = useAuthStore()
const { showCallChat, toggleCallChat, setCallChat, callChatWidth } = useCallChatPanel()

const {
  state: lkState,
  tiles: lkTiles,
  leaveGroupCall,
  toggleMute: lkToggleMute,
  toggleCamera: lkToggleCamera,
  toggleScreenShare: lkToggleScreenShare,
  inviteUsers,
  sendGroupInvite,
  joinGroupCall,
} = useLiveKitGroupCall()

// ── Invite picker ────────────────────────────────────────────────────────────
const showInvitePicker = ref(false)
const inviteSearch = ref('')
const allUsers = ref([])
const selectedInviteIds = ref([])

// upgradeMode: true when inviting from a 1:1 WebRTC call (needs LK upgrade)
const upgradeMode = ref(false)
const upgradeConvId = ref(null)
const upgradeRemoteId = ref(null)  // the 1:1 partner — auto-invited on upgrade

const filteredInviteUsers = computed(() => {
  const excluded = new Set(lkTiles.map(t => String(t.identity)))
  excluded.add(String(auth.user?.id ?? ''))
  if (upgradeRemoteId.value) excluded.add(String(upgradeRemoteId.value))
  const q = inviteSearch.value.trim().toLowerCase()
  return allUsers.value.filter(u => {
    if (excluded.has(String(u.id))) return false
    if (!q) return true
    return (u.display_name || u.username || '').toLowerCase().includes(q)
  })
})

async function _loadUsers() {
  try {
    const res = await messagesApi.listUsers()
    allUsers.value = res.data || []
  } catch { allUsers.value = [] }
}

// Called from the LiveKit active overlay (already in a group call)
async function openInvitePicker() {
  if (showInvitePicker.value && !upgradeMode.value) { closeInvitePicker(); return }
  upgradeMode.value = false
  upgradeConvId.value = null
  upgradeRemoteId.value = null
  selectedInviteIds.value = []
  inviteSearch.value = ''
  await _loadUsers()
  showInvitePicker.value = true
}

// Called from the WebRTC 1:1 video/audio overlay — will upgrade to LiveKit
async function openInvitePickerWebRTC() {
  if (showInvitePicker.value && upgradeMode.value) { closeInvitePicker(); return }
  upgradeMode.value = true
  upgradeConvId.value = state.convId
  upgradeRemoteId.value = state.remoteUserId
  selectedInviteIds.value = []
  inviteSearch.value = ''
  await _loadUsers()
  showInvitePicker.value = true
}

function closeInvitePicker() {
  showInvitePicker.value = false
  selectedInviteIds.value = []
  inviteSearch.value = ''
  upgradeMode.value = false
  upgradeConvId.value = null
  upgradeRemoteId.value = null
}

async function sendInvites() {
  if (!selectedInviteIds.value.length) return
  if (upgradeMode.value) {
    const convId = upgradeConvId.value
    const allIds = [...selectedInviteIds.value]
    if (upgradeRemoteId.value) allIds.push(upgradeRemoteId.value)
    sendGroupInvite(convId, allIds)
    closeInvitePicker()
    endCall()
    await joinGroupCall(convId, state.remoteName || '', [])
  } else {
    inviteUsers(selectedInviteIds.value)
    ui.success(t('call.invite_sent'))
    closeInvitePicker()
  }
}

function inviteInitials(u) {
  return ((u.display_name || u.username) || '?').slice(0, 2).toUpperCase()
}

const AVATAR_COLORS = ['#6366f1','#8b5cf6','#ec4899','#f97316','#14b8a6','#3b82f6','#22c55e']
function inviteAvatarBg(u) {
  const idx = (u.id || 0) % AVATAR_COLORS.length
  return { background: AVATAR_COLORS[idx] }
}

const { state, toggleMute, toggleCamera, toggleScreenShare, endCall, setRemoteStreamCallback, setLocalStreamCallback } = useWebRTCCall()

const showLiveKit = computed(
  () => lkState.phase === 'connecting' || lkState.phase === 'active'
)
const showWebRTC = computed(
  () =>
    !showLiveKit.value &&
    (state.phase === 'calling' || state.phase === 'active')
)

/** Conversation used for optional text chat (same as open DM / group room). */
const activeCallConvId = computed(() => {
  if (lkState.phase === 'connecting' || lkState.phase === 'active') return lkState.convId
  if (state.phase === 'calling' || state.phase === 'ringing' || state.phase === 'active') return state.convId
  return null
})
const { audioOutputId } = useCallSettings()

const remoteVideo = ref(null)
const localVideo  = ref(null)
const remoteAudio = ref(null)

// Store streams locally so we can re-apply when DOM elements appear
const _remoteStream = ref(null)
const _localStream  = ref(null)

function _applySinkId(el) {
  if (!el || !audioOutputId.value) return
  if (typeof el.setSinkId === 'function') {
    el.setSinkId(audioOutputId.value).catch(() => {})
  }
}

function _applyStreams() {
  nextTick(() => {
    if (_remoteStream.value) {
      if (remoteVideo.value) { remoteVideo.value.srcObject = _remoteStream.value; _applySinkId(remoteVideo.value) }
      if (remoteAudio.value) { remoteAudio.value.srcObject = _remoteStream.value; _applySinkId(remoteAudio.value) }
    }
    if (_localStream.value && localVideo.value) {
      localVideo.value.srcObject = _localStream.value
    }
  })
}

// Re-apply sink when output device changes mid-call
watch(audioOutputId, () => {
  _applySinkId(remoteVideo.value)
  _applySinkId(remoteAudio.value)
})

onMounted(() => {
  setRemoteStreamCallback((stream) => {
    _remoteStream.value = stream
    _applyStreams()
  })
  setLocalStreamCallback((stream) => {
    _localStream.value = stream
    _applyStreams()
  })
  window.addEventListener('resize', _onViewportResize)
})

// Re-apply when phase changes (video overlay appears after 'active' transition)
watch(() => state.phase, _applyStreams)
watch(() => state.hasVideo, _applyStreams)

// ── Duration timer ──────────────────────────────────────────────────────────
const durationSeconds = ref(0)
let _durationTimer = null

watch(() => state.phase, (phase) => {
  if (phase === 'active') {
    durationSeconds.value = 0
    _durationTimer = setInterval(() => { durationSeconds.value++ }, 1000)
  } else {
    clearInterval(_durationTimer)
    _durationTimer = null
    durationSeconds.value = 0
  }
})

// Watch errorMsg directly so the toast fires even when phase is already
// 'ended' (e.g. ring timeout raced the reject signal and set it first).
watch(() => state.errorMsg, (msg) => {
  if (!msg) return
  const key = msg === 'unavailable'
    ? t('call.unavailable', { name: state.remoteName || '…' })
    : msg === 'rejected'
      ? t('call.rejected', { name: state.remoteName || '…' })
      : msg === 'no_mic'
        ? t('call.no_mic')
        : msg === 'failed'
          ? t('call.failed')
          : null
  if (key) ui.error(key)
})

const formattedDuration = computed(() => {
  const s = durationSeconds.value
  const m = Math.floor(s / 60)
  return `${String(m).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
})

function lkEnd() {
  void leaveGroupCall()
}

function tileLabel(tile) {
  if (tile.isScreenShare) {
    if (tile.isLocal) return t('call.your_screen')
    const n = tile.name || t('call.group_peer', { id: tile.identity })
    return t('call.peer_screen', { name: n })
  }
  if (tile.isLocal) return t('call.group_you')
  return tile.name || t('call.group_peer', { id: tile.identity })
}

const lkDurationSeconds = ref(0)
let _lkDurTimer = null

watch(
  () => lkState.phase,
  (phase) => {
    clearInterval(_lkDurTimer)
    _lkDurTimer = null
    if (phase === 'active') {
      lkDurationSeconds.value = 0
      _lkDurTimer = setInterval(() => {
        lkDurationSeconds.value++
      }, 1000)
    } else {
      lkDurationSeconds.value = 0
    }
  }
)

const formattedLkDuration = computed(() => {
  const s = lkDurationSeconds.value
  const m = Math.floor(s / 60)
  return `${String(m).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
})

watch(
  () => lkState.errorMsg,
  (msg) => {
    if (!msg) return
    const key =
      msg === 'livekit_unavailable'
        ? t('call.livekit_unavailable')
        : msg === 'livekit_connect_failed'
          ? t('call.livekit_connect_failed')
          : null
    if (key) ui.error(key)
  }
)

// ── Windowed (floating, draggable/resizable) call mode ─────────────────────
// Lets a call shrink from the full-viewport overlay into a small movable panel,
// e.g. so the caller can keep working elsewhere while staying on the call.
const WINDOW_GEOM_KEY = 'warmdesk_call_window_geom'
const MIN_WINDOW_WIDTH = 280
const MIN_WINDOW_HEIGHT = 200
const DEFAULT_WINDOW_WIDTH = 420
const DEFAULT_WINDOW_HEIGHT = 300
const WINDOW_MARGIN = 20
const WINDOW_KEY_STEP = 24
const WINDOW_KEY_STEP_LARGE = 80

function _defaultWindowGeom() {
  const width = DEFAULT_WINDOW_WIDTH
  const height = DEFAULT_WINDOW_HEIGHT
  return {
    width,
    height,
    x: Math.max(WINDOW_MARGIN, window.innerWidth - width - WINDOW_MARGIN),
    y: Math.max(WINDOW_MARGIN, window.innerHeight - height - WINDOW_MARGIN),
  }
}

function _loadWindowGeom() {
  try {
    const raw = localStorage.getItem(WINDOW_GEOM_KEY)
    if (raw) {
      const g = JSON.parse(raw)
      if (g && Number.isFinite(g.x) && Number.isFinite(g.y) && Number.isFinite(g.width) && Number.isFinite(g.height)) return g
    }
  } catch { /* ignore malformed/unavailable storage */ }
  return _defaultWindowGeom()
}

const isWindowed = ref(false)
const windowGeom = ref(_loadWindowGeom())

function _saveWindowGeom() {
  try { localStorage.setItem(WINDOW_GEOM_KEY, JSON.stringify(windowGeom.value)) } catch { /* ignore */ }
}

function _clampWindowGeom() {
  windowGeom.value.width = Math.min(windowGeom.value.width, window.innerWidth)
  windowGeom.value.height = Math.min(windowGeom.value.height, window.innerHeight)
  const maxX = Math.max(0, window.innerWidth - windowGeom.value.width)
  const maxY = Math.max(0, window.innerHeight - windowGeom.value.height)
  windowGeom.value.x = Math.min(Math.max(0, windowGeom.value.x), maxX)
  windowGeom.value.y = Math.min(Math.max(0, windowGeom.value.y), maxY)
}

const windowStyle = computed(() => {
  if (!isWindowed.value) return null
  const g = windowGeom.value
  return { top: `${g.y}px`, left: `${g.x}px`, width: `${g.width}px`, height: `${g.height}px`, right: 'auto', bottom: 'auto' }
})

function toggleWindowed() {
  isWindowed.value = !isWindowed.value
  if (isWindowed.value) {
    _clampWindowGeom()
    // The chat sidebar is a viewport-edge-pinned panel, not meant to share
    // space with a small floating window — close it rather than overlap it.
    if (showCallChat.value) setCallChat(false)
  }
}

let _winDrag = null
function onWindowDragStart(e) {
  if (e.button !== 0) return
  e.preventDefault()
  _winDrag = { startX: e.clientX, startY: e.clientY, startLeft: windowGeom.value.x, startTop: windowGeom.value.y }
  window.addEventListener('pointermove', onWindowDragMove)
  window.addEventListener('pointerup', onWindowDragEnd, { once: true })
}
function onWindowDragMove(e) {
  if (!_winDrag) return
  const maxX = Math.max(0, window.innerWidth - windowGeom.value.width)
  const maxY = Math.max(0, window.innerHeight - windowGeom.value.height)
  windowGeom.value.x = Math.min(Math.max(0, _winDrag.startLeft + (e.clientX - _winDrag.startX)), maxX)
  windowGeom.value.y = Math.min(Math.max(0, _winDrag.startTop + (e.clientY - _winDrag.startY)), maxY)
}
function onWindowDragEnd() {
  window.removeEventListener('pointermove', onWindowDragMove)
  _winDrag = null
  _saveWindowGeom()
}

function onWindowDragHandleKeydown(e) {
  const step = e.shiftKey ? WINDOW_KEY_STEP_LARGE : WINDOW_KEY_STEP
  const maxX = Math.max(0, window.innerWidth - windowGeom.value.width)
  const maxY = Math.max(0, window.innerHeight - windowGeom.value.height)
  if (e.key === 'ArrowLeft') windowGeom.value.x = Math.max(0, windowGeom.value.x - step)
  else if (e.key === 'ArrowRight') windowGeom.value.x = Math.min(maxX, windowGeom.value.x + step)
  else if (e.key === 'ArrowUp') windowGeom.value.y = Math.max(0, windowGeom.value.y - step)
  else if (e.key === 'ArrowDown') windowGeom.value.y = Math.min(maxY, windowGeom.value.y + step)
  else return
  e.preventDefault()
  _saveWindowGeom()
}

let _winResize = null
function onWindowResizeStart(e) {
  if (e.button !== 0) return
  e.preventDefault()
  _winResize = { startX: e.clientX, startY: e.clientY, startWidth: windowGeom.value.width, startHeight: windowGeom.value.height }
  window.addEventListener('pointermove', onWindowResizeMove)
  window.addEventListener('pointerup', onWindowResizeEnd, { once: true })
}
function onWindowResizeMove(e) {
  if (!_winResize) return
  const maxWidth = window.innerWidth - windowGeom.value.x
  const maxHeight = window.innerHeight - windowGeom.value.y
  windowGeom.value.width = Math.min(Math.max(MIN_WINDOW_WIDTH, _winResize.startWidth + (e.clientX - _winResize.startX)), maxWidth)
  windowGeom.value.height = Math.min(Math.max(MIN_WINDOW_HEIGHT, _winResize.startHeight + (e.clientY - _winResize.startY)), maxHeight)
}
function onWindowResizeEnd() {
  window.removeEventListener('pointermove', onWindowResizeMove)
  _winResize = null
  _saveWindowGeom()
}

function onWindowResizeHandleKeydown(e) {
  const step = e.shiftKey ? WINDOW_KEY_STEP_LARGE : WINDOW_KEY_STEP
  const maxWidth = window.innerWidth - windowGeom.value.x
  const maxHeight = window.innerHeight - windowGeom.value.y
  if (e.key === 'ArrowLeft') windowGeom.value.width = Math.max(MIN_WINDOW_WIDTH, windowGeom.value.width - step)
  else if (e.key === 'ArrowRight') windowGeom.value.width = Math.min(maxWidth, windowGeom.value.width + step)
  else if (e.key === 'ArrowUp') windowGeom.value.height = Math.max(MIN_WINDOW_HEIGHT, windowGeom.value.height - step)
  else if (e.key === 'ArrowDown') windowGeom.value.height = Math.min(maxHeight, windowGeom.value.height + step)
  else return
  e.preventDefault()
  _saveWindowGeom()
}

function _onViewportResize() {
  if (isWindowed.value) _clampWindowGeom()
}

// A call that ends while windowed shouldn't leave the next call starting small.
watch([showWebRTC, showLiveKit], ([webrtc, liveKit]) => {
  if (!webrtc && !liveKit) isWindowed.value = false
})

onUnmounted(() => {
  setRemoteStreamCallback(null)
  setLocalStreamCallback(null)
  clearInterval(_durationTimer)
  clearInterval(_lkDurTimer)
  window.removeEventListener('resize', _onViewportResize)
  window.removeEventListener('pointermove', onWindowDragMove)
  window.removeEventListener('pointermove', onWindowResizeMove)
})

</script>

<style scoped>
/* ── VIDEO OVERLAY ─────────────────────────────────────────────────────────── */
.call-video-chat-row {
  position: fixed;
  inset: 0;
  z-index: 500;
  display: flex;
  flex-direction: row;
  align-items: stretch;
  background: #000;
}
.call-stage {
  flex: 1;
  min-width: 0;
  min-height: 0;
  position: relative;
}
.call-stage.video-overlay {
  position: relative;
  inset: auto;
  z-index: auto;
  width: 100%;
  height: 100%;
}
.call-stage.lk-grid-overlay {
  height: 100%;
}

/* ── Windowed (floating, draggable/resizable) call mode ───────────────────── */
.call-video-chat-row.is-windowed {
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 40px rgba(0,0,0,0.55);
  border: 1px solid rgba(255,255,255,0.15);
}
.call-window-drag-handle {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 32px;
  z-index: 26;
  cursor: move;
}
.call-window-drag-handle:focus-visible {
  outline: 2px solid #fff;
  outline-offset: -2px;
}
.call-window-resize-handle {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 18px;
  height: 18px;
  z-index: 26;
  cursor: nwse-resize;
}
.call-window-resize-handle::after {
  content: '';
  position: absolute;
  right: 3px;
  bottom: 3px;
  width: 8px;
  height: 8px;
  border-right: 2px solid rgba(255,255,255,0.7);
  border-bottom: 2px solid rgba(255,255,255,0.7);
}
.call-window-resize-handle:focus-visible {
  outline: 2px solid #fff;
  outline-offset: -2px;
}
/* Scale controls/pip/text down so they still fit a small floating window */
.is-windowed .video-controls { padding: 8px 6px 10px; gap: 6px 8px; }
.is-windowed .vc-btn { width: 34px; height: 34px; }
.is-windowed .vc-btn svg { width: 14px; height: 14px; }
.is-windowed .local-pip { width: 84px; height: 47px; bottom: 44px; right: 8px; }
.is-windowed .remote-name { top: 40px; left: 10px; font-size: 13px; }
.is-windowed .call-duration { top: 58px; left: 10px; font-size: 11px; }
.is-windowed .lk-grid-top { padding-top: 36px; }
.is-windowed .invite-picker { width: calc(100% - 24px); bottom: 56px; }

.video-overlay {
  position: fixed;
  inset: 0;
  z-index: 500;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remote-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.remote-video--contain {
  object-fit: contain;
}

.local-pip {
  position: absolute;
  bottom: 80px;
  right: 16px;
  width: 160px;
  height: 90px;
  border-radius: 10px;
  overflow: hidden;
  background: #1a1a1a;
  border: 2px solid rgba(255,255,255,0.15);
  box-shadow: 0 4px 16px rgba(0,0,0,0.5);
}
.local-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* mirror self-view */
}
.pip-off {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1a1a1a;
  color: rgba(255,255,255,0.4);
}

.remote-name {
  position: absolute;
  top: 20px;
  left: 24px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  text-shadow: 0 1px 6px rgba(0,0,0,0.7);
  display: flex;
  align-items: center;
  gap: 6px;
}
.remote-muted-icon {
  color: #f87171;
  flex-shrink: 0;
}
.remote-muted-icon--bar {
  color: #ef4444;
  vertical-align: -1px;
}
.call-duration {
  position: absolute;
  top: 44px;
  left: 24px;
  color: rgba(255,255,255,0.7);
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  text-shadow: 0 1px 4px rgba(0,0,0,0.6);
}

.video-controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 12px 14px;
  padding: 16px 12px 20px;
  background: linear-gradient(transparent, rgba(0,0,0,0.65));
  z-index: 25;
  box-sizing: border-box;
  max-width: 100%;
}

.vc-btn {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: none;
  background: rgba(255,255,255,0.18);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, transform 0.1s;
  backdrop-filter: blur(4px);
}
.vc-btn:hover { background: rgba(255,255,255,0.28); transform: scale(1.06); }
.vc-btn.active { background: rgba(255,255,255,0.85); color: #111; }
.vc-btn.end-btn { background: #ef4444; }
.vc-btn.end-btn:hover { background: #dc2626; transform: scale(1.06); }

/* ── AUDIO BAR ─────────────────────────────────────────────────────────────── */
.active-call-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 490;
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  padding: 10px 24px;
  box-shadow: 0 -4px 16px rgba(0,0,0,0.12);
  height: 56px;
}
/* .bar-with-chat's actual `right` offset is set inline (callChatWidth is dynamic/resizable) */
.call-bar-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.call-bar-name { font-size: 14px; font-weight: 600; color: var(--color-text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.call-bar-status { font-size: 12px; }
.status-calling { color: var(--color-text-muted); animation: pulse 1.4s ease-in-out infinite; }
.status-duration { color: #22c55e; font-variant-numeric: tabular-nums; }

.call-bar-actions { display: flex; align-items: center; gap: 10px; }
.call-bar-btn {
  width: 36px; height: 36px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: var(--color-bg);
  color: var(--color-text);
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.call-bar-btn:hover { background: var(--color-surface-raised, var(--color-bg)); }
.call-bar-btn.active { background: var(--color-primary); border-color: var(--color-primary); color: #fff; }
.call-bar-btn.end-btn { background: #ef4444; border-color: #ef4444; color: #fff; }
.call-bar-btn.end-btn:hover { opacity: 0.88; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

/* ── Transitions ────────────────────────────────────────────────────────────── */
.call-ui-enter-from,
.call-ui-leave-to { opacity: 0; transform: translateY(100%); }
.call-ui-enter-active,
.call-ui-leave-active { transition: opacity 0.2s, transform 0.2s; }

/* ── LiveKit group grid ───────────────────────────────────────────────────── */
.lk-connecting-bar {
  z-index: 490;
}
.lk-grid-overlay {
  display: flex;
  flex-direction: column;
  padding: 0 0 12px;
  box-sizing: border-box;
}
.lk-grid-top {
  flex: 0 0 auto;
  padding: 16px 20px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.lk-grid {
  flex: 1;
  min-height: 0;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 8px 16px 88px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
  align-content: start;
  overflow-y: auto;
}

/* ── Invite button ────────────────────────────────────────────────────────── */
.vc-btn.invite-btn { background: rgba(255,255,255,0.18); }
.vc-btn.invite-btn:hover { background: rgba(255,255,255,0.28); }

/* ── Invite picker panel ──────────────────────────────────────────────────── */
.invite-picker {
  position: absolute;
  bottom: 88px;
  left: 50%;
  transform: translateX(-50%);
  width: 300px;
  max-width: calc(100vw - 24px);
  background: var(--color-surface, #1e1e2e);
  border: 1px solid rgba(255,255,255,0.15);
  border-radius: 12px;
  padding: 14px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 30;
}
.invite-picker-header {
  font-size: 13px;
  font-weight: 600;
  color: #fff;
}
.invite-picker-search {
  width: 100%;
  box-sizing: border-box;
  padding: 7px 10px;
  border-radius: 7px;
  border: 1px solid rgba(255,255,255,0.2);
  background: rgba(255,255,255,0.08);
  color: #fff;
  font-size: 13px;
  outline: none;
}
.invite-picker-search::placeholder { color: rgba(255,255,255,0.4); }
.invite-picker-list {
  max-height: 200px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.invite-picker-user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.12s;
}
.invite-picker-user:hover { background: rgba(255,255,255,0.1); }
.invite-check { accent-color: #3b82f6; width: 15px; height: 15px; cursor: pointer; flex-shrink: 0; }
.invite-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.invite-avatar-img { width: 100%; height: 100%; object-fit: cover; }
.invite-avatar-initials { color: #fff; font-size: 11px; font-weight: 700; }
.invite-user-name { font-size: 13px; color: #fff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.invite-empty { font-size: 13px; color: rgba(255,255,255,0.5); text-align: center; padding: 10px 0; }
.invite-send-btn {
  width: 100%;
  padding: 9px;
  border-radius: 8px;
  border: none;
  background: #3b82f6;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.invite-send-btn:hover:not(:disabled) { background: #2563eb; }
.invite-send-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Variant: anchored above the audio bar */
.invite-picker-above-bar {
  position: fixed;
  bottom: 64px;
  left: 50%;
  transform: translateX(-50%);
}
</style>
