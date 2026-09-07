<template>
  <BaseModal
    :title="entry ? $t('timeTracking.edit_entry') : $t('timeTracking.new_entry')"
    style="--modal-width: 480px"
    @close="$emit('close')"
  >
    <div class="form-group">
      <label class="form-label" for="te-customer">{{ $t('timeTracking.customer') }}</label>
      <select id="te-customer" class="form-input" v-model="form.customer_id" @change="form.project_id = null">
        <option :value="null">{{ $t('timeTracking.no_customer') }}</option>
        <option v-for="c in allCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="te-project">{{ $t('timeTracking.project') }}</label>
      <select id="te-project" class="form-input" v-model="form.project_id">
        <option :value="null">{{ $t('timeTracking.no_project') }}</option>
        <option v-for="p in projectOptions" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="te-activity">{{ $t('timeTracking.activity') }}</label>
      <input id="te-activity" class="form-input" type="text" v-model="form.description" />
    </div>
    <div class="form-group">
      <span class="form-label" id="te-date-label">{{ $t('timeTracking.date') }}</span>
      <DatePicker :label="$t('timeTracking.date')" v-model="form.date" />
    </div>
    <div class="form-row">
      <div class="form-group">
        <label class="form-label" for="te-start">{{ $t('timeTracking.start_time') }}</label>
        <input
          id="te-start" class="form-input" type="text" maxlength="5"
          v-model="form.start_time" placeholder="09:00"
          @input="onTimeInput('start_time', $event)"
        />
      </div>
      <div class="form-group">
        <label class="form-label" for="te-end">{{ $t('timeTracking.end_time') }}</label>
        <input
          id="te-end" class="form-input" type="text" maxlength="5"
          v-model="form.end_time" placeholder="17:00"
          @input="onTimeInput('end_time', $event)"
        />
      </div>
    </div>
    <div class="form-group" v-if="locationOptions.length">
      <label class="form-label" for="te-location">{{ $t('timeTracking.location') }}</label>
      <select id="te-location" class="form-input" v-model="locationSel" @change="applyLocationDistance">
        <option value="">{{ $t('timeTracking.choose_location') }}</option>
        <option v-for="l in locationOptions" :key="l.id" :value="l.id">{{ formatLocationLabel(l) }} — {{ l.travel_distance }} {{ distanceUnit }}</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="te-distance">{{ $t('timeTracking.distance') }} ({{ distanceUnit }})</label>
      <input id="te-distance" class="form-input" type="number" min="0" step="0.1" v-model.number="form.distance" />
    </div>
    <p v-if="timeError" class="field-error" role="alert">{{ timeError }}</p>

    <template #footer>
      <button v-if="entry" type="button" class="btn btn-danger" @click="$emit('delete', entry)">
        {{ $t('common.delete') }}
      </button>
      <button type="button" class="btn" @click="$emit('close')">{{ $t('common.cancel') }}</button>
      <button type="button" class="btn btn-primary" :disabled="!formValid" @click="onSave">
        {{ $t('common.save') }}
      </button>
    </template>
  </BaseModal>
</template>

<script setup>
import { reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseModal from '@/components/common/BaseModal.vue'
import DatePicker from '@/components/common/DatePicker.vue'
import { useAuthStore } from '@/stores/auth'
import { customersApi } from '@/api/customers'
import { filterProjectsForCustomer } from '@/utils/projectFilter'
import { parseWallClock, fmtWallClock, wallClockSpanMinutes } from '@/utils/shiftTimeEntries'

const { t } = useI18n()
const auth = useAuthStore()
const distanceUnit = computed(() => auth.user?.distance_unit || 'km')

const props = defineProps({
  entry: { type: Object, default: null },      // existing TimeEntry when editing
  prefill: { type: Object, default: null },     // { date, start_time, end_time } when creating
  allCustomers: { type: Array, default: () => [] },
  allProjects: { type: Array, default: () => [] },
  ttCustomers: { type: Array, default: () => [] },
  ttProjects: { type: Array, default: () => [] },
  projects: { type: Array, default: () => [] },
})
const emit = defineEmits(['save', 'delete', 'close'])

const source = props.entry || props.prefill || {}
const form = reactive({
  customer_id: source.customer_id ?? null,
  project_id: source.project_id ?? null,
  description: source.description ?? '',
  date: source.date ? source.date.slice(0, 10) : '',
  start_time: source.start_time ?? '',
  end_time: source.end_time ?? '',
  distance: source.distance ?? null,
  location_id: source.location_id ?? null,
})

const projectOptions = computed(() => filterProjectsForCustomer(form.customer_id, {
  allProjects: props.allProjects, ttCustomers: props.ttCustomers, ttProjects: props.ttProjects, projects: props.projects,
}))

// Locations (with a standard travel distance) for the selected customer — picking one
// fills the distance field and is remembered on the entry via form.location_id; the
// distance value itself stays freely editable afterward.
const locationsByCustomer = ref({})
const locationSel = ref(source.location_id ?? '')

async function loadLocations(customerId) {
  if (!customerId || locationsByCustomer.value[customerId]) return
  try {
    const { data } = await customersApi.listLocations(customerId)
    locationsByCustomer.value = { ...locationsByCustomer.value, [customerId]: data }
  } catch {}
}
if (form.customer_id) loadLocations(form.customer_id)

watch(() => form.customer_id, (id) => {
  locationSel.value = ''
  form.location_id = null
  loadLocations(id)
})

const locationOptions = computed(() =>
  (locationsByCustomer.value[form.customer_id] || []).filter(l => l.travel_distance != null),
)

function formatLocationLabel(loc) {
  const parts = [loc.address_line1, loc.city].filter(Boolean)
  return parts.length ? parts.join(', ') : `#${loc.id}`
}

function applyLocationDistance() {
  const loc = locationOptions.value.find(l => l.id === locationSel.value)
  if (loc) {
    form.distance = loc.travel_distance
    form.location_id = loc.id
  }
}

// Auto-insert a colon after 2 digits, mirroring the existing time-popup input UX.
function onTimeInput(field, event) {
  let val = event.target.value.replace(/[^\d:]/g, '')
  const colon = val.indexOf(':')
  if (colon >= 0) val = val.slice(0, colon + 1) + val.slice(colon + 1).replace(/:/g, '')
  if (val.length === 2 && !val.includes(':')) val = val + ':'
  if (val !== event.target.value) event.target.value = val
  form[field] = val
}

const startMinutes = computed(() => parseWallClock(form.start_time))
const endMinutes = computed(() => parseWallClock(form.end_time))
// Overnight entries (e.g. 19:00 → 07:00) are valid — an end time earlier than or
// equal to the start time means the shift continues past midnight, not an error.
const spanMinutes = computed(() => wallClockSpanMinutes(startMinutes.value, endMinutes.value))

const timeError = computed(() => {
  if (!form.start_time || !form.end_time) return ''
  if (startMinutes.value < 0 || endMinutes.value < 0) return ''
  if (spanMinutes.value == null) return t('timeTracking.time_range_equal_invalid')
  return ''
})

const formValid = computed(() =>
  !!form.date && startMinutes.value >= 0 && endMinutes.value >= 0 && spanMinutes.value != null,
)

function onSave() {
  if (!formValid.value) return
  emit('save', {
    id: props.entry?.id,
    customer_id: form.customer_id || null,
    project_id: form.project_id || null,
    contract_id: props.entry?.contract_id ?? null,
    date: form.date,
    minutes: spanMinutes.value,
    description: form.description,
    is_holiday: props.entry?.is_holiday || false,
    start_time: fmtWallClock(startMinutes.value),
    end_time: fmtWallClock(endMinutes.value),
    distance: form.distance || null,
    location_id: form.location_id || null,
  })
}
</script>

<style scoped>
.field-error { margin-top: 4px; font-size: 12px; color: var(--color-danger); }
</style>
