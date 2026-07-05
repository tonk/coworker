<template>
  <div class="tt-view" :class="{ 'tt-dragging': isDragging }">
    <h1 class="sr-only">{{ $t('timeTracking.nav') }}</h1>

    <!-- ── Top bar ─────────────────────────────────────────────────────────── -->
    <div class="tt-bar">
      <div class="tt-employee" v-show="mode !== 'board-report'">
        <span class="tt-emp-label">{{ $t('timeTracking.employee') }}</span>
        <select v-if="canViewOtherUsers" class="tt-user-select" v-model="selectedUserId" @change="onUserChange">
          <option :value="0">{{ $t('timeTracking.all_employees') }}</option>
          <option v-for="u in allUsers" :key="u.id" :value="u.id">{{ userName(u) }}</option>
        </select>
        <strong v-else class="tt-emp-name">{{ displayName }}</strong>
      </div>

      <div class="tt-week-nav" v-show="mode !== 'board-report'">
        <button class="nav-btn" @click="shiftWeek(-1)" :title="$t('timeTracking.prev_week')" :aria-label="$t('timeTracking.prev_week')">&#9664;</button>
        <div class="wk-picker-wrap" ref="wkPickerRef">
          <button class="wk-label" @click="openWkPicker"
            :aria-expanded="String(wkPickerOpen)" aria-haspopup="true"
            :aria-label="$t('timeTracking.week') + ' ' + weekInfo.week + ' ' + weekInfo.year">
            {{ $t('timeTracking.week') }} &nbsp; {{ weekInfo.week }} &nbsp; {{ weekInfo.year }}
          </button>
          <div v-if="wkPickerOpen" class="wk-cal" role="dialog" aria-modal="true" :aria-label="$t('timeTracking.week') + ' ' + weekInfo.week">
            <div class="wk-cal-hdr">
              <button class="wk-cal-nav" @click.stop="calShiftMonth(-1)" aria-label="Previous month">&#9664;</button>
              <span class="wk-cal-month">{{ calMonthLabel }}</span>
              <button class="wk-cal-nav" @click.stop="calShiftMonth(1)" aria-label="Next month">&#9654;</button>
            </div>
            <div class="wk-cal-grid">
              <span class="wk-cal-dn"></span>
              <span v-for="n in calDayNames" :key="n" class="wk-cal-dn">{{ n }}</span>
              <template v-for="(row, ri) in calCells" :key="ri">
                <button class="wk-cal-wn" :class="{ 'wk-cal-wn-sel': row.isSelectedWeek }"
                  @click.stop="goToDate(row.rowStart)" :aria-label="$t('timeTracking.week') + ' ' + row.weekNum">{{ row.weekNum }}</button>
                <button v-for="(cell, ci) in row.days" :key="ci"
                  class="wk-cal-day"
                  :class="{ 'wk-cal-other': cell.otherMonth, 'wk-cal-sel': cell.inSelectedWeek, 'wk-cal-today': cell.isToday }"
                  @click.stop="goToDate(cell.date)"
                  :aria-label="formatDate(cell.date)"
                  :aria-pressed="cell.inSelectedWeek">{{ cell.day }}</button>
              </template>
            </div>
          </div>
        </div>
        <button class="nav-btn" @click="shiftWeek(1)" :title="$t('timeTracking.next_week')" :aria-label="$t('timeTracking.next_week')">&#9654;</button>
        <button class="nav-btn nav-today" @click="goToToday" :title="$t('timeTracking.today')" :aria-label="$t('timeTracking.today')" :disabled="isCurrentWeek">{{ $t('timeTracking.today') }}</button>
        <div class="nav-holidays-wrap" ref="holidaysDropRef">
          <button class="nav-btn nav-holidays" @click="holidaysDropOpen = !holidaysDropOpen"
            :aria-expanded="String(holidaysDropOpen)" aria-haspopup="true"
            :aria-label="$t('timeTracking.add_holidays_title', { year: weekInfo.year })">
            {{ $t('timeTracking.add_holidays') }}<span class="nav-hol-chevron" :class="{ open: holidaysDropOpen }" aria-hidden="true">›</span>
          </button>
          <div v-if="holidaysDropOpen" class="holidays-drop" role="menu">
            <button v-for="c in holidayLocales" :key="c.locale"
              class="hol-country-btn" role="menuitem"
              @click="addHolidays(c.locale)">
              <span v-if="isTauri" class="hol-flag" :data-country="c.flag" aria-hidden="true"></span>
              <span v-else class="hol-flag-emoji" aria-hidden="true">{{ c.emoji }}</span>{{ c.label }}
            </button>
          </div>
        </div>
      </div>

      <div class="tt-mode-tabs" role="tablist" aria-label="Time tracking mode">
        <button
          v-if="auth.timeTrackingEnabled"
          id="tab-sheet"
          class="tt-mode-btn"
          role="tab"
          :aria-selected="mode === 'sheet'"
          aria-controls="panel-sheet"
          :class="{ active: mode === 'sheet' }"
          @click="mode = 'sheet'"
        >
          {{ $t('timeTracking.tab_log') }}
        </button>
        <button
          id="tab-report"
          class="tt-mode-btn"
          role="tab"
          :aria-selected="mode === 'report'"
          aria-controls="panel-report"
          :class="{ active: mode === 'report' }"
          @click="mode = 'report'"
          v-if="auth.timeTrackingEnabled"
        >
          {{ $t('timeTracking.tab_report') }}
        </button>
        <button
          v-if="auth.canViewReports && !systemStore.isTimetrackingMode"
          id="tab-board-report"
          class="tt-mode-btn"
          role="tab"
          :aria-selected="mode === 'board-report'"
          aria-controls="panel-board-report"
          :class="{ active: mode === 'board-report' }"
          @click="mode = 'board-report'"
        >
          {{ $t('timeTracking.tab_board_report') }}
        </button>
        <button
          id="tab-invoices"
          class="tt-mode-btn"
          role="tab"
          :aria-selected="mode === 'invoices'"
          aria-controls="panel-invoices"
          :class="{ active: mode === 'invoices' }"
          @click="mode = 'invoices'; loadAllInvoices()"
        >
          {{ $t('timeTracking.tab_invoices') }}
        </button>
      </div>
      <button class="tt-mode-btn tt-rates-btn" @click="showRates = true" :title="$t('contract.rates_overview')" :aria-label="$t('contract.rates_overview')">₡</button>
      <button
        v-if="auth.timeTrackingEnabled"
        class="tt-mode-btn tt-macro-edit-btn"
        @click="openMacroEditor"
        :title="$t('timeTracking.macro_edit')"
        :aria-label="$t('timeTracking.macro_edit')"
      >⚡</button>
      <button v-if="auth.timeTrackingEnabled" class="tt-mode-btn tt-manage-btn" @click="openManageProjects" :title="$t('timeTracking.manage_tt_projects')" :aria-label="$t('timeTracking.manage_tt_projects')">⚙</button>
    </div>

    <!-- ── Weekly timesheet ────────────────────────────────────────────────── -->
    <div v-show="mode === 'sheet'" id="panel-sheet" role="tabpanel" aria-labelledby="tab-sheet" class="tt-sheet-outer">
      <div v-if="loading" class="tt-loading">{{ $t('common.loading') }}</div>
      <div v-else class="tt-scroll">
        <table class="tt-table">
          <thead>
            <tr class="tt-head">
              <th class="c-drag"></th>
              <th class="c-nr"></th>
              <th class="c-info">
                <button class="sort-btn" :class="{ 'sort-active': sortCol === 'info' }" @click="toggleSort('info')" :title="$t('timeTracking.sort_by_customer')" :aria-label="sortCol === 'info' ? (sortDir === 'asc' ? 'Sort by customer/project descending' : 'Sort by customer/project ascending') : 'Sort by customer/project ascending'">
                  <span class="sort-label">
                    <span>{{ $t('timeTracking.customer') }}</span>
                    <span class="sub">{{ $t('timeTracking.project') }}</span>
                  </span>
                  <span class="sort-icon" aria-hidden="true">{{ sortCol === 'info' ? (sortDir === 'asc' ? '△' : '▽') : '▽' }}</span>
                </button>
              </th>
              <th class="c-desc">
                <button class="sort-btn" :class="{ 'sort-active': sortCol === 'desc' }" @click="toggleSort('desc')" :title="$t('timeTracking.sort_by_activity')" :aria-label="sortCol === 'desc' ? (sortDir === 'asc' ? 'Sort by activity descending' : 'Sort by activity ascending') : 'Sort by activity ascending'">
                  <span class="sort-label">{{ $t('timeTracking.activity') }}</span>
                  <span class="sort-icon" aria-hidden="true">{{ sortCol === 'desc' ? (sortDir === 'asc' ? '△' : '▽') : '▽' }}</span>
                </button>
              </th>
              <th class="c-rate">{{ $t('timeTracking.rate') }}</th>
              <th v-for="d in weekDays" :key="d.iso" :class="['c-day', holidayDates.has(d.iso) ? 'c-day-holiday' : '', d.isToday ? 'c-day-today' : '']">
                <div class="dh-abbr">{{ d.abbr }}</div>
                <div class="dh-date">{{ d.mmdd }}</div>
                <div v-if="holidayDates.has(d.iso)" class="dh-holiday-dot" aria-hidden="true"></div>
              </th>
              <th class="c-total">{{ $t('timeTracking.total') }}</th>
              <th class="c-act"></th>
            </tr>
          </thead>

          <tbody ref="tbodyEl">
            <tr v-for="(row, idx) in sortedRows" :key="row.key"
                :class="['tt-row', idx % 2 === 1 ? 'alt' : '', deletingRow === row.key ? 'tt-row-deleting' : '', row.is_holiday ? 'tt-row-holiday' : '']" :data-key="row.key">
              <td class="c-drag"><span class="drag-handle" aria-label="Reorder row">⠿</span></td>
              <td class="c-nr">{{ idx + 1 }}</td>

              <!-- Edit mode -->
              <template v-if="editingRow === row.key">
                <td class="c-info tt-editing">
                  <select class="nr-sel" v-model="editForm.customer_id" @change="editForm.project_id = null" aria-label="Customer">
                    <option :value="null">{{ $t('timeTracking.no_customer') }}</option>
                    <option v-for="c in allCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
                  </select>
                  <select v-if="editRowContracts.length" class="nr-sel" v-model="editForm.contract_id" :aria-label="$t('timeTracking.contract')">
                    <option :value="null">{{ $t('timeTracking.no_contract') }}</option>
                    <option v-for="c in editRowContracts" :key="c.id" :value="c.id">{{ c.name }}</option>
                  </select>
                  <select class="nr-sel" v-model="editForm.project_id" aria-label="Project">
                    <option :value="null">{{ $t('timeTracking.no_project') }}</option>
                    <option v-for="p in editRowProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
                  </select>
                </td>
                <td class="c-desc tt-editing">
                  <input class="nr-desc" type="text"
                    v-model="editForm.description"
                    :placeholder="$t('timeTracking.description')"
                    @keydown.enter="confirmEditRow(row)"
                    @keydown.escape="cancelEditRow"
                  />
                </td>
              </template>

              <!-- Normal mode -->
              <template v-else>
                <td
                  class="c-info"
                  :class="{ 'tt-field-editable': !viewingOther }"
                  :title="!viewingOther ? $t('common.double_click_edit') : undefined"
                  @dblclick="onRowFieldDblClick(row)"
                >
                  <div class="rc-cust">{{ row.customer_name || '—' }}</div>
                  <div v-if="row.contract_name" class="rc-contract">{{ row.contract_name }}</div>
                  <div v-else class="rc-proj">{{ row.project_name || '—' }}</div>
                  <span class="rc-comment-wrap">
                  <button v-if="!viewingOther" class="rc-comment-btn"
                    :class="{ 'has-comment': !!rowComments[row.key] }"
                    :aria-label="rowComments[row.key] ? $t('timeTracking.edit_comment') : $t('timeTracking.add_comment')"
                    :title="rowComments[row.key] ? $t('timeTracking.edit_comment') : $t('timeTracking.add_comment')"
                    @click.stop="openRowComment(row)"
                    @dblclick.stop
                  >
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                  </button>
                  <Teleport to="body">
                    <div v-if="commentRowKey === row.key" class="rc-comment-popup" :style="commentPopupStyle" data-no-drag="true">
                      <textarea ref="commentTextareaRef" class="rc-comment-textarea" v-model="commentDraft" :placeholder="$t('timeTracking.comment_placeholder')" rows="2" @keydown.escape="closeRowComment" @keydown.enter.meta="saveRowComment(row)" @keydown.enter.ctrl="saveRowComment(row)"></textarea>
                      <div class="rc-comment-actions">
                        <button class="tp-btn tp-apply" @mousedown.prevent="saveRowComment(row)">{{ $t('common.save') }}</button>
                        <button class="tp-btn tp-clear" @mousedown.prevent="closeRowComment">{{ $t('common.cancel') }}</button>
                      </div>
                    </div>
                  </Teleport>
                </span>
                </td>
                <td
                  class="c-desc"
                  :class="{ 'tt-field-editable': !viewingOther }"
                  :title="!viewingOther ? $t('common.double_click_edit') : undefined"
                  @dblclick="onRowFieldDblClick(row)"
                >{{ row.description || '—' }}</td>
              </template>

              <td class="c-rate">
                <span v-if="rowRateLabel(row)" class="rc-rate">
                  {{ rowRateLabel(row) }}
                  <span v-if="rowHasSlots(row)" class="rc-rate-slots" aria-hidden="true" :title="$t('contract.time_slots')">✦</span>
                </span>
              </td>

              <td v-for="(d, di) in weekDays" :key="d.iso" :data-ri="idx" :data-di="di" :class="['c-day', holidayDates.has(d.iso) ? 'c-day-holiday' : '', getEntry(row, d.iso)?.is_holiday ? 'c-day-holiday-cell' : '', cellSelectionClass(idx, di), isCellCopied(row, d.iso) ? 'c-day-copied' : '', d.isToday ? 'c-day-today' : '', isCellPopupOpen(row, d.iso) ? 'c-day-popup-open' : '']" :aria-selected="isCellSelected(idx, di) ? 'true' : 'false'">
                <input
                  :key="'c' + cellRenderEpoch + '-' + row.key + '-' + d.iso"
                  :id="'tt-cell-' + idx + '-' + di"
                  type="text"
                  :inputmode="timeNotation === 'hhmm' ? 'numeric' : 'decimal'"
                  autocomplete="off"
                  spellcheck="false"
                  class="h-inp"
                  :class="{ 'h-inp-filled': !!cellVal(row, d.iso) }"
                  :placeholder="savingCell === row.key + d.iso ? '…' : ''"
                  :value="cellVal(row, d.iso)"
                  :aria-label="timeNotation === 'hhmm' ? 'Hours and minutes for ' + d.abbr + ' ' + d.mmdd : 'Hours for ' + d.abbr + ' ' + d.mmdd"
                  :aria-describedby="cellUndeclMins(row, d.iso) > 0 ? 'tt-undecl-' + idx + '-' + di : undefined"
                  :disabled="viewingOther || savingCell === row.key + d.iso || editingRow === row.key"
                  @focus="onCellFocus(idx, di)"
                  @mousedown="onCellMouseDown(idx, di, $event)"
                  @focusin="$event.target.select()"
                  @input="onCellInput($event)"
                  @paste="onCellPaste($event)"
                  @beforeinput="onCellBeforeInput"
                  @keydown.capture="onCellUndoCapture"
                  @blur="onCellBlur(row, d.iso, $event.target.value)"
                  @keydown="onCellKeydown(row, d.iso, idx, di, $event)"
                />
                <span
                  v-if="getEntry(row, d.iso)?.start_time"
                  class="cell-time-dot"
                  aria-hidden="true"
                ></span>
                <button v-if="!viewingOther"
                  class="cell-hol-toggle"
                  :class="{ 'cell-hol-on': getEntry(row, d.iso)?.is_holiday }"
                  :aria-label="$t('timeTracking.is_holiday')"
                  :aria-pressed="!!getEntry(row, d.iso)?.is_holiday"
                  :title="$t('timeTracking.is_holiday')"
                  @mousedown.prevent="toggleCellHoliday(row, d.iso)"
                  tabindex="-1"
                ></button>
                <button v-if="!viewingOther"
                  class="cell-time-toggle"
                  :class="{ 'cell-time-on': !!getEntry(row, d.iso)?.start_time }"
                  :aria-label="$t('timeTracking.set_time_range')"
                  :title="$t('timeTracking.set_time_range')"
                  @mousedown.prevent="openTimePopup(row, d.iso, $event)"
                  tabindex="-1"
                >⏱</button>
                <!-- Time range popup -->
                <div
                  v-if="timePopupKey === row.key + d.iso"
                  class="time-popup"
                  :class="{ 'time-popup-below': timePopupFlip }"
                  ref="timePopupRef"
                  role="dialog"
                  :aria-label="$t('timeTracking.set_time_range')"
                  @mousedown.stop
                >
                  <label class="tp-label">{{ $t('timeTracking.start_time') }}
                    <input type="text" class="tp-inp" v-model="timePopupStart"
                      placeholder="09:00" maxlength="5"
                      @input="onTimePopupInput('start', $event)"
                      @keydown.enter.prevent="applyTimePopup(row, d.iso)"
                      @keydown.escape.prevent="timePopupKey = ''"
                    />
                  </label>
                  <label class="tp-label">{{ $t('timeTracking.end_time') }}
                    <input type="text" class="tp-inp" v-model="timePopupEnd"
                      placeholder="17:00" maxlength="5"
                      @input="onTimePopupInput('end', $event)"
                      @keydown.enter.prevent="applyTimePopup(row, d.iso)"
                      @keydown.escape.prevent="timePopupKey = ''"
                    />
                  </label>
                  <div v-if="timePopupMinutes() !== null" class="tp-preview">
                    {{ fmtTime(timePopupMinutes()) }}
                    <span v-if="timePopupOvernight()" class="tp-overnight">{{ $t('timeTracking.time_ends_next_day') }}</span>
                  </div>
                  <div class="tp-actions">
                    <button class="tp-btn tp-apply" @mousedown.prevent="applyTimePopup(row, d.iso)">{{ $t('common.save') }}</button>
                    <button v-if="getEntry(row, d.iso)?.start_time" class="tp-btn tp-clear" @mousedown.prevent="clearTimePopup(row, d.iso)">{{ $t('common.clear') }}</button>
                  </div>
                </div>
                <!-- Distance button -->
                <span
                  v-if="getEntry(row, d.iso)?.distance"
                  class="cell-dist-dot"
                  aria-hidden="true"
                ></span>
                <button v-if="!viewingOther"
                  class="cell-dist-toggle"
                  :class="{ 'cell-dist-on': !!getEntry(row, d.iso)?.distance }"
                  :aria-label="$t('timeTracking.set_distance')"
                  :title="$t('timeTracking.set_distance')"
                  @mousedown.prevent="openDistPopup(row, d.iso, $event)"
                  tabindex="-1"
                >⇆</button>
                <!-- Distance popup -->
                <div
                  v-if="distPopupKey === row.key + d.iso"
                  class="time-popup dist-popup"
                  :class="{ 'time-popup-below': distPopupFlip }"
                  ref="distPopupRef"
                  role="dialog"
                  :aria-label="$t('timeTracking.set_distance')"
                  @mousedown.stop
                >
                  <label class="tp-label">{{ $t('timeTracking.distance') }} ({{ distanceUnit }})
                    <input type="number" class="tp-inp" v-model="distPopupVal"
                      min="0" step="0.1"
                      @keydown.enter.prevent="applyDistPopup(row, d.iso)"
                      @keydown.escape.prevent="distPopupKey = ''"
                    />
                  </label>
                  <div class="tp-actions">
                    <button class="tp-btn tp-apply" @mousedown.prevent="applyDistPopup(row, d.iso)">{{ $t('common.save') }}</button>
                    <button v-if="getEntry(row, d.iso)?.distance" class="tp-btn tp-clear" @mousedown.prevent="clearDistPopup(row, d.iso)">{{ $t('common.clear') }}</button>
                  </div>
                </div>
                <div v-if="cellUndeclMins(row, d.iso) > 0" :id="'tt-undecl-' + idx + '-' + di" class="cell-undecl">
                  <span aria-hidden="true">-{{ fmtTime(cellUndeclMins(row, d.iso)) }}</span>
                  <span class="sr-only">{{ $t('timeTracking.undeclarable') }}: {{ fmtTime(cellUndeclMins(row, d.iso)) }}</span>
                </div>
              </td>
              <td class="c-total c-rowtotal" :aria-label="rowUndecl(row) > 0 ? rowDeclTotal(row) + ', ' + $t('timeTracking.undeclarable') + ': -' + fmtTime(rowUndecl(row)) : undefined">
                <span :aria-hidden="rowUndecl(row) > 0 ? 'true' : undefined">{{ rowDeclTotal(row) }}</span>
                <span v-if="rowUndecl(row) > 0" class="row-undecl-badge" aria-hidden="true">-{{ fmtTime(rowUndecl(row)) }}</span>
              </td>

              <!-- Actions -->
              <td class="c-act">
                <template v-if="!viewingOther">
                  <template v-if="editingRow === row.key">
                    <button class="act-btn act-ok" @click="confirmEditRow(row)" :title="$t('common.save')" aria-label="Save">✓</button>
                    <button class="act-btn act-no" @click="cancelEditRow" :title="$t('common.cancel')" aria-label="Cancel">✕</button>
                  </template>
                  <template v-else-if="deletingRow === row.key">
                    <button class="act-btn act-ok" @click="confirmDeleteRow(row)" :title="$t('common.yes')" aria-label="Confirm delete">✓</button>
                    <button class="act-btn act-no" @click="cancelDeleteRow" :title="$t('common.no')" aria-label="Cancel delete">✕</button>
                  </template>
                  <template v-else>
                    <button class="act-btn act-edit" @click="startEditRow(row)" :title="$t('common.edit')" aria-label="Edit time entry">✎</button>
                    <button class="act-btn act-duplicate" @click="duplicateRow(row)" :title="$t('common.duplicate')" aria-label="Duplicate time entry">⧉</button>
                    <button v-if="row.contract_id" class="act-btn act-slots" @click="fillFromSlots(row)" :title="$t('timeTracking.fill_from_slots')" :aria-label="$t('timeTracking.fill_from_slots')">⏱</button>
                    <button class="act-btn act-standby" @click="openStandbyShift(row)" :title="$t('timeTracking.standby_shift')" :aria-label="$t('timeTracking.standby_shift')">⏳</button>
                    <button class="act-btn act-del" @click="startDeleteRow(row)" :title="$t('common.delete')" aria-label="Delete time entry">🗑</button>
                  </template>
                </template>
              </td>
            </tr>

            <!-- Inline new-row editor -->
            <tr v-if="addingRow" class="tt-row tt-newrow">
              <td class="c-drag"></td>
              <td class="c-nr">{{ allRows.length + 1 }}</td>
              <td class="c-info">
                <select class="nr-sel" v-model="newRow.customer_id" @change="newRow.project_id = null" aria-label="Customer">
                  <option :value="null">{{ $t('timeTracking.no_customer') }}</option>
                  <option v-for="c in allCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select v-if="newRowContracts.length" class="nr-sel" v-model="newRow.contract_id" :aria-label="$t('timeTracking.contract')">
                  <option :value="null">{{ $t('timeTracking.no_contract') }}</option>
                  <option v-for="c in newRowContracts" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select class="nr-sel" v-model="newRow.project_id" aria-label="Project">
                  <option :value="null">{{ $t('timeTracking.no_project') }}</option>
                  <option v-for="p in newRowProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
              </td>
              <td class="c-desc">
                <input ref="newDescRef" class="nr-desc" type="text"
                  v-model="newRow.description"
                  :placeholder="$t('timeTracking.description')"
                  @keydown.enter="confirmNewRow"
                  @keydown.escape="cancelNewRow"
                />
              </td>
              <td class="c-rate"></td>
              <td v-for="d in weekDays" :key="d.iso" :class="['c-day', holidayDates.has(d.iso) ? 'c-day-holiday' : '', d.isToday ? 'c-day-today' : '']">
                <input class="h-inp" :type="timeNotation === 'hhmm' ? 'text' : 'number'" value="" disabled />
              </td>
              <td class="c-total"></td>
              <td class="c-act">
                <button class="act-btn act-ok" @click="confirmNewRow" :title="$t('common.save')" :aria-label="$t('common.save')">✓</button>
                <button class="act-btn act-no" @click="cancelNewRow" :title="$t('common.cancel')" :aria-label="$t('common.cancel')">✕</button>
                <button v-if="newRowHasSlots" class="act-btn act-slots" @click="confirmAndFillFromSlots" :title="$t('timeTracking.fill_from_slots')" :aria-label="$t('timeTracking.fill_from_slots')">⏱</button>
              </td>
            </tr>
          </tbody>

          <tfoot>
            <tr class="tt-foot">
              <td colspan="5" class="foot-lbl">{{ $t('timeTracking.total') }}</td>
              <td v-for="d in weekDays" :key="d.iso" :class="['c-day', 'c-total', 'c-dttotal', holidayDates.has(d.iso) ? 'c-day-holiday' : '', dayOverLimit(d.iso) ? 'c-day-over' : '', d.isToday ? 'c-day-today' : '']" :title="dayExpectedLabel(d.iso)" :aria-label="dayUndeclMins(d.iso) > 0 ? dayDeclTotal(d.iso) + ', ' + $t('timeTracking.undeclarable') + ': -' + dayUndecl(d.iso) : undefined">
                <span :aria-hidden="dayUndeclMins(d.iso) > 0 ? 'true' : undefined">{{ dayDeclTotal(d.iso) }}</span>
                <span v-if="dayUndeclMins(d.iso) > 0" class="foot-undecl-inline" aria-hidden="true">-{{ dayUndecl(d.iso) }}</span>
              </td>
              <td class="c-total grand-total" :aria-label="grandUndeclTotal > 0 ? fmtTime(grandDeclarable) + ', ' + $t('timeTracking.undeclarable') + ': -' + fmtTime(grandUndeclTotal) : undefined">
                <span :aria-hidden="grandUndeclTotal > 0 ? 'true' : undefined">{{ fmtTime(grandDeclarable) }}</span>
                <span v-if="grandUndeclTotal > 0" class="foot-undecl-inline" aria-hidden="true">-{{ fmtTime(grandUndeclTotal) }}</span>
              </td>
              <td class="c-act"></td>
            </tr>
            <tr v-if="weekDistanceTotal > 0" class="tt-foot tt-foot-dist">
              <td colspan="5" class="foot-lbl foot-dist-lbl">{{ $t('timeTracking.distance_total') }} ({{ distanceUnit }})</td>
              <td v-for="d in weekDays" :key="d.iso" class="c-day c-total c-dttotal">
                <span v-if="dayDistanceTotal(d.iso) > 0">{{ fmtDistance(dayDistanceTotal(d.iso)) }}</span>
              </td>
              <td class="c-total grand-total">{{ fmtDistance(weekDistanceTotal) }}</td>
              <td class="c-act"></td>
            </tr>
          </tfoot>
        </table>
      </div>

      <!-- Bottom bar: add row + exports -->
      <div class="tt-add-bar">
        <template v-if="!viewingOther && !addingRow">
          <button class="btn-add-row" @click="startAddRow">
            ＋ {{ $t('timeTracking.add_row') }}
          </button>
          <div class="macro-run-wrap" ref="macroRunRef">
            <button
              type="button"
              class="btn-copy-prev"
              :class="{ 'macro-run-open': macroRunOpen }"
              @click="toggleMacroRun"
              :aria-expanded="String(macroRunOpen)"
              aria-haspopup="dialog"
              aria-controls="macro-run-panel"
              :aria-label="$t('timeTracking.macro_run_aria')"
            >
              ⚡ {{ $t('timeTracking.macro_button') }}
            </button>
            <div
              v-if="macroRunOpen"
              id="macro-run-panel"
              class="macro-run-drop"
              role="dialog"
              aria-modal="true"
              aria-labelledby="macro-run-title"
            >
              <p id="macro-run-title" class="macro-run-title">{{ $t('timeTracking.macro_run') }}</p>
              <label class="form-label" for="macro-run-select">{{ $t('timeTracking.macro_label') }}</label>
              <select id="macro-run-select" class="form-input" v-model="macroRunId" :aria-label="$t('timeTracking.macro_run_select')">
                <option v-for="m in macrosSorted" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
              <label class="form-label" for="macro-run-start">{{ $t('timeTracking.macro_start_on') }}</label>
              <select id="macro-run-start" class="form-input" v-model.number="macroRunStartDay" aria-describedby="macro-run-preview-text">
                <option v-for="(d, i) in weekDays" :key="d.iso" :value="i">{{ d.abbr }} {{ d.mmdd }}</option>
              </select>
              <p id="macro-run-preview-text" class="macro-run-preview">{{ macroRunPreview }}</p>
              <button type="button" class="btn btn-primary macro-run-btn" @click="runMacroFromPopout">{{ $t('timeTracking.macro_run') }}</button>
            </div>
          </div>
          <button class="btn-copy-prev" @click="copyPrevWeek" :disabled="copyingPrevWeek" :aria-label="$t('timeTracking.copy_prev_week')">
            {{ copyingPrevWeek ? '…' : '⇐' }} {{ $t('timeTracking.copy_prev_week') }}
          </button>
          <button
            type="button"
            class="btn-undo"
            :disabled="undoStack.length === 0 || viewingOther"
            :title="$t('timeTracking.undo')"
            :aria-label="$t('timeTracking.undo')"
            @click="undoLastChange"
          >↶ {{ $t('timeTracking.undo') }}</button>
        </template>
        <template v-else-if="!viewingOther">
          <button class="btn btn-primary" @click="confirmNewRow">{{ $t('common.save') }}</button>
          <button class="btn btn-secondary" @click="cancelNewRow">{{ $t('common.cancel') }}</button>
        </template>
        <div class="tt-export-group" v-if="allRows.length > 0">
          <div class="pdf-font-group">
            <label class="filter-label">{{ $t('report.pdf_font') }}</label>
            <select class="form-input" v-model="pdfFont">
              <option value="inter">Inter</option>
              <option value="roboto">Roboto</option>
              <option value="opensans">Open Sans</option>
              <option value="sourcecode">Source Code Pro</option>
              <option value="freesans">{{ $t('report.pdf_font_freesans') }}</option>
              <option value="freeserif">{{ $t('report.pdf_font_freeserif') }}</option>
              <option value="freemono">{{ $t('report.pdf_font_freemono') }}</option>
            </select>
          </div>
          <div class="pdf-font-group">
            <label class="filter-label">{{ $t('report.pdf_lang') }}</label>
            <select class="form-input" v-model="pdfLang">
              <option value="auto">{{ $t('report.pdf_lang_auto') }}</option>
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
          <button class="btn btn-secondary" @click="exportSheetXLSX">{{ $t('timeTracking.export_xlsx') }}</button>
          <button class="btn btn-secondary" @click="exportSheetPDF">{{ $t('timeTracking.export_pdf') }}</button>
          <div class="pdf-options-wrapper" ref="gridPdfRef">
            <button class="pdf-options-btn" :class="{ 'is-active': gridPdfOpen }" @click="openGridPdf" :aria-expanded="String(gridPdfOpen)" aria-haspopup="dialog">
              {{ $t('timeTracking.export_grid') }}<span class="pdf-opts-chevron" :class="{ open: gridPdfOpen }" aria-hidden="true">›</span>
            </button>
            <div v-if="gridPdfOpen" class="pdf-options-panel grid-pdf-panel" role="dialog" :aria-label="$t('timeTracking.export_grid')">
              <div class="grid-pdf-section-label">{{ $t('report.period') }}</div>
              <div class="grid-pdf-type-row" role="group" :aria-label="$t('report.period')">
                <button class="grid-pdf-type-btn" :class="{ active: gridPdfType === 'week' }" @click="gridPdfType = 'week'" role="radio" :aria-checked="String(gridPdfType === 'week')">{{ $t('report.period_week') }}</button>
                <button class="grid-pdf-type-btn" :class="{ active: gridPdfType === 'month' }" @click="gridPdfType = 'month'" role="radio" :aria-checked="String(gridPdfType === 'month')">{{ $t('report.period_month') }}</button>
                <button class="grid-pdf-type-btn" :class="{ active: gridPdfType === 'year' }" @click="gridPdfType = 'year'" role="radio" :aria-checked="String(gridPdfType === 'year')">{{ $t('report.period_year') }}</button>
              </div>
              <div class="grid-pdf-date-row">
                <template v-if="gridPdfType === 'week'">
                  <select class="form-input pdf-option-select grid-pdf-select" v-model.number="gridPdfWeek" :aria-label="$t('timeTracking.week')">
                    <option v-for="w in 53" :key="w" :value="w">{{ $t('timeTracking.week_n', { n: w }) }}</option>
                  </select>
                  <select class="form-input pdf-option-select grid-pdf-select" v-model.number="gridPdfWeekYear" :aria-label="$t('report.year')">
                    <option v-for="y in gridPdfYears" :key="y" :value="y">{{ y }}</option>
                  </select>
                </template>
                <template v-else-if="gridPdfType === 'month'">
                  <select class="form-input pdf-option-select grid-pdf-select" v-model.number="gridPdfMonth" :aria-label="$t('report.period_month')">
                    <option v-for="(name, idx) in gridPdfMonthNames" :key="idx + 1" :value="idx + 1">{{ name }}</option>
                  </select>
                  <select class="form-input pdf-option-select grid-pdf-select" v-model.number="gridPdfMonthYear" :aria-label="$t('report.year')">
                    <option v-for="y in gridPdfYears" :key="y" :value="y">{{ y }}</option>
                  </select>
                </template>
                <template v-else>
                  <select class="form-input pdf-option-select grid-pdf-select grid-pdf-select-full" v-model.number="gridPdfYear" :aria-label="$t('report.period_year')">
                    <option v-for="y in gridPdfYears" :key="y" :value="y">{{ y }}</option>
                  </select>
                </template>
              </div>
              <div class="pdf-options-divider" role="separator"></div>
              <div class="grid-pdf-actions">
                <button class="btn btn-primary grid-pdf-export-btn" @click="exportGridPDFFromPanel">{{ $t('timeTracking.export_grid') }}</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <BaseModal
      v-if="macroEditorOpen"
      :title="$t('timeTracking.macro_edit')"
      :resizable="true"
      style="--modal-width: 1180px"
      @close="macroEditorOpen = false"
    >
      <p class="macro-hint">{{ $t('timeTracking.macro_hint') }}</p>
      <div class="macro-toolbar">
        <div class="macro-toolbar-group">
          <label class="form-label" for="macro-select">{{ $t('timeTracking.macro_label') }}</label>
          <select id="macro-select" class="form-input" v-model="selectedMacroId" :aria-label="$t('timeTracking.macro_select')">
            <option v-for="m in macrosSorted" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
        <div class="macro-toolbar-actions">
          <button class="btn btn-secondary" @click="createMacroTemplate">{{ $t('timeTracking.macro_new') }}</button>
          <button class="btn btn-secondary" @click="duplicateMacroTemplate">{{ $t('timeTracking.macro_duplicate') }}</button>
          <button class="btn btn-secondary" :disabled="macroLibrary.macros.length <= 1" @click="deleteMacroTemplate">{{ $t('timeTracking.macro_delete') }}</button>
          <button class="btn btn-secondary" @click="exportMacroLibrary">{{ $t('timeTracking.macro_export_json') }}</button>
          <button class="btn btn-secondary" @click="openMacroImport">{{ $t('timeTracking.macro_import_json') }}</button>
          <input ref="macroImportRef" class="sr-only" type="file" accept="application/json,.json" @change="importMacroLibrary" />
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="macro-name">{{ $t('timeTracking.macro_name') }}</label>
        <input id="macro-name" class="form-input" type="text" v-model="activeMacro.name" />
      </div>
      <div class="macro-apply-bar">
        <div class="macro-apply-field">
          <label class="form-label" for="macro-apply-days">{{ $t('timeTracking.macro_apply_days') }}</label>
          <input
            id="macro-apply-days"
            class="form-input macro-days-input"
            type="number"
            min="1"
            max="7"
            v-model.number="activeMacro.apply_days"
            aria-describedby="macro-apply-preview"
          />
        </div>
        <div class="macro-apply-field macro-apply-toggle">
          <input id="macro-alternating" type="checkbox" v-model="activeMacro.alternating" />
          <label for="macro-alternating">{{ $t('timeTracking.macro_alternating') }}</label>
        </div>
      </div>
      <p id="macro-apply-preview" class="macro-apply-preview">{{ macroApplyPreview }}</p>
      <div
        class="macro-grid"
        :class="activeMacro.alternating ? 'macro-grid--alt' : 'macro-grid--single'"
        role="table"
        :aria-label="activeMacro.alternating ? $t('timeTracking.macro_grid_alt_aria') : $t('timeTracking.macro_grid_single_aria')"
      >
        <div class="macro-head" :class="activeMacro.alternating ? 'macro-head--alt' : 'macro-head--single'" role="row">
          <span>{{ $t('timeTracking.macro_col_customer') }}</span>
          <span>{{ $t('timeTracking.macro_col_project') }}</span>
          <span>{{ $t('timeTracking.macro_col_activity') }}</span>
          <span>{{ activeMacro.alternating ? $t('timeTracking.macro_col_pattern_a') : $t('timeTracking.macro_col_hours') }}</span>
          <span>{{ activeMacro.alternating ? $t('timeTracking.macro_col_a_start') : $t('timeTracking.macro_col_start') }}</span>
          <span>{{ activeMacro.alternating ? $t('timeTracking.macro_col_a_end') : $t('timeTracking.macro_col_end') }}</span>
          <template v-if="activeMacro.alternating">
            <span>{{ $t('timeTracking.macro_col_pattern_b') }}</span>
            <span>{{ $t('timeTracking.macro_col_b_start') }}</span>
            <span>{{ $t('timeTracking.macro_col_b_end') }}</span>
            <span>{{ $t('timeTracking.macro_col_a_km') }}</span>
            <span>{{ $t('timeTracking.macro_col_b_km') }}</span>
          </template>
          <span v-else>{{ $t('timeTracking.macro_col_distance') }}</span>
          <span>{{ $t('timeTracking.macro_col_action') }}</span>
        </div>
        <div
          v-for="(row, idx) in activeMacro.rows"
          :key="idx"
          class="macro-row"
          :class="activeMacro.alternating ? 'macro-row--alt' : 'macro-row--single'"
          role="row"
        >
          <label class="sr-only" :for="'macro-customer-' + idx">{{ $t('timeTracking.macro_row_customer', { n: idx + 1 }) }}</label>
          <select :id="'macro-customer-' + idx" class="form-input" v-model="row.customer_id" @change="onMacroCustomerChange(row)">
            <option :value="null">{{ $t('timeTracking.no_customer') }}</option>
            <option v-for="c in allCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <label class="sr-only" :for="'macro-project-' + idx">{{ $t('timeTracking.macro_row_project', { n: idx + 1 }) }}</label>
          <select :id="'macro-project-' + idx" class="form-input" v-model="row.project_id">
            <option :value="null">{{ $t('timeTracking.no_project') }}</option>
            <option v-for="p in macroProjectsForRow(row)" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <label class="sr-only" :for="'macro-activity-' + idx">{{ $t('timeTracking.macro_row_activity', { n: idx + 1 }) }}</label>
          <input :id="'macro-activity-' + idx" class="form-input" type="text" v-model="row.description" />
          <label class="sr-only" :for="'macro-day1-min-' + idx">{{ (activeMacro.alternating ? $t('timeTracking.macro_pattern_a_time') : $t('timeTracking.macro_col_hours')) + ' ' + (idx + 1) }}</label>
          <input :id="'macro-day1-min-' + idx" class="form-input" type="text" v-model="row.day1_minutes" :placeholder="timeNotation === 'hhmm' ? '1:30' : '1.5'" />
          <label class="sr-only" :for="'macro-day1-start-' + idx">{{ (activeMacro.alternating ? $t('timeTracking.macro_pattern_a_start') : $t('timeTracking.macro_col_start')) + ' ' + (idx + 1) }}</label>
          <input :id="'macro-day1-start-' + idx" class="form-input" type="text" maxlength="5" v-model="row.day1_start" placeholder="09:00" />
          <label class="sr-only" :for="'macro-day1-end-' + idx">{{ (activeMacro.alternating ? $t('timeTracking.macro_pattern_a_end') : $t('timeTracking.macro_col_end')) + ' ' + (idx + 1) }}</label>
          <input :id="'macro-day1-end-' + idx" class="form-input" type="text" maxlength="5" v-model="row.day1_end" placeholder="17:00" />
          <template v-if="activeMacro.alternating">
            <label class="sr-only" :for="'macro-day2-min-' + idx">{{ $t('timeTracking.macro_pattern_b_time') }} {{ idx + 1 }}</label>
            <input :id="'macro-day2-min-' + idx" class="form-input" type="text" v-model="row.day2_minutes" :placeholder="timeNotation === 'hhmm' ? '1:30' : '1.5'" />
            <label class="sr-only" :for="'macro-day2-start-' + idx">{{ $t('timeTracking.macro_pattern_b_start') }} {{ idx + 1 }}</label>
            <input :id="'macro-day2-start-' + idx" class="form-input" type="text" maxlength="5" v-model="row.day2_start" placeholder="09:00" />
            <label class="sr-only" :for="'macro-day2-end-' + idx">{{ $t('timeTracking.macro_pattern_b_end') }} {{ idx + 1 }}</label>
            <input :id="'macro-day2-end-' + idx" class="form-input" type="text" maxlength="5" v-model="row.day2_end" placeholder="17:00" />
            <label class="sr-only" :for="'macro-day1-dist-' + idx">{{ $t('timeTracking.macro_pattern_a_distance') }} {{ idx + 1 }}</label>
            <input :id="'macro-day1-dist-' + idx" class="form-input" type="number" min="0" step="0.1" v-model="row.day1_distance" />
            <label class="sr-only" :for="'macro-day2-dist-' + idx">{{ $t('timeTracking.macro_pattern_b_distance') }} {{ idx + 1 }}</label>
            <input :id="'macro-day2-dist-' + idx" class="form-input" type="number" min="0" step="0.1" v-model="row.day2_distance" />
          </template>
          <template v-else>
            <label class="sr-only" :for="'macro-day1-dist-' + idx">{{ $t('timeTracking.macro_col_distance') }} {{ idx + 1 }}</label>
            <input :id="'macro-day1-dist-' + idx" class="form-input" type="number" min="0" step="0.1" v-model="row.day1_distance" />
          </template>
          <button class="btn btn-secondary macro-row-remove" :disabled="activeMacro.rows.length <= 1" @click="removeMacroRow(idx)" :aria-label="$t('timeTracking.macro_row_remove', { n: idx + 1 })">{{ $t('timeTracking.macro_row_remove_btn') }}</button>
        </div>
      </div>
      <div class="macro-actions">
        <button class="btn btn-secondary" @click="addMacroRow">{{ $t('timeTracking.macro_add_row') }}</button>
      </div>
      <template #footer>
        <button class="btn" @click="macroEditorOpen = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveMacroTemplate">{{ $t('timeTracking.macro_save') }}</button>
      </template>
    </BaseModal>

    <BaseModal
      v-if="standbyRow"
      :title="$t('timeTracking.standby_shift_title')"
      :resizable="true"
      style="--modal-width: 480px"
      @close="closeStandbyShift"
    >
      <p class="standby-hint">{{ $t('timeTracking.standby_hint') }}</p>
      <div class="form-group">
        <label class="form-label" for="standby-start-date">{{ $t('timeTracking.standby_start_date') }}</label>
        <div class="date-input-row">
          <input id="standby-start-date" class="form-input" type="text" v-model="displayStandbyStartDate" :placeholder="dateOnlyFormat()" @blur="parseStandbyStartDate" />
          <label class="picker-wrap" :title="$t('common.pick_date')">
            <span class="btn-icon-xs" aria-hidden="true">&#128197;</span>
            <input type="date" class="date-picker-overlay" :value="standbyForm.start_date" :aria-label="$t('timeTracking.standby_start_date')" @change="onStandbyStartDateChange" />
          </label>
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="standby-start-time">{{ $t('timeTracking.start_time') }}</label>
        <input id="standby-start-time" class="form-input" type="text" v-model="standbyForm.start_time" placeholder="19:00" maxlength="5" @input="onStandbyTimeInput('start', $event)" />
      </div>
      <div class="form-group">
        <label class="form-label" for="standby-end-date">{{ $t('timeTracking.standby_end_date') }}</label>
        <div class="date-input-row">
          <input id="standby-end-date" class="form-input" type="text" v-model="displayStandbyEndDate" :placeholder="dateOnlyFormat()" @blur="parseStandbyEndDate" />
          <label class="picker-wrap" :title="$t('common.pick_date')">
            <span class="btn-icon-xs" aria-hidden="true">&#128197;</span>
            <input type="date" class="date-picker-overlay" :value="standbyForm.end_date" :aria-label="$t('timeTracking.standby_end_date')" @change="onStandbyEndDateChange" />
          </label>
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="standby-end-time">{{ $t('timeTracking.end_time') }}</label>
        <input id="standby-end-time" class="form-input" type="text" v-model="standbyForm.end_time" placeholder="07:00" maxlength="5" @input="onStandbyTimeInput('end', $event)" />
      </div>
      <button type="button" class="btn btn-sm standby-preset" @click="applyStandbyPreset">{{ $t('timeTracking.standby_preset_weekend') }}</button>
      <p v-if="standbyPreview" class="standby-preview">{{ standbyPreview }}</p>
      <template #footer>
        <button class="btn" @click="closeStandbyShift">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!standbyPreview || savingStandby" @click="applyStandbyShift">{{ savingStandby ? '…' : $t('timeTracking.standby_apply') }}</button>
      </template>
    </BaseModal>

    <ContractRatesModal v-if="showRates" @close="showRates = false" />

    <BaseModal
      v-if="customRangeModalOpen"
      :title="$t('timeTracking.custom_range_title')"
      style="--modal-width: 420px"
      @close="customRangeModalOpen = false"
    >
      <div class="form-row custom-range-row">
        <div class="form-group">
          <span class="form-label" id="rpt-range-start-label">{{ $t('timeTracking.custom_range_start') }}</span>
          <DatePicker :label="$t('timeTracking.custom_range_start')" v-model="customRangeForm.start" />
        </div>
        <div class="form-group">
          <span class="form-label" id="rpt-range-end-label">{{ $t('timeTracking.custom_range_end') }}</span>
          <DatePicker :label="$t('timeTracking.custom_range_end')" v-model="customRangeForm.end" />
        </div>
      </div>
      <p v-if="customRangeError" class="field-error" role="alert">{{ customRangeError }}</p>
      <template #footer>
        <button class="btn" @click="customRangeModalOpen = false">{{ $t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="!customRangeForm.start || !customRangeForm.end" @click="applyCustomRange">{{ $t('common.save') }}</button>
      </template>
    </BaseModal>

    <!-- ── Report ──────────────────────────────────────────────────────────── -->
    <div v-if="mode === 'report'" id="panel-report" role="tabpanel" aria-labelledby="tab-report" class="tt-report-outer">
      <div class="report-filters">
        <select class="form-input fi-sm" v-model="rpt.period" @change="onReportPeriodChange">
          <option value="week">{{ $t('report.week') }}</option>
          <option value="month">{{ $t('report.month') }}</option>
          <option value="year">{{ $t('report.year') }}</option>
          <option value="custom">{{ $t('report.period_custom') }}</option>
        </select>
        <input v-if="rpt.period !== 'custom'" type="number" class="form-input fi-sm fi-year"
          v-model.number="rpt.year" min="2000" max="2100" @change="loadReport" />
        <select v-if="rpt.period === 'month'" class="form-input fi-sm"
          v-model.number="rpt.month" @change="loadReport">
          <option v-for="(m, i) in months" :key="i" :value="i + 1">{{ m }}</option>
        </select>
        <select v-if="rpt.period === 'week'" class="form-input fi-sm"
          v-model.number="rpt.week" @change="loadReport">
          <option v-for="w in 53" :key="w" :value="w">
            {{ $t('timeTracking.week') }} {{ w }}
          </option>
        </select>
        <button v-if="rpt.period === 'custom'" type="button" class="btn btn-secondary fi-sm"
          @click="openCustomRangeModal" :aria-label="$t('timeTracking.custom_range_pick')">
          {{ customRangeButtonLabel }}
        </button>
        <div class="rpt-filter-group">
          <label class="filter-label" for="rpt-group-by">{{ $t('timeTracking.group_by') }}<HelpIcon i18n-key="help.fields.report_group_by" :label="$t('timeTracking.group_by')" /></label>
          <select id="rpt-group-by" class="form-input fi-sm" v-model="rpt.group_by" @change="loadReport">
            <option value="period">{{ $t('timeTracking.group_by_period') }}</option>
            <option value="customer">{{ $t('timeTracking.group_by_customer') }}</option>
            <option value="project">{{ $t('timeTracking.group_by_project') }}</option>
            <option value="customer_project">{{ $t('timeTracking.group_by_customer_project') }}</option>
          </select>
        </div>
        <button class="btn btn-secondary" @click="loadReport">{{ $t('timeTracking.refresh') }}</button>
        <div class="rpt-view-toggle" role="group" :aria-label="$t('timeTracking.view_table') + ' / ' + $t('timeTracking.view_chart')">
          <button type="button" class="rpt-view-btn" :class="{ active: reportViewMode === 'table' }"
            role="radio" :aria-checked="String(reportViewMode === 'table')" @click="reportViewMode = 'table'">
            📋 {{ $t('timeTracking.view_table') }}
          </button>
          <button type="button" class="rpt-view-btn" :class="{ active: reportViewMode === 'chart' }"
            role="radio" :aria-checked="String(reportViewMode === 'chart')" @click="reportViewMode = 'chart'">
            📊 {{ $t('timeTracking.view_chart') }}
          </button>
        </div>
        <div class="rpt-filter-group" v-if="reportViewMode === 'chart'">
          <label class="filter-label" for="rpt-chart-type">{{ $t('timeTracking.chart_type') }}</label>
          <select id="rpt-chart-type" class="form-input fi-sm" v-model="reportChartType">
            <option value="bar">{{ $t('timeTracking.chart_bar') }}</option>
            <option value="pie">{{ $t('timeTracking.chart_pie') }}</option>
            <option value="stacked">{{ $t('timeTracking.chart_stacked_bar') }}</option>
          </select>
        </div>
        <div class="rpt-filter-group" v-if="reportViewMode === 'chart'">
          <label class="filter-label" for="rpt-chart-basis">{{ $t('timeTracking.time_basis') }}</label>
          <select id="rpt-chart-basis" class="form-input fi-sm" v-model="reportChartBasis">
            <option value="declarable">{{ $t('timeTracking.time_basis_declarable') }}</option>
            <option value="total">{{ $t('timeTracking.time_basis_total') }}</option>
            <option value="undeclarable">{{ $t('timeTracking.time_basis_undeclarable') }}</option>
          </select>
        </div>
        <div class="tt-export-group" v-if="report && report.total_minutes > 0 && reportViewMode === 'chart'">
          <button class="btn btn-secondary" @click="exportReportChartPDF">{{ $t('timeTracking.export_pdf') }}</button>
          <HelpIcon i18n-key="help.fields.report_export" :label="$t('timeTracking.export_pdf')" />
        </div>
        <div class="tt-export-group" v-if="report && report.total_minutes > 0 && reportViewMode === 'table'">
          <div class="pdf-options-wrapper" ref="pdfOptionsRef">
            <button
              class="pdf-options-btn"
              :class="{ 'is-active': pdfShowAbbr || (rpt.group_by === 'customer' && pdfPageBreak) || !pdfShowPageNumbers || !pdfShowUndeclarable }"
              @click="pdfOptionsOpen = !pdfOptionsOpen"
              :aria-expanded="String(pdfOptionsOpen)"
              aria-haspopup="true"
              :aria-label="$t('timeTracking.pdf_export_options')"
            >
              {{ $t('timeTracking.pdf_export_options') }}<span class="pdf-opts-chevron" :class="{ open: pdfOptionsOpen }">›</span>
            </button>
            <div v-if="pdfOptionsOpen" class="pdf-options-panel opens-down" role="menu">
              <div class="pdf-option-selects">
                <label class="pdf-option-select-label" :for="'pdf-opt-font'">{{ $t('report.pdf_font') }}</label>
                <select id="pdf-opt-font" class="form-input pdf-option-select" v-model="pdfFont">
                  <option value="inter">Inter</option>
                  <option value="roboto">Roboto</option>
                  <option value="opensans">Open Sans</option>
                  <option value="sourcecode">Source Code Pro</option>
                  <option value="freesans">{{ $t('report.pdf_font_freesans') }}</option>
                  <option value="freeserif">{{ $t('report.pdf_font_freeserif') }}</option>
                  <option value="freemono">{{ $t('report.pdf_font_freemono') }}</option>
                </select>
                <label class="pdf-option-select-label" :for="'pdf-opt-lang'">{{ $t('report.pdf_lang') }}</label>
                <select id="pdf-opt-lang" class="form-input pdf-option-select" v-model="pdfLang">
                  <option value="auto">{{ $t('report.pdf_lang_auto') }}</option>
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
              <div class="pdf-options-divider" role="separator"></div>
              <label class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfShowAbbr)">
                <input type="checkbox" v-model="pdfShowAbbr" />
                {{ $t('timeTracking.pdf_show_abbr') }}
              </label>
              <label v-if="rpt.group_by === 'customer'" class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfPageBreak)">
                <input type="checkbox" v-model="pdfPageBreak" />
                {{ $t('timeTracking.pdf_page_break_customer') }}
              </label>
              <label class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfShowCosts)">
                <input type="checkbox" v-model="pdfShowCosts" />
                {{ $t('timeTracking.show_costs') }}
              </label>
              <label class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfShowPageNumbers)">
                <input type="checkbox" v-model="pdfShowPageNumbers" />
                {{ $t('timeTracking.pdf_show_page_numbers') }}
              </label>
              <label class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfShowUndeclarable)">
                <input type="checkbox" v-model="pdfShowUndeclarable" />
                {{ $t('timeTracking.pdf_show_undeclarable') }}
              </label>
              <label v-if="report && report.undeclarable_minutes > 0" class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfShowUndeclarableRow)">
                <input type="checkbox" v-model="pdfShowUndeclarableRow" />
                {{ $t('timeTracking.pdf_show_undeclarable_row') }}
              </label>
              <label v-if="report && report.total_distance > 0" class="pdf-option-item" role="menuitemcheckbox" :aria-checked="String(pdfShowDistance)">
                <input type="checkbox" v-model="pdfShowDistance" />
                {{ $t('timeTracking.pdf_show_distance') }}
              </label>
            </div>
          </div>
          <button class="btn btn-secondary" @click="exportReportXLSX">{{ $t('timeTracking.export_xlsx') }}</button>
          <button class="btn btn-secondary" @click="exportReportPDF">{{ $t('timeTracking.export_pdf') }}</button>
          <HelpIcon i18n-key="help.fields.report_export" :label="$t('timeTracking.export_pdf')" />
        </div>
      </div>

      <div v-if="loadingReport" class="tt-loading">{{ $t('common.loading') }}</div>
      <template v-else-if="report">
        <!-- Report header: logo + company name -->
        <div class="rpt-header">
          <div class="rpt-header-left">
            <img v-if="report.company_logo" :src="resolveAssetUrl(report.company_logo)" alt="" class="rpt-logo" @error="report.company_logo = ''" />
          </div>
          <div class="rpt-header-center">
            <div v-if="report.company_name" class="rpt-company-name">{{ report.company_name }}</div>
            <div class="rpt-title-label">{{ $t('timeTracking.title') }}</div>
            <div class="rpt-period-label">{{ report.period_label }}</div>
          </div>
          <div class="rpt-header-right"></div>
        </div>
        <div v-if="report.groups.length === 0" class="rpt-empty">{{ $t('timeTracking.no_entries') }}</div>
        <template v-else-if="reportViewMode === 'table'">
          <div v-for="grp in report.groups" :key="grp.label" class="rpt-group">
            <div class="rpt-group-hd">
              <span>{{ grp.label }}</span>
              <span class="rpt-grp-total">{{ fmtTime(grp.declarable_minutes) }}</span>
              <button
                v-if="rpt.group_by === 'customer' && grp.entries.some(e => e.contract_id)"
                class="btn btn-sm rpt-invoice-btn"
                @click="openCreateInvoiceFromGroup(grp)"
              >{{ $t('timeTracking.create_invoice') }}</button>
            </div>
            <table v-if="grp.entries.length" class="rpt-table">
              <colgroup>
                <col class="rpt-col-date" />
                <col class="rpt-col-customer" />
                <col class="rpt-col-project" />
                <col class="rpt-col-activity" />
                <col v-if="report.total_distance > 0" class="rpt-col-distance" />
                <col v-if="report.undeclarable_minutes > 0" class="rpt-col-undeclarable" />
                <col class="rpt-col-time" />
              </colgroup>
              <thead>
                <tr>
                  <th>{{ $t('timeTracking.date') }}</th>
                  <th>{{ $t('timeTracking.customer') }}</th>
                  <th>{{ $t('timeTracking.project') }}</th>
                  <th>{{ $t('timeTracking.activity') }}</th>
                  <th v-if="report.total_distance > 0" class="rpt-th-time">{{ $t('timeTracking.distance') }} ({{ distanceUnit }})</th>
                  <th v-if="report.undeclarable_minutes > 0" class="rpt-th-time">{{ $t('timeTracking.undeclarable') }}</th>
                  <th class="rpt-th-time">{{ $t('timeTracking.time') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="e in grp.entries" :key="e.id">
                  <td>{{ formatDate(e.date) }}</td>
                  <td>{{ e.customer?.name || '—' }}</td>
                  <td>{{ e.project?.name || '—' }}</td>
                  <td>{{ e.description || '—' }}</td>
                  <td v-if="report.total_distance > 0" class="rpt-th-time">{{ e.distance ? fmtDistance(e.distance) : '' }}</td>
                  <td v-if="report.undeclarable_minutes > 0" class="rpt-th-time">{{ entryUndecl(e) > 0 ? fmtTime(entryUndecl(e)) : '' }}</td>
                  <td class="rpt-th-time">{{ fmtTime(reportEntryDeclarable(e)) }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="rpt-grp-empty">{{ $t('timeTracking.no_entries_group') }}</div>
            <div v-if="grp.undeclarable_minutes > 0" class="rpt-grp-undecl-line">
              <span>{{ $t('timeTracking.undeclarable') }}</span>
              <span>-{{ fmtTime(grp.undeclarable_minutes) }}</span>
            </div>
          </div>
        </template>
        <div v-else class="rpt-chart-panel">
          <h3 class="rpt-chart-title">{{ $t('timeTracking.chart_by_activity') }}</h3>
          <div v-if="reportChartType === 'stacked' && loadingReportChartPeriod" class="tt-loading">{{ $t('common.loading') }}</div>
          <div v-else-if="!reportChartHasData" class="rpt-chart-empty">{{ $t('timeTracking.chart_no_data') }}</div>
          <template v-else>
            <div class="rpt-chart-wrap" :class="{ 'rpt-chart-wrap-pie': reportChartType === 'pie' }">
              <canvas ref="reportChartCanvas"></canvas>
            </div>
            <table v-if="reportChartType !== 'stacked'" class="sr-only">
              <caption>{{ $t('timeTracking.chart_by_activity') }}</caption>
              <thead><tr><th>{{ $t('timeTracking.activity') }}</th><th>{{ $t('timeTracking.time') }}</th></tr></thead>
              <tbody>
                <tr v-for="a in reportActivityBreakdown" :key="a.label">
                  <td>{{ a.label }}</td>
                  <td>{{ fmtTime(a.minutes) }}</td>
                </tr>
              </tbody>
            </table>
            <table v-else class="sr-only">
              <caption>{{ $t('timeTracking.chart_by_activity') }}</caption>
              <thead>
                <tr>
                  <th>{{ $t('report.period') }}</th>
                  <th v-for="s in reportStackedBreakdown.series" :key="s.label">{{ s.label }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(p, i) in reportStackedBreakdown.periods" :key="p">
                  <td>{{ p }}</td>
                  <td v-for="s in reportStackedBreakdown.series" :key="s.label">{{ fmtTime(Math.round(s.data[i] * 60)) }}</td>
                </tr>
              </tbody>
            </table>
          </template>
        </div>
        <div class="rpt-grand-total">
          <span>{{ $t('timeTracking.total') }}</span>
          <span>{{ fmtTime(report.undeclarable_minutes > 0 ? report.declarable_minutes : report.total_minutes) }}</span>
        </div>
        <div v-if="report.undeclarable_minutes > 0" class="rpt-undeclarable">
          <span>{{ $t('timeTracking.undeclarable') }}</span>
          <span>-{{ fmtTime(report.undeclarable_minutes) }}</span>
        </div>
        <div v-if="report.total_distance > 0" class="rpt-distance-total">
          <span>{{ $t('timeTracking.distance_total') }} ({{ distanceUnit }})</span>
          <span>{{ fmtDistance(report.total_distance) }}</span>
        </div>
      </template>
    </div>

    <!-- ── Board Report ──────────────────────────────────────────────────────── -->
    <div v-if="mode === 'board-report'" id="panel-board-report" role="tabpanel" aria-labelledby="tab-board-report" class="tt-board-rpt-outer">
      <BoardReportPanel />
    </div>

    <!-- ── Invoices ──────────────────────────────────────────────────────────── -->
    <div v-if="mode === 'invoices'" id="panel-invoices" role="tabpanel" aria-labelledby="tab-invoices" class="tt-invoices-outer">
      <div class="inv-toolbar">
        <select class="form-input fi-sm" v-model="invFilterCustomer" @change="loadAllInvoices">
          <option value="">{{ $t('invoice.all_customers') }}</option>
          <option v-for="c in allCustomers" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select class="form-input fi-sm" v-model="invFilterStatus" @change="loadAllInvoices">
          <option value="">{{ $t('invoice.all_statuses') }}</option>
          <option value="draft">{{ $t('invoice.status_draft') }}</option>
          <option value="sent">{{ $t('invoice.status_sent') }}</option>
          <option value="paid">{{ $t('invoice.status_paid') }}</option>
        </select>
        <button class="btn btn-secondary" @click="loadAllInvoices">{{ $t('timeTracking.refresh') }}</button>
      </div>
      <div v-if="loadingAllInvoices" class="tt-loading">{{ $t('common.loading') }}</div>
      <div v-else-if="allInvoices.length === 0" class="rpt-empty">{{ $t('invoice.no_invoices') }}</div>
      <table v-else class="inv-table">
        <thead>
          <tr>
            <th>{{ $t('invoice.invoice_number') }}</th>
            <th>{{ $t('invoice.customer') }}</th>
            <th>{{ $t('invoice.period') }}</th>
            <th>{{ $t('invoice.invoice_date') }}</th>
            <th>{{ $t('invoice.due_date') }}</th>
            <th>{{ $t('invoice.filter_status') }}</th>
            <th class="inv-col-num">{{ $t('invoice.total') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="inv in allInvoices" :key="inv.id" :class="`inv-row-${inv.status}`">
            <td class="inv-num-cell">{{ inv.invoice_number }}</td>
            <td>{{ inv.customer?.name || '—' }}</td>
            <td class="inv-muted">{{ formatDate(inv.period_start) }} – {{ formatDate(inv.period_end) }}</td>
            <td class="inv-muted">{{ formatDate(inv.created_at) }}</td>
            <td class="inv-muted">{{ inv.due_date ? formatDate(inv.due_date) : '—' }}</td>
            <td><span :class="['inv-status-badge', `inv-${inv.status}`]">{{ $t(`invoice.status_${inv.status}`) }}</span></td>
            <td class="inv-col-num inv-total-cell">{{ inv.currency }} {{ inv.total.toFixed(2) }}</td>
            <td class="inv-actions-cell">
              <a :href="invPdfUrl(inv)" target="_blank" class="btn btn-sm">PDF</a>
              <button v-if="inv.status === 'draft'" class="btn btn-sm" @click="changeAllInvStatus(inv, 'sent')">{{ $t('invoice.mark_sent') }}</button>
              <button v-if="inv.status === 'sent'" class="btn btn-sm btn-primary" @click="changeAllInvStatus(inv, 'paid')">{{ $t('invoice.mark_paid') }}</button>
              <button v-if="inv.status !== 'draft'" class="btn btn-sm" @click="changeAllInvStatus(inv, 'draft')">{{ $t('invoice.mark_draft') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

  </div>

  <!-- ── Create Invoice from report group modal ───────────────────────── -->
  <Teleport to="body">
    <div v-if="showCreateInvoiceFromGroup" class="tt-modal-backdrop" @click.self="showCreateInvoiceFromGroup = false" @keydown.escape="showCreateInvoiceFromGroup = false">
      <div class="tt-modal tt-modal-lg" role="dialog" aria-modal="true" aria-labelledby="cinv-grp-title">
        <div class="tt-modal-hd">
          <h2 id="cinv-grp-title" class="tt-modal-title">{{ $t('invoice.new_invoice') }}</h2>
          <button class="tt-modal-close" @click="showCreateInvoiceFromGroup = false" :aria-label="$t('common.close')">✕</button>
        </div>
        <!-- Step 1: form -->
        <div v-if="createInvStep === 1" class="tt-modal-body">
          <div class="cinv-field-row">
            <div class="cinv-field">
              <label for="cinv-grp-from">{{ $t('invoice.period_start') }}</label>
              <div class="date-input-wrap">
                <input id="cinv-grp-from" class="form-input" type="text" v-model="createInvDisplayFrom"
                  :placeholder="dateOnlyFormat()" @blur="parseCreateInvFrom" />
                <input type="date" tabindex="-1" aria-hidden="true" class="date-hidden-picker"
                  :value="createInvForm.period_start"
                  @change="e => { createInvForm.period_start = e.target.value; createInvDisplayFrom = formatDate(e.target.value) }" />
                <button v-if="createInvDisplayFrom" class="btn-icon-xs" @click="createInvDisplayFrom = ''; createInvForm.period_start = ''" :aria-label="$t('common.clear')">✕</button>
              </div>
            </div>
            <div class="cinv-field">
              <label for="cinv-grp-to">{{ $t('invoice.period_end') }}</label>
              <div class="date-input-wrap">
                <input id="cinv-grp-to" class="form-input" type="text" v-model="createInvDisplayTo"
                  :placeholder="dateOnlyFormat()" @blur="parseCreateInvTo" />
                <input type="date" tabindex="-1" aria-hidden="true" class="date-hidden-picker"
                  :value="createInvForm.period_end"
                  @change="e => { createInvForm.period_end = e.target.value; createInvDisplayTo = formatDate(e.target.value) }" />
                <button v-if="createInvDisplayTo" class="btn-icon-xs" @click="createInvDisplayTo = ''; createInvForm.period_end = ''" :aria-label="$t('common.clear')">✕</button>
              </div>
            </div>
          </div>
          <div class="cinv-field-row">
            <div class="cinv-field">
              <label for="cinv-grp-due">{{ $t('invoice.due_date') }}</label>
              <div class="date-input-wrap">
                <input id="cinv-grp-due" class="form-input" type="text" v-model="createInvDisplayDue"
                  :placeholder="dateOnlyFormat()" @blur="parseCreateInvDue" />
                <input type="date" tabindex="-1" aria-hidden="true" class="date-hidden-picker"
                  :value="createInvForm.due_date"
                  @change="e => { createInvForm.due_date = e.target.value; createInvDisplayDue = formatDate(e.target.value) }" />
                <button v-if="createInvDisplayDue" class="btn-icon-xs" @click="createInvDisplayDue = ''; createInvForm.due_date = ''" :aria-label="$t('common.clear')">✕</button>
              </div>
            </div>
            <div class="cinv-field">
              <label for="cinv-grp-vat">{{ $t('invoice.vat_rate') }} (%)</label>
              <input id="cinv-grp-vat" class="form-input" type="number" min="0" max="100" step="0.1" v-model.number="createInvForm.vat_rate" />
            </div>
          </div>
          <div class="cinv-field">
            <label for="cinv-grp-notes">{{ $t('invoice.notes') }}</label>
            <textarea id="cinv-grp-notes" class="form-input" rows="2" v-model="createInvForm.notes"></textarea>
          </div>
        </div>
        <!-- Step 2: preview line items -->
        <div v-else class="tt-modal-body">
          <div v-if="createInvLineItems.length === 0" class="rpt-empty">{{ $t('invoice.no_billable_entries') }}</div>
          <div v-else class="inv-preview-table-wrap">
            <table class="inv-preview-table">
              <thead>
                <tr>
                  <th>{{ $t('invoice.line_date') }}</th>
                  <th>{{ $t('invoice.line_project') }}</th>
                  <th>{{ $t('invoice.line_description') }}</th>
                  <th class="num">{{ $t('invoice.line_hours') }}</th>
                  <th v-if="createInvHasDistance" class="num">{{ $t('invoice.line_distance') }}</th>
                  <th class="num">{{ $t('invoice.line_rate') }}</th>
                  <th class="num">{{ $t('invoice.line_amount') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(li, i) in createInvLineItems" :key="i">
                  <td>{{ formatDate(li.date) }}</td>
                  <td>{{ li.project_name }}</td>
                  <td>{{ li.description }}</td>
                  <td class="num">{{ fmtMinutesInv(li.minutes) || (li.quantity > 0 ? li.quantity : '') }}</td>
                  <td v-if="createInvHasDistance" class="num">{{ li.distance > 0 ? li.distance : '—' }}</td>
                  <td class="num">{{ li.currency }} {{ li.hourly_rate > 0 ? li.hourly_rate.toFixed(2) : li.price_per_km > 0 ? li.price_per_km.toFixed(2) : li.unit_price > 0 ? li.unit_price.toFixed(2) : '' }}</td>
                  <td class="num">{{ li.currency }} {{ li.amount.toFixed(2) }}</td>
                </tr>
              </tbody>
            </table>
            <div class="inv-preview-totals">
              <div><span>{{ $t('invoice.subtotal') }}</span><strong>{{ createInvCurrency }} {{ createInvSubtotal.toFixed(2) }}</strong></div>
              <div v-if="createInvForm.vat_rate > 0"><span>{{ $t('invoice.vat_amount') }} ({{ createInvForm.vat_rate }}%)</span><strong>{{ createInvCurrency }} {{ createInvVAT.toFixed(2) }}</strong></div>
              <div class="inv-total-row"><span>{{ $t('invoice.total') }}</span><strong>{{ createInvCurrency }} {{ createInvTotal.toFixed(2) }}</strong></div>
            </div>
          </div>
        </div>
        <div class="tt-modal-ft">
          <template v-if="createInvStep === 1">
            <button class="btn btn-secondary" @click="showCreateInvoiceFromGroup = false">{{ $t('common.cancel') }}</button>
            <button class="btn btn-primary" :disabled="!createInvForm.period_start || !createInvForm.period_end" @click="createInvStep = 2">{{ $t('invoice.preview_items') }}</button>
          </template>
          <template v-else>
            <button class="btn btn-secondary" @click="createInvStep = 1">{{ $t('common.go_back') }}</button>
            <button class="btn btn-primary" :disabled="createInvLineItems.length === 0 || savingInvFromGrp" @click="saveInvoiceFromGroup">{{ $t('invoice.confirm_create') }}</button>
          </template>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- ── Manage time-tracking projects & customers modal ───────────────── -->
  <Teleport to="body">
    <div v-if="managingProjects" class="tt-modal-backdrop" @click.self="closeManageProjects" @keydown.escape="closeManageProjects">
      <div class="tt-modal" role="dialog" aria-modal="true" aria-labelledby="tt-modal-title" ref="modalRef">
        <div class="tt-modal-hd">
          <h2 id="tt-modal-title" class="tt-modal-title">{{ $t('timeTracking.manage_tt_projects') }}</h2>
          <button class="tt-modal-close" @click="closeManageProjects" :aria-label="$t('common.close')">✕</button>
        </div>

        <!-- Tabs -->
        <div class="ttp-tabs" role="tablist">
          <button
            class="ttp-tab"
            role="tab"
            :aria-selected="manageTab === 'projects'"
            :aria-controls="'ttp-panel-projects'"
            id="ttp-tab-projects"
            :class="{ active: manageTab === 'projects' }"
            @click="manageTab = 'projects'"
          >{{ $t('timeTracking.tt_tab_projects') }}</button>
          <button
            class="ttp-tab"
            role="tab"
            :aria-selected="manageTab === 'customers'"
            :aria-controls="'ttp-panel-customers'"
            id="ttp-tab-customers"
            :class="{ active: manageTab === 'customers' }"
            @click="manageTab = 'customers'"
          >{{ $t('timeTracking.tt_tab_customers') }}</button>
        </div>

        <!-- Projects tab -->
        <div v-if="manageTab === 'projects'" role="tabpanel" id="ttp-panel-projects" aria-labelledby="ttp-tab-projects">
          <p class="tt-modal-sub">{{ $t('timeTracking.tt_projects_subtitle') }}</p>
          <ul class="ttp-list" :aria-label="$t('timeTracking.tt_projects_title')">
            <li v-if="ttProjects.length === 0" class="ttp-empty">{{ $t('timeTracking.tt_projects_empty') }}</li>
            <li v-for="p in sortedTTProjects" :key="p.id" class="ttp-item">
              <template v-if="editingTTProject && editingTTProject.id === p.id">
                <label class="sr-only" :for="'ttp-proj-name-' + p.id">{{ $t('timeTracking.tt_project_name') }}</label>
                <input :id="'ttp-proj-name-' + p.id" class="ttp-name-input" v-model="editingTTProject.name" @keydown.enter="saveTTProject" @keydown.escape="cancelEditTTProject" />
                <label class="sr-only" :for="'ttp-proj-color-' + p.id">{{ $t('timeTracking.tt_project_color') }}</label>
                <input :id="'ttp-proj-color-' + p.id" type="color" class="ttp-color-input" v-model="editingTTProject.color" :aria-label="$t('timeTracking.tt_project_color')" />
                <label class="sr-only" :for="'ttp-proj-undecl-' + p.id">{{ $t('timeTracking.undeclarable') }}</label>
                <input :id="'ttp-proj-undecl-' + p.id" class="ttp-undecl-input"
                  v-model="editingTTProject.undeclStr"
                  :placeholder="timeNotation === 'hhmm' ? '0:00' : '0.00'"
                  :title="$t('timeTracking.undeclarable_per_entry')"
                  @keydown.enter="saveTTProject"
                  @keydown.escape="cancelEditTTProject"
                />
                <button class="act-btn act-ok ttp-act" @click="saveTTProject" :aria-label="$t('common.save')">✓</button>
                <button class="act-btn act-no ttp-act" @click="cancelEditTTProject" :aria-label="$t('common.cancel')">✕</button>
              </template>
              <template v-else>
                <span class="ttp-dot" :aria-hidden="true" :style="p.color ? { background: p.color } : {}"></span>
                <span class="ttp-name">{{ p.name }}</span>
                <span v-if="p.undeclarable_minutes > 0" class="ttp-undecl-badge" :title="$t('timeTracking.undeclarable_per_entry')">-{{ fmtTime(p.undeclarable_minutes) }}</span>
                <span v-if="isGlobalTTProject(p)" class="ttp-badge">{{ $t('timeTracking.tt_project_global') }}</span>
                <template v-if="canEditTTProject(p)">
                  <button class="act-btn act-edit ttp-act" @click="startEditTTProject(p)" :aria-label="$t('common.edit') + ' ' + p.name">✎</button>
                  <button class="act-btn act-del ttp-act" @click="deleteTTProject(p)" :aria-label="$t('common.delete') + ' ' + p.name">🗑</button>
                </template>
              </template>
            </li>
          </ul>
          <div class="ttp-add-form" v-if="!addingTTProject">
            <button class="btn-add-row" @click="addingTTProject = true">＋ {{ $t('timeTracking.tt_project_add') }}</button>
          </div>
          <div class="ttp-add-form ttp-add-active" v-else>
            <label class="sr-only" for="ttp-new-proj-name">{{ $t('timeTracking.tt_project_name') }}</label>
            <input id="ttp-new-proj-name" class="ttp-name-input" v-model="newTTProject.name"
              :placeholder="$t('timeTracking.tt_project_name')"
              @keydown.enter="confirmAddTTProject"
              @keydown.escape="addingTTProject = false"
              ref="newTTNameRef"
            />
            <label class="sr-only" for="ttp-new-proj-color">{{ $t('timeTracking.tt_project_color') }}</label>
            <input id="ttp-new-proj-color" type="color" class="ttp-color-input" v-model="newTTProject.color" :aria-label="$t('timeTracking.tt_project_color')" />
            <label class="sr-only" for="ttp-new-proj-undecl">{{ $t('timeTracking.undeclarable') }}</label>
            <input id="ttp-new-proj-undecl" class="ttp-undecl-input"
              v-model="newTTProject.undeclStr"
              :placeholder="timeNotation === 'hhmm' ? '0:00' : '0.00'"
              :title="$t('timeTracking.undeclarable_per_entry')"
              @keydown.enter="confirmAddTTProject"
              @keydown.escape="addingTTProject = false"
            />
            <button class="btn btn-primary btn-sm" @click="confirmAddTTProject">{{ $t('timeTracking.tt_project_save') }}</button>
            <button class="btn btn-secondary btn-sm" @click="addingTTProject = false">{{ $t('timeTracking.tt_project_cancel') }}</button>
          </div>
        </div>

        <!-- Customers tab -->
        <div v-if="manageTab === 'customers'" role="tabpanel" id="ttp-panel-customers" aria-labelledby="ttp-tab-customers">
          <p class="tt-modal-sub">{{ $t('timeTracking.tt_customers_subtitle') }}</p>
          <ul class="ttp-list" :aria-label="$t('timeTracking.tt_customers_title')">
            <li v-if="ttCustomers.length === 0" class="ttp-empty">{{ $t('timeTracking.tt_customers_empty') }}</li>
            <li v-for="c in sortedTTCustomers" :key="c.id" class="ttp-item">
              <template v-if="editingTTCustomer && editingTTCustomer.id === c.id">
                <label class="sr-only" :for="'ttp-cust-name-' + c.id">{{ $t('timeTracking.tt_customer_name') }}</label>
                <input :id="'ttp-cust-name-' + c.id" class="ttp-name-input" v-model="editingTTCustomer.name" @keydown.enter="saveTTCustomer" @keydown.escape="cancelEditTTCustomer" />
                <button class="act-btn act-ok ttp-act" @click="saveTTCustomer" :aria-label="$t('common.save')">✓</button>
                <button class="act-btn act-no ttp-act" @click="cancelEditTTCustomer" :aria-label="$t('common.cancel')">✕</button>
              </template>
              <template v-else>
                <span class="ttp-name">{{ c.name }}</span>
                <span v-if="isGlobalTTCustomer(c)" class="ttp-badge">{{ $t('timeTracking.tt_customer_global') }}</span>
                <template v-if="canEditTTCustomer(c)">
                  <button class="act-btn act-edit ttp-act" @click="startEditTTCustomer(c)" :aria-label="$t('common.edit') + ' ' + c.name">✎</button>
                  <button class="act-btn act-del ttp-act" @click="deleteTTCustomer(c)" :aria-label="$t('common.delete') + ' ' + c.name">🗑</button>
                </template>
              </template>
            </li>
          </ul>
          <div class="ttp-add-form" v-if="!addingTTCustomer">
            <button class="btn-add-row" @click="addingTTCustomer = true">＋ {{ $t('timeTracking.tt_customer_add') }}</button>
          </div>
          <div class="ttp-add-form ttp-add-active" v-else>
            <label class="sr-only" for="ttp-new-cust-name">{{ $t('timeTracking.tt_customer_name') }}</label>
            <input id="ttp-new-cust-name" class="ttp-name-input" v-model="newTTCustomer.name"
              :placeholder="$t('timeTracking.tt_customer_name')"
              @keydown.enter="confirmAddTTCustomer"
              @keydown.escape="addingTTCustomer = false"
              ref="newTTCustomerNameRef"
            />
            <button class="btn btn-primary btn-sm" @click="confirmAddTTCustomer">{{ $t('timeTracking.tt_customer_save') }}</button>
            <button class="btn btn-secondary btn-sm" @click="addingTTCustomer = false">{{ $t('timeTracking.tt_customer_cancel') }}</button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Sortable from 'sortablejs'
import { Chart, registerables } from 'chart.js'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import { useSystemStore } from '@/stores/system'
import { timeEntriesApi } from '@/api/timeEntries'
import { customersApi } from '@/api/customers'
import { projectsApi } from '@/api/projects'
import client, { triggerDownload } from '@/api/client'
import { resolveAssetUrl, getServerUrl } from '@/api/serverConfig'
import BaseModal from '@/components/common/BaseModal.vue'
import HelpIcon from '@/components/common/HelpIcon.vue'
import ContractRatesModal from '@/components/common/ContractRatesModal.vue'
import DatePicker from '@/components/common/DatePicker.vue'
const BoardReportPanel = defineAsyncComponent(() => import('@/components/reports/BoardReportPanel.vue'))
import { useDateFormat } from '@/composables/useDateFormat'
import {
  splitShiftIntoDayEntries,
  weekendStandbyDefaults,
  parseWallClock as parseShiftWallClock,
} from '@/utils/shiftTimeEntries'
import { entryUndeclMins, rowDeclarableMins } from '@/utils/timeTrackingUndecl'
import { slotCoverageOnWeekday, slotDayTypeMatches } from '@/utils/contractSlotPreview'
import { parseMacroTimeInput, parseTimeNotationMinutes } from '@/utils/timeMacroInput'

Chart.register(...registerables)

const { t, locale } = useI18n()
const route = useRoute()
const auth = useAuthStore()
const ui = useUIStore()
const systemStore = useSystemStore()
const { formatDate, dateOnlyFormat } = useDateFormat()

const showRates = ref(false)

// ── Shared lookup data ────────────────────────────────────────────────────
const customers    = ref([])    // regular CRM customers
const ttCustomers  = ref([])    // time-tracking-only customers
const projects     = ref([])    // regular board projects
const ttProjects   = ref([])    // time-tracking-only projects

const sortedTTProjects = computed(() =>
  [...ttProjects.value].sort((a, b) => a.name.localeCompare(b.name))
)
const sortedTTCustomers = computed(() =>
  [...ttCustomers.value].sort((a, b) => a.name.localeCompare(b.name))
)

// Merged lists for dropdowns
const allCustomers = computed(() => [...customers.value, ...ttCustomers.value])
const allProjects  = computed(() => [...projects.value, ...ttProjects.value])

// Contracts per customer — lazy-loaded when a customer is selected in add/edit row
const contractsByCustomer = ref({})

async function loadContractsForCustomer(customerId) {
  if (!customerId || contractsByCustomer.value[customerId]) return
  try {
    const { data } = await customersApi.listContracts(customerId)
    contractsByCustomer.value = { ...contractsByCustomer.value, [customerId]: data }
  } catch {}
}

// ── User switching (admin / time_tracking_viewer only) ────────────────────
const canViewOtherUsers = computed(() => auth.isAdmin || !!auth.user?.time_tracking_viewer)
const allUsers = ref([])
const selectedUserId = ref(auth.user?.id ?? null)
const viewingOther = computed(() => !!selectedUserId.value && selectedUserId.value !== auth.user?.id)

function userName(u) {
  return [u.first_name, u.last_name].filter(Boolean).join(' ') || u.display_name || u.username
}

const displayName = computed(() => {
  if (canViewOtherUsers.value) {
    if (selectedUserId.value === 0) return t('timeTracking.all_employees')
    const u = allUsers.value.find(u => u.id === selectedUserId.value) || auth.user
    return u ? userName(u) : ''
  }
  return auth.user ? userName(auth.user) : ''
})

function onUserChange() {
  localRows.value = []
  editingRow.value = null
  deletingRow.value = null
  addingRow.value = false
  clearCellSelection()
  clearUndoStack()
  loadWeek()
  loadReport()
}

// ── PDF export options ────────────────────────────────────────────────────
const pdfFont = ref(localStorage.getItem('timeTracking.pdfFont') || 'inter')
watch(pdfFont, v => localStorage.setItem('timeTracking.pdfFont', v))
const pdfLang = ref(localStorage.getItem('timeTracking.pdfLang') || 'auto')
watch(pdfLang, v => localStorage.setItem('timeTracking.pdfLang', v))
const pdfPageBreak = ref(localStorage.getItem('timeTracking.pdfPageBreak') === '1')
watch(pdfPageBreak, v => localStorage.setItem('timeTracking.pdfPageBreak', v ? '1' : '0'))
const pdfShowAbbr = ref(localStorage.getItem('timeTracking.pdfShowAbbr') === '1')
watch(pdfShowAbbr, v => localStorage.setItem('timeTracking.pdfShowAbbr', v ? '1' : '0'))
const pdfShowCosts = ref(localStorage.getItem('timeTracking.pdfShowCosts') === '1')
watch(pdfShowCosts, v => localStorage.setItem('timeTracking.pdfShowCosts', v ? '1' : '0'))
const pdfShowPageNumbers = ref(localStorage.getItem('timeTracking.pdfShowPageNumbers') !== '0')
watch(pdfShowPageNumbers, v => localStorage.setItem('timeTracking.pdfShowPageNumbers', v ? '1' : '0'))
const pdfShowUndeclarable = ref(localStorage.getItem('timeTracking.pdfShowUndeclarable') !== '0')
watch(pdfShowUndeclarable, v => localStorage.setItem('timeTracking.pdfShowUndeclarable', v ? '1' : '0'))
const pdfShowUndeclarableRow = ref(localStorage.getItem('timeTracking.pdfShowUndeclarableRow') === '1')
watch(pdfShowUndeclarableRow, v => localStorage.setItem('timeTracking.pdfShowUndeclarableRow', v ? '1' : '0'))
const pdfShowDistance = ref(localStorage.getItem('timeTracking.pdfShowDistance') === '1')
watch(pdfShowDistance, v => localStorage.setItem('timeTracking.pdfShowDistance', v ? '1' : '0'))
const pdfOptionsOpen = ref(false)
const pdfOptionsRef = ref(null)
const gridPdfOpen = ref(false)
const gridPdfRef = ref(null)
const gridPdfType = ref('week')
const gridPdfWeek = ref(1)
const gridPdfWeekYear = ref(new Date().getFullYear())
const gridPdfMonth = ref(1)
const gridPdfMonthYear = ref(new Date().getFullYear())
const gridPdfYear = ref(new Date().getFullYear())
const gridPdfYears = computed(() => {
  const y = new Date().getFullYear()
  return Array.from({ length: 5 }, (_, i) => y - i)
})
const gridPdfMonthNames = computed(() => {
  const fmt = new Intl.DateTimeFormat(locale.value || 'en', { month: 'long' })
  return Array.from({ length: 12 }, (_, i) => fmt.format(new Date(2024, i, 1)))
})

// ── Mode ──────────────────────────────────────────────────────────────────
// Honour ?tab=board-report from the /reports redirect; also default to board-report
// when the user only has canViewReports and no time-tracking access.
const _initialMode = (() => {
  if (route.query.tab === 'board-report' && auth.canViewReports) return 'board-report'
  if (!auth.timeTrackingEnabled && auth.canViewReports) return 'board-report'
  return 'sheet'
})()
const mode = ref(_initialMode)

watch(mode, (activeMode) => {
  ui.setHelpContext(`timeTracking.${activeMode}`)
}, { immediate: true })

// ── Week navigation ───────────────────────────────────────────────────────
const anchor = ref(new Date())

// 0 = Sunday-start, 1 = Monday-start (ISO default)
const weekStartDay = computed(() => auth.user?.week_start === 'sunday' ? 0 : 1)

const weekStart = computed(() => {
  const d = new Date(anchor.value)
  d.setHours(0, 0, 0, 0)
  const day = d.getDay()
  if (weekStartDay.value === 0) {
    d.setDate(d.getDate() - day)         // back to most-recent Sunday
  } else {
    d.setDate(d.getDate() + (day === 0 ? -6 : 1 - day)) // back to Monday
  }
  return d
})

const weekInfo = computed(() => {
  // Use the Thursday of the 7-day span to derive a stable ISO week/year label
  const thu = new Date(weekStart.value)
  thu.setDate(thu.getDate() + (weekStartDay.value === 0 ? 4 : 3))
  const d = new Date(Date.UTC(thu.getFullYear(), thu.getMonth(), thu.getDate()))
  const y0 = new Date(Date.UTC(d.getUTCFullYear(), 0, 1))
  return { year: d.getUTCFullYear(), week: Math.ceil(((d - y0) / 86400000 + 1) / 7) }
})

const weekDays = computed(() => {
  const abbr = new Intl.DateTimeFormat(undefined, { weekday: 'short' })
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(weekStart.value)
    d.setDate(d.getDate() + i)
    // Use local date components so the ISO key matches the visible day name;
    // toISOString() would give the UTC date which is one day behind in UTC+ zones.
    const iso = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const todayIso = (() => { const t = new Date(); return `${t.getFullYear()}-${String(t.getMonth() + 1).padStart(2, '0')}-${String(t.getDate()).padStart(2, '0')}` })()
    return { iso, mmdd: iso.slice(5), abbr: abbr.format(d), isToday: iso === todayIso }
  })
})

function shiftWeek(delta) {
  const d = new Date(anchor.value)
  d.setDate(d.getDate() + delta * 7)
  anchor.value = d
  localRows.value = []
  editingRow.value = null
  deletingRow.value = null
  addingRow.value = false
  clearCellSelection()
  clearUndoStack()
  loadWeek()
}

const isCurrentWeek = computed(() => {
  const { week, year } = weekInfo.value
  // Find the week start for today under the current weekStartDay setting
  const today = new Date(); today.setHours(0, 0, 0, 0)
  const todayDay = today.getDay()
  const todayWS = new Date(today)
  if (weekStartDay.value === 0) {
    todayWS.setDate(today.getDate() - todayDay)
  } else {
    todayWS.setDate(today.getDate() + (todayDay === 0 ? -6 : 1 - todayDay))
  }
  const thu = new Date(todayWS)
  thu.setDate(todayWS.getDate() + (weekStartDay.value === 0 ? 4 : 3))
  const d = new Date(Date.UTC(thu.getFullYear(), thu.getMonth(), thu.getDate()))
  const y0 = new Date(Date.UTC(d.getUTCFullYear(), 0, 1))
  const todayWeek = Math.ceil(((d - y0) / 86400000 + 1) / 7)
  return week === todayWeek && year === d.getUTCFullYear()
})

function goToToday() {
  anchor.value = new Date()
  localRows.value = []
  editingRow.value = null
  deletingRow.value = null
  addingRow.value = false
  clearCellSelection()
  clearUndoStack()
  loadWeek()
}

// ── Week picker calendar ──────────────────────────────────────────────────
const wkPickerOpen = ref(false)
const wkPickerRef  = ref(null)
const calAnchor    = ref(null)

function openWkPicker() {
  if (!wkPickerOpen.value) {
    const s = weekStart.value
    calAnchor.value = new Date(s.getFullYear(), s.getMonth(), 1)
  }
  wkPickerOpen.value = !wkPickerOpen.value
}

function calShiftMonth(delta) {
  const d = new Date(calAnchor.value)
  d.setMonth(d.getMonth() + delta)
  calAnchor.value = d
}

const calMonthLabel = computed(() => {
  if (!calAnchor.value) return ''
  return new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric' }).format(calAnchor.value)
})

const calDayNames = computed(() => {
  const fmt = new Intl.DateTimeFormat(undefined, { weekday: 'narrow' })
  // 2024-01-01 is Monday; 2023-12-31 is Sunday — pick start based on preference
  const base = weekStartDay.value === 0 ? new Date(2023, 11, 31) : new Date(2024, 0, 1)
  return Array.from({ length: 7 }, (_, i) => fmt.format(new Date(base.getFullYear(), base.getMonth(), base.getDate() + i)))
})

function isoWeekNum(d, startDay = 1) {
  // Find Thursday of the 7-day week starting on startDay to get the ISO week number
  const thu = new Date(d)
  thu.setDate(d.getDate() + (startDay === 0 ? 4 : 3))
  const y0 = new Date(Date.UTC(thu.getFullYear(), 0, 1))
  const td = new Date(Date.UTC(thu.getFullYear(), thu.getMonth(), thu.getDate()))
  return Math.ceil(((td - y0) / 86400000 + 1) / 7)
}

const calCells = computed(() => {
  if (!calAnchor.value) return []
  const year = calAnchor.value.getFullYear()
  const month = calAnchor.value.getMonth()
  const first = new Date(year, month, 1)
  const dayOfWeek = first.getDay()
  // Offset back to the nearest week-start day
  const startOffset = weekStartDay.value === 0
    ? -dayOfWeek                               // Sunday start
    : (dayOfWeek === 0 ? -6 : 1 - dayOfWeek)  // Monday start
  const start = new Date(year, month, 1 + startOffset)
  const ws = weekStart.value
  const we = new Date(ws); we.setDate(we.getDate() + 7)
  const today = new Date(); today.setHours(0, 0, 0, 0)
  return Array.from({ length: 6 }, (_, row) => {
    const rowStart = new Date(start); rowStart.setDate(rowStart.getDate() + row * 7)
    const days = Array.from({ length: 7 }, (_, col) => {
      const d = new Date(rowStart); d.setDate(d.getDate() + col)
      return { date: d, day: d.getDate(), otherMonth: d.getMonth() !== month,
        inSelectedWeek: d >= ws && d < we, isToday: d.getTime() === today.getTime() }
    })
    return { weekNum: isoWeekNum(rowStart, weekStartDay.value), rowStart, days, isSelectedWeek: days.some(c => c.inSelectedWeek) }
  })
})

function goToDate(date) {
  anchor.value = new Date(date)
  wkPickerOpen.value = false
  localRows.value = []
  editingRow.value = null
  deletingRow.value = null
  addingRow.value = false
  clearCellSelection()
  clearUndoStack()
  loadWeek()
}

function onWkPickerDocClick(e) {
  if (wkPickerRef.value && !wkPickerRef.value.contains(e.target)) wkPickerOpen.value = false
}

const holidaysDropOpen = ref(false)
const holidaysDropRef  = ref(null)

const isTauri = !!window.__TAURI_INTERNALS__

const holidayLocales = [
  { locale: 'en', flag: 'GB', emoji: '🇬🇧', label: 'United Kingdom' },
  { locale: 'nl', flag: 'NL', emoji: '🇳🇱', label: 'Nederland' },
  { locale: 'de', flag: 'DE', emoji: '🇩🇪', label: 'Deutschland' },
  { locale: 'fr', flag: 'FR', emoji: '🇫🇷', label: 'France' },
  { locale: 'es', flag: 'ES', emoji: '🇪🇸', label: 'España' },
  { locale: 'da', flag: 'DK', emoji: '🇩🇰', label: 'Danmark' },
  { locale: 'sv', flag: 'SE', emoji: '🇸🇪', label: 'Sverige' },
  { locale: 'nb', flag: 'NO', emoji: '🇳🇴', label: 'Norge' },
  { locale: 'fi', flag: 'FI', emoji: '🇫🇮', label: 'Suomi' },
  { locale: 'is', flag: 'IS', emoji: '🇮🇸', label: 'Ísland' },
  { locale: 'pt', flag: 'PT', emoji: '🇵🇹', label: 'Portugal' },
  { locale: 'it', flag: 'IT', emoji: '🇮🇹', label: 'Italia' },
]

function onHolidaysDocClick(e) {
  if (holidaysDropRef.value && !holidaysDropRef.value.contains(e.target)) {
    holidaysDropOpen.value = false
  }
}

function onNavEscapeKey(e) {
  if (e.key !== 'Escape') return
  if (wkPickerOpen.value) { wkPickerOpen.value = false; e.stopPropagation(); return }
  if (holidaysDropOpen.value) { holidaysDropOpen.value = false; e.stopPropagation(); return }
  if (macroRunOpen.value) { macroRunOpen.value = false; e.stopPropagation() }
}

async function addHolidays(loc) {
  holidaysDropOpen.value = false
  const year = weekInfo.value.year
  try {
    const { data } = await timeEntriesApi.addHolidays({ year, locale: loc })
    if (data.added === 0 && data.skipped > 0) {
      ui.info(t('timeTracking.holidays_all_exist', { year, count: data.skipped }))
    } else if (data.skipped === 0) {
      ui.success(t('timeTracking.holidays_added', { year, count: data.added }))
    } else {
      ui.success(t('timeTracking.holidays_added_skipped', { year, added: data.added, skipped: data.skipped }))
    }
    if (data.added > 0) loadWeek()
  } catch {
    ui.error(t('timeTracking.holidays_error'))
  }
}

// ── Entry state ───────────────────────────────────────────────────────────
const rawEntries = ref([])  // TimeEntry[] from API for this week
const localRows  = ref([])  // rows added locally but with no entries yet
const rowComments = ref({}) // { [rowKey]: string } row-level comments
const loading    = ref(false)
const savingCell = ref('')  // "rowKey+dateISO" while a save is in flight

function rowKey(customerId, projectId, description) {
  return `${customerId ?? ''}|${projectId ?? ''}|${description ?? ''}`
}

function parseRowKey(key) {
  const i1 = key.indexOf('|')
  const i2 = key.indexOf('|', i1 + 1)
  if (i1 < 0 || i2 < 0) return { customer_id: null, project_id: null, description: '' }
  const cid = key.slice(0, i1)
  const pid = key.slice(i1 + 1, i2)
  return {
    customer_id: cid ? Number(cid) : null,
    project_id:  pid ? Number(pid) : null,
    description: key.slice(i2 + 1),
  }
}

function rowFromKey(key) {
  const { customer_id, project_id, description } = parseRowKey(key)
  const cust = allCustomers.value.find(c => c.id === customer_id)
  const proj = allProjects.value.find(p => p.id === project_id)
  return {
    key,
    customer_id,
    customer_name: cust?.name || '',
    project_id,
    project_name:  proj?.name || '',
    contract_id:   null,
    contract_name: '',
    description,
    is_holiday: false,
  }
}

function weekOrderParams() {
  return { year: weekInfo.value.year, week: weekInfo.value.week }
}

function entryKeySet() {
  const s = new Set()
  for (const e of rawEntries.value) {
    s.add(rowKey(e.customer_id, e.project_id, e.description))
  }
  return s
}

function restoreEmptyRowsFromOrder(keys) {
  if (!keys?.length) return
  const seen = entryKeySet()
  for (const k of keys) {
    if (seen.has(k)) continue
    if (localRows.value.some(r => r.key === k)) continue
    localRows.value.push(rowFromKey(k))
  }
}

function ensureLocalRow(row) {
  const k = row.key
  if (entryKeySet().has(k)) return
  if (localRows.value.some(r => r.key === k)) return
  localRows.value.push({
    key:           k,
    customer_id:   row.customer_id,
    customer_name: row.customer_name || '',
    project_id:    row.project_id,
    project_name:  row.project_name || '',
    contract_id:   row.contract_id   || null,
    contract_name: row.contract_name || '',
    description:   row.description || '',
    is_holiday:    false,
  })
}

// Rows derived from existing entries this week
const entryRows = computed(() => {
  const seen = new Map()
  for (const e of rawEntries.value) {
    const k = rowKey(e.customer_id, e.project_id, e.description)
    if (!seen.has(k)) {
      seen.set(k, {
        key:           k,
        customer_id:   e.customer_id,
        customer_name: e.customer?.name || '',
        project_id:    e.project_id,
        project_name:  e.project?.name || '',
        contract_id:   e.contract_id   || null,
        contract_name: e.contract?.name || '',
        description:   e.description || '',
        is_holiday:    e.is_holiday || false,
        min_id:        e.id,
      })
    } else {
      if (e.is_holiday) seen.get(k).is_holiday = true
      if (e.id < seen.get(k).min_id) seen.get(k).min_id = e.id
    }
  }
  return [...seen.values()].sort((a, b) => a.min_id - b.min_id)
})

// ISO dates that have a holiday entry this week
const holidayDates = computed(() => {
  const s = new Set()
  for (const e of rawEntries.value) {
    if (e.is_holiday) s.add(e.date?.slice(0, 10) ?? '')
  }
  return s
})

// All rows = entry-derived rows + locally-added rows not yet in entries
const allRows = computed(() => {
  const keys = new Set(entryRows.value.map(r => r.key))
  const extras = localRows.value.filter(r => !keys.has(r.key))
  return [...entryRows.value, ...extras]
})

// ── Sorting ───────────────────────────────────────────────────────────────
const sortCol = ref(null)   // null | 'info' | 'desc'
const sortDir = ref('asc')  // 'asc' | 'desc'

// Stable display order — only re-sorted when the user explicitly clicks a
// sort header; data updates (cell saves) preserve the current order so rows
// don't jump while the user is entering time.
const _keyOrder = ref(null)
// Server-loaded order, used to seed _keyOrder on first allRows update.
const _serverOrder = ref(null)

let _saveOrderTimer = null
function _scheduleSaveOrder(keys) {
  clearTimeout(_saveOrderTimer)
  _saveOrderTimer = setTimeout(() => {
    const comments = { ...rowComments.value }
    for (const k of Object.keys(comments)) {
      if (!comments[k]) delete comments[k]
    }
    timeEntriesApi.setRowOrder(keys, comments, weekOrderParams()).catch(() => {})
  }, 600)
}

watch(allRows, (rows) => {
  if (_keyOrder.value === null) {
    const saved = _serverOrder.value
    if (saved) {
      const keySet = new Set(rows.map(r => r.key))
      const valid = saved.filter(k => keySet.has(k))
      const fresh = rows.map(r => r.key).filter(k => !valid.includes(k))
      _keyOrder.value = [...valid, ...fresh]
    } else {
      _keyOrder.value = rows.map(r => r.key)
    }
    return
  }
  const keySet = new Set(rows.map(r => r.key))
  const stable = _keyOrder.value.filter(k => keySet.has(k))
  const fresh  = rows.map(r => r.key).filter(k => !stable.includes(k))

  if (stable.length === 0 && fresh.length > 0) {
    const saved = _serverOrder.value
    if (saved) {
      const valid = saved.filter(k => keySet.has(k))
      const remaining = fresh.filter(k => !valid.includes(k))
      _keyOrder.value = [...valid, ...remaining]
    } else {
      _keyOrder.value = [...stable, ...fresh]
    }
  } else {
    _keyOrder.value = [...stable, ...fresh]
  }
  _scheduleSaveOrder(_keyOrder.value)
}, { immediate: true })

const tbodyEl = ref(null)
let rowSortable = null

watch(tbodyEl, (newEl) => {
  if (rowSortable) {
    rowSortable.destroy()
    rowSortable = null
  }
  if (newEl) initRowSortable()
})

function applySortOrder() {
  const sorted = [...allRows.value].sort((a, b) => {
    let va, vb
    if (sortCol.value === 'info') {
      va = ((a.customer_name || '') + '\x00' + (a.project_name || '')).toLowerCase()
      vb = ((b.customer_name || '') + '\x00' + (b.project_name || '')).toLowerCase()
    } else {
      va = (a.description || '').toLowerCase()
      vb = (b.description || '').toLowerCase()
    }
    if (va < vb) return sortDir.value === 'asc' ? -1 : 1
    if (va > vb) return sortDir.value === 'asc' ? 1 : -1
    return 0
  })
  _keyOrder.value = sorted.map(r => r.key)
  _scheduleSaveOrder(_keyOrder.value)
}

function toggleSort(col) {
  if (sortCol.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortCol.value = col
    sortDir.value = 'asc'
  }
  applySortOrder()
}

const sortedRows = computed(() => {
  if (!_keyOrder.value) return allRows.value
  const byKey = Object.fromEntries(allRows.value.map(r => [r.key, r]))
  return _keyOrder.value.map(k => byKey[k]).filter(Boolean)
})

function getEntry(row, dateISO) {
  return rawEntries.value.find(
    e => rowKey(e.customer_id, e.project_id, e.description) === row.key
      && e.date.slice(0, 10) === dateISO
  ) ?? null
}

function cellVal(row, dateISO) {
  const e = getEntry(row, dateISO)
  return e && e.minutes ? fmtTime(e.minutes) : ''
}

function rowDeclTotal(row) {
  const entries = weekDays.value.map(d => getEntry(row, d.iso))
  return fmtTime(rowDeclarableMins(entries))
}

function dayTotal(dateISO) {
  return fmtTime(dayRawMinutes(dateISO))
}

const dayOfWeekField = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat']

function dayRawMinutes(dateISO) {
  return allRows.value.reduce((s, row) => s + (getEntry(row, dateISO)?.minutes || 0), 0)
}

function dayExpectedMinutes(dateISO) {
  const dayIndex = new Date(dateISO + 'T12:00:00').getDay()
  const field = dayOfWeekField[dayIndex] + '_work_start'
  const val = auth.user?.[field]
  if (!val) return 0
  const [h, m] = val.split(':').map(Number)
  return h * 60 + m
}

function dayOverLimit(dateISO) {
  const expected = dayExpectedMinutes(dateISO)
  if (expected <= 0) return false
  return dayRawMinutes(dateISO) > expected
}

function dayExpectedLabel(dateISO) {
  const m = dayExpectedMinutes(dateISO)
  if (m <= 0) return ''
  return t('timeTracking.allotted') + ' ' + fmtTime(m)
}

// ── Undeclarable helpers — derived from the project's undeclarable_minutes ─
// For each entry: undecl = min(entry.minutes, project.undeclarable_minutes)
function reportEntryDeclarable(entry) {
  const pu = entry.project?.undeclarable_minutes || 0
  return pu > 0 ? Math.max(0, entry.minutes - pu) : entry.minutes
}

function entryUndecl(entry) {
  return entryUndeclMins(entry)
}

function rowUndecl(row) {
  return weekDays.value.reduce((s, d) => {
    const e = getEntry(row, d.iso)
    return s + (e ? entryUndecl(e) : 0)
  }, 0)
}

function cellUndeclMins(row, dateISO) {
  const e = getEntry(row, dateISO)
  return e ? entryUndecl(e) : 0
}

function dayUndeclMins(dateISO) {
  return allRows.value.reduce((s, row) => {
    const e = getEntry(row, dateISO)
    return s + (e ? entryUndecl(e) : 0)
  }, 0)
}

function dayUndecl(dateISO) {
  const m = dayUndeclMins(dateISO)
  return m > 0 ? fmtTime(m) : ''
}

function dayDeclTotal(dateISO) {
  const raw = dayRawMinutes(dateISO)
  if (raw === 0) return fmtTime(0)
  return fmtTime(Math.max(0, raw - dayUndeclMins(dateISO)))
}


const grandUndeclTotal = computed(() =>
  rawEntries.value.reduce((s, e) => s + entryUndecl(e), 0)
)

const grandDeclarable = computed(() => {
  const total = rawEntries.value.reduce((s, e) => s + e.minutes, 0)
  return Math.max(0, total - grandUndeclTotal.value)
})

function dayDistanceTotal(dateISO) {
  return rawEntries.value
    .filter(e => e.date?.startsWith(dateISO) && e.distance)
    .reduce((s, e) => s + (e.distance || 0), 0)
}

const weekDistanceTotal = computed(() =>
  rawEntries.value.reduce((s, e) => s + (e.distance || 0), 0)
)

// ── Load week ─────────────────────────────────────────────────────────────
async function loadWeek() {
  localRows.value = []
  timePopupKey.value = ''
  distPopupKey.value = ''
  loading.value = true
  try {
    const from = weekDays.value[0].iso
    const to   = weekDays.value[6].iso
    const params = { from, to }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const [{ data }, orderRes] = await Promise.all([
      timeEntriesApi.list(params),
      timeEntriesApi.getRowOrder(weekOrderParams()).catch(() => ({ data: { keys: [] } })),
    ])
    rawEntries.value = data
    // Eagerly load contracts for all customers that appear in this week's entries
    // so the rate column is populated without waiting for a row interaction.
    const contractCustomerIds = [...new Set(
      data.filter(e => e.contract_id && e.customer_id).map(e => e.customer_id)
    )]
    contractCustomerIds.forEach(id => loadContractsForCustomer(id))
    rowComments.value = orderRes.data?.comments || {}
    const keys = orderRes.data?.keys
    if (keys?.length) {
      _serverOrder.value = keys
      _keyOrder.value = null
      restoreEmptyRowsFromOrder(keys)
    }
  } catch {
    ui.error(t('timeTracking.load_error'))
  } finally {
    loading.value = false
  }
}

// ── Cell undo ─────────────────────────────────────────────────────────────
const undoStack = ref([])
const undoInProgress = ref(false)
const cellRenderEpoch = ref(0)
const MAX_UNDO = 50

function isUndoShortcut(event) {
  if (!(event.ctrlKey || event.metaKey) || event.shiftKey || event.altKey) return false
  return event.code === 'KeyZ' || (typeof event.key === 'string' && event.key.toLowerCase() === 'z')
}

function isUndoBlockedTarget(el) {
  if (!(el instanceof HTMLElement) || !el.closest('.tt-view')) return true
  if (el.closest('.modal-backdrop')) return true
  if (el.closest('.tt-editing')) return true
  const tag = el.tagName
  if (tag === 'SELECT') return true
  if (tag === 'INPUT' || tag === 'TEXTAREA') {
    return !el.classList.contains('h-inp') && !el.classList.contains('tp-inp')
  }
  return false
}

function shouldHandleTimeTrackingUndo(event) {
  if (mode.value !== 'sheet' || viewingOther.value || undoInProgress.value || undoStack.value.length === 0) {
    return false
  }
  if (event.type === 'keydown' && !isUndoShortcut(event)) return false
  if (event.type === 'beforeinput' && event.inputType !== 'historyUndo') return false
  return !isUndoBlockedTarget(event.target)
}

function triggerTimeTrackingUndo(event) {
  if (!shouldHandleTimeTrackingUndo(event)) return false
  event.preventDefault()
  event.stopPropagation()
  event.stopImmediatePropagation()
  void undoLastChange()
  return true
}

function onCellUndoCapture(event) {
  triggerTimeTrackingUndo(event)
}

function onWindowUndoCapture(event) {
  triggerTimeTrackingUndo(event)
}

function onWindowBeforeInputCapture(event) {
  triggerTimeTrackingUndo(event)
}

function onCellBeforeInput(event) {
  if (event.inputType === 'historyUndo') {
    triggerTimeTrackingUndo(event)
  } else if (event.inputType === 'historyRedo') {
    event.preventDefault()
  }
}

function snapshotCell(row, dateISO) {
  const entry = getEntry(row, dateISO)
  return entry ? { ...entry } : null
}

function pushUndoBatch(items) {
  const changes = items.map(({ row, dateISO, before }) => ({
    rowKey: row.key,
    customer_id: row.customer_id,
    project_id: row.project_id,
    description: row.description,
    dateISO,
    before: before ? { ...before } : null,
  }))
  if (!changes.length) return
  undoStack.value.push({ type: 'cells', changes })
  if (undoStack.value.length > MAX_UNDO) undoStack.value.shift()
}

function pushUndoCopyPrevWeek(snapshot) {
  undoStack.value.push({ type: 'copy_prev_week', ...snapshot })
  if (undoStack.value.length > MAX_UNDO) undoStack.value.shift()
}

function undoCopyPrevWeek(action) {
  const addedSet = new Set(action.addedKeys)
  localRows.value = localRows.value.filter(r => !addedSet.has(r.key))
  _keyOrder.value = [...action.prevKeyOrder]
  rowComments.value = { ...action.prevComments }
  _scheduleSaveOrder(_keyOrder.value)
  cellRenderEpoch.value++
}

function clearUndoStack() {
  undoStack.value = []
}

function rowForUndo(change) {
  return sortedRows.value.find(r => r.key === change.rowKey) ?? {
    key: change.rowKey,
    customer_id: change.customer_id,
    project_id: change.project_id,
    description: change.description,
    customer_name: '',
    project_name: '',
  }
}

async function restoreCellChange(change) {
  const row = rowForUndo(change)
  const existing = getEntry(row, change.dateISO)
  const before = change.before

  if (before === null) {
    if (existing) {
      await timeEntriesApi.remove(existing.id)
      rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
      ensureLocalRow(row)
    }
    return
  }

  const payload = {
    customer_id: change.customer_id ?? null,
    project_id: change.project_id ?? null,
    date: change.dateISO,
    minutes: before.minutes,
    description: change.description,
    is_holiday: before.is_holiday ?? false,
    start_time: before.start_time ?? null,
    end_time: before.end_time ?? null,
    distance: before.distance ?? null,
  }

  if (existing) {
    const { data } = await timeEntriesApi.update(existing.id, payload)
    const idx = rawEntries.value.findIndex(e => e.id === existing.id)
    if (idx >= 0) rawEntries.value[idx] = data
    else rawEntries.value.push(data)
  } else {
    const { data } = await timeEntriesApi.create(payload)
    rawEntries.value.push(data)
    localRows.value = localRows.value.filter(r => r.key !== change.rowKey)
  }
}

async function undoLastChange() {
  if (undoInProgress.value || viewingOther.value || undoStack.value.length === 0) return false
  undoInProgress.value = true
  const action = undoStack.value.pop()
  try {
    if (action.type === 'copy_prev_week') {
      undoCopyPrevWeek(action)
    } else {
      for (const change of action.changes) {
        await restoreCellChange(change)
      }
      refreshCellInputs()
      cellRenderEpoch.value++
    }
    return true
  } catch {
    undoStack.value.push(action)
    ui.error(t('timeTracking.undo_error'))
    return false
  } finally {
    undoInProgress.value = false
  }
}

function refreshCellInputs() {
  nextTick(() => {
    for (let ri = 0; ri < sortedRows.value.length; ri++) {
      for (let di = 0; di < weekDays.value.length; di++) {
        const inp = document.getElementById(`tt-cell-${ri}-${di}`)
        if (!inp) continue
        const row = sortedRows.value[ri]
        const iso = weekDays.value[di]?.iso
        if (row && iso) inp.value = cellVal(row, iso)
      }
    }
  })
}

// ── Cell save ─────────────────────────────────────────────────────────────
async function onCellBlur(row, dateISO, rawVal) {
  if (undoInProgress.value) return
  const minutes = parseTimeInput(rawVal)
  const existing = getEntry(row, dateISO)

  if (minutes === (existing?.minutes || 0)) return   // no change

  const before = snapshotCell(row, dateISO)
  const ck = row.key + dateISO
  savingCell.value = ck

  try {
    let changed = false
    if (minutes === 0 && existing) {
      if (existing.is_holiday) {
        // Keep holiday markers at 0 minutes rather than deleting them
        const { data } = await timeEntriesApi.update(existing.id, {
          customer_id:  row.customer_id  || null,
          project_id:   row.project_id   || null,
          contract_id:  row.contract_id  || null,
          date:         dateISO,
          minutes:      0,
          description:  row.description,
          is_holiday:   true,
          start_time:   existing.start_time || null,
          end_time:     existing.end_time   || null,
        })
        const idx = rawEntries.value.findIndex(e => e.id === existing.id)
        rawEntries.value[idx] = data
        changed = true
      } else {
        await timeEntriesApi.remove(existing.id)
        rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
        ensureLocalRow(row)
        changed = true
      }
    } else if (minutes > 0 && existing) {
      const { data } = await timeEntriesApi.update(existing.id, {
        customer_id:  row.customer_id  || null,
        project_id:   row.project_id   || null,
        contract_id:  row.contract_id  || null,
        date:         dateISO,
        minutes,
        description:  row.description,
        is_holiday:   existing.is_holiday,
        start_time:   existing.start_time || null,
        end_time:     existing.end_time   || null,
      })
      const idx = rawEntries.value.findIndex(e => e.id === existing.id)
      rawEntries.value[idx] = data
      changed = true
    } else if (minutes > 0) {
      const { data } = await timeEntriesApi.create({
        customer_id:  row.customer_id  || null,
        project_id:   row.project_id   || null,
        contract_id:  row.contract_id  || null,
        date:         dateISO,
        minutes,
        description:  row.description,
      })
      rawEntries.value.push(data)
      localRows.value = localRows.value.filter(r => r.key !== row.key)
      changed = true
    }
    if (changed) pushUndoBatch([{ row, dateISO, before }])
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
  }
}

// ── Rate column helpers ───────────────────────────────────────────────────
function rowContract(row) {
  if (!row.contract_id || !row.customer_id) return null
  const contracts = contractsByCustomer.value[row.customer_id] || []
  return contracts.find(c => c.id === row.contract_id) || null
}

function rowRateLabel(row) {
  const c = rowContract(row)
  if (!c || c.price_per_hour == null) return ''
  const currency = c.currency || '€'
  return `${c.price_per_hour} ${currency}/h`
}

function rowHasSlots(row) {
  return !!(rowContract(row)?.time_slots?.length)
}

// ── Edit row ──────────────────────────────────────────────────────────────
const editingRow = ref(null)
const editForm   = ref({ customer_id: null, project_id: null, contract_id: null, description: '' })
const deletingRow = ref(null)

const editRowContracts = computed(() => {
  const id = editForm.value.customer_id
  if (!id) return []
  return contractsByCustomer.value[id] || []
})

const editRowProjects = computed(() => {
  if (!editForm.value.customer_id) return allProjects.value
  if (ttCustomers.value.some(c => c.id === editForm.value.customer_id))
    return ttProjects.value
  return [...projects.value.filter(p => p.customer_id === editForm.value.customer_id), ...ttProjects.value]
})

function onRowFieldDblClick(row) {
  if (viewingOther.value) return
  if (deletingRow.value === row.key) return
  startEditRow(row)
}

function startEditRow(row) {
  cancelDeleteRow()
  editForm.value = { customer_id: row.customer_id, project_id: row.project_id, contract_id: row.contract_id || null, description: row.description }
  loadContractsForCustomer(row.customer_id)
  editingRow.value = row.key
}

function duplicateRow(row) {
  let desc = row.description || ''
  const m = desc.match(/\(copy(?: (\d+))?\)$/)
  if (m) {
    const n = m[1] ? parseInt(m[1]) + 1 : 2
    desc = desc.replace(/\(copy(?: \d+)?\)$/, `(copy ${n})`)
  } else {
    desc = desc ? `${desc} (copy)` : '(copy)'
  }
  const k = rowKey(row.customer_id, row.project_id, desc)
  if (allRows.value.find(x => x.key === k)) return
  localRows.value.push({
    key:           k,
    customer_id:   row.customer_id,
    customer_name: row.customer_name,
    project_id:    row.project_id,
    project_name:  row.project_name,
    contract_id:   row.contract_id   || null,
    contract_name: row.contract_name || '',
    description:   desc,
    is_holiday:    false,
  })
  nextTick(() => {
    const srcIdx = _keyOrder.value.indexOf(row.key)
    const dstIdx = _keyOrder.value.indexOf(k)
    if (srcIdx !== -1 && dstIdx !== -1) {
      const arr = [..._keyOrder.value]
      arr.splice(dstIdx, 1)
      arr.splice(srcIdx + 1, 0, k)
      _keyOrder.value = arr
    }
  })
}

function cancelEditRow() {
  editingRow.value = null
}

async function confirmEditRow(row) {
  const r = editForm.value
  const cust = allCustomers.value.find(c => c.id === r.customer_id)
  const proj = allProjects.value.find(p => p.id === r.project_id)
  const contracts = contractsByCustomer.value[r.customer_id] || []
  const cont = contracts.find(c => c.id === r.contract_id)
  const newKey = rowKey(r.customer_id, r.project_id, r.description)

  const toUpdate = rawEntries.value.filter(
    e => rowKey(e.customer_id, e.project_id, e.description) === row.key
  )
  try {
    for (const e of toUpdate) {
      const { data } = await timeEntriesApi.update(e.id, {
        customer_id:  r.customer_id  || null,
        project_id:   r.project_id   || null,
        contract_id:  r.contract_id  || null,
        date:         e.date.slice(0, 10),
        minutes:      e.minutes,
        description:  r.description,
        is_holiday:   e.is_holiday,
      })
      const idx = rawEntries.value.findIndex(x => x.id === e.id)
      rawEntries.value[idx] = data
    }
    const li = localRows.value.findIndex(x => x.key === row.key)
    if (li >= 0) {
      localRows.value[li] = {
        ...localRows.value[li],
        key:           newKey,
        customer_id:   r.customer_id,
        customer_name: cust?.name || '',
        project_id:    r.project_id,
        project_name:  proj?.name || '',
        contract_id:   r.contract_id || null,
        contract_name: cont?.name || '',
        description:   r.description,
      }
    }
    if (newKey !== row.key && !_keyOrder.value.includes(newKey)) {
      _keyOrder.value = _keyOrder.value.map(k => k === row.key ? newKey : k)
    }
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    editingRow.value = null
  }
}

function startDeleteRow(row) {
  cancelEditRow()
  deletingRow.value = row.key
}

function cancelDeleteRow() {
  deletingRow.value = null
}

async function confirmDeleteRow(row) {
  const toDelete = rawEntries.value.filter(
    e => rowKey(e.customer_id, e.project_id, e.description) === row.key
  )
  try {
    for (const e of toDelete) {
      await timeEntriesApi.remove(e.id)
    }
    rawEntries.value = rawEntries.value.filter(
      e => rowKey(e.customer_id, e.project_id, e.description) !== row.key
    )
    localRows.value = localRows.value.filter(r => r.key !== row.key)
  } catch {
    ui.error(t('timeTracking.delete_error'))
  } finally {
    deletingRow.value = null
  }
}

// ── Row comments ───────────────────────────────────────────────────────────
const commentRowKey = ref(null)
const commentDraft  = ref('')
const commentTextareaRef = ref(null)
const commentPopupStyle = ref({})

let _commentCloseHandler = null

function openRowComment(row) {
  cancelEditRow()
  const btn = document.activeElement
  if (btn) {
    const r = btn.getBoundingClientRect()
    commentPopupStyle.value = {
      position: 'fixed',
      left: `${r.left}px`,
      top: `${r.bottom + 4}px`,
      zIndex: 10000,
    }
  } else {
    commentPopupStyle.value = {
      position: 'fixed',
      left: '100px',
      top: '200px',
      zIndex: 10000,
    }
  }
  commentRowKey.value = row.key
  commentDraft.value  = rowComments.value[row.key] || ''
  _commentCloseHandler = (e) => {
    if (commentRowKey.value && !e.target.closest('.rc-comment-popup') && !e.target.closest('.rc-comment-btn')) {
      closeRowComment()
    }
  }
  document.addEventListener('mousedown', _commentCloseHandler, true)
  nextTick(() => {
    const el = commentTextareaRef.value
    if (el) {
      if (Array.isArray(el)) el[0]?.focus()
      else el.focus()
    }
  })
}

function closeRowComment() {
  commentRowKey.value = null
  commentDraft.value  = ''
  commentPopupStyle.value = {}
  if (_commentCloseHandler) {
    document.removeEventListener('mousedown', _commentCloseHandler, true)
    _commentCloseHandler = null
  }
}

function saveRowComment(row) {
  const text = commentDraft.value?.trim() || ''
  if (text) {
    rowComments.value[row.key] = text
  } else {
    delete rowComments.value[row.key]
  }
  closeRowComment()
  _scheduleSaveOrder(_keyOrder.value || [])
}

// ── Per-cell holiday toggle ───────────────────────────────────────────────
async function toggleCellHoliday(row, dateISO) {
  const existing = getEntry(row, dateISO)
  const before = snapshotCell(row, dateISO)
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    if (existing) {
      if (existing.is_holiday && existing.minutes === 0) {
        // 0-minute holiday marker — removing holiday means removing the entry entirely
        await timeEntriesApi.remove(existing.id)
        rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
        ensureLocalRow(row)
      } else {
        const { data } = await timeEntriesApi.update(existing.id, {
          customer_id: row.customer_id || null,
          project_id:  row.project_id  || null,
          date:        dateISO,
          minutes:     existing.minutes,
          description: row.description,
          is_holiday:  !existing.is_holiday,
        })
        const idx = rawEntries.value.findIndex(e => e.id === existing.id)
        rawEntries.value[idx] = data
      }
    } else {
      const { data } = await timeEntriesApi.create({
        customer_id:  row.customer_id  || null,
        project_id:   row.project_id   || null,
        date:         dateISO,
        minutes:      0,
        description:  row.description,
        is_holiday:   true,
      })
      rawEntries.value.push(data)
      localRows.value = localRows.value.filter(r => r.key !== row.key)
    }
    pushUndoBatch([{ row, dateISO, before }])
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
  }
}

// ── Cell selection, copy / paste ──────────────────────────────────────────
// copiedBlock holds a rectangular grid of copied cells. It persists across
// week navigation until the user copies something else or presses Escape.
const copiedBlock = ref(null) // { height, width, cells[][], sourceKeys[] }
const selectionAnchor = ref(null) // { rowIdx, dayIdx }
const selectionFocus = ref(null)  // { rowIdx, dayIdx }
const selectionFromKeyboard = ref(false)
const isDragging = ref(false)
let dragStartX = 0
let dragStartY = 0

function clearCellSelection() {
  selectionAnchor.value = null
  selectionFocus.value = null
}

function getSelectionRect() {
  if (!selectionAnchor.value || !selectionFocus.value) return null
  const r0 = Math.min(selectionAnchor.value.rowIdx, selectionFocus.value.rowIdx)
  const r1 = Math.max(selectionAnchor.value.rowIdx, selectionFocus.value.rowIdx)
  const d0 = Math.min(selectionAnchor.value.dayIdx, selectionFocus.value.dayIdx)
  const d1 = Math.max(selectionAnchor.value.dayIdx, selectionFocus.value.dayIdx)
  return { r0, r1, d0, d1, height: r1 - r0 + 1, width: d1 - d0 + 1 }
}

function snapshotCellData(row, dateISO) {
  const entry = getEntry(row, dateISO)
  return {
    minutes:   entry?.minutes    ?? 0,
    startTime: entry?.start_time ?? null,
    endTime:   entry?.end_time   ?? null,
    isHoliday: entry?.is_holiday ?? false,
    distance:  entry?.distance   ?? null,
  }
}

function isCellCopied(row, dateISO) {
  return copiedBlock.value?.sourceKeys?.includes(row.key + dateISO) ?? false
}

function isCellPopupOpen(row, dateISO) {
  const key = row.key + dateISO
  return timePopupKey.value === key || distPopupKey.value === key
}

function clampSelection(rowIdx, dayIdx) {
  return {
    rowIdx: Math.max(0, Math.min(rowIdx, sortedRows.value.length - 1)),
    dayIdx: Math.max(0, Math.min(dayIdx, weekDays.value.length - 1)),
  }
}

function isCellSelected(rowIdx, dayIdx) {
  if (!selectionAnchor.value || !selectionFocus.value) return false
  const r0 = Math.min(selectionAnchor.value.rowIdx, selectionFocus.value.rowIdx)
  const r1 = Math.max(selectionAnchor.value.rowIdx, selectionFocus.value.rowIdx)
  const d0 = Math.min(selectionAnchor.value.dayIdx, selectionFocus.value.dayIdx)
  const d1 = Math.max(selectionAnchor.value.dayIdx, selectionFocus.value.dayIdx)
  return rowIdx >= r0 && rowIdx <= r1 && dayIdx >= d0 && dayIdx <= d1
}

function cellSelectionClass(rowIdx, dayIdx) {
  return isCellSelected(rowIdx, dayIdx) ? 'c-day-selected' : ''
}

function getSelectedCells() {
  if (!selectionAnchor.value || !selectionFocus.value) return []
  const r0 = Math.min(selectionAnchor.value.rowIdx, selectionFocus.value.rowIdx)
  const r1 = Math.max(selectionAnchor.value.rowIdx, selectionFocus.value.rowIdx)
  const d0 = Math.min(selectionAnchor.value.dayIdx, selectionFocus.value.dayIdx)
  const d1 = Math.max(selectionAnchor.value.dayIdx, selectionFocus.value.dayIdx)
  const cells = []
  for (let ri = r0; ri <= r1; ri++) {
    const row = sortedRows.value[ri]
    if (!row) continue
    for (let di = d0; di <= d1; di++) {
      const iso = weekDays.value[di]?.iso
      if (iso) cells.push({ row, dateISO: iso })
    }
  }
  return cells
}

function focusCellInput(rowIdx, dayIdx) {
  nextTick(() => {
    document.getElementById(`tt-cell-${rowIdx}-${dayIdx}`)?.focus()
  })
}

function onCellFocus(rowIdx, dayIdx) {
  if (selectionFromKeyboard.value) {
    selectionFromKeyboard.value = false
    return
  }
  selectionAnchor.value = { rowIdx, dayIdx }
  selectionFocus.value = { rowIdx, dayIdx }
}

function onCellMouseDown(rowIdx, dayIdx, event) {
  if (viewingOther.value) return
  if (event.shiftKey && selectionAnchor.value) {
    event.preventDefault()
    selectionFocus.value = clampSelection(rowIdx, dayIdx)
    selectionFromKeyboard.value = true
    focusCellInput(selectionFocus.value.rowIdx, selectionFocus.value.dayIdx)
    return
  }
  event.preventDefault()
  selectionAnchor.value = { rowIdx, dayIdx }
  selectionFocus.value = { rowIdx, dayIdx }
  isDragging.value = true
  dragStartX = event.clientX
  dragStartY = event.clientY
  document.addEventListener('pointermove', onPointerMove)
  document.addEventListener('pointerup', onPointerUp)
  document.addEventListener('pointercancel', onPointerUp)
}

function onPointerMove(event) {
  if (!isDragging.value) return
  event.preventDefault()
  const td = document.elementFromPoint(event.clientX, event.clientY)?.closest('[data-ri]')
  if (!td) return
  const ri = parseInt(td.dataset.ri, 10)
  const di = parseInt(td.dataset.di, 10)
  if (isNaN(ri) || isNaN(di)) return
  selectionFocus.value = clampSelection(ri, di)
}

function onPointerUp(event) {
  if (!isDragging.value) return
  isDragging.value = false
  document.removeEventListener('pointermove', onPointerMove)
  document.removeEventListener('pointerup', onPointerUp)
  document.removeEventListener('pointercancel', onPointerUp)
  const dx = event.clientX - dragStartX
  const dy = event.clientY - dragStartY
  if (Math.sqrt(dx * dx + dy * dy) < 4 && selectionAnchor.value) {
    focusCellInput(selectionAnchor.value.rowIdx, selectionAnchor.value.dayIdx)
  }
}

function moveCellSelection(rowIdx, dayIdx, extend) {
  const next = clampSelection(rowIdx, dayIdx)
  if (extend) {
    if (!selectionAnchor.value) selectionAnchor.value = next
    selectionFocus.value = next
  } else {
    selectionAnchor.value = next
    selectionFocus.value = next
  }
  selectionFromKeyboard.value = true
  focusCellInput(next.rowIdx, next.dayIdx)
}

function copyCellData(row, dateISO, rowIdx, dayIdx) {
  let rect = getSelectionRect()
  if (!rect && rowIdx != null && dayIdx != null) {
    selectionAnchor.value = { rowIdx, dayIdx }
    selectionFocus.value = { rowIdx, dayIdx }
    rect = getSelectionRect()
  }
  if (!rect) return

  const cells = []
  const sourceKeys = []
  for (let ri = rect.r0; ri <= rect.r1; ri++) {
    const rowArr = []
    const srcRow = sortedRows.value[ri]
    for (let di = rect.d0; di <= rect.d1; di++) {
      const iso = weekDays.value[di]?.iso
      if (!srcRow || !iso) {
        rowArr.push(null)
        continue
      }
      rowArr.push(snapshotCellData(srcRow, iso))
      sourceKeys.push(srcRow.key + iso)
    }
    cells.push(rowArr)
  }

  copiedBlock.value = {
    height: rect.height,
    width: rect.width,
    cells,
    sourceKeys,
  }
}

async function pasteCellDataOne(row, dateISO, src) {
  const existing = getEntry(row, dateISO)
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    if (src.minutes > 0 || src.isHoliday) {
      if (existing) {
        const { data } = await timeEntriesApi.update(existing.id, {
          customer_id: row.customer_id || null,
          project_id:  row.project_id  || null,
          date:        dateISO,
          minutes:     src.minutes,
          description: row.description,
          is_holiday:  src.isHoliday,
          start_time:  src.startTime,
          end_time:    src.endTime,
          distance:    src.distance,
        })
        const idx = rawEntries.value.findIndex(e => e.id === existing.id)
        rawEntries.value[idx] = data
      } else {
        const { data } = await timeEntriesApi.create({
          customer_id:  row.customer_id  || null,
          project_id:   row.project_id   || null,
          date:         dateISO,
          minutes:      src.minutes,
          description:  row.description,
          is_holiday:   src.isHoliday,
          start_time:   src.startTime,
          end_time:     src.endTime,
          distance:     src.distance,
        })
        rawEntries.value.push(data)
        localRows.value = localRows.value.filter(r => r.key !== row.key)
      }
    } else if (existing) {
      await timeEntriesApi.remove(existing.id)
      rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
      ensureLocalRow(row)
    }
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
  }
}

async function clearCellDataOne(row, dateISO) {
  const existing = getEntry(row, dateISO)
  if (!existing) return
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    await timeEntriesApi.remove(existing.id)
    rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
    ensureLocalRow(row)
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
  }
}

async function pasteCellData(row, dateISO, rowIdx, dayIdx) {
  const block = copiedBlock.value
  if (!block) return

  let targets = []

  if (block.height === 1 && block.width === 1) {
    const src = block.cells[0][0]
    const selected = getSelectedCells()
    if (selected.length > 1) {
      targets = selected.map(cell => ({ row: cell.row, dateISO: cell.dateISO, src }))
    } else {
      targets = [{ row, dateISO, src }]
    }
  } else {
    let pasteR = rowIdx
    let pasteD = dayIdx
    if (selectionAnchor.value) {
      pasteR = Math.min(selectionAnchor.value.rowIdx, selectionFocus.value?.rowIdx ?? rowIdx)
      pasteD = Math.min(selectionAnchor.value.dayIdx, selectionFocus.value?.dayIdx ?? dayIdx)
    }
    for (let r = 0; r < block.height; r++) {
      for (let c = 0; c < block.width; c++) {
        const cellData = block.cells[r]?.[c]
        if (cellData == null) continue
        const ri = pasteR + r
        const di = pasteD + c
        const destRow = sortedRows.value[ri]
        const iso = weekDays.value[di]?.iso
        if (!destRow || !iso) continue
        targets.push({ row: destRow, dateISO: iso, src: cellData })
      }
    }
  }

  if (!targets.length) return

  const undoItems = targets.map(cell => ({
    row: cell.row,
    dateISO: cell.dateISO,
    before: snapshotCell(cell.row, cell.dateISO),
  }))
  try {
    for (const cell of targets) {
      await pasteCellDataOne(cell.row, cell.dateISO, cell.src)
    }
    pushUndoBatch(undoItems)
    refreshCellInputs()
  } catch {
    ui.error(t('timeTracking.save_error'))
  }
}

async function cutCellData(row, dateISO, rowIdx, dayIdx) {
  copyCellData(row, dateISO, rowIdx, dayIdx)
  const rect = getSelectionRect()
  if (!rect) return
  const targets = []
  for (let ri = rect.r0; ri <= rect.r1; ri++) {
    const srcRow = sortedRows.value[ri]
    for (let di = rect.d0; di <= rect.d1; di++) {
      const iso = weekDays.value[di]?.iso
      if (srcRow && iso) targets.push({ row: srcRow, dateISO: iso })
    }
  }
  const undoItems = targets.map(cell => ({
    row: cell.row,
    dateISO: cell.dateISO,
    before: snapshotCell(cell.row, cell.dateISO),
  }))
  try {
    for (const cell of targets) {
      await clearCellDataOne(cell.row, cell.dateISO)
    }
    pushUndoBatch(undoItems)
    refreshCellInputs()
  } catch {
    ui.error(t('timeTracking.save_error'))
  }
}

function onCellKeydown(row, dateISO, rowIdx, dayIdx, event) {
  if (isUndoShortcut(event)) return

  if (event.key === 'Enter') {
    event.target.blur()
    return
  }
  if (event.key === 'Escape') {
    copiedBlock.value = null
    clearCellSelection()
    event.target.blur()
    return
  }

  if (event.key === 'Delete' || event.key === 'Backspace') {
    const { selectionStart, selectionEnd, value } = event.target
    if (selectionStart === 0 && selectionEnd === value.length) {
      event.preventDefault()
      onCellBlur(row, dateISO, '')
    }
    return
  }

  if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
    if (!isAllowedCellChar(event.key)) {
      event.preventDefault()
      return
    }
  }

  const arrowDelta = {
    ArrowUp: [-1, 0],
    ArrowDown: [1, 0],
    ArrowLeft: [0, -1],
    ArrowRight: [0, 1],
  }
  if (arrowDelta[event.key] && !viewingOther.value && editingRow.value !== row.key) {
    event.preventDefault()
    const [dr, dd] = arrowDelta[event.key]
    if (event.shiftKey) {
      if (!selectionAnchor.value) {
        selectionAnchor.value = { rowIdx, dayIdx }
        selectionFocus.value = { rowIdx, dayIdx }
      }
      const cur = selectionFocus.value ?? { rowIdx, dayIdx }
      moveCellSelection(cur.rowIdx + dr, cur.dayIdx + dd, true)
    } else {
      moveCellSelection(rowIdx + dr, dayIdx + dd, false)
    }
    return
  }

  if (event.shiftKey && event.key === 'Insert' && copiedBlock.value) {
    event.preventDefault()
    pasteCellData(row, dateISO, rowIdx, dayIdx)
    return
  }

  const mod = event.ctrlKey || event.metaKey
  if (!mod) return
  if (event.key === 'c') {
    event.preventDefault()
    const hasMulti = selectionAnchor.value && selectionFocus.value &&
      (selectionAnchor.value.rowIdx !== selectionFocus.value.rowIdx ||
       selectionAnchor.value.dayIdx !== selectionFocus.value.dayIdx)
    if (!hasMulti) {
      selectionAnchor.value = { rowIdx, dayIdx }
      selectionFocus.value = { rowIdx, dayIdx }
      event.target.select()
    }
    copyCellData(row, dateISO, rowIdx, dayIdx)
  } else if (event.key === 'x') {
    event.preventDefault()
    const hasMulti = selectionAnchor.value && selectionFocus.value &&
      (selectionAnchor.value.rowIdx !== selectionFocus.value.rowIdx ||
       selectionAnchor.value.dayIdx !== selectionFocus.value.dayIdx)
    if (!hasMulti) {
      selectionAnchor.value = { rowIdx, dayIdx }
      selectionFocus.value = { rowIdx, dayIdx }
      event.target.select()
    }
    cutCellData(row, dateISO, rowIdx, dayIdx)
  } else if (event.key === 'v' && copiedBlock.value) {
    event.preventDefault()
    pasteCellData(row, dateISO, rowIdx, dayIdx)
  }
}

function isAllowedCellChar(key) {
  if (timeNotation.value === 'hhmm') return /^[\d:]$/.test(key)
  return /^[\d.,]$/.test(key)
}

function sanitizeCellInput(raw) {
  if (timeNotation.value === 'hhmm') {
    let val = String(raw).replace(/[^\d:]/g, '')
    const colon = val.indexOf(':')
    if (colon >= 0) {
      val = val.slice(0, colon + 1) + val.slice(colon + 1).replace(/:/g, '')
    }
    return val
  }
  let val = String(raw).replace(/[^\d.,]/g, '').replace(',', '.')
  const dot = val.indexOf('.')
  if (dot >= 0) val = val.slice(0, dot + 1) + val.slice(dot + 1).replace(/\./g, '')
  return val
}

function onCellInput(event) {
  const el = event.target
  let val = sanitizeCellInput(el.value)
  if (timeNotation.value === 'hhmm' && val.length === 2 && !val.includes(':')) {
    val = val + ':'
  }
  if (val !== el.value) el.value = val
}

function onCellPaste(event) {
  event.preventDefault()
  const text = event.clipboardData?.getData('text') ?? ''
  const sanitized = sanitizeCellInput(text)
  if (!sanitized) return
  const el = event.target
  const start = el.selectionStart ?? el.value.length
  const end = el.selectionEnd ?? el.value.length
  el.value = sanitizeCellInput(el.value.slice(0, start) + sanitized + el.value.slice(end))
  onCellInput({ target: el })
}

// ── Per-cell distance popup ───────────────────────────────────────────────
const distPopupKey  = ref('')
const distPopupFlip = ref(false)
const distPopupVal  = ref('')
const distPopupRef  = ref(null)

function openDistPopup(row, dateISO, event) {
  const key = row.key + dateISO
  if (distPopupKey.value === key) { distPopupKey.value = ''; return }
  const existing = getEntry(row, dateISO)
  distPopupVal.value = existing?.distance != null ? String(existing.distance) : ''
  const rect = event?.currentTarget?.getBoundingClientRect()
  const scrollEl = event?.currentTarget?.closest('.tt-scroll')
  const scrollTop = scrollEl ? scrollEl.getBoundingClientRect().top : 0
  distPopupFlip.value = rect ? (rect.top - scrollTop) < 200 : false
  distPopupKey.value = key
}

async function applyDistPopup(row, dateISO) {
  const raw = parseFloat(distPopupVal.value)
  const dist = isNaN(raw) || raw <= 0 ? null : raw
  const existing = getEntry(row, dateISO)
  if (!existing) { distPopupKey.value = ''; return }
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    const { data } = await timeEntriesApi.update(existing.id, {
      customer_id: row.customer_id || null,
      project_id:  row.project_id  || null,
      date:        dateISO,
      minutes:     existing.minutes,
      description: row.description,
      is_holiday:  existing.is_holiday,
      start_time:  existing.start_time || null,
      end_time:    existing.end_time   || null,
      distance:    dist,
    })
    const idx = rawEntries.value.findIndex(e => e.id === existing.id)
    rawEntries.value[idx] = data
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
    distPopupKey.value = ''
  }
}

async function clearDistPopup(row, dateISO) {
  const existing = getEntry(row, dateISO)
  if (!existing) { distPopupKey.value = ''; return }
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    const { data } = await timeEntriesApi.update(existing.id, {
      customer_id: row.customer_id || null,
      project_id:  row.project_id  || null,
      date:        dateISO,
      minutes:     existing.minutes,
      description: row.description,
      is_holiday:  existing.is_holiday,
      start_time:  existing.start_time || null,
      end_time:    existing.end_time   || null,
      distance:    null,
    })
    const idx = rawEntries.value.findIndex(e => e.id === existing.id)
    rawEntries.value[idx] = data
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
    distPopupKey.value = ''
  }
}

// ── Per-cell time range popup ─────────────────────────────────────────────
const timePopupKey   = ref('')        // row.key + dateISO when open
const timePopupFlip  = ref(false)     // true → open downward instead of upward
const timePopupStart = ref('')
const timePopupEnd   = ref('')
const timePopupRef   = ref(null)

// Format minutes-since-midnight as "HH:MM" for storage/display.
function fmtWallClock(mins) {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}

function openTimePopup(row, dateISO, event) {
  const key = row.key + dateISO
  if (timePopupKey.value === key) {
    timePopupKey.value = ''
    return
  }
  const existing = getEntry(row, dateISO)
  timePopupStart.value = existing?.start_time || ''
  timePopupEnd.value   = existing?.end_time   || ''
  const rect = event?.currentTarget?.getBoundingClientRect()
  const scrollEl = event?.currentTarget?.closest('.tt-scroll')
  const scrollTop = scrollEl ? scrollEl.getBoundingClientRect().top : 0
  timePopupFlip.value = rect ? (rect.top - scrollTop) < 200 : false
  timePopupKey.value   = key
}

// Parse a wall-clock string to minutes-since-midnight.
// Accepts "H:MM", "HH:MM", "HH:" (→ HH:00), or bare "HH" digits (→ HH:00).
// Returns -1 when invalid.
function parseWallClock(s) {
  if (!s) return -1
  const trimmed = s.trim()
  const full = trimmed.match(/^(\d{1,2}):(\d{0,2})$/)
  if (full) {
    const h = parseInt(full[1])
    const min = full[2] ? parseInt(full[2]) : 0
    if (h > 23 || min > 59) return -1
    return h * 60 + min
  }
  // Bare 1-2 digits: "20" → 20:00, "9" → 09:00
  const bare = trimmed.match(/^(\d{1,2})$/)
  if (bare) {
    const h = parseInt(bare[1])
    if (h > 23) return -1
    return h * 60
  }
  return -1
}

// Auto-insert colon while typing (e.g. "09" → "09:" after 2 digits).
function onTimePopupInput(field, event) {
  let val = event.target.value.replace(/[^\d:]/g, '')
  // Keep only the first colon; absorb any duplicate the user typed
  const colon = val.indexOf(':')
  if (colon >= 0) {
    val = val.slice(0, colon + 1) + val.slice(colon + 1).replace(/:/g, '')
  }
  // Auto-insert colon after 2 digits if user hasn't typed one yet
  if (val.length === 2 && !val.includes(':')) {
    val = val + ':'
  }
  if (val !== event.target.value) event.target.value = val
  if (field === 'start') timePopupStart.value = val
  else timePopupEnd.value = val
}

function timePopupOvernight() {
  const s = parseWallClock(timePopupStart.value)
  const e = parseWallClock(timePopupEnd.value)
  return s >= 0 && e >= 0 && e <= s
}

function timePopupMinutes() {
  const s = parseWallClock(timePopupStart.value)
  const e = parseWallClock(timePopupEnd.value)
  if (s < 0 || e < 0) return null
  if (e > s) return e - s
  if (e === s) return null
  return (24 * 60 - s) + e
}

async function applyTimePopup(row, dateISO) {
  // Normalise to HH:MM before saving so storage is always consistent.
  const sm = parseWallClock(timePopupStart.value)
  const em = parseWallClock(timePopupEnd.value)
  const start = sm >= 0 ? fmtWallClock(sm) : null
  const end   = em >= 0 ? fmtWallClock(em) : null
  let calcMins = null
  if (start && end && sm >= 0 && em >= 0 && sm !== em) {
    calcMins = em > sm ? (em - sm) : ((24 * 60 - sm) + em)
  }

  const existing = getEntry(row, dateISO)
  const before = snapshotCell(row, dateISO)
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    let changed = false
    if (existing) {
      const minutes = calcMins !== null ? calcMins : existing.minutes
      const { data } = await timeEntriesApi.update(existing.id, {
        customer_id: row.customer_id || null,
        project_id:  row.project_id  || null,
        date:        dateISO,
        minutes,
        description: row.description,
        is_holiday:  existing.is_holiday,
        start_time:  start,
        end_time:    end,
      })
      const idx = rawEntries.value.findIndex(e => e.id === existing.id)
      rawEntries.value[idx] = data
      changed = true
    } else if (calcMins !== null) {
      const { data } = await timeEntriesApi.create({
        customer_id:  row.customer_id || null,
        project_id:   row.project_id  || null,
        date:         dateISO,
        minutes:      calcMins,
        description:  row.description,
        start_time:   start,
        end_time:     end,
      })
      rawEntries.value.push(data)
      localRows.value = localRows.value.filter(r => r.key !== row.key)
      changed = true
    }
    if (changed) pushUndoBatch([{ row, dateISO, before }])
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
    timePopupKey.value = ''
  }
}

async function clearTimePopup(row, dateISO) {
  const existing = getEntry(row, dateISO)
  if (!existing) { timePopupKey.value = ''; return }
  const before = snapshotCell(row, dateISO)
  const ck = row.key + dateISO
  savingCell.value = ck
  try {
    const { data } = await timeEntriesApi.update(existing.id, {
      customer_id: row.customer_id || null,
      project_id:  row.project_id  || null,
      date:        dateISO,
      minutes:     existing.minutes,
      description: row.description,
      is_holiday:  existing.is_holiday,
      start_time:  null,
      end_time:    null,
    })
    const idx = rawEntries.value.findIndex(e => e.id === existing.id)
    rawEntries.value[idx] = data
    pushUndoBatch([{ row, dateISO, before }])
  } catch {
    ui.error(t('timeTracking.save_error'))
  } finally {
    if (savingCell.value === ck) savingCell.value = ''
    timePopupKey.value = ''
  }
}

// ── Standby shift (multi-day) ─────────────────────────────────────────────
const standbyRow = ref(null)
const savingStandby = ref(false)
const standbyForm = ref({ start_date: '', start_time: '19:00', end_date: '', end_time: '07:00' })
const displayStandbyStartDate = ref('')
const displayStandbyEndDate = ref('')

function syncStandbyDateDisplays() {
  displayStandbyStartDate.value = standbyForm.value.start_date ? formatDate(standbyForm.value.start_date) : ''
  displayStandbyEndDate.value = standbyForm.value.end_date ? formatDate(standbyForm.value.end_date) : ''
}

function _parseStandbyDate(displayRef, isoKey) {
  const val = displayRef.value.trim()
  if (!val) {
    standbyForm.value[isoKey] = ''
    return
  }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY')
  const mPos = fmt.indexOf('MM')
  const dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4), 10)
  const m = parseInt(val.slice(mPos, mPos + 2), 10)
  const d = parseInt(val.slice(dPos, dPos + 2), 10)
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayRef.value = standbyForm.value[isoKey] ? formatDate(standbyForm.value[isoKey]) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  standbyForm.value[isoKey] = iso
  displayRef.value = formatDate(iso)
}

function parseStandbyStartDate() { _parseStandbyDate(displayStandbyStartDate, 'start_date') }
function parseStandbyEndDate() { _parseStandbyDate(displayStandbyEndDate, 'end_date') }

function onStandbyStartDateChange(e) {
  standbyForm.value.start_date = e.target.value
  displayStandbyStartDate.value = e.target.value ? formatDate(e.target.value) : ''
}

function onStandbyEndDateChange(e) {
  standbyForm.value.end_date = e.target.value
  displayStandbyEndDate.value = e.target.value ? formatDate(e.target.value) : ''
}

const standbySegments = computed(() => {
  if (!standbyForm.value.start_date || !standbyForm.value.end_date) return []
  return splitShiftIntoDayEntries(
    standbyForm.value.start_date,
    standbyForm.value.start_time,
    standbyForm.value.end_date,
    standbyForm.value.end_time,
  )
})

const standbyPreview = computed(() => {
  const segs = standbySegments.value
  if (!segs.length) return ''
  const total = segs.reduce((s, seg) => s + seg.minutes, 0)
  return t('timeTracking.standby_preview', { count: segs.length, hours: fmtTime(total) })
})

function onStandbyTimeInput(field, event) {
  let val = event.target.value.replace(/[^\d:]/g, '')
  // Keep only the first colon; absorb any duplicate the user typed
  const colon = val.indexOf(':')
  if (colon >= 0) {
    val = val.slice(0, colon + 1) + val.slice(colon + 1).replace(/:/g, '')
  }
  if (val.length === 2 && !val.includes(':')) {
    val = val + ':'
  }
  if (val !== event.target.value) event.target.value = val
  if (field === 'start') standbyForm.value.start_time = val
  else standbyForm.value.end_time = val
}

function openStandbyShift(row) {
  standbyRow.value = row
  const defaults = weekendStandbyDefaults(weekDays.value.map(d => d.iso))
  standbyForm.value = { ...defaults }
  syncStandbyDateDisplays()
}

function closeStandbyShift() {
  standbyRow.value = null
  savingStandby.value = false
  displayStandbyStartDate.value = ''
  displayStandbyEndDate.value = ''
}

function applyStandbyPreset() {
  const defaults = weekendStandbyDefaults(weekDays.value.map(d => d.iso))
  standbyForm.value = { ...defaults }
  syncStandbyDateDisplays()
}

async function applyStandbyShift() {
  parseStandbyStartDate()
  parseStandbyEndDate()
  const row = standbyRow.value
  const segs = standbySegments.value
  if (!row || !segs.length) {
    ui.error(t('timeTracking.standby_invalid_range'))
    return
  }
  if (parseShiftWallClock(standbyForm.value.start_time) < 0 || parseShiftWallClock(standbyForm.value.end_time) < 0) {
    ui.error(t('timeTracking.standby_invalid_range'))
    return
  }

  savingStandby.value = true
  const undoItems = segs.map(seg => ({
    row,
    dateISO: seg.date,
    before: snapshotCell(row, seg.date),
  }))
  try {
    for (const seg of segs) {
      const existing = getEntry(row, seg.date)
      if (existing) {
        const { data } = await timeEntriesApi.update(existing.id, {
          customer_id: row.customer_id || null,
          project_id:  row.project_id  || null,
          date:        seg.date,
          minutes:     seg.minutes,
          description: row.description,
          is_holiday:  existing.is_holiday,
          start_time:  seg.start_time,
          end_time:    seg.end_time,
        })
        const idx = rawEntries.value.findIndex(e => e.id === existing.id)
        rawEntries.value[idx] = data
      } else {
        const { data } = await timeEntriesApi.create({
          customer_id:  row.customer_id || null,
          project_id:   row.project_id  || null,
          date:         seg.date,
          minutes:      seg.minutes,
          description:  row.description,
          start_time:   seg.start_time,
          end_time:     seg.end_time,
        })
        rawEntries.value.push(data)
      }
    }
    localRows.value = localRows.value.filter(r => r.key !== row.key)
    pushUndoBatch(undoItems)
    ui.success(t('timeTracking.standby_success'))
    closeStandbyShift()
    loadWeek()
  } catch {
    ui.error(t('timeTracking.standby_error'))
  } finally {
    savingStandby.value = false
  }
}

function onTimePopupDocClick(e) {
  if (!timePopupKey.value) return
  // timePopupRef may be an array when inside v-for
  const el = Array.isArray(timePopupRef.value) ? timePopupRef.value[0] : timePopupRef.value
  if (el && !el.contains(e.target) && !e.target.closest('.cell-time-toggle')) {
    timePopupKey.value = ''
  }
}

function onDistPopupDocClick(e) {
  if (!distPopupKey.value) return
  const el = Array.isArray(distPopupRef.value) ? distPopupRef.value[0] : distPopupRef.value
  if (el && !el.contains(e.target) && !e.target.closest('.cell-dist-toggle')) {
    distPopupKey.value = ''
  }
}

// ── Add row ───────────────────────────────────────────────────────────────
const addingRow   = ref(false)
const newDescRef  = ref(null)
const newRow      = ref({ customer_id: null, project_id: null, contract_id: null, description: '' })
const macroEditorOpen = ref(false)
const macroRunOpen = ref(false)
const macroRunRef = ref(null)
const macroStorageKey = 'timeTracking.macroTemplates.v2'

function makeMacroRow(description = '') {
  return {
    customer_id: null,
    project_id: null,
    description,
    day1_minutes: '',
    day1_start: '',
    day1_end: '',
    day2_minutes: '',
    day2_start: '',
    day2_end: '',
    day1_distance: '',
    day2_distance: '',
  }
}

function defaultMacroRows() {
  return [
    makeMacroRow('Travel to location'),
    makeMacroRow('Preparing for teaching'),
    makeMacroRow('Teaching'),
    makeMacroRow('Travel home'),
  ]
}

function normalizeMacroApplyDays(value) {
  return Math.max(1, Math.min(7, Number(value) || 5))
}

function makeMacroTemplate(id, name = 'Teaching block') {
  return {
    id,
    name,
    apply_days: 5,
    alternating: false,
    rows: defaultMacroRows(),
  }
}

function normalizeMacroRow(rawRow, fallbackDescription = '') {
  const base = makeMacroRow(fallbackDescription)
  const row = { ...base, ...rawRow }
  row.customer_id = row.customer_id != null && row.customer_id !== '' ? Number(row.customer_id) : null
  row.project_id = row.project_id != null && row.project_id !== '' ? Number(row.project_id) : null
  row.day1_minutes = row.day1_minutes == null ? '' : String(row.day1_minutes)
  row.day2_minutes = row.day2_minutes == null ? '' : String(row.day2_minutes)
  row.day1_start = row.day1_start == null ? '' : String(row.day1_start)
  row.day1_end = row.day1_end == null ? '' : String(row.day1_end)
  row.day2_start = row.day2_start == null ? '' : String(row.day2_start)
  row.day2_end = row.day2_end == null ? '' : String(row.day2_end)
  row.day1_distance = row.day1_distance == null ? '' : String(row.day1_distance)
  row.day2_distance = row.day2_distance == null ? '' : String(row.day2_distance)
  return row
}

function cloneMacroTemplate(tpl) {
  return {
    id: tpl.id,
    name: tpl.name,
    apply_days: normalizeMacroApplyDays(tpl.apply_days),
    alternating: !!tpl.alternating,
    rows: tpl.rows.map((r, idx) => normalizeMacroRow(r, defaultMacroRows()[idx]?.description || '')),
  }
}

function createDefaultMacroLibrary() {
  const first = makeMacroTemplate(1, 'Teaching block')
  return { nextId: 2, macros: [first] }
}

function parseMacroLibraryPayload(parsed) {
  if (!parsed || !Array.isArray(parsed.macros) || !parsed.macros.length) return null
  const macros = parsed.macros.map((m, idx) => ({
    id: Number(m.id) || (idx + 1),
    name: String(m.name || `Macro ${idx + 1}`),
    apply_days: normalizeMacroApplyDays(m.apply_days),
    alternating: !!m.alternating,
    rows: Array.isArray(m.rows)
      ? m.rows.map((r, ri) => normalizeMacroRow(r, defaultMacroRows()[ri]?.description || ''))
      : defaultMacroRows(),
  }))
  const maxId = macros.reduce((mx, m) => Math.max(mx, m.id), 0)
  return {
    nextId: Math.max(Number(parsed.nextId) || 1, maxId + 1),
    macros,
  }
}

function loadMacroLibrary() {
  try {
    const raw = localStorage.getItem(macroStorageKey)
    if (!raw) return createDefaultMacroLibrary()
    const parsed = JSON.parse(raw)
    return parseMacroLibraryPayload(parsed) || createDefaultMacroLibrary()
  } catch {
    return createDefaultMacroLibrary()
  }
}

const macroLibrary = ref(loadMacroLibrary())
const selectedMacroId = ref(macroLibrary.value.macros[0]?.id || null)
const activeMacro = ref(cloneMacroTemplate(macroLibrary.value.macros[0] || makeMacroTemplate(1)))
const macroRunId = ref(macroLibrary.value.macros[0]?.id || null)
const macroRunStartDay = ref(0)
const macroImportRef = ref(null)
const macrosSorted = computed(() =>
  [...macroLibrary.value.macros].sort((a, b) =>
    (a.name || '').localeCompare(b.name || '', undefined, { sensitivity: 'base' }),
  ),
)
function macroById(id) {
  if (id == null || id === '') return null
  const numId = Number(id)
  return macroLibrary.value.macros.find(m => m.id === numId || m.id === id) || null
}
function defaultMacroStartDayIndex() {
  const todayIdx = weekDays.value.findIndex(d => d.isToday)
  return todayIdx >= 0 ? todayIdx : 0
}
function toggleMacroRun() {
  macroRunOpen.value = !macroRunOpen.value
  if (macroRunOpen.value) {
    if (!macroRunId.value || !macroById(macroRunId.value)) {
      macroRunId.value = macrosSorted.value[0]?.id ?? null
    }
    macroRunStartDay.value = defaultMacroStartDayIndex()
  }
}
function openMacroEditor() {
  macroRunOpen.value = false
  if (macroRunId.value && macroById(macroRunId.value)) {
    selectedMacroId.value = macroRunId.value
    activeMacro.value = cloneMacroTemplate(macroById(macroRunId.value))
  }
  macroEditorOpen.value = true
}
function onMacroRunDocClick(e) {
  if (macroRunRef.value && !macroRunRef.value.contains(e.target)) {
    macroRunOpen.value = false
  }
}
const macroRunPreview = computed(() => {
  const macro = macroById(macroRunId.value)
  if (!macro) return t('timeTracking.macro_run_choose')
  const dayCount = normalizeMacroApplyDays(macro.apply_days)
  const startIdx = Math.max(0, Math.min(6, Number(macroRunStartDay.value) || 0))
  const days = weekDays.value.slice(startIdx, startIdx + dayCount)
  const labels = days.map(d => `${d.abbr} ${d.mmdd}`)
  if (!labels.length) return t('timeTracking.macro_run_no_days')
  const countLabel = labels.length === 1
    ? t('timeTracking.macro_day_one')
    : t('timeTracking.macro_day_many', { count: labels.length })
  const dayList = `${labels.join(', ')} (${countLabel})`
  if (macro.alternating) {
    const patternLabels = days.map((_, i) =>
      i % 2 === 0 ? t('timeTracking.macro_pattern_a') : t('timeTracking.macro_pattern_b'),
    ).join(', ')
    return t('timeTracking.macro_run_fill_alt', { days: dayList, patterns: patternLabels })
  }
  return t('timeTracking.macro_run_fill_same', { days: dayList })
})
const macroApplyPreview = computed(() => {
  const dayCount = normalizeMacroApplyDays(activeMacro.value.apply_days)
  const countLabel = dayCount === 1
    ? t('timeTracking.macro_day_one')
    : t('timeTracking.macro_day_many', { count: dayCount })
  if (activeMacro.value.alternating) {
    return t('timeTracking.macro_apply_preview_alt', { count: countLabel })
  }
  return t('timeTracking.macro_apply_preview_same', { count: countLabel })
})

function applyMacroLibraryState(lib) {
  macroLibrary.value = lib
  if (!macroById(selectedMacroId.value)) {
    selectedMacroId.value = lib.macros[0]?.id ?? null
  }
  const selected = macroById(selectedMacroId.value)
  if (selected) activeMacro.value = cloneMacroTemplate(selected)
  if (!macroById(macroRunId.value)) {
    macroRunId.value = lib.macros[0]?.id ?? null
  }
}

async function loadMacroLibraryFromServer() {
  try {
    const { data } = await timeEntriesApi.getMacroLibrary()
    const lib = parseMacroLibraryPayload(data?.library)
    if (!lib) return
    applyMacroLibraryState(lib)
    localStorage.setItem(macroStorageKey, JSON.stringify(lib))
  } catch {
    // keep local fallback
  }
}

watch(selectedMacroId, (id) => {
  const found = macroLibrary.value.macros.find(m => m.id === id)
  if (found) activeMacro.value = cloneMacroTemplate(found)
})

const newRowContracts = computed(() => {
  const id = newRow.value.customer_id
  if (!id) return []
  return contractsByCustomer.value[id] || []
})

const newRowProjects = computed(() => {
  if (!newRow.value.customer_id) return allProjects.value
  if (ttCustomers.value.some(c => c.id === newRow.value.customer_id))
    return ttProjects.value
  return [...projects.value.filter(p => p.customer_id === newRow.value.customer_id), ...ttProjects.value]
})

function macroProjectsForRow(row) {
  if (!row.customer_id) return allProjects.value
  if (ttCustomers.value.some(c => c.id === row.customer_id)) return ttProjects.value
  return [...projects.value.filter(p => p.customer_id === row.customer_id), ...ttProjects.value]
}

function onMacroCustomerChange(row) {
  row.project_id = null
}

function addMacroRow() {
  activeMacro.value.rows.push(makeMacroRow(''))
}

function removeMacroRow(idx) {
  if (activeMacro.value.rows.length <= 1) return
  activeMacro.value.rows.splice(idx, 1)
}

function createMacroTemplate() {
  const id = macroLibrary.value.nextId++
  const tpl = makeMacroTemplate(id, t('timeTracking.macro_default_name', { id }))
  macroLibrary.value.macros.push(cloneMacroTemplate(tpl))
  selectedMacroId.value = id
  activeMacro.value = cloneMacroTemplate(tpl)
  persistMacroLibrary()
}

function duplicateMacroTemplate() {
  const id = macroLibrary.value.nextId++
  const base = activeMacro.value
  const tpl = {
    id,
    name: `${base.name || t('timeTracking.macro_label')}${t('timeTracking.macro_copy_suffix')}`,
    apply_days: normalizeMacroApplyDays(base.apply_days),
    alternating: !!base.alternating,
    rows: base.rows.map((r, idx) => normalizeMacroRow(r, defaultMacroRows()[idx]?.description || '')),
  }
  macroLibrary.value.macros.push(cloneMacroTemplate(tpl))
  selectedMacroId.value = id
  activeMacro.value = cloneMacroTemplate(tpl)
  persistMacroLibrary()
}

function deleteMacroTemplate() {
  if (macroLibrary.value.macros.length <= 1) return
  const idx = macroLibrary.value.macros.findIndex(m => m.id === selectedMacroId.value)
  if (idx < 0) return
  const deletedId = macroLibrary.value.macros[idx].id
  macroLibrary.value.macros.splice(idx, 1)
  const next = macroLibrary.value.macros[Math.max(0, idx - 1)]
  selectedMacroId.value = next.id
  activeMacro.value = cloneMacroTemplate(next)
  if (macroRunId.value === deletedId) macroRunId.value = next.id
  persistMacroLibrary()
}

function persistMacroLibrary() {
  localStorage.setItem(macroStorageKey, JSON.stringify(macroLibrary.value))
  timeEntriesApi.setMacroLibrary(macroLibrary.value).catch(() => {})
}

function openMacroImport() {
  macroImportRef.value?.click()
}

async function exportMacroLibrary() {
  try {
    const data = JSON.stringify(macroLibrary.value, null, 2)
    await triggerDownload(data, 'warmdesk-time-macros.json', 'application/json')
  } catch {
    ui.error(t('timeTracking.macro_export_failed'))
  }
}

async function importMacroLibrary(event) {
  try {
    const file = event?.target?.files?.[0]
    if (!file) return
    const text = await file.text()
    const parsed = JSON.parse(text)
    const lib = parseMacroLibraryPayload(parsed)
    if (!lib) throw new Error('invalid')
    macroLibrary.value = lib
    selectedMacroId.value = lib.macros[0].id
    activeMacro.value = cloneMacroTemplate(lib.macros[0])
    persistMacroLibrary()
    ui.success(t('timeTracking.macro_imported'))
  } catch {
    ui.error(t('timeTracking.macro_import_invalid'))
  } finally {
    if (event?.target) event.target.value = ''
  }
}

function saveMacroTemplate() {
  const name = (activeMacro.value.name || '').trim() || t('timeTracking.macro_default_name', { id: activeMacro.value.id })
  const payload = {
    id: activeMacro.value.id,
    name,
    apply_days: normalizeMacroApplyDays(activeMacro.value.apply_days),
    alternating: !!activeMacro.value.alternating,
    rows: activeMacro.value.rows.map((r, idx) => normalizeMacroRow(r, defaultMacroRows()[idx]?.description || '')),
  }
  const idx = macroLibrary.value.macros.findIndex(m => m.id === payload.id)
  if (idx >= 0) macroLibrary.value.macros[idx] = payload
  else macroLibrary.value.macros.push(payload)
  selectedMacroId.value = payload.id
  activeMacro.value = cloneMacroTemplate(payload)
  persistMacroLibrary()
  ui.success(t('timeTracking.macro_saved'))
}

function ensureMacroRow(meta) {
  const cust = allCustomers.value.find(c => c.id === meta.customer_id)
  const proj = allProjects.value.find(p => p.id === meta.project_id)
  const k = rowKey(meta.customer_id, meta.project_id, meta.description)
  let row = allRows.value.find(x => x.key === k)
  if (!row) {
    const newLocal = {
      key:           k,
      customer_id:   meta.customer_id,
      customer_name: cust?.name || '',
      project_id:    meta.project_id,
      project_name:  proj?.name || '',
      contract_id:   null,
      contract_name: '',
      description:   meta.description,
      is_holiday:    false,
    }
    localRows.value.push(newLocal)
    row = newLocal
  }
  return row
}

function normalizeMacroTimePair(startRaw, endRaw) {
  const start = String(startRaw || '').trim()
  const end = String(endRaw || '').trim()
  if (!start && !end) return { start: null, end: null, valid: true, hasValue: false }
  const valid = parseShiftWallClock(start) >= 0 && parseShiftWallClock(end) >= 0
  if (!valid) return { start: null, end: null, valid: false, hasValue: true }
  return { start, end, valid: true, hasValue: true }
}

async function upsertMacroCell(row, dateISO, minutes, distance, startRaw, endRaw) {
  const existing = getEntry(row, dateISO)
  const tp = normalizeMacroTimePair(startRaw, endRaw)
  if (!tp.valid) throw new Error('invalid-time-range')
  if (minutes <= 0 && !distance) {
    if (existing) {
      await timeEntriesApi.remove(existing.id)
      rawEntries.value = rawEntries.value.filter(e => e.id !== existing.id)
      ensureLocalRow(row)
    }
    return
  }
  const payload = {
    customer_id: row.customer_id || null,
    project_id: row.project_id || null,
    contract_id: row.contract_id || null,
    date: dateISO,
    minutes,
    description: row.description,
    is_holiday: existing?.is_holiday || false,
    start_time: tp.hasValue ? tp.start : (existing?.start_time || null),
    end_time: tp.hasValue ? tp.end : (existing?.end_time || null),
    distance: distance || null,
  }
  if (existing) {
    const { data } = await timeEntriesApi.update(existing.id, payload)
    const idx = rawEntries.value.findIndex(e => e.id === existing.id)
    if (idx >= 0) rawEntries.value[idx] = data
  } else {
    const { data } = await timeEntriesApi.create(payload)
    rawEntries.value.push(data)
    localRows.value = localRows.value.filter(r => r.key !== row.key)
  }
}

async function applyMacroTemplate(macro = activeMacro.value, startDayIndex = 0) {
  const dayCount = normalizeMacroApplyDays(macro.apply_days)
  const alternating = !!macro.alternating
  const startIdx = Math.max(0, Math.min(6, Number(startDayIndex) || 0))
  const targetDays = weekDays.value.slice(startIdx, startIdx + dayCount).map(d => d.iso).filter(Boolean)
  if (!targetDays.length) return
  const undoItems = []
  try {
    for (const rowDef of macro.rows) {
      const desc = (rowDef.description || '').trim()
      if (!desc) continue
      const row = ensureMacroRow({
        customer_id: rowDef.customer_id != null ? Number(rowDef.customer_id) : null,
        project_id: rowDef.project_id != null ? Number(rowDef.project_id) : null,
        description: desc,
      })
      const dayPattern = [
        {
          minutes: parseMacroTimeInput(rowDef.day1_minutes, timeNotation.value),
          distance: Math.max(0, parseFloat(rowDef.day1_distance) || 0),
          start: rowDef.day1_start,
          end: rowDef.day1_end,
        },
        {
          minutes: parseMacroTimeInput(rowDef.day2_minutes, timeNotation.value),
          distance: Math.max(0, parseFloat(rowDef.day2_distance) || 0),
          start: rowDef.day2_start,
          end: rowDef.day2_end,
        },
      ]
      for (let di = 0; di < targetDays.length; di++) {
        const dateISO = targetDays[di]
        const pattern = alternating ? dayPattern[di % 2] : dayPattern[0]
        undoItems.push({ row, dateISO, before: snapshotCell(row, dateISO) })
        await upsertMacroCell(row, dateISO, pattern.minutes, pattern.distance, pattern.start, pattern.end)
      }
    }
    if (undoItems.length) {
      pushUndoBatch(undoItems)
      refreshCellInputs()
      macroRunOpen.value = false
      ui.success(t('timeTracking.macro_applied'))
    }
  } catch (err) {
    if (err?.message === 'invalid-time-range') ui.error(t('timeTracking.macro_invalid_time'))
    else ui.error(t('timeTracking.save_error'))
  }
}

async function runMacroFromPopout() {
  const macro = macroById(macroRunId.value)
  if (!macro) {
    ui.error(t('timeTracking.macro_choose_to_run'))
    return
  }
  await applyMacroTemplate(macro, macroRunStartDay.value)
}

const newRowHasSlots = computed(() => {
  if (!newRow.value.contract_id) return false
  const contracts = contractsByCustomer.value[newRow.value.customer_id] || []
  const c = contracts.find(x => x.id === newRow.value.contract_id)
  return !!(c?.time_slots?.length)
})

watch(() => newRow.value.customer_id, (id) => {
  newRow.value.contract_id = null
  loadContractsForCustomer(id)
})

watch(() => editForm.value.customer_id, (id) => {
  editForm.value.contract_id = null
  loadContractsForCustomer(id)
})

function startAddRow() {
  newRow.value = { customer_id: null, project_id: null, contract_id: null, description: '' }
  addingRow.value = true
  nextTick(() => newDescRef.value?.focus())
}

function confirmNewRow() {
  const r = newRow.value
  const cust = allCustomers.value.find(c => c.id === r.customer_id)
  const proj = allProjects.value.find(p => p.id === r.project_id)
  const contracts = contractsByCustomer.value[r.customer_id] || []
  const cont = contracts.find(c => c.id === r.contract_id)
  const k = rowKey(r.customer_id, r.project_id, r.description)
  if (!allRows.value.find(x => x.key === k)) {
    localRows.value.push({
      key:           k,
      customer_id:   r.customer_id,
      customer_name: cust?.name || '',
      project_id:    r.project_id,
      project_name:  proj?.name || '',
      contract_id:   r.contract_id || null,
      contract_name: cont?.name || '',
      description:   r.description,
      is_holiday:    false,
    })
  }
  addingRow.value = false
  return k
}

function mergeIntervals(intervals) {
  if (!intervals.length) return []
  const sorted = [...intervals].sort((a, b) => a[0] - b[0])
  const merged = [[sorted[0][0], sorted[0][1]]]
  for (let i = 1; i < sorted.length; i++) {
    const [s, e] = sorted[i]
    const last = merged[merged.length - 1]
    if (s <= last[1]) { if (e > last[1]) last[1] = e }
    else merged.push([s, e])
  }
  return merged
}

async function fillFromSlots(row) {
  if (!row.contract_id) return
  if (!contractsByCustomer.value[row.customer_id]) {
    await loadContractsForCustomer(row.customer_id)
  }
  const contracts = contractsByCustomer.value[row.customer_id] || []
  const contract = contracts.find(c => c.id === row.contract_id)
  if (!contract?.time_slots?.length) {
    ui.error(t('timeTracking.no_slots_on_contract'))
    return
  }
  const saves = []
  for (const day of weekDays.value) {
    const dow = (new Date(day.iso).getDay() + 6) % 7 // 0=Mon…6=Sun

    // Slots that START on this day (slotDayTypeMatches = true).
    // Multi-day overnight slots (e.g. Fri 19:00→Mon 07:00) only match on the anchor day,
    // so Sat/Sun have no starting slots even though they are fully covered.
    const startingSlots = contract.time_slots.filter(s => slotDayTypeMatches(s.day_type, dow))
    // Whole-day slots (empty start/end) aren't computed by slotCoverageOnWeekday.
    const hasWholeDaySlot = startingSlots.some(s => !s.start_time && !s.end_time)

    const allIntervals = []
    for (const slot of contract.time_slots) {
      for (const iv of slotCoverageOnWeekday(slot, dow)) allIntervals.push(iv)
    }
    const merged = mergeIntervals(allIntervals)
    let totalMinutes = merged.reduce((s, [a, b]) => s + b - a, 0)
    if (hasWholeDaySlot) totalMinutes = 1440

    if (totalMinutes <= 0) continue
    const existing = getEntry(row, day.iso)
    const payload = {
      customer_id: row.customer_id || null,
      project_id:  row.project_id  || null,
      contract_id: row.contract_id || null,
      date:        day.iso,
      minutes:     totalMinutes,
      description: row.description,
      is_holiday:  false,
    }
    const startingTimedSlots = startingSlots.filter(s => s.start_time && s.end_time)
    if (startingTimedSlots.length >= 1) {
      // A slot starts on this day with explicit times — pick the best one.
      const uniqueTimes = new Set(startingTimedSlots.map(s => `${s.start_time}/${s.end_time}`))
      let chosen
      if (uniqueTimes.size === 1) {
        chosen = startingTimedSlots[0]
      } else {
        const sorted = [...startingTimedSlots].sort((a, b) => {
          const aOv = a.end_time < a.start_time ? 1 : 0
          const bOv = b.end_time < b.start_time ? 1 : 0
          if (aOv !== bOv) return bOv - aOv           // overnight slots first
          return b.start_time.localeCompare(a.start_time) // then latest start
        })
        chosen = sorted[0]
      }
      payload.start_time = chosen.start_time
      payload.end_time   = chosen.end_time
    } else if (hasWholeDaySlot) {
      // Whole-day slot (empty times) starts here → 00:00/00:00
      payload.start_time = '00:00'
      payload.end_time   = '00:00'
    } else if (merged.length > 0 && merged[0][0] === 0) {
      // Coverage starts at midnight: this day is an overnight continuation from a
      // prior anchor day (e.g. Fri 19:00→Mon 07:00 covering Sat and Sun fully).
      const lastEnd = merged[merged.length - 1][1]
      payload.start_time = '00:00'
      payload.end_time   = lastEnd >= 1440 ? '00:00'
        : `${String(Math.floor(lastEnd / 60)).padStart(2, '0')}:${String(lastEnd % 60).padStart(2, '0')}`
    }
    saves.push(existing
      ? timeEntriesApi.update(existing.id, payload)
      : timeEntriesApi.create(payload)
    )
  }
  if (!saves.length) {
    ui.error(t('timeTracking.no_slots_match_week'))
    return
  }
  try {
    await Promise.all(saves)
    await loadWeek()
    ui.success(t('timeTracking.slots_filled'))
  } catch {
    ui.error(t('timeTracking.save_error'))
  }
}

async function confirmAndFillFromSlots() {
  const rowData = { ...newRow.value }
  const key = confirmNewRow()
  await nextTick()
  const row = allRows.value.find(x => x.key === key)
  if (row) await fillFromSlots(row)
}

function cancelNewRow() {
  addingRow.value = false
}

// ── Copy previous week ────────────────────────────────────────────────────
const copyingPrevWeek = ref(false)

async function copyPrevWeek() {
  copyingPrevWeek.value = true
  try {
    const prevStart = new Date(weekStart.value)
    prevStart.setDate(prevStart.getDate() - 7)
    const prevEnd = new Date(prevStart)
    prevEnd.setDate(prevEnd.getDate() + 6)

    const toLocalISO = d => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const listParams = {
      from: toLocalISO(prevStart),
      to:   toLocalISO(prevEnd),
    }
    const prevWeekInfo = (() => {
      const thu = new Date(prevStart)
      thu.setDate(thu.getDate() + (weekStartDay.value === 0 ? 4 : 3))
      const d = new Date(Date.UTC(thu.getFullYear(), thu.getMonth(), thu.getDate()))
      const y0 = new Date(Date.UTC(d.getUTCFullYear(), 0, 1))
      return { year: d.getUTCFullYear(), week: Math.ceil(((d - y0) / 86400000 + 1) / 7) }
    })()
    const orderParams = { year: prevWeekInfo.year, week: prevWeekInfo.week }

    const [{ data }, orderRes] = await Promise.all([
      timeEntriesApi.list(listParams),
      timeEntriesApi.getRowOrder(orderParams).catch(() => ({ data: { keys: [] } })),
    ])

    const savedOrder = orderRes.data?.keys || []
    const prevComments = orderRes.data?.comments || {}
    const rowMap = new Map()
    for (const e of data) {
      const k = rowKey(e.customer_id, e.project_id, e.description)
      if (!rowMap.has(k)) {
        rowMap.set(k, {
          key:           k,
          customer_id:   e.customer_id,
          customer_name: e.customer?.name || '',
          project_id:    e.project_id,
          project_name:  e.project?.name || '',
          description:   e.description || '',
        })
      }
    }

    let orderedKeys
    if (savedOrder.length) {
      orderedKeys = savedOrder
    } else if (data.length) {
      orderedKeys = []
      const seen = new Set()
      for (const e of [...data].sort((a, b) => a.id - b.id)) {
        const k = rowKey(e.customer_id, e.project_id, e.description)
        if (!seen.has(k)) {
          seen.add(k)
          orderedKeys.push(k)
        }
      }
    } else {
      ui.info(t('timeTracking.copy_prev_nothing'))
      return
    }

    const existing = new Set(allRows.value.map(r => r.key))
    const toAdd = []
    for (const k of orderedKeys) {
      if (existing.has(k)) continue
      toAdd.push(rowMap.get(k) || rowFromKey(k))
      existing.add(k)
    }

    if (toAdd.length === 0) {
      ui.info(t('timeTracking.copy_prev_nothing'))
      return
    }

    if (allRows.value.length > 0) {
      const msg = toAdd.length === 1
        ? t('timeTracking.copy_prev_confirm_one')
        : t('timeTracking.copy_prev_confirm', { count: toAdd.length })
      if (!await ui.confirm(msg, { confirmLabel: t('timeTracking.copy_prev_confirm_action') })) return
    }

    const undoSnapshot = {
      addedKeys: toAdd.map(r => r.key),
      prevKeyOrder: [...(_keyOrder.value || [])],
      prevComments: { ...rowComments.value },
    }

    for (const row of toAdd) {
      localRows.value.push(row)
      if (prevComments[row.key]) {
        rowComments.value[row.key] = prevComments[row.key]
      }
    }
    const newKeys = toAdd.map(r => r.key)
    _keyOrder.value = [...(_keyOrder.value || []), ...newKeys]
    _scheduleSaveOrder(_keyOrder.value)
    pushUndoCopyPrevWeek(undoSnapshot)
  } catch {
    ui.error(t('timeTracking.copy_prev_error'))
  } finally {
    copyingPrevWeek.value = false
  }
}

// ── Report ────────────────────────────────────────────────────────────────
const now  = new Date()
const rpt  = ref({ period: 'month', year: now.getFullYear(), month: now.getMonth() + 1, week: currentISOWeek(), group_by: 'period', start_date: '', end_date: '' })
const report       = ref(null)
const loadingReport = ref(false)

const months = computed(() => {
  const fmt = new Intl.DateTimeFormat(undefined, { month: 'long' })
  return Array.from({ length: 12 }, (_, i) => fmt.format(new Date(2000, i, 1)))
})

function currentISOWeek() {
  const today = new Date()
  return isoWeekNum(today, weekStartDay.value)
}

// Params shared by the report JSON load and both export endpoints.
function reportBaseParams() {
  const params = {
    period:   rpt.value.period,
    year:     rpt.value.year,
    month:    rpt.value.month,
    week:     rpt.value.week,
    group_by: rpt.value.group_by,
  }
  if (rpt.value.period === 'custom') {
    params.start_date = rpt.value.start_date
    params.end_date = rpt.value.end_date
  }
  if (canViewOtherUsers.value) params.user_id = selectedUserId.value
  return params
}

async function loadReport() {
  loadingReport.value = true
  try {
    const { data } = await timeEntriesApi.report(reportBaseParams())
    report.value = data
  } catch {
    ui.error(t('timeTracking.load_error'))
  } finally {
    loadingReport.value = false
  }
  if (reportViewMode.value === 'chart') {
    if (reportChartType.value === 'stacked') await ensureReportChartPeriodData()
    await nextTick(); renderReportChart()
  }
}

// ── Report chart ─────────────────────────────────────────────────────────
const reportViewMode  = ref('table') // 'table' | 'chart'
const reportChartType = ref('bar')   // 'bar' | 'pie' | 'stacked'
const reportChartBasis = ref('declarable') // 'declarable' | 'total' | 'undeclarable'
const reportChartCanvas = ref(null)
let reportChartInstance = null

// The stacked-by-time chart always needs entries bucketed by period (year →
// month, month → week, week → day — see backend's buildGroups). When the
// table itself is grouped by period this is just `report`; otherwise it's
// fetched separately so the stacked chart doesn't depend on the table's own
// "Group by" choice.
const reportChartPeriodReport = ref(null)
const loadingReportChartPeriod = ref(false)

async function ensureReportChartPeriodData() {
  if (rpt.value.group_by === 'period') { reportChartPeriodReport.value = null; return }
  loadingReportChartPeriod.value = true
  try {
    const { data } = await timeEntriesApi.report({ ...reportBaseParams(), group_by: 'period' })
    reportChartPeriodReport.value = data
  } catch {
    reportChartPeriodReport.value = null
    ui.error(t('timeTracking.load_error'))
  } finally {
    loadingReportChartPeriod.value = false
  }
}

const REPORT_CHART_MAX_SLICES = 7

function cssVar(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

// Aggregates every entry across all report groups by activity (description),
// independent of the report's own group_by (period/customer/project), since
// the chart's job is a single breakdown "by activity" regardless of grouping.
const reportActivityBreakdown = computed(() => {
  if (!report.value) return []
  const totals = new Map()
  const customers = new Map()
  for (const grp of report.value.groups) {
    for (const e of grp.entries) {
      const label = e.description?.trim() || t('timeTracking.no_activity')
      const minutes = reportChartBasis.value === 'total' ? e.minutes
        : reportChartBasis.value === 'undeclarable' ? entryUndecl(e)
        : reportEntryDeclarable(e)
      totals.set(label, (totals.get(label) || 0) + minutes)
      if (!customers.has(label)) customers.set(label, new Set())
      customers.get(label).add(e.customer?.name || t('timeTracking.no_customer'))
    }
  }
  const sorted = Array.from(totals, ([label, minutes]) => ({
    label,
    minutes,
    customer: Array.from(customers.get(label) || []).join(', '),
  }))
    .filter(a => a.minutes > 0)
    .sort((a, b) => b.minutes - a.minutes)
  if (sorted.length <= REPORT_CHART_MAX_SLICES + 1) return sorted
  const top = sorted.slice(0, REPORT_CHART_MAX_SLICES)
  const otherMinutes = sorted.slice(REPORT_CHART_MAX_SLICES).reduce((s, a) => s + a.minutes, 0)
  top.push({ label: t('timeTracking.chart_other'), minutes: otherMinutes, customer: '' })
  return top
})

function reportChartColorAt(index, isOther) {
  return isOther ? cssVar('--chart-cat-other') : cssVar(`--chart-cat-${(index % REPORT_CHART_MAX_SLICES) + 1}`)
}

function reportChartColors(data) {
  return data.map((a, i) => reportChartColorAt(i, a.label === t('timeTracking.chart_other')))
}

// Same "top N + Other" breakdown as reportActivityBreakdown, but one series
// per period so activity totals can be stacked over time (year → month,
// month → week, week → day — whatever the backend bucketed for period
// grouping). Reads from the table's own `report` when it's already grouped
// by period, otherwise from the separately-fetched `reportChartPeriodReport`
// so the stacked chart works regardless of the table's "Group by" setting.
const reportStackedBreakdown = computed(() => {
  const source = rpt.value.group_by === 'period' ? report.value : reportChartPeriodReport.value
  if (!source) return null
  const periods = source.groups.map(g => g.label)
  const grandTotals = new Map()
  const customersByLabel = new Map()
  const perPeriodTotals = source.groups.map(grp => {
    const totals = new Map()
    for (const e of grp.entries) {
      const label = e.description?.trim() || t('timeTracking.no_activity')
      const minutes = reportChartBasis.value === 'total' ? e.minutes
        : reportChartBasis.value === 'undeclarable' ? entryUndecl(e)
        : reportEntryDeclarable(e)
      if (minutes <= 0) continue
      totals.set(label, (totals.get(label) || 0) + minutes)
      grandTotals.set(label, (grandTotals.get(label) || 0) + minutes)
      if (!customersByLabel.has(label)) customersByLabel.set(label, new Set())
      customersByLabel.get(label).add(e.customer?.name || t('timeTracking.no_customer'))
    }
    return totals
  })
  const ranked = Array.from(grandTotals, ([label, minutes]) => ({ label, minutes }))
    .sort((a, b) => b.minutes - a.minutes)
  const topLabels = new Set(ranked.slice(0, REPORT_CHART_MAX_SLICES).map(a => a.label))
  const otherLabel = t('timeTracking.chart_other')
  const seriesLabels = ranked.length > REPORT_CHART_MAX_SLICES
    ? [...topLabels, otherLabel]
    : [...topLabels]
  const series = seriesLabels.map(label => ({
    label,
    isOther: label === otherLabel,
    customer: label === otherLabel ? '' : Array.from(customersByLabel.get(label) || []).join(', '),
    data: perPeriodTotals.map(totals => {
      if (label === otherLabel) {
        let sum = 0
        for (const [actLabel, minutes] of totals) {
          if (!topLabels.has(actLabel)) sum += minutes
        }
        return sum / 60
      }
      return (totals.get(label) || 0) / 60
    }),
  }))
  return { periods, series }
})

const reportChartHasData = computed(() => {
  if (reportChartType.value === 'stacked') {
    const s = reportStackedBreakdown.value
    return !!s && s.periods.length > 0 && s.series.length > 0
  }
  return reportActivityBreakdown.value.length > 0
})

function renderReportChart() {
  if (!reportChartCanvas.value) return
  reportChartInstance?.destroy()
  reportChartInstance = null
  const textColor = cssVar('--color-text')

  if (reportChartType.value === 'stacked') {
    const stacked = reportStackedBreakdown.value
    if (!stacked || !stacked.series.length) return
    const colors = stacked.series.map((s, i) => reportChartColorAt(i, s.isOther))
    reportChartInstance = new Chart(reportChartCanvas.value, {
      type: 'bar',
      data: {
        labels: stacked.periods,
        datasets: stacked.series.map((s, i) => ({
          label: s.label,
          data: s.data,
          backgroundColor: colors[i] + 'cc',
          borderColor: colors[i],
          borderWidth: 1,
        })),
      },
      options: {
        responsive: true,
        color: textColor,
        plugins: {
          legend: { display: true, position: 'top', labels: { color: textColor } },
          tooltip: { callbacks: {
            title: ctx => stacked.series[ctx[0]?.datasetIndex]?.customer || '',
            label: ctx => `${ctx.dataset.label}: ${fmtTime(Math.round(ctx.parsed.y * 60))}`,
          } },
        },
        scales: {
          x: { stacked: true, ticks: { color: textColor }, grid: { display: false } },
          y: { stacked: true, beginAtZero: true, ticks: { color: textColor }, grid: { color: cssVar('--color-border') } },
        },
      },
    })
    return
  }

  const data = reportActivityBreakdown.value
  if (!data.length) return
  const colors = reportChartColors(data)
  const isPie = reportChartType.value === 'pie'
  reportChartInstance = new Chart(reportChartCanvas.value, {
    type: isPie ? 'pie' : 'bar',
    data: {
      labels: data.map(a => a.label),
      datasets: [{
        data: data.map(a => a.minutes / 60),
        backgroundColor: isPie ? colors : colors.map(c => c + 'cc'),
        borderColor: isPie ? cssVar('--color-surface') : colors,
        borderWidth: isPie ? 2 : 1,
      }],
    },
    options: {
      responsive: true,
      color: textColor,
      aspectRatio: isPie ? 1.6 : undefined,
      plugins: {
        legend: { display: isPie, position: 'right', labels: { color: textColor } },
        tooltip: { callbacks: {
          title: ctx => data[ctx[0]?.dataIndex]?.customer || '',
          label: ctx => {
            const hours = isPie ? ctx.parsed : ctx.parsed.y
            return `${ctx.label}: ${fmtTime(Math.round(hours * 60))}`
          },
        } },
      },
      scales: isPie ? undefined : {
        x: { ticks: { color: textColor }, grid: { display: false } },
        y: { beginAtZero: true, ticks: { color: textColor }, grid: { color: cssVar('--color-border') } },
      },
    },
  })
}

watch(reportChartType, async (type) => {
  if (type === 'stacked') await ensureReportChartPeriodData()
  await nextTick(); renderReportChart()
})

watch(reportChartBasis, async () => {
  await nextTick(); renderReportChart()
})

watch(reportViewMode, async (mode) => {
  if (mode !== 'chart') return
  if (reportChartType.value === 'stacked') await ensureReportChartPeriodData()
  await nextTick(); renderReportChart()
})

// ── Custom date range picker ─────────────────────────────────────────────
const customRangeModalOpen = ref(false)
const customRangeForm = ref({ start: '', end: '' })
const customRangeError = ref('')

function onReportPeriodChange() {
  if (rpt.value.period === 'custom') {
    openCustomRangeModal()
    return
  }
  loadReport()
}

function openCustomRangeModal() {
  customRangeForm.value.start = rpt.value.start_date || ''
  customRangeForm.value.end = rpt.value.end_date || ''
  customRangeError.value = ''
  customRangeModalOpen.value = true
}

function applyCustomRange() {
  if (!customRangeForm.value.start || !customRangeForm.value.end) return
  if (customRangeForm.value.end < customRangeForm.value.start) {
    customRangeError.value = t('timeTracking.custom_range_invalid')
    return
  }
  rpt.value.start_date = customRangeForm.value.start
  rpt.value.end_date = customRangeForm.value.end
  customRangeModalOpen.value = false
  loadReport()
}

const customRangeButtonLabel = computed(() => {
  if (rpt.value.start_date && rpt.value.end_date) {
    return `${formatDate(rpt.value.start_date)} – ${formatDate(rpt.value.end_date)}`
  }
  return t('timeTracking.custom_range_pick')
})

// ── Distance unit ─────────────────────────────────────────────────────────
const distanceUnit = computed(() => auth.user?.distance_unit || 'km')

function fmtDistance(val) {
  if (!val) return '0'
  const n = parseFloat(val)
  if (isNaN(n)) return '0'
  return n === Math.floor(n) ? String(Math.floor(n)) : n.toFixed(1)
}

// ── Time notation ─────────────────────────────────────────────────────────
const timeNotation = computed(() => auth.user?.time_notation || 'decimal')

function fmtTime(minutes) {
  if (!minutes) return timeNotation.value === 'hhmm' ? '0:00' : '0.00'
  if (timeNotation.value === 'hhmm') {
    const h = Math.floor(minutes / 60)
    const m = minutes % 60
    return `${h}:${String(m).padStart(2, '0')}`
  }
  return (minutes / 60).toFixed(2)
}

function parseTimeInput(val) {
  return parseTimeNotationMinutes(val, timeNotation.value)
}

// ── Exports ───────────────────────────────────────────────────────────────

// Weekly timesheet → XLSX: rows are customer/project combos, columns are days.
async function exportSheetXLSX() {
  try {
    const params = { start_date: weekDays.value[0].iso }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const data = await timeEntriesApi.sheetXLSX(params)
    await triggerDownload(data, `time-tracking-week${weekInfo.value.week}-${weekInfo.value.year}.xlsx`, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
  } catch (e) {
    console.error('[export] sheet XLSX failed:', e)
    ui.error(String(e?.message || e))
  }
}

// Weekly timesheet → PDF: delegates to backend report with period=week.
async function exportSheetPDF() {
  try {
    const params = { period: 'week', start_date: weekDays.value[0].iso, font: pdfFont.value, lang: pdfLang.value }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const data = await timeEntriesApi.reportPDF(params)
    await triggerDownload(data, `time-tracking-week${weekInfo.value.week}-${weekInfo.value.year}.pdf`, 'application/pdf')
  } catch (e) {
    console.error('[export] sheet PDF failed:', e)
    ui.error(String(e?.message || e))
  }
}

// Open grid PDF panel, initialising selectors from the current week view.
function openGridPdf() {
  if (!gridPdfOpen.value) {
    const d = new Date(weekDays.value[0].iso)
    gridPdfWeek.value = weekInfo.value.week
    gridPdfWeekYear.value = weekInfo.value.year
    gridPdfMonth.value = d.getMonth() + 1
    gridPdfMonthYear.value = d.getFullYear()
    gridPdfYear.value = d.getFullYear()
  }
  gridPdfOpen.value = !gridPdfOpen.value
}

// Export grid PDF using the period and date selected in the panel.
async function exportGridPDFFromPanel() {
  gridPdfOpen.value = false
  try {
    const params = { grid: gridPdfType.value, font: pdfFont.value, lang: pdfLang.value }
    let slug
    if (gridPdfType.value === 'week') {
      params.year = gridPdfWeekYear.value
      params.week = gridPdfWeek.value
      slug = `week${gridPdfWeek.value}-${gridPdfWeekYear.value}`
    } else if (gridPdfType.value === 'month') {
      params.year = gridPdfMonthYear.value
      params.month = gridPdfMonth.value
      slug = `month${String(gridPdfMonth.value).padStart(2, '0')}-${gridPdfMonthYear.value}`
    } else {
      params.year = gridPdfYear.value
      slug = `year${gridPdfYear.value}`
    }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const data = await timeEntriesApi.gridPDF(params)
    await triggerDownload(data, `timesheet-grid-${slug}.pdf`, 'application/pdf')
  } catch (e) {
    console.error('[export] grid PDF failed:', e)
    ui.error(String(e?.message || e))
  }
}

// Weekly timesheet → Grid PDF (week, month, or year view).
async function exportGridPDF(gridType) {
  try {
    const d = new Date(weekDays.value[0].iso)
    const params = { grid: gridType, font: pdfFont.value, lang: pdfLang.value }
    if (gridType === 'week') {
      params.start_date = weekDays.value[0].iso
    } else if (gridType === 'month') {
      params.year = d.getFullYear()
      params.month = d.getMonth() + 1
    } else {
      params.year = d.getFullYear()
    }
    if (canViewOtherUsers.value) params.user_id = selectedUserId.value
    const data = await timeEntriesApi.gridPDF(params)
    const y = params.year ?? weekInfo.value.year
    const slug = gridType === 'week' ? `week${weekInfo.value.week}-${y}` : gridType === 'month' ? `month${String(params.month).padStart(2,'0')}-${y}` : `year${y}`
    await triggerDownload(data, `timesheet-grid-${slug}.pdf`, 'application/pdf')
  } catch (e) {
    console.error('[export] grid PDF failed:', e)
    ui.error(String(e?.message || e))
  }
}

// Report tab → XLSX: date-list grouped by period.
async function exportReportXLSX() {
  if (!report.value) return
  try {
    const params = reportBaseParams()
    params.distance_unit = distanceUnit.value
    const data = await timeEntriesApi.reportXLSX(params)
    const slug = report.value.period_label.replace(/\s+/g, '-').toLowerCase()
    await triggerDownload(data, `time-tracking-${slug}.xlsx`, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
  } catch (e) {
    console.error('[export] report XLSX failed:', e)
    ui.error(String(e?.message || e))
  }
}

// Report tab → PDF: delegates to backend.
async function exportReportPDF() {
  if (!report.value) return
  try {
    const params = reportBaseParams()
    params.font = pdfFont.value
    params.lang = pdfLang.value
    if (pdfShowAbbr.value) params.show_abbr = '1'
    if (pdfPageBreak.value && rpt.value.group_by === 'customer') params.page_break = 'customer'
    if (pdfShowCosts.value) params.show_costs = '1'
    params.show_page_numbers = pdfShowPageNumbers.value ? '1' : '0'
    params.show_undeclarable = pdfShowUndeclarable.value ? '1' : '0'
    if (pdfShowUndeclarableRow.value) params.show_undeclarable_row = '1'
    if (pdfShowDistance.value) params.show_distance = '1'
    params.distance_unit = distanceUnit.value
    const data = await timeEntriesApi.reportPDF(params)
    const slug = report.value.period_label.replace(/\s+/g, '-').toLowerCase()
    await triggerDownload(data, `time-tracking-${slug}.pdf`, 'application/pdf')
  } catch (e) {
    console.error('[export] report PDF failed:', e)
    ui.error(String(e?.message || e))
  }
}

// Report tab → Chart PDF: renders the current chart type server-side.
async function exportReportChartPDF() {
  if (!report.value) return
  try {
    const params = reportBaseParams()
    params.chart_type = reportChartType.value
    params.chart_basis = reportChartBasis.value
    params.font = pdfFont.value
    params.lang = pdfLang.value
    const data = await timeEntriesApi.reportChartPDF(params)
    const slug = report.value.period_label.replace(/\s+/g, '-').toLowerCase()
    await triggerDownload(data, `time-tracking-chart-${slug}.pdf`, 'application/pdf')
  } catch (e) {
    console.error('[export] report chart PDF failed:', e)
    ui.error(String(e?.message || e))
  }
}


// ── Time-tracking-only project & customer management ─────────────────────
const managingProjects     = ref(false)
const manageTab            = ref('projects')
const modalRef             = ref(null)

// Projects
const addingTTProject      = ref(false)
const newTTProject         = ref({ name: '', color: '#6366f1', undeclStr: '' })
const editingTTProject     = ref(null)
const newTTNameRef         = ref(null)

// Customers
const addingTTCustomer     = ref(false)
const newTTCustomer        = ref({ name: '' })
const editingTTCustomer    = ref(null)
const newTTCustomerNameRef = ref(null)

function openManageProjects() {
  managingProjects.value = true
  nextTick(() => {
    const first = modalRef.value?.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])')
    first?.focus()
  })
}

function closeManageProjects() {
  managingProjects.value = false
  addingTTProject.value = false
  editingTTProject.value = null
  newTTProject.value = { name: '', color: '#6366f1', undeclStr: '' }
  addingTTCustomer.value = false
  editingTTCustomer.value = null
}

watch(addingTTProject, (v) => {
  if (v) nextTick(() => newTTNameRef.value?.focus())
})

async function loadTTProjects() {
  try {
    const { data } = await projectsApi.listTimeTracking()
    ttProjects.value = data ?? []
  } catch {
    ui.error(t('timeTracking.tt_project_load_error'))
  }
}

function isGlobalTTProject(p) {
  return p.created_by_id !== auth.user?.id
}

function canEditTTProject(p) {
  return auth.isAdmin || p.created_by_id === auth.user?.id
}

async function confirmAddTTProject() {
  const name = newTTProject.value.name.trim()
  if (!name) return
  try {
    const { data } = await projectsApi.createTimeTracking({
      name,
      color: newTTProject.value.color,
      undeclarable_minutes: parseTimeInput(newTTProject.value.undeclStr),
    })
    ttProjects.value.push(data)
    newTTProject.value = { name: '', color: '#6366f1', undeclStr: '' }
    addingTTProject.value = false
  } catch {
    ui.error(t('timeTracking.tt_project_save_error'))
  }
}

function startEditTTProject(p) {
  const undeclStr = p.undeclarable_minutes > 0 ? fmtTime(p.undeclarable_minutes) : ''
  editingTTProject.value = { id: p.id, name: p.name, color: p.color || '#6366f1', undeclStr }
}

function cancelEditTTProject() {
  editingTTProject.value = null
}

async function saveTTProject() {
  const e = editingTTProject.value
  if (!e || !e.name.trim()) return
  try {
    const { data } = await projectsApi.updateTimeTracking(e.id, {
      name: e.name.trim(),
      color: e.color,
      undeclarable_minutes: parseTimeInput(e.undeclStr),
    })
    const idx = ttProjects.value.findIndex(p => p.id === e.id)
    if (idx >= 0) ttProjects.value[idx] = data
    editingTTProject.value = null
  } catch {
    ui.error(t('timeTracking.tt_project_save_error'))
  }
}

async function deleteTTProject(p) {
  if (!await ui.confirm(t('timeTracking.tt_project_delete_confirm'), { destructive: true })) return
  try {
    await projectsApi.deleteTimeTracking(p.id)
    ttProjects.value = ttProjects.value.filter(x => x.id !== p.id)
  } catch {
    ui.error(t('timeTracking.tt_project_delete_error'))
  }
}

// TT Customer helpers
function isGlobalTTCustomer(c) {
  return c.created_by_id !== auth.user?.id
}

function canEditTTCustomer(c) {
  return auth.isAdmin || c.created_by_id === auth.user?.id
}

watch(addingTTCustomer, (v) => {
  if (v) nextTick(() => newTTCustomerNameRef.value?.focus())
})

async function confirmAddTTCustomer() {
  const name = newTTCustomer.value.name.trim()
  if (!name) return
  try {
    const { data } = await customersApi.createTimeTracking({ name })
    ttCustomers.value.push(data)
    newTTCustomer.value = { name: '' }
    addingTTCustomer.value = false
  } catch {
    ui.error(t('timeTracking.tt_customer_save_error'))
  }
}

function startEditTTCustomer(c) {
  editingTTCustomer.value = { id: c.id, name: c.name }
}

function cancelEditTTCustomer() {
  editingTTCustomer.value = null
}

async function saveTTCustomer() {
  const e = editingTTCustomer.value
  if (!e || !e.name.trim()) return
  try {
    const { data } = await customersApi.updateTimeTracking(e.id, { name: e.name.trim() })
    const idx = ttCustomers.value.findIndex(c => c.id === e.id)
    if (idx >= 0) ttCustomers.value[idx] = data
    editingTTCustomer.value = null
  } catch {
    ui.error(t('timeTracking.tt_customer_save_error'))
  }
}

async function deleteTTCustomer(c) {
  if (!await ui.confirm(t('timeTracking.tt_customer_delete_confirm'), { destructive: true })) return
  try {
    await customersApi.deleteTimeTracking(c.id)
    ttCustomers.value = ttCustomers.value.filter(x => x.id !== c.id)
  } catch {
    ui.error(t('timeTracking.tt_customer_delete_error'))
  }
}

// ── PDF options dropdown — close on outside click ─────────────────────────
function onDocClick(e) {
  if (pdfOptionsRef.value && !pdfOptionsRef.value.contains(e.target)) {
    pdfOptionsOpen.value = false
  }
  if (gridPdfRef.value && !gridPdfRef.value.contains(e.target)) {
    gridPdfOpen.value = false
  }
}
onMounted(() => {
  window.addEventListener('keydown', onWindowUndoCapture, true)
  window.addEventListener('beforeinput', onWindowBeforeInputCapture, true)
  document.addEventListener('keydown', onNavEscapeKey)
  document.addEventListener('mousedown', onDocClick)
  document.addEventListener('mousedown', onHolidaysDocClick)
  document.addEventListener('mousedown', onMacroRunDocClick)
  document.addEventListener('mousedown', onWkPickerDocClick)
  document.addEventListener('mousedown', onTimePopupDocClick)
  document.addEventListener('mousedown', onDistPopupDocClick)
})

function initRowSortable() {
  if (!tbodyEl.value) return
  if (rowSortable) {
    rowSortable.destroy()
    rowSortable = null
  }
  rowSortable = Sortable.create(tbodyEl.value, {
    animation: 150,
    handle: '.drag-handle',
    draggable: 'tr.tt-row',
    ghostClass: 'sortable-ghost',
    onEnd({ item, oldIndex, newIndex }) {
      if (oldIndex === newIndex) return
      // Revert SortableJS's DOM move before Vue re-renders so the vdom
      // patch starts from the known pre-drag order.
      const rows = Array.from(tbodyEl.value.querySelectorAll('tr.tt-row:not(.tt-newrow)'))
      if (newIndex < oldIndex) {
        tbodyEl.value.insertBefore(item, rows[oldIndex + 1] ?? null)
      } else {
        tbodyEl.value.insertBefore(item, rows[oldIndex])
      }
      const arr = [..._keyOrder.value]
      const [moved] = arr.splice(oldIndex, 1)
      arr.splice(newIndex, 0, moved)
      _keyOrder.value = arr
      _scheduleSaveOrder(arr)
    },
  })
}
onUnmounted(() => {
  reportChartInstance?.destroy()
  reportChartInstance = null
  rowSortable?.destroy()
  rowSortable = null
  window.removeEventListener('keydown', onWindowUndoCapture, true)
  window.removeEventListener('beforeinput', onWindowBeforeInputCapture, true)
  document.removeEventListener('keydown', onNavEscapeKey)
  document.removeEventListener('mousedown', onDocClick)
  document.removeEventListener('mousedown', onHolidaysDocClick)
  document.removeEventListener('mousedown', onMacroRunDocClick)
  document.removeEventListener('mousedown', onWkPickerDocClick)
  document.removeEventListener('mousedown', onTimePopupDocClick)
  document.removeEventListener('mousedown', onDistPopupDocClick)
  document.removeEventListener('pointermove', onPointerMove)
  document.removeEventListener('pointerup', onPointerUp)
  document.removeEventListener('pointercancel', onPointerUp)
  ui.setHelpContext(null)
})

// ── Init ──────────────────────────────────────────────────────────────────
onMounted(async () => {
  const fetches = [
    customersApi.list().catch(() => ({ data: [] })),
    customersApi.listTimeTracking().catch(() => ({ data: [] })),
    projectsApi.list().catch(() => ({ data: [] })),
    projectsApi.listTimeTracking().catch(() => ({ data: [] })),
  ]
  if (canViewOtherUsers.value) {
    fetches.push(client.get('/users').catch(() => ({ data: [] })))
  }
  const results = await Promise.all(fetches)
  customers.value   = results[0]?.data ?? []
  ttCustomers.value = results[1]?.data ?? []
  projects.value    = (results[2]?.data ?? []).filter(p => !p.archived)
  ttProjects.value  = results[3]?.data ?? []
  if (canViewOtherUsers.value) {
    allUsers.value = results[4].data
    selectedUserId.value = auth.user?.id ?? null
  }
  try {
    const { data } = await timeEntriesApi.getRowOrder()
    if (data.keys?.length) _serverOrder.value = data.keys
  } catch {}
  await loadMacroLibraryFromServer()
  await loadWeek()
  await nextTick()
  initRowSortable()
  await loadReport()
})

// ── Invoices ──────────────────────────────────────────────────────────────
const allInvoices       = ref([])
const loadingAllInvoices = ref(false)
const invFilterCustomer  = ref('')
const invFilterStatus    = ref('')

async function loadAllInvoices() {
  loadingAllInvoices.value = true
  try {
    const params = {}
    if (invFilterCustomer.value) params.customer_id = invFilterCustomer.value
    if (invFilterStatus.value)   params.status       = invFilterStatus.value
    const { data } = await customersApi.listAllInvoices(params)
    allInvoices.value = data || []
  } catch {
    allInvoices.value = []
  } finally {
    loadingAllInvoices.value = false
  }
}

function invPdfUrl(inv) {
  const server = getServerUrl()
  const base = server ? `${server}/api/v1` : '/api/v1'
  const lang = pdfLang.value === 'auto' ? (auth.user?.locale || 'en') : pdfLang.value
  const du = distanceUnit.value
  return `${base}/customers/${inv.customer_id}/invoices/${inv.id}/pdf?font=${pdfFont.value}&lang=${lang}&distance_unit=${du}`
}

async function changeAllInvStatus(inv, status) {
  try {
    const { data } = await customersApi.updateInvoice(inv.customer_id, inv.id, { status })
    const idx = allInvoices.value.findIndex(i => i.id === inv.id)
    if (idx >= 0) allInvoices.value[idx] = data
  } catch {
    ui.error(t('invoice.invoices'))
  }
}

// ── Create invoice from report group ─────────────────────────────────────
const showCreateInvoiceFromGroup = ref(false)
const createInvGroup   = ref(null)
const createInvStep    = ref(1)
const createInvLineItems = ref([])
const createInvForm    = ref({ period_start: '', period_end: '', due_date: '', vat_rate: 0, notes: '' })
const savingInvFromGrp = ref(false)

const createInvDisplayFrom = ref('')
const createInvDisplayTo   = ref('')
const createInvDisplayDue  = ref('')

function _parseCreateInvDate(displayRef, isoKey) {
  const val = displayRef.value.trim()
  if (!val) { createInvForm.value[isoKey] = ''; return }
  const fmt = dateOnlyFormat()
  const yPos = fmt.indexOf('YYYY'), mPos = fmt.indexOf('MM'), dPos = fmt.indexOf('DD')
  const y = parseInt(val.slice(yPos, yPos + 4))
  const m = parseInt(val.slice(mPos, mPos + 2))
  const d = parseInt(val.slice(dPos, dPos + 2))
  if (!y || m < 1 || m > 12 || d < 1 || d > 31) {
    displayRef.value = createInvForm.value[isoKey] ? formatDate(createInvForm.value[isoKey]) : ''
    return
  }
  const iso = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`
  createInvForm.value[isoKey] = iso
  displayRef.value = formatDate(iso)
}
function parseCreateInvFrom() { _parseCreateInvDate(createInvDisplayFrom, 'period_start') }
function parseCreateInvTo()   { _parseCreateInvDate(createInvDisplayTo,   'period_end') }
function parseCreateInvDue()  { _parseCreateInvDate(createInvDisplayDue,  'due_date') }

const createInvHasDistance = computed(() => createInvLineItems.value.some(li => li.distance > 0))
const createInvSubtotal    = computed(() => createInvLineItems.value.reduce((s, li) => s + li.amount, 0))
const createInvVAT         = computed(() => createInvSubtotal.value * createInvForm.value.vat_rate / 100)
const createInvTotal       = computed(() => createInvSubtotal.value + createInvVAT.value)
const createInvCurrency    = computed(() => createInvLineItems.value[0]?.currency || '€')

function fmtMinutesInv(mins) {
  if (!mins) return ''
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

function openCreateInvoiceFromGroup(grp) {
  createInvGroup.value = grp
  const entries = grp.entries || []
  const dates = entries.map(e => e.date).filter(Boolean).sort()
  const fromIso = dates[0] || new Date().toISOString().slice(0, 10)
  const toIso   = dates[dates.length - 1] || fromIso
  createInvForm.value = { period_start: fromIso, period_end: toIso, due_date: '', vat_rate: 0, notes: '' }
  createInvDisplayFrom.value = formatDate(fromIso)
  createInvDisplayTo.value   = formatDate(toIso)
  createInvDisplayDue.value  = ''
  // Build line items from entries that have a contract
  const items = []
  for (const entry of entries) {
    if (!entry.contract_id) continue
    const rate   = entry.contract?.price_per_hour ?? 0
    const km     = entry.contract?.price_per_km ?? 0
    const hours  = entry.minutes / 60
    const dist   = entry.distance ?? 0
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
  items.sort((a, b) => a.date.localeCompare(b.date))
  createInvLineItems.value = items
  createInvStep.value = 1
  showCreateInvoiceFromGroup.value = true
}

async function saveInvoiceFromGroup() {
  const grp = createInvGroup.value
  if (!grp?.customer_id) return
  savingInvFromGrp.value = true
  try {
    await customersApi.createInvoice(grp.customer_id, {
      period_start: createInvForm.value.period_start,
      period_end:   createInvForm.value.period_end,
      due_date:     createInvForm.value.due_date || undefined,
      vat_rate:     createInvForm.value.vat_rate,
      notes:        createInvForm.value.notes,
      line_items:   createInvLineItems.value,
      currency:     createInvCurrency.value,
    })
    showCreateInvoiceFromGroup.value = false
    ui.success(t('invoice.invoices'))
    if (mode.value === 'invoices') await loadAllInvoices()
  } catch {
    ui.error('Failed to create invoice')
  } finally {
    savingInvFromGrp.value = false
  }
}
</script>

<style scoped>
/* ── Shell ── */
.tt-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--color-bg);
  font-size: 13px;
}

/* ── Top bar ── */
.tt-bar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 16px;
  padding: 0 16px;
  height: 48px;
  background: var(--color-primary);
  color: #fff;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.tt-emp-label { font-size: 11px; opacity: .7; margin-right: 6px; }
.tt-emp-name  { font-weight: 600; }
.tt-user-select {
  background: rgba(255,255,255,.15);
  border: 1px solid rgba(255,255,255,.3);
  border-radius: 4px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  padding: 3px 8px;
  cursor: pointer;
}
.tt-user-select option { background: var(--color-surface); color: var(--color-text); }

.tt-week-nav {
  display: flex;
  align-items: center;
  gap: 10px;
}
.wk-picker-wrap { position: relative; }
.wk-label {
  font-weight: 600; font-size: 14px; letter-spacing: .5px;
  background: none; border: none; color: #fff; cursor: pointer; padding: 2px 6px;
  border-radius: 4px; transition: background .12s;
}
.wk-label:hover { background: rgba(255,255,255,.2); }
.wk-cal {
  position: absolute; top: calc(100% + 6px); left: 50%; transform: translateX(-50%);
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-sm); box-shadow: 0 6px 20px rgba(0,0,0,.18);
  padding: 10px; z-index: 300; min-width: 220px;
}
.wk-cal-hdr {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 8px;
}
.wk-cal-month { font-size: 13px; font-weight: 600; color: var(--color-text); }
.wk-cal-nav {
  background: none; border: none; cursor: pointer; padding: 2px 6px;
  color: var(--color-text-muted); font-size: 10px; border-radius: 3px;
}
.wk-cal-nav:hover { background: var(--color-surface-2); }
.wk-cal-grid { display: grid; grid-template-columns: auto repeat(7, 1fr); gap: 2px; }
.wk-cal-dn { font-size: 10px; font-weight: 600; color: var(--color-text-muted); text-align: center; padding: 2px 0 4px; }
.wk-cal-wn {
  background: none; border: none; cursor: pointer; padding: 4px 5px 4px 2px;
  font-size: 10px; font-weight: 600; color: var(--color-text-muted);
  text-align: right; border-radius: 3px; line-height: 1.4; white-space: nowrap;
}
.wk-cal-wn:hover { background: var(--color-surface-2); color: var(--color-text); }
.wk-cal-wn.wk-cal-wn-sel { color: var(--color-primary); }
.wk-cal-day {
  background: none; border: none; cursor: pointer; padding: 4px 0;
  font-size: 12px; text-align: center; border-radius: 3px; color: var(--color-text);
  line-height: 1.4;
}
.wk-cal-day:hover { background: var(--color-surface-2); }
.wk-cal-day.wk-cal-other { color: var(--color-text-muted); opacity: .5; }
.wk-cal-day.wk-cal-sel { background: var(--color-primary); color: #fff; font-weight: 600; }
.wk-cal-day.wk-cal-sel:hover { background: var(--color-primary); }
.wk-cal-day.wk-cal-today:not(.wk-cal-sel) { font-weight: 700; color: var(--color-primary); }
.nav-btn {
  background: rgba(255,255,255,.2);
  border: none;
  color: #fff;
  width: 26px;
  height: 26px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}
.nav-btn:hover { background: rgba(255,255,255,.35); }
.nav-today { width: auto; padding: 0 8px; font-size: 11px; font-weight: 600; letter-spacing: .03em; }
.nav-today:disabled { opacity: .4; cursor: default; }
.nav-today:disabled:hover { background: rgba(255,255,255,.2); }
.nav-holidays-wrap { position: relative; margin-left: 4px; }
.nav-holidays { width: auto; padding: 0 10px; font-size: 11px; font-weight: 600; letter-spacing: .03em; display: inline-flex; align-items: center; gap: 4px; }
.nav-hol-chevron { display: inline-block; font-style: normal; font-size: 11px; transform: rotate(90deg); transition: transform .15s; line-height: 1; }
.nav-hol-chevron.open { transform: rotate(-90deg); }
.holidays-drop {
  position: absolute; top: calc(100% + 4px); right: 0;
  min-width: 180px; padding: 4px 0;
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-sm); box-shadow: 0 4px 12px rgba(0,0,0,.15); z-index: 200;
}
.hol-country-btn {
  display: flex; align-items: center; gap: 8px; width: 100%;
  padding: 7px 14px; border: none; background: none;
  font-size: 13px; color: var(--color-text); cursor: pointer; text-align: left;
}
.hol-country-btn:hover { background: var(--color-surface-2); }
.hol-flag {
  display: inline-block; width: 26px; height: 17px; border-radius: 2px; flex-shrink: 0;
  box-shadow: 0 0 0 1px rgba(0,0,0,.18);
}
.hol-flag-emoji { font-size: 18px; line-height: 1; flex-shrink: 0; }
/* Horizontal tricolors */
.hol-flag[data-country="NL"] { background: linear-gradient(to bottom, #AE1C28 0% 33.4%, #fff 33.4% 66.6%, #21468B 66.6% 100%); }
.hol-flag[data-country="DE"] { background: linear-gradient(to bottom, #222 0% 33.4%, #DD0000 33.4% 66.6%, #FFCE00 66.6% 100%); }
.hol-flag[data-country="ES"] { background: linear-gradient(to bottom, #AA151B 0% 25%, #F1BF00 25% 75%, #AA151B 75% 100%); }
/* Vertical tricolors */
.hol-flag[data-country="FR"] { background: linear-gradient(to right, #002395 0% 33.4%, #fff 33.4% 66.6%, #ED2939 66.6% 100%); }
.hol-flag[data-country="IT"] { background: linear-gradient(to right, #009246 0% 33.4%, #fff 33.4% 66.6%, #CE2B37 66.6% 100%); }
.hol-flag[data-country="PT"] { background: linear-gradient(to right, #006600 0% 38%, #CC0000 38% 100%); }
/* Nordic cross / other — approximated as bicolour split */
.hol-flag[data-country="GB"] { background: linear-gradient(135deg, #012169 0% 50%, #CF101A 50% 100%); }
.hol-flag[data-country="DK"] { background: linear-gradient(to right, #C60C30 0% 35%, #fff 35% 45%, #C60C30 45% 100%); }
.hol-flag[data-country="SE"] { background: linear-gradient(to right, #006AA7 0% 35%, #FECC00 35% 45%, #006AA7 45% 100%); }
.hol-flag[data-country="NO"] { background: linear-gradient(to right, #EF2B2D 0% 30%, #fff 30% 40%, #002868 40% 48%, #fff 48% 58%, #EF2B2D 58% 100%); }
.hol-flag[data-country="FI"] { background: linear-gradient(to right, #fff 0% 35%, #003580 35% 45%, #fff 45% 100%); }
.hol-flag[data-country="IS"] { background: linear-gradient(to right, #003897 0% 30%, #fff 30% 40%, #DC143C 40% 48%, #fff 48% 58%, #003897 58% 100%); }


/* Holiday highlights */
.c-day { position: relative; }
.cell-hol-toggle {
  position: absolute;
  bottom: 3px; right: 3px;
  width: 10px; height: 10px;
  padding: 0; border: none;
  border-radius: 50%;
  background: var(--color-text-muted, #888);
  opacity: 0;
  cursor: pointer;
  transition: opacity 0.15s, background 0.15s;
}
td.c-day:hover .cell-hol-toggle,
td.c-day:focus-within .cell-hol-toggle,
.cell-hol-toggle.cell-hol-on { opacity: 1; }
.cell-hol-toggle.cell-hol-on {
  background: var(--color-warning, #f59e0b);
  box-shadow: 0 0 5px var(--color-warning, #f59e0b);
}

/* Clock button — time range toggle */
.cell-time-toggle {
  position: absolute;
  bottom: 3px; left: 3px;
  width: 14px; height: 14px;
  padding: 0; border: none;
  border-radius: 3px;
  background: transparent;
  color: var(--color-text-muted, #888);
  font-size: 10px;
  line-height: 1;
  opacity: 0;
  cursor: pointer;
  transition: opacity 0.15s, color 0.15s;
  display: flex; align-items: center; justify-content: center;
}
td.c-day:hover .cell-time-toggle,
td.c-day:focus-within .cell-time-toggle,
.cell-time-toggle.cell-time-on { opacity: 1; }
.cell-time-toggle.cell-time-on { color: var(--color-primary); }
.cell-time-toggle:focus-visible { opacity: 1; }

/* Blue dot indicator for cells that have a time range set */
.cell-time-dot {
  position: absolute;
  top: 3px; left: 3px;
  width: 5px; height: 5px;
  border-radius: 50%;
  background: var(--color-primary);
  pointer-events: none;
}

/* Distance toggle button */
.cell-dist-toggle {
  position: absolute;
  bottom: 3px; left: 19px;
  width: 14px; height: 14px;
  padding: 0; border: none;
  border-radius: 3px;
  background: transparent;
  color: var(--color-text-muted, #888);
  font-size: 10px;
  line-height: 1;
  opacity: 0;
  cursor: pointer;
  transition: opacity 0.15s, color 0.15s;
  display: flex; align-items: center; justify-content: center;
}
td.c-day:hover .cell-dist-toggle,
td.c-day:focus-within .cell-dist-toggle,
.cell-dist-toggle.cell-dist-on { opacity: 1; }
.cell-dist-toggle.cell-dist-on { color: var(--color-success, #16a34a); }
.cell-dist-toggle:focus-visible { opacity: 1; }

/* Green dot indicator for cells that have distance set */
.cell-dist-dot {
  position: absolute;
  top: 3px; left: 11px;
  width: 5px; height: 5px;
  border-radius: 50%;
  background: var(--color-success, #16a34a);
  pointer-events: none;
}

/* Distance footer row */
.tt-foot-dist .foot-dist-lbl {
  font-weight: 500;
  color: var(--color-text-muted);
  font-size: 11px;
}
.tt-foot-dist td {
  font-size: 11px;
  color: var(--color-text-muted);
}

/* Distance total in report tab */
.rpt-distance-total {
  display: flex;
  justify-content: space-between;
  padding: 6px 14px;
  font-size: 13px;
  color: var(--color-text-muted);
  border-top: 1px solid var(--color-border);
  margin-top: 4px;
}

/* Time range popup */
.time-popup {
  position: absolute;
  bottom: calc(100% + 4px);
  top: auto;
  left: 0;
  z-index: 300;
  background: var(--color-surface);
  border: 2px solid var(--color-primary);
  border-radius: var(--radius-sm);
  box-shadow:
    0 4px 16px rgba(0, 0, 0, 0.18),
    0 0 0 1px color-mix(in srgb, var(--color-primary) 30%, transparent);
  padding: 10px;
  min-width: 150px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.time-popup.time-popup-below { bottom: auto; top: calc(100% + 4px); }
.tp-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 11px;
  color: var(--color-text-muted);
}
.tp-inp {
  width: 100%;
  padding: 3px 6px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  background: var(--color-input-bg, var(--color-surface-2));
  color: var(--color-text);
  font-size: 12px;
}
.tp-inp:focus {
  border-color: var(--color-primary);
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-primary) 28%, transparent);
}
.tp-preview {
  text-align: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary);
}
.tp-overnight {
  display: block;
  font-size: 10px;
  font-weight: 500;
  color: var(--color-text-muted);
  margin-top: 2px;
}
.tp-actions {
  display: flex;
  gap: 4px;
}
.tp-btn {
  flex: 1;
  padding: 3px 6px;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
}
.tp-apply {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.tp-apply:hover { opacity: .9; }
.tp-clear {
  background: var(--color-surface-2);
  color: var(--color-text-muted);
}
.tp-clear:hover { background: var(--color-surface-3, var(--color-surface-2)); }

.standby-hint,
.standby-preview {
  font-size: 12px;
  color: var(--color-text-muted);
  margin: 0 0 12px;
}
.standby-preview {
  margin: 12px 0 0;
  font-weight: 600;
  color: var(--color-primary);
}
.standby-preset { margin-top: 4px; }

.field-error {
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-danger, #ef4444);
}

.date-input-row { display: flex; align-items: center; gap: 6px; }
.date-input-row .form-input { flex: 1; }
.picker-wrap { position: relative; display: inline-flex; flex-shrink: 0; cursor: pointer; }
.date-picker-overlay { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }
.btn-icon-xs {
  background: none; border: none; cursor: pointer; color: var(--color-text-muted);
  padding: 2px 4px; font-size: 13px; line-height: 1; border-radius: 3px; flex-shrink: 0;
}
.btn-icon-xs:hover { background: var(--color-bg); color: var(--color-text); }
.macro-hint {
  margin: 0 0 12px;
  color: var(--color-text-muted);
  font-size: 13px;
}
.macro-toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}
.macro-toolbar-group {
  min-width: 260px;
}
.macro-toolbar-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.macro-grid {
  display: grid;
  gap: 6px;
}
.macro-head--single,
.macro-row--single {
  display: grid;
  grid-template-columns: 1.1fr 1.1fr 1.4fr 0.7fr 0.7fr 0.7fr 0.7fr auto;
  gap: 6px;
  align-items: center;
}
.macro-head--alt,
.macro-row--alt {
  display: grid;
  grid-template-columns: 1.1fr 1.1fr 1.4fr 0.7fr 0.7fr 0.7fr 0.7fr 0.7fr 0.7fr 0.7fr 0.7fr auto;
  gap: 6px;
  align-items: center;
}
.macro-head {
  font-size: 11px;
  font-weight: 700;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 0 2px;
}
.macro-row .form-input {
  min-width: 0;
}
.macro-actions {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.macro-apply-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 6px;
}
.macro-apply-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.macro-apply-toggle {
  flex-direction: row;
  align-items: center;
  gap: 8px;
  padding-bottom: 6px;
}
.macro-apply-toggle input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary);
  cursor: pointer;
  flex-shrink: 0;
}
.macro-apply-toggle label {
  margin: 0;
  font-size: 13px;
  cursor: pointer;
}
.macro-row-remove {
  white-space: nowrap;
}
.macro-days-input {
  width: 72px;
}
.macro-apply-preview {
  margin: 0 0 10px;
  color: var(--color-text-muted);
  font-size: 12px;
}
.c-day-holiday { box-shadow: inset 0 0 0 9999px rgba(0, 0, 0, 0.18); }
.tt-head th.c-day-today { background: color-mix(in srgb, var(--color-primary) 18%, var(--color-surface)); color: var(--color-primary); }
td.c-day-today { box-shadow: inset 0 0 0 9999px color-mix(in srgb, var(--color-primary) 10%, transparent); }
.dh-holiday-dot {
  width: 7px; height: 7px; border-radius: 50%;
  background: var(--color-warning, #f59e0b); margin: 3px auto 0;
  box-shadow: 0 0 4px var(--color-warning, #f59e0b);
}
.tt-row-holiday { background: color-mix(in srgb, var(--color-warning, #f59e0b) 22%, transparent) !important; }
.tt-row-holiday .c-info, .tt-row-holiday .c-desc { font-style: italic; opacity: .9; }
.c-day-holiday-cell { box-shadow: inset 0 0 0 9999px rgba(0, 0, 0, 0.30); }
.c-day-copied { outline: 2px dashed var(--color-primary); outline-offset: -2px; }
td.c-day-selected {
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--color-primary) 55%, transparent);
}
td.c-day-today.c-day-selected {
  box-shadow:
    inset 0 0 0 9999px color-mix(in srgb, var(--color-primary) 10%, transparent),
    inset 0 0 0 2px color-mix(in srgb, var(--color-primary) 55%, transparent);
}
td.c-day-holiday.c-day-selected {
  box-shadow:
    inset 0 0 0 9999px rgba(0, 0, 0, 0.18),
    inset 0 0 0 2px color-mix(in srgb, var(--color-primary) 55%, transparent);
}
td.c-day-holiday-cell.c-day-selected {
  box-shadow:
    inset 0 0 0 9999px rgba(0, 0, 0, 0.30),
    inset 0 0 0 2px color-mix(in srgb, var(--color-primary) 55%, transparent);
}
td.c-day-selected .h-inp { background: transparent; }

.tt-dragging { user-select: none; }
.tt-dragging .c-day button { pointer-events: none; }
.tt-dragging .h-inp { caret-color: transparent; }

td.c-day.c-day-popup-open {
  box-shadow: inset 0 0 0 2px var(--color-primary);
}
td.c-day-today.c-day-popup-open {
  box-shadow:
    inset 0 0 0 9999px color-mix(in srgb, var(--color-primary) 10%, transparent),
    inset 0 0 0 2px var(--color-primary);
}
td.c-day-holiday.c-day-popup-open,
td.c-day-holiday-cell.c-day-popup-open {
  box-shadow:
    inset 0 0 0 9999px rgba(0, 0, 0, 0.18),
    inset 0 0 0 2px var(--color-primary);
}

.tt-mode-tabs { display: flex; gap: 2px; margin-left: auto; }
.tt-mode-btn {
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 14px;
  border: 1px solid rgba(255,255,255,.4);
  border-radius: 4px;
  background: transparent;
  color: rgba(255,255,255,.8);
  cursor: pointer;
  font-size: 12px;
}
.tt-mode-btn.active {
  background: #fff;
  color: var(--color-primary);
  font-weight: 600;
  border-color: #fff;
}

/* ── Sheet wrapper ── */
.tt-sheet-outer {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}
.tt-scroll {
  flex: 1;
  overflow: auto;
}
.tt-loading {
  padding: 40px;
  text-align: center;
  color: var(--color-text-muted);
}

/* ── Table ── */
.tt-table {
  border-collapse: collapse;
  min-width: 100%;
  table-layout: fixed;
}

/* Column widths */
.c-nr   { width: 40px; }
.c-info { width: 210px; overflow: visible; }
.c-desc { width: 180px; }
.c-rate { width: 88px; }
.c-day  { width: 82px; }
.c-total { width: 70px; }
.c-act  { width: 72px; }

/* Sticky left columns */
.c-nr, .c-info, .c-desc {
  position: sticky;
  left: 0;
  z-index: 2;
  background: inherit;
}
.c-info { left: 40px; }
.c-desc { left: 250px; }

/* Head */
.tt-head th {
  background: var(--color-surface);
  border-bottom: 2px solid var(--color-border);
  border-right: 1px solid var(--color-border);
  padding: 6px 8px;
  text-align: center;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
}
.tt-head .c-nr   { text-align: center; }
.tt-head .c-info { text-align: left; display: table-cell; }
.tt-head .c-info .sub { display: block; font-weight: 400; font-size: 11px; color: var(--color-text-muted); }
.tt-head .c-desc { text-align: left; }

.sort-btn {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  font: inherit;
  font-weight: 600;
  color: inherit;
  text-align: left;
  display: flex;
  align-items: flex-start;
  gap: 4px;
  width: 100%;
}
.sort-btn:hover { color: var(--color-primary); }
.sort-label { flex: 1; }
.sort-icon {
  font-size: 13px;
  color: var(--color-text-muted);
  opacity: .35;
  flex-shrink: 0;
  margin-top: 1px;
}
.sort-btn.sort-active .sort-icon {
  color: var(--color-primary);
  opacity: 1;
}

.dh-abbr { font-weight: 600; }
.dh-date { font-size: 11px; color: var(--color-text-muted); }

/* Rows */
.tt-row td {
  border-bottom: 1px solid var(--color-border);
  border-right: 1px solid var(--color-border);
  padding: 3px 6px;
  vertical-align: middle;
  background: var(--color-surface);
}
.tt-row.alt td { background: var(--color-bg); }
.c-nr { text-align: center; color: var(--color-text-muted); font-size: 12px; }

.rc-cust { font-weight: 600; font-size: 12px; line-height: 1.3; color: var(--color-text); }
.rc-proj { font-size: 11px; color: var(--color-text-muted); line-height: 1.3; }
.rc-contract { font-size: 11px; color: var(--color-primary); line-height: 1.3; }
.rc-rate { font-size: 12px; color: var(--color-text-muted); white-space: nowrap; }
.rc-rate-slots { color: var(--color-primary); font-size: 10px; margin-left: 2px; cursor: default; }
.rc-comment-wrap { position: relative; display: inline-block; }
.rc-comment-btn {
  display: inline-flex; align-items: center; justify-content: center;
  margin-top: 3px; padding: 1px 3px; border: 1px solid var(--color-border);
  border-radius: 4px; background: transparent; cursor: pointer;
  color: var(--color-text-muted); line-height: 1;
  opacity: .5; transition: opacity .15s;
}
.rc-comment-btn:hover { opacity: 1; background: var(--color-bg); border-color: var(--color-primary); color: var(--color-primary); }
.rc-comment-btn.has-comment { opacity: 1; color: var(--color-primary); border-color: var(--color-primary); background: color-mix(in srgb, var(--color-primary) 8%, transparent); }
.rc-comment-popup {
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: 8px; padding: 8px; box-shadow: 0 4px 16px rgba(0,0,0,.15);
  min-width: 240px;
}
.rc-comment-textarea {
  width: 100%; resize: vertical; min-height: 40px;
  border: 1px solid var(--color-border); border-radius: 4px;
  padding: 4px 6px; font-size: 12px; line-height: 1.4;
  background: var(--color-bg); color: var(--color-text);
  font-family: inherit;
}
.rc-comment-textarea:focus { outline: none; border-color: var(--color-primary); }
.rc-comment-actions { display: flex; gap: 4px; margin-top: 4px; justify-content: flex-end; }
.rc-comment-actions .tp-btn { font-size: 11px; padding: 2px 8px; }
.c-desc  { font-size: 12px; color: var(--color-text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tt-field-editable { cursor: pointer; }

/* Hour input cells */
.c-day {
  --h-inp-pad-x: 6px;
  padding: 2px;
}
.h-inp {
  width: 100%;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  padding: 4px var(--h-inp-pad-x);
  text-align: right;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  background: var(--color-surface);
  color: var(--color-text);
  outline: none;
  box-sizing: border-box;
  -moz-appearance: textfield;
}
.h-inp::-webkit-outer-spin-button,
.h-inp::-webkit-inner-spin-button { -webkit-appearance: none; }
.h-inp:focus { border-color: var(--color-primary); background: var(--color-surface); box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-primary) 20%, transparent); }
.h-inp.h-inp-filled { background: var(--color-surface); font-weight: 600; }
.h-inp:disabled { background: var(--color-bg); cursor: not-allowed; }
.h-inp::placeholder { color: transparent; }
.h-inp:focus::placeholder { color: var(--color-text-muted); opacity: .5; }

/* Totals */
.c-total { text-align: right; font-weight: 600; font-size: 13px; padding: 4px 10px; color: var(--color-text); }
.c-rowtotal { color: var(--color-text); }

/* Undeclarable row-total badge */
.row-undecl-badge {
  display: block;
  font-size: 10px;
  font-weight: 500;
  color: var(--color-text-muted);
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}

/* Undeclarable amount shown below a day cell's time value */
.cell-undecl {
  font-size: 10px;
  font-weight: 500;
  color: var(--color-text-muted);
  text-align: right;
  line-height: 1.2;
  padding: 1px calc(var(--h-inp-pad-x) + 1px) 0 var(--h-inp-pad-x);
  font-variant-numeric: tabular-nums;
}

/* Undeclarable amount shown inline below footer day totals */
.foot-undecl-inline {
  display: block;
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-muted);
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}

/* Footer */
.tt-foot td {
  background: var(--color-surface);
  border-top: 2px solid var(--color-border);
  border-right: 1px solid var(--color-border);
  padding: 6px 8px;
  font-weight: 700;
  color: var(--color-text);
}
.foot-lbl { text-align: right; }
.c-dttotal { text-align: right; font-size: 13px; color: var(--color-text); }
.c-dttotal.c-day-over {
  color: var(--color-text);
  background: color-mix(in srgb, var(--color-warning) 14%, var(--color-surface));
}
.grand-total { color: var(--color-text); font-size: 14px; font-weight: 800; }


/* Add-row bar */
.tt-add-bar {
  border-top: 1px solid var(--color-border);
  padding: 8px 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  background: var(--color-surface);
  flex-shrink: 0;
}
.btn-add-row {
  background: none;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius);
  padding: 5px 14px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-primary);
}
.btn-add-row:hover { background: var(--color-bg); border-color: var(--color-primary); }
.btn-copy-prev {
  background: none;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius);
  padding: 5px 14px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text-muted);
}
.btn-copy-prev:hover:not(:disabled) { background: var(--color-bg); border-color: var(--color-text-muted); color: var(--color-text); }
.btn-copy-prev:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-copy-prev.macro-run-open {
  background: var(--color-bg);
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.macro-run-wrap { position: relative; }
.macro-run-drop {
  position: absolute;
  left: 0;
  bottom: calc(100% + 6px);
  width: min(320px, calc(100vw - 24px));
  padding: 12px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.18);
  z-index: 250;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.macro-run-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  color: var(--color-text);
}
.macro-run-preview {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 12px;
  line-height: 1.4;
}
.macro-run-btn { align-self: stretch; }
.btn-undo {
  font-size: 13px;
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text-muted);
  cursor: pointer;
}
.btn-undo:hover:not(:disabled) { background: var(--color-bg); border-color: var(--color-text-muted); color: var(--color-text); }
.btn-undo:disabled { opacity: 0.5; cursor: not-allowed; }
.tt-export-group { margin-left: auto; display: flex; gap: 6px; align-items: flex-end; }
.pdf-options-wrapper { position: relative; }
.pdf-options-btn {
  display: inline-flex; align-items: center; gap: 6px;
  height: 36px; padding: 0 12px;
  border: 1px solid var(--color-border); border-radius: var(--radius-sm);
  background: var(--color-surface); color: var(--color-text);
  font-size: 13px; cursor: pointer; white-space: nowrap;
  transition: border-color .15s, color .15s;
}
.pdf-options-btn:hover { border-color: var(--color-primary); }
.pdf-options-btn.is-active { border-color: var(--color-primary); color: var(--color-primary); }
.pdf-opts-chevron {
  display: inline-block; font-style: normal; font-size: 11px;
  transform: rotate(90deg); transition: transform .15s;
}
.pdf-opts-chevron.open { transform: rotate(-90deg); }
.pdf-options-panel {
  position: absolute; bottom: calc(100% + 4px); top: auto; right: 0;
  min-width: 280px; padding: 6px 0;
  background: var(--color-surface); border: 1px solid var(--color-border);
  border-radius: var(--radius-sm); box-shadow: 0 4px 12px rgba(0,0,0,.12);
  z-index: 120;
}
.pdf-options-panel.opens-down { bottom: auto; top: calc(100% + 4px); }
.pdf-option-selects {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: center;
  gap: 6px 10px;
  padding: 8px 14px;
}
.pdf-option-select-label {
  font-size: 11px; font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase; letter-spacing: 0.04em;
  white-space: nowrap;
}
.pdf-option-select { font-size: 13px; }
.pdf-options-divider { height: 1px; background: var(--color-border); margin: 2px 0; }
.pdf-option-item {
  display: flex; align-items: center; gap: 8px;
  padding: 7px 14px; font-size: 13px; color: var(--color-text);
  cursor: pointer; user-select: none;
}
.pdf-option-item:hover { background: var(--color-surface-2); }
.pdf-option-item input[type="checkbox"] { accent-color: var(--color-primary); cursor: pointer; flex-shrink: 0; }
.grid-pdf-panel { min-width: 230px; padding: 10px 0 8px; }
.grid-pdf-section-label {
  font-size: 11px; font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase; letter-spacing: 0.04em;
  padding: 0 14px 6px;
}
.grid-pdf-type-row { display: flex; padding: 0 14px 10px; }
.grid-pdf-type-btn {
  flex: 1; padding: 5px 4px;
  font-size: 13px; cursor: pointer; line-height: 1.3;
  border: 1px solid var(--color-border);
  background: var(--color-surface); color: var(--color-text);
  transition: background .12s, color .12s, border-color .12s;
}
.grid-pdf-type-btn:first-child { border-radius: var(--radius-sm) 0 0 var(--radius-sm); }
.grid-pdf-type-btn:last-child { border-radius: 0 var(--radius-sm) var(--radius-sm) 0; }
.grid-pdf-type-btn:not(:first-child) { margin-left: -1px; }
.grid-pdf-type-btn.active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); z-index: 1; position: relative; }
.grid-pdf-type-btn:hover:not(.active) { background: var(--color-surface-2); border-color: var(--color-primary); z-index: 1; position: relative; }
.grid-pdf-date-row { display: flex; gap: 8px; padding: 0 14px 8px; }
.grid-pdf-select { flex: 1; min-width: 0; }
.grid-pdf-select-full { width: 100%; }
.grid-pdf-actions { display: flex; justify-content: flex-end; padding: 6px 14px 2px; }
.grid-pdf-export-btn { font-size: 13px; height: 30px; padding: 0 14px; }
.filter-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

/* Action column */
.c-act {
  padding: 0 4px;
  text-align: center;
  white-space: nowrap;
}
.tt-head .c-act { background: var(--color-surface); border-bottom: 2px solid var(--color-border); border-right: 1px solid var(--color-border); }
.tt-row .c-act { background: inherit; }
.tt-foot .c-act { background: var(--color-surface); border-top: 2px solid var(--color-border); }

.act-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
  background: transparent;
  color: var(--color-text-muted);
  opacity: 0;
  transition: opacity .15s, background .1s;
}
.tt-row:hover .act-btn { opacity: 1; }
.act-btn:focus-visible { opacity: 1; }
.act-btn:hover { background: var(--color-bg); color: var(--color-text); }

.act-edit:hover { color: var(--color-primary); }
.act-duplicate:hover { color: var(--color-primary); }
.act-standby:hover { color: var(--color-primary); }
.act-slots:hover  { color: var(--color-primary); }
.act-del:hover  { color: var(--color-danger); }
.act-ok  { opacity: 1 !important; color: var(--color-success); }
.act-ok:hover { background: color-mix(in srgb, var(--color-success) 12%, transparent); }
.act-no  { opacity: 1 !important; color: var(--color-danger); }
.act-no:hover { background: color-mix(in srgb, var(--color-danger) 12%, transparent); }

/* Modal action buttons are always visible (no hover-parent needed) */
.ttp-act { opacity: 1 !important; }

/* Deleting row highlight */
.tt-row-deleting td { background: color-mix(in srgb, var(--color-danger) 8%, var(--color-surface)) !important; }

/* Editing row highlight */
.tt-editing { background: color-mix(in srgb, var(--color-warning, #f59e0b) 8%, var(--color-surface)) !important; }

/* New row editor */
.tt-newrow td { background: color-mix(in srgb, var(--color-warning, #f59e0b) 8%, var(--color-surface)) !important; }
.nr-sel {
  width: 100%;
  font-size: 12px;
  padding: 3px 4px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  display: block;
  margin-bottom: 2px;
  background: var(--color-surface);
  color: var(--color-text);
}
.nr-desc {
  width: 100%;
  font-size: 12px;
  padding: 3px 6px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  box-sizing: border-box;
  background: var(--color-surface);
  color: var(--color-text);
}
/* ── Report ── */
.tt-report-outer { flex: 1; overflow: auto; padding: 20px 24px; }
.report-filters { display: flex; align-items: flex-end; gap: 8px; flex-wrap: wrap; margin-bottom: 20px; }
.fi-sm { min-width: 80px; width: auto; }
.fi-year { width: 80px; }
.rpt-filter-group { display: flex; flex-direction: column; gap: 4px; }

.rpt-view-toggle { display: flex; border: 1px solid var(--color-border); border-radius: var(--radius-sm); overflow: hidden; }
.rpt-view-btn {
  border: none;
  background: var(--color-surface);
  color: var(--color-text-muted);
  padding: 0 12px;
  height: 36px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: background .15s, color .15s;
}
.rpt-view-btn + .rpt-view-btn { border-left: 1px solid var(--color-border); }
.rpt-view-btn:hover { background: var(--color-bg); }
.rpt-view-btn.active { background: var(--color-primary); color: #fff; }
.rpt-view-btn:focus-visible { outline: 2px solid var(--color-primary); outline-offset: -2px; }

.rpt-chart-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 20px;
  margin-bottom: 16px;
}
.rpt-chart-title { font-size: 14px; font-weight: 600; margin: 0 0 16px; color: var(--color-text); }
.rpt-chart-empty { padding: 48px; text-align: center; color: var(--color-text-muted); font-size: 14px; }
/* Opt Chart.js canvases out of the app-wide zoom: Chart.js's own size/DPR
   measurement and coordinate math assume "CSS pixel width" and "rendered
   width" always match, which CSS `zoom` breaks (confirmed in Firefox) and
   causes hover positions to drift off the actual bars. Canceling the
   inherited zoom here keeps Chart.js in the self-consistent, unzoomed
   environment it already handles correctly. */
.rpt-chart-wrap { max-width: 900px; margin: 0 auto; zoom: calc(1 / var(--app-zoom, 1)); }
.rpt-chart-wrap-pie { max-width: 640px; }

/* Report header: logo + company name + period */
.rpt-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 3px solid var(--color-primary);
}
.rpt-header-left { flex: 0 0 auto; min-width: 60px; }
.rpt-header-center { flex: 1; text-align: center; }
.rpt-header-right { flex: 0 0 auto; min-width: 60px; }
.rpt-logo { max-height: 52px; max-width: 160px; object-fit: contain; }
.rpt-company-name { font-size: 20px; font-weight: 800; color: var(--color-text); letter-spacing: -0.02em; }
.rpt-title-label { font-size: 12px; font-weight: 600; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.08em; margin-top: 4px; }
.rpt-period-label { font-size: 18px; font-weight: 700; color: var(--color-primary); margin-top: 4px; }

.rpt-group {
  margin-bottom: 16px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}
.rpt-group-hd {
  display: flex;
  justify-content: space-between;
  padding: 8px 14px;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  font-weight: 600;
  font-size: 13px;
}
.rpt-grp-total { color: var(--color-primary); }
.rpt-grp-empty { padding: 10px 14px; color: var(--color-text-muted); font-size: 12px; }
.rpt-grp-undecl-line {
  display: flex;
  justify-content: space-between;
  padding: 5px 14px;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-muted);
  background: var(--color-bg);
  border-top: 1px dashed var(--color-border);
}

.rpt-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  table-layout: fixed;
}
.rpt-col-date     { width: 110px; }
.rpt-col-customer { width: 18%; }
.rpt-col-project  { width: 18%; }
.rpt-col-activity { /* fills remaining space */ }
.rpt-col-distance { width: 90px; }
.rpt-col-undeclarable { width: 90px; }
.rpt-col-time     { width: 80px; }
.rpt-table th,
.rpt-table td {
  padding: 6px 14px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rpt-table th { color: var(--color-text-muted); font-weight: 500; background: var(--color-surface); }
.rpt-th-time { text-align: right; }

.rpt-grand-total {
  display: flex;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--color-primary);
  color: #fff;
  font-weight: 700;
  border-radius: var(--radius);
  margin-top: 8px;
  font-size: 14px;
}
.rpt-undeclarable {
  display: flex;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--color-bg);
  color: var(--color-text-muted);
  font-weight: 500;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  margin-top: 4px;
  font-size: 13px;
}
.rpt-declarable {
  background: color-mix(in srgb, var(--color-success, #38a169) 80%, #1a5c38);
}
.rpt-empty { color: var(--color-text-muted); padding: 20px 0; text-align: center; }

/* ── Modal tabs ── */
.ttp-tabs {
  display: flex;
  border-bottom: 2px solid var(--color-border);
  flex-shrink: 0;
}
.ttp-tab {
  flex: 1;
  padding: 8px 16px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-muted);
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: color .15s;
}
.ttp-tab.active {
  color: var(--color-primary);
  font-weight: 700;
  border-bottom-color: var(--color-primary);
}
.ttp-tab:hover:not(.active) { color: var(--color-text); }

/* ── Manage-projects button ── */
.tt-macro-edit-btn,
.tt-manage-btn {
  padding: 0 10px;
  font-size: 14px;
}

/* ── Modal backdrop & dialog ── */
.tt-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.tt-modal {
  background: var(--color-surface);
  border-radius: var(--radius);
  box-shadow: 0 8px 32px rgba(0,0,0,.25);
  width: 480px;
  max-width: 96vw;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.tt-modal-hd {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px 10px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}
.tt-modal-title {
  font-weight: 700;
  font-size: 15px;
  margin: 0;
}
.tt-modal-close {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 2px 6px;
  border-radius: 3px;
}
.tt-modal-close:hover { background: var(--color-bg); color: var(--color-text); }
.tt-modal-sub {
  padding: 8px 16px 4px;
  font-size: 12px;
  color: var(--color-text-muted);
  flex-shrink: 0;
}

/* ── TT project list ── */
.ttp-list {
  list-style: none;
  padding: 0;
  margin: 0;
  overflow-y: auto;
  flex: 1;
}
.ttp-empty {
  padding: 20px 16px;
  color: var(--color-text-muted);
  font-size: 13px;
  text-align: center;
}
.ttp-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--color-border);
  font-size: 13px;
}
.ttp-item:last-child { border-bottom: none; }
.ttp-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--color-border);
  flex-shrink: 0;
}
.ttp-name { flex: 1; font-weight: 500; }
.ttp-badge {
  font-size: 10px;
  font-weight: 600;
  color: var(--color-text-muted);
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 1px 7px;
  white-space: nowrap;
}
.ttp-name-input {
  flex: 1;
  font-size: 13px;
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  background: var(--color-bg);
  color: var(--color-text);
}
.ttp-name-input:focus { outline: none; border-color: var(--color-primary); }
.ttp-undecl-input {
  width: 62px;
  font-size: 12px;
  padding: 4px 6px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  background: var(--color-bg);
  color: var(--color-text);
  text-align: right;
}
.ttp-undecl-input:focus { outline: none; border-color: var(--color-primary); }
.ttp-undecl-badge {
  font-size: 11px;
  font-weight: 500;
  color: var(--color-text-muted);
  white-space: nowrap;
}
.ttp-color-input {
  width: 32px;
  height: 28px;
  padding: 2px;
  border: 1px solid var(--color-border);
  border-radius: 3px;
  cursor: pointer;
  background: none;
}

/* ── Add form ── */
.ttp-add-form {
  padding: 10px 16px;
  border-top: 1px solid var(--color-border);
  flex-shrink: 0;
}
.ttp-add-active {
  display: flex;
  align-items: center;
  gap: 8px;
}
.btn-sm { height: 28px; padding: 0 12px; font-size: 12px; }

/* ── Drag handle ── */
.c-drag { width: 28px; min-width: 28px; padding: 0 !important; text-align: center; }
.drag-handle {
  display: inline-flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; margin: 0; padding: 0;
  border: none; border-radius: 3px;
  background: none; color: var(--color-text-muted);
  cursor: grab; font-size: 13px; line-height: 1;
  user-select: none; -webkit-user-select: none;
}
.drag-handle:hover { background: var(--color-bg); color: var(--color-text); }
.drag-handle:active { cursor: grabbing; }
.tt-row.sortable-ghost { opacity: 0.3; }
.sortable-fallback {
  opacity: 0.9 !important;
  box-shadow: 0 4px 12px rgba(0,0,0,.2) !important;
}
/* ── Board Report ── */
.tt-board-rpt-outer { flex: 1; overflow: auto; }

/* ── Invoices tab ── */
.tt-invoices-outer {
  flex: 1;
  overflow: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.inv-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.fi-sm { height: 28px; padding: 0 8px; font-size: 12px; min-width: 140px; }
.inv-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.inv-table th, .inv-table td {
  padding: 6px 10px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  white-space: nowrap;
}
.inv-table th { font-weight: 600; color: var(--color-text-muted); font-size: 11px; }
.inv-col-num { text-align: right !important; }
.inv-total-cell { font-weight: 600; }
.inv-num-cell { font-family: monospace; }
.inv-muted { color: var(--color-text-muted); }
.inv-actions-cell { display: flex; gap: 4px; align-items: center; }
.inv-status-badge {
  display: inline-block;
  padding: 2px 7px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .03em;
}
.inv-draft { background: var(--color-bg-alt, #f0f0f0); color: var(--color-text-muted); }
.inv-sent  { background: #dbeafe; color: #1d4ed8; }
.inv-paid  { background: #d1fae5; color: #065f46; }
.inv-row-draft td { opacity: .85; }
.inv-row-paid  td { opacity: .9; }

/* ── Create Invoice modal ── */
.tt-modal-lg { max-width: 700px; width: 95vw; }
.tt-modal-body { padding: 16px 20px; overflow-y: auto; max-height: 60vh; }
.tt-modal-ft { display: flex; justify-content: flex-end; gap: 8px; padding: 12px 20px; border-top: 1px solid var(--color-border); }
.cinv-field-row { display: flex; gap: 16px; margin-bottom: 12px; }
.cinv-field { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.cinv-field label { font-size: 12px; font-weight: 500; color: var(--color-text-muted); }
.cinv-field .form-input { height: 30px; font-size: 13px; }

/* ── Report invoice button ── */
.rpt-invoice-btn { margin-left: 8px; font-size: 11px; padding: 2px 8px; }

/* ── Invoice preview table (reuse from CustomerDetailView) ── */
.inv-preview-table-wrap { overflow-x: auto; }
.inv-preview-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  margin-bottom: 12px;
}
.inv-preview-table th,
.inv-preview-table td {
  padding: 5px 8px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  white-space: nowrap;
}
.inv-preview-table th { font-weight: 600; color: var(--color-text-muted); }
.inv-preview-table td.num,
.inv-preview-table th.num { text-align: right; }
.inv-preview-totals {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  font-size: 13px;
}
.inv-preview-totals > div {
  display: flex;
  gap: 16px;
  justify-content: flex-end;
  min-width: 260px;
}
.inv-preview-totals span { color: var(--color-text-muted); }
.inv-total-row strong { color: var(--color-primary); font-size: 15px; }
</style>
