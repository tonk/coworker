<template>
  <main class="admin-main">
      <div class="admin-container">
        <h1>{{ $t('admin.panel') }}</h1>

        <div class="tabs">
          <button :class="['tab', { active: tab === 'users' }]" @click="tab = 'users'">{{ $t('admin.users') }}</button>
          <button :class="['tab', { active: tab === 'projects' }]" @click="tab = 'projects'; loadProjects()">{{ $t('admin.projects') }}</button>
          <button :class="['tab', { active: tab === 'settings' }]" @click="tab = 'settings'; loadSettings()">{{ $t('admin.settings') }}</button>
        </div>

        <!-- Users tab -->
        <div v-if="tab === 'users'">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="openCreateUser">+ {{ $t('admin.create_user') }}</button>
          </div>

          <div v-if="loading" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>

          <table v-else class="data-table">
            <thead>
              <tr>
                <th>{{ $t('admin.user') }}</th>
                <th>{{ $t('admin.global_role') }}</th>
                <th>{{ $t('admin.last_login') }}</th>
                <th>Status</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>
                  <strong>{{ user.display_name || user.username }}</strong>
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
                  </select>
                </td>
                <td>
                  <small>{{ user.last_login_at ? formatDateTime(user.last_login_at) : '-' }}</small>
                </td>
                <td>
                  <span :class="['badge', user.is_active ? 'badge-active' : 'badge-inactive']">
                    {{ user.is_active ? $t('admin.active') : $t('admin.inactive') }}
                  </span>
                  <span v-if="user.totp_enabled" class="badge badge-mfa" :title="$t('mfa.title')">MFA</span>
                </td>
                <td class="actions-cell">
                  <button class="btn btn-secondary btn-sm" @click="openEditUser(user)">{{ $t('common.edit') }}</button>
                  <button class="btn btn-secondary btn-sm" @click="toggleActive(user)">
                    {{ user.is_active ? $t('admin.deactivate') : $t('admin.activate') }}
                  </button>
                  <button class="btn btn-danger btn-sm" @click="deleteUser(user)">{{ $t('common.delete') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Projects tab -->
        <div v-if="tab === 'projects'">
          <div class="tab-toolbar">
            <button class="btn btn-primary btn-sm" @click="showCreateProject = true">+ {{ $t('project.new_project') }}</button>
          </div>
          <div v-if="loadingProjects" class="loading-state">
            <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
          </div>

          <table v-else class="data-table">
            <thead>
              <tr>
                <th>{{ $t('project.project_name') }}</th>
                <th>{{ $t('admin.owner') }}</th>
                <th>Status</th>
                <th>{{ $t('admin.open_cards') }}</th>
                <th>{{ $t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="project in projects" :key="project.id">
                <td>
                  <strong>{{ project.name }}</strong>
                  <br><small>{{ project.slug }}</small>
                  <br><small>{{ project.description }}</small>
                </td>
                <td>
                  <small>{{ project.created_by?.display_name || project.created_by?.username }}</small>
                </td>
                <td>
                  <span :class="['badge', project.is_archived ? 'badge-inactive' : 'badge-active']">
                    {{ project.is_archived ? $t('admin.archived') : $t('admin.active') }}
                  </span>
                </td>
                <td>
                  <span class="open-cards-count">{{ project.open_card_count }}</span>
                </td>
                <td class="actions-cell">
                  <button class="btn btn-secondary btn-sm" @click="openEditProject(project)">{{ $t('common.edit') }}</button>
                  <button class="btn btn-secondary btn-sm" @click="toggleArchive(project)">
                    {{ project.is_archived ? $t('admin.unarchive') : $t('project.archive') }}
                  </button>
                  <button class="btn btn-danger btn-sm" @click="deleteProject(project)">{{ $t('common.delete') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Settings tab -->
        <div v-if="tab === 'settings'">
          <div class="settings-section">
            <h2>{{ $t('admin.system_settings') }}</h2>

            <div class="form-group" style="max-width:400px">
              <label class="toggle-row">
                <span>{{ $t('admin.registration_enabled') }}</span>
                <input type="checkbox" v-model="systemSettings.registration_enabled" @change="saveGeneralSettings" />
              </label>
              <p class="form-hint">{{ $t('admin.registration_hint') }}</p>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="toggle-row">
                <span>{{ $t('mfa.enforce_label') }}</span>
                <input type="checkbox" v-model="systemSettings.mfa_required" @change="saveMFASettings" />
              </label>
              <p class="form-hint">{{ $t('mfa.enforce_hint') }}</p>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('admin.session_timeout') }}</label>
              <div class="form-row" style="align-items:center;gap:8px;max-width:240px">
                <input class="form-input" type="number" min="0" v-model.number="systemSettings.session_timeout_minutes" @change="saveGeneralSettings" style="width:100px" />
                <span class="form-hint" style="margin:0">{{ $t('admin.session_timeout_unit') }}</span>
              </div>
              <p class="form-hint">{{ $t('admin.session_timeout_hint') }}</p>
            </div>

            <h3 class="settings-subsection">{{ $t('admin.global_defaults_title') }}</h3>
            <p class="form-hint" style="margin-bottom:16px">{{ $t('admin.global_defaults_hint') }}</p>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('settings.date_time_format') }}</label>
              <select class="form-input" v-model="systemSettings.default_date_time_format" @change="saveGeneralSettings">
                <option value="YYYY-MM-DD HH:mm">YYYY-MM-DD HH:mm (ISO)</option>
                <option value="DD/MM/YYYY HH:mm">DD/MM/YYYY HH:mm</option>
                <option value="MM/DD/YYYY hh:mm a">MM/DD/YYYY hh:mm a</option>
                <option value="DD-MM-YYYY HH:mm">DD-MM-YYYY HH:mm</option>
                <option value="DD.MM.YYYY HH:mm">DD.MM.YYYY HH:mm</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('settings.timezone') }}</label>
              <select class="form-input" v-model="systemSettings.default_timezone" @change="saveGeneralSettings">
                <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('settings.theme') }}</label>
              <select class="form-input" v-model="systemSettings.default_theme" @change="saveGeneralSettings">
                <option value="light">{{ $t('settings.theme_light') }}</option>
                <option value="dark">{{ $t('settings.theme_dark') }}</option>
                <option value="system">{{ $t('settings.theme_system') }}</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('settings.font') }}</label>
              <select class="form-input" v-model="systemSettings.default_font" @change="saveGeneralSettings">
                <option value="system">{{ $t('settings.font_system') }}</option>
                <option value="Inter, sans-serif">Inter</option>
                <option value="'Roboto', sans-serif">Roboto</option>
                <option value="'Open Sans', sans-serif">Open Sans</option>
                <option value="'Source Code Pro', monospace">Source Code Pro (monospace)</option>
                <option value="Georgia, serif">Georgia (serif)</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('settings.font_size') }}</label>
              <select class="form-input" v-model="systemSettings.default_font_size" @change="saveGeneralSettings">
                <option value="12">12px</option>
                <option value="13">13px</option>
                <option value="14">14px</option>
                <option value="15">15px</option>
                <option value="16">16px</option>
                <option value="18">18px</option>
              </select>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('common.language') }}</label>
              <select class="form-input" v-model="systemSettings.default_locale" @change="saveGeneralSettings">
                <option value="en">English</option>
                <option value="nl">Nederlands</option>
                <option value="de">Deutsch</option>
                <option value="fr">Français</option>
                <option value="es">Español</option>
              </select>
            </div>

            <h3 class="settings-subsection">{{ $t('admin.project_defaults_title') }}</h3>
            <p class="form-hint" style="margin-bottom:16px">{{ $t('admin.default_columns_hint') }}</p>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('admin.default_columns') }}</label>
              <textarea class="form-input" v-model="systemSettings.default_columns" rows="4" style="font-family:monospace;resize:vertical" :placeholder="'Backlog\nIn Progress\nDone'"></textarea>
              <p class="form-hint">{{ $t('admin.default_columns_each_line') }}</p>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('admin.default_labels') }}</label>
              <textarea class="form-input" v-model="systemSettings.default_labels" rows="4" style="font-family:monospace;resize:vertical" :placeholder="'Bug\nFeature\nDesign\nContent'"></textarea>
              <p class="form-hint">{{ $t('admin.default_labels_each_line') }}</p>
            </div>

            <div style="max-width:400px;margin-top:8px">
              <button class="btn btn-primary btn-sm" @click="saveGeneralSettings">{{ $t('common.save') }}</button>
            </div>

            <h3 class="settings-subsection">{{ $t('admin.smtp_title') }}</h3>
            <p class="form-hint" style="margin-bottom:16px">{{ $t('admin.smtp_hint') }}</p>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:3">
                <label class="form-label">{{ $t('admin.smtp_host') }}</label>
                <input class="form-input" v-model="systemSettings.smtp_host" :placeholder="$t('admin.smtp_host_placeholder')" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label">{{ $t('admin.smtp_port') }}</label>
                <input class="form-input" v-model="systemSettings.smtp_port" type="number" placeholder="587" />
              </div>
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('admin.smtp_from') }}</label>
              <input class="form-input" v-model="systemSettings.smtp_from" type="email" placeholder="noreply@example.com" />
            </div>

            <div class="form-row" style="max-width:500px">
              <div class="form-group" style="flex:1">
                <label class="form-label">{{ $t('admin.smtp_username') }}</label>
                <input class="form-input" v-model="systemSettings.smtp_username" autocomplete="off" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label">{{ $t('admin.smtp_password') }}</label>
                <input class="form-input" v-model="systemSettings.smtp_password" type="password" autocomplete="new-password" :placeholder="smtpPasswordPlaceholder" />
              </div>
            </div>

            <div class="form-actions" style="max-width:500px">
              <button class="btn btn-primary" @click="saveSmtpSettings">{{ $t('common.save') }}</button>
            </div>

            <div class="form-group" style="max-width:500px;margin-top:16px">
              <label class="form-label">{{ $t('admin.smtp_test_title') }}</label>
              <div style="display:flex;gap:8px">
                <input class="form-input" v-model="smtpTestEmail" type="email" :placeholder="$t('admin.smtp_test_placeholder')" style="flex:1" />
                <button class="btn btn-secondary" :disabled="smtpTestSending || !smtpTestEmail" @click="sendSmtpTest">
                  {{ smtpTestSending ? $t('admin.smtp_test_sending') : $t('admin.smtp_test_send') }}
                </button>
              </div>
            </div>

            <h3 class="settings-subsection">{{ $t('admin.branding_title') }}</h3>
            <p class="form-hint" style="margin-bottom:16px">{{ $t('admin.branding_hint') }}</p>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('admin.company_name') }}</label>
              <input class="form-input" v-model="systemSettings.company_name" :placeholder="$t('admin.company_name_placeholder')" />
            </div>

            <div class="form-group" style="max-width:400px">
              <label class="form-label">{{ $t('admin.company_logo') }}</label>
              <input class="form-input" v-model="systemSettings.company_logo" :placeholder="'https://...'" style="margin-bottom:8px" />
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

            <div class="form-actions" style="max-width:400px">
              <button class="btn btn-primary" @click="saveBrandingSettings">{{ $t('common.save') }}</button>
            </div>

            <h3 class="settings-subsection">{{ $t('admin.password_policy_title') }}</h3>
            <p class="form-hint" style="margin-bottom:16px">{{ $t('admin.password_policy_hint') }}</p>

            <div class="form-group" style="max-width:240px">
              <label class="form-label">{{ $t('admin.password_min_length') }}</label>
              <input class="form-input" type="number" min="8" max="128" v-model.number="systemSettings.password_min_length" style="width:100px" />
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

            <div style="max-width:400px;margin-top:8px">
              <button class="btn btn-primary btn-sm" @click="savePasswordPolicy">{{ $t('common.save') }}</button>
            </div>
          </div>
        </div>
      </div>
  </main>

  <!-- Create User Modal -->
  <BaseModal v-if="showCreateUser" :title="$t('admin.create_user')" @close="showCreateUser = false">
    <div class="form-row">
      <div class="form-group">
        <label class="form-label">{{ $t('settings.first_name') }}</label>
        <input class="form-input" v-model="newUser.first_name" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('settings.last_name') }}</label>
        <input class="form-input" v-model="newUser.last_name" />
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('auth.username') }} *</label>
      <input class="form-input" v-model="newUser.username" required />
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('auth.email') }} *</label>
      <input class="form-input" v-model="newUser.email" type="email" required />
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('auth.password') }} *</label>
      <input class="form-input" v-model="newUser.password" type="password" required minlength="8" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('admin.global_role') }}</label>
      <select class="form-input" v-model="newUser.global_role">
        <option value="user">{{ $t('admin.role_user') }}</option>
        <option value="admin">{{ $t('admin.role_admin') }}</option>
        <option value="viewer">{{ $t('admin.role_viewer') }}</option>
        <option value="metrics">{{ $t('admin.role_metrics') }}</option>
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
    <template #footer>
      <button class="btn btn-secondary" @click="showCreateUser = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="submitCreateUser">{{ $t('common.create') }}</button>
    </template>
  </BaseModal>

  <!-- Edit User Modal -->
    <BaseModal v-if="editUser" :title="$t('admin.edit_user')" @close="editUser = null">
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">{{ $t('settings.first_name') }}</label>
          <input class="form-input" v-model="editUser.first_name" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ $t('settings.last_name') }}</label>
          <input class="form-input" v-model="editUser.last_name" />
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('settings.display_name') }}</label>
        <input class="form-input" v-model="editUser.display_name" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('auth.email') }}</label>
        <input class="form-input" v-model="editUser.email" type="email" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('settings.avatar_url') }}</label>
        <input class="form-input" v-model="editUser.avatar_url" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('common.language') }}</label>
        <select class="form-input" v-model="editUser.locale">
          <option value="en">English</option>
          <option value="nl">Nederlands</option>
          <option value="de">Deutsch</option>
          <option value="fr">Français</option>
          <option value="es">Español</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('auth.password') }} <span class="form-label-hint">(leave blank to keep current)</span></label>
        <input class="form-input" v-model="editUser._newPassword" type="password" autocomplete="new-password" minlength="8" placeholder="New password…" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('mfa.title') }}</label>
        <div v-if="editUser.totp_enabled" style="display:flex;align-items:center;gap:12px">
          <span class="badge badge-mfa">{{ $t('mfa.enabled') }}</span>
          <button type="button" class="btn btn-danger btn-sm" @click="adminResetMFA(editUser)">{{ $t('admin.reset_mfa') }}</button>
        </div>
        <span v-else class="badge badge-inactive">{{ $t('mfa.disabled') }}</span>
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
      <template #footer>
        <button class="btn btn-secondary" @click="editUser = null">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEditUser">{{ $t('common.save') }}</button>
      </template>
  </BaseModal>

  <!-- Create Project Modal -->
  <BaseModal v-if="showCreateProject" :title="$t('project.new_project')" @close="showCreateProject = false; newProject.key_prefix = ''; prefixTouched = false">
    <div class="form-group">
      <label class="form-label">{{ $t('project.project_name') }} *</label>
      <input class="form-input" v-model="newProject.name" autofocus />
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('project.key_prefix') }} *</label>
      <div style="display:flex;align-items:center;gap:8px">
        <input
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
      <label class="form-label">{{ $t('project.description') }}</label>
      <textarea class="form-input" v-model="newProject.description" rows="3"></textarea>
    </div>
    <div class="form-group">
      <label class="form-label">{{ $t('project.color') }}</label>
      <input type="color" class="form-input" v-model="newProject.color" style="height:40px;padding:4px;width:80px" />
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showCreateProject = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="!newProject.name.trim() || !newProject.key_prefix.trim()" @click="submitCreateProject">{{ $t('common.create') }}</button>
    </template>
  </BaseModal>

  <!-- Edit Project Modal -->
    <BaseModal v-if="editProject" :title="$t('admin.edit_project')" @close="editProject = null">
      <div class="form-group">
        <label class="form-label">{{ $t('project.project_name') }}</label>
        <input class="form-input" v-model="editProject.name" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('project.description') }}</label>
        <textarea class="form-input" v-model="editProject.description" rows="3"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label">{{ $t('project.color') }}</label>
        <input type="color" class="form-input" v-model="editProject.color" style="height:40px;padding:4px" />
      </div>
      <template #footer>
        <button class="btn btn-secondary" @click="editProject = null">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEditProject">{{ $t('common.save') }}</button>
      </template>
  </BaseModal>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import BaseModal from '@/components/common/BaseModal.vue'
import { adminApi } from '@/api/admin'
import { attachmentsApi } from '@/api/attachments'
import { resolveAssetUrl } from '@/api/serverConfig'
import { useUIStore } from '@/stores/ui'
import { useSidebarStore } from '@/stores/sidebar'
import { useDateFormat } from '@/composables/useDateFormat'

const ui = useUIStore()
const sidebarStore = useSidebarStore()
const { formatDateTime } = useDateFormat()
const tab = ref('users')

const users = ref([])
const loading = ref(true)

const projects = ref([])
const loadingProjects = ref(false)
let projectsLoaded = false

const editUser = ref(null)
const editProject = ref(null)
const showCreateUser = ref(false)
const showCreateProject = ref(false)
const newProject = ref({ name: '', description: '', color: '#6366f1', key_prefix: '' })
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

watch(() => newProject.value.name, (name) => {
  if (!prefixTouched.value) newProject.value.key_prefix = autoPrefix(name)
})
const newUser = ref({ username: '', email: '', password: '', first_name: '', last_name: '', global_role: 'user' })
const userProjectIds = ref([])

const systemSettings = ref({
  registration_enabled: true,
  mfa_required: false,
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
  company_name: '',
  company_logo: '',
  default_columns: 'Backlog',
  default_labels: 'Bug\nFeature\nDesign\nContent',
  password_min_length: 8,
  password_require_upper: false,
  password_require_lower: false,
  password_require_digit: false,
  password_require_special: false,
})
// True when the server has a password saved (so we show a placeholder instead of the value)
const smtpPasswordSet = ref(false)
const smtpPasswordPlaceholder = computed(() => smtpPasswordSet.value ? '••••••••' : '')
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

onMounted(async () => {
  try {
    const { data } = await adminApi.listUsers()
    users.value = data
  } finally {
    loading.value = false
  }
})

async function loadProjects() {
  if (projectsLoaded) return
  loadingProjects.value = true
  try {
    const { data } = await adminApi.listProjects()
    projects.value = data
    projectsLoaded = true
  } finally {
    loadingProjects.value = false
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
    systemSettings.value.registration_enabled    = data.registration_enabled !== 'false'
    systemSettings.value.mfa_required             = data.mfa_required === 'true' || data.mfa_required === true
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
    systemSettings.value.company_name             = data.company_name || ''
    systemSettings.value.company_logo             = data.company_logo || ''
    systemSettings.value.default_columns          = data.default_columns || 'Backlog'
    systemSettings.value.default_labels           = data.default_labels || 'Bug\nFeature\nDesign\nContent'
    systemSettings.value.password_min_length      = parseInt(data.password_min_length) || 8
    systemSettings.value.password_require_upper   = data.password_require_upper === 'true'
    systemSettings.value.password_require_lower   = data.password_require_lower === 'true'
    systemSettings.value.password_require_digit   = data.password_require_digit === 'true'
    systemSettings.value.password_require_special = data.password_require_special === 'true'
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

async function saveBrandingSettings() {
  try {
    await adminApi.updateSystemSettings({
      company_name: systemSettings.value.company_name,
      company_logo: systemSettings.value.company_logo,
    })
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function savePasswordPolicy() {
  try {
    await adminApi.updateSystemSettings({
      password_min_length:      systemSettings.value.password_min_length,
      password_require_upper:   systemSettings.value.password_require_upper,
      password_require_lower:   systemSettings.value.password_require_lower,
      password_require_digit:   systemSettings.value.password_require_digit,
      password_require_special: systemSettings.value.password_require_special,
    })
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function saveGeneralSettings() {
  try {
    const payload = {
      registration_enabled:     systemSettings.value.registration_enabled,
      session_timeout_minutes:  systemSettings.value.session_timeout_minutes,
      default_date_time_format: systemSettings.value.default_date_time_format,
      default_timezone:         systemSettings.value.default_timezone,
      default_theme:            systemSettings.value.default_theme,
      default_font:             systemSettings.value.default_font,
      default_font_size:        systemSettings.value.default_font_size,
      default_locale:           systemSettings.value.default_locale,
      default_columns:          systemSettings.value.default_columns,
      default_labels:           systemSettings.value.default_labels,
    }
    await adminApi.updateSystemSettings(payload)
    ui.success('Settings saved')
  } catch {
    ui.error('Failed to save settings')
  }
}

async function saveMFASettings() {
  try {
    await adminApi.updateSystemSettings({ mfa_required: systemSettings.value.mfa_required })
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
  showCreateUser.value = true
  loadProjects()
}

async function submitCreateUser() {
  try {
    const { data } = await adminApi.createUser(newUser.value)
    if (userProjectIds.value.length) {
      await adminApi.setUserProjects(data.id, userProjectIds.value)
    }
    users.value.push(data)
    showCreateUser.value = false
    newUser.value = { username: '', email: '', password: '', first_name: '', last_name: '', global_role: 'user' }
    userProjectIds.value = []
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

async function deleteUser(user) {
  if (!confirm(`Delete user ${user.username}?`)) return
  try {
    await adminApi.deleteUser(user.id)
    users.value = users.value.filter(u => u.id !== user.id)
    sidebarStore.fetchAllUsers()
    ui.success('User deleted')
  } catch {
    ui.error('Failed to delete user')
  }
}

async function adminResetMFA(user) {
  if (!confirm(`Disable MFA for ${user.display_name || user.username}?`)) return
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

async function openEditUser(user) {
  editUser.value = { ...user, _newPassword: '' }
  userProjectIds.value = []
  loadProjects()
  try {
    const { data } = await adminApi.getUserProjects(user.id)
    userProjectIds.value = data.project_ids || []
  } catch {}
}

async function saveEditUser() {
  try {
    const payload = {
      first_name: editUser.value.first_name,
      last_name: editUser.value.last_name,
      display_name: editUser.value.display_name,
      email: editUser.value.email,
      avatar_url: editUser.value.avatar_url,
      locale: editUser.value.locale
    }
    if (editUser.value._newPassword) {
      payload.password = editUser.value._newPassword
    }
    const { data } = await adminApi.updateUser(editUser.value.id, payload)
    await adminApi.setUserProjects(editUser.value.id, userProjectIds.value)
    const idx = users.value.findIndex(u => u.id === data.id)
    if (idx >= 0) users.value[idx] = data
    editUser.value = null
    userProjectIds.value = []
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
    newProject.value = { name: '', description: '', color: '#6366f1', key_prefix: '' }
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
      color: editProject.value.color
    })
    const idx = projects.value.findIndex(p => p.id === data.id)
    if (idx >= 0) projects.value[idx] = data
    editProject.value = null
    ui.success('Project updated')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to update project')
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
  if (!confirm(`Delete project "${project.name}"?`)) return
  try {
    await adminApi.deleteProject(project.id)
    projects.value = projects.value.filter(p => p.id !== project.id)
    sidebarStore.allProjects = sidebarStore.allProjects.filter(p => p.id !== project.id)
    ui.success('Project deleted')
  } catch {
    ui.error('Failed to delete project')
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
.data-table th, .data-table td { padding: 12px 16px; text-align: left; border-bottom: 1px solid var(--color-border); font-size: 13px; }
.data-table th { font-weight: 600; color: var(--color-text-muted); font-size: 12px; background: var(--color-bg); }
.data-table small { color: var(--color-text-muted); font-size: 11px; }
.email { color: var(--color-text-muted); }
.actions-cell { display: flex; gap: 6px; flex-wrap: wrap; }

.badge-admin { background: #ede9fe; color: #5b21b6; }
.badge-user { background: #f1f5f9; color: #64748b; }

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
.badge-mfa { background: #dbeafe; color: #1e40af; margin-left: 6px; }
[data-theme="dark"] .badge-active { background: #14532d; color: #86efac; }
[data-theme="dark"] .badge-inactive { background: #450a0a; color: #fca5a5; }
[data-theme="dark"] .badge-mfa { background: #1e3a5f; color: #93c5fd; }
.open-cards-count { font-weight: 600; color: var(--color-primary); }

.loading-state { display: flex; justify-content: center; padding: 60px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.tab-toolbar { margin-bottom: 16px; }
.settings-section { max-width: 560px; }
.settings-section h2 { font-size: 16px; font-weight: 600; margin-bottom: 16px; }
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
</style>
