<template>
  <main class="admin-main">
      <div class="admin-container">
        <h1>{{ $t('admin.panel') }}</h1>

        <div class="tabs" role="tablist" :aria-label="$t('admin.panel')">
          <button :class="['tab', { active: tab === 'users' }]" @click="tab = 'users'" role="tab" :aria-selected="tab === 'users'" aria-controls="tab-panel-users" id="tab-btn-users">{{ $t('admin.users') }}</button>
          <button :class="['tab', { active: tab === 'groups' }]" @click="tab = 'groups'; loadGroups()" role="tab" :aria-selected="tab === 'groups'" aria-controls="tab-panel-groups" id="tab-btn-groups">{{ $t('groups.title') }}</button>
          <button :class="['tab', { active: tab === 'customers' }]" @click="tab = 'customers'; loadAdminCustomers()" role="tab" :aria-selected="tab === 'customers'" aria-controls="tab-panel-customers" id="tab-btn-customers">{{ $t('customer.customers') }}</button>
          <button v-if="!systemStore.isTimetrackingMode" :class="['tab', { active: tab === 'projects' }]" @click="tab = 'projects'; loadProjects()" role="tab" :aria-selected="tab === 'projects'" aria-controls="tab-panel-projects" id="tab-btn-projects">{{ $t('admin.projects') }}</button>
          <button :class="['tab', { active: tab === 'settings' }]" @click="tab = 'settings'; loadSettings()" role="tab" :aria-selected="tab === 'settings'" aria-controls="tab-panel-settings" id="tab-btn-settings">{{ $t('admin.settings') }}</button>
          <button :class="['tab', { active: tab === 'news' }]" @click="tab = 'news'; loadNews()" role="tab" :aria-selected="tab === 'news'" aria-controls="tab-panel-news" id="tab-btn-news">{{ $t('admin.news_tab') }}</button>
          <button :class="['tab', { active: tab === 'time-tracking' }]" @click="tab = 'time-tracking'; loadAdminTTProjects(); loadAdminTTCustomers()" role="tab" :aria-selected="tab === 'time-tracking'" aria-controls="tab-panel-time-tracking" id="tab-btn-time-tracking">{{ $t('admin.time_tracking') }}</button>
          <button v-if="!systemStore.isTimetrackingMode" :class="['tab', { active: tab === 'sla' }]" @click="tab = 'sla'" role="tab" :aria-selected="tab === 'sla'" aria-controls="tab-panel-sla" id="tab-btn-sla">{{ $t('sla.title') }}</button>
          <button v-if="!systemStore.isTimetrackingMode" :class="['tab', { active: tab === 'macros' }]" @click="tab = 'macros'" role="tab" :aria-selected="tab === 'macros'" aria-controls="tab-panel-macros" id="tab-btn-macros">{{ $t('macro.title') }}</button>
          <button v-if="!systemStore.isTimetrackingMode" :class="['tab', { active: tab === 'ticket-checklists' }]" @click="tab = 'ticket-checklists'" role="tab" :aria-selected="tab === 'ticket-checklists'" aria-controls="tab-panel-ticket-checklists" id="tab-btn-ticket-checklists">{{ $t('ticketChecklist.title') }}</button>
          <button :class="['tab', { active: tab === 'backup' }]" @click="tab = 'backup'; loadBackups(); loadSettings()" role="tab" :aria-selected="tab === 'backup'" aria-controls="tab-panel-backup" id="tab-btn-backup">{{ $t('admin.backup_tab') }}</button>
        </div>

        <!-- Users tab -->
        <div v-show="tab === 'users'" role="tabpanel" id="tab-panel-users" aria-labelledby="tab-btn-users">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="openCreateUser">+ {{ $t('admin.new_user') }}</button>
            <label class="toggle-label" style="display:flex;align-items:center;gap:6px;font-size:13px;cursor:pointer">
              <input type="checkbox" v-model="showInactiveUsers" />
              {{ $t('admin.show_inactive') }}
            </label>
            <input v-model="userSearch" class="form-input admin-search" type="search" :placeholder="$t('common.search')" aria-label="Search users" />
          </div>

          <div v-if="loading" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>

          <table v-else class="data-table">
            <thead>
              <tr>
                <th style="width: 48px;"></th>
                <th class="sortable-th" @click="toggleUserSort()">
                  {{ $t('admin.user') }}
                  <span class="sort-indicator">{{ userSortDir === 'asc' ? '△' : '▽' }}</span>
                </th>
                <th>{{ $t('admin.global_role') }}</th>
                <th>{{ $t('admin.last_login') }}</th>
                <th>{{ $t('admin.last_password_change') }}</th>
                <th>{{ $t('admin.mfa_enabled') }}</th>
                <th>{{ $t('common.status') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in sortedUsers" :key="user.id" :style="user.deleted_at ? 'opacity:0.5' : ''">
                <td>
                  <div class="member-avatar-wrap">
                    <img v-if="getUserAvatar(user)" :src="getUserAvatar(user)" class="member-avatar" alt="" />
                    <span v-else class="member-avatar-initials" :style="getAvatarColor(user)">{{ getInitials(user) }}</span>
                  </div>
                </td>
                <td>
                  <button type="button" class="name-link" @click="openEditUser(user)">{{ user.display_name || user.username }}</button>
                  <br>
                  <small>{{ user.first_name }} {{ user.last_name }}</small>
                  <br><small class="email">{{ user.email }}</small>
                </td>
                <td>
                  <select class="role-select" :value="user.global_role" @change="setRole(user, $event.target.value)">
                    <option value="admin">{{ $t('admin.role_admin') }}</option>
                    <option value="user">{{ $t('admin.role_user') }}</option>
                    <option value="viewer">{{ $t('admin.role_viewer') }}</option>
                    <option value="metrics">{{ $t('admin.role_metrics') }}</option>
                    <option value="backup">{{ $t('admin.role_backup') }}</option>
                    <option value="customer">{{ $t('admin.role_customer') }}</option>
                  </select>
                </td>
                <td>
                  <small>{{ user.last_login_at ? formatDateTime(user.last_login_at) : '-' }}</small>
                </td>
                <td>
                  <small>{{ user.password_changed_at ? formatDateTime(user.password_changed_at) : '-' }}</small>
                </td>
                <td>
                  <span v-if="user.totp_enabled" class="badge badge-mfa" :title="$t('mfa.title')">MFA</span>
                  <span v-else class="text-muted">-</span>
                </td>
                <td>
                  <span :class="['badge', user.is_active ? 'badge-active' : 'badge-inactive']">
                    {{ user.is_active ? $t('admin.active') : $t('admin.inactive') }}
                  </span>
                </td>
                <td>
                  <div class="actions-cell">
                    <template v-if="user.deleted_at">
                      <button class="btn btn-ghost btn-sm" @click="restoreUser(user)">{{ $t('admin.restore') }}</button>
                      <button class="btn btn-ghost btn-sm btn-danger" @click="purgeUser(user)">{{ $t('admin.purge') }}</button>
                    </template>
                    <template v-else>
                      <button class="btn btn-ghost btn-sm" @click="openEditUser(user)">{{ $t('common.edit') }}</button>
                      <template v-if="user.id !== auth.user?.id">
                        <button class="btn btn-ghost btn-sm" @click="toggleActive(user)">
                          {{ user.is_active ? $t('admin.deactivate') : $t('admin.activate') }}
                        </button>
                        <button class="btn btn-ghost btn-sm btn-danger" @click="deleteUser(user)">{{ $t('common.delete') }}</button>
                      </template>
                    </template>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>

          <div v-if="sortedUsers.length" class="feature-matrix">
            <h3>{{ $t('feature.features') }}</h3>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width:42px;"></th>
                  <th>{{ $t('admin.user') }}</th>
                  <th style="text-align:center;width:70px;">{{ $t('feature.board') }}</th>
                  <th style="text-align:center;width:70px;">{{ $t('feature.chat') }}</th>
                  <th style="text-align:center;width:90px;">{{ $t('feature.time_tracking') }}</th>
                  <th style="text-align:center;width:100px;">{{ $t('admin.timetracking_viewer') }}</th>
                  <th style="text-align:center;width:90px;">{{ $t('feature.helpdesk') }}</th>
                  <th style="text-align:center;width:60px;">{{ $t('admin.mfa_enabled') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="user in sortedUsers" :key="'feat-'+user.id">
                  <td>
                    <div class="member-avatar-wrap">
                      <img v-if="getUserAvatar(user)" :src="getUserAvatar(user)" class="member-avatar" alt="" />
                      <span v-else class="member-avatar-initials" :style="getAvatarColor(user)">{{ getInitials(user) }}</span>
                    </div>
                  </td>
                  <td>{{ user.display_name || user.username }}</td>
                  <td style="text-align:center;">
                    <template v-if="user.global_role === 'admin'"><span class="feat-check feat-always">✓</span></template>
                    <template v-else-if="user.global_role === 'customer'"><span class="feat-check feat-off">—</span></template>
                    <span v-else-if="user.global_role === 'metrics' || user.global_role === 'backup'" class="feat-check feat-off">—</span>
                    <input v-else type="checkbox" class="feat-toggle" :checked="user.board_enabled !== false" @change="toggleFeature(user, 'board_enabled', $event.target.checked)" />
                  </td>
                  <td style="text-align:center;">
                    <template v-if="user.global_role === 'admin'"><span class="feat-check feat-always">✓</span></template>
                    <template v-else-if="user.global_role === 'customer'"><span class="feat-check feat-off">—</span></template>
                    <span v-else-if="user.global_role === 'metrics' || user.global_role === 'backup'" class="feat-check feat-off">—</span>
                    <input v-else type="checkbox" class="feat-toggle" :checked="user.chat_enabled !== false" @change="toggleFeature(user, 'chat_enabled', $event.target.checked)" />
                  </td>
                  <td style="text-align:center;">
                    <template v-if="user.global_role === 'customer'"><span class="feat-check feat-off">—</span></template>
                    <span v-else-if="user.global_role === 'metrics' || user.global_role === 'backup'" class="feat-check feat-off">—</span>
                    <input v-else type="checkbox" class="feat-toggle" :checked="!!user.time_tracking_enabled" @change="toggleFeature(user, 'time_tracking_enabled', $event.target.checked)" />
                  </td>
                  <td style="text-align:center;">
                    <template v-if="user.global_role === 'customer'"><span class="feat-check feat-off">—</span></template>
                    <span v-else-if="user.global_role === 'metrics' || user.global_role === 'backup'" class="feat-check feat-off">—</span>
                    <input v-else type="checkbox" class="feat-toggle" :checked="!!user.time_tracking_viewer" @change="toggleFeature(user, 'time_tracking_viewer', $event.target.checked)" />
                  </td>
                  <td style="text-align:center;">
                    <template v-if="user.global_role === 'admin'"><span class="feat-check feat-always">✓</span></template>
                    <template v-else-if="user.global_role === 'customer'"><span class="feat-check feat-always">✓</span></template>
                    <span v-else-if="user.global_role === 'metrics' || user.global_role === 'backup'" class="feat-check feat-off">—</span>
                    <input v-else type="checkbox" class="feat-toggle" :checked="!!user.helpdesk_enabled" @change="toggleFeature(user, 'helpdesk_enabled', $event.target.checked)" />
                  </td>
                  <td style="text-align:center;">
                    <span v-if="user.totp_enabled" class="feat-check feat-on">✓</span>
                    <span v-else class="feat-check feat-off">—</span>
                  </td>
                </tr>
              </tbody>
            </table>
            <p v-if="sortedUsers.some(u => u.global_role === 'admin')" style="font-size:12px;color:var(--color-text-muted);margin-top:8px;">{{ $t('admin.admins_bypass_features') }}</p>
            <p v-if="sortedUsers.some(u => u.global_role === 'customer')" style="font-size:12px;color:var(--color-text-muted);margin-top:4px;">{{ $t('admin.customer_role_hint') }}</p>
          </div>
        </div>

        <!-- Projects tab -->
        <div v-show="tab === 'projects'" role="tabpanel" id="tab-panel-projects" aria-labelledby="tab-btn-projects">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="showCreateProject = true">+ {{ $t('project.new_project') }}</button>
            <label class="toggle-label" style="display:flex;align-items:center;gap:6px;font-size:13px;cursor:pointer">
              <input type="checkbox" v-model="showDeletedProjects" @change="loadProjects()" />
              {{ $t('admin.show_deleted') }}
            </label>
            <input v-model="projectSearch" class="form-input admin-search" type="search" :placeholder="$t('common.search')" aria-label="Search projects" />
          </div>
          <div v-if="loadingProjects" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>

          <table v-else class="data-table">
            <thead>
              <tr>
                <th class="sortable-th" @click="toggleProjectSort()">
                  {{ $t('project.project_name') }}
                  <span class="sort-indicator">{{ projectSortDir === 'asc' ? '△' : '▽' }}</span>
                </th>
                <th>{{ $t('admin.owner') }}</th>
                <th>{{ $t('common.status') }}</th>
                <th>{{ $t('admin.open_cards') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="project in sortedProjects" :key="project.id">
                <td>
                  <div class="entity-cell">
                    <img v-if="projectAvatar(project)" :src="projectAvatar(project)" class="entity-avatar" alt="" />
                    <span v-else class="project-dot" :style="{ background: project.color || '#6366f1' }"></span>
                    <button v-if="!showDeletedProjects" type="button" class="name-link" @click="openEditProject(project)">{{ project.name }}</button>
                    <strong v-else>{{ project.name }}</strong>
                  </div>
                  <br>
                  <small>{{ project.slug }} &middot; <code>{{ project.key_prefix }}</code></small>
                  <br><small>{{ project.description }}</small>
                </td>
                <td>
                  <small>{{ project.created_by?.display_name || project.created_by?.username }}</small>
                </td>
                <td>
                  <span v-if="showDeletedProjects" class="badge badge-deleted">{{ $t('admin.deleted') }}</span>
                  <span v-else :class="['badge', project.is_archived ? 'badge-inactive' : 'badge-active']">
                    {{ project.is_archived ? $t('admin.archived') : $t('admin.active') }}
                  </span>
                </td>
                <td>
                  <span class="open-cards-count">{{ project.open_card_count }}</span>
                </td>
                <td>
                  <div class="actions-cell">
                  <template v-if="showDeletedProjects">
                    <button class="btn btn-ghost btn-sm" @click="restoreProject(project)">{{ $t('admin.restore') }}</button>
                    <button class="btn btn-ghost btn-sm btn-danger" @click="purgeProject(project)">{{ $t('admin.purge_project') }}</button>
                  </template>
                  <template v-else>
                    <button class="btn btn-ghost btn-sm" @click="openEditProject(project)">{{ $t('common.edit') }}</button>
                    <button class="btn btn-ghost btn-sm" @click="toggleArchive(project)">
                      {{ project.is_archived ? $t('admin.unarchive') : $t('project.archive') }}
                    </button>
                    <button class="btn btn-ghost btn-sm btn-danger" @click="deleteProject(project)">{{ $t('common.delete') }}</button>
                  </template>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Groups tab -->
        <div v-show="tab === 'groups'" role="tabpanel" id="tab-panel-groups" aria-labelledby="tab-btn-groups">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="openCreateGroup">+ {{ $t('groups.create') }}</button>
            <input v-model="groupSearch" class="form-input admin-search" type="search" :placeholder="$t('common.search')" aria-label="Search groups" />
          </div>
          <div v-if="loadingGroups" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>
          <table v-else class="data-table">
            <thead>
              <tr>
                <th class="sortable-th" @click="toggleGroupSort()">
                  {{ $t('groups.name') }}
                  <span class="sort-indicator">{{ groupSortDir === 'asc' ? '△' : '▽' }}</span>
                </th>
                <th>{{ $t('groups.members') }}</th>
                <th>{{ $t('admin.projects') }}</th>
                <th>{{ $t('customer.title') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="g in sortedGroups" :key="g.id">
                <td>
                  <div class="entity-cell">
                    <img v-if="groupAvatar(g)" :src="groupAvatar(g)" class="entity-avatar" alt="" />
                    <button type="button" class="name-link" @click="openEditGroup(g)">{{ g.name }}</button>
                  </div>
                  <div v-if="g.description" style="font-size:12px;color:var(--color-text-muted)">{{ g.description }}</div>
                </td>
                <td>{{ g.member_count }}</td>
                <td>{{ g.project_count }}</td>
                <td>{{ g.customer_count }}</td>
                <td>
                  <div class="actions-cell">
                    <button class="btn btn-ghost btn-sm" @click="openEditGroup(g)">{{ $t('common.edit') }}</button>
                    <button class="btn btn-ghost btn-sm" @click="openGroupDetail(g)">{{ $t('groups.members') }}</button>
                    <button class="btn btn-ghost btn-sm btn-danger" @click="deleteGroup(g)">{{ $t('common.delete') }}</button>
                  </div>
                </td>
              </tr>
              <tr v-if="!groups.length">
                <td colspan="5" style="text-align:center;color:var(--color-text-muted)">{{ $t('groups.no_groups') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Customers tab -->
        <div v-show="tab === 'customers'" role="tabpanel" id="tab-panel-customers" aria-labelledby="tab-btn-customers">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="showCreateCustomer = true">+ {{ $t('customer.new_customer') }}</button>
            <input v-model="customerSearch" class="form-input admin-search" type="search" :placeholder="$t('common.search')" aria-label="Search customers" />
          </div>
          <div v-if="loadingCustomers" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>
          <table v-else class="data-table">
            <thead>
              <tr>
                <th style="width: 48px;"></th>
                <th class="sortable-th" @click="toggleCustomerSort()">
                  {{ $t('customer.name') }}
                  <span class="sort-indicator">{{ customerSortDir === 'asc' ? '△' : '▽' }}</span>
                </th>
                <th>{{ $t('contract.contracts') }}</th>
                <th>{{ $t('customer.projects') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="c in sortedCustomers" :key="c.id">
                <td>
                  <div class="customer-avatar-container">
                    <img v-if="c.logo_url" :src="resolveAssetUrl(c.logo_url)" class="customer-avatar" alt="" />
                    <div v-else class="customer-avatar-placeholder">{{ c.name.charAt(0) }}</div>
                  </div>
                </td>
                <td>
                  <RouterLink :to="`/customers/${c.id}`" style="font-weight:600">{{ c.name }}</RouterLink>
                  <div v-if="c.description" style="font-size:12px;color:var(--color-text-muted)">{{ c.description }}</div>
                </td>
                <td>{{ c.contract_count }}</td>
                <td>{{ c.project_count }}</td>
                <td>
                  <div class="actions-cell">
                    <button class="btn btn-ghost btn-sm" @click="openEditCustomer(c)">{{ $t('common.edit') }}</button>
                    <button class="btn btn-ghost btn-sm btn-danger" @click="deleteAdminCustomer(c)">{{ $t('common.delete') }}</button>
                  </div>
                </td>
              </tr>
              <tr v-if="!adminCustomers.length">
                <td colspan="5" style="text-align:center;color:var(--color-text-muted)">{{ $t('customer.no_customers') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Settings tab -->
        <div v-show="tab === 'settings'" role="tabpanel" id="tab-panel-settings" aria-labelledby="tab-btn-settings">
          <div class="settings-section">
            <h2>{{ $t('admin.system_settings') }}</h2>

            <div class="form-group" style="max-width:400px">
              <label class="toggle-row">
                <span>{{ $t('admin.registration_enabled') }}<HelpIcon i18n-key="admin.registration_hint" :label="$t('admin.registration_enabled')" /></span>
                <input type="checkbox" v-model="systemSettings.registration_enabled" @change="saveGeneralSettings" />
              </label>
            </div>

            <div class="mfa-settings-group">
              <h3 class="mfa-settings-heading">{{ $t('mfa.admin_group_title') }}</h3>
              <p class="form-hint mfa-settings-hint">{{ $t('mfa.admin_group_hint') }}</p>

              <div class="form-group" style="max-width:400px">
                <label class="toggle-row">
                  <span>{{ $t('mfa.enforce_label') }}<HelpIcon i18n-key="mfa.enforce_hint" :label="$t('mfa.enforce_label')" /></span>
                  <input type="checkbox" v-model="systemSettings.mfa_required" @change="saveMFASettings" />
                </label>
              </div>

              <div class="form-group" style="max-width:400px">
                <label class="form-label" for="sys-mfa-remember-devices">
                  {{ $t('mfa.remember_devices_label') }}
                  <HelpIcon i18n-key="mfa.remember_devices_hint" :label="$t('mfa.remember_devices_label')" />
                </label>
                <select id="sys-mfa-remember-devices" class="form-input" v-model="systemSettings.mfa_remember_devices" @change="onMfaRememberDevicesChange">
                  <option value="disabled">{{ $t('mfa.remember_devices_disabled') }}</option>
                  <option value="week">{{ $t('mfa.remember_devices_week') }}</option>
                  <option value="week_month">{{ $t('mfa.remember_devices_week_month') }}</option>
                </select>
              </div>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-allowed-ips">
                {{ $t('admin.allowed_ips_label') }}
                <HelpIcon i18n-key="admin.allowed_ips_hint" :label="$t('admin.allowed_ips_label')" />
              </label>
              <input id="sys-allowed-ips" class="form-input" v-model="systemSettings.allowed_ips"
                :placeholder="$t('admin.allowed_ips_placeholder')"
                spellcheck="false" autocorrect="off" autocapitalize="off"
                @change="saveSecuritySettings" />
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="toggle-row">
                <span>{{ $t('admin.scrum_storypoints_enabled') }}<HelpIcon i18n-key="admin.scrum_storypoints_hint" :label="$t('admin.scrum_storypoints_enabled')" /></span>
                <input type="checkbox" v-model="systemSettings.scrum_storypoints_enabled" @change="saveGeneralSettings" />
              </label>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="toggle-row">
                <span>{{ $t('admin.gravatar_enabled') }}<HelpIcon i18n-key="admin.gravatar_hint" :label="$t('admin.gravatar_enabled')" /></span>
                <input type="checkbox" v-model="systemSettings.gravatar_enabled" @change="saveGeneralSettings" />
              </label>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="toggle-row">
                <span>{{ $t('admin.external_image_proxy_enabled') }}<HelpIcon i18n-key="admin.external_image_proxy_hint" :label="$t('admin.external_image_proxy_enabled')" /></span>
                <input type="checkbox" v-model="systemSettings.external_image_proxy_enabled" @change="saveGeneralSettings" />
              </label>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-session-timeout">
                {{ $t('admin.session_timeout') }}
                <HelpIcon i18n-key="admin.session_timeout_hint" :label="$t('admin.session_timeout')" />
              </label>
              <div class="form-row" style="align-items:center;gap:8px">
                <input id="sys-session-timeout" class="form-input" type="number" min="0" v-model.number="systemSettings.session_timeout_minutes" @change="saveGeneralSettings" style="width:120px" />
                <span class="form-hint" style="margin:0">{{ $t('admin.session_timeout_unit') }}</span>
              </div>
            </div>

            <h3 class="settings-subsection">
              {{ $t('admin.global_defaults_title') }}
              <HelpIcon i18n-key="admin.global_defaults_hint" :label="$t('admin.global_defaults_title')" />
            </h3>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-datetime-format">{{ $t('settings.date_time_format') }}</label>
              <select id="sys-datetime-format" class="form-input" v-model="systemSettings.default_date_time_format" @change="saveGeneralSettings">
                <option value="YYYY-MM-DD HH:mm">YYYY-MM-DD HH:mm (ISO)</option>
                <option value="DD/MM/YYYY HH:mm">DD/MM/YYYY HH:mm</option>
                <option value="MM/DD/YYYY hh:mm a">MM/DD/YYYY hh:mm a</option>
                <option value="DD-MM-YYYY HH:mm">DD-MM-YYYY HH:mm</option>
                <option value="DD.MM.YYYY HH:mm">DD.MM.YYYY HH:mm</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-timezone">{{ $t('settings.timezone') }}</label>
              <select id="sys-timezone" class="form-input" v-model="systemSettings.default_timezone" @change="saveGeneralSettings">
                <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-theme">{{ $t('settings.theme') }}</label>
              <select id="sys-theme" class="form-input" v-model="systemSettings.default_theme" @change="saveGeneralSettings">
                <option value="light">{{ $t('settings.theme_light') }}</option>
                <option value="dark">{{ $t('settings.theme_dark') }}</option>
                <option value="system">{{ $t('settings.theme_system') }}</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-font">{{ $t('settings.font') }}</label>
              <select id="sys-font" class="form-input" v-model="systemSettings.default_font" @change="saveGeneralSettings">
                <option value="system">{{ $t('settings.font_system') }}</option>
                <option value="Inter, sans-serif">Inter</option>
                <option value="'Roboto', sans-serif">Roboto</option>
                <option value="'Open Sans', sans-serif">Open Sans</option>
                <option value="'Source Code Pro', monospace">Source Code Pro (monospace)</option>
                <option value="Georgia, serif">Georgia (serif)</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-font-size">{{ $t('settings.font_size') }}</label>
              <select id="sys-font-size" class="form-input" v-model="systemSettings.default_font_size" @change="saveGeneralSettings">
                <option value="12">12px</option>
                <option value="13">13px</option>
                <option value="14">14px</option>
                <option value="15">15px</option>
                <option value="16">16px</option>
                <option value="18">18px</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-locale">{{ $t('common.language') }}</label>
              <select id="sys-locale" class="form-input" v-model="systemSettings.default_locale" @change="saveGeneralSettings">
                <option value="en">English</option>
                <option value="nl">Nederlands</option>
                <option value="de">Deutsch</option>
                <option value="fr">Français</option>
                <option value="es">Español</option>
                <option value="da">Dansk</option>
                <option value="sv">Svenska</option>
                <option value="nb">Norsk</option>
                <option value="fi">Suomi</option>
                <option value="is">Íslenska</option>
                <option value="pt">Português</option>
                <option value="it">Italiano</option>
              </select>
            </div>

            <h3 class="settings-subsection">
              {{ $t('admin.project_defaults_title') }}
              <HelpIcon i18n-key="admin.default_columns_hint" :label="$t('admin.project_defaults_title')" />
            </h3>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-default-columns">
                {{ $t('admin.default_columns') }}
                <HelpIcon i18n-key="admin.default_columns_each_line" :label="$t('admin.default_columns')" />
              </label>
              <textarea id="sys-default-columns" class="form-input" v-model="systemSettings.default_columns" rows="4" style="font-family:monospace;resize:vertical" :placeholder="'Backlog\nIn Progress\nDone'"></textarea>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-default-labels">{{ $t('admin.default_labels') }}</label>
              <textarea id="sys-default-labels" class="form-input" v-model="systemSettings.default_labels" rows="4" style="font-family:monospace;resize:vertical" :placeholder="'Bug\nFeature\nDesign\nContent'"></textarea>
              <p class="form-hint">{{ $t('admin.default_labels_each_line') }}</p>
            </div>

            <div style="max-width:400px;margin-top:8px">
              <button class="btn btn-primary btn-sm" @click="saveGeneralSettings">{{ $t('common.save') }}</button>
            </div>

            <h3 class="settings-subsection">
              {{ $t('admin.smtp_title') }}
              <HelpIcon i18n-key="admin.smtp_hint" :label="$t('admin.smtp_title')" />
            </h3>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:3">
                <label class="form-label" for="sys-smtp-host">{{ $t('admin.smtp_host') }}</label>
                <input id="sys-smtp-host" class="form-input" v-model="systemSettings.smtp_host" :placeholder="$t('admin.smtp_host_placeholder')" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-smtp-port">{{ $t('admin.smtp_port') }}</label>
                <input id="sys-smtp-port" class="form-input" v-model="systemSettings.smtp_port" type="number" placeholder="587" />
              </div>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-smtp-from">{{ $t('admin.smtp_from') }}</label>
              <input id="sys-smtp-from" class="form-input" v-model="systemSettings.smtp_from" type="text" placeholder="WarmDesk &lt;noreply@example.com&gt;" />
            </div>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-smtp-username">{{ $t('admin.smtp_username') }}</label>
                <input id="sys-smtp-username" class="form-input" v-model="systemSettings.smtp_username" autocomplete="off" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-smtp-password">{{ $t('admin.smtp_password') }}</label>
                <input id="sys-smtp-password" class="form-input" v-model="systemSettings.smtp_password" type="password" autocomplete="new-password" :placeholder="smtpPasswordPlaceholder" />
              </div>
            </div>

            <div class="form-actions" style="max-width:500px">
              <button class="btn btn-primary" @click="saveSmtpSettings">{{ $t('common.save') }}</button>
            </div>

            <div class="form-group" style="max-width:500px;margin-top:16px">
              <label class="form-label" for="sys-smtp-test">{{ $t('admin.smtp_test_title') }}</label>
              <div style="display:flex;gap:8px">
                <input id="sys-smtp-test" class="form-input" v-model="smtpTestEmail" type="email" :placeholder="$t('admin.smtp_test_placeholder')" style="flex:1" />
                <button class="btn btn-secondary" :disabled="smtpTestSending || !smtpTestEmail" @click="sendSmtpTest">
                  {{ smtpTestSending ? $t('admin.smtp_test_sending') : $t('admin.smtp_test_send') }}
                </button>
              </div>
            </div>

            <h3 class="settings-subsection">
              {{ $t('admin.imap_title') }}
              <HelpIcon i18n-key="admin.imap_hint" :label="$t('admin.imap_title')" />
            </h3>

            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.imap_enabled') }}</span>
                <input type="checkbox" v-model="systemSettings.imap_enabled" />
              </label>
            </div>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:3">
                <label class="form-label" for="sys-imap-host">{{ $t('admin.imap_host') }}</label>
                <input id="sys-imap-host" class="form-input" v-model="systemSettings.imap_host" :placeholder="$t('admin.imap_host_placeholder')" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-imap-port">{{ $t('admin.imap_port') }}</label>
                <input id="sys-imap-port" class="form-input" v-model="systemSettings.imap_port" type="number" placeholder="993" />
              </div>
            </div>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-imap-username">{{ $t('admin.imap_username') }}</label>
                <input id="sys-imap-username" class="form-input" v-model="systemSettings.imap_username" autocomplete="off" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-imap-password">{{ $t('admin.imap_password') }}</label>
                <input id="sys-imap-password" class="form-input" v-model="systemSettings.imap_password" type="password" autocomplete="new-password" :placeholder="imapPasswordPlaceholder" :disabled="systemSettings.imap_auth_mechanism === 'oauth2'" />
              </div>
            </div>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-imap-auth">{{ $t('admin.imap_auth_mechanism') }}<HelpIcon i18n-key="help.fields.imap_auth_mechanism" :label="$t('admin.imap_auth_mechanism')" /></label>
                <select id="sys-imap-auth" class="form-input" v-model="systemSettings.imap_auth_mechanism">
                  <option value="plain">{{ $t('admin.imap_auth_plain') }}</option>
                  <option value="oauth2">{{ $t('admin.imap_auth_oauth2') }}</option>
                </select>
              </div>
              <div class="form-group" style="flex:1" v-if="systemSettings.imap_auth_mechanism === 'oauth2'">
                <label class="form-label" for="sys-imap-oauth2-provider">{{ $t('admin.imap_oauth2_provider') }}<HelpIcon i18n-key="help.fields.imap_oauth2_provider" :label="$t('admin.imap_oauth2_provider')" /></label>
                <select id="sys-imap-oauth2-provider" class="form-input" v-model="systemSettings.imap_oauth2_provider">
                  <option value="">{{ $t('common.select') }}</option>
                  <option value="google">{{ $t('admin.imap_oauth2_google') }}</option>
                  <option value="office365">{{ $t('admin.imap_oauth2_office365') }}</option>
                </select>
              </div>
            </div>

            <div class="form-group" v-if="systemSettings.imap_auth_mechanism === 'oauth2'" style="max-width:500px">
              <template v-if="imapOAuth2Connected">
                <span class="badge badge-success" style="margin-right:8px">{{ $t('admin.imap_oauth2_connected') }}</span>
                <button class="btn btn-secondary btn-sm" @click="disconnectImapOAuth2">{{ $t('admin.imap_oauth2_revoke') }}</button>
              </template>
              <template v-else>
                <button class="btn btn-secondary btn-sm" @click="authorizeImapOAuth2" :disabled="imapOAuth2Connecting || !systemSettings.imap_oauth2_provider">
                  {{ imapOAuth2Connecting ? $t('common.loading') : $t('admin.imap_oauth2_authorize') }}
                </button>
              </template>
            </div>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-imap-mailbox">{{ $t('admin.imap_mailbox') }}</label>
                <input id="sys-imap-mailbox" class="form-input" v-model="systemSettings.imap_mailbox" placeholder="INBOX" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label" for="sys-imap-poll">{{ $t('admin.imap_poll_interval') }}<HelpIcon i18n-key="help.fields.imap_poll_interval" :label="$t('admin.imap_poll_interval')" /></label>
                <input id="sys-imap-poll" class="form-input" v-model="systemSettings.imap_poll_interval" type="number" placeholder="60" />
              </div>
            </div>

            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.imap_use_tls') }}</span>
                <input type="checkbox" v-model="systemSettings.imap_use_tls" />
              </label>
              <p class="form-hint">{{ $t('admin.imap_use_tls_hint') }}</p>
            </div>

            <div class="form-actions" style="max-width:500px">
              <button class="btn btn-primary" @click="saveImapSettings">{{ $t('common.save') }}</button>
              <button class="btn btn-secondary btn-sm" @click="testImap" :disabled="imapTesting">{{ imapTesting ? $t('common.loading') : $t('admin.imap_test') }}</button>
              <button class="btn btn-secondary btn-sm" @click="pollImap" :disabled="imapPolling">{{ imapPolling ? $t('common.loading') : $t('admin.imap_poll') }}</button>
            </div>

            <h3 class="settings-subsection">
              {{ $t('admin.branding_title') }}
              <HelpIcon i18n-key="admin.branding_hint" :label="$t('admin.branding_title')" />
            </h3>

            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.login_branding_enabled') }}<HelpIcon i18n-key="admin.login_branding_hint" :label="$t('admin.login_branding_enabled')" /></span>
                <input type="checkbox" v-model="systemSettings.login_branding_enabled" />
              </label>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-company-name">{{ $t('admin.company_name') }}</label>
              <input id="sys-company-name" class="form-input" v-model="systemSettings.company_name" :placeholder="$t('admin.company_name_placeholder')" />
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-company-logo">{{ $t('admin.company_logo') }}</label>
              <input id="sys-company-logo" class="form-input" v-model="systemSettings.company_logo" :placeholder="'https://...'" style="margin-bottom:8px" />
              <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
                <button class="btn btn-secondary btn-sm" @click="$refs.logoFileInput.click()">{{ $t('admin.company_logo_upload') }}</button>
                <button v-if="systemSettings.company_logo" class="btn btn-danger btn-sm" @click="clearCompanyLogo">{{ $t('common.clear') }}</button>
                <span class="form-hint" style="margin:0">{{ $t('admin.company_logo_hint') }}</span>
              </div>
              <input ref="logoFileInput" type="file" accept="image/*" style="display:none" @change="onLogoFileSelected" />
              <div v-if="systemSettings.company_logo" style="margin-top:8px">
                <span class="form-hint">{{ $t('admin.company_logo_preview') }}</span>
                <div style="margin-top:6px;padding:8px;border:1px solid var(--color-border);border-radius:var(--radius);display:inline-block;background:var(--color-bg)">
                  <img :src="resolveAssetUrl(systemSettings.company_logo)" alt="Logo preview" style="max-height:60px;max-width:200px;object-fit:contain" @error="systemSettings.company_logo=''" />
                </div>
              </div>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label" for="sys-company-logo-dark">{{ $t('admin.company_logo_dark') }}</label>
              <input id="sys-company-logo-dark" class="form-input" v-model="systemSettings.company_logo_dark" :placeholder="'https://...'" style="margin-bottom:8px" />
              <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
                <button class="btn btn-secondary btn-sm" @click="$refs.logoDarkFileInput.click()">{{ $t('admin.company_logo_upload') }}</button>
                <button v-if="systemSettings.company_logo_dark" class="btn btn-danger btn-sm" @click="clearCompanyLogoDark">{{ $t('common.clear') }}</button>
                <span class="form-hint" style="margin:0">{{ $t('admin.company_logo_hint') }}</span>
              </div>
              <input ref="logoDarkFileInput" type="file" accept="image/*" style="display:none" @change="onLogoDarkFileSelected" />
              <div v-if="systemSettings.company_logo_dark" style="margin-top:8px">
                <span class="form-hint">{{ $t('admin.company_logo_preview') }}</span>
                <div style="margin-top:6px;padding:8px;border:1px solid var(--color-border);border-radius:var(--radius);display:inline-block;background:var(--color-surface)">
                  <img :src="resolveAssetUrl(systemSettings.company_logo_dark)" alt="Logo dark preview" style="max-height:60px;max-width:200px;object-fit:contain" @error="systemSettings.company_logo_dark=''" />
                </div>
              </div>
            </div>

            <div class="form-actions" style="max-width:400px">
              <button class="btn btn-primary" @click="saveBrandingSettings">{{ $t('common.save') }}</button>
            </div>

            <h3 class="settings-subsection">
              {{ $t('admin.password_policy_title') }}
              <HelpIcon i18n-key="admin.password_policy_hint" :label="$t('admin.password_policy_title')" />
            </h3>

            <div class="form-group" style="max-width:240px">
              <label class="form-label" for="sys-pwd-min-len">{{ $t('admin.password_min_length') }}<HelpIcon i18n-key="help.fields.password_min_length" :label="$t('admin.password_min_length')" /></label>
              <input id="sys-pwd-min-len" class="form-input" type="number" min="8" max="128" v-model.number="systemSettings.password_min_length" style="width:100px" />
            </div>

            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.password_require_upper') }}</span>
                <input type="checkbox" v-model="systemSettings.password_require_upper" />
              </label>
            </div>
            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.password_require_lower') }}</span>
                <input type="checkbox" v-model="systemSettings.password_require_lower" />
              </label>
            </div>
            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.password_require_digit') }}</span>
                <input type="checkbox" v-model="systemSettings.password_require_digit" />
              </label>
            </div>
            <div class="form-group">
              <label class="toggle-row">
                <span>{{ $t('admin.password_require_special') }}</span>
                <input type="checkbox" v-model="systemSettings.password_require_special" />
              </label>
            </div>

            <div class="form-group" style="max-width:400px;margin-top:8px">
              <label class="form-label" for="sys-pwd-change-period">{{ $t('admin.password_change_period') }}</label>
              <div class="form-row" style="align-items:center;gap:8px">
                <input id="sys-pwd-change-period" class="form-input" type="number" min="0" max="3650" v-model.number="systemSettings.password_change_period_days" style="width:100px" />
                <span class="form-hint" style="margin:0">{{ $t('admin.password_change_period_unit') }}</span>
              </div>
              <p class="form-hint">{{ $t('admin.password_change_period_hint') }}</p>
            </div>

            <div style="max-width:400px;margin-top:8px">
              <button class="btn btn-primary btn-sm" @click="savePasswordPolicy">{{ $t('common.save') }}</button>
            </div>
          </div>

          <!-- Metrics access log -->
          <div style="margin-top:32px">
            <h3 class="form-section-title">{{ $t('admin.metrics_access_title') }}</h3>
            <div v-if="systemSettings.metrics_last_access" style="font-size:13px;color:var(--color-text-muted);display:flex;flex-direction:column;gap:4px">
              <span>{{ $t('admin.metrics_last_access') }}: <strong>{{ formatDateTime(systemSettings.metrics_last_access) }}</strong></span>
              <span>{{ $t('admin.metrics_last_access_result') }}: <strong :style="{ color: systemSettings.metrics_last_access_success === 'true' ? 'var(--color-success)' : 'var(--color-danger)' }">{{ systemSettings.metrics_last_access_success === 'true' ? $t('admin.metrics_access_success') : $t('admin.metrics_access_failed') }}</strong></span>
            </div>
            <span v-else style="font-size:13px;color:var(--color-text-muted)">{{ $t('admin.metrics_never_accessed') }}</span>
          </div>

        </div>

        <!-- News tab -->
        <div v-show="tab === 'news'" role="tabpanel" id="tab-panel-news" aria-labelledby="tab-btn-news">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="openCreateNews">+ {{ $t('admin.news_create') }}</button>
          </div>
          <div v-if="newsLoading" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>
          <table v-else class="data-table">
            <thead>
              <tr>
                <th>{{ $t('admin.news_title_col') }}</th>
                <th>{{ $t('admin.news_start_date') }}</th>
                <th>{{ $t('admin.news_end_date') }}</th>
                <th>{{ $t('admin.news_active') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!newsItems.length">
                <td colspan="5" style="text-align:center;color:var(--color-text-muted)">{{ $t('admin.news_empty') }}</td>
              </tr>
              <tr v-for="item in newsItems" :key="item.id">
                <td><button type="button" class="name-link" @click="openEditNews(item)">{{ item.title }}</button><br><small style="color:var(--color-text-muted);white-space:pre-wrap">{{ item.text?.slice(0, 80) }}{{ item.text?.length > 80 ? '…' : '' }}</small></td>
                <td><small>{{ item.start_date ? formatDateTime(item.start_date) : '—' }}</small></td>
                <td><small>{{ item.end_date ? formatDateTime(item.end_date) : '—' }}</small></td>
                <td>
                  <span :class="['badge', item.active ? 'badge-active' : 'badge-inactive']">
                    {{ item.active ? $t('admin.active') : $t('admin.inactive') }}
                  </span>
                </td>
                <td>
                  <div class="actions-cell">
                    <button class="btn btn-ghost btn-sm" @click="openEditNews(item)">{{ $t('common.edit') }}</button>
                    <button class="btn btn-ghost btn-sm btn-danger" @click="deleteNewsItem(item)">{{ $t('common.delete') }}</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Time tracking tab -->
        <div v-show="tab === 'time-tracking'" role="tabpanel" id="tab-panel-time-tracking" aria-labelledby="tab-btn-time-tracking">
          <div class="tab-toolbar">
            <h2 class="tab-section-title">{{ $t('admin.tt_projects_title') }}</h2>
            <button v-if="!addingTTProject && !editingTTProject" class="btn btn-primary btn-sm" @click="addingTTProject = true; nextTick(() => newTTProjRef?.focus())">+ {{ $t('timeTracking.tt_project_add') }}</button>
          </div>
          <div v-if="loadingTTProjects" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>
          <div v-else>
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ $t('timeTracking.tt_project_name') }}</th>
                  <th>{{ $t('admin.owner') }}</th>
                  <th>{{ $t('timeTracking.undeclarable') }}</th>
                  <th>{{ $t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!editingTTProject && addingTTProject">
                  <td colspan="4" style="padding:8px">
                    <div class="tt-admin-add-row">
                      <label class="sr-only" for="adm-tt-proj-name">{{ $t('timeTracking.tt_project_name') }}</label>
                      <input id="adm-tt-proj-name" class="form-input" v-model="newTTProject.name" :placeholder="$t('timeTracking.tt_project_name')" @keydown.enter="confirmAddTTProject" @keydown.escape="addingTTProject = false" ref="newTTProjRef" style="flex:1" />
                      <label class="sr-only" for="adm-tt-proj-color">{{ $t('timeTracking.tt_project_color') }}</label>
                      <input id="adm-tt-proj-color" type="color" class="tt-admin-color" v-model="newTTProject.color" :aria-label="$t('timeTracking.tt_project_color')" />
                      <label class="sr-only" for="adm-tt-proj-undecl">{{ $t('timeTracking.undeclarable') }}</label>
                      <input id="adm-tt-proj-undecl" class="form-input" v-model="newTTProject.undeclStr" :placeholder="'0:00'" style="width:80px" :title="$t('timeTracking.undeclarable_per_entry')" @keydown.enter="confirmAddTTProject" @keydown.escape="addingTTProject = false" />
                      <button class="btn btn-primary btn-sm" @click="confirmAddTTProject">{{ $t('common.create') }}</button>
                      <button class="btn btn-secondary btn-sm" @click="addingTTProject = false">{{ $t('common.cancel') }}</button>
                    </div>
                  </td>
                </tr>
                <tr v-for="p in adminTTProjects" :key="p.id">
                  <template v-if="editingTTProject && editingTTProject.id === p.id">
                    <td colspan="4" style="padding:8px">
                      <div class="tt-admin-add-row">
                        <input class="form-input" v-model="editingTTProject.name" @keydown.enter="saveEditTTProject" @keydown.escape="editingTTProject = null" style="flex:1" />
                        <input type="color" class="tt-admin-color" v-model="editingTTProject.color" :aria-label="$t('timeTracking.tt_project_color')" />
                        <input class="form-input" v-model="editingTTProject.undeclStr" :placeholder="'0:00'" style="width:80px" :title="$t('timeTracking.undeclarable_per_entry')" @keydown.enter="saveEditTTProject" @keydown.escape="editingTTProject = null" />
                        <button class="btn btn-primary btn-sm" @click="saveEditTTProject">{{ $t('common.save') }}</button>
                        <button class="btn btn-secondary btn-sm" @click="editingTTProject = null">{{ $t('common.cancel') }}</button>
                      </div>
                    </td>
                  </template>
                  <template v-else>
                    <td>
                      <span class="tt-admin-dot" :style="{ background: p.color || '#6366f1' }"></span>
                      <button type="button" class="name-link" @click="startEditTTProject(p)">{{ p.name }}</button>
                      <span v-if="isGlobalTTEntity(p)" class="ttp-badge">{{ $t('timeTracking.tt_project_global') }}</span>
                    </td>
                    <td><small>{{ p.created_by?.display_name || p.created_by?.username }}</small></td>
                    <td><span v-if="p.undeclarable_minutes > 0" class="tt-admin-undecl-badge">-{{ fmtTTTime(p.undeclarable_minutes) }}</span></td>
                    <td>
                      <div class="actions-cell">
                        <button class="btn btn-ghost btn-sm" @click="startEditTTProject(p)">{{ $t('common.edit') }}</button>
                        <button class="btn btn-ghost btn-sm btn-danger" @click="deleteTTProject(p)">{{ $t('common.delete') }}</button>
                      </div>
                    </td>
                  </template>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="tab-toolbar" style="margin-top:48px">
            <h2 class="tab-section-title">{{ $t('admin.tt_customers_title') }}</h2>
            <button v-if="!addingTTCustomer && !editingTTCustomer" class="btn btn-primary btn-sm" @click="addingTTCustomer = true; nextTick(() => newTTCustRef?.focus())">+ {{ $t('timeTracking.tt_customer_add') }}</button>
          </div>
          <div v-if="loadingTTCustomers" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>
          <div v-else>
            <table class="data-table">
              <thead>
                <tr>
                  <th>{{ $t('timeTracking.tt_customer_name') }}</th>
                  <th>{{ $t('admin.owner') }}</th>
                  <th>{{ $t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!editingTTCustomer && addingTTCustomer">
                  <td colspan="3" style="padding:8px">
                    <div class="tt-admin-add-row">
                      <label class="sr-only" for="adm-tt-cust-name">{{ $t('timeTracking.tt_customer_name') }}</label>
                      <input id="adm-tt-cust-name" class="form-input" v-model="newTTCustomer.name" :placeholder="$t('timeTracking.tt_customer_name')" @keydown.enter="confirmAddTTCustomer" @keydown.escape="addingTTCustomer = false" ref="newTTCustRef" style="flex:1" />
                      <button class="btn btn-primary btn-sm" @click="confirmAddTTCustomer">{{ $t('common.create') }}</button>
                      <button class="btn btn-secondary btn-sm" @click="addingTTCustomer = false">{{ $t('common.cancel') }}</button>
                    </div>
                  </td>
                </tr>
                <tr v-for="c in adminTTCustomers" :key="c.id">
                  <template v-if="editingTTCustomer && editingTTCustomer.id === c.id">
                    <td colspan="3" style="padding:8px">
                      <div class="tt-admin-add-row">
                        <input class="form-input" v-model="editingTTCustomer.name" @keydown.enter="saveEditTTCustomer" @keydown.escape="editingTTCustomer = null" style="flex:1" />
                        <button class="btn btn-primary btn-sm" @click="saveEditTTCustomer">{{ $t('common.save') }}</button>
                        <button class="btn btn-secondary btn-sm" @click="editingTTCustomer = null">{{ $t('common.cancel') }}</button>
                      </div>
                    </td>
                  </template>
                  <template v-else>
                    <td>
                      <button type="button" class="name-link" @click="startEditTTCustomer(c)">{{ c.name }}</button>
                      <span v-if="isGlobalTTEntity(c)" class="ttp-badge">{{ $t('timeTracking.tt_customer_global') }}</span>
                    </td>
                    <td><small>{{ c.created_by?.display_name || c.created_by?.username }}</small></td>
                    <td>
                      <div class="actions-cell">
                        <button class="btn btn-ghost btn-sm" @click="startEditTTCustomer(c)">{{ $t('common.edit') }}</button>
                        <button class="btn btn-ghost btn-sm btn-danger" @click="deleteTTCustomer(c)">{{ $t('common.delete') }}</button>
                      </div>
                    </td>
                  </template>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- SLA policies tab -->
        <div v-show="tab === 'sla'" role="tabpanel" id="tab-panel-sla" aria-labelledby="tab-btn-sla">
          <SlaPoliciesTab />
        </div>

        <!-- Macros tab -->
        <div v-show="tab === 'macros'" role="tabpanel" id="tab-panel-macros" aria-labelledby="tab-btn-macros">
          <MacrosTab />
        </div>

        <!-- Ticket checklist templates tab -->
        <div v-show="tab === 'ticket-checklists'" role="tabpanel" id="tab-panel-ticket-checklists" aria-labelledby="tab-btn-ticket-checklists">
          <TicketChecklistTemplatesTab />
        </div>

        <!-- Backup / Restore tab -->
        <div v-show="tab === 'backup'" role="tabpanel" id="tab-panel-backup" aria-labelledby="tab-btn-backup">
          <!-- Scheduled backups -->
          <div style="margin-bottom:24px">
            <h3 class="form-section-title">{{ $t('admin.backup_schedule_title') }}</h3>
            <div class="form-row" style="align-items:end;flex-wrap:wrap;gap:12px;margin-top:8px">
              <div class="form-group" style="margin-bottom:0;flex:0 0 160px">
                <label class="form-label" for="sys-backup-schedule">{{ $t('admin.backup_schedule_label') }}<HelpIcon i18n-key="help.fields.backup_schedule" :label="$t('admin.backup_schedule_label')" /></label>
                <select id="sys-backup-schedule" class="form-input" v-model="systemSettings.backup_schedule">
                  <option value="disabled">{{ $t('admin.backup_schedule_disabled') }}</option>
                  <option value="6h">{{ $t('admin.backup_schedule_6h') }}</option>
                  <option value="8h">{{ $t('admin.backup_schedule_8h') }}</option>
                  <option value="12h">{{ $t('admin.backup_schedule_12h') }}</option>
                  <option value="24h">{{ $t('admin.backup_schedule_24h') }}</option>
                </select>
              </div>
              <div class="form-group" style="margin-bottom:0;flex:0 0 160px" v-if="systemSettings.backup_schedule !== 'disabled'">
                <label class="form-label" for="sys-backup-time">{{ $t('admin.backup_start_time') }}</label>
                <input
                  id="sys-backup-time"
                  type="text"
                  class="form-input"
                  v-model="backupStartTimeDisplay"
                  :placeholder="backupTimePlaceholder"
                  @blur="onBackupStartTimeBlur"
                />
              </div>
              <div class="form-group" style="flex:0 0 100px" v-if="systemSettings.backup_schedule !== 'disabled'">
                <label class="form-label" for="sys-backup-keep">{{ $t('admin.backup_keep_label') }}</label>
                <input id="sys-backup-keep" class="form-input" type="number" min="1" max="100" v-model.number="systemSettings.backup_keep" />
              </div>
            </div>

            <div v-if="systemSettings.backup_schedule !== 'disabled'" style="margin-top:8px;display:flex;flex-direction:column;gap:4px;font-size:13px;color:var(--color-text-muted)">
              <span>{{ $t('admin.backup_last_run') }}: <strong>{{ backupLastRun }}</strong></span>
              <span>{{ $t('admin.backup_next_run') }}: <strong>{{ backupNextRun }}</strong></span>
            </div>

            <div style="margin-top:8px;display:flex;flex-direction:column;gap:6px;max-width:450px">
              <label class="toggle-row">
                <span>{{ $t('admin.backup_email_enabled') }}</span>
                <input type="checkbox" v-model="systemSettings.backup_email_enabled" />
              </label>
              <input v-if="systemSettings.backup_email_enabled" class="form-input" v-model="systemSettings.backup_email_address" :placeholder="$t('admin.smtp_test_placeholder')" style="margin-top:4px" />
            </div>

            <div style="margin-top:12px">
              <button class="btn btn-primary btn-sm" @click="saveBackupSchedule">{{ $t('admin.backup_schedule_save') }}</button>
            </div>
          </div>

          <!-- Manual backup -->
          <div style="margin-bottom:24px">
            <h3 class="form-section-title">{{ $t('admin.backup_title') }}</h3>
            <p class="form-hint">{{ $t('admin.backup_description') }}</p>
            <button class="btn btn-primary btn-sm" @click="createBackup" :disabled="backupCreating">
              {{ backupCreating ? $t('admin.backup_creating') : $t('admin.backup_button') }}
            </button>
            <span v-if="backupSuccess" style="margin-left:12px;color:var(--color-success);font-size:13px">{{ $t('admin.backup_success', { filename: backupSuccess }) }}</span>
            <span v-if="backupError" style="margin-left:12px;color:var(--color-danger);font-size:13px">{{ $t('admin.backup_failed') }}</span>
          </div>

          <!-- Backup list -->
          <div>
            <table class="data-table" v-if="backups.length">
              <thead>
                <tr>
                  <th>{{ $t('timeTracking.date') }}</th>
                  <th>{{ $t('common.filename') }}</th>
                  <th>{{ $t('common.file_size') }}</th>
                  <th>{{ $t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="b in backups" :key="b.filename">
                  <td>{{ formatDateTime(b.created_at) }}</td>
                  <td><code>{{ b.filename }}</code></td>
                  <td>{{ formatBytes(b.size) }}</td>
                  <td>
                    <div class="actions-cell">
                      <button class="btn btn-secondary btn-sm" @click="downloadBackup(b.filename)">{{ $t('admin.backup_download') }}</button>
                      <button class="btn btn-secondary btn-sm" @click="confirmRestoreBackup(b.filename)">{{ $t('admin.backup_restore') }}</button>
                      <button class="btn btn-danger btn-sm" @click="deleteBackup(b)">{{ $t('common.delete') }}</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
            <p v-else class="form-hint">{{ $t('admin.backup_list_empty') }}</p>
          </div>
        </div>
      </div>

  </main>

  <!-- Create User Modal -->
  <BaseModal v-if="showCreateUser" :title="$t('admin.new_user')" @close="showCreateUser = false">
    <div class="form-row">
      <div class="form-group">
        <label class="form-label" for="new-user-first">{{ $t('settings.first_name') }}</label>
        <input id="new-user-first" class="form-input" v-model="newUser.first_name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="new-user-last">{{ $t('settings.last_name') }}</label>
        <input id="new-user-last" class="form-input" v-model="newUser.last_name" />
      </div>
    </div>
    <div class="form-group">
      <label class="form-label" for="new-user-username">{{ $t('auth.username') }} *</label>
      <input id="new-user-username" class="form-input" v-model="newUser.username" required />
    </div>
    <div class="form-group">
      <label class="form-label" for="new-user-email">{{ $t('auth.email') }} *</label>
      <input id="new-user-email" class="form-input" v-model="newUser.email" type="email" required />
    </div>
    <div class="form-group">
      <label class="form-label" for="new-user-password">{{ $t('auth.password') }} *</label>
      <input id="new-user-password" class="form-input" v-model="newUser.password" type="password" required minlength="8" />
    </div>
    <div class="form-group">
      <label class="form-label" for="new-user-role">{{ $t('admin.global_role') }}</label>
      <select id="new-user-role" class="form-input" v-model="newUser.global_role">
        <option value="user">{{ $t('admin.role_user') }}</option>
        <option value="admin">{{ $t('admin.role_admin') }}</option>
        <option value="viewer">{{ $t('admin.role_viewer') }}</option>
        <option value="metrics">{{ $t('admin.role_metrics') }}</option>
        <option value="backup">{{ $t('admin.role_backup') }}</option>
      </select>
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('admin.assign_projects') }}</label>
      <div class="labels-picker">
        <span
          v-for="p in projects"
          :key="p.id"
          class="label-chip project-chip"
          :class="{ active: userProjectIds.includes(p.id) }"
          :style="{ borderColor: p.color || '#6366f1', color: userProjectIds.includes(p.id) ? '#fff' : (p.color || '#6366f1'), background: userProjectIds.includes(p.id) ? (p.color || '#6366f1') : 'transparent' }"
          @click="toggleUserProject(p.id)"
        >{{ p.name }}</span>
        <span v-if="!projects.length" class="form-hint">No projects yet</span>
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('admin.assign_customers') }} <span class="form-label-hint">({{ $t('admin.assign_customers_hint') }})</span></label>
      <div class="labels-picker">
        <span
          v-for="cu in allCustomers"
          :key="cu.id"
          class="label-chip customer-chip-wrap"
          :class="{ active: userCustomerIds.includes(cu.id) }"
          :style="{ borderColor: '#0ea5e9', color: userCustomerIds.includes(cu.id) ? '#fff' : '#0ea5e9', background: userCustomerIds.includes(cu.id) ? '#0ea5e9' : 'transparent' }"
          @click="toggleUserCustomer(cu.id)"
        >
          {{ cu.name }}
          <span
            v-if="userCustomerIds.includes(cu.id)"
            class="admin-toggle"
            :class="{ 'is-admin': userCustomerAdminIds.includes(cu.id) }"
            :title="userCustomerAdminIds.includes(cu.id) ? $t('admin.role_admin') : $t('admin.role_user')"
            @click.stop="toggleCustomerAdmin(cu.id, $event)"
          >{{ userCustomerAdminIds.includes(cu.id) ? 'A' : 'M' }}</span>
        </span>
        <span v-if="!allCustomers.length" class="form-hint">No customers yet</span>
      </div>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showCreateUser = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="submitCreateUser">{{ $t('common.create') }}</button>
    </template>
  </BaseModal>

  <!-- Edit User Modal -->
    <BaseModal v-if="editUser" :title="$t('admin.edit_user')" :resizable="true" @close="editUser = null">
      <div style="display: flex; gap: 16px; margin-bottom: 16px; align-items: flex-start;">
        <div class="member-avatar-wrap" style="width: 64px; height: 64px;">
          <img v-if="getUserAvatar(editUser)" :src="getUserAvatar(editUser)" style="width: 64px; height: 64px; border-radius: 50%; object-fit: cover; border: 1px solid var(--color-border); flex-shrink: 0;" alt="" />
          <span v-else class="member-avatar-initials" style="width: 64px; height: 64px; font-size: 20px;" :style="getAvatarColor(editUser)">{{ getInitials(editUser) }}</span>
        </div>
        <div style="flex: 1;">
          <div class="form-row" style="margin-bottom: 0;">
            <div class="form-group">
              <label class="form-label" for="edit-user-first">{{ $t('settings.first_name') }}</label>
              <input id="edit-user-first" class="form-input" v-model="editUser.first_name" />
            </div>
            <div class="form-group">
              <label class="form-label" for="edit-user-last">{{ $t('settings.last_name') }}</label>
              <input id="edit-user-last" class="form-input" v-model="editUser.last_name" />
            </div>
          </div>
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-user-role">{{ $t('admin.global_role') }}</label>
        <select id="edit-user-role" class="form-input" v-model="editUser.global_role">
          <option value="user">{{ $t('admin.role_user') }}</option>
          <option value="admin">{{ $t('admin.role_admin') }}</option>
          <option value="metrics">{{ $t('admin.role_metrics') }}</option>
          <option value="backup">{{ $t('admin.role_backup') }}</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-user-display">{{ $t('settings.display_name') }}</label>
        <input id="edit-user-display" class="form-input" v-model="editUser.display_name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-user-email">{{ $t('auth.email') }}</label>
        <input id="edit-user-email" class="form-input" v-model="editUser.email" type="email" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-user-avatar">{{ $t('settings.avatar_url') }}</label>
        <input id="edit-user-avatar" class="form-input" v-model="editUser.avatar_url" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-user-locale">{{ $t('common.language') }}</label>
        <select id="edit-user-locale" class="form-input" v-model="editUser.locale">
          <option value="en">English</option>
          <option value="nl">Nederlands</option>
          <option value="de">Deutsch</option>
          <option value="fr">Français</option>
          <option value="es">Español</option>
          <option value="da">Dansk</option>
          <option value="sv">Svenska</option>
          <option value="nb">Norsk</option>
          <option value="fi">Suomi</option>
          <option value="is">Íslenska</option>
          <option value="pt">Português</option>
          <option value="it">Italiano</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-user-password">{{ $t('auth.password') }} <span class="form-label-hint">(leave blank to keep current)</span></label>
        <input id="edit-user-password" class="form-input" v-model="editUser._newPassword" type="password" autocomplete="new-password" minlength="8" placeholder="New password…" />
      </div>
      <template v-if="editUser.global_role !== 'metrics' && editUser.global_role !== 'backup'">
        <div class="form-group">
          <label class="form-label">{{ $t('admin.timetracking_viewer') }}</label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="editUser.time_tracking_viewer" />
            {{ $t('admin.timetracking_viewer_hint') }}
          </label>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('feature.features') }}</label>
          <div style="display:flex;flex-direction:column;gap:6px">
            <label class="checkbox-label">
              <input type="checkbox" v-model="editUser.time_tracking_enabled" />
              {{ $t('feature.time_tracking') }}
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="editUser.board_enabled" />
              {{ $t('feature.board') }}
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="editUser.chat_enabled" />
              {{ $t('feature.chat') }}
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="editUser.helpdesk_enabled" />
              {{ $t('feature.helpdesk') }}
            </label>
          </div>
        </div>
      </template>
      <div class="form-group">
        <label class="form-label">{{ $t('mfa.title') }}</label>
        <div v-if="editUser.totp_enabled" style="display:flex;align-items:center;gap:12px">
          <span class="badge badge-mfa">{{ $t('mfa.enabled') }}</span>
          <button type="button" class="btn btn-danger btn-sm" @click="adminResetMFA(editUser)">{{ $t('admin.reset_mfa') }}</button>
        </div>
        <span v-else class="badge badge-inactive">{{ $t('mfa.disabled') }}</span>
      </div>
      <template v-if="editUser.global_role !== 'metrics' && editUser.global_role !== 'backup'">
        <div class="form-group">
          <label class="form-label">{{ $t('groups.title') }}</label>
          <div class="labels-picker">
            <span
              v-for="g in groups"
              :key="g.id"
              class="label-chip"
              :class="{ active: userGroupIds.includes(g.id) }"
              :style="{ borderColor: '#8b5cf6', color: userGroupIds.includes(g.id) ? '#fff' : '#8b5cf6', background: userGroupIds.includes(g.id) ? '#8b5cf6' : 'transparent' }"
              @click="userGroupIds.includes(g.id) ? userGroupIds.splice(userGroupIds.indexOf(g.id), 1) : userGroupIds.push(g.id)"
            >{{ g.name }}</span>
            <span v-if="!groups.length" class="form-hint">{{ $t('groups.no_groups') }}</span>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('admin.assign_projects') }}</label>
          <div class="labels-picker">
            <span
              v-for="p in projects"
              :key="p.id"
              class="label-chip project-chip"
              :class="{ active: userProjectIds.includes(p.id) }"
              :style="{ borderColor: p.color || '#6366f1', color: userProjectIds.includes(p.id) ? '#fff' : (p.color || '#6366f1'), background: userProjectIds.includes(p.id) ? (p.color || '#6366f1') : 'transparent' }"
              @click="toggleUserProject(p.id)"
            >{{ p.name }}</span>
            <span v-if="!projects.length" class="form-hint">No projects yet</span>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('admin.assign_customers') }} <span class="form-label-hint">({{ $t('admin.assign_customers_hint') }})</span></label>
          <div class="labels-picker">
            <span
              v-for="cu in allCustomers"
              :key="cu.id"
              class="label-chip customer-chip-wrap"
              :class="{ active: userCustomerIds.includes(cu.id) }"
              :style="{ borderColor: '#0ea5e9', color: userCustomerIds.includes(cu.id) ? '#fff' : '#0ea5e9', background: userCustomerIds.includes(cu.id) ? '#0ea5e9' : 'transparent' }"
              @click="toggleUserCustomer(cu.id)"
            >
              {{ cu.name }}
              <span
                v-if="userCustomerIds.includes(cu.id)"
                class="admin-toggle"
                :class="{ 'is-admin': userCustomerAdminIds.includes(cu.id) }"
                :title="userCustomerAdminIds.includes(cu.id) ? $t('admin.role_admin') : $t('admin.role_user')"
                @click.stop="toggleCustomerAdmin(cu.id, $event)"
              >{{ userCustomerAdminIds.includes(cu.id) ? 'A' : 'M' }}</span>
            </span>
            <span v-if="!allCustomers.length" class="form-hint">No customers yet</span>
          </div>
        </div>
      </template>
      <!-- API Keys — shown only for metrics-role users -->
      <div v-if="editUser.global_role === 'metrics'" class="form-group">
        <label class="form-label">{{ $t('admin.api_keys') }}</label>

        <!-- Shown-once key result -->
        <div v-if="newApiKeyResult" style="background:var(--color-surface-raised);border:1px solid var(--color-border);border-radius:6px;padding:10px 12px;margin-bottom:10px;font-size:13px">
          <p style="margin:0 0 6px;color:var(--color-text-muted)">{{ $t('admin.api_key_shown_once') }}</p>
          <div style="display:flex;align-items:center;gap:8px">
            <code style="flex:1;word-break:break-all;background:var(--color-surface);padding:4px 8px;border-radius:4px;font-size:12px">{{ newApiKeyResult.key }}</code>
            <button type="button" class="btn btn-secondary btn-sm" @click="copyApiKey(newApiKeyResult.key)" aria-label="{{ $t('admin.api_key_copy') }}">{{ $t('admin.api_key_copy') }}</button>
          </div>
        </div>

        <!-- Existing keys -->
        <table v-if="editUserApiKeys.length" class="data-table" style="margin-bottom:10px">
          <thead>
            <tr>
              <th>{{ $t('common.name') }}</th>
              <th>{{ $t('admin.api_key_prefix') }}</th>
              <th>{{ $t('admin.api_key_last_used') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="k in editUserApiKeys" :key="k.id">
              <td>{{ k.name }}</td>
              <td><code>{{ k.key_prefix }}…</code></td>
              <td>{{ k.last_used_at ? k.last_used_at : $t('admin.api_key_never_used') }}</td>
              <td><button type="button" class="btn btn-danger btn-sm" @click="deleteUserApiKey(k.id)">{{ $t('admin.api_key_revoke') }}</button></td>
            </tr>
          </tbody>
        </table>
        <p v-else-if="!newApiKeyResult" class="form-hint">{{ $t('admin.api_key_none') }}</p>

        <!-- Create new key -->
        <div style="display:flex;gap:8px;align-items:center">
          <input class="form-input" v-model="newApiKeyName" :placeholder="$t('admin.api_key_name_placeholder')" style="flex:1" @keyup.enter="createUserApiKey" />
          <button type="button" class="btn btn-secondary btn-sm" @click="createUserApiKey" :disabled="!newApiKeyName.trim()">{{ $t('admin.api_key_generate') }}</button>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" @click="openLoginHistory(editUser)">{{ $t('admin.login_history') }}</button>
        <button class="btn btn-secondary" @click="editUser = null">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEditUser">{{ $t('common.save') }}</button>
      </template>
  </BaseModal>

  <!-- Login History Modal -->
  <BaseModal v-if="loginHistoryUser" :title="$t('admin.login_history') + ' — ' + (loginHistoryUser.display_name || loginHistoryUser.username)" width="min(95vw, 1080px)" @close="loginHistoryUser = null">
    <div v-if="loginHistoryLoading" style="text-align:center;padding:24px">
      <div class="spinner" style="width:28px;height:28px;border-width:3px;display:inline-block"></div>
    </div>
    <p v-else-if="!loginHistory.length" style="color:var(--color-text-muted);text-align:center;padding:16px 0">{{ $t('admin.login_history_empty') }}</p>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th style="white-space:nowrap">{{ $t('timeTracking.date') }}</th>
          <th>{{ $t('admin.login_history_event') }}</th>
          <th style="min-width:120px">{{ $t('admin.login_history_actor') }}</th>
          <th style="white-space:nowrap">{{ $t('admin.login_history_ip') }}</th>
          <th style="min-width:200px">{{ $t('admin.login_history_client') }}</th>
          <th>{{ $t('admin.login_history_detail') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="entry in loginHistory" :key="entry.id">
          <td style="white-space:nowrap">{{ formatDateTime(entry.created_at) }}</td>
          <td>
            <span :class="['audit-badge', 'audit-badge--' + entry.event]">
              {{ $t('admin.event_' + entry.event, entry.event) }}
            </span>
          </td>
          <td style="color:var(--color-text-muted);font-size:12px">{{ entry.actor_id !== entry.user_id ? entry.actor_username : '' }}</td>
          <td style="white-space:nowrap"><code>{{ entry.ip }}</code></td>
          <td style="font-size:12px">{{ entry.client }}</td>
          <td style="color:var(--color-text-muted);font-size:12px">{{ entry.detail }}</td>
        </tr>
      </tbody>
    </table>
    <template #footer>
      <button class="btn btn-secondary" @click="loginHistoryUser = null">{{ $t('common.close') }}</button>
    </template>
  </BaseModal>

  <!-- Create Project Modal -->
  <BaseModal v-if="showCreateProject" :title="$t('project.new_project')" @close="showCreateProject = false; newProject.key_prefix = ''; prefixTouched = false">
    <div class="form-group">
      <label class="form-label" for="new-proj-name">{{ $t('project.project_name') }} *</label>
      <input id="new-proj-name" class="form-input" v-model="newProject.name" autofocus />
    </div>
    <div class="form-group">
      <label class="form-label" for="new-proj-prefix">{{ $t('project.key_prefix') }} *</label>
      <div style="display:flex;align-items:center;gap:8px">
        <input
          id="new-proj-prefix"
          class="form-input"
          style="width:120px;text-transform:uppercase;font-family:monospace"
          :value="newProject.key_prefix"
          maxlength="10"
          @input="e => { const v = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, ''); e.target.value = v; newProject.key_prefix = v; prefixTouched = true }"
        />
        <span style="font-size:13px;color:var(--color-text-muted)">{{ $t('project.key_prefix_hint') }} &nbsp;<code style="color:var(--color-primary)">{{ newProject.key_prefix || '???' }}-1</code></span>
      </div>
    </div>
    <div class="form-group">
      <label class="form-label" for="new-proj-desc">{{ $t('project.description') }}</label>
      <textarea id="new-proj-desc" class="form-input" v-model="newProject.description" rows="3"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="new-proj-color">{{ $t('project.color') }}</label>
      <input id="new-proj-color" type="color" class="form-input" v-model="newProject.color" style="height:40px;padding:4px;width:80px" :aria-label="$t('project.color')" />
    </div>
    <div class="form-group">
      <label class="form-label" for="new-proj-avatar">Avatar</label>
      <div style="display:flex;gap:8px;align-items:center">
        <input id="new-proj-avatar" class="form-input" v-model="newProject.avatar" placeholder="https://... or /uploads/..." />
        <button type="button" class="btn btn-secondary btn-sm" @click="$refs.newProjectAvatarInput.click()">Upload</button>
        <button v-if="newProject.avatar" type="button" class="btn btn-danger btn-sm" @click="newProject.avatar = ''">Clear</button>
      </div>
      <input ref="newProjectAvatarInput" type="file" accept="image/*" style="display:none" @change="onNewProjectAvatarSelected" />
    </div>
    <div class="form-group">
      <label class="form-label" for="new-proj-type">{{ $t('sprint.board_type') }}</label>
      <select id="new-proj-type" class="form-input" v-model="newProject.board_type" style="max-width:280px">
        <option value="kanban">{{ $t('sprint.board_type_kanban') }}</option>
        <option value="scrum">{{ $t('sprint.board_type_scrum') }}</option>
      </select>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showCreateProject = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="!newProject.name.trim() || !newProject.key_prefix.trim()" @click="submitCreateProject">{{ $t('common.create') }}</button>
    </template>
  </BaseModal>

  <!-- Edit Project Modal -->
    <BaseModal v-if="editProject" :title="$t('admin.edit_project')" @close="editProject = null">
      <div class="form-group">
        <label class="form-label" for="edit-proj-name">{{ $t('project.project_name') }}</label>
        <input id="edit-proj-name" class="form-input" v-model="editProject.name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-proj-desc">{{ $t('project.description') }}</label>
        <textarea id="edit-proj-desc" class="form-input" v-model="editProject.description" rows="3"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-proj-color">{{ $t('project.color') }}</label>
        <input id="edit-proj-color" type="color" class="form-input" v-model="editProject.color" style="height:40px;padding:4px" :aria-label="$t('project.color')" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-proj-avatar">Avatar</label>
        <div style="display:flex;gap:8px;align-items:center">
          <input id="edit-proj-avatar" class="form-input" v-model="editProject.avatar" placeholder="https://... or /uploads/..." />
          <button type="button" class="btn btn-secondary btn-sm" @click="$refs.editProjectAvatarInput.click()">Upload</button>
          <button v-if="editProject.avatar" type="button" class="btn btn-danger btn-sm" @click="editProject.avatar = ''">Clear</button>
        </div>
        <input ref="editProjectAvatarInput" type="file" accept="image/*" style="display:none" @change="onEditProjectAvatarSelected" />
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="editProject = null">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEditProject">{{ $t('common.save') }}</button>
      </template>
  </BaseModal>

  <!-- Create / Edit group modal -->
  <BaseModal v-if="showGroupForm" :title="editingGroup ? $t('groups.edit') : $t('groups.create')" @close="showGroupForm = false">
    <div class="form-group">
      <label class="form-label" for="group-form-name">{{ $t('groups.name') }}</label>
      <input id="group-form-name" class="form-input" v-model="groupForm.name" />
    </div>
    <div class="form-group">
      <label class="form-label" for="group-form-desc">{{ $t('groups.description') }}</label>
      <textarea id="group-form-desc" class="form-input" v-model="groupForm.description" rows="3"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="group-form-avatar">Avatar</label>
      <div style="display:flex;gap:8px;align-items:center">
        <input id="group-form-avatar" class="form-input" v-model="groupForm.avatar" placeholder="https://... or /uploads/..." />
        <button type="button" class="btn btn-secondary btn-sm" @click="$refs.groupAvatarInput.click()">Upload</button>
        <button v-if="groupForm.avatar" type="button" class="btn btn-danger btn-sm" @click="groupForm.avatar = ''">Clear</button>
      </div>
      <input ref="groupAvatarInput" type="file" accept="image/*" style="display:none" @change="onGroupAvatarSelected" />
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showGroupForm = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="saveGroup">{{ $t('common.save') }}</button>
    </template>
  </BaseModal>

  <!-- Group detail modal: members + access -->
  <BaseModal v-if="activeGroup" :title="activeGroup.name" @close="activeGroup = null" style="--modal-width:700px">
    <div v-if="loadingGroupDetail" class="loading-state" style="padding:40px">
      <div class="spinner" style="width:28px;height:28px;border-width:3px"></div>
    </div>
    <div v-else-if="groupDetail">
      <!-- Members -->
      <h3 style="margin:0 0 8px">{{ $t('groups.members') }}</h3>
      <div style="display:flex;gap:8px;margin-bottom:12px">
        <select class="form-input" v-model="groupAddUserId" style="flex:1">
          <option value="">— {{ $t('groups.add_member') }} —</option>
          <option v-for="u in usersNotInGroup" :key="u.id" :value="u.id">
            {{ u.display_name || u.username }} ({{ u.email }})
          </option>
        </select>
        <button class="btn btn-primary btn-sm" :disabled="!groupAddUserId" @click="addGroupMember">{{ $t('common.add') }}</button>
      </div>
      <div v-if="!groupDetail.members.length" class="empty-hint">{{ $t('groups.no_members') }}</div>
      <div v-else class="members-list" style="margin-bottom:16px">
        <div v-for="m in groupDetail.members" :key="m.user_id" class="member-row">
          <div class="member-avatar-wrap">
            <img v-if="getUserAvatar(m.user)" :src="getUserAvatar(m.user)" class="member-avatar" alt="" />
            <span v-else class="member-avatar-initials" :style="getAvatarColor(m.user)">{{ getInitials(m.user) }}</span>
          </div>
          <div class="member-info">
            <span class="member-name">{{ m.user.display_name || m.user.username }}</span>
            <span class="member-email">{{ m.user.email }}</span>
          </div>
          <button class="icon-btn icon-danger" @click="removeGroupMember(m.user_id)" title="Remove">✕</button>
        </div>
      </div>

      <!-- Project access -->
      <h3 style="margin:0 0 8px">{{ $t('groups.project_access') }}</h3>
      <div style="display:flex;gap:8px;margin-bottom:12px">
        <select class="form-input" v-model="groupAddProjectId" style="flex:1">
          <option value="">— {{ $t('groups.add_project') }} —</option>
          <option v-for="p in allProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <select class="form-input" v-model="groupAddProjectRole" style="width:110px">
          <option value="viewer">{{ $t('project.roles.viewer') }}</option>
          <option value="member">{{ $t('project.roles.member') }}</option>
          <option value="owner">{{ $t('project.roles.owner') }}</option>
        </select>
        <button class="btn btn-primary btn-sm" :disabled="!groupAddProjectId" @click="addGroupProjectAccess">{{ $t('common.add') }}</button>
      </div>
      <div v-if="!groupDetail.project_access.length" class="empty-hint">{{ $t('groups.no_project_access') }}</div>
      <table v-else class="data-table" style="margin-bottom:16px">
        <thead><tr><th>{{ $t('project.project') }}</th><th>{{ $t('groups.role') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="pa in groupDetail.project_access" :key="pa.project_id">
            <td>{{ pa.project.name }}</td>
            <td>
              <select class="role-select" :value="pa.role" @change="updateGroupProjectRole(pa, $event.target.value)">
                <option value="viewer">{{ $t('project.roles.viewer') }}</option>
                <option value="member">{{ $t('project.roles.member') }}</option>
                <option value="owner">{{ $t('project.roles.owner') }}</option>
              </select>
            </td>
            <td><button class="icon-btn icon-danger" @click="removeGroupProjectAccess(pa)">✕</button></td>
          </tr>
        </tbody>
      </table>

      <!-- Customer access -->
      <h3 style="margin:0 0 8px">{{ $t('groups.customer_access') }}</h3>
      <div style="display:flex;gap:8px;margin-bottom:12px">
        <select class="form-input" v-model="groupAddCustomerId" style="flex:1">
          <option value="">— {{ $t('groups.add_customer') }} —</option>
          <option v-for="cu in allCustomers" :key="cu.id" :value="cu.id">{{ cu.name }}</option>
        </select>
        <select class="form-input" v-model="groupAddCustomerRole" style="width:110px">
          <option value="viewer">{{ $t('project.roles.viewer') }}</option>
          <option value="member">{{ $t('project.roles.member') }}</option>
          <option value="owner">{{ $t('project.roles.owner') }}</option>
        </select>
        <button class="btn btn-primary btn-sm" :disabled="!groupAddCustomerId" @click="addGroupCustomerAccess">{{ $t('common.add') }}</button>
      </div>
      <div v-if="!groupDetail.customer_access.length" class="empty-hint">{{ $t('groups.no_customer_access') }}</div>
      <table v-else class="data-table">
        <thead><tr><th>{{ $t('customer.title') }}</th><th>{{ $t('groups.role') }}</th><th></th></tr></thead>
        <tbody>
          <tr v-for="ca in groupDetail.customer_access" :key="ca.customer_id">
            <td>{{ ca.customer.name }}</td>
            <td>
              <select class="role-select" :value="ca.role" @change="updateGroupCustomerRole(ca, $event.target.value)">
                <option value="viewer">{{ $t('project.roles.viewer') }}</option>
                <option value="member">{{ $t('project.roles.member') }}</option>
                <option value="owner">{{ $t('project.roles.owner') }}</option>
              </select>
            </td>
            <td><button class="icon-btn icon-danger" @click="removeGroupCustomerAccess(ca)">✕</button></td>
          </tr>
        </tbody>
      </table>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="activeGroup = null">{{ $t('common.close') }}</button>
    </template>
  </BaseModal>

  <!-- Create / edit customer modal -->
  <BaseModal
    v-if="showCreateCustomer || editingCustomer"
    :title="editingCustomer ? $t('customer.edit') : $t('customer.new_customer')"
    @close="showCreateCustomer = false; editingCustomer = null"
  >
    <div class="form-group">
      <label class="form-label" for="admin-cust-name">{{ $t('customer.name') }}</label>
      <input id="admin-cust-name" class="form-input" v-model="customerForm.name" />
    </div>
    <div class="form-group">
      <label class="form-label" for="admin-cust-desc">{{ $t('customer.description') }}</label>
      <textarea id="admin-cust-desc" class="form-input" v-model="customerForm.description" rows="3"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label" for="admin-cust-logo">{{ $t('customer.logo_url') }}</label>
      <input id="admin-cust-logo" class="form-input" v-model="customerForm.logo_url" placeholder="https://..." />
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showCreateCustomer = false; editingCustomer = null">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="!customerForm.name.trim()" @click="editingCustomer ? saveEditCustomer() : submitCreateCustomer()">
        {{ editingCustomer ? $t('common.save') : $t('common.create') }}
      </button>
    </template>
  </BaseModal>

  <!-- News create/edit modal -->
  <BaseModal
    v-if="showNewsModal"
    :title="editingNews ? $t('admin.news_edit') : $t('admin.news_create')"
    @close="showNewsModal = false; editingNews = null"
  >
    <div class="form-group">
      <label class="form-label">{{ $t('admin.news_title_col') }}</label>
      <input class="form-input" v-model="newsForm.title" required autofocus />
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('admin.news_text') }}</label>
      <div class="md-editor">
        <div class="md-editor-tabs" role="tablist">
          <button
            role="tab"
            :aria-selected="newsEditorTab === 'edit'"
            aria-controls="news-md-panel-edit"
            id="news-md-tab-edit"
            :class="['md-tab', { active: newsEditorTab === 'edit' }]"
            type="button"
            @click="newsEditorTab = 'edit'"
          >{{ $t('common.edit') }}</button>
          <button
            role="tab"
            :aria-selected="newsEditorTab === 'preview'"
            aria-controls="news-md-panel-preview"
            id="news-md-tab-preview"
            :class="['md-tab', { active: newsEditorTab === 'preview' }]"
            type="button"
            @click="newsEditorTab = 'preview'"
          >{{ $t('common.preview') }}</button>
        </div>
        <textarea
          v-if="newsEditorTab === 'edit'"
          id="news-md-panel-edit"
          role="tabpanel"
          aria-labelledby="news-md-tab-edit"
          class="form-input md-textarea"
          v-model="newsForm.text"
          rows="8"
          :placeholder="$t('admin.news_text_placeholder')"
          required
        ></textarea>
        <div
          v-else
          id="news-md-panel-preview"
          role="tabpanel"
          aria-labelledby="news-md-tab-preview"
          class="md-preview markdown-body"
          v-html="renderMarkdown(newsForm.text)"
        ></div>
      </div>
    </div>
    <div class="form-row">
      <div class="form-group">
        <label class="form-label">{{ $t('admin.news_start_date') }}</label>
        <DateTimeInput
          v-model="newsForm.start_date"
          :label="$t('admin.news_start_date')"
          :hint="$t('admin.news_date_hint')"
        />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('admin.news_end_date') }}</label>
        <DateTimeInput
          v-model="newsForm.end_date"
          :label="$t('admin.news_end_date')"
        />
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('admin.news_sidebar_color') }}</label>
      <div class="news-color-picker">
        <button
          v-for="c in newsColorPresets"
          :key="c"
          type="button"
          class="news-color-swatch"
          :class="{ selected: newsForm.sidebar_color === c }"
          :style="{ background: c }"
          :title="c"
          :aria-label="c"
          :aria-pressed="newsForm.sidebar_color === c"
          @click="newsForm.sidebar_color = newsForm.sidebar_color === c ? '' : c"
        ></button>
        <label class="news-color-custom" :title="$t('admin.news_sidebar_color_custom')">
          <span
            class="news-color-swatch"
            :class="{ selected: newsForm.sidebar_color && !newsColorPresets.includes(newsForm.sidebar_color) }"
            :style="{ background: newsForm.sidebar_color && !newsColorPresets.includes(newsForm.sidebar_color) ? newsForm.sidebar_color : 'transparent', border: '2px dashed var(--color-border)' }"
          >+</span>
          <input type="color" class="news-color-hidden" aria-label="Custom news item color" @input="newsForm.sidebar_color = $event.target.value" />
        </label>
        <button
          v-if="newsForm.sidebar_color"
          type="button"
          class="btn btn-ghost btn-sm"
          style="padding:2px 8px;font-size:11px"
          @click="newsForm.sidebar_color = ''"
        >&#x2715;</button>
      </div>
      <div v-if="newsForm.sidebar_color" class="news-color-preview">
        <div class="news-color-preview-bar" :style="{ borderLeftColor: newsForm.sidebar_color }"></div>
        <span style="font-size:12px;color:var(--color-text-muted)">{{ newsForm.sidebar_color }}</span>
      </div>
    </div>
    <div class="form-group" style="display:flex;align-items:center;gap:10px">
      <input type="checkbox" id="news-active" v-model="newsForm.active" style="width:auto" />
      <label for="news-active" class="form-label" style="margin:0">{{ $t('admin.news_active') }}</label>
    </div>
    <div class="form-group" style="display:flex;align-items:center;gap:10px">
      <input type="checkbox" id="news-show-on-login" v-model="newsForm.show_on_login" style="width:auto" />
      <label for="news-show-on-login" class="form-label" style="margin:0">{{ $t('admin.news_show_on_login') }}</label>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showNewsModal = false; editingNews = null">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="!newsForm.title.trim() || !newsForm.text.trim() || newsSaving" @click="saveNews">
        {{ editingNews ? $t('common.save') : $t('common.create') }}
      </button>
    </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { RouterLink } from 'vue-router'
import BaseModal from '@/components/common/BaseModal.vue'
import DateTimeInput from '@/components/common/DateTimeInput.vue'
import { adminApi } from '@/api/admin'
import client, { triggerDownload } from '@/api/client'
import { groupsApi } from '@/api/groups'
import { customersApi } from '@/api/customers'
import { projectsApi } from '@/api/projects'
import { attachmentsApi } from '@/api/attachments'
import { newsApi } from '@/api/news'
import { resolveAssetUrl } from '@/api/serverConfig'
import { useUIStore } from '@/stores/ui'
import { useSystemStore } from '@/stores/system'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'
import { useDateFormat } from '@/composables/useDateFormat'
import SlaPoliciesTab from '@/components/admin/SlaPoliciesTab.vue'
import MacrosTab from '@/components/admin/MacrosTab.vue'
import TicketChecklistTemplatesTab from '@/components/admin/TicketChecklistTemplatesTab.vue'
import HelpIcon from '@/components/common/HelpIcon.vue'

const { t } = useI18n()
const ui = useUIStore()
const sidebarStore = useSidebarStore()
const systemStore = useSystemStore()
const auth = useAuthStore()
const { formatDateTime } = useDateFormat()
const tab = ref('users')

watch(tab, (activeTab) => {
  ui.setHelpContext(`admin.${activeTab}`)
}, { immediate: true })

onBeforeUnmount(() => {
  ui.setHelpContext(null)
})

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

// ── News ─────────────────────────────────────────────────────────────────────
const newsItems = ref([])
const newsLoading = ref(false)
const showNewsModal = ref(false)
const editingNews = ref(null)
const newsSaving = ref(false)
const newsForm = ref({ title: '', text: '', start_date: null, end_date: null, active: true, sidebar_color: '' })
const newsEditorTab = ref('edit')
const newsColorPresets = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#f97316', '#64748b']

async function loadNews() {
  newsLoading.value = true
  try {
    const r = await newsApi.adminList()
    newsItems.value = r.data || []
  } finally {
    newsLoading.value = false
  }
}
function openCreateNews() {
  editingNews.value = null
  newsEditorTab.value = 'edit'
  newsForm.value = { title: '', text: '', start_date: null, end_date: null, active: true, sidebar_color: '', show_on_login: false }
  showNewsModal.value = true
}
function openEditNews(item) {
  editingNews.value = item
  newsEditorTab.value = 'edit'
  newsForm.value = {
    title: item.title,
    text: item.text,
    start_date: item.start_date || null,
    end_date: item.end_date || null,
    active: item.active,
    sidebar_color: item.sidebar_color || '',
    show_on_login: !!item.show_on_login,
  }
  showNewsModal.value = true
}
async function saveNews() {
  newsSaving.value = true
  try {
    const payload = {
      title: newsForm.value.title,
      text: newsForm.value.text,
      start_date: newsForm.value.start_date,
      end_date: newsForm.value.end_date,
      active: newsForm.value.active,
      sidebar_color: newsForm.value.sidebar_color,
      show_on_login: newsForm.value.show_on_login,
    }
    if (editingNews.value) {
      await newsApi.adminUpdate(editingNews.value.id, payload)
    } else {
      await newsApi.adminCreate(payload)
    }
    showNewsModal.value = false
    editingNews.value = null
    await loadNews()
  } catch (e) {
    ui.error(e.response?.data?.error || t('common.error'))
  } finally {
    newsSaving.value = false
  }
}
async function deleteNewsItem(item) {
  if (!await ui.confirm(t('admin.news_delete_confirm', { title: item.title }))) return
  await newsApi.adminDelete(item.id)
  await loadNews()
}

const users = ref([])
const loading = ref(true)

const projects = ref([])
const loadingProjects = ref(false)
const showDeletedProjects = ref(false)
let projectsLoaded = false

const showInactiveUsers = ref(false)
const userSearch = ref('')
const projectSearch = ref('')
const groupSearch = ref('')
const customerSearch = ref('')

const editUser = ref(null)
const editUserApiKeys = ref([])
const loginHistoryUser = ref(null)
const loginHistory = ref([])
const loginHistoryLoading = ref(false)
const newApiKeyName = ref('')
const newApiKeyResult = ref(null) // { key, name } shown once after creation
const editProject = ref(null)
const showCreateUser = ref(false)
const showDeletedUsers = ref(false)
const showCreateProject = ref(false)
const newProject = ref({ name: '', description: '', color: '#6366f1', avatar: '', key_prefix: '', board_type: 'kanban' })
const prefixTouched = ref(false)

// Mirror GenerateKeyPrefix from the backend so the field auto-fills as the user types.
function autoPrefix(name) {
  const words = name.toUpperCase().split(/[^A-Z0-9]+/).filter(Boolean)
  let r = ''
  for (const w of words) { if (r.length >= 3) break; r += w[0] }
  if (r.length < 3 && words.length > 0) {
    for (let i = 1; i < words[0].length && r.length < 3; i++) r += words[0][i]
  }
  while (r.length < 3) r += 'X'
  return r.slice(0, 3)
}

function getUserAvatar(user) {
  if (!user) return null
  return resolveAssetUrl(user.avatar_url || user.gravatar_url)
}

function getInitials(user) {
  return (user.display_name || user.username || '?').slice(0, 2).toUpperCase()
}

function getAvatarColor(user) {
  const colors = ["#6366f1", "#8b5cf6", "#ec4899", "#f59e0b", "#10b981", "#3b82f6", "#ef4444"]
  const charCode = (user?.username?.charCodeAt(0) || 0)
  return { background: colors[charCode % colors.length] }
}

watch(() => newProject.value.name, (name) => {
  if (!prefixTouched.value) newProject.value.key_prefix = autoPrefix(name)
})
const newUser = ref({ username: '', email: '', password: '', first_name: '', last_name: '', global_role: 'user' })
const userProjectIds = ref([])
const userCustomerIds = ref([])
const userCustomerAdminIds = ref([])
const userGroupIds = ref([])

const allCustomers = ref([])
let customersLoaded = false

// ─── Groups ───────────────────────────────────────────────────────────────────
const groups = ref([])
const loadingGroups = ref(false)
const showGroupForm = ref(false)
const editingGroup = ref(null)
const groupForm = ref({ name: '', description: '', avatar: '' })
const activeGroup = ref(null)
const groupDetail = ref(null)
const loadingGroupDetail = ref(false)
const groupAddUserId = ref('')
const groupAddProjectId = ref('')
const groupAddProjectRole = ref('member')
const groupAddCustomerId = ref('')
const groupAddCustomerRole = ref('member')

const allProjects = ref([])
let projectsListLoaded = false

const usersNotInGroup = computed(() => {
  if (!groupDetail.value) return users.value
  const inGroup = new Set(groupDetail.value.members.map(m => m.user_id))
  return users.value.filter(u => !inGroup.has(u.id))
})

// ─── Customers ────────────────────────────────────────────────────────────────
const adminCustomers = ref([])
const loadingCustomers = ref(false)
const showCreateCustomer = ref(false)

const userSortDir = ref('asc')
const groupSortDir = ref('asc')
const customerSortDir = ref('asc')
const projectSortDir = ref('asc')

function toggleUserSort() { userSortDir.value = userSortDir.value === 'asc' ? 'desc' : 'asc' }
function toggleGroupSort() { groupSortDir.value = groupSortDir.value === 'asc' ? 'desc' : 'asc' }
function toggleCustomerSort() { customerSortDir.value = customerSortDir.value === 'asc' ? 'desc' : 'asc' }
function toggleProjectSort() { projectSortDir.value = projectSortDir.value === 'asc' ? 'desc' : 'asc' }

const sortedUsers = computed(() => {
  const q = userSearch.value.trim().toLowerCase()
  const mul = userSortDir.value === 'asc' ? 1 : -1
  let list = showInactiveUsers.value ? users.value : users.value.filter(u => u.is_active)
  if (q) list = list.filter(u =>
    (u.display_name || '').toLowerCase().includes(q) ||
    (u.username || '').toLowerCase().includes(q) ||
    (u.email || '').toLowerCase().includes(q) ||
    (u.first_name || '').toLowerCase().includes(q) ||
    (u.last_name || '').toLowerCase().includes(q))
  return [...list].sort((a, b) => {
    const an = (a.display_name || a.username || '').toLowerCase()
    const bn = (b.display_name || b.username || '').toLowerCase()
    return mul * an.localeCompare(bn)
  })
})

const sortedGroups = computed(() => {
  const q = groupSearch.value.trim().toLowerCase()
  const mul = groupSortDir.value === 'asc' ? 1 : -1
  const list = q
    ? groups.value.filter(g =>
        (g.name || '').toLowerCase().includes(q) ||
        (g.description || '').toLowerCase().includes(q))
    : groups.value
  return [...list].sort((a, b) =>
    mul * (a.name || '').toLowerCase().localeCompare((b.name || '').toLowerCase())
  )
})

const sortedCustomers = computed(() => {
  const q = customerSearch.value.trim().toLowerCase()
  const mul = customerSortDir.value === 'asc' ? 1 : -1
  const list = q
    ? adminCustomers.value.filter(c =>
        (c.name || '').toLowerCase().includes(q) ||
        (c.description || '').toLowerCase().includes(q))
    : adminCustomers.value
  return [...list].sort((a, b) =>
    mul * (a.name || '').toLowerCase().localeCompare((b.name || '').toLowerCase())
  )
})

const sortedProjects = computed(() => {
  const q = projectSearch.value.trim().toLowerCase()
  const mul = projectSortDir.value === 'asc' ? 1 : -1
  const list = q
    ? projects.value.filter(p =>
        (p.name || '').toLowerCase().includes(q) ||
        (p.slug || '').toLowerCase().includes(q) ||
        (p.key_prefix || '').toLowerCase().includes(q) ||
        (p.description || '').toLowerCase().includes(q))
    : projects.value
  return [...list].sort((a, b) =>
    mul * (a.name || '').toLowerCase().localeCompare((b.name || '').toLowerCase())
  )
})
const editingCustomer = ref(null)
const customerForm = ref({ name: '', description: '', logo_url: '' })
let adminCustomersLoaded = false

async function loadGroups() {
  loadingGroups.value = true
  try {
    const { data } = await groupsApi.list()
    groups.value = data || []
  } finally {
    loadingGroups.value = false
  }
}

async function loadAllProjects() {
  if (projectsListLoaded) return
  try {
    const { data } = await adminApi.listProjects()
    allProjects.value = data || []
    projectsListLoaded = true
  } catch {}
}

function openCreateGroup() {
  editingGroup.value = null
  groupForm.value = { name: '', description: '', avatar: '' }
  showGroupForm.value = true
}

function openEditGroup(g) {
  editingGroup.value = g
  groupForm.value = { name: g.name, description: g.description || '', avatar: g.avatar || '' }
  showGroupForm.value = true
}

async function saveGroup() {
  try {
    if (editingGroup.value) {
      await groupsApi.update(editingGroup.value.id, groupForm.value)
      ui.success(t('common.saved'))
    } else {
      await groupsApi.create(groupForm.value)
      ui.success(t('common.saved'))
    }
    showGroupForm.value = false
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function deleteGroup(g) {
  if (!await ui.confirm(t('groups.delete_confirm'))) return
  try {
    await groupsApi.delete(g.id)
    groups.value = groups.value.filter(x => x.id !== g.id)
  } catch {
    ui.error(t('common.error'))
  }
}

async function openGroupDetail(g) {
  activeGroup.value = g
  groupDetail.value = null
  groupAddUserId.value = ''
  groupAddProjectId.value = ''
  groupAddProjectRole.value = 'member'
  groupAddCustomerId.value = ''
  groupAddCustomerRole.value = 'member'
  loadingGroupDetail.value = true
  try {
    await Promise.all([loadAllCustomers(), loadAllProjects()])
    const { data } = await groupsApi.get(g.id)
    groupDetail.value = data
  } catch {
    ui.error(t('common.error'))
    activeGroup.value = null
  } finally {
    loadingGroupDetail.value = false
  }
}

async function addGroupMember() {
  if (!groupAddUserId.value) return
  try {
    await groupsApi.addMember(activeGroup.value.id, groupAddUserId.value)
    const { data } = await groupsApi.get(activeGroup.value.id)
    groupDetail.value = data
    groupAddUserId.value = ''
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function removeGroupMember(userId) {
  try {
    await groupsApi.removeMember(activeGroup.value.id, userId)
    const { data } = await groupsApi.get(activeGroup.value.id)
    groupDetail.value = data
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function addGroupProjectAccess() {
  if (!groupAddProjectId.value) return
  try {
    await groupsApi.setProjectAccess(activeGroup.value.id, groupAddProjectId.value, groupAddProjectRole.value)
    const { data } = await groupsApi.get(activeGroup.value.id)
    groupDetail.value = data
    groupAddProjectId.value = ''
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function updateGroupProjectRole(pa, role) {
  try {
    await groupsApi.setProjectAccess(activeGroup.value.id, pa.project_id, role)
    pa.role = role
  } catch {
    ui.error(t('common.error'))
  }
}

async function removeGroupProjectAccess(pa) {
  try {
    await groupsApi.removeProjectAccess(activeGroup.value.id, pa.project_id)
    const { data } = await groupsApi.get(activeGroup.value.id)
    groupDetail.value = data
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function addGroupCustomerAccess() {
  if (!groupAddCustomerId.value) return
  try {
    await groupsApi.setCustomerAccess(activeGroup.value.id, groupAddCustomerId.value, groupAddCustomerRole.value)
    const { data } = await groupsApi.get(activeGroup.value.id)
    groupDetail.value = data
    groupAddCustomerId.value = ''
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function updateGroupCustomerRole(ca, role) {
  try {
    await groupsApi.setCustomerAccess(activeGroup.value.id, ca.customer_id, role)
    ca.role = role
  } catch {
    ui.error(t('common.error'))
  }
}

async function removeGroupCustomerAccess(ca) {
  try {
    await groupsApi.removeCustomerAccess(activeGroup.value.id, ca.customer_id)
    const { data } = await groupsApi.get(activeGroup.value.id)
    groupDetail.value = data
    await loadGroups()
  } catch {
    ui.error(t('common.error'))
  }
}

async function loadAllCustomers() {
  if (customersLoaded) return
  try {
    const { data } = await customersApi.list()
    allCustomers.value = data || []
    customersLoaded = true
  } catch {}
}

function toggleUserCustomer(id) {
  const idx = userCustomerIds.value.indexOf(id)
  if (idx >= 0) {
    userCustomerIds.value.splice(idx, 1)
    // Also remove admin flag when removing access entirely
    const ai = userCustomerAdminIds.value.indexOf(id)
    if (ai >= 0) userCustomerAdminIds.value.splice(ai, 1)
  } else {
    userCustomerIds.value.push(id)
  }
}

function toggleCustomerAdmin(id, event) {
  event.stopPropagation()
  if (!userCustomerIds.value.includes(id)) return
  const idx = userCustomerAdminIds.value.indexOf(id)
  if (idx >= 0) userCustomerAdminIds.value.splice(idx, 1)
  else userCustomerAdminIds.value.push(id)
}

const systemSettings = ref({
  registration_enabled: true,
  mfa_required: false,
  mfa_remember_devices: 'week_month',
  scrum_storypoints_enabled: false,
  gravatar_enabled: true,
  external_image_proxy_enabled: true,
  session_timeout_minutes: 60,
  default_date_time_format: 'YYYY-MM-DD HH:mm',
  default_timezone: 'UTC',
  default_theme: 'system',
  default_font: 'system',
  default_font_size: '14',
  default_locale: 'en',
  smtp_host: '',
  smtp_port: '587',
  smtp_from: '',
  smtp_username: '',
  smtp_password: '',
  imap_enabled: false,
  imap_host: '',
  imap_port: '993',
  imap_username: '',
  imap_password: '',
  imap_use_tls: true,
  imap_mailbox: 'INBOX',
  imap_poll_interval: '60',
  imap_auth_mechanism: 'plain',
  imap_oauth2_provider: '',
  company_name: '',
  company_logo: '',
  company_logo_dark: '',
  default_columns: 'Backlog',
  default_labels: 'Bug\nFeature\nDesign\nContent',
  password_min_length: 8,
  password_require_upper: false,
  password_require_lower: false,
  password_require_digit: false,
  password_require_special: false,
  backup_schedule: 'disabled',
  backup_start_time: '',
  backup_keep: 10,
  backup_last_run: '',
  backup_email_enabled: false,
  backup_email_address: '',
  metrics_last_access: '',
  metrics_last_access_success: '',
})
const mfaRememberDevicesPrevious = ref('week_month')
// True when the server has a password saved (so we show a placeholder instead of the value)
const smtpPasswordSet = ref(false)
const smtpPasswordPlaceholder = computed(() => smtpPasswordSet.value ? '••••••••' : '')
const imapPasswordSet = ref(false)
const imapPasswordPlaceholder = computed(() => imapPasswordSet.value ? '••••••••' : '')
const imapTesting = ref(false)
const imapPolling = ref(false)
const imapOAuth2Connected = ref(false)
const imapOAuth2Connecting = ref(false)
const smtpTestEmail = ref('')
const smtpTestSending = ref(false)
let settingsLoading = false

const timezones = [
  'UTC',
  'Europe/Amsterdam', 'Europe/Berlin', 'Europe/Brussels', 'Europe/London',
  'Europe/Madrid', 'Europe/Paris', 'Europe/Rome', 'Europe/Stockholm',
  'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
  'America/Toronto', 'America/Vancouver', 'America/Sao_Paulo',
  'Asia/Dubai', 'Asia/Istanbul', 'Asia/Jerusalem', 'Asia/Kolkata',
  'Asia/Singapore', 'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Seoul',
  'Australia/Sydney', 'Pacific/Auckland'
]

async function loadUsers() {
  loading.value = true
  try {
    const { data } = await adminApi.listUsers(showDeletedUsers.value)
    users.value = data
  } finally {
    loading.value = false
  }
}

onMounted(loadUsers)

async function loadProjects() {
  loadingProjects.value = true
  projectsLoaded = false
  try {
    const { data } = await adminApi.listProjects(showDeletedProjects.value)
    projects.value = data
    projectsLoaded = true
  } finally {
    loadingProjects.value = false
  }
}

async function restoreProject(project) {
  try {
    await adminApi.restoreProject(project.id)
    projects.value = projects.value.filter(p => p.id !== project.id)
    ui.success(`Project "${project.name}" restored`)
  } catch {
    ui.error('Failed to restore project')
  }
}

async function purgeProject(project) {
  if (!await ui.confirm(t('admin.purge_project_confirm', { name: project.name }))) return
  try {
    await adminApi.purgeProject(project.id)
    projects.value = projects.value.filter(p => p.id !== project.id)
    ui.success(t('admin.purge_project_success', { name: project.name }))
  } catch {
    ui.error('Failed to permanently delete project')
  }
}

function toggleUserProject(id) {
  const idx = userProjectIds.value.indexOf(id)
  if (idx >= 0) userProjectIds.value.splice(idx, 1)
  else userProjectIds.value.push(id)
}

async function loadSettings() {
  if (settingsLoading) return
  settingsLoading = true
  try {
    const { data } = await adminApi.getSystemSettings()
    systemSettings.value.registration_enabled        = data.registration_enabled !== 'false'
    systemSettings.value.mfa_required                = data.mfa_required === 'true' || data.mfa_required === true
    systemSettings.value.mfa_remember_devices        = data.mfa_remember_devices || 'week_month'
    mfaRememberDevicesPrevious.value                 = systemSettings.value.mfa_remember_devices
    systemSettings.value.scrum_storypoints_enabled   = data.scrum_storypoints_enabled === 'true' || data.scrum_storypoints_enabled === true
    systemSettings.value.gravatar_enabled             = data.gravatar_enabled !== 'false' && data.gravatar_enabled !== false
    systemSettings.value.external_image_proxy_enabled = data.external_image_proxy_enabled !== 'false' && data.external_image_proxy_enabled !== false
    systemSettings.value.session_timeout_minutes  = parseInt(data.session_timeout_minutes) || 0
    systemSettings.value.default_date_time_format = data.default_date_time_format || 'YYYY-MM-DD HH:mm'
    systemSettings.value.default_timezone         = data.default_timezone || 'UTC'
    systemSettings.value.default_theme            = data.default_theme || 'system'
    systemSettings.value.default_font             = data.default_font || 'system'
    systemSettings.value.default_font_size        = data.default_font_size || '14'
    systemSettings.value.default_locale           = data.default_locale || 'en'
    systemSettings.value.smtp_host                = data.smtp_host || ''
    systemSettings.value.smtp_port                = data.smtp_port || '587'
    systemSettings.value.smtp_from                = data.smtp_from || ''
    systemSettings.value.smtp_username            = data.smtp_username || ''
    // Password is never sent back from the server — show placeholder if one is set
    smtpPasswordSet.value = data.smtp_password_set === 'true'
    systemSettings.value.smtp_password            = ''
    systemSettings.value.imap_enabled       = data.imap_enabled === 'true'
    systemSettings.value.imap_host          = data.imap_host || ''
    systemSettings.value.imap_port          = data.imap_port || '993'
    systemSettings.value.imap_username      = data.imap_username || ''
    imapPasswordSet.value = data.imap_password_set === 'true'
    systemSettings.value.imap_password      = ''
    systemSettings.value.imap_use_tls       = data.imap_use_tls !== 'false'
    systemSettings.value.imap_mailbox       = data.imap_mailbox || 'INBOX'
    systemSettings.value.imap_poll_interval = data.imap_poll_interval || '60'
    systemSettings.value.imap_auth_mechanism  = data.imap_auth_mechanism || 'plain'
    systemSettings.value.imap_oauth2_provider = data.imap_oauth2_provider || ''
    imapOAuth2Connected.value = data.imap_auth_mechanism === 'oauth2' && data.imap_access_token_set === 'true'
    systemSettings.value.company_name             = data.company_name || ''
    systemSettings.value.company_logo             = data.company_logo || ''
    systemSettings.value.company_logo_dark        = data.company_logo_dark || ''
    systemSettings.value.login_branding_enabled   = data.login_branding_enabled === 'true'
    systemSettings.value.default_columns          = data.default_columns || 'Backlog'
    systemSettings.value.default_labels           = data.default_labels || 'Bug\nFeature\nDesign\nContent'
    systemSettings.value.password_min_length      = parseInt(data.password_min_length) || 8
    systemSettings.value.password_require_upper   = data.password_require_upper === 'true'
    systemSettings.value.password_require_lower   = data.password_require_lower === 'true'
    systemSettings.value.password_require_digit   = data.password_require_digit === 'true'
    systemSettings.value.password_require_special = data.password_require_special === 'true'
    systemSettings.value.backup_schedule           = data.backup_schedule || 'disabled'
    systemSettings.value.backup_start_time         = data.backup_start_time || ''
    systemSettings.value.backup_keep               = parseInt(data.backup_keep) || 10
    systemSettings.value.backup_last_run           = data.backup_last_run || ''
    systemSettings.value.backup_email_enabled       = data.backup_email_enabled === 'true'
    systemSettings.value.backup_email_address       = data.backup_email_address || ''
    systemSettings.value.allowed_ips                = data.allowed_ips || ''
    systemSettings.value.metrics_last_access        = data.metrics_last_access || ''
    systemSettings.value.metrics_last_access_success = data.metrics_last_access_success || ''
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to load settings')
  } finally {
    settingsLoading = false
  }
}

async function onLogoFileSelected(e) {
  const file = e.target.files[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    systemSettings.value.company_logo = data.url
  } catch {
    ui.error('Failed to upload image')
  }
}

async function clearCompanyLogo() {
  systemSettings.value.company_logo = ''
  await saveBrandingSettings()
}

async function onLogoDarkFileSelected(e) {
  const file = e.target.files[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    systemSettings.value.company_logo_dark = data.url
  } catch {
    ui.error('Failed to upload image')
  }
}

async function clearCompanyLogoDark() {
  systemSettings.value.company_logo_dark = ''
  await saveBrandingSettings()
}

async function saveBrandingSettings() {
  try {
    await adminApi.updateSystemSettings({
      company_name: systemSettings.value.company_name,
      company_logo: systemSettings.value.company_logo,
      company_logo_dark: systemSettings.value.company_logo_dark,
      login_branding_enabled: systemSettings.value.login_branding_enabled,
    })
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

const backingUp = ref(false)
const backups = ref([])
const restoringBackup = ref(null)
const backupStartTimeDisplay = ref('')

const prefers12HourTime = computed(() => {
  const fmt = auth.user?.date_time_format || systemStore.defaults.date_time_format || 'YYYY-MM-DD HH:mm'
  return fmt.includes('hh') && fmt.includes('a')
})

const backupTimePlaceholder = computed(() => (prefers12HourTime.value ? 'hh:mm am' : 'HH:mm'))

function formatBackupStartTime(raw) {
  if (!raw) return ''
  const [hhStr, mmStr] = raw.split(':')
  const hh = Number(hhStr)
  const mm = Number(mmStr)
  if (!Number.isInteger(hh) || !Number.isInteger(mm) || hh < 0 || hh > 23 || mm < 0 || mm > 59) return ''
  if (!prefers12HourTime.value) return `${String(hh).padStart(2, '0')}:${String(mm).padStart(2, '0')}`
  const suffix = hh < 12 ? 'am' : 'pm'
  const h12 = hh % 12 || 12
  return `${String(h12).padStart(2, '0')}:${String(mm).padStart(2, '0')} ${suffix}`
}

function parseBackupStartTime(input) {
  const val = (input || '').trim()
  if (!val) return ''

  if (prefers12HourTime.value) {
    const m = val.match(/^(\d{1,2}):(\d{2})\s*([ap]m)$/i)
    if (!m) return null
    let hh = Number(m[1])
    const mm = Number(m[2])
    const mer = m[3].toLowerCase()
    if (!Number.isInteger(hh) || !Number.isInteger(mm) || hh < 1 || hh > 12 || mm < 0 || mm > 59) return null
    if (mer === 'am') hh = hh === 12 ? 0 : hh
    if (mer === 'pm') hh = hh === 12 ? 12 : hh + 12
    return `${String(hh).padStart(2, '0')}:${String(mm).padStart(2, '0')}`
  }

  const m = val.match(/^(\d{1,2}):(\d{2})$/)
  if (!m) return null
  const hh = Number(m[1])
  const mm = Number(m[2])
  if (!Number.isInteger(hh) || !Number.isInteger(mm) || hh < 0 || hh > 23 || mm < 0 || mm > 59) return null
  return `${String(hh).padStart(2, '0')}:${String(mm).padStart(2, '0')}`
}

function onBackupStartTimeBlur() {
  const parsed = parseBackupStartTime(backupStartTimeDisplay.value)
  if (parsed === null) {
    backupStartTimeDisplay.value = formatBackupStartTime(systemSettings.value.backup_start_time)
    return
  }
  systemSettings.value.backup_start_time = parsed
  backupStartTimeDisplay.value = formatBackupStartTime(parsed)
}

watch([() => systemSettings.value.backup_start_time, prefers12HourTime], () => {
  backupStartTimeDisplay.value = formatBackupStartTime(systemSettings.value.backup_start_time)
}, { immediate: true })

const backupLastRun = computed(() => {
  const v = systemSettings.value.backup_last_run
  return v ? formatDateTime(v) : '–'
})

const backupNextRun = computed(() => {
  const hours = { '6h': 6, '8h': 8, '12h': 12, '24h': 24 }[systemSettings.value.backup_schedule]
  if (!hours) return '–'
  const startTime = systemSettings.value.backup_start_time
  if (startTime) {
    const [hh, mm] = startTime.split(':').map(Number)
    const now = new Date()
    const midnight = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0)
    let anchor = new Date(midnight.getTime() + hh * 3600000 + mm * 60000)
    if (anchor > now) anchor = new Date(anchor.getTime() - 24 * 3600000)
    const intervalMs = hours * 3600000
    const n = Math.floor((now - anchor) / intervalMs)
    const nextSlot = new Date(anchor.getTime() + (n + 1) * intervalMs)
    return formatDateTime(nextSlot.toISOString())
  }
  const lastRun = systemSettings.value.backup_last_run
  if (!lastRun) return '–'
  return formatDateTime(new Date(new Date(lastRun).getTime() + hours * 3600000).toISOString())
})

async function saveBackupSchedule() {
  try {
    await adminApi.updateSystemSettings({
      backup_schedule: systemSettings.value.backup_schedule,
      backup_start_time: systemSettings.value.backup_start_time,
      backup_keep: systemSettings.value.backup_keep,
      backup_email_enabled: systemSettings.value.backup_email_enabled,
      backup_email_address: systemSettings.value.backup_email_address,
    })
    ui.success('Backup schedule saved')
  } catch {
    ui.error('Failed to save backup schedule')
  }
}

async function loadBackups() {
  try {
    const { data } = await adminApi.listBackups()
    backups.value = data
  } catch {
    // silently ignore — backups dir may not exist yet
  }
}

async function createBackup() {
  backingUp.value = true
  try {
    const { data } = await adminApi.backupDatabase()
    ui.success(`Backup created: ${data.filename}`)
    await loadBackups()
  } catch {
    ui.error('Backup failed')
  } finally {
    backingUp.value = false
  }
}

async function restoreBackup(b) {
  if (!await ui.confirm(`Replace the current database with "${b.filename}"? All changes since this backup will be lost.`)) return
  restoringBackup.value = b.filename
  try {
    await adminApi.restoreBackup(b.filename)
    ui.success(`Database restored from ${b.filename}`)
  } catch {
    ui.error('Restore failed')
  } finally {
    restoringBackup.value = null
  }
}

async function downloadBackup(b) {
  try {
    const data = await adminApi.downloadBackup(b.filename)
    await triggerDownload(data, b.filename, 'application/octet-stream')
  } catch {
    ui.error('Failed to download backup')
  }
}

async function deleteBackup(b) {
  if (!await ui.confirm(`Delete backup "${b.filename}"?`)) return
  try {
    await adminApi.deleteBackup(b.filename)
    backups.value = backups.value.filter(x => x.filename !== b.filename)
    ui.success('Backup deleted')
  } catch {
    ui.error('Failed to delete backup')
  }
}

function formatBytes(bytes) {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' kB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

async function savePasswordPolicy() {
  try {
    await adminApi.updateSystemSettings({
      password_min_length:           systemSettings.value.password_min_length,
      password_require_upper:        systemSettings.value.password_require_upper,
      password_require_lower:        systemSettings.value.password_require_lower,
      password_require_digit:        systemSettings.value.password_require_digit,
      password_require_special:      systemSettings.value.password_require_special,
      password_change_period_days:   systemSettings.value.password_change_period_days ?? 0,
    })
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function saveGeneralSettings() {
  try {
    const payload = {
      registration_enabled:          systemSettings.value.registration_enabled,
      scrum_storypoints_enabled:     systemSettings.value.scrum_storypoints_enabled,
      gravatar_enabled:              systemSettings.value.gravatar_enabled,
      external_image_proxy_enabled:  systemSettings.value.external_image_proxy_enabled,
      session_timeout_minutes:       systemSettings.value.session_timeout_minutes,
      default_date_time_format:      systemSettings.value.default_date_time_format,
      default_timezone:              systemSettings.value.default_timezone,
      default_theme:                 systemSettings.value.default_theme,
      default_font:                  systemSettings.value.default_font,
      default_font_size:             systemSettings.value.default_font_size,
      default_locale:                systemSettings.value.default_locale,
      default_columns:               systemSettings.value.default_columns,
      default_labels:                systemSettings.value.default_labels,
    }
    await adminApi.updateSystemSettings(payload)
    await systemStore.fetchSettings()
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function onMfaRememberDevicesChange() {
  const newVal = systemSettings.value.mfa_remember_devices
  if (newVal === 'disabled' && !confirm($t('mfa.remember_devices_disable_confirm'))) {
    systemSettings.value.mfa_remember_devices = mfaRememberDevicesPrevious.value
    return
  }
  await saveMFASettings()
  mfaRememberDevicesPrevious.value = systemSettings.value.mfa_remember_devices
}

async function saveMFASettings() {
  try {
    await adminApi.updateSystemSettings({
      mfa_required: systemSettings.value.mfa_required,
      mfa_remember_devices: systemSettings.value.mfa_remember_devices,
    })
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function saveSecuritySettings() {
  try {
    await adminApi.updateSystemSettings({ allowed_ips: systemSettings.value.allowed_ips })
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function saveSmtpSettings() {
  try {
    const payload = {
      smtp_host:     systemSettings.value.smtp_host,
      smtp_port:     String(systemSettings.value.smtp_port || '587'),
      smtp_from:     systemSettings.value.smtp_from,
      smtp_username: systemSettings.value.smtp_username,
    }
    // Only include password if the admin typed something new
    if (systemSettings.value.smtp_password) {
      payload.smtp_password = systemSettings.value.smtp_password
    }
    await adminApi.updateSystemSettings(payload)
    if (systemSettings.value.smtp_password) {
      smtpPasswordSet.value = true
      systemSettings.value.smtp_password = ''
    }
    ui.success('Settings saved')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save settings')
  }
}

async function saveImapSettings() {
  try {
    const payload = {
      imap_enabled:         systemSettings.value.imap_enabled,
      imap_host:            systemSettings.value.imap_host,
      imap_port:            String(systemSettings.value.imap_port || '993'),
      imap_username:        systemSettings.value.imap_username,
      imap_use_tls:         systemSettings.value.imap_use_tls,
      imap_mailbox:         systemSettings.value.imap_mailbox,
      imap_poll_interval:   String(systemSettings.value.imap_poll_interval || '60'),
      imap_auth_mechanism:  systemSettings.value.imap_auth_mechanism,
      imap_oauth2_provider: systemSettings.value.imap_oauth2_provider,
    }
    if (systemSettings.value.imap_password) {
      payload.imap_password = systemSettings.value.imap_password
    }
    await adminApi.updateSystemSettings(payload)
    if (systemSettings.value.imap_password) {
      imapPasswordSet.value = true
      systemSettings.value.imap_password = ''
    }
    ui.success('Settings saved')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to save settings')
  }
}

async function testImap() {
  imapTesting.value = true
  try {
    const body = {
      host:            systemSettings.value.imap_host,
      port:            Number(systemSettings.value.imap_port) || 993,
      username:        systemSettings.value.imap_username,
      use_tls:         systemSettings.value.imap_use_tls,
      mailbox:         systemSettings.value.imap_mailbox,
      auth_mechanism:  systemSettings.value.imap_auth_mechanism || 'plain',
    }
    if (systemSettings.value.imap_password) body.password = systemSettings.value.imap_password
    await client.post('/admin/imap/test', body)
    ui.success(t('admin.imap_test_success'))
  } catch (e) {
    ui.error(e.response?.data?.error || t('admin.imap_test_failed'))
  } finally {
    imapTesting.value = false
  }
}

async function pollImap() {
  imapPolling.value = true
  try {
    await client.post('/admin/imap/poll')
    ui.success(t('admin.imap_poll_success'))
  } catch (e) {
    ui.error(e.response?.data?.error || t('admin.imap_poll_failed'))
  } finally {
    imapPolling.value = false
  }
}

async function authorizeImapOAuth2() {
  const provider = systemSettings.value.imap_oauth2_provider
  if (!provider) {
    ui.error('Select an email provider first')
    return
  }
  imapOAuth2Connecting.value = true
  try {
    const { data } = await client.get('/admin/imap/oauth2/auth-url', { params: { provider } })
    // Open the OAuth2 authorization in a popup
    const popup = window.open(data.url, 'oauth2-auth', 'width=600,height=700')
    if (!popup) {
      ui.error('Popup was blocked. Allow popups for this site and try again.')
      return
    }
    // Poll until the popup closes (callback closes it on success)
    const checkPopup = setInterval(async () => {
      if (popup.closed) {
        clearInterval(checkPopup)
        imapOAuth2Connecting.value = false
        // Refresh status
        try {
          const statusRes = await client.get('/admin/imap/oauth2/status')
          imapOAuth2Connected.value = statusRes.data.connected
          if (statusRes.data.connected) {
            systemSettings.value.imap_auth_mechanism = 'oauth2'
            systemSettings.value.imap_oauth2_provider = statusRes.data.provider
            ui.success('Connected')
          }
        } catch {
          // ignore
        }
      }
    }, 500)
  } catch (e) {
    imapOAuth2Connecting.value = false
    ui.error(e.response?.data?.error || 'Authorization failed')
  }
}

async function disconnectImapOAuth2() {
  try {
    await client.post('/admin/imap/oauth2/disconnect')
    imapOAuth2Connected.value = false
    systemSettings.value.imap_auth_mechanism = 'plain'
    systemSettings.value.imap_oauth2_provider = ''
    ui.success('Disconnected')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to disconnect')
  }
}

async function sendSmtpTest() {
  smtpTestSending.value = true
  try {
    await adminApi.sendTestEmail(smtpTestEmail.value)
    ui.success('Test email sent to ' + smtpTestEmail.value)
    smtpTestEmail.value = ''
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to send test email')
  } finally {
    smtpTestSending.value = false
  }
}

function openCreateUser() {
  userProjectIds.value = []
  userCustomerIds.value = []
  userCustomerAdminIds.value = []
  showCreateUser.value = true
  loadProjects()
  loadAllCustomers()
}

async function submitCreateUser() {
  try {
    const { data } = await adminApi.createUser(newUser.value)
    if (userProjectIds.value.length) {
      await adminApi.setUserProjects(data.id, userProjectIds.value)
    }
    const customerRoles = {}
    userCustomerIds.value.forEach(id => { customerRoles[id] = userCustomerAdminIds.value.includes(id) ? 'admin' : 'member' })
    await adminApi.setUserCustomers(data.id, userCustomerIds.value, customerRoles)
    users.value.push(data)
    showCreateUser.value = false
    newUser.value = { username: '', email: '', password: '', first_name: '', last_name: '', global_role: 'user' }
    userProjectIds.value = []
    userCustomerIds.value = []
    userCustomerAdminIds.value = []
    sidebarStore.fetchAllUsers()
    ui.success('User created')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create user')
  }
}

async function setRole(user, newRole) {
  await adminApi.updateUser(user.id, { global_role: newRole })
  user.global_role = newRole
}

async function toggleActive(user) {
  await adminApi.updateUser(user.id, { is_active: !user.is_active })
  user.is_active = !user.is_active
}

const featureLabels = {
  board_enabled: 'feature.board',
  chat_enabled: 'feature.chat',
  time_tracking_enabled: 'feature.time_tracking',
  time_tracking_viewer: 'admin.timetracking_viewer',
  helpdesk_enabled: 'feature.helpdesk',
}

async function toggleFeature(user, field, value) {
  const idx = users.value.findIndex(u => u.id === user.id)
  if (idx === -1) return
  try {
    await adminApi.updateUser(user.id, { [field]: value })
    users.value[idx] = { ...users.value[idx], [field]: value }
    const label = t(featureLabels[field] || field)
    ui.success(`${label} ${value ? 'enabled' : 'disabled'} for ${user.display_name || user.username}`)
  } catch (e) {
    ui.error('Failed to update feature access')
  }
}

async function deleteUser(user) {
  if (!await ui.confirm(`Delete user ${user.username}?`)) return
  try {
    await adminApi.deleteUser(user.id)
    users.value = users.value.filter(u => u.id !== user.id)
    sidebarStore.fetchAllUsers()
    ui.success('User deleted')
  } catch {
    ui.error('Failed to delete user')
  }
}

async function restoreUser(user) {
  try {
    await adminApi.restoreUser(user.id)
    users.value = users.value.filter(u => u.id !== user.id)
    ui.success(`User ${user.username} restored`)
  } catch {
    ui.error('Failed to restore user')
  }
}

async function purgeUser(user) {
  if (!await ui.confirm(t('admin.purge_user_confirm', { name: user.username }))) return
  try {
    await adminApi.purgeUser(user.id)
    users.value = users.value.filter(u => u.id !== user.id)
    ui.success(`User ${user.username} permanently deleted`)
  } catch {
    ui.error('Failed to permanently delete user')
  }
}

async function adminResetMFA(user) {
  if (!await ui.confirm(`Disable MFA for ${user.display_name || user.username}?`)) return
  try {
    const { data } = await adminApi.disableUserMFA(user.id)
    // Update both the modal and the list
    editUser.value = { ...editUser.value, totp_enabled: false }
    const idx = users.value.findIndex(u => u.id === user.id)
    if (idx >= 0) users.value[idx] = { ...users.value[idx], totp_enabled: false }
    ui.success('MFA disabled for ' + (data.display_name || data.username))
  } catch {
    ui.error('Failed to disable MFA')
  }
}

async function openLoginHistory(user) {
  loginHistoryUser.value = user
  loginHistory.value = []
  loginHistoryLoading.value = true
  try {
    const { data } = await adminApi.getUserLoginHistory(user.id)
    loginHistory.value = data || []
  } catch {
    ui.error(t('common.error'))
  } finally {
    loginHistoryLoading.value = false
  }
}

async function openEditUser(user) {
  editUser.value = { ...user, _newPassword: '' }
  editUserApiKeys.value = []
  newApiKeyName.value = ''
  newApiKeyResult.value = null
  userProjectIds.value = []
  userCustomerIds.value = []
  userCustomerAdminIds.value = []
  userGroupIds.value = []
  loadProjects()
  loadAllCustomers()
  loadGroups()
  const calls = [
    adminApi.getUserProjects(user.id),
    adminApi.getUserCustomers(user.id),
    adminApi.getUserGroups(user.id),
  ]
  if (user.global_role === 'metrics') {
    calls.push(adminApi.listUserApiKeys(user.id))
  }
  try {
    const [projRes, custRes, grpRes, keysRes] = await Promise.all(calls)
    userProjectIds.value = projRes.data.project_ids || []
    userCustomerIds.value = custRes.data.customer_ids || []
    const roles = custRes.data.customer_roles || {}
    userCustomerAdminIds.value = Object.entries(roles)
      .filter(([, role]) => role === 'admin')
      .map(([id]) => Number(id))
    userGroupIds.value = grpRes.data.group_ids || []
    if (keysRes) editUserApiKeys.value = keysRes.data || []
  } catch {}
}

async function createUserApiKey() {
  if (!newApiKeyName.value.trim()) return
  try {
    const { data } = await adminApi.createUserApiKey(editUser.value.id, newApiKeyName.value.trim())
    newApiKeyResult.value = { key: data.key, name: data.name, id: data.id, key_prefix: data.key_prefix }
    editUserApiKeys.value.push({ id: data.id, name: data.name, key_prefix: data.key_prefix, created_at: data.created_at })
    newApiKeyName.value = ''
  } catch {
    ui.error('Failed to create API key')
  }
}

async function deleteUserApiKey(keyId) {
  if (!await ui.confirm(t('admin.api_key_revoke_confirm'))) return
  try {
    await adminApi.deleteUserApiKey(editUser.value.id, keyId)
    editUserApiKeys.value = editUserApiKeys.value.filter(k => k.id !== keyId)
    if (newApiKeyResult.value?.id === keyId) newApiKeyResult.value = null
  } catch {
    ui.error('Failed to revoke API key')
  }
}

function copyApiKey(key) {
  navigator.clipboard.writeText(key)
  ui.success(t('admin.api_key_copied'))
}

async function saveEditUser() {
  try {
    const isServiceAccount = editUser.value.global_role === 'metrics' || editUser.value.global_role === 'backup'
    const payload = {
      global_role: editUser.value.global_role,
      first_name: editUser.value.first_name,
      last_name: editUser.value.last_name,
      display_name: editUser.value.display_name,
      email: editUser.value.email,
      avatar_url: editUser.value.avatar_url,
      locale: editUser.value.locale,
      time_tracking_viewer: isServiceAccount ? false : (editUser.value.time_tracking_viewer ?? false),
      time_tracking_enabled: isServiceAccount ? false : (editUser.value.time_tracking_enabled ?? false),
      board_enabled: isServiceAccount ? false : (editUser.value.board_enabled ?? true),
      chat_enabled: isServiceAccount ? false : (editUser.value.chat_enabled ?? true),
      helpdesk_enabled: isServiceAccount ? false : (editUser.value.helpdesk_enabled ?? false)
    }
    if (editUser.value._newPassword) {
      payload.password = editUser.value._newPassword
    }
    const { data } = await adminApi.updateUser(editUser.value.id, payload)
    await adminApi.setUserProjects(editUser.value.id, userProjectIds.value)
    const customerRoles = {}
    userCustomerIds.value.forEach(id => { customerRoles[id] = userCustomerAdminIds.value.includes(id) ? 'admin' : 'member' })
    await adminApi.setUserCustomers(editUser.value.id, userCustomerIds.value, customerRoles)
    await adminApi.setUserGroups(editUser.value.id, userGroupIds.value)
    const idx = users.value.findIndex(u => u.id === data.id)
    if (idx >= 0) users.value[idx] = data
    editUser.value = null
    userProjectIds.value = []
    userCustomerIds.value = []
    userCustomerAdminIds.value = []
    userGroupIds.value = []
    ui.success('User updated')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update user')
  }
}

async function submitCreateProject() {
  try {
    const { data } = await adminApi.createProject(newProject.value)
    projects.value.unshift(data)
    showCreateProject.value = false
    newProject.value = { name: '', description: '', color: '#6366f1', avatar: '', key_prefix: '', board_type: 'kanban' }
    prefixTouched.value = false
    ui.success('Project created')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create project')
  }
}

function openEditProject(project) {
  editProject.value = { ...project }
}

async function saveEditProject() {
  try {
    const { data } = await adminApi.updateProject(editProject.value.id, {
      name: editProject.value.name,
      description: editProject.value.description,
      color: editProject.value.color,
      avatar: editProject.value.avatar ?? ''
    })
    const idx = projects.value.findIndex(p => p.id === data.id)
    if (idx >= 0) projects.value[idx] = data
    editProject.value = null
    ui.success('Project updated')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update project')
  }
}

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

function groupAvatar(group) {
  return resolveAssetUrl(group?.avatar || '')
}

async function onNewProjectAvatarSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    newProject.value.avatar = data.url
  } catch {
    ui.error('Failed to upload avatar')
  }
}

async function onEditProjectAvatarSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    editProject.value.avatar = data.url
  } catch {
    ui.error('Failed to upload avatar')
  }
}

async function onGroupAvatarSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    groupForm.value.avatar = data.url
  } catch {
    ui.error('Failed to upload avatar')
  }
}

async function toggleArchive(project) {
  try {
    const { data } = await adminApi.updateProject(project.id, { is_archived: !project.is_archived })
    const idx = projects.value.findIndex(p => p.id === project.id)
    if (idx >= 0) projects.value[idx] = data
  } catch {
    ui.error('Failed to update project')
  }
}

async function deleteProject(project) {
  if (!await ui.confirm(`Delete project "${project.name}"?`)) return
  try {
    await adminApi.deleteProject(project.id)
    projects.value = projects.value.filter(p => p.id !== project.id)
    sidebarStore.allProjects = sidebarStore.allProjects.filter(p => p.id !== project.id)
    ui.success('Project deleted')
  } catch {
    ui.error('Failed to delete project')
  }
}

async function loadAdminCustomers() {
  if (adminCustomersLoaded) return
  loadingCustomers.value = true
  try {
    const { data } = await customersApi.list()
    adminCustomers.value = data || []
    adminCustomersLoaded = true
  } catch {
    ui.error('Failed to load customers')
  } finally {
    loadingCustomers.value = false
  }
}

async function submitCreateCustomer() {
  try {
    const { data } = await customersApi.create({
      name: customerForm.value.name.trim(),
      description: customerForm.value.description,
      logo_url: customerForm.value.logo_url,
    })
    adminCustomers.value.push(data)
    showCreateCustomer.value = false
    customerForm.value = { name: '', description: '', logo_url: '' }
    ui.success('Customer created')
  } catch (e) {
    ui.error(e?.response?.data?.error || 'Failed to create customer')
  }
}

function openEditCustomer(c) {
  editingCustomer.value = { ...c }
  customerForm.value = { name: c.name, description: c.description || '', logo_url: c.logo_url || '' }
}

async function saveEditCustomer() {
  try {
    const { data } = await customersApi.update(editingCustomer.value.id, customerForm.value)
    const idx = adminCustomers.value.findIndex(c => c.id === editingCustomer.value.id)
    if (idx >= 0) adminCustomers.value[idx] = data
    editingCustomer.value = null
    customerForm.value = { name: '', description: '', logo_url: '' }
    ui.success('Customer updated')
  } catch {
    ui.error('Failed to update customer')
  }
}

async function deleteAdminCustomer(c) {
  if (!await ui.confirm(`Delete customer "${c.name}"? All linked projects will be detached.`)) return
  try {
    await customersApi.delete(c.id)
    adminCustomers.value = adminCustomers.value.filter(ac => ac.id !== c.id)
    ui.success('Customer deleted')
  } catch {
    ui.error('Failed to delete customer')
  }
}

// ── Time-tracking projects & customers ──────────────────────────────────────
const adminTTProjects  = ref([])
const adminTTCustomers = ref([])
const loadingTTProjects  = ref(false)
const loadingTTCustomers = ref(false)
const addingTTProject    = ref(false)
const addingTTCustomer   = ref(false)
const newTTProject       = ref({ name: '', color: '#6366f1', undeclStr: '' })
const newTTCustomer      = ref({ name: '' })
const editingTTProject   = ref(null)
const editingTTCustomer  = ref(null)
const newTTProjRef       = ref(null)
const newTTCustRef       = ref(null)

function fmtTTTime(minutes) {
  if (!minutes) return '0:00'
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return `${h}:${String(m).padStart(2, '0')}`
}

function parseTTTime(val) {
  if (!val && val !== 0) return 0
  const s = String(val)
  if (s.includes(':')) {
    const [h, m] = s.split(':')
    return (parseInt(h) || 0) * 60 + (parseInt(m) || 0)
  }
  return Math.round((parseFloat(s) || 0) * 60)
}

function isGlobalTTEntity(e) {
  return e.created_by_id !== auth.user?.id
}

async function loadAdminTTProjects() {
  loadingTTProjects.value = true
  try {
    const { data } = await projectsApi.listTimeTracking()
    adminTTProjects.value = data
  } catch {
    ui.error(t('timeTracking.tt_project_load_error'))
  } finally {
    loadingTTProjects.value = false
  }
}

async function loadAdminTTCustomers() {
  loadingTTCustomers.value = true
  try {
    const { data } = await customersApi.listTimeTracking()
    adminTTCustomers.value = data
  } catch {
    ui.error(t('timeTracking.tt_customer_load_error'))
  } finally {
    loadingTTCustomers.value = false
  }
}

async function confirmAddTTProject() {
  const name = newTTProject.value.name.trim()
  if (!name) return
  try {
    const { data } = await projectsApi.createTimeTracking({
      name,
      color: newTTProject.value.color,
      undeclarable_minutes: parseTTTime(newTTProject.value.undeclStr),
    })
    adminTTProjects.value.push(data)
    newTTProject.value = { name: '', color: '#6366f1', undeclStr: '' }
    addingTTProject.value = false
  } catch {
    ui.error(t('timeTracking.tt_project_save_error'))
  }
}

async function confirmAddTTCustomer() {
  const name = newTTCustomer.value.name.trim()
  if (!name) return
  try {
    const { data } = await customersApi.createTimeTracking({ name })
    adminTTCustomers.value.push(data)
    newTTCustomer.value = { name: '' }
    addingTTCustomer.value = false
  } catch {
    ui.error(t('timeTracking.tt_customer_save_error'))
  }
}

function startEditTTProject(p) {
  const undeclStr = p.undeclarable_minutes > 0 ? fmtTTTime(p.undeclarable_minutes) : ''
  editingTTProject.value = { id: p.id, name: p.name, color: p.color || '#6366f1', undeclStr }
}

async function saveEditTTProject() {
  const e = editingTTProject.value
  if (!e || !e.name.trim()) return
  try {
    const { data } = await projectsApi.updateTimeTracking(e.id, {
      name: e.name.trim(),
      color: e.color,
      undeclarable_minutes: parseTTTime(e.undeclStr),
    })
    const idx = adminTTProjects.value.findIndex(p => p.id === e.id)
    if (idx >= 0) adminTTProjects.value[idx] = data
    editingTTProject.value = null
  } catch {
    ui.error(t('timeTracking.tt_project_save_error'))
  }
}

async function deleteTTProject(p) {
  if (!await ui.confirm(t('timeTracking.tt_project_delete_confirm'))) return
  try {
    await projectsApi.deleteTimeTracking(p.id)
    adminTTProjects.value = adminTTProjects.value.filter(x => x.id !== p.id)
  } catch {
    ui.error(t('timeTracking.tt_project_delete_error'))
  }
}

function startEditTTCustomer(c) {
  editingTTCustomer.value = { id: c.id, name: c.name }
}

async function saveEditTTCustomer() {
  const e = editingTTCustomer.value
  if (!e || !e.name.trim()) return
  try {
    const { data } = await customersApi.updateTimeTracking(e.id, { name: e.name.trim() })
    const idx = adminTTCustomers.value.findIndex(c => c.id === e.id)
    if (idx >= 0) adminTTCustomers.value[idx] = data
    editingTTCustomer.value = null
  } catch {
    ui.error(t('timeTracking.tt_customer_save_error'))
  }
}

async function deleteTTCustomer(c) {
  if (!await ui.confirm(t('timeTracking.tt_customer_delete_confirm'))) return
  try {
    await customersApi.deleteTimeTracking(c.id)
    adminTTCustomers.value = adminTTCustomers.value.filter(x => x.id !== c.id)
  } catch {
    ui.error(t('timeTracking.tt_customer_delete_error'))
  }
}
</script>

<style scoped>
.admin-main { flex: 1; padding: 32px 24px; }
.admin-container { max-width: 1100px; margin: 0 auto; }
h1 { font-size: 22px; font-weight: 700; margin-bottom: 24px; }

.tabs { display: flex; gap: 4px; margin-bottom: 24px; border-bottom: 1px solid var(--color-border); }
.tab {
  padding: 10px 20px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-muted);
  margin-bottom: -1px;
}
.tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }

.data-table { width: 100%; border-collapse: collapse; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
.data-table th, .data-table td { padding: 12px 16px; text-align: left; border-bottom: 1px solid var(--color-border); font-size: 13px; vertical-align: middle; }
.data-table th { font-weight: 600; color: var(--color-text-muted); font-size: 12px; background: var(--color-bg); }
.sortable-th { cursor: pointer; user-select: none; white-space: nowrap; }
.sortable-th:hover { color: var(--color-text); }
.sort-indicator { margin-left: 4px; opacity: 0.6; font-size: 11px; }
.data-table small { color: var(--color-text-muted); font-size: 11px; }
.email { color: var(--color-text-muted); }
.actions-cell { display: flex; gap: 6px; flex-wrap: wrap; }

.badge-admin { background: #ede9fe; color: #5b21b6; }
.badge-user { background: #f1f5f9; color: #64748b; }

.audit-badge { display: inline-block; padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 600; white-space: nowrap; }
.audit-badge--login_ok { background: #dcfce7; color: #166534; }
.audit-badge--login_failed { background: #fee2e2; color: #991b1b; }
.audit-badge--logout { background: #f1f5f9; color: #64748b; }
.audit-badge--password_changed { background: #fef9c3; color: #854d0e; }
.audit-badge--password_reset { background: #fef9c3; color: #854d0e; }
.audit-badge--mfa_ok { background: #dbeafe; color: #1e40af; }
.audit-badge--mfa_failed { background: #fee2e2; color: #991b1b; }
.audit-badge--mfa_challenge { background: #ede9fe; color: #5b21b6; }
.audit-badge--mfa_trusted_device { background: #f0fdf4; color: #166534; }
.audit-badge--mfa_disabled { background: #fef9c3; color: #854d0e; }
.audit-badge--email_changed { background: #fef9c3; color: #854d0e; }
.audit-badge--passkey_registered { background: #dbeafe; color: #1e40af; }
.audit-badge--passkey_deleted { background: #fee2e2; color: #991b1b; }
.audit-badge--api_key_created { background: #dbeafe; color: #1e40af; }
.audit-badge--api_key_deleted { background: #fee2e2; color: #991b1b; }
.audit-badge--admin_user_created { background: #dcfce7; color: #166534; }
.audit-badge--admin_user_updated { background: #fef9c3; color: #854d0e; }
.audit-badge--admin_user_deleted { background: #fee2e2; color: #991b1b; }
.audit-badge--admin_user_restored { background: #dbeafe; color: #1e40af; }
.audit-badge--admin_user_purged { background: #fee2e2; color: #991b1b; }
.audit-badge--admin_mfa_disabled { background: #fef9c3; color: #854d0e; }

.role-select {
  font-size: 12px;
  padding: 3px 6px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
  cursor: pointer;
}
.badge-active { background: #dcfce7; color: #166534; }
.badge-inactive { background: #fee2e2; color: #991b1b; }
.badge-deleted { background: #f1f5f9; color: #64748b; }
.badge-mfa { background: #dbeafe; color: #1e40af; margin-left: 6px; }
[data-theme="dark"] .badge-active { background: #14532d; color: #86efac; }
[data-theme="dark"] .badge-inactive { background: #450a0a; color: #fca5a5; }
[data-theme="dark"] .badge-mfa { background: #1e3a5f; color: #93c5fd; }
.open-cards-count { font-weight: 600; color: var(--color-primary); }
.entity-cell { display: flex; align-items: center; gap: 8px; }
.entity-avatar {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}
.project-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.loading-state { display: flex; justify-content: center; padding: 60px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.tab-toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.admin-search { width: 260px; margin-left: auto; }
.settings-section { max-width: 560px; }
.settings-section h2 { font-size: 16px; font-weight: 600; margin-bottom: 16px; }
.settings-section h3 { font-size: 14px; font-weight: 600; margin-bottom: 8px; }
.mfa-settings-group {
  max-width: 400px;
  margin-bottom: 20px;
  padding: 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--color-surface) 92%, var(--color-primary) 8%);
}
.mfa-settings-heading { font-size: 14px; font-weight: 600; margin: 0 0 6px; }
.mfa-settings-hint { margin: 0 0 14px; }
.mfa-settings-group .form-group:last-child { margin-bottom: 0; }
.settings-subsection { font-size: 14px; font-weight: 600; margin-top: 28px; margin-bottom: 4px; color: var(--color-text); }
.toggle-row { display: flex; align-items: center; justify-content: space-between; font-size: 14px; font-weight: 500; cursor: pointer; }
.toggle-row input[type=checkbox] { width: 18px; height: 18px; cursor: pointer; }
.form-hint { font-size: 12px; color: var(--color-text-muted); margin-top: 4px; }
.form-label-hint { font-size: 11px; color: var(--color-text-muted); font-weight: 400; }

.labels-picker { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 6px; }
.label-chip {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: 1.5px solid currentColor;
  transition: background 0.15s, color 0.15s;
  user-select: none;
}
.label-chip:hover { opacity: 0.85; }

.customer-chip-wrap { display: inline-flex; align-items: center; gap: 4px; }

.admin-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  font-size: 9px;
  font-weight: 700;
  background: rgba(255,255,255,0.3);
  color: inherit;
  cursor: pointer;
  flex-shrink: 0;
  line-height: 1;
}
.admin-toggle.is-admin { background: rgba(0,0,0,0.25); }
.admin-toggle:hover { background: rgba(0,0,0,0.2); }

.members-list { display: flex; flex-direction: column; gap: 4px; }
.member-row { display: flex; align-items: center; gap: 10px; padding: 6px 8px; border-radius: var(--radius-sm); background: var(--color-bg); }
.member-avatar { width: 28px; height: 28px; border-radius: 50%; object-fit: cover; flex-shrink: 0; }
.member-avatar-wrap { width: 28px; height: 28px; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.member-avatar-initials { width: 100%; height: 100%; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 10px; font-weight: 600; }
.member-info { display: flex; flex-direction: column; flex: 1; min-width: 0; }
.member-name { font-size: 13px; font-weight: 500; color: var(--color-text); }
.member-email { font-size: 11px; color: var(--color-text-muted); }

.customer-avatar-container { display: flex; align-items: center; justify-content: center; }
.customer-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  object-fit: contain;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
}
.customer-avatar-placeholder {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: var(--color-text-muted);
  font-size: 14px;
}

/* ── Markdown editor ── */
.md-editor { border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
.md-editor-tabs { display: flex; background: var(--color-bg-alt); border-bottom: 1px solid var(--color-border); }
.md-tab {
  background: none; border: none; padding: 6px 16px; font-size: 13px; cursor: pointer;
  color: var(--color-text-muted); border-bottom: 2px solid transparent; margin-bottom: -1px;
}
.md-tab:hover { color: var(--color-text); }
.md-tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); font-weight: 600; }
.md-textarea { border: none !important; border-radius: 0 !important; resize: vertical; min-height: 160px; }
.md-preview {
  min-height: 160px; padding: 12px 14px;
  font-size: 14px; line-height: 1.6; color: var(--color-text);
  background: var(--color-surface);
}

/* markdown-body styles are global in main.css */

/* ── News sidebar color picker ── */
.news-color-picker {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}
.news-color-swatch {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text-muted);
  transition: transform .1s, border-color .1s;
  flex-shrink: 0;
}
.news-color-swatch:hover { transform: scale(1.15); }
.news-color-swatch.selected { border-color: var(--color-text); transform: scale(1.1); }
.news-color-custom {
  cursor: pointer;
  position: relative;
}
.news-color-hidden {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  pointer-events: none;
}
.news-color-custom:hover .news-color-swatch { transform: scale(1.15); }
.news-color-preview {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}
.tt-admin-add-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.tt-admin-color {
  width: 36px;
  height: 30px;
  padding: 2px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  cursor: pointer;
  background: none;
}
.tt-admin-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}
.tt-admin-undecl-badge {
  font-size: 11px;
  color: var(--color-danger, #ef4444);
  background: color-mix(in srgb, var(--color-danger, #ef4444) 10%, transparent);
  padding: 1px 6px;
  border-radius: 4px;
  white-space: nowrap;
}
.ttp-badge {
  font-size: 10px;
  color: var(--color-text-muted);
  background: var(--color-surface-alt);
  padding: 1px 5px;
  border-radius: 3px;
  margin-left: 6px;
  white-space: nowrap;
  vertical-align: middle;
}
.tab-section-title {
  font-size: 15px;
  font-weight: 600;
  margin: 0;
}
.news-color-preview-bar {
  width: 48px;
  height: 20px;
  border-left: 4px solid var(--color-primary);
  border-radius: 2px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-left-width: 4px;
}

.feature-matrix { margin-top: 32px; }
.feature-matrix h3 { font-size: 16px; margin: 0 0 12px; }
.feat-check { font-weight: 700; font-size: 16px; }
.feat-always { color: var(--color-text-muted); }
.feat-on { color: var(--color-primary); }
.feat-off { color: var(--color-text-muted); }
.feat-toggle { width: 18px; height: 18px; cursor: pointer; }
.name-link { background: none; border: none; padding: 0; font: inherit; font-weight: 600; color: var(--color-primary); cursor: pointer; text-align: left; }
.name-link:hover { text-decoration: underline; }
.name-link:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; border-radius: 2px; }
</style>
