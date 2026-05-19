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
      <div v-if="lkState.phase === 'connecting'" class="active-call-bar lk-connecting-bar" :class="{ 'bar-with-chat': showCallChat }">
        <div class="call-bar-info">
          <div class="call-bar-name">{{ lkState.title || $t('call.group_call') }}</div>
          <div class="call-bar-status"><span class="status-calling">{{ $t('call.joining_group') }}</span></div>
        </div>
      </div>

      <!-- ── LiveKit: active (grid) ───────────────────────────────────────── -->
      <div v-else-if="lkState.phase === 'active'" class="call-video-chat-row">
        <div class="call-stage lk-grid-overlay" @click.self="closeInvitePicker">
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
        <div v-if="showInvitePicker" class="invite-picker" @click.stop>
          <div class="invite-picker-header">{{ $t('call.invite_to_call') }}</div>
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
        </div>
        <CallChatSidebar
          v-if="showCallChat && activeCallConvId"
          :conversation-id="activeCallConvId"
          @close="setCallChat(false)"
        />
      </div>

      <!-- ── VIDEO OVERLAY (active + video) — WebRTC 1:1 ─────────────────────── -->
      <div v-else-if="state.phase === 'active' && state.hasVideo" class="call-video-chat-row">
        <div class="call-stage video-overlay">
        <!-- Remote video fills the background -->
        <video ref="remoteVideo" autoplay playsinline class="remote-video"></video>

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
        <div class="remote-name">{{ state.remoteName }}</div>
        <div class="call-duration">{{ formattedDuration }}</div>

        <!-- Invite picker for 1:1 → group upgrade -->
        <div v-if="showInvitePicker && upgradeMode" class="invite-picker" @click.stop>
          <div class="invite-picker-header">{{ $t('call.invite_to_call') }}</div>
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
      <div v-else-if="showWebRTC" class="active-call-bar" :class="{ 'bar-with-chat': showCallChat }">
        <!-- Hidden audio element for remote stream in audio-only mode -->
        <audio ref="remoteAudio" autoplay playsinline></audio>

        <!-- Invite picker anchored above the audio bar -->
        <div v-if="showInvitePicker && upgradeMode" class="invite-picker invite-picker-above-bar" @click.stop>
          <div class="invite-picker-header">{{ $t('call.invite_to_call') }}</div>
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
          <div class="call-bar-name">{{ state.remoteName || '…' }}</div>
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
const { showCallChat, toggleCallChat, setCallChat } = useCallChatPanel()

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

onUnmounted(() => {
  setRemoteStreamCallback(null)
  setLocalStreamCallback(null)
  clearInterval(_durationTimer)
  clearInterval(_lkDurTimer)
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
.active-call-bar.bar-with-chat {
  right: 340px;
}
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
