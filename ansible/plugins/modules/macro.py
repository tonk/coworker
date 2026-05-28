# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: macro
short_description: Manage WarmDesk macros
version_added: "0.4.2"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete WarmDesk macros via the admin API.
  - Macros are sequences of actions (set status, set priority, set type,
    add tag, add message) that can be applied to helpdesk tickets.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires admin credentials (C(global_role=admin)) or a token obtained from
    an admin account.
  - The macro name is used as the idempotency key.
  - Deleting a macro that does not exist is a no-op (no error).
  - Action comparison is order-sensitive — reordering actions counts as a change.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  name:
    description:
      - Display name for the macro.  Used as the idempotency key.
    type: str
    required: true

  description:
    description:
      - Optional description of what the macro does.
    type: str

  actions:
    description:
      - List of actions to execute when the macro is applied.
      - Each action is a dict with C(type) and C(value) keys.
      - "Supported types: C(set_status), C(set_priority), C(set_type), C(add_tag), C(add_message)."
      - "C(set_status) values: C(new), C(open), C(pending), C(pending_close), C(closed)."
      - "C(set_priority) values: C(low), C(medium), C(high), C(critical)."
      - "C(set_type) values: C(incident), C(problem), C(service_request), C(change_request)."
      - "C(add_tag) value: any tag string."
      - "C(add_message) value: message body, supports placeholders C({email}), C({fname}), C({name}), C({subject}), C({ticket_id}), C({agent}), C({agent_fname})."
    type: list
    elements: dict
    suboptions:
      type:
        description: Action type.
        type: str
        required: true
        choices:
          - set_status
          - set_priority
          - set_type
          - add_tag
          - add_message
      value:
        description: Action value.
        type: str
        required: true

  is_active:
    description:
      - Whether this macro is active and visible to helpdesk users.
    type: bool
    default: true

  sort_order:
    description:
      - Numeric sort order; lower values appear first in the macro list.
    type: int
    default: 0

  state:
    description:
      - C(present) ensures the macro exists with the specified settings.
      - C(absent) removes the macro if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create a macro that closes a ticket and adds a message
  ansilabnl.warmdesk.macro:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Close and thank customer
    description: Sets status to closed and sends a closing message
    actions:
      - type: set_status
        value: closed
      - type: add_message
        value: "Hi {fname}, your ticket {ticket_id} has been resolved. Thanks!"
    is_active: true
    state: present

- name: Create a critical triage macro
  ansilabnl.warmdesk.macro:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    name: Escalate to critical
    actions:
      - type: set_priority
        value: critical
      - type: set_type
        value: incident
      - type: add_tag
        value: escalated
    sort_order: 10

- name: Disable a macro without deleting it
  ansilabnl.warmdesk.macro:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    name: Old Macro
    is_active: false

- name: Delete a macro
  ansilabnl.warmdesk.macro:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Old Macro
    state: absent
"""

RETURN = r"""
macro:
  description:
    - The macro object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the macro was deleted (or did not exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric macro ID.
      returned: always
      type: int
      sample: 5
    name:
      description: Display name.
      returned: always
      type: str
      sample: Close and thank customer
    description:
      description: Optional description.
      returned: always
      type: str
      sample: Sets status to closed and sends a closing message
    actions:
      description: List of macro actions.
      returned: always
      type: list
      elements: dict
      contains:
        type:
          description: Action type.
          type: str
          sample: set_status
        value:
          description: Action value.
          type: str
          sample: closed
    is_active:
      description: Whether the macro is active.
      returned: always
      type: bool
      sample: true
    sort_order:
      description: Sort order position.
      returned: always
      type: int
      sample: 0
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2026-05-28T10:00:00Z"
    updated_at:
      description: ISO-8601 last-update timestamp.
      returned: always
      type: str
      sample: "2026-05-28T10:00:00Z"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


def _find_macro(client, name):
    macros = client.get('/admin/macros')
    for m in macros:
        if m.get('name') == name:
            return m
    return None


def _actions_equal(desired, existing):
    """Return True when both action lists are identical (order-sensitive)."""
    if desired is None:
        return True
    existing_actions = existing.get('actions') or []
    if len(desired) != len(existing_actions):
        return False
    for d, e in zip(desired, existing_actions):
        if d.get('type') != e.get('type') or d.get('value') != e.get('value'):
            return False
    return True


def _build_update(p, existing):
    """Return (body, changed) for a PUT request."""
    body = {}
    changed = False

    if p.get('description') is not None:
        if existing.get('description') != p['description']:
            body['description'] = p['description']
            changed = True

    if p.get('actions') is not None:
        if not _actions_equal(p['actions'], existing):
            body['actions'] = p['actions']
            changed = True

    if p.get('is_active') is not None:
        if existing.get('is_active') != p['is_active']:
            body['is_active'] = p['is_active']
            changed = True

    if p.get('sort_order') is not None:
        if existing.get('sort_order') != p['sort_order']:
            body['sort_order'] = p['sort_order']
            changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        name=dict(type='str', required=True),
        description=dict(type='str'),
        actions=dict(
            type='list',
            elements='dict',
            options=dict(
                type=dict(
                    type='str',
                    required=True,
                    choices=['set_status', 'set_priority', 'set_type', 'add_tag', 'add_message'],
                ),
                value=dict(type='str', required=True),
            ),
        ),
        is_active=dict(type='bool'),
        sort_order=dict(type='int'),
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
        existing = _find_macro(client, p['name'])

        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, macro=None)
            if not module.check_mode:
                client.delete('/admin/macros/%d' % existing['id'])
            module.exit_json(changed=True, macro=None)

        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, macro=None)
            body = dict(name=p['name'])
            if p.get('description') is not None:
                body['description'] = p['description']
            if p.get('actions') is not None:
                body['actions'] = p['actions']
            body['is_active'] = p['is_active'] if p.get('is_active') is not None else True
            if p.get('sort_order') is not None:
                body['sort_order'] = p['sort_order']
            macro = client.post('/admin/macros', body)
            module.exit_json(changed=True, macro=macro)

        update_body, changed = _build_update(p, existing)

        if not changed:
            module.exit_json(changed=False, macro=existing)

        if module.check_mode:
            module.exit_json(changed=True, macro=existing)

        macro = client.put('/admin/macros/%d' % existing['id'], update_body)
        module.exit_json(changed=True, macro=macro)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
