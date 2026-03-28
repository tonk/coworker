<template>
  <main class="settings-main">
      <div class="settings-container">
        <div class="settings-header">
          <RouterLink :to="`/projects/${slug}`" class="btn btn-ghost btn-sm">← Back</RouterLink>
          <h1>{{ $t('project.settings') }}: {{ project?.name }}</h1>
        </div>

        <div class="settings-tabs">
          <button :class="['tab', { active: tab === 'general' }]" @click="tab = 'general'">General</button>
          <button :class="['tab', { active: tab === 'members' }]" @click="tab = 'members'">{{ $t('project.members') }}</button>
          <button :class="['tab', { active: tab === 'labels' }]" @click="tab = 'labels'">{{ $t('project.labels') }}</button>
          <button :class="['tab', { active: tab === 'apikeys' }]" @click="tab = 'apikeys'; loadApiKeys()">{{ $t('apikeys.tab') }}</button>
          <button :class="['tab', { active: tab === 'webhooks' }]" @click="tab = 'webhooks'; loadWebhooks()">Webhooks</button>
        </div>

        <!-- General Tab -->
        <div v-if="tab === 'general'" class="tab-content">
          <div class="form-group">
            <label class="form-label">{{ $t('project.project_name') }}</label>
            <input class="form-input" v-model="form.name" style="max-width:400px" />
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('project.description') }}</label>
            <textarea class="form-input" v-model="form.description" rows="3" style="max-width:400px"></textarea>
          </div>
          <div class="form-group">
            <label class="form-label">{{ $t('project.color') }}</label>
            <input type="color" class="form-input" v-model="form.color" style="height:40px;padding:4px;width:80px" />
          </div>
          <button class="btn btn-primary" @click="saveGeneral">{{ $t('common.save') }}</button>

          <div class="danger-zone">
            <h3>Danger Zone</h3>
            <button class="btn btn-danger" @click="confirmDelete">{{ $t('project.delete') }}</button>
          </div>
        </div>

        <!-- Members Tab -->
        <div v-if="tab === 'members'" class="tab-content">
          <div class="section-action">
            <button class="btn btn-primary btn-sm" @click="showInvite = true; invite.userIds = []; inviteSearch = ''">+ {{ $t('project.invite_member') }}</button>
          </div>
          <table class="data-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>{{ $t('project.role') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in members" :key="m.id">
                <td>{{ m.user.display_name || m.user.username }}<br><small>{{ m.user.email }}</small></td>
                <td>
                  <select class="form-input" style="width:auto" v-model="m.role" @change="updateRole(m)">
                    <option value="owner">{{ $t('project.roles.owner') }}</option>
                    <option value="member">{{ $t('project.roles.member') }}</option>
                    <option value="viewer">{{ $t('project.roles.viewer') }}</option>
                  </select>
                </td>
                <td>
                  <button class="btn btn-danger btn-sm" @click="removeMember(m)">{{ $t('common.delete') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- API Keys Tab -->
        <div v-if="tab === 'apikeys'" class="tab-content">
          <div class="form-group" style="max-width:400px">
            <label class="form-label">{{ $t('apikeys.key_name') }}</label>
            <input class="form-input" v-model="newKeyName" :placeholder="$t('apikeys.key_name_placeholder')" />
          </div>
          <button class="btn btn-primary btn-sm" :disabled="!newKeyName.trim()" @click="generateKey">{{ $t('apikeys.generate') }}</button>

          <div v-if="generatedKey" class="new-key-box">
            <p class="new-key-notice">{{ $t('apikeys.copy_notice') }}</p>
            <code class="new-key-value">{{ generatedKey }}</code>
            <button class="btn btn-secondary btn-sm" @click="copyKey">{{ $t('apikeys.copy') }}</button>
          </div>

          <table class="data-table" style="margin-top:24px">
            <thead>
              <tr>
                <th>{{ $t('apikeys.name') }}</th>
                <th>{{ $t('apikeys.prefix') }}</th>
                <th>{{ $t('apikeys.last_used') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="apiKeys.length === 0">
                <td colspan="4" style="text-align:center;color:var(--color-text-muted)">{{ $t('apikeys.no_keys') }}</td>
              </tr>
              <tr v-for="key in apiKeys" :key="key.id">
                <td>{{ key.name }}</td>
                <td><code>{{ key.key_prefix }}…</code></td>
                <td>{{ key.last_used_at ? formatDateTime(key.last_used_at) : '—' }}</td>
                <td><button class="btn btn-danger btn-sm" @click="revokeKey(key)">{{ $t('apikeys.revoke') }}</button></td>
              </tr>
            </tbody>
          </table>

          <div class="api-docs">
            <h3>{{ $t('apikeys.docs_title') }}</h3>
            <p>{{ $t('apikeys.docs_auth') }}: <code>X-API-Key: cwk_…</code></p>
            <div class="api-endpoint">
              <span class="method post">POST</span>
              <code>/api/v1/ticket/{{ slug }}/cards</code>
              <span class="endpoint-desc">{{ $t('apikeys.docs_add') }}</span>
            </div>
            <pre class="api-body">{"title": "…", "description": "…", "column_id": 1}</pre>
            <div class="api-endpoint">
              <span class="method post">POST</span>
              <code>/api/v1/ticket/{{ slug }}/cards/:cardId/comments</code>
              <span class="endpoint-desc">{{ $t('apikeys.docs_comment') }}</span>
            </div>
            <pre class="api-body">{"body": "…"}</pre>
            <div class="api-endpoint">
              <span class="method patch">PATCH</span>
              <code>/api/v1/ticket/{{ slug }}/cards/:cardId/move</code>
              <span class="endpoint-desc">{{ $t('apikeys.docs_move') }}</span>
            </div>
            <pre class="api-body">{"column_id": 2, "position": 1000}</pre>
          </div>
        </div>

        <!-- Webhooks Tab -->
        <div v-if="tab === 'webhooks'" class="tab-content">
          <div class="form-group" style="max-width:420px">
            <label class="form-label">Webhook name</label>
            <input class="form-input" v-model="newWebhookName" placeholder="e.g. CI Bot" />
          </div>
          <div class="form-group" style="max-width:420px">
            <label class="form-label">Type</label>
            <select class="form-input" v-model="newWebhookType">
              <option value="generic">Generic (plain JSON)</option>
              <option value="gitea">Gitea / Forgejo</option>
            </select>
          </div>
          <button class="btn btn-primary btn-sm" :disabled="!newWebhookName.trim()" @click="createWebhook">Create Webhook</button>

          <div v-if="createdWebhookToken" class="new-key-box" style="margin-top:16px">
            <p class="new-key-notice">Copy this token now — it won't be shown again.</p>
            <code class="new-key-value">{{ createdWebhookToken }}</code>
            <button class="btn btn-secondary btn-sm" @click="copyWebhookToken">Copy</button>
          </div>

          <!-- Generic webhook docs -->
          <div class="webhook-docs" style="margin-top:20px">
            <h4 style="margin:0 0 6px">Generic webhook</h4>
            <p style="font-size:13px;color:var(--color-text-muted);margin:0 0 6px">
              <code>POST /api/v1/webhooks/&lt;token&gt;</code> — body: <code>{"text": "...", "username": "Bot"}</code>
            </p>
            <h4 style="margin:12px 0 6px">Gitea / Forgejo webhook</h4>
            <p style="font-size:13px;color:var(--color-text-muted);margin:0 0 4px">
              Create a webhook of type <strong>Gitea / Forgejo</strong> above, then in your Gitea/Forgejo repository go to
              <strong>Settings → Webhooks → Add Webhook → Gitea</strong> and set:
            </p>
            <ul style="font-size:13px;color:var(--color-text-muted);margin:0 0 4px;padding-left:18px">
              <li>Target URL: <code>{{ baseUrl }}/api/v1/gitea-webhook/&lt;token&gt;</code></li>
              <li>Content type: <code>application/json</code></li>
              <li>Secret: leave empty (the token in the URL authenticates the request)</li>
            </ul>
            <p style="font-size:13px;color:var(--color-text-muted);margin:0">
              Supported events: push, issues, pull request, comments, create, delete, release, fork.
            </p>
          </div>

          <table class="data-table" style="margin-top:24px">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Token (hint)</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!webhooks.length">
                <td colspan="5" style="text-align:center;color:var(--color-text-muted)">No webhooks yet</td>
              </tr>
              <tr v-for="wh in webhooks" :key="wh.id">
                <td>{{ wh.name }}</td>
                <td><span class="webhook-type-badge" :class="wh.type">{{ wh.type === 'gitea' ? 'Gitea/Forgejo' : 'Generic' }}</span></td>
                <td><code>…{{ wh.token_hint }}</code></td>
                <td>{{ formatDateTime(wh.created_at) }}</td>
                <td style="display:flex;gap:6px">
                  <button class="btn btn-secondary btn-sm" @click="regenerateWebhook(wh)">Regenerate</button>
                  <button class="btn btn-danger btn-sm" @click="deleteWebhook(wh)">Delete</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Labels Tab -->
        <div v-if="tab === 'labels'" class="tab-content">
          <div class="section-action">
            <button class="btn btn-primary btn-sm" @click="showAddLabel = true">+ Add Label</button>
          </div>
          <div class="labels-list">
            <div v-for="label in labels" :key="label.id" class="label-row">
              <span class="label-preview" :style="{ background: label.color }">{{ label.name }}</span>
              <button class="btn btn-danger btn-sm" @click="deleteLabel(label)">{{ $t('common.delete') }}</button>
            </div>
          </div>
        </div>
      </div>
  </main>

  <BaseModal v-if="showInvite" :title="$t('project.invite_member')" @close="showInvite = false">
      <div class="form-group">
        <label class="form-label">{{ $t('project.select_user') }}</label>
        <div class="invite-search-wrap">
          <input class="form-input invite-search" v-model="inviteSearch" placeholder="Filter users…" />
        </div>
        <div class="invite-user-list">
          <label
            v-for="u in filteredInvitableUsers"
            :key="u.id"
            class="invite-user-row"
          >
            <input type="checkbox" :value="u.id" v-model="invite.userIds" class="invite-checkbox" />
            <span class="invite-avatar">{{ (u.display_name || u.username).slice(0, 2).toUpperCase() }}</span>
            <span class="invite-name">{{ u.display_name || u.username }}</span>
            <span class="invite-email">{{ u.email }}</span>
          </label>
          <div v-if="!filteredInvitableUsers.length" class="invite-empty">No users available</div>
        </div>
        <div v-if="invite.userIds.length" class="invite-selected-count">
          {{ invite.userIds.length }} user{{ invite.userIds.length > 1 ? 's' : '' }} selected
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('project.role') }}</label>
        <select class="form-input" v-model="invite.role">
          <option value="member">{{ $t('project.roles.member') }}</option>
          <option value="viewer">{{ $t('project.roles.viewer') }}</option>
          <option value="owner">{{ $t('project.roles.owner') }}</option>
        </select>
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="showInvite = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!invite.userIds.length" @click="sendInvite">{{ $t('project.invite_member') }}</button>
      </template>
  </BaseModal>

  <BaseModal v-if="showAddLabel" title="Add Label" @close="showAddLabel = false">
      <div class="form-group">
        <label class="form-label">Name</label>
        <input class="form-input" v-model="newLabel.name" autofocus />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('project.color') }}</label>
        <input type="color" class="form-input" v-model="newLabel.color" style="height:40px;padding:4px;width:80px" />
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="showAddLabel = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="createLabel">{{ $t('common.create') }}</button>
      </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import BaseModal from '@/components/common/BaseModal.vue'
import { useProjectStore } from '@/stores/project'
import { useUIStore } from '@/stores/ui'
import { projectsApi } from '@/api/projects'
import { authApi } from '@/api/auth'
import { useDateFormat } from '@/composables/useDateFormat'
import client from '@/api/client'

const route = useRoute()
const { formatDateTime } = useDateFormat()
const router = useRouter()
const slug = computed(() => route.params.slug)
const projectStore = useProjectStore()
const ui = useUIStore()

const tab = ref('general')
const project = ref(null)
const members = ref([])
const labels = ref([])
const showInvite = ref(false)
const showAddLabel = ref(false)
const invite = ref({ userIds: [], role: 'member' })
const inviteSearch = ref('')
const allUsers = ref([])
const newLabel = ref({ name: '', color: '#6366f1' })
const form = ref({ name: '', description: '', color: '' })
const apiKeys = ref([])
const newKeyName = ref('')
const generatedKey = ref('')

// Webhooks state
const webhooks = ref([])
const newWebhookName = ref('')
const newWebhookType = ref('generic')
const createdWebhookToken = ref('')
const baseUrl = computed(() => window.location.origin)

// Users not yet in the project
const invitableUsers = computed(() => {
  const memberIds = new Set(members.value.map(m => m.user_id || m.user?.id))
  return allUsers.value.filter(u => !memberIds.has(u.id))
})

const filteredInvitableUsers = computed(() => {
  const q = inviteSearch.value.toLowerCase()
  if (!q) return invitableUsers.value
  return invitableUsers.value.filter(u =>
    (u.display_name || '').toLowerCase().includes(q) ||
    u.username.toLowerCase().includes(q) ||
    u.email.toLowerCase().includes(q)
  )
})

onMounted(async () => {
  const data = await projectStore.fetchProject(slug.value)
  project.value = data
  form.value = { name: data.name, description: data.description || '', color: data.color || '#6366f1' }
  loadMembers()
  loadLabels()
  // Load all active users for the invite dropdown
  try {
    const { data: users } = await client.get('/users')
    allUsers.value = users || []
  } catch {}
})

async function loadMembers() {
  const { data } = await projectsApi.listMembers(slug.value)
  members.value = data
}

async function loadLabels() {
  const { data } = await projectsApi.listLabels(slug.value)
  labels.value = data
}

async function saveGeneral() {
  try {
    await projectStore.updateProject(slug.value, form.value)
    ui.success('Saved')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed')
  }
}

async function confirmDelete() {
  if (!confirm('Delete this project? This cannot be undone.')) return
  try {
    await projectStore.deleteProject(slug.value)
    router.push('/')
  } catch (e) {
    ui.error('Failed to delete project')
  }
}

async function updateRole(member) {
  try {
    await projectsApi.updateMemberRole(slug.value, member.user.id, member.role)
  } catch (e) {
    ui.error('Failed to update role')
    loadMembers()
  }
}

async function removeMember(member) {
  if (!confirm('Remove this member?')) return
  await projectsApi.removeMember(slug.value, member.user.id)
  loadMembers()
}

async function sendInvite() {
  if (!invite.value.userIds.length) return
  const users = allUsers.value.filter(u => invite.value.userIds.includes(u.id))
  let failed = 0
  for (const user of users) {
    try {
      await projectsApi.inviteMember(slug.value, { login: user.username, role: invite.value.role })
    } catch {
      failed++
    }
  }
  showInvite.value = false
  invite.value = { userIds: [], role: 'member' }
  inviteSearch.value = ''
  loadMembers()
  if (failed === 0) {
    ui.success(users.length > 1 ? `${users.length} members invited` : 'Member invited')
  } else {
    ui.error(`${failed} invitation(s) failed`)
  }
}

async function createLabel() {
  try {
    await projectsApi.createLabel(slug.value, newLabel.value)
    showAddLabel.value = false
    newLabel.value = { name: '', color: '#6366f1' }
    loadLabels()
  } catch (e) {
    ui.error('Failed to create label')
  }
}

async function deleteLabel(label) {
  if (!confirm('Delete this label?')) return
  await projectsApi.deleteLabel(slug.value, label.id)
  loadLabels()
}

async function loadApiKeys() {
  const { data } = await authApi.listApiKeys()
  apiKeys.value = data
}

async function generateKey() {
  try {
    const { data } = await authApi.createApiKey(newKeyName.value.trim())
    generatedKey.value = data.key
    newKeyName.value = ''
    loadApiKeys()
  } catch (e) {
    ui.error('Failed to generate key')
  }
}

async function revokeKey(key) {
  if (!confirm('Revoke this API key?')) return
  await authApi.deleteApiKey(key.id)
  loadApiKeys()
}

function copyKey() {
  navigator.clipboard.writeText(generatedKey.value)
  ui.success('Copied!')
}

async function loadWebhooks() {
  try {
    const { data } = await projectsApi.listWebhooks(slug.value)
    webhooks.value = data
  } catch {}
}

async function createWebhook() {
  try {
    const { data } = await projectsApi.createWebhook(slug.value, {
      name: newWebhookName.value.trim(),
      type: newWebhookType.value,
    })
    createdWebhookToken.value = data.token
    newWebhookName.value = ''
    newWebhookType.value = 'generic'
    await loadWebhooks()
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create webhook')
  }
}

async function deleteWebhook(wh) {
  if (!confirm('Delete this webhook?')) return
  await projectsApi.deleteWebhook(slug.value, wh.id)
  await loadWebhooks()
}

async function regenerateWebhook(wh) {
  if (!confirm('Regenerate token? The old token will stop working immediately.')) return
  try {
    const { data } = await projectsApi.regenerateWebhook(slug.value, wh.id)
    createdWebhookToken.value = data.token
    await loadWebhooks()
  } catch (e) {
    ui.error('Failed to regenerate token')
  }
}

function copyWebhookToken() {
  navigator.clipboard.writeText(createdWebhookToken.value)
  ui.success('Copied!')
}
</script>

<style scoped>
.settings-main { flex: 1; padding: 32px 24px; }
.settings-container { max-width: 800px; margin: 0 auto; }
.settings-header { display: flex; align-items: center; gap: 16px; margin-bottom: 28px; }
.settings-header h1 { font-size: 20px; font-weight: 700; }

.settings-tabs { display: flex; gap: 0; border-bottom: 2px solid var(--color-border); margin-bottom: 28px; }
.tab { padding: 10px 20px; background: transparent; border: none; font-size: 14px; font-weight: 500; color: var(--color-text-muted); cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px; }
.tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.tab:hover { color: var(--color-text); }

.tab-content { padding-top: 8px; }

.section-action { margin-bottom: 16px; }

.data-table { width: 100%; border-collapse: collapse; }
.data-table th, .data-table td { padding: 10px 12px; text-align: left; border-bottom: 1px solid var(--color-border); font-size: 13px; }
.data-table th { font-weight: 600; color: var(--color-text-muted); font-size: 12px; }
.data-table small { color: var(--color-text-muted); font-size: 11px; }

.labels-list { display: flex; flex-direction: column; gap: 8px; }
.label-row { display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; background: #f8fafc; border-radius: var(--radius-sm); }
.label-preview { padding: 3px 10px; border-radius: 9999px; color: #fff; font-size: 12px; font-weight: 600; }

.danger-zone { margin-top: 40px; padding: 20px; border: 1px solid #fecaca; border-radius: var(--radius); background: #fff5f5; }
.danger-zone h3 { color: var(--color-danger); margin-bottom: 12px; font-size: 14px; }

.new-key-box { margin-top: 16px; padding: 16px; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); display: flex; flex-direction: column; gap: 8px; max-width: 560px; }
.new-key-notice { font-size: 13px; color: var(--color-text-muted); margin: 0; }
.new-key-value { font-size: 13px; word-break: break-all; background: var(--color-bg); padding: 8px; border-radius: var(--radius-sm); border: 1px solid var(--color-border); }

.api-docs { margin-top: 36px; padding: 20px; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); }
.api-docs h3 { font-size: 15px; font-weight: 700; margin-bottom: 16px; }
.api-docs p { font-size: 13px; margin-bottom: 12px; }
.api-endpoint { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
.api-endpoint code { font-size: 13px; }
.endpoint-desc { font-size: 12px; color: var(--color-text-muted); }
.api-body { font-size: 12px; background: var(--color-bg); padding: 8px 12px; border-radius: var(--radius-sm); border: 1px solid var(--color-border); margin-bottom: 16px; }
.method { font-size: 11px; font-weight: 700; padding: 2px 6px; border-radius: 4px; color: #fff; }
.method.post { background: #10b981; }

/* ── Invite multi-select ─────────────────────────────────── */
.invite-search-wrap { margin-bottom: 6px; }
.invite-search { width: 100%; }

.invite-user-list {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  max-height: 220px;
  overflow-y: auto;
}

.invite-user-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background .1s;
  user-select: none;
}
.invite-user-row:not(:last-child) { border-bottom: 1px solid var(--color-border); }
.invite-user-row:hover { background: var(--color-bg); }

.invite-checkbox { flex-shrink: 0; width: 15px; height: 15px; accent-color: var(--color-primary); cursor: pointer; }

.invite-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--color-primary);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.invite-name { font-size: 13px; font-weight: 500; flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.invite-email { font-size: 11px; color: var(--color-text-muted); white-space: nowrap; }
.invite-empty { padding: 16px; text-align: center; color: var(--color-text-muted); font-size: 13px; }

.invite-selected-count {
  margin-top: 6px;
  font-size: 12px;
  color: var(--color-primary);
  font-weight: 500;
}
.method.patch { background: #f59e0b; }

.webhook-type-badge { font-size: 11px; font-weight: 600; padding: 2px 7px; border-radius: 9999px; }
.webhook-type-badge.gitea { background: #fef3c7; color: #92400e; }
.webhook-type-badge.generic { background: var(--color-surface); color: var(--color-text-muted); border: 1px solid var(--color-border); }

.webhook-docs { padding: 16px; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); max-width: 640px; }
.webhook-docs h4 { font-size: 13px; font-weight: 700; }
</style>
