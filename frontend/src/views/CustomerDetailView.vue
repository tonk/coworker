<template>
  <div class="customer-detail-page">
    <div v-if="loading" class="loading">{{ $t('common.loading') }}</div>
    <template v-else-if="detail">

      <!-- Customer header -->
      <div class="cust-header">
        <div class="cust-logo-wrap">
          <img v-if="detail.customer.logo_url" :src="resolveAssetUrl(detail.customer.logo_url)" class="cust-logo" alt="" />
          <span v-else class="cust-logo-placeholder">{{ detail.customer.name[0] }}</span>
        </div>
        <div class="cust-info">
          <div v-if="!editingName" class="cust-name-row">
            <h1 class="cust-name">{{ detail.customer.name }}</h1>
            <button v-if="canManage" class="icon-btn" @click="startEditName" :aria-label="$t('common.edit')" title="Edit">✎</button>
          </div>
          <div v-else class="cust-name-row">
            <input class="form-input name-input" v-model="nameEdit" :aria-label="$t('customer.name')" @keydown.enter="saveNameEdit" @keydown.escape="editingName = false" />
            <button class="btn btn-primary btn-sm" @click="saveNameEdit">{{ $t('common.save') }}</button>
            <button class="btn btn-sm" @click="editingName = false">{{ $t('common.cancel') }}</button>
          </div>
          <p v-if="detail.customer.description" class="cust-desc">{{ detail.customer.description }}</p>
        </div>
        <div class="cust-actions">
          <button
            class="star-btn"
            :class="{ active: detail.customer.is_favorite }"
            @click="toggleFav"
            :aria-label="detail.customer.is_favorite ? $t('customer.unstar') : $t('customer.star')"
            :aria-pressed="detail.customer.is_favorite"
          >{{ detail.customer.is_favorite ? '★' : '☆' }}</button>
          <button v-if="canManage" class="btn btn-sm" @click="openEdit">{{ $t('customer.edit') }}</button>
          <button v-if="auth.isAdmin" class="btn btn-sm btn-danger" @click="doDelete">{{ $t('common.delete') }}</button>
        </div>
      </div>

      <!-- Contracts -->
      <section class="contracts-section">
        <div class="section-header-row">
          <h2>{{ $t('contract.contracts') }}</h2>
          <button v-if="canManage" class="btn btn-primary btn-sm" @click="showAddContract = true">
            + {{ $t('contract.new_contract') }}
          </button>
        </div>

        <div v-if="!detail.contracts.length && !detail.projects.length" class="empty-state">
          {{ $t('contract.no_contracts') }}
        </div>

        <div v-for="grp in detail.contracts" :key="grp.id" class="contract-block">
          <div class="contract-header">
            <div class="contract-title-row">
              <span class="contract-icon">📋</span>
              <strong>{{ grp.name }}</strong>
              <span v-if="grp.start_date || grp.end_date" class="contract-dates">
                {{ formatDate(grp.start_date) }}{{ grp.end_date ? ' – ' + formatDate(grp.end_date) : '' }}
              </span>
              <span v-if="grp.price_per_hour != null" class="contract-rate">{{ grp.price_per_hour }} {{ grp.currency }}/h</span>
            </div>
            <div v-if="grp.time_slots && grp.time_slots.length" class="slot-list">
              <div v-for="slot in grp.time_slots" :key="slot.id" class="slot-item">
                <span class="slot-time">{{ formatSlotTimeRange(slot) }}</span>
                <span v-if="slot.day_type && slot.day_type !== 'all'" class="slot-days">{{ $t('contract.slot_days_' + slot.day_type) }}</span>
                <span v-if="slot.label" class="slot-label">{{ slot.label }}</span>
                <span v-if="slot.multiplication_factor != null" class="slot-factor">×{{ slot.multiplication_factor }}</span>
                <span v-if="slot.hourly_rate != null" class="slot-rate">{{ slot.hourly_rate }} {{ grp.currency }}/h</span>
              </div>
            </div>
            <div class="contract-actions">
              <button v-if="canManage" class="icon-btn" @click="editContract(grp)" :aria-label="$t('contract.edit')" title="Edit">✎</button>
              <button v-if="auth.isAdmin" class="icon-btn icon-danger" @click="deleteContract(grp)" :aria-label="$t('common.delete')" title="Delete">✕</button>
            </div>
          </div>
          <p v-if="grp.description" class="contract-desc">{{ grp.description }}</p>
          <div v-if="grp.projects.length" class="projects-mini-grid">
            <RouterLink
              v-for="p in grp.projects"
              :key="p.id"
              :to="`/projects/${p.slug}`"
              class="project-mini-tile"
            >
              <img v-if="projectAvatar(p)" :src="projectAvatar(p)" class="proj-avatar" alt="" />
              <span v-else class="proj-dot" :style="{ background: p.color || '#6366f1' }"></span>
              <span>{{ p.name }}</span>
            </RouterLink>
          </div>
          <div v-else class="empty-state-sm">{{ $t('project.no_projects') }}</div>
        </div>

        <!-- Unassigned projects -->
        <div v-if="detail.projects.length" class="contract-block unassigned-block">
          <div class="contract-header">
            <div class="contract-title-row">
              <span class="contract-icon">📁</span>
              <strong>{{ $t('customer.unassigned') }}</strong>
            </div>
          </div>
          <div class="projects-mini-grid">
            <RouterLink
              v-for="p in detail.projects"
              :key="p.id"
              :to="`/projects/${p.slug}`"
              class="project-mini-tile"
            >
              <img v-if="projectAvatar(p)" :src="projectAvatar(p)" class="proj-avatar" alt="" />
              <span v-else class="proj-dot" :style="{ background: p.color || '#6366f1' }"></span>
              <span>{{ p.name }}</span>
            </RouterLink>
          </div>
        </div>
      </section>

      <!-- Members section (visible to admins and customer-admins) -->
      <section v-if="canManage" class="members-section">
        <div class="section-header-row">
          <h2>{{ $t('customer.members') }}</h2>
          <button class="btn btn-primary btn-sm" @click="openAddMember">+ {{ $t('customer.add_member') }}</button>
        </div>
        <div v-if="members.length === 0" class="empty-state-sm" style="padding:16px 0">
          {{ $t('customer.no_members') }}
        </div>
        <div v-else class="members-list">
          <div v-for="m in members" :key="m.user_id" class="member-row">
            <img :src="resolveAssetUrl(m.avatar_url || m.gravatar_url)" class="member-avatar" alt="" />
            <div class="member-info">
              <span class="member-name">{{ m.display_name || m.username }}</span>
              <span class="member-email">{{ m.email }}</span>
            </div>
            <span :class="['role-badge', m.role === 'admin' ? 'role-admin' : 'role-member']">
              {{ m.role === 'admin' ? $t('customer.role_admin') : $t('customer.role_member') }}
            </span>
            <div class="member-actions">
              <button
                v-if="m.role === 'member'"
                class="btn btn-sm"
                @click="setMemberRole(m.user_id, 'admin')"
                :aria-label="$t('customer.promote')"
                :title="$t('customer.promote')"
              >↑</button>
              <button
                v-else-if="auth.isAdmin || m.user_id !== authUserId"
                class="btn btn-sm"
                @click="setMemberRole(m.user_id, 'member')"
                :aria-label="$t('customer.demote')"
                :title="$t('customer.demote')"
              >↓</button>
              <button
                v-if="auth.isAdmin || m.user_id !== authUserId"
                class="icon-btn icon-danger"
                @click="removeMember(m.user_id)"
                :aria-label="$t('common.delete') + ' ' + (m.display_name || m.username)"
                title="Remove"
              >✕</button>
            </div>
          </div>
        </div>

        <!-- Groups with access -->
        <div style="margin-top:20px">
          <h3 style="font-size:13px;font-weight:600;color:var(--color-text-muted);text-transform:uppercase;letter-spacing:.04em;margin-bottom:10px">
            {{ $t('groups.groups_with_access') }}
          </h3>
          <div v-if="!customerGroups.length" style="color:var(--color-text-muted);font-size:13px;margin-bottom:10px">{{ $t('groups.no_group_access') }}</div>
          <div v-for="g in customerGroups" :key="g.group_id" class="member-row" style="align-items:center">
            <img v-if="groupAvatar(g.group)" :src="groupAvatar(g.group)" class="group-avatar" alt="" />
            <div class="member-info">
              <span class="member-name">{{ g.group.name }}</span>
              <span class="member-email">{{ g.members.map(m => m.user.display_name || m.user.username).join(', ') || '—' }}</span>
            </div>
            <span :class="['role-badge', g.role === 'admin' ? 'role-admin' : 'role-member']">{{ g.role }}</span>
            <button v-if="auth.isAdmin" class="icon-btn icon-danger" style="margin-left:8px" @click="removeGroupFromCustomer(g.group_id)" :aria-label="$t('common.delete') + ' ' + g.group.name" title="Remove">✕</button>
          </div>
          <div v-if="auth.isAdmin" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap;margin-top:10px">
            <select class="form-input" v-model="addGroupId" style="flex:1;min-width:160px">
              <option value="">— {{ $t('groups.add_group') }} —</option>
              <option v-for="g in groupsNotOnCustomer" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
            <select class="form-input" v-model="addGroupRole" style="width:110px">
              <option value="member">{{ $t('customer.role_member') }}</option>
              <option value="admin">{{ $t('customer.role_admin') }}</option>
            </select>
            <button class="btn btn-primary btn-sm" :disabled="!addGroupId" @click="addGroupToCustomer">{{ $t('common.add') }}</button>
          </div>
        </div>
      </section>

    </template>

    <!-- Edit customer modal -->
    <BaseModal v-if="showEdit" :title="$t('customer.edit')" @close="showEdit = false">
      <div class="form-group">
        <label class="form-label" for="edit-cust-name">{{ $t('customer.name') }}</label>
        <input id="edit-cust-name" class="form-input" v-model="editForm.name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-cust-desc">{{ $t('customer.description') }}</label>
        <textarea id="edit-cust-desc" class="form-input" v-model="editForm.description" rows="3"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-cust-logo">{{ $t('customer.logo_url') }}</label>
        <div style="display:flex;gap:8px;align-items:center">
          <input id="edit-cust-logo" class="form-input" v-model="editForm.logo_url" placeholder="https://..." style="flex:1" />
          <button type="button" class="btn btn-secondary btn-sm" @click="$refs.logoFileInput.click()">{{ $t('customer.upload_logo') }}</button>
        </div>
        <input ref="logoFileInput" type="file" accept="image/*" style="display:none" @change="onLogoFileSelected" />
      </div>
      <template #footer>
        <button class="btn" @click="showEdit = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEdit">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>

    <!-- Add / edit contract modal -->
    <BaseModal
      v-if="showAddContract || editingContract"
      :title="editingContract ? $t('contract.edit') : $t('contract.new_contract')"
      :resizable="true"
      style="--modal-width: 720px"
      @close="closeContractModal"
    >
      <div class="form-group">
        <label class="form-label" for="contract-name">{{ $t('contract.name') }}</label>
        <input id="contract-name" class="form-input" v-model="contractForm.name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="contract-desc">{{ $t('contract.description') }}</label>
        <textarea id="contract-desc" class="form-input" v-model="contractForm.description" rows="2"></textarea>
      </div>
      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label" for="contract-start">{{ $t('contract.start_date') }}</label>
          <div class="date-input-row">
            <input id="contract-start" class="form-input" type="text" v-model="displayContractStartDate" :placeholder="dateOnlyFormat()" @blur="parseContractStartDate" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="contractForm.start_date" @change="onContractStartDateChange" />
            </label>
            <button v-if="displayContractStartDate" class="btn-icon-xs" @click="displayContractStartDate = ''; contractForm.start_date = ''" title="Clear">×</button>
          </div>
        </div>
        <div class="form-group half">
          <label class="form-label" for="contract-end">{{ $t('contract.end_date') }}</label>
          <div class="date-input-row">
            <input id="contract-end" class="form-input" type="text" v-model="displayContractEndDate" :placeholder="dateOnlyFormat()" @blur="parseContractEndDate" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="contractForm.end_date" @change="onContractEndDateChange" />
            </label>
            <button v-if="displayContractEndDate" class="btn-icon-xs" @click="displayContractEndDate = ''; contractForm.end_date = ''" title="Clear">×</button>
          </div>
        </div>
      </div>
      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label" for="contract-rate">{{ $t('contract.price_per_hour') }}</label>
          <input id="contract-rate" class="form-input" type="number" min="0" step="0.01" v-model="contractForm.price_per_hour" />
        </div>
        <div class="form-group half">
          <label class="form-label" for="contract-currency">{{ $t('contract.currency') }}</label>
          <select id="contract-currency" class="form-input" v-model="contractForm.currency">
            <option value="€">€ (EUR)</option>
            <option value="USD">USD ($)</option>
            <option value="GBP">GBP (£)</option>
            <option value="CHF">CHF</option>
            <option value="SEK">SEK (kr)</option>
            <option value="NOK">NOK (kr)</option>
            <option value="DKK">DKK (kr)</option>
            <option value="PLN">PLN (zł)</option>
            <option value="CZK">CZK (Kč)</option>
          </select>
        </div>
      </div>
      <div class="form-group">
        <div class="slots-header">
          <label class="form-label">{{ $t('contract.time_slots') }}</label>
          <button type="button" class="btn btn-sm" @click="addTimeSlot">+ {{ $t('contract.add_time_slot') }}</button>
        </div>
        <p class="form-hint">{{ $t('contract.slot_overnight_hint') }}</p>
        <div class="slots-list">
          <div v-for="(slot, idx) in contractForm.time_slots" :key="idx" class="slot-card">
            <div class="slot-card-top">
              <input class="form-input" type="text" v-model="slot.label" :aria-label="$t('contract.slot_label')" :placeholder="$t('contract.slot_label')" />
              <button type="button" class="btn-icon-xs slot-remove" @click="removeTimeSlot(idx)" :aria-label="$t('contract.remove_slot')">✕</button>
            </div>
            <div class="slot-card-bottom">
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_start') }}</label>
                <input class="form-input" type="time" v-model="slot.start_time" :aria-label="$t('contract.slot_start')" />
              </div>
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_end') }}</label>
                <input class="form-input" type="time" v-model="slot.end_time" :aria-label="$t('contract.slot_end')" />
              </div>
              <div class="slot-field slot-field-days">
                <label class="slot-field-label">{{ $t('contract.slot_days') }}</label>
                <select class="form-input" v-model="slot.day_type" :aria-label="$t('contract.slot_days')">
                  <option value="all">{{ $t('contract.slot_days_all') }}</option>
                  <option value="weekdays">{{ $t('contract.slot_days_weekdays') }}</option>
                  <option value="weekends">{{ $t('contract.slot_days_weekends') }}</option>
                  <option value="monday">{{ $t('contract.slot_days_monday') }}</option>
                  <option value="tuesday">{{ $t('contract.slot_days_tuesday') }}</option>
                  <option value="wednesday">{{ $t('contract.slot_days_wednesday') }}</option>
                  <option value="thursday">{{ $t('contract.slot_days_thursday') }}</option>
                  <option value="friday">{{ $t('contract.slot_days_friday') }}</option>
                  <option value="saturday">{{ $t('contract.slot_days_saturday') }}</option>
                  <option value="sunday">{{ $t('contract.slot_days_sunday') }}</option>
                </select>
              </div>
              <div v-if="isOvernightSlot(slot)" class="slot-field slot-field-days">
                <label class="slot-field-label">{{ $t('contract.slot_end_day_offset') }}</label>
                <select class="form-input" v-model.number="slot.end_day_offset" :aria-label="$t('contract.slot_end_day_offset')">
                  <option v-for="n in 6" :key="n" :value="n">{{ $t('contract.slot_end_day_offset_' + n) }}</option>
                </select>
              </div>
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_factor') }}</label>
                <input class="form-input" type="number" min="0" step="0.01" v-model="slot.multiplication_factor" :aria-label="$t('contract.slot_factor')" :placeholder="'×'" />
              </div>
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_rate') }}</label>
                <input class="form-input" type="number" min="0" step="0.01" v-model="slot.hourly_rate" :aria-label="$t('contract.slot_rate')" :placeholder="contractForm.currency + '/h'" />
              </div>
            </div>
            <div v-if="slotPreviewReady(slot)" class="slot-preview">
              <span class="slot-preview-label">{{ $t('contract.slot_preview') }}</span>
              <div class="slot-preview-week" role="img" :aria-label="slotPreviewAria(slot)">
                <div
                  v-for="day in slotPreviewDays(slot)"
                  :key="day.key"
                  class="slot-preview-day"
                  :class="{ 'slot-preview-day-active': day.active }"
                >
                  <span class="slot-preview-dow">{{ day.label }}</span>
                  <div class="slot-preview-track" aria-hidden="true">
                    <div
                      v-for="(seg, segIdx) in day.segments"
                      :key="segIdx"
                      class="slot-preview-seg"
                      :style="{ left: seg.left + '%', width: seg.width + '%' }"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="btn" @click="closeContractModal">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveContract" :disabled="!contractForm.name.trim()">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>
  <!-- Add member modal -->
  <BaseModal v-if="showAddMember" :title="$t('customer.add_member')" @close="showAddMember = false">
    <div class="form-group">
      <input class="form-input" v-model="memberSearch" :aria-label="$t('common.search')" :placeholder="$t('common.search') + '…'" />
    </div>
    <div class="user-picker-list">
      <div
        v-for="u in filteredUsers"
        :key="u.id"
        class="user-picker-row"
        :class="{ selected: pendingMemberIds.includes(u.id) }"
        @click="togglePendingMember(u.id)"
      >
        <img :src="resolveAssetUrl(u.avatar_url || u.gravatar_url)" class="member-avatar" alt="" />
        <div class="member-info">
          <span class="member-name">{{ u.display_name || u.username }}</span>
          <span class="member-email">{{ u.email }}</span>
        </div>
        <span v-if="pendingMemberIds.includes(u.id)" class="check-mark">✓</span>
      </div>
      <div v-if="filteredUsers.length === 0" class="empty-state-sm" style="padding:8px 0">
        {{ $t('common.no_results') }}
      </div>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showAddMember = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="pendingMemberIds.length === 0" @click="confirmAddMembers">{{ $t('common.add') }}</button>
    </template>
  </BaseModal>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCustomersStore } from '@/stores/customers'
import { useUIStore } from '@/stores/ui'
import { customersApi } from '@/api/customers'
import { groupsApi } from '@/api/groups'
import { attachmentsApi } from '@/api/attachments'
import { resolveAssetUrl } from '@/api/serverConfig'
import BaseModal from '@/components/common/BaseModal.vue'
import { useDateFormat } from '@/composables/useDateFormat'
import client from '@/api/client'
import { buildSlotPreviewDays, slotPreviewReady as slotPreviewReadyFn } from '@/utils/contractSlotPreview'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const custStore = useCustomersStore()
const ui = useUIStore()

const loading = ref(true)
const detail = ref(null)

const showEdit = ref(false)
const editForm = ref({ name: '', description: '', logo_url: '' })

const { formatDate, dateOnlyFormat } = useDateFormat()

const showAddContract = ref(false)
const editingContract = ref(null)
const contractForm = ref({ name: '', description: '', start_date: '', end_date: '', price_per_hour: null, currency: '€', time_slots: [] })
const emptySlot = () => ({ label: '', start_time: '', end_time: '', day_type: 'all', end_day_offset: 1, multiplication_factor: null, hourly_rate: null })

function isOvernightSlot(slot) {
  return !!(slot.start_time && slot.end_time && slot.end_time <= slot.start_time)
}

function formatSlotTimeRange(slot) {
  if (!slot.start_time || !slot.end_time) return ''
  if (!isOvernightSlot(slot)) return `${slot.start_time}–${slot.end_time}`
  const offset = slot.end_day_offset > 0 ? slot.end_day_offset : 1
  if (offset === 1) return `${slot.start_time}–${slot.end_time}`
  return `${slot.start_time} → +${offset}d ${slot.end_time}`
}

const slotPreviewDowLabels = computed(() => [
  t('contract.slot_preview_dow_mon'),
  t('contract.slot_preview_dow_tue'),
  t('contract.slot_preview_dow_wed'),
  t('contract.slot_preview_dow_thu'),
  t('contract.slot_preview_dow_fri'),
  t('contract.slot_preview_dow_sat'),
  t('contract.slot_preview_dow_sun'),
])

function slotPreviewReady(slot) {
  return slotPreviewReadyFn(slot)
}

function slotPreviewDays(slot) {
  return buildSlotPreviewDays(slot, slotPreviewDowLabels.value)
}

function slotPreviewAria(slot) {
  const range = formatSlotTimeRange(slot)
  const days = slot.day_type && slot.day_type !== 'all'
    ? t('contract.slot_days_' + slot.day_type)
    : t('contract.slot_days_all')
  return t('contract.slot_preview_aria', { range, days })
}
const displayContractStartDate = ref('')
const displayContractEndDate   = ref('')

function _parseContractDate(displayRef, isoKey) {
  const val = displayRef.value.trim()
  if (!val) { contractForm.value[isoKey] = ''; return }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY'), mPos = fmt.indexOf('MM'), dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayRef.value = contractForm.value[isoKey] ? formatDate(contractForm.value[isoKey]) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  contractForm.value[isoKey] = iso
  displayRef.value = formatDate(iso)
}

function parseContractStartDate() { _parseContractDate(displayContractStartDate, 'start_date') }
function parseContractEndDate()   { _parseContractDate(displayContractEndDate,   'end_date')   }

function onContractStartDateChange(e) {
  const iso = e.target.value
  contractForm.value.start_date = iso
  displayContractStartDate.value = iso ? formatDate(iso) : ''
}
function onContractEndDateChange(e) {
  const iso = e.target.value
  contractForm.value.end_date = iso
  displayContractEndDate.value = iso ? formatDate(iso) : ''
}

const editingName = ref(false)
const nameEdit = ref('')

const canManage = computed(() => auth.isAdmin || detail.value?.customer?.my_role === 'admin')

const authUserId = computed(() => auth.user?.id)

// ── Member management ──────────────────────────────────────────────────────
const members = ref([])
const customerGroups = ref([])
const allGroups = ref([])
let allGroupsLoaded = false
const addGroupId = ref('')
const addGroupRole = ref('member')

const groupsNotOnCustomer = computed(() => {
  const assigned = new Set(customerGroups.value.map(g => g.group_id))
  return allGroups.value.filter(g => !assigned.has(g.id))
})

const showAddMember = ref(false)
const allUsers = ref([])
let allUsersLoaded = false
const memberSearch = ref('')
const pendingMemberIds = ref([])

const filteredUsers = computed(() => {
  const q = memberSearch.value.toLowerCase()
  const existingIds = new Set(members.value.map(m => m.user_id))
  return allUsers.value.filter(u => {
    if (existingIds.has(u.id)) return false
    if (!q) return true
    return (u.display_name || '').toLowerCase().includes(q) ||
           u.username.toLowerCase().includes(q) ||
           u.email.toLowerCase().includes(q)
  })
})

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

function groupAvatar(group) {
  return resolveAssetUrl(group?.avatar || '')
}

async function loadMembers() {
  try {
    const { data } = await customersApi.listMembers(custId.value)
    members.value = data || []
  } catch {}
  try {
    const { data } = await groupsApi.listCustomerGroups(custId.value)
    customerGroups.value = data || []
  } catch {}
  if (auth.isAdmin && !allGroupsLoaded) {
    try {
      const { data } = await groupsApi.list()
      allGroups.value = data || []
      allGroupsLoaded = true
    } catch {}
  }
}

async function addGroupToCustomer() {
  if (!addGroupId.value) return
  try {
    await groupsApi.setCustomerAccess(addGroupId.value, custId.value, addGroupRole.value)
    const { data } = await groupsApi.listCustomerGroups(custId.value)
    customerGroups.value = data || []
    addGroupId.value = ''
  } catch {
    ui.error('Failed to add group')
  }
}

async function removeGroupFromCustomer(groupId) {
  try {
    await groupsApi.removeCustomerAccess(groupId, custId.value)
    customerGroups.value = customerGroups.value.filter(g => g.group_id !== groupId)
  } catch {
    ui.error('Failed to remove group')
  }
}

async function openAddMember() {
  memberSearch.value = ''
  pendingMemberIds.value = []
  if (!allUsersLoaded) {
    try {
      const { data } = await client.get('/users')
      allUsers.value = data || []
      allUsersLoaded = true
    } catch {}
  }
  showAddMember.value = true
}

function togglePendingMember(id) {
  const idx = pendingMemberIds.value.indexOf(id)
  if (idx >= 0) pendingMemberIds.value.splice(idx, 1)
  else pendingMemberIds.value.push(id)
}

async function confirmAddMembers() {
  const newMembers = [
    ...members.value.map(m => ({ user_id: m.user_id, role: m.role })),
    ...pendingMemberIds.value.map(id => ({ user_id: id, role: 'member' })),
  ]
  try {
    await customersApi.setMembers(custId.value, newMembers)
    await loadMembers()
    showAddMember.value = false
  } catch {
    ui.error('Failed to add members')
  }
}

async function removeMember(userId) {
  const newMembers = members.value
    .filter(m => m.user_id !== userId)
    .map(m => ({ user_id: m.user_id, role: m.role }))
  try {
    await customersApi.setMembers(custId.value, newMembers)
    await loadMembers()
  } catch {
    ui.error('Failed to remove member')
  }
}

async function setMemberRole(userId, role) {
  const newMembers = members.value.map(m =>
    m.user_id === userId ? { user_id: m.user_id, role } : { user_id: m.user_id, role: m.role }
  )
  try {
    await customersApi.setMembers(custId.value, newMembers)
    await loadMembers()
  } catch {
    ui.error('Failed to update role')
  }
}

const custId = computed(() => Number(route.params.id))

onMounted(() => load())
watch(custId, () => load())

async function load() {
  loading.value = true
  try {
    const { data } = await customersApi.get(custId.value)
    detail.value = data
    if (data?.customer?.my_role === 'admin' || auth.isAdmin) {
      await loadMembers()
    }
  } catch {
    ui.error('Customer not found')
    router.push('/customers')
  } finally {
    loading.value = false
  }
}

async function toggleFav() {
  await custStore.toggleFavorite(custId.value)
  detail.value.customer.is_favorite = !detail.value.customer.is_favorite
}

function startEditName() {
  nameEdit.value = detail.value.customer.name
  editingName.value = true
}

async function saveNameEdit() {
  if (!nameEdit.value.trim()) return
  try {
    await customersApi.update(custId.value, { name: nameEdit.value.trim(), description: detail.value.customer.description, logo_url: detail.value.customer.logo_url })
    detail.value.customer.name = nameEdit.value.trim()
    editingName.value = false
    await custStore.fetchCustomers()
  } catch {
    ui.error('Failed to update')
  }
}

function openEdit() {
  editForm.value = { name: detail.value.customer.name, description: detail.value.customer.description, logo_url: detail.value.customer.logo_url }
  showEdit.value = true
}

async function onLogoFileSelected(e) {
  const file = e.target.files[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    editForm.value.logo_url = data.url
  } catch {
    ui.error('Failed to upload image')
  }
}

async function saveEdit() {
  try {
    await customersApi.update(custId.value, editForm.value)
    await load()
    await custStore.fetchCustomers()
    showEdit.value = false
    ui.success('Saved')
  } catch {
    ui.error('Failed to save')
  }
}

async function doDelete() {
  if (!confirm(this.$t?.('customer.delete_confirm') || 'Delete this customer?')) return
  try {
    await customersApi.delete(custId.value)
    await custStore.fetchCustomers()
    router.push('/customers')
  } catch {
    ui.error('Failed to delete')
  }
}

function editContract(grp) {
  editingContract.value = grp
  contractForm.value = {
    name: grp.name,
    description: grp.description || '',
    start_date: grp.start_date ? grp.start_date.split('T')[0] : '',
    end_date:   grp.end_date   ? grp.end_date.split('T')[0]   : '',
    price_per_hour: grp.price_per_hour != null ? grp.price_per_hour : null,
    currency:       grp.currency || '€',
    time_slots: (grp.time_slots || []).map(s => ({
      label:                s.label || '',
      start_time:           s.start_time || '',
      end_time:             s.end_time || '',
      day_type:             s.day_type || 'all',
      end_day_offset:       s.end_day_offset > 0 ? s.end_day_offset : 1,
      multiplication_factor: s.multiplication_factor != null ? s.multiplication_factor : null,
      hourly_rate:           s.hourly_rate != null ? s.hourly_rate : null,
    })),
  }
  displayContractStartDate.value = contractForm.value.start_date ? formatDate(contractForm.value.start_date) : ''
  displayContractEndDate.value   = contractForm.value.end_date   ? formatDate(contractForm.value.end_date)   : ''
}

function addTimeSlot() {
  contractForm.value.time_slots.push(emptySlot())
}

function removeTimeSlot(idx) {
  contractForm.value.time_slots.splice(idx, 1)
}

function closeContractModal() {
  showAddContract.value = false
  editingContract.value = null
  contractForm.value = { name: '', description: '', start_date: '', end_date: '', price_per_hour: null, currency: '€', time_slots: [] }
  displayContractStartDate.value = ''
  displayContractEndDate.value   = ''
}

async function saveContract() {
  const payload = {
    name:           contractForm.value.name,
    description:    contractForm.value.description,
    start_date:     contractForm.value.start_date || '',
    end_date:       contractForm.value.end_date   || '',
    price_per_hour: contractForm.value.price_per_hour,
    currency:       contractForm.value.currency || '€',
    time_slots:     contractForm.value.time_slots
      .filter(s => s.start_time && s.end_time)
      .map(s => ({
        label:                s.label || '',
        start_time:           s.start_time,
        end_time:             s.end_time,
        day_type:             s.day_type || 'all',
        end_day_offset:       isOvernightSlot(s) ? (s.end_day_offset > 0 ? s.end_day_offset : 1) : 0,
        multiplication_factor: s.multiplication_factor !== null && s.multiplication_factor !== '' ? parseFloat(s.multiplication_factor) : null,
        hourly_rate:           s.hourly_rate !== null && s.hourly_rate !== '' ? parseFloat(s.hourly_rate) : null,
      })),
  }
  try {
    if (editingContract.value) {
      await customersApi.updateContract(custId.value, editingContract.value.id, payload)
    } else {
      await customersApi.createContract(custId.value, payload)
    }
    await load()
    closeContractModal()
  } catch {
    ui.error('Failed to save contract')
  }
}

async function deleteContract(grp) {
  if (!confirm('Delete contract "' + grp.name + '"?')) return
  try {
    await customersApi.deleteContract(custId.value, grp.id)
    await load()
  } catch {
    ui.error('Failed to delete contract')
  }
}

</script>

<style scoped>
.customer-detail-page {
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}

.loading { color: var(--color-text-muted); padding: 48px; text-align: center; }

.cust-header {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--color-border);
}

.cust-logo-wrap { flex-shrink: 0; }
.cust-logo { width: 64px; height: 64px; border-radius: 12px; object-fit: contain; }
.cust-logo-placeholder {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
}

.cust-info { flex: 1; }
.cust-name-row { display: flex; align-items: center; gap: 8px; }
.cust-name { font-size: 22px; font-weight: 700; margin: 0; }
.name-input { font-size: 18px; }
.cust-desc { margin: 6px 0 0; color: var(--color-text-muted); font-size: 14px; }

.cust-actions { display: flex; gap: 8px; align-items: center; flex-shrink: 0; }
.star-btn {
  background: none;
  border: none;
  font-size: 22px;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 2px 4px;
}
.star-btn.active { color: #f59e0b; }

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.section-header-row h2 { font-size: 18px; font-weight: 600; margin: 0; }

.contracts-section { display: flex; flex-direction: column; gap: 16px; }

.contract-block {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 16px;
}

.contract-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.contract-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.contract-icon { font-size: 16px; }

.contract-dates {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-left: 4px;
}

.contract-rate {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary);
  margin-left: 8px;
  white-space: nowrap;
}

.contract-actions { display: flex; gap: 4px; }

.contract-desc {
  font-size: 13px;
  color: var(--color-text-muted);
  margin: 0 0 12px;
}

.projects-mini-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.project-mini-tile {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  text-decoration: none;
  font-size: 13px;
  color: var(--color-text);
  background: var(--color-bg);
  transition: border-color .12s;
}
.project-mini-tile:hover { border-color: var(--color-primary); text-decoration: none; }

.proj-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.proj-avatar {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}

.unassigned-block { opacity: .85; }

.empty-state { color: var(--color-text-muted); padding: 32px; text-align: center; }
.empty-state-sm { color: var(--color-text-muted); font-size: 12px; padding: 4px 0; }

.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-muted);
  font-size: 14px;
  padding: 2px 6px;
  border-radius: 4px;
}
.icon-btn:hover { background: var(--color-bg); color: var(--color-text); }
.icon-danger:hover { color: var(--color-danger, #ef4444); }

.btn { padding: 6px 14px; border-radius: 6px; border: 1px solid var(--color-border); background: var(--color-surface); color: var(--color-text); cursor: pointer; font-size: 13px; }
.btn-primary { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.btn-danger { background: var(--color-danger, #ef4444); color: #fff; border-color: var(--color-danger, #ef4444); }
.btn-sm { height: 28px; padding: 0 10px; font-size: 12px; }
.btn:disabled { opacity: .5; cursor: not-allowed; }

.form-group { margin-bottom: 12px; }
.form-label { display: block; font-size: 12px; font-weight: 600; margin-bottom: 4px; color: var(--color-text-muted); }
.form-input { width: 100%; padding: 8px 10px; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-bg); color: var(--color-text); font-size: 14px; box-sizing: border-box; }
.detail-row { display: flex; gap: 12px; }
.half { flex: 1; }

.date-input-row { display: flex; align-items: center; gap: 6px; }
.date-input-row .form-input { flex: 1; }
.picker-wrap { position: relative; display: inline-flex; cursor: pointer; }
.date-picker-overlay { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
.btn-icon-xs {
  background: none; border: none; cursor: pointer; color: var(--color-text-muted);
  padding: 2px 4px; font-size: 13px; line-height: 1; border-radius: 3px; flex-shrink: 0;
}
.btn-icon-xs:hover { background: var(--color-bg); color: var(--color-text); }

.members-section { margin-top: 32px; }

.members-list { display: flex; flex-direction: column; gap: 8px; }

.member-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.member-avatar { width: 32px; height: 32px; border-radius: 50%; object-fit: cover; flex-shrink: 0; }
.group-avatar {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}

.member-info { flex: 1; min-width: 0; }
.member-name { display: block; font-size: 13px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.member-email { display: block; font-size: 11px; color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.role-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 99px;
  flex-shrink: 0;
}
.role-admin { background: #0ea5e9; color: #fff; }
.role-member { background: var(--color-bg); color: var(--color-text-muted); border: 1px solid var(--color-border); }

.member-actions { display: flex; gap: 4px; flex-shrink: 0; }

.user-picker-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  margin-top: 4px;
}

.user-picker-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
}
.user-picker-row:last-child { border-bottom: none; }
.user-picker-row:hover { background: var(--color-bg); }
.user-picker-row.selected { background: color-mix(in srgb, var(--color-primary) 8%, transparent); }

.check-mark { font-size: 14px; color: var(--color-primary); font-weight: 700; flex-shrink: 0; }

/* Time slots — contract display */
.slot-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 6px;
}
.slot-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: var(--color-text-muted);
}
.slot-time { font-variant-numeric: tabular-nums; white-space: nowrap; }
.slot-days { font-size: 11px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 4px; padding: 1px 5px; white-space: nowrap; }
.slot-label { color: var(--color-text); }
.slot-factor { font-weight: 600; color: var(--color-primary); }
.slot-rate { font-weight: 600; color: var(--color-primary); }

/* Time slots — contract form */
.slots-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.slots-header .form-label { margin-bottom: 0; }
.form-hint { font-size: 12px; color: var(--color-text-muted); margin: 0 0 8px; }
.slots-list { display: flex; flex-direction: column; gap: 8px; }
.slot-card {
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.slot-card-top {
  display: flex;
  gap: 6px;
  align-items: center;
}
.slot-card-top .form-input { flex: 1; }
.slot-card-bottom {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.slot-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}
.slot-field-days { flex: 1; min-width: 110px; }
.slot-field .form-input { padding: 5px 7px; font-size: 13px; width: 90px; }
.slot-field-days .form-input { width: 100%; }
.slot-field-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .04em;
  color: var(--color-text-muted);
}
.slot-remove { color: var(--color-danger, #ef4444); flex-shrink: 0; }
.slot-remove:hover { background: color-mix(in srgb, var(--color-danger, #ef4444) 10%, transparent); }

.slot-preview {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-top: 4px;
  border-top: 1px dashed var(--color-border);
}
.slot-preview-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .04em;
  color: var(--color-text-muted);
}
.slot-preview-week {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
}
.slot-preview-day {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.slot-preview-dow {
  font-size: 10px;
  text-align: center;
  color: var(--color-text-muted);
}
.slot-preview-day-active .slot-preview-dow {
  color: var(--color-text);
  font-weight: 600;
}
.slot-preview-track {
  position: relative;
  height: 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}
.slot-preview-seg {
  position: absolute;
  top: 0;
  bottom: 0;
  background: color-mix(in srgb, var(--color-primary) 55%, transparent);
  border-radius: 2px;
  min-width: 2px;
}
</style>
