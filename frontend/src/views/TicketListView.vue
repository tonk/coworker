<template>
  <main class="ticket-list-main">
    <div class="ticket-list-container">
      <DashboardNews />

      <!-- Customer tab bar — matches CustomerDetailView -->
      <div v-if="customer" class="cust-mini-header">
        <RouterLink :to="`/customers/${customerId}`" class="cust-mini-name">{{ customer.name }}</RouterLink>
      </div>
      <div class="cust-tabs" role="tablist" :aria-label="customer?.name || $t('ticket.tickets')">
        <RouterLink
          role="tab"
          :aria-selected="false"
          class="tab"
          :to="`/customers/${customerId}`"
        >{{ $t('customer.tab_overview') }}</RouterLink>
        <RouterLink
          role="tab"
          :aria-selected="false"
          class="tab"
          :to="`/customers/${customerId}?tab=invoices`"
        >{{ $t('invoice.invoices') }}</RouterLink>
        <RouterLink
          role="tab"
          :aria-selected="false"
          class="tab"
          :to="`/customers/${customerId}?tab=contacts`"
        >{{ $t('customer.contacts') }}</RouterLink>
        <span role="tab" aria-selected="true" class="tab active" aria-current="page">{{ $t('ticket.tickets') }}</span>
      </div>

      <header class="ticket-list-header">
        <div class="header-actions">
          <div class="view-toggle" role="tablist">
            <button role="tab" :aria-selected="viewMode === 'cards'" :class="['view-toggle-btn', { active: viewMode === 'cards' }]" @click="viewMode = 'cards'">☰ {{ $t('ticket.card_view') }}</button>
            <button role="tab" :aria-selected="viewMode === 'group'" :class="['view-toggle-btn', { active: viewMode === 'group' }]" @click="viewMode = 'group'">⊞ {{ $t('ticket.group_view') }}</button>
            <button role="tab" :aria-selected="viewMode === 'list'" :class="['view-toggle-btn', { active: viewMode === 'list' }]" @click="viewMode = 'list'">☷ {{ $t('ticket.list_view') }}</button>
          </div>
          <button :class="['btn btn-sm', showSpam ? 'btn-warning' : 'btn-secondary']" @click="showSpam = !showSpam">
            {{ showSpam ? $t('ticket.hide_spam') : $t('ticket.show_spam') }}
            <span v-if="!showSpam && spamCount > 0" class="spam-count-badge">{{ spamCount }}</span>
          </button>
          <button class="btn btn-primary btn-sm" @click="showCreate = true">+ {{ $t('ticket.new_ticket') }}</button>
        </div>
      </header>

      <div v-if="loading" class="loading-state">
        <div class="spinner" style="width:32px;height:32px;border-width:3px"></div>
      </div>

      <template v-else>
        <div v-if="!tickets.length" class="empty-state">
          {{ $t('ticket.no_tickets') }}
        </div>

        <!-- Card view -->
        <template v-else-if="viewMode === 'cards'">
          <div v-if="regularTickets.length" class="ticket-grid">
            <div
              v-for="t in regularTickets"
              :key="t.id"
              class="ticket-card"
              :class="'ticket-' + t.status"
              @click="openTicket(t)"
              role="button"
              tabindex="0"
              @keydown.enter="openTicket(t)"
              @keydown.space.prevent="openTicket(t)"
              :aria-label="t.title"
            >
              <div class="ticket-card-header">
                <span class="ticket-id">#{{ t.id }}</span>
                <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
              </div>
              <h3 class="ticket-card-title"><span v-if="t.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ t.title }}</h3>
              <div class="ticket-card-meta">
                <span class="ticket-status" :class="'status-' + t.status">{{ $t('ticket.status_' + t.status) }}</span>
                <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                  <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">
                  {{ t.assigned_to.display_name || t.assigned_to.username }}
                </span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>

          <div v-if="pendingReminderTickets.length" class="pending-reminder-divider">
            <span class="pending-reminder-label">{{ $t('ticket.pending_reminders') }}</span>
          </div>

          <div v-if="pendingReminderTickets.length" class="ticket-grid">
            <div
              v-for="t in pendingReminderTickets"
              :key="t.id"
              class="ticket-card"
              :class="'ticket-' + t.status"
              @click="openTicket(t)"
              role="button"
              tabindex="0"
              @keydown.enter="openTicket(t)"
              @keydown.space.prevent="openTicket(t)"
              :aria-label="t.title"
            >
              <div class="ticket-card-header">
                <span class="ticket-id">#{{ t.id }}</span>
                <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
              </div>
              <h3 class="ticket-card-title"><span v-if="t.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ t.title }}</h3>
              <div class="ticket-card-meta">
                <span class="ticket-status" :class="'status-' + t.status">{{ $t('ticket.status_' + t.status) }}</span>
                <span v-if="t.reminder_at" class="reminder-badge">
                  {{ $t('ticket.reminder') }}: {{ formatDate(t.reminder_at) }}
                </span>
                <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                  <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">
                  {{ t.assigned_to.display_name || t.assigned_to.username }}
                </span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>

          <div v-if="pendingCloseTickets.length" class="pending-reminder-divider">
            <span class="pending-reminder-label">{{ $t('ticket.resolved_closed') }}</span>
          </div>

          <div v-if="pendingCloseTickets.length" class="ticket-grid">
            <div
              v-for="t in pendingCloseTickets"
              :key="t.id"
              class="ticket-card"
              :class="'ticket-' + t.status"
              @click="openTicket(t)"
              role="button"
              tabindex="0"
              @keydown.enter="openTicket(t)"
              @keydown.space.prevent="openTicket(t)"
              :aria-label="t.title"
            >
              <div class="ticket-card-header">
                <span class="ticket-id">#{{ t.id }}</span>
                <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
              </div>
              <h3 class="ticket-card-title"><span v-if="t.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ t.title }}</h3>
              <div class="ticket-card-meta">
                <span class="ticket-status" :class="'status-' + t.status">{{ $t('ticket.status_' + t.status) }}</span>
                <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                  <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">
                  {{ t.assigned_to.display_name || t.assigned_to.username }}
                </span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>

          <div v-if="closedTickets.length" class="pending-reminder-divider">
            <span class="pending-reminder-label">{{ $t('ticket.status_closed') }}</span>
          </div>

          <div v-if="closedTickets.length" class="ticket-grid">
            <div
              v-for="t in closedTickets"
              :key="t.id"
              class="ticket-card"
              :class="'ticket-' + t.status"
              @click="openTicket(t)"
              role="button"
              tabindex="0"
              @keydown.enter="openTicket(t)"
              @keydown.space.prevent="openTicket(t)"
              :aria-label="t.title"
            >
              <div class="ticket-card-header">
                <span class="ticket-id">#{{ t.id }}</span>
                <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
              </div>
              <h3 class="ticket-card-title"><span v-if="t.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ t.title }}</h3>
              <div class="ticket-card-meta">
                <span class="ticket-status" :class="'status-' + t.status">{{ $t('ticket.status_' + t.status) }}</span>
                <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                <span v-if="t.tags?.length" class="ticket-tags">
                  <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                  <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
                </span>
                <span v-if="t.assigned_to" class="ticket-assignee">
                  {{ t.assigned_to.display_name || t.assigned_to.username }}
                </span>
                <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
              </div>
            </div>
          </div>
        </template>

        <!-- Group view -->
        <template v-else-if="viewMode === 'group'">
          <div class="group-sub-toggle">
            <button :class="['group-sub-btn', { active: groupSubMode === 'cards' }]" @click="groupSubMode = 'cards'">☰ {{ $t('ticket.card_view') }}</button>
            <button :class="['group-sub-btn', { active: groupSubMode === 'list' }]" @click="groupSubMode = 'list'">☷ {{ $t('ticket.list_view') }}</button>
          </div>
          <div v-for="g in groupedTickets" :key="g.status" class="group-section">
            <div class="group-header">
              <h2 class="group-title">{{ g.label }}</h2>
              <span class="group-count">{{ g.tickets.length }}</span>
            </div>
            <div v-if="g.tickets.length && groupSubMode === 'cards'" class="ticket-grid">
              <div
                v-for="t in g.tickets"
                :key="t.id"
                class="ticket-card"
                :class="'ticket-' + t.status"
                @click="openTicket(t)"
                role="button"
                tabindex="0"
                @keydown.enter="openTicket(t)"
                @keydown.space.prevent="openTicket(t)"
                :aria-label="t.title"
              >
                <div class="ticket-card-header">
                  <span class="ticket-id">#{{ t.id }}</span>
                  <span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span>
                  <span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span>
                </div>
                <h3 class="ticket-card-title"><span v-if="t.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ t.title }}</h3>
                <div class="ticket-card-meta">
                  <span v-if="t.status === 'pending' && t.reminder_at" class="reminder-badge">
                    {{ $t('ticket.reminder') }}: {{ formatDate(t.reminder_at) }}
                  </span>
                  <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                  <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                  <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                  <span v-if="t.tags?.length" class="ticket-tags">
                    <span v-for="tag in t.tags.slice(0, 3)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                    <span v-if="t.tags.length > 3" class="mini-tag more">+{{ t.tags.length - 3 }}</span>
                  </span>
                  <span v-if="t.assigned_to" class="ticket-assignee">
                    {{ t.assigned_to.display_name || t.assigned_to.username }}
                  </span>
                  <span class="ticket-date">{{ formatDate(t.created_at) }}</span>
                </div>
              </div>
            </div>
            <table v-else-if="g.tickets.length && groupSubMode === 'list'" class="group-table">
              <thead>
                <tr>
                  <th class="th-sort" :class="groupSortClass('id')" @click="groupToggleSort('id')"># <span class="sort-arrow">{{ groupSortArrow('id') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('title')" @click="groupToggleSort('title')">{{ $t('ticket.title') }} <span class="sort-arrow">{{ groupSortArrow('title') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('priority')" @click="groupToggleSort('priority')">{{ $t('ticket.priority') }} <span class="sort-arrow">{{ groupSortArrow('priority') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('type')" @click="groupToggleSort('type')">{{ $t('ticket.type') }} <span class="sort-arrow">{{ groupSortArrow('type') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('assigned_to')" @click="groupToggleSort('assigned_to')">{{ $t('ticket.assigned_to') }} <span class="sort-arrow">{{ groupSortArrow('assigned_to') }}</span></th>
                  <th class="th-sort" :class="groupSortClass('created_at')" @click="groupToggleSort('created_at')">{{ $t('ticket.created_at') }} <span class="sort-arrow">{{ groupSortArrow('created_at') }}</span></th>
                  <th>{{ $t('ticket.tags') }}</th>
                  <th>SLA</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="t in g.tickets" :key="t.id" class="table-row" @click="openTicket(t)" tabindex="0" @keydown.enter="openTicket(t)">
                  <td class="td-id">#{{ t.id }}</td>
                  <td class="td-title"><span v-if="t.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ t.title }}</td>
                  <td><span class="ticket-priority" :class="'pri-' + t.priority">{{ t.priority }}</span></td>
                  <td><span class="ticket-type" :class="'type-' + t.type">{{ $t('ticket.type_' + t.type) }}</span></td>
                  <td class="td-assignee">{{ t.assigned_to?.display_name || t.assigned_to?.username || '—' }}</td>
                  <td class="td-date">{{ formatDate(t.created_at) }}</td>
                  <td>
                    <span v-if="t.tags?.length" class="ticket-tags">
                      <span v-for="tag in t.tags.slice(0, 2)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                      <span v-if="t.tags.length > 2" class="mini-tag more">+{{ t.tags.length - 2 }}</span>
                    </span>
                  </td>
                  <td>
                    <span v-if="t.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(t)">{{ $t('sla.breached') }}</span>
                    <span v-else-if="slaWarning(t)" class="sla-badge sla-warning" :title="slaTitle(t)">{{ $t('sla.warning') }}</span>
                    <span v-else-if="t.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(t)">{{ $t('sla.on_track') }}</span>
                    <span v-else>—</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <!-- List view -->
        <template v-else-if="viewMode === 'list'">
          <template v-for="section in listSections" :key="section.key">
            <div class="pending-reminder-divider"><span class="pending-reminder-label">{{ section.label }}</span></div>
            <table class="ticket-table">
              <thead>
                <tr>
                  <th class="th-sort" :class="sortClass('id')" @click="toggleSort('id')"># <span class="sort-arrow">{{ sortArrow('id') }}</span></th>
                  <th class="th-sort th-title" :class="sortClass('title')" @click="toggleSort('title')">{{ $t('ticket.title') }} <span class="sort-arrow">{{ sortArrow('title') }}</span></th>
                  <th class="th-sort" :class="sortClass('priority')" @click="toggleSort('priority')">{{ $t('ticket.priority') }} <span class="sort-arrow">{{ sortArrow('priority') }}</span></th>
                  <th class="th-sort" :class="sortClass('type')" @click="toggleSort('type')">{{ $t('ticket.type') }} <span class="sort-arrow">{{ sortArrow('type') }}</span></th>
                  <th class="th-sort" :class="sortClass('assigned_to')" @click="toggleSort('assigned_to')">{{ $t('ticket.assigned_to') }} <span class="sort-arrow">{{ sortArrow('assigned_to') }}</span></th>
                  <th class="th-sort" :class="sortClass('created_at')" @click="toggleSort('created_at')">{{ $t('ticket.created_at') }} <span class="sort-arrow">{{ sortArrow('created_at') }}</span></th>
                  <th class="th-tags">{{ $t('ticket.tags') }}</th>
                  <th>SLA</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="tk in section.tickets" :key="tk.id" class="table-row" @click="openTicket(tk)" tabindex="0" @keydown.enter="openTicket(tk)" :aria-label="tk.title">
                  <td class="td-id">#{{ tk.id }}</td>
                  <td class="td-title"><span v-if="tk.is_spam" class="spam-tag">{{ $t('ticket.spam') }}</span>{{ tk.title }}</td>
                  <td><span class="ticket-priority" :class="'pri-' + tk.priority">{{ tk.priority }}</span></td>
                  <td><span class="ticket-type" :class="'type-' + tk.type">{{ $t('ticket.type_' + tk.type) }}</span></td>
                  <td class="td-assignee">{{ tk.assigned_to?.display_name || tk.assigned_to?.username || '—' }}</td>
                  <td class="td-date">{{ formatDate(tk.created_at) }}</td>
                  <td>
                    <span v-if="tk.tags?.length" class="ticket-tags">
                      <span v-for="tag in tk.tags.slice(0, 2)" :key="tag.id" class="mini-tag">#{{ tag.name }}</span>
                      <span v-if="tk.tags.length > 2" class="mini-tag more">+{{ tk.tags.length - 2 }}</span>
                    </span>
                  </td>
                  <td>
                    <span v-if="tk.sla_response_breached" class="sla-badge sla-breach" :title="slaTitle(tk)">{{ $t('sla.breached') }}</span>
                    <span v-else-if="slaWarning(tk)" class="sla-badge sla-warning" :title="slaTitle(tk)">{{ $t('sla.warning') }}</span>
                    <span v-else-if="tk.sla_policy_id" class="sla-badge sla-ok" :title="slaTitle(tk)">{{ $t('sla.on_track') }}</span>
                    <span v-else>—</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </template>
        </template>
      </template>
    </div>

    <BaseModal v-if="showCreate" :title="$t('ticket.new_ticket')" @close="showCreate = false">
      <form @submit.prevent="submitCreate">
        <div class="form-group">
          <label>{{ $t('ticket.title') }}</label>
          <input v-model="newTicket.title" class="form-input" required :placeholder="$t('ticket.title_placeholder')" />
        </div>
        <div class="form-group">
          <label>{{ $t('ticket.description') }}</label>
          <div class="md-editor">
            <div class="md-editor-tabs" role="tablist">
              <button role="tab" :aria-selected="descTab === 'edit'" aria-controls="ticket-create-panel-edit" id="ticket-create-tab-edit" :class="['md-tab', { active: descTab === 'edit' }]" type="button" @click="descTab = 'edit'">{{ $t('common.edit') }}</button>
              <button role="tab" :aria-selected="descTab === 'preview'" aria-controls="ticket-create-panel-preview" id="ticket-create-tab-preview" :class="['md-tab', { active: descTab === 'preview' }]" type="button" @click="descTab = 'preview'">{{ $t('common.preview') }}</button>
            </div>
            <textarea v-if="descTab === 'edit'" id="ticket-create-panel-edit" role="tabpanel" aria-labelledby="ticket-create-tab-edit" v-model="newTicket.description" class="form-input md-textarea" rows="4" @paste="onDescPaste"></textarea>
            <div v-else id="ticket-create-panel-preview" role="tabpanel" aria-labelledby="ticket-create-tab-preview" class="md-preview markdown-body" v-html="renderMarkdown(newTicket.description)"></div>
          </div>
          <div v-if="pendingFiles.length" class="pending-attachments">
            <span class="pending-label">{{ $t('ticket.pending_files') }}</span>
            <AttachmentList :attachments="pendingFiles" :can-delete="true" @remove="removePending" />
          </div>
          <FileUploadButton @files-selected="onFilesSelected" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>{{ $t('ticket.type') }}</label>
            <select v-model="newTicket.type" class="form-input">
              <option value="incident">{{ $t('ticket.type_incident') }}</option>
              <option value="problem">{{ $t('ticket.type_problem') }}</option>
              <option value="service_request">{{ $t('ticket.type_service_request') }}</option>
              <option value="change_request">{{ $t('ticket.type_change_request') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>{{ $t('ticket.priority') }}</label>
            <select v-model="newTicket.priority" class="form-input">
              <option value="low">{{ $t('ticket.priority_low') }}</option>
              <option value="medium" selected>{{ $t('ticket.priority_medium') }}</option>
              <option value="high">{{ $t('ticket.priority_high') }}</option>
              <option value="critical">{{ $t('ticket.priority_critical') }}</option>
            </select>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>{{ $t('ticket.group') }}</label>
            <select v-model="newTicket.group_id" class="form-input">
              <option :value="null">—</option>
              <option v-for="g in customerGroups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>{{ $t('ticket.owner') }}</label>
            <select v-model="newTicket.owner_id" class="form-input">
              <option :value="null">—</option>
              <option v-for="u in customerUsers" :key="u.id" :value="u.id">{{ u.display_name || u.username }}</option>
            </select>
          </div>
        </div>
        <div class="form-group">
          <label>{{ $t('ticket.tags') }}</label>
          <div class="tags-editor">
            <div class="tags-list" v-if="newTicketTags.length">
              <span v-for="(tag, i) in newTicketTags" :key="i" class="tag-chip">
                #{{ tag }}
                <button class="tag-remove" @click="newTicketTags.splice(i, 1)" title="Remove tag" aria-label="Remove tag">×</button>
              </span>
            </div>
            <input class="form-input tag-input" v-model="newTagName" :placeholder="$t('ticket.add_tag_placeholder')" @keydown.enter.prevent="addNewTag" @keydown.comma.prevent="addNewTag" />
          </div>
        </div>
        <div class="modal-footer" slot="footer">
          <button type="button" class="btn btn-secondary" @click="showCreate = false">{{ $t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary">{{ $t('common.create') }}</button>
        </div>
      </form>
    </BaseModal>
  </main>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import client from '@/api/client'
import { customersApi } from '@/api/customers'
import { ticketsApi } from '@/api/tickets'
import { attachmentsApi } from '@/api/attachments'
import { useUIStore } from '@/stores/ui'
import BaseModal from '@/components/common/BaseModal.vue'
import DashboardNews from '@/components/common/DashboardNews.vue'
import AttachmentList from '@/components/common/AttachmentList.vue'
import FileUploadButton from '@/components/common/FileUploadButton.vue'
import { useDateFormat } from '@/composables/useDateFormat'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const { t } = useI18n()
const { formatDate } = useDateFormat()

function renderMarkdown(text) {
  return DOMPurify.sanitize(marked.parse(text || ''))
}

const route = useRoute()
const router = useRouter()
const ui = useUIStore()

const customerId = computed(() => Number(route.params.id))
const customer = ref(null)
const tickets = ref([])
const loading = ref(true)
const showCreate = ref(false)
const descTab = ref('edit')
const newTicket = ref({ title: '', description: '', type: 'incident', priority: 'medium', group_id: null, owner_id: null })
const newTicketTags = ref([])
const newTagName = ref('')
const pendingFiles = ref([])
const customerGroups = ref([])
const customerUsers = ref([])
const viewMode = ref(localStorage.getItem('ticket_view_mode') || 'cards')
const showSpam = ref(false)

watch(viewMode, (val) => {
  localStorage.setItem('ticket_view_mode', val)
})

const spamCount = computed(() => tickets.value.filter(t => t.is_spam).length)
const visibleTickets = computed(() => showSpam.value ? tickets.value : tickets.value.filter(t => !t.is_spam))
const groupSubMode = ref(localStorage.getItem('ticket_group_sub_mode') || 'cards')

watch(groupSubMode, (val) => {
  localStorage.setItem('ticket_group_sub_mode', val)
})
const sortField = ref('created_at')
const sortDir = ref(-1)
const groupSortField = ref('created_at')
const groupSortDir = ref(-1)

const priorityRank = { low: 1, medium: 2, high: 3, critical: 4 }

function isPendingReminder(t) {
  return t.status === 'pending' && t.reminder_at
}

function isResolvedOrClosed(t) {
  return t.status === 'pending_close' || t.status === 'closed'
}

const regularTickets = computed(() => {
  return visibleTickets.value.filter(t => !isPendingReminder(t) && !isResolvedOrClosed(t))
})

const pendingReminderTickets = computed(() => {
  return visibleTickets.value
    .filter(t => isPendingReminder(t))
    .sort((a, b) => new Date(a.reminder_at) - new Date(b.reminder_at))
})

const pendingCloseTickets = computed(() => {
  return visibleTickets.value.filter(t => t.status === 'pending_close')
})

const closedTickets = computed(() => {
  return visibleTickets.value.filter(t => t.status === 'closed')
})

const statusGroups = [
  { status: 'new', label: 'New' },
  { status: 'open', label: 'Open' },
  { status: 'pending', label: 'Pending reminder' },
  { status: 'pending_close', label: 'Pending close' },
  { status: 'closed', label: 'Closed' },
]

const groupedTickets = computed(() => {
  const f = groupSortField.value
  const d = groupSortDir.value
  return statusGroups.map(g => ({
    ...g,
    tickets: visibleTickets.value.filter(t => t.status === g.status).sort((a, b) => {
      let va, vb
      if (f === 'priority') {
        va = priorityRank[a.priority] || 0
        vb = priorityRank[b.priority] || 0
      } else if (f === 'type') {
        va = a.type
        vb = b.type
      } else if (f === 'assigned_to') {
        va = (a.assigned_to?.display_name || a.assigned_to?.username || '').toLowerCase()
        vb = (b.assigned_to?.display_name || b.assigned_to?.username || '').toLowerCase()
      } else if (f === 'id') {
        va = a.id
        vb = b.id
      } else if (f === 'title') {
        va = a.title.toLowerCase()
        vb = b.title.toLowerCase()
      } else {
        va = a[f]
        vb = b[f]
      }
      if (va < vb) return -1 * d
      if (va > vb) return 1 * d
      return 0
    }),
  })).filter(g => g.tickets.length > 0)
})

function sortByField(arr, f, d) {
  return [...arr].sort((a, b) => {
    let va, vb
    if (f === 'priority') { va = priorityRank[a.priority] || 0; vb = priorityRank[b.priority] || 0 }
    else if (f === 'type') { va = a.type; vb = b.type }
    else if (f === 'assigned_to') { va = (a.assigned_to?.display_name || a.assigned_to?.username || '').toLowerCase(); vb = (b.assigned_to?.display_name || b.assigned_to?.username || '').toLowerCase() }
    else if (f === 'id') { va = a.id; vb = b.id }
    else if (f === 'title') { va = a.title.toLowerCase(); vb = b.title.toLowerCase() }
    else { va = a[f]; vb = b[f] }
    if (va < vb) return -1 * d
    if (va > vb) return 1 * d
    return 0
  })
}

const listSections = computed(() => [
  { key: 'new',            label: t('ticket.status_new'),        tickets: visibleTickets.value.filter(tk => tk.status === 'new') },
  { key: 'open',           label: t('ticket.status_open'),       tickets: visibleTickets.value.filter(tk => tk.status === 'open') },
  { key: 'pending',        label: t('ticket.status_pending'),    tickets: visibleTickets.value.filter(tk => tk.status === 'pending' && !tk.reminder_at) },
  { key: 'pending-remind', label: t('ticket.pending_reminders'), tickets: visibleTickets.value.filter(tk => tk.status === 'pending' && tk.reminder_at).sort((a, b) => new Date(a.reminder_at) - new Date(b.reminder_at)) },
  { key: 'pending-close',  label: t('ticket.resolved_closed'),   tickets: visibleTickets.value.filter(tk => tk.status === 'pending_close') },
  { key: 'closed',         label: t('ticket.status_closed'),     tickets: visibleTickets.value.filter(tk => tk.status === 'closed') },
].filter(s => s.tickets.length > 0).map(s => ({ ...s, tickets: sortByField(s.tickets, sortField.value, sortDir.value) }))
)

const sortedTickets = computed(() => sortByField(visibleTickets.value, sortField.value, sortDir.value))

function toggleSort(field) {
  if (sortField.value === field) {
    sortDir.value *= -1
  } else {
    sortField.value = field
    sortDir.value = -1
  }
}

function sortClass(field) {
  if (sortField.value !== field) return ''
  return sortDir.value === -1 ? 'sort-desc' : 'sort-asc'
}

function sortArrow(field) {
  if (sortField.value !== field) return '▽'
  return sortDir.value === -1 ? '▽' : '△'
}

function groupToggleSort(field) {
  if (groupSortField.value === field) {
    groupSortDir.value *= -1
  } else {
    groupSortField.value = field
    groupSortDir.value = -1
  }
}

function groupSortClass(field) {
  if (groupSortField.value !== field) return ''
  return groupSortDir.value === -1 ? 'sort-desc' : 'sort-asc'
}

function groupSortArrow(field) {
  if (groupSortField.value !== field) return '▽'
  return groupSortDir.value === -1 ? '▽' : '△'
}

async function fetchData() {
  loading.value = true
  pendingFiles.value = []
  try {
    const { data } = await customersApi.get(customerId.value)
    customer.value = data.customer || data
  } catch {}
  try {
    const { data } = await ticketsApi.list(customerId.value, { include_spam: true })
    tickets.value = data || []
  } catch {}
  try {
    const { data } = await client.get(`/customers/${customerId.value}/groups`)
    customerGroups.value = data || []
  } catch {}
  try {
    const { data } = await client.get(`/customers/${customerId.value}/members`)
    customerUsers.value = data || []
  } catch {}
  loading.value = false
}

watch(() => route.params.id, () => {
  fetchData()
})

onMounted(fetchData)

function openTicket(t) {
  router.push(`/customers/${customerId.value}/tickets/${t.id}`)
}

function slaWarning(t) {
  if (!t.sla_response_deadline && !t.sla_resolution_deadline) return false
  if (t.status === 'pending_close') return false
  const now = Date.now()
  if (t.sla_response_deadline && !t.first_response_at) {
    const deadline = new Date(t.sla_response_deadline).getTime()
    if (deadline - now < 3600000 && deadline - now > 0) return true
  }
  if (t.sla_resolution_deadline) {
    const deadline = new Date(t.sla_resolution_deadline).getTime()
    if (deadline - now < 3600000 && deadline - now > 0) return true
  }
  return false
}

function slaTitle(t) {
  const parts = []
  if (t.sla_policy?.name) parts.push(t.sla_policy.name)
  if (t.sla_response_deadline) {
    parts.push(`Response: ${formatDate(t.sla_response_deadline)}`)
  }
  if (t.sla_resolution_deadline) {
    parts.push(`Resolution: ${formatDate(t.sla_resolution_deadline)}`)
  }
  return parts.join(' | ')
}

async function submitCreate() {
  try {
    const res = await ticketsApi.create(customerId.value, newTicket.value)
    const ticket = res.data || res

    if (pendingFiles.value.length) {
      const filesToUpload = [...pendingFiles.value]
      pendingFiles.value = []
      filesToUpload.forEach(pf => { if (pf._previewUrl) URL.revokeObjectURL(pf._previewUrl) })
      for (const pf of filesToUpload) {
        const fd = new FormData()
        fd.append('file', pf._file)
        fd.append('owner_type', 'ticket')
        fd.append('owner_id', String(ticket.id))
        try {
          await attachmentsApi.upload(fd)
        } catch {}
      }
    }

    if (newTicketTags.value.length) {
      for (const tagName of newTicketTags.value) {
        try {
          await ticketsApi.addTag(customerId.value, ticket.id, tagName)
        } catch {}
      }
    }

    tickets.value.unshift(ticket)
    showCreate.value = false
    descTab.value = 'edit'
    newTicket.value = { title: '', description: '', type: 'incident', priority: 'medium', group_id: null, owner_id: null }
    newTicketTags.value = []
    newTagName.value = ''
    ui.success('Ticket created')
  } catch (e) {
    ui.error(e.response?.data?.error || 'Failed to create ticket')
  }
}

function onFilesSelected(files) {
  for (const f of files) {
    pendingFiles.value.push({
      id: Math.random(),
      filename: f.name,
      size_bytes: f.size,
      mime_type: f.type || 'application/octet-stream',
      _file: f,
      _previewUrl: f.type?.startsWith('image/') ? URL.createObjectURL(f) : null,
    })
  }
}

function removePending(a) {
  if (a._previewUrl) URL.revokeObjectURL(a._previewUrl)
  pendingFiles.value = pendingFiles.value.filter(p => p.id !== a.id)
}

function addNewTag() {
  const name = newTagName.value.trim().replace(/^#/, '')
  if (!name || newTicketTags.value.includes(name)) return
  newTicketTags.value.push(name)
  newTagName.value = ''
}

async function onDescPaste(e) {
  const items = Array.from(e.clipboardData?.items || [])
  const images = items.filter(it => it.kind === 'file' && it.type.startsWith('image/'))
  if (images.length) {
    e.preventDefault()
    const files = await Promise.all(images.map(it => {
      const file = it.getAsFile()
      if (file) return file
      return it.getType(it.type).then(blob => {
        const ext = it.type.split('/')[1]?.split('+')[0] || 'png'
        return new File([blob], `clipboard.${ext}`, { type: it.type })
      })
    }))
    const valid = files.filter(Boolean)
    if (valid.length) {
      onFilesSelected(valid)
      ui.success(valid.length > 1 ? `${valid.length} images pasted` : 'Image pasted')
    }
    return
  }
  if (window.__TAURI_INTERNALS__ && navigator.clipboard?.read) {
    try {
      const clipItems = await navigator.clipboard.read()
      const files = []
      for (const item of clipItems) {
        for (const type of item.types) {
          if (type.startsWith('image/')) {
            const blob = await item.getType(type)
            const ext = type.split('/')[1]?.split('+')[0] || 'png'
            files.push(new File([blob], `paste.${ext}`, { type }))
          }
        }
      }
      if (files.length) { e.preventDefault(); onFilesSelected(files); ui.success('Image pasted') }
    } catch {}
  }
}
</script>

<style scoped>
.ticket-list-main { padding: 24px; margin: 0 auto; }
.ticket-list-main:has(.ticket-table) { max-width: 100%; padding: 24px 32px; }
.ticket-list-main:not(:has(.ticket-table)) { max-width: 1200px; }

.cust-mini-header { margin-bottom: 4px; }
.cust-mini-name { font-size: 13px; font-weight: 600; color: var(--color-text-muted); text-decoration: none; }
.cust-mini-name:hover { color: var(--color-primary); }

.cust-tabs {
  display: flex;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 20px;
}
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
  text-decoration: none;
  display: inline-block;
  transition: color 0.15s;
}
.tab:hover { color: var(--color-text); }
.tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.tab:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }

.ticket-list-header { display: flex; align-items: center; gap: 8px; margin-bottom: 20px; flex-wrap: wrap; }
.header-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.view-toggle { display: flex; border: 1px solid var(--color-border); border-radius: 6px; overflow: hidden; }
.view-toggle-btn { background: none; border: none; padding: 5px 12px; font-size: 12px; cursor: pointer; color: var(--color-text-muted); transition: background .15s, color .15s; }
.view-toggle-btn:not(:last-child) { border-right: 1px solid var(--color-border); }
.view-toggle-btn:hover { background: var(--color-bg-alt); }
.view-toggle-btn.active { background: var(--color-primary); color: #fff; }
.loading-state { display: flex; justify-content: center; padding: 48px; }
.empty-state { text-align: center; padding: 64px 24px; color: var(--color-text-muted); }
.ticket-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px; }
.ticket-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: box-shadow .15s, border-color .15s;
}
.ticket-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,.08); border-color: var(--color-primary); }
.ticket-card:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }
.ticket-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.ticket-id { font-size: 11px; font-weight: 700; color: var(--color-text-muted); }
.ticket-type { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; }
.type-incident { background: #fecaca; color: #b91c1c; }
.type-problem { background: #fed7aa; color: #9a3412; }
.type-service_request { background: #d1fae5; color: #065f46; }
.type-change_request { background: #dbeafe; color: #1e40af; }
.mini-tag { font-size: 10px; background: var(--color-bg-alt); padding: 1px 5px; border-radius: 3px; color: var(--color-text-muted); }
.spam-tag { display: inline-block; font-size: 10px; font-weight: 700; background: #fecaca; color: #b91c1c; padding: 1px 5px; border-radius: 3px; margin-right: 5px; text-transform: uppercase; letter-spacing: 0.03em; }
[data-theme="dark"] .spam-tag { background: #7f1d1d; color: #fca5a5; }
.spam-count-badge { display: inline-flex; align-items: center; justify-content: center; min-width: 18px; height: 18px; padding: 0 5px; border-radius: 9px; background: #ef4444; color: #fff; font-size: 11px; font-weight: 700; line-height: 1; margin-left: 4px; }
.btn-warning { background: #f59e0b; color: #fff; border-color: #f59e0b; }
.btn-warning:hover { background: #d97706; border-color: #d97706; }
.mini-tag.more { font-weight: 700; }
.ticket-tags { display: flex; gap: 3px; align-items: center; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.tags-editor { display: flex; flex-direction: column; gap: 6px; }
.tags-list { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-chip { display: inline-flex; align-items: center; gap: 3px; font-size: 12px; background: var(--color-bg-alt); padding: 2px 8px; border-radius: 4px; }
.tag-remove { background: none; border: none; cursor: pointer; font-size: 14px; line-height: 1; color: var(--color-text-muted); padding: 0; }
.tag-input { flex: 1; }
.ticket-priority { font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.pri-low { background: #e0f2fe; color: #0369a1; }
.pri-medium { background: #fef3c7; color: #92400e; }
.pri-high { background: #fde68a; color: #b45309; }
.pri-critical { background: #fecaca; color: #b91c1c; }
.ticket-card-title { font-size: 14px; font-weight: 600; margin: 0 0 8px; line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.ticket-card-meta { display: flex; gap: 8px; align-items: center; font-size: 11px; color: var(--color-text-muted); flex-wrap: wrap; }
.ticket-status { padding: 2px 6px; border-radius: 4px; font-weight: 600; }
.status-new { background: #dbeafe; color: #1e40af; }
.status-open { background: #fef3c7; color: #92400e; }
.status-pending { background: #f0e6ff; color: #6b21a8; }
.status-pending_close { background: #d1fae5; color: #065f46; }
.status-closed { background: #e5e7eb; color: #374151; }
.ticket-assignee { display: flex; align-items: center; gap: 4px; }
.sla-badge { font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; }
.sla-ok { background: #d1fae5; color: #065f46; }
.sla-warning { background: #fef3c7; color: #92400e; }
.sla-breach { background: #fecaca; color: #b91c1c; }
.pending-reminder-divider {
  display: flex; align-items: center; gap: 12px;
  margin: 40px 0 16px;
  color: var(--color-primary);
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.pending-reminder-divider::before {
  content: '';
  flex: 1;
  height: 2px;
  background: var(--color-primary);
  opacity: 0.4;
}
.pending-reminder-divider::after {
  content: '';
  flex: 1;
  height: 2px;
  background: var(--color-primary);
  opacity: 0.4;
}
.reminder-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 3px;
  background: #fef3c7;
  color: #92400e;
}

/* Group view */
.group-section { margin-bottom: 32px; }
.group-header { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; }
.group-title { font-size: 16px; font-weight: 700; margin: 0; }
.group-count { font-size: 11px; font-weight: 600; color: #fff; background: var(--color-primary); padding: 1px 7px; border-radius: 9999px; line-height: 20px; }
.group-section:not(:last-child)::after { content: ''; display: block; height: 1px; background: var(--color-border); margin-top: 32px; }
.group-sub-toggle { display: flex; gap: 4px; margin-bottom: 20px; }
.group-sub-btn { font-size: 12px; font-weight: 600; padding: 4px 12px; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-bg); color: var(--color-text-muted); cursor: pointer; transition: all .15s; }
.group-sub-btn.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.group-sub-btn:hover:not(.active) { border-color: var(--color-primary); color: var(--color-primary); }
.group-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.group-table th { text-align: left; padding: 6px 8px; border-bottom: 2px solid var(--color-border); font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: var(--color-text-muted); white-space: nowrap; }
.group-table td { padding: 6px 8px; border-bottom: 1px solid var(--color-border); vertical-align: middle; }
.group-table .table-row { cursor: pointer; }
.group-table .table-row:hover { background: var(--color-bg-alt); }

/* Table view */
.ticket-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.ticket-table th { text-align: left; padding: 8px 10px; border-bottom: 2px solid var(--color-border); font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; color: var(--color-text-muted); white-space: nowrap; user-select: none; }
.ticket-table td { padding: 8px 10px; border-bottom: 1px solid var(--color-border); vertical-align: middle; }
.th-sort { cursor: pointer; }
.th-sort:hover { color: var(--color-text); }
.sort-arrow { font-size: 10px; margin-left: 2px; }
.th-sort.sort-desc .sort-arrow { color: var(--color-primary); }
.th-sort.sort-asc .sort-arrow { color: var(--color-primary); }
.th-title { min-width: 200px; }
.td-title { font-weight: 600; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.td-id { font-family: monospace; font-size: 12px; color: var(--color-text-muted); font-weight: 700; }
.td-assignee { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 120px; color: var(--color-text-muted); }
.td-date { white-space: nowrap; color: var(--color-text-muted); font-size: 12px; }
.table-row { cursor: pointer; transition: background .1s; }
.table-row:hover { background: var(--color-bg-alt); }
.table-row:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }

/* Dark mode */
[data-theme="dark"] .type-incident { background: #3a1015; color: #ef4444; }
[data-theme="dark"] .type-problem { background: #3a2010; color: #f97316; }
[data-theme="dark"] .type-service_request { background: #0a2a1a; color: #34d399; }
[data-theme="dark"] .type-change_request { background: #0a1a3a; color: #60a5fa; }
[data-theme="dark"] .pri-low { background: #1a3a24; color: #6fcf97; }
[data-theme="dark"] .pri-medium { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .pri-high { background: #3a2010; color: #f2994a; }
[data-theme="dark"] .pri-critical { background: #3a1015; color: #eb5757; }
[data-theme="dark"] .status-new { background: #0a1a3a; color: #60a5fa; }
[data-theme="dark"] .status-open { background: #3a3010; color: #f2c94c; }
[data-theme="dark"] .status-pending { background: #2a1040; color: #c084fc; }
[data-theme="dark"] .status-pending_close { background: #0a2a1a; color: #34d399; }
[data-theme="dark"] .status-closed { background: #1f2937; color: #9ca3af; }

/* Markdown editor tabs */
:deep(.md-editor) { border: 1px solid var(--color-border); border-radius: var(--radius); overflow: hidden; }
:deep(.md-editor-tabs) { display: flex; background: var(--color-bg-alt); border-bottom: 1px solid var(--color-border); }
:deep(.md-tab) { padding: 6px 16px; font-size: 12px; font-weight: 600; cursor: pointer; background: none; border: none; border-bottom: 2px solid transparent; color: var(--color-text-muted); transition: color .15s, border-color .15s; }
:deep(.md-tab:hover) { color: var(--color-text); }
:deep(.md-tab.active) { color: var(--color-primary); border-bottom-color: var(--color-primary); }
:deep(.md-textarea) { border: none !important; border-radius: 0 !important; resize: vertical; }
:deep(.md-preview) { padding: 10px 12px; min-height: 80px; }
.pending-attachments { background: var(--color-bg-alt); border-radius: 6px; padding: 8px; margin-top: 4px; }
.pending-label { font-size: 11px; font-weight: 600; color: var(--color-text-muted); display: block; margin-bottom: 4px; }
</style>
