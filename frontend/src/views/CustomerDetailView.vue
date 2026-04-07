<template>
  <div class="customer-detail-page">
    <div v-if="loading" class="loading">{{ $t('common.loading') }}</div>
    <template v-else-if="detail">

      <!-- Customer header -->
      <div class="cust-header">
        <div class="cust-logo-wrap">
          <img v-if="detail.customer.logo_url" :src="detail.customer.logo_url" class="cust-logo" alt="" />
          <span v-else class="cust-logo-placeholder">{{ detail.customer.name[0] }}</span>
        </div>
        <div class="cust-info">
          <div v-if="!editingName" class="cust-name-row">
            <h1 class="cust-name">{{ detail.customer.name }}</h1>
            <button v-if="canManage" class="icon-btn" @click="startEditName" title="Edit">✎</button>
          </div>
          <div v-else class="cust-name-row">
            <input class="form-input name-input" v-model="nameEdit" @keydown.enter="saveNameEdit" @keydown.escape="editingName = false" />
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
          >{{ detail.customer.is_favorite ? '★' : '☆' }}</button>
          <button v-if="canManage" class="btn btn-sm" @click="showEdit = true">{{ $t('customer.edit') }}</button>
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
            </div>
            <div class="contract-actions">
              <button v-if="canManage" class="icon-btn" @click="editContract(grp)" title="Edit">✎</button>
              <button v-if="auth.isAdmin" class="icon-btn icon-danger" @click="deleteContract(grp)" title="Delete">✕</button>
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
              <span class="proj-dot" :style="{ background: p.color || '#6366f1' }"></span>
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
              <span class="proj-dot" :style="{ background: p.color || '#6366f1' }"></span>
              <span>{{ p.name }}</span>
            </RouterLink>
          </div>
        </div>
      </section>

    </template>

    <!-- Edit customer modal -->
    <BaseModal v-if="showEdit" :title="$t('customer.edit')" @close="showEdit = false">
      <div class="form-group">
        <label class="form-label">{{ $t('customer.name') }}</label>
        <input class="form-input" v-model="editForm.name" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('customer.description') }}</label>
        <textarea class="form-input" v-model="editForm.description" rows="3"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('customer.logo_url') }}</label>
        <input class="form-input" v-model="editForm.logo_url" placeholder="https://..." />
      </div>
      <template #footer>
        <button class="btn" @click="showEdit = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEdit">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>

    <!-- Add / edit contract modal -->
    <BaseModal v-if="showAddContract || editingContract" :title="editingContract ? $t('contract.edit') : $t('contract.new_contract')" @close="closeContractModal">
      <div class="form-group">
        <label class="form-label">{{ $t('contract.name') }}</label>
        <input class="form-input" v-model="contractForm.name" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('contract.description') }}</label>
        <textarea class="form-input" v-model="contractForm.description" rows="2"></textarea>
      </div>
      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label">{{ $t('contract.start_date') }}</label>
          <input class="form-input" type="date" v-model="contractForm.start_date" />
        </div>
        <div class="form-group half">
          <label class="form-label">{{ $t('contract.end_date') }}</label>
          <input class="form-input" type="date" v-model="contractForm.end_date" />
        </div>
      </div>
      <template #footer>
        <button class="btn" @click="closeContractModal">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveContract" :disabled="!contractForm.name.trim()">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCustomersStore } from '@/stores/customers'
import { useUIStore } from '@/stores/ui'
import { customersApi } from '@/api/customers'
import BaseModal from '@/components/common/BaseModal.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const custStore = useCustomersStore()
const ui = useUIStore()

const loading = ref(true)
const detail = ref(null)

const showEdit = ref(false)
const editForm = ref({ name: '', description: '', logo_url: '' })

const showAddContract = ref(false)
const editingContract = ref(null)
const contractForm = ref({ name: '', description: '', start_date: '', end_date: '' })

const editingName = ref(false)
const nameEdit = ref('')

const canManage = computed(() => auth.isAdmin || auth.user?.global_role === 'user')

const custId = computed(() => Number(route.params.id))

onMounted(() => load())

async function load() {
  loading.value = true
  try {
    const { data } = await customersApi.get(custId.value)
    detail.value = data
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
  }
}

function closeContractModal() {
  showAddContract.value = false
  editingContract.value = null
  contractForm.value = { name: '', description: '', start_date: '', end_date: '' }
}

async function saveContract() {
  const payload = {
    name:        contractForm.value.name,
    description: contractForm.value.description,
    start_date:  contractForm.value.start_date || '',
    end_date:    contractForm.value.end_date   || '',
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

function formatDate(dt) {
  if (!dt) return ''
  return dt.split('T')[0]
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
.btn-sm { padding: 4px 10px; font-size: 12px; }
.btn:disabled { opacity: .5; cursor: not-allowed; }

.form-group { margin-bottom: 12px; }
.form-label { display: block; font-size: 12px; font-weight: 600; margin-bottom: 4px; color: var(--color-text-muted); }
.form-input { width: 100%; padding: 8px 10px; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-bg); color: var(--color-text); font-size: 14px; box-sizing: border-box; }
.detail-row { display: flex; gap: 12px; }
.half { flex: 1; }
</style>
