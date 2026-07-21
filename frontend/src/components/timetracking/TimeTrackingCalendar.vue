<template>
  <div class="tt-calendar">
    <p class="cal-hint sr-only">{{ $t('timeTracking.calendar_keyboard_hint') }}</p>

    <div class="cal-toolbar">
      <div class="cal-zoom" role="group" :aria-label="$t('timeTracking.calendar_zoom')">
        <button
          type="button" class="btn btn-ghost btn-sm cal-zoom-btn"
          :disabled="zoomIndex <= 0"
          :title="$t('timeTracking.calendar_zoom_out')" :aria-label="$t('timeTracking.calendar_zoom_out')"
          @click="zoomOut"
        >−</button>
        <button
          type="button" class="btn btn-ghost btn-sm cal-zoom-btn"
          :disabled="zoomIndex >= ZOOM_LEVELS.length - 1"
          :title="$t('timeTracking.calendar_zoom_in')" :aria-label="$t('timeTracking.calendar_zoom_in')"
          @click="zoomIn"
        >+</button>
      </div>
    </div>

    <TimeTrackingCalendarWeekGrid
      v-if="viewGranularity === 'week'"
      :week-days="weekDays"
      :entries="entries"
      :px-per-hour="pxPerHour"
      :customer-name="customerNameFor"
      :project-name="projectNameFor"
      :customer-color="customerColorFor"
      :read-only="readOnly"
      @slot-click="onSlotClick"
      @slot-contextmenu="onSlotContextMenu"
      @block-contextmenu="onBlockContextMenu"
      @block-open="openEditModal"
      @block-move="(payload) => $emit('move-entry', payload)"
      @block-resize="(payload) => $emit('resize-entry', payload)"
    />
    <!-- Month view is a future granularity; the emit payload shapes above are already
         month-ready (newDate/newStartTime/newEndTime/newMinutes), so adding a
         TimeTrackingCalendarMonthGrid sibling here won't require changes elsewhere. -->

    <ContextMenu
      v-if="ctxMenu"
      :x="ctxMenu.x"
      :y="ctxMenu.y"
      :items="ctxMenu.items"
      @select="onCtxSelect"
      @close="ctxMenu = null"
    />

    <TimeEntryModal
      v-if="modalState"
      :entry="modalState.entry"
      :prefill="modalState.prefill"
      :all-customers="allCustomers"
      :all-projects="allProjects"
      :tt-customers="ttCustomers"
      :tt-projects="ttProjects"
      :projects="projects"
      @save="onModalSave"
      @delete="onModalDelete"
      @close="modalState = null"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import TimeTrackingCalendarWeekGrid from './TimeTrackingCalendarWeekGrid.vue'
import ContextMenu from '@/components/common/ContextMenu.vue'
import TimeEntryModal from './TimeEntryModal.vue'
import { parseWallClock, fmtWallClock } from '@/utils/shiftTimeEntries'
import { assignCustomerColors, NO_CUSTOMER_COLOR } from '@/utils/calendarColors'

const ZOOM_STORAGE_KEY = 'tt_calendar_zoom'
const ZOOM_LEVELS = [20, 30, 45, 60, 90, 120, 160] // px per hour

function loadZoomIndex() {
  const stored = Number(localStorage.getItem(ZOOM_STORAGE_KEY))
  const idx = ZOOM_LEVELS.indexOf(stored)
  return idx >= 0 ? idx : ZOOM_LEVELS.indexOf(60)
}

const props = defineProps({
  entries: { type: Array, default: () => [] },
  weekDays: { type: Array, required: true },
  allCustomers: { type: Array, default: () => [] },
  allProjects: { type: Array, default: () => [] },
  ttCustomers: { type: Array, default: () => [] },
  ttProjects: { type: Array, default: () => [] },
  projects: { type: Array, default: () => [] },
  viewGranularity: { type: String, default: 'week' }, // 'week' | 'month' (month not yet implemented)
  readOnly: { type: Boolean, default: false },
})
const emit = defineEmits(['save-entry', 'delete-entry', 'move-entry', 'resize-entry'])

const { t } = useI18n()

const ctxMenu = ref(null)
const modalState = ref(null) // { entry } | { prefill }

const zoomIndex = ref(loadZoomIndex())
const pxPerHour = computed(() => ZOOM_LEVELS[zoomIndex.value])

function zoomIn() {
  if (zoomIndex.value >= ZOOM_LEVELS.length - 1) return
  zoomIndex.value++
  localStorage.setItem(ZOOM_STORAGE_KEY, String(ZOOM_LEVELS[zoomIndex.value]))
}

function zoomOut() {
  if (zoomIndex.value <= 0) return
  zoomIndex.value--
  localStorage.setItem(ZOOM_STORAGE_KEY, String(ZOOM_LEVELS[zoomIndex.value]))
}

function customerNameFor(entry) {
  return entry.customer?.name || ''
}

function projectNameFor(entry) {
  return entry.project?.name || ''
}

const customerColorMap = computed(() => assignCustomerColors(props.allCustomers))

function customerColorFor(entry) {
  if (entry.customer_id == null) return NO_CUSTOMER_COLOR
  return customerColorMap.value.get(entry.customer_id) || NO_CUSTOMER_COLOR
}

function onSlotClick({ date, startTime, endTime }) {
  if (props.readOnly) return
  openCreateModal({ date, startTime, endTime })
}

function onSlotContextMenu({ x, y, date, startTime }) {
  if (props.readOnly) return
  ctxMenu.value = {
    x, y,
    items: [{ key: 'create', label: t('timeTracking.new_entry') }],
    target: { date, startTime },
  }
}

function onBlockContextMenu({ x, y, entry }) {
  if (props.readOnly) return
  ctxMenu.value = {
    x, y,
    items: [
      { key: 'edit', label: t('common.edit') },
      { key: 'delete', label: t('common.delete'), danger: true },
    ],
    target: entry,
  }
}

function onCtxSelect(key) {
  const target = ctxMenu.value?.target
  ctxMenu.value = null
  if (!target) return
  if (key === 'create') openCreateModal(target)
  else if (key === 'edit') openEditModal(target)
  else if (key === 'delete') emit('delete-entry', target)
}

function openEditModal(entry) {
  if (props.readOnly) return
  modalState.value = { entry, prefill: null }
}

function openCreateModal({ date, startTime, endTime }) {
  const finalEndTime = endTime || fmtWallClock(Math.min(24 * 60 - 1, parseWallClock(startTime) + 60))
  modalState.value = { entry: null, prefill: { date, start_time: startTime, end_time: finalEndTime } }
}

function onModalSave(payload) {
  emit('save-entry', payload)
  modalState.value = null
}

function onModalDelete(entry) {
  emit('delete-entry', entry)
  modalState.value = null
}
</script>

<style scoped>
.tt-calendar {
  display: flex;
  flex-direction: column;
  height: 70vh;
  min-height: 420px;
}

.cal-toolbar {
  display: flex;
  justify-content: flex-end;
  padding: 0 0 6px;
}

.cal-zoom { display: flex; gap: 2px; }

.cal-zoom-btn {
  width: 24px;
  height: 24px;
  padding: 0;
  font-size: 14px;
  line-height: 1;
}
</style>
