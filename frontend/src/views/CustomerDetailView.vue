<template>
  <div class="customer-detail-page">
    <div v-if="loading" class="loading">{{ $t('common.loading') }}</div>
    <template v-else-if="detail">

      <!-- Customer header -->
      <div class="cust-header">
        <div class="cust-logo-wrap">
          <img v-if="detail.customer.logo_url" :src="resolveAssetUrl(detail.customer.logo_url)" class="cust-logo" alt="" />
          <span v-else class="cust-logo-placeholder">{{ detail.customer.name[0] }}</span>
        </div>
        <div class="cust-info">
          <div v-if="!editingName" class="cust-name-row">
            <h1 class="cust-name">{{ detail.customer.name }}</h1>
            <button v-if="canManage" class="icon-btn" @click="startEditName" :aria-label="$t('common.edit')" title="Edit">✎</button>
          </div>
          <div v-else class="cust-name-row">
            <input class="form-input name-input" v-model="nameEdit" :aria-label="$t('customer.name')" @keydown.enter="saveNameEdit" @keydown.escape="editingName = false" />
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
            :aria-label="detail.customer.is_favorite ? $t('customer.unstar') : $t('customer.star')"
            :aria-pressed="detail.customer.is_favorite"
          >{{ detail.customer.is_favorite ? '★' : '☆' }}</button>
          <button v-if="canManage" class="btn btn-sm" @click="openEdit">{{ $t('customer.edit') }}</button>
          <button v-if="auth.isAdmin" class="btn btn-sm btn-danger" @click="doDelete">{{ $t('common.delete') }}</button>
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="cust-tabs" role="tablist" :aria-label="detail.customer.name">
        <button
          id="tab-btn-overview"
          role="tab"
          :aria-selected="activeTab === 'overview'"
          aria-controls="tab-panel-overview"
          :class="['tab', { active: activeTab === 'overview' }]"
          @click="activeTab = 'overview'"
        >{{ $t('customer.tab_overview') }}</button>
        <button
          id="tab-btn-invoices"
          role="tab"
          :aria-selected="activeTab === 'invoices'"
          aria-controls="tab-panel-invoices"
          :class="['tab', { active: activeTab === 'invoices' }]"
          @click="activeTab = 'invoices'"
        >{{ $t('invoice.invoices') }}</button>
        <button
          id="tab-btn-contacts"
          role="tab"
          :aria-selected="activeTab === 'contacts'"
          aria-controls="tab-panel-contacts"
          :class="['tab', { active: activeTab === 'contacts' }]"
          @click="activeTab = 'contacts'"
        >{{ $t('customer.contacts') }}</button>
        <button
          v-if="auth.helpdeskEnabled"
          role="tab"
          :aria-selected="false"
          :class="['tab']"
          @click="router.push(`/customers/${custId}/tickets`)"
        >{{ $t('ticket.tickets') }}</button>
      </div>

      <!-- Overview tab: contracts + members -->
      <div id="tab-panel-overview" role="tabpanel" aria-labelledby="tab-btn-overview" v-show="activeTab === 'overview'">

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
              <button v-if="canManage" class="contract-name-btn" @click="editContract(grp)" :aria-label="$t('contract.edit') + ': ' + grp.name">{{ grp.name }}</button>
              <strong v-else>{{ grp.name }}</strong>
              <span v-if="grp.start_date || grp.end_date" class="contract-dates">
                {{ formatDate(grp.start_date) }}{{ grp.end_date ? ' – ' + formatDate(grp.end_date) : '' }}
              </span>
              <span v-if="grp.price_per_hour != null" class="contract-rate">{{ grp.price_per_hour }} {{ grp.currency }}/h</span>
              <span v-if="grp.price_per_km != null" class="contract-rate">{{ grp.price_per_km }} {{ grp.currency }}/{{ distanceUnit }}</span>
            </div>
            <div v-if="grp.time_slots && grp.time_slots.length" class="slot-indicator-wrap">
              <span class="slot-badge" tabindex="0" :aria-label="grp.time_slots.length + ' ' + $t('contract.time_slots')">
                🕒 {{ grp.time_slots.length }}
              </span>
              <div class="slot-popup" role="tooltip">
                <div v-for="slot in grp.time_slots" :key="slot.id" class="slot-popup-item">
                  <span class="slot-popup-time">{{ formatSlotTimeRange(slot) }}</span>
                  <span v-if="slot.day_type && slot.day_type !== 'all'" class="slot-popup-days">{{ formatDayType(slot.day_type, $t) }}</span>
                  <span v-if="slot.label" class="slot-popup-label">{{ slot.label }}</span>
                  <span v-if="slot.multiplication_factor != null" class="slot-popup-factor">×{{ slot.multiplication_factor }}</span>
                  <span v-if="slot.hourly_rate != null" class="slot-popup-rate">{{ slot.hourly_rate }} {{ grp.currency }}/h</span>
                </div>
              </div>
            </div>
            <div class="contract-actions">
              <button v-if="canManage" class="icon-btn" @click="editContract(grp)" :aria-label="$t('contract.edit')" title="Edit">✎</button>
              <button v-if="auth.isAdmin" class="icon-btn icon-danger" @click="deleteContract(grp)" :aria-label="$t('common.delete')" title="Delete">✕</button>
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
              <img v-if="projectAvatar(p)" :src="projectAvatar(p)" class="proj-avatar" alt="" />
              <span v-else class="proj-dot" :style="{ background: p.color || '#6366f1' }"></span>
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
              <img v-if="projectAvatar(p)" :src="projectAvatar(p)" class="proj-avatar" alt="" />
              <span v-else class="proj-dot" :style="{ background: p.color || '#6366f1' }"></span>
              <span>{{ p.name }}</span>
            </RouterLink>
          </div>
        </div>
      </section>

      <!-- Members section (visible to admins and customer-admins) -->
      <section v-if="canManage" class="members-section">
        <div class="section-header-row">
          <h2>{{ $t('customer.members') }}</h2>
          <button class="btn btn-primary btn-sm" @click="openAddMember">+ {{ $t('customer.add_member') }}</button>
        </div>
        <div v-if="members.length === 0" class="empty-state-sm" style="padding:16px 0">
          {{ $t('customer.no_members') }}
        </div>
        <div v-else class="members-list">
          <div v-for="m in members" :key="m.user_id" class="member-row">
            <img :src="resolveAssetUrl(m.avatar_url || m.gravatar_url)" class="member-avatar" alt="" />
            <div class="member-info">
              <span class="member-name">{{ m.display_name || m.username }}</span>
              <span class="member-email">{{ m.email }}</span>
            </div>
            <span :class="['role-badge', m.role === 'admin' ? 'role-admin' : 'role-member']">
              {{ m.role === 'admin' ? $t('customer.role_admin') : $t('customer.role_member') }}
            </span>
            <div class="member-actions">
              <button
                v-if="m.role === 'member'"
                class="btn btn-sm"
                @click="setMemberRole(m.user_id, 'admin')"
                :aria-label="$t('customer.promote')"
                :title="$t('customer.promote')"
              >↑</button>
              <button
                v-else-if="auth.isAdmin || m.user_id !== authUserId"
                class="btn btn-sm"
                @click="setMemberRole(m.user_id, 'member')"
                :aria-label="$t('customer.demote')"
                :title="$t('customer.demote')"
              >↓</button>
              <button
                v-if="auth.isAdmin || m.user_id !== authUserId"
                class="icon-btn icon-danger"
                @click="removeMember(m.user_id)"
                :aria-label="$t('common.delete') + ' ' + (m.display_name || m.username)"
                title="Remove"
              >✕</button>
            </div>
          </div>
        </div>

        <!-- Groups with access -->
        <div style="margin-top:20px">
          <h3 style="font-size:13px;font-weight:600;color:var(--color-text-muted);text-transform:uppercase;letter-spacing:.04em;margin-bottom:10px">
            {{ $t('groups.groups_with_access') }}
          </h3>
          <div v-if="!customerGroups.length" style="color:var(--color-text-muted);font-size:13px;margin-bottom:10px">{{ $t('groups.no_group_access') }}</div>
          <div v-for="g in customerGroups" :key="g.group_id" class="member-row" style="align-items:center">
            <img v-if="groupAvatar(g.group)" :src="groupAvatar(g.group)" class="group-avatar" alt="" />
            <div class="member-info">
              <span class="member-name">{{ g.group.name }}</span>
              <span class="member-email">{{ g.members.map(m => m.user.display_name || m.user.username).join(', ') || '—' }}</span>
            </div>
            <span :class="['role-badge', g.role === 'admin' ? 'role-admin' : 'role-member']">{{ g.role }}</span>
            <button v-if="auth.isAdmin" class="icon-btn icon-danger" style="margin-left:8px" @click="removeGroupFromCustomer(g.group_id)" :aria-label="$t('common.delete') + ' ' + g.group.name" title="Remove">✕</button>
          </div>
          <div v-if="auth.isAdmin" style="display:flex;gap:8px;align-items:center;flex-wrap:wrap;margin-top:10px">
            <select class="form-input" v-model="addGroupId" style="flex:1;min-width:160px">
              <option value="">— {{ $t('groups.add_group') }} —</option>
              <option v-for="g in groupsNotOnCustomer" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
            <select class="form-input" v-model="addGroupRole" style="width:110px">
              <option value="member">{{ $t('customer.role_member') }}</option>
              <option value="admin">{{ $t('customer.role_admin') }}</option>
            </select>
            <button class="btn btn-primary btn-sm" :disabled="!addGroupId" @click="addGroupToCustomer">{{ $t('common.add') }}</button>
          </div>
        </div>
      </section>

      </div><!-- end tab-panel-overview -->

      <!-- Invoices tab -->
      <div id="tab-panel-invoices" role="tabpanel" aria-labelledby="tab-btn-invoices" v-show="activeTab === 'invoices'">

      <!-- Invoices section -->
      <section class="invoices-section">
        <div class="section-header-row">
          <h2>{{ $t('invoice.invoices') }}</h2>
          <button v-if="canManage" class="btn btn-primary btn-sm" @click="openNewInvoice">
            + {{ $t('invoice.new_invoice') }}
          </button>
        </div>

        <div v-if="invoicesLoading" class="empty-state-sm">{{ $t('common.loading') }}</div>
        <div v-else-if="invoices.length === 0" class="empty-state-sm">{{ $t('invoice.no_invoices') }}</div>
        <div v-else class="invoice-list">
          <div v-for="inv in invoices" :key="inv.id" :class="['invoice-row', { 'invoice-row--overdue': isInvoiceOverdue(inv) }]">
            <div class="invoice-row-main">
              <span class="invoice-number">{{ inv.invoice_number }}</span>
              <span v-if="inv.credited_invoice_id" class="invoice-credits">↩ {{ $t('invoice.credits_invoice') }} #{{ inv.credited_invoice_id }}</span>
              <span v-if="creditNoteByOriginal[inv.id]" class="invoice-credit-issued">↩ {{ $t('invoice.credit_note_issued') }}: {{ creditNoteByOriginal[inv.id].invoice_number }}</span>
              <span class="invoice-period">{{ formatDate(inv.period_start) }} – {{ formatDate(inv.period_end) }}</span>
              <span v-if="isInvoiceOverdue(inv)" class="invoice-status inv-overdue">{{ $t('invoice.overdue') }}</span>
              <span v-else :class="['invoice-status', `inv-${inv.status}`]">{{ $t(`invoice.status_${inv.status}`) }}</span>
              <span class="invoice-total">{{ inv.currency }} {{ inv.total.toFixed(2) }}</span>
              <span v-if="inv.due_date" :class="['invoice-due', { 'invoice-due--overdue': isInvoiceOverdue(inv) }]">{{ $t('invoice.due_date') }}: {{ formatDate(inv.due_date) }}</span>
              <span v-if="inv.status === 'paid' && inv.payment_date" class="invoice-paid-info">
                {{ $t('invoice.payment_info') }}: {{ formatDate(inv.payment_date) }}<template v-if="inv.payment_reference"> · {{ inv.payment_reference }}</template><template v-if="inv.payment_method"> · {{ $t(`invoice.payment_method_${inv.payment_method}`) }}</template>
              </span>
            </div>
            <div class="invoice-row-actions">
              <a :href="invoicePdfUrl(inv.id)" target="_blank" class="btn btn-sm" :aria-label="$t('invoice.download_pdf')">PDF</a>
              <template v-if="canManage">
                <button
                  v-if="inv.status !== 'credit_note'"
                  class="icon-btn"
                  @click="openEditLines(inv)"
                  :aria-label="$t('invoice.edit_lines')"
                  :title="$t('invoice.edit_lines')"
                >✎</button>
                <button
                  v-if="inv.status === 'draft' || inv.status === 'sent'"
                  class="btn btn-sm"
                  @click="openSendEmail(inv)"
                >{{ $t('invoice.send_email') }}</button>
                <button v-if="inv.status === 'draft'" class="btn btn-sm" @click="changeInvoiceStatus(inv, 'sent')">{{ $t('invoice.mark_sent') }}</button>
                <button v-if="inv.status === 'sent'" class="btn btn-sm btn-primary" @click="openRecordPayment(inv)">{{ $t('invoice.mark_paid') }}</button>
                <button v-if="inv.status !== 'draft' && inv.status !== 'credit_note'" class="btn btn-sm" @click="changeInvoiceStatus(inv, 'draft')">{{ $t('invoice.mark_draft') }}</button>
                <button
                  v-if="inv.status === 'sent' || inv.status === 'paid'"
                  class="btn btn-sm"
                  @click="doIssueCreditNote(inv)"
                >{{ $t('invoice.issue_credit_note') }}</button>
                <button v-if="inv.status === 'draft'" class="icon-btn icon-danger" @click="doDeleteInvoice(inv)" :aria-label="$t('common.delete')" title="Delete">✕</button>
                <button v-if="inv.status === 'credit_note'" class="btn btn-sm btn-danger" @click="doDeleteInvoice(inv)">{{ $t('invoice.void_credit_note') }}</button>
              </template>
            </div>
          </div>
        </div>
      </section>

      </div><!-- end tab-panel-invoices -->

      <!-- Contacts tab -->
      <div id="tab-panel-contacts" role="tabpanel" aria-labelledby="tab-btn-contacts" v-show="activeTab === 'contacts'">

      <!-- Contacts section -->
      <section class="contacts-section">
        <div class="section-header-row">
          <h2>{{ $t('customer.contacts') }}</h2>
          <button v-if="canManage" class="btn btn-primary btn-sm" @click="openAddContact">+ {{ $t('customer.add_contact') }}</button>
        </div>
        <div v-if="!contacts.length" class="empty-state-sm" style="padding:16px 0">
          {{ $t('customer.no_contacts') }}
        </div>
        <div v-else class="contacts-list">
          <div v-for="ct in contacts" :key="ct.id" class="contact-row">
            <div class="contact-info">
              <span class="contact-name">{{ ct.name }}</span>
              <span v-if="ct.department" class="contact-detail">{{ ct.department }}</span>
              <span v-if="ct.phone" class="contact-detail">
                <span aria-hidden="true">📞</span> {{ ct.phone }}
              </span>
              <span v-if="ct.email" class="contact-detail">
                <span aria-hidden="true">✉</span> {{ ct.email }}
              </span>
            </div>
            <span v-if="ct.is_primary" class="badge-primary-contact">{{ $t('customer.contact_primary') }}</span>
            <div v-if="canManage" class="contact-actions">
              <button class="btn btn-sm" @click="openEditContact(ct)" :aria-label="$t('customer.edit_contact') + ': ' + ct.name">✎</button>
              <button class="icon-btn icon-danger" @click="removeContact(ct)" :aria-label="$t('common.delete') + ' ' + ct.name">✕</button>
            </div>
          </div>
        </div>
      </section>

      </div><!-- end tab-panel-contacts -->

    </template>

    <!-- Add / edit contact modal -->
    <BaseModal
      v-if="showContactModal"
      :title="editingContact ? $t('customer.edit_contact') : $t('customer.add_contact')"
      @close="showContactModal = false"
    >
      <div class="form-group">
        <label class="form-label" for="ct-name">{{ $t('customer.contact_name') }}</label>
        <input id="ct-name" class="form-input" type="text" v-model="contactForm.name" required />
      </div>
      <div class="form-group">
        <label class="form-label" for="ct-dept">{{ $t('customer.contact_department') }}</label>
        <input id="ct-dept" class="form-input" type="text" v-model="contactForm.department" />
      </div>
      <div class="form-group">
        <label class="form-label" for="ct-phone">{{ $t('customer.contact_phone') }}</label>
        <input id="ct-phone" class="form-input" type="tel" v-model="contactForm.phone" />
      </div>
      <div class="form-group">
        <label class="form-label" for="ct-email">{{ $t('customer.contact_email') }}</label>
        <input id="ct-email" class="form-input" type="email" v-model="contactForm.email" />
      </div>
      <div class="form-group form-check-row">
        <input id="ct-primary" type="checkbox" v-model="contactForm.is_primary" />
        <label for="ct-primary">{{ $t('customer.contact_is_primary') }}</label>
      </div>
      <template #footer>
        <button class="btn" @click="showContactModal = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!contactForm.name.trim()" @click="saveContact">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>

    <!-- Edit customer modal -->
    <BaseModal v-if="showEdit" :title="$t('customer.edit')" @close="showEdit = false">
      <div class="form-group">
        <label class="form-label" for="edit-cust-name">{{ $t('customer.name') }}</label>
        <input id="edit-cust-name" class="form-input" v-model="editForm.name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-cust-desc">{{ $t('customer.description') }}</label>
        <textarea id="edit-cust-desc" class="form-input" v-model="editForm.description" rows="3"></textarea>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-cust-logo">{{ $t('customer.logo_url') }}</label>
        <div style="display:flex;gap:8px;align-items:center">
          <input id="edit-cust-logo" class="form-input" v-model="editForm.logo_url" placeholder="https://..." style="flex:1" />
          <button type="button" class="btn btn-secondary btn-sm" @click="$refs.logoFileInput.click()">{{ $t('customer.upload_logo') }}</button>
        </div>
        <input ref="logoFileInput" type="file" accept="image/*" style="display:none" @change="onLogoFileSelected" />
      </div>
      <h4 class="form-section-title">{{ $t('customer.billing_section') }}</h4>
      <div class="form-group">
        <label class="form-label" for="edit-cust-street">{{ $t('customer.billing_street') }}</label>
        <input id="edit-cust-street" class="form-input" v-model="editForm.billing_street" />
      </div>
      <div class="form-group form-row">
        <div class="form-group half">
          <label class="form-label" for="edit-cust-postal">{{ $t('customer.billing_postal_code') }}</label>
          <input id="edit-cust-postal" class="form-input" v-model="editForm.billing_postal_code" />
        </div>
        <div class="form-group half">
          <label class="form-label" for="edit-cust-city">{{ $t('customer.billing_city') }}</label>
          <input id="edit-cust-city" class="form-input" v-model="editForm.billing_city" />
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="edit-cust-country">{{ $t('customer.billing_country') }}</label>
        <input id="edit-cust-country" class="form-input" v-model="editForm.billing_country" />
      </div>
      <div class="form-group form-row">
        <div class="form-group half">
          <label class="form-label" for="edit-cust-vat">{{ $t('customer.vat_number') }}</label>
          <input id="edit-cust-vat" class="form-input" v-model="editForm.vat_number" />
        </div>
        <div class="form-group half">
          <label class="form-label" for="edit-cust-po">{{ $t('customer.po_reference') }}</label>
          <input id="edit-cust-po" class="form-input" v-model="editForm.po_reference" />
        </div>
      </div>
      <template #footer>
        <button class="btn" @click="showEdit = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveEdit">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>

    <!-- Add / edit contract modal -->
    <BaseModal
      v-if="showAddContract || editingContract"
      :title="editingContract ? $t('contract.edit') : $t('contract.new_contract')"
      :resizable="true"
      style="--modal-width: 720px"
      @close="closeContractModal"
    >
      <div class="form-group">
        <label class="form-label" for="contract-name">{{ $t('contract.name') }}</label>
        <input id="contract-name" class="form-input" v-model="contractForm.name" />
      </div>
      <div class="form-group">
        <label class="form-label" for="contract-desc">{{ $t('contract.description') }}</label>
        <textarea id="contract-desc" class="form-input" v-model="contractForm.description" rows="2"></textarea>
      </div>
      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label" for="contract-start">{{ $t('contract.start_date') }}</label>
          <div class="date-input-row">
            <input id="contract-start" class="form-input" type="text" v-model="displayContractStartDate" :placeholder="dateOnlyFormat()" @blur="parseContractStartDate" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="contractForm.start_date" @change="onContractStartDateChange" />
            </label>
            <button v-if="displayContractStartDate" class="btn-icon-xs" @click="displayContractStartDate = ''; contractForm.start_date = ''" title="Clear">×</button>
          </div>
        </div>
        <div class="form-group half">
          <label class="form-label" for="contract-end">{{ $t('contract.end_date') }}</label>
          <div class="date-input-row">
            <input id="contract-end" class="form-input" type="text" v-model="displayContractEndDate" :placeholder="dateOnlyFormat()" @blur="parseContractEndDate" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="contractForm.end_date" @change="onContractEndDateChange" />
            </label>
            <button v-if="displayContractEndDate" class="btn-icon-xs" @click="displayContractEndDate = ''; contractForm.end_date = ''" title="Clear">×</button>
          </div>
        </div>
      </div>
      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label" for="contract-rate">{{ $t('contract.price_per_hour') }}</label>
          <input id="contract-rate" class="form-input" type="number" min="0" step="0.01" v-model="contractForm.price_per_hour" />
        </div>
        <div class="form-group half">
          <label class="form-label" for="contract-rate-km">{{ $t('contract.price_per_km', { unit: distanceUnit }) }}</label>
          <input id="contract-rate-km" class="form-input" type="number" min="0" step="0.01" v-model="contractForm.price_per_km" />
        </div>
      </div>
      <div class="detail-row">
        <div class="form-group half">
          <label class="form-label" for="contract-currency">{{ $t('contract.currency') }}</label>
          <select id="contract-currency" class="form-input" v-model="contractForm.currency">
            <option value="€">€ (EUR)</option>
            <option value="USD">USD ($)</option>
            <option value="GBP">GBP (£)</option>
            <option value="CHF">CHF</option>
            <option value="SEK">SEK (kr)</option>
            <option value="NOK">NOK (kr)</option>
            <option value="DKK">DKK (kr)</option>
            <option value="PLN">PLN (zł)</option>
            <option value="CZK">CZK (Kč)</option>
          </select>
        </div>
      </div>
      <div class="form-group">
        <div class="slots-header">
          <label class="form-label">{{ $t('contract.time_slots') }}</label>
          <button type="button" class="btn btn-sm" @click="addTimeSlot">+ {{ $t('contract.add_time_slot') }}</button>
        </div>
        <p class="form-hint">{{ $t('contract.slot_overnight_hint') }}</p>
        <div class="slots-list">
          <div v-for="(slot, idx) in contractForm.time_slots" :key="idx" class="slot-card">
            <div class="slot-card-top">
              <input class="form-input" type="text" v-model="slot.label" :aria-label="$t('contract.slot_label')" :placeholder="$t('contract.slot_label')" />
              <button type="button" class="btn-icon-xs slot-remove" @click="removeTimeSlot(idx)" :aria-label="$t('contract.remove_slot')">✕</button>
            </div>
            <div class="slot-card-bottom">
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_start') }}</label>
                <input class="form-input" type="text"
                  :value="formatSlotTime(slot.start_time)"
                  :placeholder="slotTimePlaceholder"
                  :aria-label="$t('contract.slot_start')"
                  @input="slot.start_time = parseSlotTime($event.target.value) ?? slot.start_time"
                  @blur="onSlotTimeBlur(slot, 'start_time', $event)" />
              </div>
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_end') }}</label>
                <input class="form-input" type="text"
                  :value="formatSlotTime(slot.end_time)"
                  :placeholder="slotTimePlaceholder"
                  :aria-label="$t('contract.slot_end')"
                  @input="slot.end_time = parseSlotTime($event.target.value) ?? slot.end_time"
                  @blur="onSlotTimeBlur(slot, 'end_time', $event)" />
              </div>
              <div class="slot-field slot-field-days">
                <label class="slot-field-label">{{ $t('contract.slot_days') }} <HelpIcon i18n-key="help.fields.slot_day_type" align="start" /></label>
                <div class="slot-day-checks" role="group" :aria-label="$t('contract.slot_days')">
                  <label
                    v-for="(abbr, idx) in slotDowAbbrevs"
                    :key="idx"
                    class="slot-day-check"
                    :class="{ active: isSlotDayChecked(slot, idx) }"
                  >
                    <input
                      type="checkbox"
                      class="sr-only"
                      :checked="isSlotDayChecked(slot, idx)"
                      @change="toggleSlotDay(slot, idx)"
                      :aria-label="slotDowFullNames[idx]"
                    />
                    {{ abbr }}
                  </label>
                </div>
              </div>
              <div v-if="isOvernightSlot(slot)" class="slot-field slot-field-days">
                <label class="slot-field-label">{{ $t('contract.slot_end_day_offset') }} <HelpIcon i18n-key="help.fields.slot_end_day" align="start" /></label>
                <select class="form-input" v-model.number="slot.end_day_offset" :aria-label="$t('contract.slot_end_day_offset')">
                  <option v-for="n in 6" :key="n" :value="n">{{ $t('contract.slot_end_day_offset_' + n) }}</option>
                </select>
              </div>
              <div v-if="slotDurationMinutes(slot) !== null" class="slot-field slot-field-duration">
                <label class="slot-field-label">{{ $t('contract.slot_duration') }}</label>
                <div class="slot-duration-summary">
                  <template v-if="getCheckedDayIndices(slot.day_type).length > 1">
                    <span v-for="i in getCheckedDayIndices(slot.day_type)" :key="i" class="slot-dur-item">
                      {{ slotDowAbbrevs[i] }}: {{ formatMinutes(slotDurationMinutes(slot)) }}
                    </span>
                    <span class="slot-dur-total">= {{ formatMinutes(getCheckedDayIndices(slot.day_type).length * slotDurationMinutes(slot)) }}</span>
                  </template>
                  <template v-else>
                    <span class="slot-dur-single">{{ formatMinutes(slotDurationMinutes(slot)) }}</span>
                  </template>
                </div>
              </div>
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_factor') }} <HelpIcon i18n-key="help.fields.slot_factor" align="start" /></label>
                <input class="form-input" type="number" min="0" step="0.01" v-model="slot.multiplication_factor" :aria-label="$t('contract.slot_factor')" :placeholder="'×'" />
              </div>
              <div class="slot-field">
                <label class="slot-field-label">{{ $t('contract.slot_rate') }} <HelpIcon i18n-key="help.fields.slot_rate" align="start" /></label>
                <input class="form-input" type="number" min="0" step="0.01" v-model="slot.hourly_rate" :aria-label="$t('contract.slot_rate')" :placeholder="contractForm.currency + '/h'" />
              </div>
            </div>
            <div v-if="slotPreviewReady(slot)" class="slot-preview">
              <span class="slot-preview-label">{{ $t('contract.slot_preview') }}</span>
              <div class="slot-preview-week" role="img" :aria-label="slotPreviewAria(slot)">
                <div
                  v-for="day in slotPreviewDays(slot)"
                  :key="day.key"
                  class="slot-preview-day"
                  :class="{ 'slot-preview-day-active': day.active }"
                >
                  <span class="slot-preview-dow">{{ day.label }}</span>
                  <div class="slot-preview-track" aria-hidden="true">
                    <div
                      v-for="(seg, segIdx) in day.segments"
                      :key="segIdx"
                      class="slot-preview-seg"
                      :style="{ left: seg.left + '%', width: seg.width + '%', background: slotSegColor(idx) }"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="btn" @click="closeContractModal">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveContract" :disabled="!contractForm.name.trim()">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>
  <!-- Add member modal -->
  <BaseModal v-if="showAddMember" :title="$t('customer.add_member')" @close="showAddMember = false">
    <div class="form-group">
      <input class="form-input" v-model="memberSearch" :aria-label="$t('common.search')" :placeholder="$t('common.search') + '…'" />
    </div>
    <div class="user-picker-list">
      <div
        v-for="u in filteredUsers"
        :key="u.id"
        class="user-picker-row"
        :class="{ selected: pendingMemberIds.includes(u.id) }"
        @click="togglePendingMember(u.id)"
      >
        <img :src="resolveAssetUrl(u.avatar_url || u.gravatar_url)" class="member-avatar" alt="" />
        <div class="member-info">
          <span class="member-name">{{ u.display_name || u.username }}</span>
          <span class="member-email">{{ u.email }}</span>
        </div>
        <span v-if="pendingMemberIds.includes(u.id)" class="check-mark">✓</span>
      </div>
      <div v-if="filteredUsers.length === 0" class="empty-state-sm" style="padding:8px 0">
        {{ $t('common.no_results') }}
      </div>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="showAddMember = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="pendingMemberIds.length === 0" @click="confirmAddMembers">{{ $t('common.add') }}</button>
    </template>
  </BaseModal>

  <!-- Create Invoice modal -->
  <BaseModal v-if="showNewInvoice" :title="$t('invoice.create_from_entries')" @close="closeNewInvoice" style="--modal-width:700px">
    <!-- Step 1: period selection -->
    <template v-if="invoiceStep === 1">
      <div v-if="invoiceTemplates.length > 0" class="form-group">
        <label class="form-label" for="inv-tmpl">{{ $t('invoice.load_template') }}</label>
        <select id="inv-tmpl" class="form-input" @change="applyInvoiceTemplate($event.target.value); $event.target.value = ''">
          <option value="">— {{ $t('invoice.load_template') }} —</option>
          <option v-for="tmpl in invoiceTemplates" :key="tmpl.id" :value="tmpl.id">{{ tmpl.name }}</option>
        </select>
      </div>
      <div class="form-group form-row">
        <div class="form-group half">
          <label class="form-label" for="inv-from">{{ $t('invoice.period_start') }}</label>
          <div class="date-input-row">
            <input id="inv-from" class="form-input" type="text" v-model="displayInvFrom" :placeholder="dateOnlyFormat()" @blur="parseInvFrom" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="invoiceForm.period_start" @change="e => { invoiceForm.period_start = e.target.value; displayInvFrom = e.target.value ? formatDate(e.target.value) : '' }" />
            </label>
            <button v-if="displayInvFrom" class="btn-icon-xs" @click="displayInvFrom = ''; invoiceForm.period_start = ''" title="Clear">×</button>
          </div>
        </div>
        <div class="form-group half">
          <label class="form-label" for="inv-to">{{ $t('invoice.period_end') }}</label>
          <div class="date-input-row">
            <input id="inv-to" class="form-input" type="text" v-model="displayInvTo" :placeholder="dateOnlyFormat()" @blur="parseInvTo" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="invoiceForm.period_end" @change="e => { invoiceForm.period_end = e.target.value; displayInvTo = e.target.value ? formatDate(e.target.value) : '' }" />
            </label>
            <button v-if="displayInvTo" class="btn-icon-xs" @click="displayInvTo = ''; invoiceForm.period_end = ''" title="Clear">×</button>
          </div>
        </div>
      </div>
      <div class="form-group form-row">
        <div class="form-group half">
          <label class="form-label" for="inv-due">{{ $t('invoice.due_date') }}</label>
          <div class="date-input-row">
            <input id="inv-due" class="form-input" type="text" v-model="displayInvDue" :placeholder="dateOnlyFormat()" @blur="parseInvDue" />
            <label class="picker-wrap" :title="$t('common.pick_date')">
              <span class="btn-icon-xs">&#128197;</span>
              <input type="date" class="date-picker-overlay" :value="invoiceForm.due_date" @change="e => { invoiceForm.due_date = e.target.value; displayInvDue = e.target.value ? formatDate(e.target.value) : '' }" />
            </label>
            <button v-if="displayInvDue" class="btn-icon-xs" @click="displayInvDue = ''; invoiceForm.due_date = ''" title="Clear">×</button>
          </div>
        </div>
        <div class="form-group half">
          <label class="form-label" for="inv-vat">{{ $t('invoice.vat_rate') }}</label>
          <input id="inv-vat" class="form-input" type="number" min="0" max="100" step="0.1" v-model.number="invoiceForm.vat_rate" />
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="inv-notes">{{ $t('invoice.notes') }}</label>
        <textarea id="inv-notes" class="form-input" rows="2" v-model="invoiceForm.notes"></textarea>
      </div>
    </template>

    <!-- Step 2: line item preview -->
    <template v-else-if="invoiceStep === 2">
      <div v-if="invoiceLineItems.length === 0" class="empty-state-sm" style="padding:16px 0">
        {{ $t('invoice.no_billable_entries') }}
      </div>
      <div v-else class="inv-preview-table-wrap">
        <table class="inv-preview-table">
          <thead>
            <tr>
              <th>{{ $t('invoice.line_date') }}</th>
              <th>{{ $t('invoice.line_project') }}</th>
              <th>{{ $t('invoice.line_description') }}</th>
              <th class="num">{{ $t('invoice.line_hours') }}</th>
              <th v-if="invoiceHasDistance" class="num">{{ $t('invoice.line_distance') }}</th>
              <th class="num">{{ $t('invoice.line_rate') }}</th>
              <th class="num">{{ $t('invoice.line_amount') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(li, i) in invoiceLineItems" :key="i">
              <td>{{ li.date ? formatDate(li.date) : '' }}</td>
              <td>{{ li.project_name }}</td>
              <td>{{ li.description }}</td>
              <td class="num">{{ fmtMinutes(li.minutes) }}</td>
              <td v-if="invoiceHasDistance" class="num">{{ li.distance > 0 ? li.distance.toFixed(1) : '' }}</td>
              <td class="num">{{ li.hourly_rate > 0 ? li.hourly_rate.toFixed(2) : '' }}</td>
              <td class="num">{{ li.currency }} {{ li.amount.toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="inv-preview-totals">
        <span>{{ $t('invoice.subtotal') }}: <strong>{{ invoicePreviewCurrency }} {{ invoicePreviewSubtotal.toFixed(2) }}</strong></span>
        <span v-if="invoiceForm.vat_rate > 0">VAT {{ invoiceForm.vat_rate }}%: <strong>{{ invoicePreviewCurrency }} {{ invoicePreviewVAT.toFixed(2) }}</strong></span>
        <span>{{ $t('invoice.total') }}: <strong>{{ invoicePreviewCurrency }} {{ invoicePreviewTotal.toFixed(2) }}</strong></span>
      </div>
    </template>

    <template #footer>
      <button class="btn" @click="closeNewInvoice">{{ $t('common.cancel') }}</button>
      <button v-if="invoiceStep === 1" class="btn btn-primary" :disabled="!invoiceForm.period_start || !invoiceForm.period_end" @click="previewInvoice">
        {{ $t('invoice.preview_items') }} →
      </button>
      <template v-if="invoiceStep === 2">
        <button class="btn btn-secondary" @click="invoiceStep = 1">← {{ $t('common.go_back') }}</button>
        <button class="btn btn-primary" :disabled="invoiceLineItems.length === 0 || savingInvoice" @click="saveInvoice">
          {{ $t('invoice.confirm_create') }}
        </button>
      </template>
    </template>
  </BaseModal>

  <!-- Edit line items modal -->
  <BaseModal
    v-if="showEditLines"
    :title="$t('invoice.edit_lines')"
    @close="showEditLines = false"
    style="--modal-width:760px"
    aria-labelledby="edit-lines-title"
  >
    <h2 id="edit-lines-title" class="sr-only">{{ $t('invoice.edit_lines') }}</h2>
    <div class="inv-edit-table-wrap">
      <table class="inv-preview-table">
        <thead>
          <tr>
            <th>{{ $t('invoice.line_date') }}</th>
            <th>{{ $t('invoice.line_description') }}</th>
            <th class="num">{{ $t('invoice.manual_line_qty') }}</th>
            <th class="num">{{ $t('invoice.manual_line_price') }}</th>
            <th class="num">{{ $t('invoice.line_amount') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(li, idx) in editLineItems" :key="idx">
            <template v-if="li.is_comment">
              <td colspan="5" style="padding:2px 4px">
                <input
                  class="form-input form-input-sm"
                  style="width:100%;font-style:italic;color:var(--color-text-muted)"
                  type="text"
                  v-model="li.description"
                  :placeholder="$t('invoice.comment_placeholder')"
                  :aria-label="$t('invoice.add_comment_line')"
                />
              </td>
            </template>
            <template v-else>
              <td><input class="form-input form-input-sm" type="text" v-model="li.date" :aria-label="$t('invoice.line_date')" /></td>
              <td><input class="form-input form-input-sm" type="text" v-model="li.description" :aria-label="$t('invoice.line_description')" /></td>
              <td class="num">
                <input
                  v-if="li.is_manual"
                  class="form-input form-input-sm num-input"
                  type="number" min="0" step="0.01"
                  v-model.number="li.quantity"
                  @input="li.amount = +(li.quantity * li.unit_price).toFixed(2)"
                  :aria-label="$t('invoice.manual_line_qty')"
                />
                <span v-else>{{ li.minutes > 0 ? fmtMinutes(li.minutes) : '' }}</span>
              </td>
              <td class="num">
                <input
                  v-if="li.is_manual"
                  class="form-input form-input-sm num-input"
                  type="number" min="0" step="0.01"
                  v-model.number="li.unit_price"
                  @input="li.amount = +(li.quantity * li.unit_price).toFixed(2)"
                  :aria-label="$t('invoice.manual_line_price')"
                />
                <span v-else>{{ li.hourly_rate > 0 ? li.hourly_rate.toFixed(2) : (li.price_per_km > 0 ? li.price_per_km.toFixed(2) : '') }}</span>
              </td>
              <td class="num">
                <input
                  class="form-input form-input-sm num-input"
                  type="number" step="0.01"
                  v-model.number="li.amount"
                  :aria-label="$t('invoice.line_amount')"
                />
              </td>
            </template>
            <td>
              <button class="icon-btn icon-danger" @click="editLineItems.splice(idx,1)" :aria-label="$t('common.delete')">✕</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div style="margin-top:8px;display:flex;gap:8px">
      <button class="btn btn-sm" @click="addManualLine">+ {{ $t('invoice.add_manual_line') }}</button>
      <button class="btn btn-sm" @click="addCommentLine">+ {{ $t('invoice.add_comment_line') }}</button>
    </div>
    <div class="inv-preview-totals" style="margin-top:8px">
      <span>{{ $t('invoice.total') }}: <strong>{{ editingInvoice?.currency }} {{ editLinesTotal.toFixed(2) }}</strong></span>
    </div>
    <template #footer>
      <button class="btn" @click="showEditLines = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="saveEditLines">{{ $t('common.save') }}</button>
    </template>
  </BaseModal>

  <!-- Send by email modal -->
  <BaseModal
    v-if="showSendEmail"
    :title="$t('invoice.send_email')"
    @close="showSendEmail = false"
    style="--modal-width:560px"
    aria-labelledby="send-email-title"
  >
    <h2 id="send-email-title" class="sr-only">{{ $t('invoice.send_email') }}</h2>
    <div class="form-group">
      <label class="form-label" for="send-to">{{ $t('invoice.send_email_to') }}</label>
      <input id="send-to" class="form-input" type="email" v-model="sendEmailForm.to" />
    </div>
    <div class="form-group">
      <label class="form-label" for="send-subject">{{ $t('invoice.send_email_subject') }}</label>
      <input id="send-subject" class="form-input" type="text" v-model="sendEmailForm.subject" />
    </div>
    <div class="form-group">
      <div class="email-body-header">
        <label class="form-label">{{ $t('invoice.send_email_body') }}</label>
        <div class="email-body-tabs" role="tablist">
          <button
            role="tab"
            :aria-selected="!sendEmailPreview"
            :class="['email-tab', { active: !sendEmailPreview }]"
            @click="sendEmailPreview = false"
          >{{ $t('common.write') }}</button>
          <button
            role="tab"
            :aria-selected="sendEmailPreview"
            :class="['email-tab', { active: sendEmailPreview }]"
            @click="sendEmailPreview = true"
          >{{ $t('common.preview') }}</button>
        </div>
      </div>
      <textarea
        v-if="!sendEmailPreview"
        id="send-body"
        class="form-input email-body-textarea"
        rows="8"
        v-model="sendEmailForm.body"
        :placeholder="$t('invoice.send_email_md_hint')"
      ></textarea>
      <div
        v-else
        class="email-body-preview markdown-body"
        v-html="sendEmailRendered"
        aria-live="polite"
      ></div>
    </div>
    <template #footer>
      <button class="btn" @click="showSendEmail = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" :disabled="!sendEmailForm.to || sendingEmail" @click="doSendEmail">
        {{ sendingEmail ? $t('invoice.send_email_sending') : $t('invoice.send_email') }}
      </button>
    </template>
  </BaseModal>

  <!-- Record payment modal -->
  <BaseModal
    v-if="showRecordPayment"
    :title="$t('invoice.record_payment')"
    @close="showRecordPayment = false"
    style="--modal-width:440px"
    aria-labelledby="record-payment-title"
  >
    <h2 id="record-payment-title" class="sr-only">{{ $t('invoice.record_payment') }}</h2>
    <div class="form-group">
      <label class="form-label" for="pay-date">{{ $t('invoice.payment_date') }}</label>
      <div class="date-input-row">
        <input id="pay-date" class="form-input" type="text" v-model="paymentForm.display_payment_date" :placeholder="dateOnlyFormat()" @blur="parsePaymentDate" />
        <label class="picker-wrap" :title="$t('common.pick_date')">
          <span class="btn-icon-xs">&#128197;</span>
          <input type="date" class="date-picker-overlay" :value="paymentForm.payment_date" @change="e => { paymentForm.payment_date = e.target.value; paymentForm.display_payment_date = e.target.value ? formatDate(e.target.value) : '' }" />
        </label>
        <button v-if="paymentForm.display_payment_date" class="btn-icon-xs" @click="paymentForm.display_payment_date = ''; paymentForm.payment_date = ''" title="Clear">×</button>
      </div>
    </div>
    <div class="form-group">
      <label class="form-label" for="pay-amount">{{ $t('invoice.payment_amount') }}</label>
      <input id="pay-amount" class="form-input" type="number" step="0.01" v-model.number="paymentForm.payment_amount" />
    </div>
    <div class="form-group">
      <label class="form-label" for="pay-ref">{{ $t('invoice.payment_reference') }}</label>
      <input id="pay-ref" class="form-input" type="text" v-model="paymentForm.payment_reference" />
    </div>
    <div class="form-group">
      <label class="form-label" for="pay-method">{{ $t('invoice.payment_method') }}</label>
      <select id="pay-method" class="form-input" v-model="paymentForm.payment_method">
        <option value="">—</option>
        <option value="bank">{{ $t('invoice.payment_method_bank') }}</option>
        <option value="card">{{ $t('invoice.payment_method_card') }}</option>
        <option value="cash">{{ $t('invoice.payment_method_cash') }}</option>
        <option value="other">{{ $t('invoice.payment_method_other') }}</option>
      </select>
    </div>
    <template #footer>
      <button class="btn" @click="showRecordPayment = false">{{ $t('common.cancel') }}</button>
      <button class="btn btn-primary" @click="doRecordPayment">{{ $t('invoice.mark_paid') }}</button>
    </template>
  </BaseModal>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCustomersStore } from '@/stores/customers'
import { useUIStore } from '@/stores/ui'
import { customersApi } from '@/api/customers'
import { timeEntriesApi } from '@/api/timeEntries'
import { groupsApi } from '@/api/groups'
import { attachmentsApi } from '@/api/attachments'
import { resolveAssetUrl, getServerUrl } from '@/api/serverConfig'
import BaseModal from '@/components/common/BaseModal.vue'
import HelpIcon from '@/components/common/HelpIcon.vue'
import { useDateFormat } from '@/composables/useDateFormat'
import client from '@/api/client'
import { buildSlotPreviewDays, slotPreviewReady as slotPreviewReadyFn, formatDayType } from '@/utils/contractSlotPreview'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const custStore = useCustomersStore()
const ui = useUIStore()

const distanceUnit = computed(() => auth.user?.distance_unit || 'km')

const loading = ref(true)
const detail = ref(null)
const VALID_TABS = ['overview', 'invoices', 'contacts']
const activeTab = ref(VALID_TABS.includes(route.query.tab) ? route.query.tab : 'overview')

const showEdit = ref(false)
const editForm = ref({ name: '', description: '', logo_url: '', billing_street: '', billing_city: '', billing_postal_code: '', billing_country: '', vat_number: '', po_reference: '' })

const { formatDate, dateOnlyFormat } = useDateFormat()

// ── Invoices ──────────────────────────────────────────────────────────────────
const invoices = ref([])
const invoicesLoading = ref(false)
const showNewInvoice = ref(false)
const invoiceStep = ref(1)
const savingInvoice = ref(false)
const invoiceLineItems = ref([])
const invoiceForm = ref({ period_start: '', period_end: '', due_date: '', vat_rate: 0, notes: '' })

const displayInvFrom = ref('')
const displayInvTo   = ref('')
const displayInvDue  = ref('')

function _parseInvDate(displayRef, isoKey) {
  const val = displayRef.value.trim()
  if (!val) { invoiceForm.value[isoKey] = ''; return }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY'), mPos = fmt.indexOf('MM'), dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayRef.value = invoiceForm.value[isoKey] ? formatDate(invoiceForm.value[isoKey]) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  invoiceForm.value[isoKey] = iso
  displayRef.value = formatDate(iso)
}
function parseInvFrom() { _parseInvDate(displayInvFrom, 'period_start') }
function parseInvTo()   { _parseInvDate(displayInvTo,   'period_end') }
function parseInvDue()  { _parseInvDate(displayInvDue,  'due_date') }

// Map original invoice id → credit note, so we can show the tag on the original.
const creditNoteByOriginal = computed(() => {
  const map = {}
  for (const inv of invoices.value) {
    if (inv.status === 'credit_note' && inv.credited_invoice_id) {
      map[inv.credited_invoice_id] = inv
    }
  }
  return map
})

const invoiceHasDistance = computed(() => invoiceLineItems.value.some(li => li.distance > 0))
const invoicePreviewSubtotal = computed(() => invoiceLineItems.value.reduce((s, li) => s + li.amount, 0))
const invoicePreviewVAT = computed(() => invoicePreviewSubtotal.value * invoiceForm.value.vat_rate / 100)
const invoicePreviewTotal = computed(() => invoicePreviewSubtotal.value + invoicePreviewVAT.value)
const invoicePreviewCurrency = computed(() => invoiceLineItems.value[0]?.currency || '€')

function invoicePdfUrl(invoiceId) {
  const server = getServerUrl()
  const base = server ? `${server}/api/v1` : '/api/v1'
  const lang = auth.user?.locale || 'en'
  const du = distanceUnit.value
  return `${base}/customers/${custId.value}/invoices/${invoiceId}/pdf?lang=${lang}&distance_unit=${du}`
}

function fmtMinutes(mins) {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

async function loadInvoices() {
  invoicesLoading.value = true
  try {
    const { data } = await customersApi.listInvoices(custId.value)
    invoices.value = data || []
  } catch {
    invoices.value = []
  } finally {
    invoicesLoading.value = false
  }
}

// Invoice templates for the new-invoice modal
const invoiceTemplates = ref([])
async function loadInvoiceTemplates() {
  try {
    const { data } = await customersApi.listInvoiceTemplates()
    invoiceTemplates.value = data || []
  } catch {
    invoiceTemplates.value = []
  }
}
function applyInvoiceTemplate(idStr) {
  const tmpl = invoiceTemplates.value.find(t => String(t.id) === String(idStr))
  if (!tmpl) return
  if (tmpl.default_vat_rate) invoiceForm.value.vat_rate = tmpl.default_vat_rate
  if (tmpl.notes) invoiceForm.value.notes = tmpl.notes
  if (tmpl.line_items) {
    try {
      const lines = JSON.parse(tmpl.line_items)
      if (lines.length) {
        invoiceLineItems.value = lines.map(li => ({
          ...li,
          currency: li.currency || tmpl.default_currency || '€',
          amount: (li.quantity || 0) * (li.unit_price || 0),
          is_manual: true,
        }))
        invoiceStep.value = 2
      }
    } catch { /* ignore */ }
  }
}

function openNewInvoice() {
  const now = new Date()
  const y = now.getFullYear(), m = now.getMonth()
  const firstDay = new Date(y, m, 1)
  const lastDay = new Date(y, m + 1, 0)
  const fromIso = firstDay.toISOString().slice(0, 10)
  const toIso   = lastDay.toISOString().slice(0, 10)
  invoiceForm.value = { period_start: fromIso, period_end: toIso, due_date: '', vat_rate: 0, notes: '' }
  displayInvFrom.value = formatDate(fromIso)
  displayInvTo.value   = formatDate(toIso)
  displayInvDue.value  = ''
  invoiceLineItems.value = []
  invoiceStep.value = 1
  showNewInvoice.value = true
}

function closeNewInvoice() {
  showNewInvoice.value = false
}

async function previewInvoice() {
  try {
    const { data } = await timeEntriesApi.list({
      customer_id: custId.value,
      from: invoiceForm.value.period_start,
      to:   invoiceForm.value.period_end,
      user_id: 0,  // all users (admin only; non-admins see own entries)
    })
    const entries = Array.isArray(data) ? data : []
    const items = []
    for (const entry of entries) {
      if (!entry.contract_id) continue
      const rate = entry.contract?.price_per_hour ?? 0
      const km   = entry.contract?.price_per_km ?? 0
      const hours = entry.minutes / 60
      const dist  = entry.distance ?? 0
      const amount = (hours * rate) + (dist * km)
      if (amount <= 0 && entry.minutes <= 0) continue
      items.push({
        date:         entry.date ? entry.date.slice(0, 10) : '',
        project_name: entry.project?.name || '',
        description:  entry.description || '',
        minutes:      entry.minutes,
        hourly_rate:  rate,
        distance:     dist,
        price_per_km: km,
        amount:       amount,
        currency:     entry.contract?.currency || '€',
      })
    }
    // sort by date ascending
    items.sort((a, b) => a.date.localeCompare(b.date))
    invoiceLineItems.value = items
    invoiceStep.value = 2
  } catch {
    ui.error('Failed to load time entries')
  }
}

async function saveInvoice() {
  savingInvoice.value = true
  try {
    await customersApi.createInvoice(custId.value, {
      period_start: invoiceForm.value.period_start,
      period_end:   invoiceForm.value.period_end,
      due_date:     invoiceForm.value.due_date || undefined,
      vat_rate:     invoiceForm.value.vat_rate,
      notes:        invoiceForm.value.notes,
      line_items:   invoiceLineItems.value,
      currency:     invoicePreviewCurrency.value,
    })
    await loadInvoices()
    closeNewInvoice()
    ui.success(t('invoice.invoices'))
  } catch {
    ui.error('Failed to create invoice')
  } finally {
    savingInvoice.value = false
  }
}

async function changeInvoiceStatus(inv, status) {
  try {
    const { data } = await customersApi.updateInvoice(custId.value, inv.id, { status })
    const idx = invoices.value.findIndex(i => i.id === inv.id)
    if (idx >= 0) invoices.value[idx] = data
  } catch {
    ui.error('Failed to update invoice')
  }
}

async function doDeleteInvoice(inv) {
  const msg = inv.status === 'credit_note'
    ? t('invoice.void_credit_note_confirm', { number: inv.invoice_number })
    : t('invoice.delete_confirm')
  if (!await ui.confirm(msg, { destructive: true })) return
  try {
    await customersApi.deleteInvoice(custId.value, inv.id)
    invoices.value = invoices.value.filter(i => i.id !== inv.id)
  } catch {
    ui.error('Failed to delete invoice')
  }
}

// ── Edit line items ──────────────────────────────────────────────────────────
const showEditLines = ref(false)
const editingInvoice = ref(null)
const editLineItems = ref([])
const editLinesTotal = computed(() => editLineItems.value.reduce((s, li) => s + (li.amount || 0), 0))

function openEditLines(inv) {
  editingInvoice.value = inv
  try {
    editLineItems.value = JSON.parse(inv.line_items || '[]').map(li => ({ ...li }))
  } catch {
    editLineItems.value = []
  }
  showEditLines.value = true
}

function addManualLine() {
  const today = new Date().toISOString().slice(0, 10)
  editLineItems.value.push({
    date: today,
    project_name: '',
    description: '',
    minutes: 0,
    hourly_rate: 0,
    distance: 0,
    price_per_km: 0,
    quantity: 1,
    unit_price: 0,
    amount: 0,
    currency: editingInvoice.value?.currency || '€',
    is_manual: true,
  })
}

function addCommentLine() {
  editLineItems.value.push({
    date: '',
    project_name: '',
    description: '',
    minutes: 0,
    hourly_rate: 0,
    distance: 0,
    price_per_km: 0,
    amount: 0,
    currency: editingInvoice.value?.currency || '€',
    is_comment: true,
  })
}

async function saveEditLines() {
  const inv = editingInvoice.value
  try {
    const { data } = await customersApi.updateInvoice(custId.value, inv.id, {
      line_items: editLineItems.value,
      vat_rate: inv.vat_rate,
    })
    const idx = invoices.value.findIndex(i => i.id === inv.id)
    if (idx >= 0) invoices.value[idx] = data
    showEditLines.value = false
    ui.success(t('common.saved'))
  } catch {
    ui.error('Failed to save line items')
  }
}

// ── Send by email ─────────────────────────────────────────────────────────────
const showSendEmail = ref(false)
const sendEmailInvoice = ref(null)
const sendEmailForm = ref({ to: '', subject: '', body: '' })
const sendingEmail = ref(false)
const sendEmailPreview = ref(false)
const sendEmailRendered = computed(() =>
  DOMPurify.sanitize(marked.parse(sendEmailForm.value.body || ''))
)

function openSendEmail(inv) {
  sendEmailInvoice.value = inv
  sendEmailPreview.value = false
  const primary = contacts.value.find(c => c.is_primary)
  const toEmail = primary?.email || detail.value?.customer?.email || ''
  const firstName = primary ? primary.name.split(' ')[0] : 'customer'
  sendEmailForm.value = {
    to: toEmail,
    subject: `${t('invoice.invoices')} ${inv.invoice_number}`,
    body: `Dear ${firstName},\n\nPlease find your invoice **${inv.invoice_number}** attached.\n\nKind regards`,
  }
  showSendEmail.value = true
}

async function doSendEmail() {
  const inv = sendEmailInvoice.value
  sendingEmail.value = true
  try {
    const { data } = await customersApi.sendInvoice(custId.value, inv.id, {
      to: sendEmailForm.value.to,
      subject: sendEmailForm.value.subject,
      body: sendEmailForm.value.body,
      body_html: DOMPurify.sanitize(marked.parse(sendEmailForm.value.body || '')),
    })
    const idx = invoices.value.findIndex(i => i.id === inv.id)
    if (idx >= 0) invoices.value[idx] = data
    showSendEmail.value = false
    ui.success(t('invoice.send_email'))
  } catch (err) {
    const msg = err?.response?.data?.error || 'Failed to send email'
    if (msg.includes('not configured')) {
      ui.error(t('invoice.no_email_configured'))
    } else {
      ui.error(msg)
    }
  } finally {
    sendingEmail.value = false
  }
}

// ── Overdue check ─────────────────────────────────────────────────────────────
const _today = new Date()
_today.setHours(0, 0, 0, 0)
function isInvoiceOverdue(inv) {
  return inv.status === 'sent' && !!inv.due_date && new Date(inv.due_date) < _today
}

// ── Record payment ────────────────────────────────────────────────────────────
const showRecordPayment = ref(false)
const paymentInvoice = ref(null)
const paymentForm = ref({ payment_date: '', display_payment_date: '', payment_amount: null, payment_reference: '', payment_method: '' })

function openRecordPayment(inv) {
  paymentInvoice.value = inv
  const today = new Date().toISOString().slice(0, 10)
  paymentForm.value = {
    payment_date: today,
    display_payment_date: formatDate(today),
    payment_amount: inv.total,
    payment_reference: '',
    payment_method: '',
  }
  showRecordPayment.value = true
}

function parsePaymentDate() {
  const val = paymentForm.value.display_payment_date.trim()
  if (!val) { paymentForm.value.payment_date = ''; return }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY'), mPos = fmt.indexOf('MM'), dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    paymentForm.value.display_payment_date = paymentForm.value.payment_date ? formatDate(paymentForm.value.payment_date) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  paymentForm.value.payment_date = iso
  paymentForm.value.display_payment_date = formatDate(iso)
}

async function doRecordPayment() {
  const inv = paymentInvoice.value
  try {
    const { data } = await customersApi.updateInvoice(custId.value, inv.id, {
      status: 'paid',
      payment_date: paymentForm.value.payment_date,
      payment_amount: paymentForm.value.payment_amount,
      payment_reference: paymentForm.value.payment_reference,
      payment_method: paymentForm.value.payment_method || undefined,
    })
    const idx = invoices.value.findIndex(i => i.id === inv.id)
    if (idx >= 0) invoices.value[idx] = data
    showRecordPayment.value = false
    ui.success(t('invoice.payment_info'))
  } catch {
    ui.error('Failed to record payment')
  }
}

// ── Credit notes ──────────────────────────────────────────────────────────────
async function doIssueCreditNote(inv) {
  if (!await ui.confirm(t('invoice.credit_note_confirm', { number: inv.invoice_number }), { destructive: false })) return
  try {
    const { data } = await customersApi.createCreditNote(custId.value, inv.id)
    invoices.value.unshift(data)
    ui.success(t('invoice.credit_note'))
  } catch {
    ui.error('Failed to create credit note')
  }
}

// ── Contacts ─────────────────────────────────────────────────────────────────
const contacts = ref([])
const showContactModal = ref(false)
const editingContact = ref(null)
const contactForm = ref({ name: '', department: '', phone: '', email: '', is_primary: false })

watch(() => detail.value?.contacts, (v) => { if (v) contacts.value = [...v] }, { immediate: true })

function openAddContact() {
  editingContact.value = null
  contactForm.value = { name: '', department: '', phone: '', email: '', is_primary: contacts.value.length === 0 }
  showContactModal.value = true
}

function openEditContact(ct) {
  editingContact.value = ct
  contactForm.value = { name: ct.name, department: ct.department || '', phone: ct.phone || '', email: ct.email || '', is_primary: ct.is_primary }
  showContactModal.value = true
}

async function saveContact() {
  try {
    if (editingContact.value) {
      const { data } = await customersApi.updateContact(custId.value, editingContact.value.id, contactForm.value)
      const idx = contacts.value.findIndex(c => c.id === data.id)
      if (idx >= 0) contacts.value.splice(idx, 1, data)
      if (data.is_primary) contacts.value.forEach(c => { if (c.id !== data.id) c.is_primary = false })
    } else {
      const { data } = await customersApi.createContact(custId.value, contactForm.value)
      if (data.is_primary) contacts.value.forEach(c => { c.is_primary = false })
      contacts.value.push(data)
      contacts.value.sort((a, b) => (b.is_primary ? 1 : 0) - (a.is_primary ? 1 : 0))
    }
    showContactModal.value = false
  } catch {
    ui.error('Failed to save contact')
  }
}

async function removeContact(ct) {
  if (!await ui.confirm(t('customer.delete_contact_confirm'), { destructive: true })) return
  try {
    await customersApi.deleteContact(custId.value, ct.id)
    contacts.value = contacts.value.filter(c => c.id !== ct.id)
  } catch {
    ui.error('Failed to delete contact')
  }
}

const prefers12HourSlotTime = computed(() => {
  const fmt = auth.user?.date_time_format || 'YYYY-MM-DD HH:mm'
  return fmt.includes('hh') && fmt.includes('a')
})

function formatSlotTime(raw) {
  if (!raw) return ''
  const [hhStr, mmStr] = raw.split(':')
  const hh = Number(hhStr), mm = Number(mmStr)
  if (!Number.isInteger(hh) || !Number.isInteger(mm)) return ''
  if (!prefers12HourSlotTime.value) return `${String(hh).padStart(2,'0')}:${String(mm).padStart(2,'0')}`
  const suffix = hh < 12 ? 'am' : 'pm'
  const h12 = hh % 12 || 12
  return `${String(h12).padStart(2,'0')}:${String(mm).padStart(2,'0')} ${suffix}`
}

function parseSlotTime(input) {
  const val = (input || '').trim()
  if (!val) return ''
  if (prefers12HourSlotTime.value) {
    const m = val.match(/^(\d{1,2}):(\d{2})\s*([ap]m)$/i)
    if (!m) return null
    let hh = Number(m[1]); const mm = Number(m[2]); const mer = m[3].toLowerCase()
    if (hh < 1 || hh > 12 || mm < 0 || mm > 59) return null
    if (mer === 'am') hh = hh === 12 ? 0 : hh
    else hh = hh === 12 ? 12 : hh + 12
    return `${String(hh).padStart(2,'0')}:${String(mm).padStart(2,'0')}`
  }
  const m = val.match(/^(\d{1,2}):(\d{2})$/)
  if (!m) return null
  const hh = Number(m[1]), mm = Number(m[2])
  if (hh < 0 || hh > 23 || mm < 0 || mm > 59) return null
  return `${String(hh).padStart(2,'0')}:${String(mm).padStart(2,'0')}`
}

function onSlotTimeBlur(slot, field, event) {
  const parsed = parseSlotTime(event.target.value)
  if (parsed === null) {
    event.target.value = formatSlotTime(slot[field])
  } else {
    slot[field] = parsed
    event.target.value = formatSlotTime(parsed)
  }
}

const slotTimePlaceholder = computed(() => prefers12HourSlotTime.value ? 'hh:mm am' : 'HH:mm')

const showAddContract = ref(false)
const editingContract = ref(null)
const contractForm = ref({ name: '', description: '', start_date: '', end_date: '', price_per_hour: null, price_per_km: null, currency: '€', time_slots: [] })
const emptySlot = () => ({ label: '', start_time: '', end_time: '', day_type: 'all', end_day_offset: 1, multiplication_factor: null, hourly_rate: null })

function isOvernightSlot(slot) {
  return !!(slot.start_time && slot.end_time && slot.end_time <= slot.start_time)
}

function formatSlotTimeRange(slot) {
  if (!slot.start_time || !slot.end_time) return ''
  if (!isOvernightSlot(slot)) return `${slot.start_time}–${slot.end_time}`
  const offset = slot.end_day_offset > 0 ? slot.end_day_offset : 1
  if (offset === 1) return `${slot.start_time}–${slot.end_time}`
  return `${slot.start_time} → +${offset}d ${slot.end_time}`
}

const slotPreviewDowLabels = computed(() => [
  t('contract.slot_preview_dow_mon'),
  t('contract.slot_preview_dow_tue'),
  t('contract.slot_preview_dow_wed'),
  t('contract.slot_preview_dow_thu'),
  t('contract.slot_preview_dow_fri'),
  t('contract.slot_preview_dow_sat'),
  t('contract.slot_preview_dow_sun'),
])

// Compact abbrevs and full names for the day-of-week checkbox row
const slotDowAbbrevs = computed(() => slotPreviewDowLabels.value)
const slotDowFullNames = computed(() => [
  t('contract.slot_days_monday'),
  t('contract.slot_days_tuesday'),
  t('contract.slot_days_wednesday'),
  t('contract.slot_days_thursday'),
  t('contract.slot_days_friday'),
  t('contract.slot_days_saturday'),
  t('contract.slot_days_sunday'),
])

const ISO_DOW_KEYS = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday']

function getCheckedDayIndices(dayType) {
  if (!dayType || dayType === 'all') return [0, 1, 2, 3, 4, 5, 6]
  if (dayType === 'weekdays') return [0, 1, 2, 3, 4]
  if (dayType === 'weekends') return [5, 6]
  return dayType.split(',')
    .map(d => ISO_DOW_KEYS.indexOf(d.trim()))
    .filter(i => i >= 0)
}

function isSlotDayChecked(slot, idx) {
  return getCheckedDayIndices(slot.day_type).includes(idx)
}

function toggleSlotDay(slot, idx) {
  const checked = new Set(getCheckedDayIndices(slot.day_type))
  if (checked.has(idx)) {
    checked.delete(idx)
    if (checked.size === 0) checked.add(idx) // keep at least one day
  } else {
    checked.add(idx)
  }
  const arr = [...checked].sort((a, b) => a - b)
  if (arr.length === 7) {
    slot.day_type = 'all'
  } else if (arr.join(',') === '0,1,2,3,4') {
    slot.day_type = 'weekdays'
  } else if (arr.join(',') === '5,6') {
    slot.day_type = 'weekends'
  } else if (arr.length === 1) {
    slot.day_type = ISO_DOW_KEYS[arr[0]]
  } else {
    slot.day_type = arr.map(i => ISO_DOW_KEYS[i]).join(',')
  }
}

const SLOT_SEG_COLORS = [
  null,                        // idx 0 — falls back to CSS default (primary)
  'rgba(245,158,11,0.65)',      // amber
  'rgba(16,185,129,0.65)',      // emerald
  'rgba(168,85,247,0.65)',      // violet
  'rgba(239,68,68,0.65)',       // red
  'rgba(20,184,166,0.65)',      // teal
]
function slotSegColor(idx) {
  return SLOT_SEG_COLORS[idx % SLOT_SEG_COLORS.length]
}

function slotPreviewReady(slot) {
  return slotPreviewReadyFn(slot)
}

function slotDurationMinutes(slot) {
  if (!slot.start_time || !slot.end_time) return null
  const [sh, sm] = slot.start_time.split(':').map(Number)
  const [eh, em] = slot.end_time.split(':').map(Number)
  if (isNaN(sh) || isNaN(sm) || isNaN(eh) || isNaN(em)) return null
  const startMins = sh * 60 + sm
  const endMins = eh * 60 + em
  if (endMins > startMins) return endMins - startMins
  const offset = (slot.end_day_offset > 0 ? slot.end_day_offset : 1)
  return (24 * 60 - startMins) + endMins + (offset - 1) * 24 * 60
}

function formatMinutes(mins) {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m === 0 ? `${h}h` : `${h}h ${m}m`
}

function slotPreviewDays(slot) {
  return buildSlotPreviewDays(slot, slotPreviewDowLabels.value)
}

function slotPreviewAria(slot) {
  const range = formatSlotTimeRange(slot)
  const days = formatDayType(slot.day_type || 'all', t)
  return t('contract.slot_preview_aria', { range, days })
}
const displayContractStartDate = ref('')
const displayContractEndDate   = ref('')

function _parseContractDate(displayRef, isoKey) {
  const val = displayRef.value.trim()
  if (!val) { contractForm.value[isoKey] = ''; return }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY'), mPos = fmt.indexOf('MM'), dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayRef.value = contractForm.value[isoKey] ? formatDate(contractForm.value[isoKey]) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  contractForm.value[isoKey] = iso
  displayRef.value = formatDate(iso)
}

function parseContractStartDate() { _parseContractDate(displayContractStartDate, 'start_date') }
function parseContractEndDate()   { _parseContractDate(displayContractEndDate,   'end_date')   }

function onContractStartDateChange(e) {
  const iso = e.target.value
  contractForm.value.start_date = iso
  displayContractStartDate.value = iso ? formatDate(iso) : ''
}
function onContractEndDateChange(e) {
  const iso = e.target.value
  contractForm.value.end_date = iso
  displayContractEndDate.value = iso ? formatDate(iso) : ''
}

const editingName = ref(false)
const nameEdit = ref('')

const canManage = computed(() => auth.isAdmin || detail.value?.customer?.my_role === 'admin')

const authUserId = computed(() => auth.user?.id)

// ── Member management ──────────────────────────────────────────────────────
const members = ref([])
const customerGroups = ref([])
const allGroups = ref([])
let allGroupsLoaded = false
const addGroupId = ref('')
const addGroupRole = ref('member')

const groupsNotOnCustomer = computed(() => {
  const assigned = new Set(customerGroups.value.map(g => g.group_id))
  return allGroups.value.filter(g => !assigned.has(g.id))
})

const showAddMember = ref(false)
const allUsers = ref([])
let allUsersLoaded = false
const memberSearch = ref('')
const pendingMemberIds = ref([])

const filteredUsers = computed(() => {
  const q = memberSearch.value.toLowerCase()
  const existingIds = new Set(members.value.map(m => m.user_id))
  return allUsers.value.filter(u => {
    if (existingIds.has(u.id)) return false
    if (!q) return true
    return (u.display_name || '').toLowerCase().includes(q) ||
           u.username.toLowerCase().includes(q) ||
           u.email.toLowerCase().includes(q)
  })
})

function projectAvatar(project) {
  return resolveAssetUrl(project?.avatar || '')
}

function groupAvatar(group) {
  return resolveAssetUrl(group?.avatar || '')
}

async function loadMembers() {
  try {
    const { data } = await customersApi.listMembers(custId.value)
    members.value = data || []
  } catch {}
  try {
    const { data } = await groupsApi.listCustomerGroups(custId.value)
    customerGroups.value = data || []
  } catch {}
  if (auth.isAdmin && !allGroupsLoaded) {
    try {
      const { data } = await groupsApi.list()
      allGroups.value = data || []
      allGroupsLoaded = true
    } catch {}
  }
}

async function addGroupToCustomer() {
  if (!addGroupId.value) return
  try {
    await groupsApi.setCustomerAccess(addGroupId.value, custId.value, addGroupRole.value)
    const { data } = await groupsApi.listCustomerGroups(custId.value)
    customerGroups.value = data || []
    addGroupId.value = ''
  } catch {
    ui.error('Failed to add group')
  }
}

async function removeGroupFromCustomer(groupId) {
  try {
    await groupsApi.removeCustomerAccess(groupId, custId.value)
    customerGroups.value = customerGroups.value.filter(g => g.group_id !== groupId)
  } catch {
    ui.error('Failed to remove group')
  }
}

async function openAddMember() {
  memberSearch.value = ''
  pendingMemberIds.value = []
  if (!allUsersLoaded) {
    try {
      const { data } = await client.get('/users')
      allUsers.value = data || []
      allUsersLoaded = true
    } catch {}
  }
  showAddMember.value = true
}

function togglePendingMember(id) {
  const idx = pendingMemberIds.value.indexOf(id)
  if (idx >= 0) pendingMemberIds.value.splice(idx, 1)
  else pendingMemberIds.value.push(id)
}

async function confirmAddMembers() {
  const newMembers = [
    ...members.value.map(m => ({ user_id: m.user_id, role: m.role })),
    ...pendingMemberIds.value.map(id => ({ user_id: id, role: 'member' })),
  ]
  try {
    await customersApi.setMembers(custId.value, newMembers)
    await loadMembers()
    showAddMember.value = false
  } catch {
    ui.error('Failed to add members')
  }
}

async function removeMember(userId) {
  const newMembers = members.value
    .filter(m => m.user_id !== userId)
    .map(m => ({ user_id: m.user_id, role: m.role }))
  try {
    await customersApi.setMembers(custId.value, newMembers)
    await loadMembers()
  } catch {
    ui.error('Failed to remove member')
  }
}

async function setMemberRole(userId, role) {
  const newMembers = members.value.map(m =>
    m.user_id === userId ? { user_id: m.user_id, role } : { user_id: m.user_id, role: m.role }
  )
  try {
    await customersApi.setMembers(custId.value, newMembers)
    await loadMembers()
  } catch {
    ui.error('Failed to update role')
  }
}

const custId = computed(() => Number(route.params.id))

onMounted(() => { load(); loadInvoiceTemplates() })
watch(custId, () => load())

async function load() {
  loading.value = true
  try {
    const { data } = await customersApi.get(custId.value)
    detail.value = data
    if (data?.customer?.my_role === 'admin' || auth.isAdmin) {
      await loadMembers()
    }
    await loadInvoices()
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
  const c = detail.value.customer
  editForm.value = {
    name: c.name, description: c.description || '', logo_url: c.logo_url || '',
    billing_street: c.billing_street || '', billing_city: c.billing_city || '',
    billing_postal_code: c.billing_postal_code || '', billing_country: c.billing_country || '',
    vat_number: c.vat_number || '', po_reference: c.po_reference || '',
  }
  showEdit.value = true
}

async function onLogoFileSelected(e) {
  const file = e.target.files[0]
  if (!file) return
  e.target.value = ''
  try {
    const { data } = await attachmentsApi.uploadImage(file)
    editForm.value.logo_url = data.url
  } catch {
    ui.error('Failed to upload image')
  }
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
  if (!await ui.confirm(this.$t?.('customer.delete_confirm') || 'Delete this customer?', { destructive: true })) return
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
    price_per_hour: grp.price_per_hour != null ? grp.price_per_hour : null,
    price_per_km:   grp.price_per_km != null ? grp.price_per_km : null,
    currency:       grp.currency || '€',
    time_slots: (grp.time_slots || []).map(s => ({
      label:                s.label || '',
      start_time:           s.start_time || '',
      end_time:             s.end_time || '',
      day_type:             s.day_type || 'all',
      end_day_offset:       s.end_day_offset > 0 ? s.end_day_offset : 1,
      multiplication_factor: s.multiplication_factor != null ? s.multiplication_factor : null,
      hourly_rate:           s.hourly_rate != null ? s.hourly_rate : null,
    })),
  }
  displayContractStartDate.value = contractForm.value.start_date ? formatDate(contractForm.value.start_date) : ''
  displayContractEndDate.value   = contractForm.value.end_date   ? formatDate(contractForm.value.end_date)   : ''
}

function addTimeSlot() {
  contractForm.value.time_slots.push(emptySlot())
}

function removeTimeSlot(idx) {
  contractForm.value.time_slots.splice(idx, 1)
}

function closeContractModal() {
  showAddContract.value = false
  editingContract.value = null
  contractForm.value = { name: '', description: '', start_date: '', end_date: '', price_per_hour: null, currency: '€', time_slots: [] }
  displayContractStartDate.value = ''
  displayContractEndDate.value   = ''
}

async function saveContract() {
  const payload = {
    name:           contractForm.value.name,
    description:    contractForm.value.description,
    start_date:     contractForm.value.start_date || '',
    end_date:       contractForm.value.end_date   || '',
    price_per_hour: contractForm.value.price_per_hour,
    price_per_km:   contractForm.value.price_per_km,
    currency:       contractForm.value.currency || '€',
    time_slots:     contractForm.value.time_slots
      .filter(s => s.start_time && s.end_time)
      .map(s => ({
        label:                s.label || '',
        start_time:           s.start_time,
        end_time:             s.end_time,
        day_type:             s.day_type || 'all',
        end_day_offset:       isOvernightSlot(s) ? (s.end_day_offset > 0 ? s.end_day_offset : 1) : 0,
        multiplication_factor: s.multiplication_factor !== null && s.multiplication_factor !== '' ? parseFloat(s.multiplication_factor) : null,
        hourly_rate:           s.hourly_rate !== null && s.hourly_rate !== '' ? parseFloat(s.hourly_rate) : null,
      })),
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
  if (!await ui.confirm('Delete contract "' + grp.name + '"?', { destructive: true })) return
  try {
    await customersApi.deleteContract(custId.value, grp.id)
    await load()
  } catch {
    ui.error('Failed to delete contract')
  }
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
  margin-bottom: 0;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--color-border);
}

.cust-tabs {
  display: flex;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 28px;
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
  transition: color 0.15s;
}
.tab:hover { color: var(--color-text); }
.tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.tab:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }

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

.contract-name-btn {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  font-weight: 700;
  color: var(--color-text);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: transparent;
  text-underline-offset: 2px;
  transition: text-decoration-color 0.15s;
}
.contract-name-btn:hover,
.contract-name-btn:focus-visible {
  text-decoration-color: var(--color-text);
  outline: none;
}

.contract-dates {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-left: 4px;
}

.contract-rate {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary);
  margin-left: 8px;
  white-space: nowrap;
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
.proj-avatar {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  object-fit: cover;
  border: 1px solid var(--color-border);
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


.form-group { margin-bottom: 12px; }
.form-label { display: block; font-size: 12px; font-weight: 600; margin-bottom: 4px; color: var(--color-text-muted); }
.form-input { width: 100%; padding: 8px 10px; border: 1px solid var(--color-border); border-radius: 6px; background: var(--color-bg); color: var(--color-text); font-size: 14px; box-sizing: border-box; }
.detail-row { display: flex; gap: 12px; }
.half { flex: 1; }

.date-input-row { display: flex; align-items: center; gap: 6px; }
.date-input-row .form-input { flex: 1; }
.picker-wrap { position: relative; display: inline-flex; cursor: pointer; }
.date-picker-overlay { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
.btn-icon-xs {
  background: none; border: none; cursor: pointer; color: var(--color-text-muted);
  padding: 2px 4px; font-size: 13px; line-height: 1; border-radius: 3px; flex-shrink: 0;
}
.btn-icon-xs:hover { background: var(--color-bg); color: var(--color-text); }

.contacts-section { margin-top: 32px; }

.contacts-list { display: flex; flex-direction: column; gap: 8px; }

.contact-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.contact-info { flex: 1; min-width: 0; display: flex; flex-wrap: wrap; align-items: center; gap: 4px 16px; }
.contact-name { font-size: 14px; font-weight: 600; white-space: nowrap; }
.contact-detail { font-size: 12px; color: var(--color-text-muted); white-space: nowrap; }

.badge-primary-contact {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 99px;
  background: #0ea5e9;
  color: #fff;
  flex-shrink: 0;
}

.contact-actions { display: flex; gap: 4px; flex-shrink: 0; }

.form-check-row { display: flex; align-items: center; gap: 8px; }
.form-check-row input[type="checkbox"] { width: 16px; height: 16px; flex-shrink: 0; }
.form-check-row label { font-size: 13px; color: var(--color-text); cursor: pointer; }

.members-section { margin-top: 32px; }

.members-list { display: flex; flex-direction: column; gap: 8px; }

.member-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.member-avatar { width: 32px; height: 32px; border-radius: 50%; object-fit: cover; flex-shrink: 0; }
.group-avatar {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid var(--color-border);
  flex-shrink: 0;
}

.member-info { flex: 1; min-width: 0; }
.member-name { display: block; font-size: 13px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.member-email { display: block; font-size: 11px; color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.role-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 99px;
  flex-shrink: 0;
}
.role-admin { background: #0ea5e9; color: #fff; }
.role-member { background: var(--color-bg); color: var(--color-text-muted); border: 1px solid var(--color-border); }

.member-actions { display: flex; gap: 4px; flex-shrink: 0; }

.user-picker-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  margin-top: 4px;
}

.user-picker-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
}
.user-picker-row:last-child { border-bottom: none; }
.user-picker-row:hover { background: var(--color-bg); }
.user-picker-row.selected { background: color-mix(in srgb, var(--color-primary) 8%, transparent); }

.check-mark { font-size: 14px; color: var(--color-primary); font-weight: 700; flex-shrink: 0; }

/* Time slots — contract display (compact badge + hover popup) */
.slot-indicator-wrap {
  position: relative;
  display: inline-flex;
  margin-top: 6px;
}
.slot-badge {
  font-size: 11px;
  background: var(--color-surface-alt);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 1px 7px;
  cursor: default;
  white-space: nowrap;
  user-select: none;
}
.slot-badge:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
.slot-popup {
  display: none;
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: 6px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 8px 10px;
  box-shadow: 0 4px 12px rgba(0,0,0,.15);
  z-index: 10;
  min-width: 240px;
  flex-direction: column;
  gap: 5px;
}
.slot-indicator-wrap:hover .slot-popup,
.slot-indicator-wrap:focus-within .slot-popup {
  display: flex;
}
.slot-popup-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-muted);
  white-space: nowrap;
}
.slot-popup-time { font-variant-numeric: tabular-nums; white-space: nowrap; }
.slot-popup-days { font-size: 11px; background: var(--color-bg); border: 1px solid var(--color-border); border-radius: 4px; padding: 1px 5px; white-space: nowrap; }
.slot-popup-label { color: var(--color-text); }
.slot-popup-factor { font-weight: 600; color: var(--color-primary); }
.slot-popup-rate { font-weight: 600; color: var(--color-primary); }

/* Time slots — contract form */
.slots-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.slots-header .form-label { margin-bottom: 0; }
.form-hint { font-size: 12px; color: var(--color-text-muted); margin: 0 0 8px; }
.slots-list { display: flex; flex-direction: column; gap: 8px; }
.slot-card {
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.slot-card-top {
  display: flex;
  gap: 6px;
  align-items: center;
}
.slot-card-top .form-input { flex: 1; }
.slot-card-bottom {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.slot-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}
.slot-field-days { flex: 1; min-width: 110px; }
.slot-field-duration { flex: 1; min-width: 140px; }
.slot-duration-summary { display: flex; flex-wrap: wrap; gap: 4px 10px; align-items: baseline; font-size: 12px; padding-top: 2px; }
.slot-dur-item { color: var(--color-text); }
.slot-dur-total { color: var(--color-primary); font-weight: 600; white-space: nowrap; }
.slot-dur-single { color: var(--color-primary); font-weight: 600; }
.slot-field .form-input { padding: 5px 7px; font-size: 13px; width: 90px; }
.slot-field-days .form-input { width: 100%; }
.slot-day-checks { display: flex; gap: 3px; flex-wrap: wrap; }
.slot-day-check { display: flex; align-items: center; justify-content: center; width: 26px; height: 26px; border-radius: 4px; border: 1px solid var(--color-border); font-size: 11px; font-weight: 600; cursor: pointer; user-select: none; color: var(--color-text-muted); background: var(--color-surface); transition: background 0.12s, color 0.12s, border-color 0.12s; }
.slot-day-check.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.slot-day-check:hover { border-color: var(--color-primary); color: var(--color-primary); }
.slot-day-check.active:hover { color: #fff; }
.slot-field-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .04em;
  color: var(--color-text-muted);
}
.slot-remove { color: var(--color-danger, #ef4444); flex-shrink: 0; }
.slot-remove:hover { background: color-mix(in srgb, var(--color-danger, #ef4444) 10%, transparent); }

.slot-preview {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-top: 4px;
  border-top: 1px dashed var(--color-border);
}
.slot-preview-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .04em;
  color: var(--color-text-muted);
}
.slot-preview-week {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
}
.slot-preview-day {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.slot-preview-dow {
  font-size: 10px;
  text-align: center;
  color: var(--color-text-muted);
}
.slot-preview-day-active .slot-preview-dow {
  color: var(--color-text);
  font-weight: 600;
}
.slot-preview-track {
  position: relative;
  height: 10px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}
.slot-preview-seg {
  position: absolute;
  top: 0;
  bottom: 0;
  background: color-mix(in srgb, var(--color-primary) 55%, transparent);
  border-radius: 2px;
  min-width: 2px;
}

/* ── Invoices ── */
.invoices-section { margin-top: 32px; }

.invoice-list { display: flex; flex-direction: column; gap: 8px; }

.invoice-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  flex-wrap: wrap;
}

.invoice-row-main { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; flex: 1; min-width: 0; }
.invoice-row-actions { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }

.invoice-number { font-weight: 700; font-size: 13px; white-space: nowrap; }
.invoice-period { font-size: 12px; color: var(--color-text-muted); white-space: nowrap; }
.invoice-total  { font-size: 13px; font-weight: 600; white-space: nowrap; }
.invoice-due    { font-size: 12px; color: var(--color-text-muted); white-space: nowrap; }

.invoice-status {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 99px;
  white-space: nowrap;
  flex-shrink: 0;
}
.inv-draft       { background: var(--color-bg); color: var(--color-text-muted); border: 1px solid var(--color-border); }
.inv-sent        { background: #dbeafe; color: #1d4ed8; }
.inv-overdue     { background: color-mix(in srgb, var(--color-danger, #e53e3e) 15%, transparent); color: var(--color-danger, #e53e3e); }
.inv-paid        { background: #dcfce7; color: #15803d; }
.inv-credit_note { background: #fef9c3; color: #854d0e; }

.invoice-row--overdue { background: color-mix(in srgb, var(--color-danger, #e53e3e) 5%, var(--color-surface)); }
.invoice-due--overdue { color: var(--color-danger, #e53e3e); font-weight: 600; }

.inv-preview-table-wrap { overflow-x: auto; margin-bottom: 12px; }
.inv-preview-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.inv-preview-table th,
.inv-preview-table td {
  padding: 5px 8px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  white-space: nowrap;
}
.inv-preview-table th {
  background: var(--color-bg);
  font-weight: 600;
  color: var(--color-text-muted);
  font-size: 11px;
}
.inv-preview-table td.num,
.inv-preview-table th.num { text-align: right; }
.inv-preview-table tbody tr:nth-child(even) { background: var(--color-bg); }

.inv-preview-totals {
  display: flex;
  gap: 20px;
  justify-content: flex-end;
  font-size: 13px;
  padding: 8px 0 0;
  flex-wrap: wrap;
}
.inv-preview-totals strong { color: var(--color-primary); }

.email-body-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
.email-body-header .form-label { margin-bottom: 0; }
.email-body-tabs { display: flex; gap: 2px; }
.email-tab {
  background: none; border: 1px solid var(--color-border); border-radius: 4px;
  padding: 2px 10px; font-size: 12px; cursor: pointer; color: var(--color-text-muted);
}
.email-tab.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.email-tab:hover:not(.active) { border-color: var(--color-primary); color: var(--color-primary); }
.email-body-textarea { resize: vertical; font-family: monospace; font-size: 13px; }
.email-body-preview {
  min-height: 140px; max-height: 340px; overflow-y: auto;
  border: 1px solid var(--color-border); border-radius: 6px;
  padding: 10px 14px; background: var(--color-bg); font-size: 14px; line-height: 1.6;
}
.email-body-preview :deep(p) { margin: 0 0 .7em; }
.email-body-preview :deep(strong) { font-weight: 700; }
.email-body-preview :deep(em) { font-style: italic; }
.email-body-preview :deep(ul), .email-body-preview :deep(ol) { padding-left: 1.4em; margin: 0 0 .7em; }
.email-body-preview :deep(a) { color: var(--color-primary); }
.email-body-preview :deep(code) { background: var(--color-surface); border-radius: 3px; padding: 1px 4px; font-size: .9em; }
.invoice-credits       { font-size: 11px; color: var(--color-text-muted); font-style: italic; white-space: nowrap; }
.invoice-credit-issued { font-size: 11px; color: #b45309; background: #fef3c7; padding: 1px 6px; border-radius: 4px; white-space: nowrap; }
.invoice-paid-info { font-size: 11px; color: #15803d; white-space: nowrap; }

.inv-edit-table-wrap { overflow-x: auto; margin-bottom: 4px; }
.form-input-sm { padding: 4px 6px; font-size: 12px; }
.num-input { width: 80px; text-align: right; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }
</style>
