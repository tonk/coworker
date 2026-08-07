# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: timetrack_macro_library
short_description: Manage macros in a user's WarmDesk time-tracking macro library
version_added: "0.6.0"
author: "Ton Kersten (@tonk)"
description:
  - Add, update, or remove a single named macro from the current user's
    time-tracking macro library via C(GET/PUT /api/v1/time-entries/macro-library).
  - The module fetches the library, makes the required change, and writes
    the updated library back in a single round-trip.
  - Idempotent — a second run with identical parameters produces no change.
  - Supports check mode.
notes:
  - Requires that C(time_tracking_enabled) is set for the authenticated user's
    account.  The server returns 403 otherwise.
  - The macro I(name) is the idempotency key.  Renaming a macro is not
    supported; use C(state=absent) then C(state=present) with the new name.
  - C(rows) are matched by position.  Supplying fewer rows than the existing
    macro silently truncates; supplying more rows appends them.
  - The API rejects an empty macro library outright, so C(state=absent)
    fails with a clear message when asked to remove the last remaining
    macro — add a replacement first, or leave at least one macro in place.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  name:
    description:
      - Display name of the macro.  Used as the idempotency key.
    type: str
    required: true

  apply_days:
    description:
      - Number of days the macro is applied over (1–7).
    type: int
    default: 5

  alternating:
    description:
      - When C(true) the macro uses alternating day patterns (day1/day2).
    type: bool
    default: false

  rows:
    description:
      - List of time-entry rows for this macro.
      - Each row is a dict of time-entry fields.
      - When omitted on C(state=present) a default set of four rows is used.
    type: list
    elements: dict
    suboptions:
      customer_id:
        description: Customer ID for this row (null for none).
        type: int
      project_id:
        description: Project ID for this row (null for none).
        type: int
      description:
        description: Free-text description for the time entry.
        type: str
        default: ""
      day1_minutes:
        description: Minutes for day-1 entries (empty string or number).
        type: raw
      day1_start:
        description: Start time for day-1 entries (HH:MM or empty string).
        type: str
      day1_end:
        description: End time for day-1 entries (HH:MM or empty string).
        type: str
      day2_minutes:
        description: Minutes for day-2 entries when alternating is enabled.
        type: raw
      day2_start:
        description: Start time for day-2 entries.
        type: str
      day2_end:
        description: End time for day-2 entries.
        type: str
      day1_distance:
        description: Distance for day-1 travel entries.
        type: raw
      day2_distance:
        description: Distance for day-2 travel entries.
        type: raw

  state:
    description:
      - C(present) ensures the macro exists with the specified settings.
      - C(absent) removes the macro if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Add a teaching-block macro
  ansilabnl.warmdesk.timetrack_macro_library:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Teaching block
    apply_days: 5
    rows:
      - description: Travel to location
        day1_minutes: 60
        day2_minutes: 60
      - description: Preparing for teaching
        day1_minutes: 30
        day2_minutes: 30
      - description: Teaching
        day1_minutes: 360
        day2_minutes: 360
      - description: Travel home
        day1_minutes: 60
        day2_minutes: 60
    state: present

- name: Enable alternating days on a macro
  ansilabnl.warmdesk.timetrack_macro_library:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    name: Teaching block
    alternating: true

- name: Remove a macro from the library
  ansilabnl.warmdesk.timetrack_macro_library:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Old Macro
    state: absent
"""

RETURN = r"""
macro:
  description:
    - The macro dict as it exists in the library after the operation.
    - C(null) when C(state=absent) and the macro was removed (or did not exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric macro ID within the library.
      type: int
      sample: 1
    name:
      description: Display name.
      type: str
      sample: Teaching block
    apply_days:
      description: Number of days the macro covers.
      type: int
      sample: 5
    alternating:
      description: Whether the macro uses alternating day patterns.
      type: bool
      sample: false
    rows:
      description: List of time-entry row dicts.
      type: list
      elements: dict

library:
  description: The full library dict after the operation.
  returned: always
  type: dict
  contains:
    nextId:
      description: Next auto-increment ID for new macros.
      type: int
      sample: 3
    macros:
      description: All macros in the library.
      type: list
      elements: dict
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)

_DEFAULT_ROWS = [
    {'customer_id': None, 'project_id': None, 'description': 'Travel to location',
     'day1_minutes': '', 'day1_start': '', 'day1_end': '',
     'day2_minutes': '', 'day2_start': '', 'day2_end': '',
     'day1_distance': '', 'day2_distance': ''},
    {'customer_id': None, 'project_id': None, 'description': 'Preparing for teaching',
     'day1_minutes': '', 'day1_start': '', 'day1_end': '',
     'day2_minutes': '', 'day2_start': '', 'day2_end': '',
     'day1_distance': '', 'day2_distance': ''},
    {'customer_id': None, 'project_id': None, 'description': 'Teaching',
     'day1_minutes': '', 'day1_start': '', 'day1_end': '',
     'day2_minutes': '', 'day2_start': '', 'day2_end': '',
     'day1_distance': '', 'day2_distance': ''},
    {'customer_id': None, 'project_id': None, 'description': 'Travel home',
     'day1_minutes': '', 'day1_start': '', 'day1_end': '',
     'day2_minutes': '', 'day2_start': '', 'day2_end': '',
     'day1_distance': '', 'day2_distance': ''},
]

_ROW_FIELDS = [
    'customer_id', 'project_id', 'description',
    'day1_minutes', 'day1_start', 'day1_end',
    'day2_minutes', 'day2_start', 'day2_end',
    'day1_distance', 'day2_distance',
]


def _normalize_row(raw, fallback_description=''):
    """Normalise a raw row dict from module params or the server."""
    base = {
        'customer_id': None,
        'project_id': None,
        'description': fallback_description,
        'day1_minutes': '',
        'day1_start': '',
        'day1_end': '',
        'day2_minutes': '',
        'day2_start': '',
        'day2_end': '',
        'day1_distance': '',
        'day2_distance': '',
    }
    base.update({k: v for k, v in raw.items() if k in _ROW_FIELDS})
    # customer_id / project_id: keep None or convert to int
    for id_field in ('customer_id', 'project_id'):
        val = base[id_field]
        if val is None or val == '':
            base[id_field] = None
        else:
            try:
                base[id_field] = int(val)
            except (TypeError, ValueError):
                base[id_field] = None
    # Numeric-ish fields: keep as string (frontend convention)
    for str_field in ('day1_minutes', 'day2_minutes', 'day1_distance', 'day2_distance'):
        if base[str_field] is None:
            base[str_field] = ''
        else:
            base[str_field] = str(base[str_field])
    return base


def _rows_equal(rows_a, rows_b):
    """Return True when two normalised row lists are identical."""
    if len(rows_a) != len(rows_b):
        return False
    for a, b in zip(rows_a, rows_b):
        if a != b:
            return False
    return True


def _fetch_library(client):
    """Return the library dict, creating a default one if none exists."""
    result = client.get('/time-entries/macro-library')
    raw = result.get('library') if result else None
    if raw and isinstance(raw, dict) and isinstance(raw.get('macros'), list):
        return raw
    return {'nextId': 1, 'macros': []}


def _find_macro(library, name):
    for m in library.get('macros', []):
        if m.get('name') == name:
            return m
    return None


def _macro_matches(existing, desired_name, desired_apply_days, desired_alternating, desired_rows):
    """Return True when the existing macro already has the desired state."""
    if existing.get('apply_days') != desired_apply_days:
        return False
    if bool(existing.get('alternating')) != desired_alternating:
        return False
    existing_rows = [_normalize_row(r) for r in (existing.get('rows') or [])]
    if not _rows_equal(existing_rows, desired_rows):
        return False
    return True


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        name=dict(type='str', required=True),
        apply_days=dict(type='int', default=5),
        alternating=dict(type='bool', default=False),
        rows=dict(
            type='list',
            elements='dict',
            options=dict(
                customer_id=dict(type='int'),
                project_id=dict(type='int'),
                description=dict(type='str', default=''),
                day1_minutes=dict(type='raw'),
                day1_start=dict(type='str', default=''),
                day1_end=dict(type='str', default=''),
                day2_minutes=dict(type='raw'),
                day2_start=dict(type='str', default=''),
                day2_end=dict(type='str', default=''),
                day1_distance=dict(type='raw'),
                day2_distance=dict(type='raw'),
            ),
        ),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']
    client = WarmDeskClient.from_module(module)

    try:
        library = _fetch_library(client)

        if state == 'absent':
            existing = _find_macro(library, p['name'])
            if existing is None:
                module.exit_json(changed=False, macro=None, library=library)
            remaining = [m for m in library['macros'] if m.get('name') != p['name']]
            if not remaining:
                # UpdateTimeMacroLibrary rejects an empty macros array outright
                # (HTTP 400 "invalid macro library") — surface that up front
                # with an actionable message instead of a raw API error.
                module.fail_json(
                    msg='Cannot remove macro "%s": it is the last macro in the '
                        'library, and the API does not allow an empty macro '
                        'library. Add a replacement macro first, or leave at '
                        'least one macro in place.' % p['name']
                )
            if not module.check_mode:
                library['macros'] = remaining
                client.put('/time-entries/macro-library', {'library': library})
            module.exit_json(changed=True, macro=None, library=library)

        # --- state == 'present' ---
        apply_days = max(1, min(7, p['apply_days']))
        alternating = bool(p['alternating'])

        if p.get('rows') is not None:
            desired_rows = [_normalize_row(r) for r in p['rows']]
        else:
            desired_rows = [_normalize_row(r) for r in _DEFAULT_ROWS]

        existing = _find_macro(library, p['name'])

        if existing is not None:
            if _macro_matches(existing, p['name'], apply_days, alternating, desired_rows):
                module.exit_json(changed=False, macro=existing, library=library)
            if module.check_mode:
                module.exit_json(changed=True, macro=existing, library=library)
            existing['apply_days'] = apply_days
            existing['alternating'] = alternating
            existing['rows'] = desired_rows
            client.put('/time-entries/macro-library', {'library': library})
            module.exit_json(changed=True, macro=existing, library=library)

        # New macro
        if module.check_mode:
            module.exit_json(changed=True, macro=None, library=library)

        next_id = library.get('nextId') or 1
        if not isinstance(next_id, int):
            try:
                next_id = int(next_id)
            except (TypeError, ValueError):
                next_id = 1

        new_macro = {
            'id': next_id,
            'name': p['name'],
            'apply_days': apply_days,
            'alternating': alternating,
            'rows': desired_rows,
        }
        library['macros'].append(new_macro)
        library['nextId'] = next_id + 1
        client.put('/time-entries/macro-library', {'library': library})
        module.exit_json(changed=True, macro=new_macro, library=library)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
