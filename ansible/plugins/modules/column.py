# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: column
short_description: Manage Kanban columns in a WarmDesk project
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Creates, updates, or deletes a Kanban column inside a WarmDesk project.
  - Idempotent — a column is identified by I(project) slug + I(name). Running the
    module again with the same parameters produces no change.
  - Requires at least I(admin) role on the project (or global admin).
options:
  project:
    description:
      - Slug of the WarmDesk project that owns this column (e.g. C(my-project)).
    type: str
    required: true
  name:
    description:
      - Display name of the column (e.g. C(In Progress)).
      - Used as the idempotency key together with I(project).
    type: str
    required: true
  color:
    description:
      - Hex colour code for the column header (e.g. C(#4CAF50)).
      - Optional; leave unset to keep the server default (empty string).
    type: str
    required: false
  wip_limit:
    description:
      - Work-in-progress limit for the column. C(0) means no limit.
      - Optional; leave unset to keep the existing value.
    type: int
    required: false
  state:
    description:
      - C(present) — ensure the column exists and its properties match.
      - C(absent) — ensure the column does not exist. The server rejects
        deletion when the column still contains cards; move them first.
    type: str
    choices: [present, absent]
    default: present
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
notes:
  - Check mode is fully supported; no changes are made to the server.
  - Clearing the WIP limit on an C(existing) column (C(wip_limit: 0)) is sent
    as C(wip_limit_clear=true) rather than C(wip_limit=0) — the update
    endpoint rejects C(wip_limit) values below 1 outright, unlike creation,
    which normalises C(0) to "no limit" automatically.
seealso:
  - module: ansilabnl.warmdesk.card
  - module: ansilabnl.warmdesk.label
'''

EXAMPLES = r'''
- name: Ensure "In Progress" column exists (no WIP limit)
  ansilabnl.warmdesk.column:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: In Progress
    color: "#1E88E5"

- name: Set a WIP limit of 3 on the "In Review" column
  ansilabnl.warmdesk.column:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: In Review
    color: "#FB8C00"
    wip_limit: 3

- name: Remove the "Archive" column (must be empty)
  ansilabnl.warmdesk.column:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: Archive
    state: absent

- name: Idempotent column setup using username/password
  ansilabnl.warmdesk.column:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    project: ops-board
    name: Backlog
    color: "#9E9E9E"
    wip_limit: 0
  register: col_result

- name: Show what changed
  ansible.builtin.debug:
    var: col_result.column
'''

RETURN = r'''
changed:
  description: Whether the module made any change on the server.
  type: bool
  returned: always
column:
  description:
    - The column object as returned by the WarmDesk API after the operation.
    - C(null) when I(state=absent) and the column did not exist, or after
      a successful deletion.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric database ID of the column.
      type: int
      sample: 7
    name:
      description: Display name.
      type: str
      sample: In Progress
    color:
      description: Hex colour string (may be empty).
      type: str
      sample: "#1E88E5"
    wip_limit:
      description: Work-in-progress limit; C(null) or C(0) means unlimited.
      type: int
      sample: 3
    position:
      description: Float used for ordering columns on the board.
      type: float
      sample: 2000.0
    project_id:
      description: ID of the owning project.
      type: int
      sample: 1
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)


def _find_column(columns, name):
    """Return the first column dict whose name matches, or None."""
    for col in columns:
        if col.get('name') == name:
            return col
    return None


def _needs_update(existing, color, wip_limit):
    """Return True if any desired attribute differs from the current value."""
    if color is not None and existing.get('color') != color:
        return True
    if wip_limit is not None and (existing.get('wip_limit') or 0) != wip_limit:
        return True
    return False


def _wip_limit_update_fields(wip_limit):
    """Return the PUT-body fragment for wip_limit on an *existing* column.

    UpdateColumn rejects wip_limit < 1 outright (unlike CreateColumn, which
    silently normalises 0 to "no limit"), so clearing an existing column's
    limit must go through the dedicated wip_limit_clear flag instead of
    sending wip_limit=0.
    """
    if wip_limit is None:
        return {}
    if wip_limit < 1:
        return {'wip_limit_clear': True}
    return {'wip_limit': wip_limit}


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        name=dict(type='str', required=True),
        color=dict(type='str', required=False),
        wip_limit=dict(type='int', required=False),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    project = module.params['project']
    name = module.params['name']
    color = module.params['color']
    wip_limit = module.params['wip_limit']
    state = module.params['state']

    try:
        client = WarmDeskClient.from_module(module)
        columns = client.get('/projects/%s/columns' % project)
    except WarmDeskAPIError as exc:
        module.fail_json(msg='Failed to list columns: %s (HTTP %s)' % (exc.message, exc.status))

    existing = _find_column(columns, name)

    # ------------------------------------------------------------------ absent
    if state == 'absent':
        if existing is None:
            module.exit_json(changed=False, column=None)
        if module.check_mode:
            module.exit_json(changed=True, column=None)
        try:
            client.delete('/projects/%s/columns/%d' % (project, existing['id']))
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to delete column "%s": %s (HTTP %s)' % (name, exc.message, exc.status)
            )
        module.exit_json(changed=True, column=None)

    # ----------------------------------------------------------------- present
    if existing is None:
        # Create
        if module.check_mode:
            module.exit_json(changed=True, column=None)
        body = {'name': name}
        if color is not None:
            body['color'] = color
        if wip_limit is not None:
            body['wip_limit'] = wip_limit
        try:
            created = client.post('/projects/%s/columns' % project, body)
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to create column "%s": %s (HTTP %s)' % (name, exc.message, exc.status)
            )
        module.exit_json(changed=True, column=created)

    # Update if needed
    if not _needs_update(existing, color, wip_limit):
        module.exit_json(changed=False, column=existing)

    if module.check_mode:
        module.exit_json(changed=True, column=existing)

    body = {'name': name}
    if color is not None:
        body['color'] = color
    body.update(_wip_limit_update_fields(wip_limit))
    try:
        updated = client.put(
            '/projects/%s/columns/%d' % (project, existing['id']), body
        )
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Failed to update column "%s": %s (HTTP %s)' % (name, exc.message, exc.status)
        )
    module.exit_json(changed=True, column=updated)


def main():
    run_module()


if __name__ == '__main__':
    main()
