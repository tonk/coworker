# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: epic
short_description: Manage WarmDesk epics
version_added: "0.5.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete Scrum epics on a WarmDesk project board.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires at least I(member) role on the project (or global admin).
  - The epic name is used as the idempotency key within a project.
  - Deleting an epic that does not exist is a no-op (no error).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  project:
    description:
      - Slug of the WarmDesk project the epic belongs to (e.g. C(my-project)).
    type: str
    required: true

  name:
    description:
      - Display name of the epic.  Used as the idempotency key within a project.
    type: str
    required: true

  description:
    description:
      - Optional description for the epic.
    type: str

  color:
    description:
      - Hex colour string for the epic label (e.g. C(#6366f1)).
      - Defaults to C(#6366f1) server-side when omitted on creation.
    type: str

  status:
    description:
      - Lifecycle status of the epic.
      - C(open) means the epic is in progress; C(closed) means it is done.
    type: str
    choices: [open, closed]

  state:
    description:
      - C(present) ensures the epic exists with the specified attributes.
      - C(absent) removes the epic if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create an epic for the authentication milestone
  ansilabnl.warmdesk.epic:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Auth overhaul
    description: Consolidate all authentication flows into a single service.
    color: "#ef4444"

- name: Close a completed epic
  ansilabnl.warmdesk.epic:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Auth overhaul
    status: closed

- name: Reopen a previously closed epic
  ansilabnl.warmdesk.epic:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Auth overhaul
    status: open

- name: Remove an epic
  ansilabnl.warmdesk.epic:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Auth overhaul
    state: absent
"""

RETURN = r"""
epic:
  description:
    - The epic object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the epic was deleted (or did not exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric epic ID.
      returned: always
      type: int
      sample: 3
    project_id:
      description: Numeric ID of the owning project.
      returned: always
      type: int
      sample: 1
    name:
      description: Epic name.
      returned: always
      type: str
      sample: Auth overhaul
    description:
      description: Epic description.
      returned: always
      type: str
      sample: Consolidate all authentication flows.
    color:
      description: Hex colour string.
      returned: always
      type: str
      sample: "#6366f1"
    status:
      description: C(open) or C(closed).
      returned: always
      type: str
      sample: open
    card_count:
      description: Total number of cards in this epic.
      returned: always
      type: int
      sample: 12
    done_count:
      description: Number of completed cards in this epic.
      returned: always
      type: int
      sample: 7
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2026-01-10T08:00:00Z"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


def _find_epic(client, project_slug, name):
    """Return the epic dict whose name matches, or None."""
    epics = client.get('/projects/%s/epics' % project_slug)
    for e in epics:
        if e.get('name') == name:
            return e
    return None


def _build_update_body(p, existing):
    body = {}
    changed = False
    for param_key, api_key in (
        ('description', 'description'),
        ('color', 'color'),
        ('status', 'status'),
    ):
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            body[api_key] = desired
            changed = True
    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        name=dict(type='str', required=True),
        description=dict(type='str'),
        color=dict(type='str'),
        status=dict(type='str', choices=['open', 'closed']),
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
        existing = _find_epic(client, p['project'], p['name'])

        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, epic=None)
            if not module.check_mode:
                client.delete('/projects/%s/epics/%d' % (p['project'], existing['id']))
            module.exit_json(changed=True, epic=None)

        # state=present — CREATE
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, epic=None)
            body = dict(name=p['name'])
            for optional in ('description', 'color'):
                if p.get(optional) is not None:
                    body[optional] = p[optional]
            epic = client.post('/projects/%s/epics' % p['project'], body)
            # status defaults to 'open' on create; apply if caller wants 'closed'
            if p.get('status') and p['status'] != epic.get('status', 'open'):
                epic = client.put('/projects/%s/epics/%d' % (p['project'], epic['id']),
                                  {'status': p['status']})
            module.exit_json(changed=True, epic=epic)

        # state=present — UPDATE
        update_body, changed = _build_update_body(p, existing)

        if not changed:
            module.exit_json(changed=False, epic=existing)
        if module.check_mode:
            module.exit_json(changed=True, epic=existing)

        epic = client.put('/projects/%s/epics/%d' % (p['project'], existing['id']), update_body)
        module.exit_json(changed=True, epic=epic)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
