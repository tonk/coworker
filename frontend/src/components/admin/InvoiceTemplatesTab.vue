<template>
  <div>
    <div class="tab-toolbar">
      <button class="btn btn-primary btn-sm" @click="startCreate">+ {{ $t('admin.invoice_template_new') }}</button>
    </div>

    <div v-if="editing" class="tmpl-form-card">
      <h3 class="tmpl-form-title">{{ editing.id ? $t('admin.invoice_template_edit') : $t('admin.invoice_template_new') }}</h3>

      <div class="tmpl-form-row">
        <div class="tmpl-form-field" style="flex:2">
          <label class="tmpl-label" for="tmpl-name">{{ $t('admin.invoice_template_name') }}</label>
          <input id="tmpl-name" class="form-input" v-model="form.name" :placeholder="$t('admin.invoice_template_name')" />
        </div>
        <div class="tmpl-form-field" style="flex:0 0 120px">
          <label class="tmpl-label" for="tmpl-currency">{{ $t('admin.invoice_template_currency') }}</label>
          <input id="tmpl-currency" class="form-input" v-model="form.default_currency" placeholder="€" />
        </div>
        <div class="tmpl-form-field" style="flex:0 0 140px">
          <label class="tmpl-label" for="tmpl-vat">{{ $t('admin.invoice_template_vat') }}</label>
          <input id="tmpl-vat" class="form-input" type="number" min="0" max="100" step="0.5" v-model.number="form.default_vat_rate" />
        </div>
      </div>

      <div class="tmpl-form-field">
        <label class="tmpl-label" for="tmpl-notes">{{ $t('admin.invoice_template_notes') }}</label>
        <textarea id="tmpl-notes" class="form-input tmpl-notes" v-model="form.notes" rows="2" />
      </div>

      <div class="tmpl-line-items">
        <div class="tmpl-li-header">
          <span class="tmpl-label">{{ $t('invoice.line_description') }}</span>
          <span class="tmpl-label tmpl-col-qty">{{ $t('invoice.line_hours') }}</span>
          <span class="tmpl-label tmpl-col-price">{{ $t('invoice.line_rate') }}</span>
          <span class="tmpl-label tmpl-col-amount">{{ $t('invoice.line_amount') }}</span>
          <span class="tmpl-col-del"></span>
        </div>
        <div v-for="(li, idx) in formLines" :key="idx" class="tmpl-li-row">
          <input class="form-input tmpl-li-desc" v-model="li.description" :placeholder="$t('invoice.manual_line_desc')" />
          <input class="form-input tmpl-col-qty" type="number" min="0" step="0.25" v-model.number="li.quantity" />
          <input class="form-input tmpl-col-price" type="number" min="0" step="0.01" v-model.number="li.unit_price" />
          <span class="tmpl-col-amount tmpl-amount-val">{{ (li.quantity * li.unit_price).toFixed(2) }}</span>
          <button class="btn-icon-xs tmpl-col-del" @click="formLines.splice(idx, 1)" :aria-label="$t('common.delete')">✕</button>
        </div>
        <button class="btn btn-sm tmpl-add-line" @click="addLine">+ {{ $t('invoice.add_manual_line') }}</button>
      </div>

      <div class="tmpl-form-actions">
        <button class="btn" @click="cancelEdit">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="confirmSave">{{ $t('common.save') }}</button>
      </div>
    </div>

    <div v-if="loading" class="tmpl-empty">{{ $t('common.loading') }}</div>
    <div v-else-if="templates.length === 0 && !editing" class="tmpl-empty">{{ $t('admin.invoice_template_no') }}</div>
    <div v-else class="tmpl-list">
      <div v-for="t in templates" :key="t.id" class="tmpl-item">
        <div class="tmpl-item-info">
          <button class="tmpl-item-name" @click="startEdit(t)" :aria-label="$t('common.edit') + ' ' + t.name">{{ t.name }}</button>
          <span class="tmpl-item-meta">
            {{ t.default_currency || '€' }}
            <template v-if="t.default_vat_rate"> · {{ t.default_vat_rate }}% VAT</template>
            <template v-if="lineCount(t)"> · {{ lineCount(t) }} {{ lineCount(t) === 1 ? $t('invoice.line_description').toLowerCase() : $t('invoice.line_amount').toLowerCase() }}</template>
          </span>
          <span v-if="t.notes" class="tmpl-item-notes">{{ t.notes }}</span>
        </div>
        <div class="tmpl-item-actions">
          <button class="btn btn-sm btn-danger" @click="deleteTemplate(t)" :aria-label="$t('common.delete') + ' ' + t.name">{{ $t('common.delete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { customersApi } from '@/api/customers'
import { useUIStore } from '@/stores/ui'

const { t } = useI18n()
const ui = useUIStore()

const templates = ref([])
const loading = ref(false)
const editing = ref(null)
const form = ref(emptyForm())
const formLines = ref([])

function emptyForm() {
  return { name: '', default_vat_rate: 0, default_currency: '€', notes: '' }
}

function lineCount(tmpl) {
  try { return JSON.parse(tmpl.line_items || '[]').length } catch { return 0 }
}

async function load() {
  loading.value = true
  try {
    const { data } = await customersApi.listInvoiceTemplates()
    templates.value = data || []
  } catch {
    templates.value = []
  } finally {
    loading.value = false
  }
}

function startCreate() {
  editing.value = {}
  form.value = emptyForm()
  formLines.value = []
}

function startEdit(tmpl) {
  editing.value = tmpl
  form.value = {
    name: tmpl.name,
    default_vat_rate: tmpl.default_vat_rate,
    default_currency: tmpl.default_currency || '€',
    notes: tmpl.notes || '',
  }
  try { formLines.value = JSON.parse(tmpl.line_items || '[]') } catch { formLines.value = [] }
}

function cancelEdit() {
  editing.value = null
  formLines.value = []
}

function addLine() {
  formLines.value.push({ description: '', quantity: 1, unit_price: 0, amount: 0, currency: form.value.default_currency || '€', is_manual: true })
}

async function confirmSave() {
  if (!form.value.name.trim()) return
  const lines = formLines.value.map(li => ({ ...li, amount: (li.quantity || 0) * (li.unit_price || 0) }))
  const payload = {
    ...form.value,
    line_items: JSON.stringify(lines),
  }
  try {
    if (editing.value?.id) {
      const { data } = await customersApi.updateInvoiceTemplate(editing.value.id, payload)
      const idx = templates.value.findIndex(x => x.id === data.id)
      if (idx >= 0) templates.value[idx] = data
    } else {
      const { data } = await customersApi.createInvoiceTemplate(payload)
      templates.value.push(data)
      templates.value.sort((a, b) => a.name.localeCompare(b.name))
    }
    ui.success(t('admin.invoice_template_saved'))
    cancelEdit()
  } catch {
    ui.error(t('common.error'))
  }
}

async function deleteTemplate(tmpl) {
  if (!await ui.confirm(t('common.delete') + ' "' + tmpl.name + '"?', { destructive: true })) return
  try {
    await customersApi.deleteInvoiceTemplate(tmpl.id)
    templates.value = templates.value.filter(x => x.id !== tmpl.id)
    ui.success(t('admin.invoice_template_deleted'))
  } catch {
    ui.error(t('common.error'))
  }
}

onMounted(load)
</script>

<style scoped>
.tmpl-form-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 20px;
  margin-bottom: 20px;
}
.tmpl-form-title { font-size: 15px; font-weight: 600; margin-bottom: 16px; }
.tmpl-form-row { display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 12px; }
.tmpl-form-field { display: flex; flex-direction: column; gap: 4px; flex: 1; min-width: 140px; }
.tmpl-notes { resize: vertical; }
.tmpl-label { font-size: 12px; font-weight: 600; color: var(--color-text-muted); }

.tmpl-line-items { margin-top: 16px; }
.tmpl-li-header { display: grid; grid-template-columns: 1fr 80px 100px 80px 28px; gap: 8px; padding: 0 0 4px; }
.tmpl-li-row { display: grid; grid-template-columns: 1fr 80px 100px 80px 28px; gap: 8px; margin-bottom: 6px; align-items: center; }
.tmpl-li-desc { width: 100%; }
.tmpl-col-qty, .tmpl-col-price, .tmpl-col-amount, .tmpl-col-del { text-align: right; }
.tmpl-col-amount .tmpl-label, .tmpl-amount-val { font-variant-numeric: tabular-nums; }
.tmpl-add-line { margin-top: 8px; }

.tmpl-form-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px; }

.tmpl-empty { color: var(--color-text-muted); font-size: 14px; padding: 32px 0; text-align: center; }

.tmpl-list { display: flex; flex-direction: column; gap: 8px; margin-top: 20px; }
.tmpl-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 12px 16px;
}
.tmpl-item-info { display: flex; flex-direction: column; gap: 2px; }
.tmpl-item-name {
  font-weight: 600;
  font-size: 14px;
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  color: var(--color-primary);
  text-align: left;
}
.tmpl-item-name:hover,
.tmpl-item-name:focus-visible { text-decoration: underline; }
.tmpl-item-name:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.tmpl-item-meta { font-size: 12px; color: var(--color-text-muted); }
.tmpl-item-notes { font-size: 12px; color: var(--color-text-secondary); }
.tmpl-item-actions { display: flex; gap: 8px; flex-shrink: 0; }
</style>
