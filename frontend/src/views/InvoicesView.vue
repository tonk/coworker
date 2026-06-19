<template>
  <main class="invoices-main">
    <div class="invoices-container">
      <div class="page-header">
        <h1>{{ $t('invoice.invoices') }}</h1>
      </div>

      <!-- Revenue summary -->
      <div v-if="!loading && allInvoices.length > 0" class="summary-bar">
        <div class="summary-card">
          <span class="summary-label">{{ $t('invoice.total_invoiced') }}</span>
          <span class="summary-value">{{ primaryCurrency }} {{ summary.totalInvoiced.toFixed(2) }}</span>
        </div>
        <div class="summary-card">
          <span class="summary-label">{{ $t('invoice.outstanding') }}</span>
          <span class="summary-value">{{ primaryCurrency }} {{ summary.outstanding.toFixed(2) }}</span>
        </div>
        <div class="summary-card summary-card--overdue" :class="{ 'has-value': summary.overdue > 0 }">
          <span class="summary-label">{{ $t('invoice.overdue_amount') }}</span>
          <span class="summary-value">{{ primaryCurrency }} {{ summary.overdue.toFixed(2) }}</span>
        </div>
        <div class="summary-card summary-card--paid">
          <span class="summary-label">{{ $t('invoice.paid_total') }}</span>
          <span class="summary-value">{{ primaryCurrency }} {{ summary.paid.toFixed(2) }}</span>
        </div>
      </div>

      <div class="filter-bar">
        <select class="form-input filter-select" v-model="filterCustomer" :aria-label="$t('invoice.customer')">
          <option value="">{{ $t('invoice.all_customers') }}</option>
          <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select class="form-input filter-select" v-model="filterStatus" :aria-label="$t('invoice.filter_status')">
          <option value="">{{ $t('invoice.all_statuses') }}</option>
          <option value="draft">{{ $t('invoice.status_draft') }}</option>
          <option value="sent">{{ $t('invoice.status_sent') }}</option>
          <option value="overdue">{{ $t('invoice.overdue') }}</option>
          <option value="paid">{{ $t('invoice.status_paid') }}</option>
          <option value="credit_note">{{ $t('invoice.status_credit_note') }}</option>
        </select>
      </div>

      <div v-if="loading" class="empty-state">{{ $t('common.loading') }}</div>
      <div v-else-if="invoices.length === 0" class="empty-state">{{ $t('invoice.no_invoices') }}</div>
      <div v-else class="invoice-table-wrap">
        <table class="invoice-table">
          <thead>
            <tr>
              <th>{{ $t('invoice.invoice_number') }}</th>
              <th>{{ $t('invoice.customer') }}</th>
              <th>{{ $t('invoice.period') }}</th>
              <th>{{ $t('invoice.filter_status') }}</th>
              <th class="col-right">{{ $t('invoice.total') }}</th>
              <th>{{ $t('invoice.due_date') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="inv in invoices" :key="inv.id" :class="{ 'row-overdue': isOverdue(inv) }">
              <td class="col-number">
                <RouterLink :to="`/customers/${inv.customer_id}`" class="inv-link">
                  {{ inv.invoice_number }}
                </RouterLink>
                <span v-if="inv.credited_invoice_id" class="inv-credit-mark" :title="$t('invoice.credits_invoice') + ' #' + inv.credited_invoice_id">↩</span>
              </td>
              <td>
                <RouterLink v-if="inv.customer" :to="`/customers/${inv.customer_id}`" class="customer-link">
                  {{ inv.customer.name }}
                </RouterLink>
                <span v-else class="text-muted">#{{ inv.customer_id }}</span>
              </td>
              <td class="col-period">{{ formatDate(inv.period_start) }} – {{ formatDate(inv.period_end) }}</td>
              <td>
                <span v-if="isOverdue(inv)" class="status-badge status-overdue">{{ $t('invoice.overdue') }}</span>
                <span v-else :class="['status-badge', `status-${inv.status}`]">{{ $t(`invoice.status_${inv.status}`) }}</span>
              </td>
              <td class="col-right col-total">{{ inv.currency }} {{ inv.total.toFixed(2) }}</td>
              <td class="col-date">
                <span v-if="inv.due_date" :class="{ 'overdue-date': isOverdue(inv) }">{{ formatDate(inv.due_date) }}</span>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="col-actions">
                <a :href="pdfUrl(inv)" target="_blank" class="btn btn-sm" :aria-label="$t('invoice.download_pdf')">PDF</a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </main>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useCustomersStore } from '@/stores/customers'
import { customersApi } from '@/api/customers'
import { useDateFormat } from '@/composables/useDateFormat'
import { getServerUrl } from '@/api/serverConfig'

const { t: $t } = useI18n()
const auth = useAuthStore()
const customersStore = useCustomersStore()
const { formatDate } = useDateFormat()

const allInvoices = ref([])
const loading = ref(false)
const filterCustomer = ref('')
const filterStatus = ref('')

const today = new Date()
today.setHours(0, 0, 0, 0)

function isOverdue(inv) {
  return inv.status === 'sent' && !!inv.due_date && new Date(inv.due_date) < today
}

const customers = computed(() => [...customersStore.customers].sort((a, b) => a.name.localeCompare(b.name)))

const invoices = computed(() => {
  if (!filterStatus.value) return allInvoices.value
  if (filterStatus.value === 'overdue') return allInvoices.value.filter(isOverdue)
  return allInvoices.value.filter(inv => inv.status === filterStatus.value)
})

const primaryCurrency = computed(() => {
  const currencies = allInvoices.value.map(inv => inv.currency).filter(Boolean)
  if (!currencies.length) return '€'
  const freq = {}
  for (const c of currencies) freq[c] = (freq[c] || 0) + 1
  return Object.entries(freq).sort((a, b) => b[1] - a[1])[0][0]
})

const summary = computed(() => {
  let totalInvoiced = 0, outstanding = 0, overdue = 0, paid = 0
  for (const inv of allInvoices.value) {
    if (inv.status === 'draft' || inv.status === 'credit_note') continue
    if (inv.status === 'sent') {
      totalInvoiced += inv.total
      if (isOverdue(inv)) overdue += inv.total
      else outstanding += inv.total
    } else if (inv.status === 'paid') {
      totalInvoiced += inv.total
      paid += inv.total
    }
  }
  return { totalInvoiced, outstanding, overdue, paid }
})

async function load() {
  loading.value = true
  try {
    const params = {}
    if (filterCustomer.value) params.customer_id = filterCustomer.value
    const { data } = await customersApi.listAllInvoices(params)
    allInvoices.value = data || []
  } catch {
    allInvoices.value = []
  } finally {
    loading.value = false
  }
}

function pdfUrl(inv) {
  const server = getServerUrl()
  const base = server ? `${server}/api/v1` : '/api/v1'
  const lang = auth.user?.locale || 'en'
  return `${base}/customers/${inv.customer_id}/invoices/${inv.id}/pdf?lang=${lang}`
}

watch(filterCustomer, load)
onMounted(load)
</script>

<style scoped>
.invoices-main { flex: 1; padding: 32px 24px; overflow-y: auto; }
.invoices-container { max-width: 1000px; margin: 0 auto; }

.page-header { margin-bottom: 20px; }
h1 { font-size: 22px; font-weight: 700; }

.summary-bar {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}
.summary-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.summary-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.summary-value {
  font-size: 18px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.summary-card--overdue.has-value .summary-value { color: var(--color-danger, #e53e3e); }
.summary-card--paid .summary-value { color: var(--color-success); }

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}
.filter-select { width: auto; min-width: 160px; }

.empty-state {
  color: var(--color-text-muted);
  font-size: 14px;
  padding: 32px 0;
  text-align: center;
}

.invoice-table-wrap { overflow-x: auto; }
.invoice-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
  font-size: 13px;
}
.invoice-table th,
.invoice-table td {
  padding: 10px 14px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  vertical-align: middle;
  white-space: nowrap;
}
.invoice-table th {
  font-weight: 600;
  color: var(--color-text-muted);
  font-size: 12px;
  background: var(--color-bg);
}
.invoice-table tbody tr:last-child td { border-bottom: none; }
.invoice-table tbody tr:hover { background: var(--color-surface-raised); }
.invoice-table tbody tr.row-overdue { background: color-mix(in srgb, var(--color-danger, #e53e3e) 6%, transparent); }
.invoice-table tbody tr.row-overdue:hover { background: color-mix(in srgb, var(--color-danger, #e53e3e) 10%, transparent); }

.col-right { text-align: right; }
.col-total { font-variant-numeric: tabular-nums; font-weight: 500; }
.col-period { color: var(--color-text-secondary); }
.col-date { color: var(--color-text-secondary); }
.col-number { font-weight: 500; }
.col-actions { text-align: right; }

.overdue-date { color: var(--color-danger, #e53e3e); font-weight: 600; }

.inv-link, .customer-link {
  color: var(--color-text);
  text-decoration: none;
}
.inv-link:hover, .customer-link:hover { color: var(--color-primary); text-decoration: underline; }

.inv-credit-mark {
  margin-left: 4px;
  font-size: 11px;
  color: var(--color-text-muted);
}

.text-muted { color: var(--color-text-muted); }

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
}
.status-draft    { background: color-mix(in srgb, var(--color-text-muted) 15%, transparent); color: var(--color-text-muted); }
.status-sent     { background: color-mix(in srgb, var(--color-primary) 15%, transparent);    color: var(--color-primary); }
.status-overdue  { background: color-mix(in srgb, var(--color-danger, #e53e3e) 15%, transparent); color: var(--color-danger, #e53e3e); }
.status-paid     { background: color-mix(in srgb, var(--color-success) 15%, transparent);    color: var(--color-success); }
.status-credit_note { background: color-mix(in srgb, var(--color-warning) 15%, transparent); color: var(--color-warning); }
</style>
