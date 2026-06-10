<template>
  <BaseModal
    :title="$t('contract.rates_overview_title')"
    width="min(95vw, 900px)"
    @close="$emit('close')"
  >
    <div v-if="loading" class="crm-loading">{{ $t('common.loading') }}</div>

    <div v-else-if="!items.length" class="crm-empty">
      {{ $t('contract.rates_overview_empty') }}
    </div>

    <div v-else class="crm-body">
      <section
        v-for="cust in items"
        :key="cust.customer_id"
        class="crm-customer"
        :aria-labelledby="'crm-cust-' + cust.customer_id"
      >
        <h4 :id="'crm-cust-' + cust.customer_id" class="crm-cust-name">{{ cust.customer_name }}</h4>

        <div
          v-for="contract in cust.contracts"
          :key="contract.id"
          class="crm-contract"
        >
          <div class="crm-contract-header">
            <span class="crm-contract-name">{{ contract.name }}</span>
            <span
              v-if="contract.price_per_hour != null"
              class="crm-base-rate"
            >
              {{ $t('contract.base_rate') }}: {{ contract.price_per_hour }} {{ contract.currency }}
            </span>
          </div>

          <table class="crm-table" :aria-label="contract.name + ' ' + $t('contract.time_slots')">
            <thead>
              <tr>
                <th scope="col">{{ $t('contract.slot_label') }}</th>
                <th scope="col">{{ $t('contract.slot_start') }} – {{ $t('contract.slot_end') }}</th>
                <th scope="col">{{ $t('contract.slot_days') }}</th>
                <th scope="col">{{ $t('contract.slot_factor') }}</th>
                <th scope="col">{{ $t('contract.slot_rate') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="slotItem in contract.time_slots" :key="slotItem.id">
                <td>{{ slotItem.label || '—' }}</td>
                <td>
                  {{ slotItem.start_time }} – {{ slotItem.end_time }}
                  <span
                    v-if="isOvernight(slotItem)"
                    class="crm-overnight"
                    :title="$t('contract.slot_overnight_hint')"
                  >
                    ({{ slotItem.end_day_offset > 0 ? $t('contract.slot_end_day_offset_' + slotItem.end_day_offset) : $t('contract.slot_overnight') }})
                  </span>
                </td>
                <td>{{ $t('contract.slot_days_' + slotItem.day_type) }}</td>
                <td>{{ slotItem.multiplication_factor != null ? '×' + slotItem.multiplication_factor : '—' }}</td>
                <td>{{ slotItem.hourly_rate != null ? slotItem.hourly_rate + ' ' + contract.currency : '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </BaseModal>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { customersApi } from '@/api/customers'
import BaseModal from '@/components/common/BaseModal.vue'

defineEmits(['close'])

const loading = ref(true)
const items = ref([])

function isOvernight(slotItem) {
  if (!slotItem.start_time || !slotItem.end_time) return false
  return slotItem.end_time < slotItem.start_time
}

onMounted(async () => {
  try {
    const res = await customersApi.listAllContractRates()
    items.value = res.data
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.crm-loading,
.crm-empty {
  color: var(--color-text-muted);
  text-align: center;
  padding: 32px 16px;
}

.crm-body {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.crm-customer {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.crm-cust-name {
  font-size: 15px;
  font-weight: 700;
  margin: 0;
  padding-bottom: 4px;
  border-bottom: 2px solid var(--color-primary);
  color: var(--color-text);
}

.crm-contract {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-left: 8px;
}

.crm-contract-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  flex-wrap: wrap;
}

.crm-contract-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.crm-base-rate {
  font-size: 12px;
  color: var(--color-text-muted);
}

.crm-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.crm-table th,
.crm-table td {
  padding: 6px 10px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.crm-table thead th {
  font-weight: 600;
  color: var(--color-text-muted);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: var(--color-surface-raised, var(--color-surface));
}

.crm-table tbody tr:last-child td {
  border-bottom: none;
}

.crm-table tbody tr:hover td {
  background: var(--color-surface-hover, var(--color-surface));
}

.crm-overnight {
  font-size: 11px;
  color: var(--color-text-muted);
  white-space: nowrap;
}
</style>
