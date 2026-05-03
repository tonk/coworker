<template>
  <div class="customers-page">
    <div class="page-header">
      <h1>{{ $t('customer.customers') }}</h1>
      <button v-if="canManage" class="btn btn-primary" @click="showCreate = true">
        + {{ $t('customer.new_customer') }}
      </button>
    </div>

    <div v-if="loading" class="loading">{{ $t('common.loading') }}</div>
    <div v-else-if="!customers.length" class="empty-state">{{ $t('customer.no_customers') }}</div>

    <div v-else class="customers-grid">
      <RouterLink
        v-for="c in customers"
        :key="c.id"
        :to="`/customers/${c.id}`"
        class="customer-tile"
      >
        <div class="tile-header">
          <img v-if="c.logo_url" :src="resolveAssetUrl(c.logo_url)" class="cust-logo" alt="" />
          <span v-else class="cust-logo-placeholder">{{ c.name[0] }}</span>
          <button
            class="star-btn"
            :class="{ active: c.is_favorite }"
            @click.prevent="toggleFav(c)"
            :title="c.is_favorite ? $t('customer.unstar') : $t('customer.star')"
          >{{ c.is_favorite ? '★' : '☆' }}</button>
        </div>
        <div class="tile-name">{{ c.name }}</div>
        <div v-if="c.description" class="tile-desc">{{ c.description }}</div>
        <div class="tile-meta">
          <span>{{ c.contract_count }} {{ $t('contract.contracts').toLowerCase() }}</span>
          <span>{{ c.project_count }} {{ $t('project.projects').toLowerCase() }}</span>
        </div>
      </RouterLink>
    </div>

    <!-- Create dialog -->
    <BaseModal v-if="showCreate" :title="$t('customer.new_customer')" @close="showCreate = false">
      <div class="form-group">
        <label class="form-label">{{ $t('customer.name') }}</label>
        <input class="form-input" v-model="form.name" :placeholder="$t('customer.name')" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('customer.description') }}</label>
        <textarea class="form-input" v-model="form.description" rows="3"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('customer.logo_url') }}</label>
        <input class="form-input" v-model="form.logo_url" placeholder="https://..." />
      </div>
      <template #footer>
        <button class="btn" @click="showCreate = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="doCreate" :disabled="!form.name.trim()">
          {{ $t('common.create') }}
        </button>
      </template>
    </BaseModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCustomersStore } from '@/stores/customers'
import { useUIStore } from '@/stores/ui'
import { customersApi } from '@/api/customers'
import { resolveAssetUrl } from '@/api/serverConfig'
import BaseModal from '@/components/common/BaseModal.vue'

const auth = useAuthStore()
const custStore = useCustomersStore()
const ui = useUIStore()

const loading = ref(true)
const showCreate = ref(false)
const form = ref({ name: '', description: '', logo_url: '' })

const customers = computed(() => custStore.customers)

const canManage = computed(() => auth.isAdmin)

onMounted(async () => {
  await custStore.fetchCustomers()
  loading.value = false
})

async function toggleFav(c) {
  await custStore.toggleFavorite(c.id)
}

async function doCreate() {
  if (!form.value.name.trim()) return
  try {
    await customersApi.create({
      name: form.value.name.trim(),
      description: form.value.description,
      logo_url: form.value.logo_url,
    })
    await custStore.fetchCustomers()
    showCreate.value = false
    form.value = { name: '', description: '', logo_url: '' }
    ui.success('Customer created')
  } catch (e) {
    ui.error(e?.response?.data?.error || 'Failed to create customer')
  }
}
</script>

<style scoped>
.customers-page {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
}

.loading, .empty-state {
  color: var(--color-text-muted);
  text-align: center;
  padding: 48px;
}

.customers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.customer-tile {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 16px;
  text-decoration: none;
  color: var(--color-text);
  transition: box-shadow .15s, border-color .15s;
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative;
}
.customer-tile:hover {
  box-shadow: 0 2px 12px rgba(0,0,0,.1);
  border-color: var(--color-primary);
  text-decoration: none;
}

.tile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.cust-logo {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: contain;
}

.cust-logo-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
}

.star-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 2px 4px;
  transition: color .1s;
}
.star-btn.active { color: #f59e0b; }
.star-btn:hover { color: #f59e0b; }

.tile-name {
  font-weight: 600;
  font-size: 15px;
}

.tile-desc {
  font-size: 12px;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tile-meta {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text);
  cursor: pointer;
  font-size: 14px;
}
.btn-primary {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.btn:disabled { opacity: .5; cursor: not-allowed; }
</style>
