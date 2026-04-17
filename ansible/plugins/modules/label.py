# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: label
short_description: Manage card labels in a WarmDesk project
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Creates, updates, or deletes a label inside a WarmDesk project.
  - Labels are project-scoped coloured tags that can be attached to cards.
  - Idempotent — a label is identified by I(project) slug + I(name). Running
    the module again with unchanged parameters produces no change.
  - Requires at least I(member) role on the project (or global admin).
options:
  project:
    description:
      - Slug of the WarmDesk project that owns this label (e.g. C(my-project)).
    type: str
    required: true
  name:
    description:
      - Display name of the label (e.g. C(bug), C(enhancement)).
      - Used as the idempotency key together with I(project).
    type: str
    required: true
  color:
    description:
      - Hex colour string for the label badge (e.g. C(#E53935)).
      - Required when creating a new label (I(state=present) and label does
        not yet exist). Ignored when I(state=absent).
    type: str
    required: false
  state:
    description:
      - C(present) — ensure the label exists; update I(color) if it differs.
      - C(absent) — ensure the label does not exist.
    type: str
    choices: [present, absent]
    default: present
extends_documentation_fragment:
  - ansilab.warmdesk.connection
notes:
  - Check mode is fully supported; no changes are made to the server.
  - Deleting a label does not remove it from cards that already carry it; the
    badge simply disappears from those cards.
seealso:
  - module: ansilab.warmdesk.card
  - module: ansilab.warmdesk.column
'''

EXAMPLES = r'''
- name: Ensure a "bug" label exists in red
  ansilab.warmdesk.label:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: bug
    color: "#E53935"

- name: Create multiple labels from a variable list
  ansilab.warmdesk.label:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: "{{ item.name }}"
    color: "{{ item.color }}"
  loop:
    - {name: bug,         color: "#E53935"}
    - {name: enhancement, color: "#43A047"}
    - {name: question,    color: "#1E88E5"}
    - {name: wontfix,     color: "#757575"}

- name: Rename the colour of the "urgent" label
  ansilab.warmdesk.label:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: urgent
    color: "#B71C1C"

- name: Remove a deprecated label
  ansilab.warmdesk.label:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    name: legacy
    state: absent

- name: Provision labels and capture results
  ansilab.warmdesk.label:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    project: ops-board
    name: infra
    color: "#6D4C41"
  register: label_result

- name: Show the label that was created or updated
  ansible.builtin.debug:
    var: label_result.label
'''

RETURN = r'''
changed:
  description: Whether the module made any change on the server.
  type: bool
  returned: always
label:
  description:
    - The label object as returned by the WarmDesk API after the operation.
    - C(null) when I(state=absent) and the label did not exist, or after a
      successful deletion.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric database ID of the label.
      type: int
      sample: 3
    name:
      description: Display name.
      type: str
      sample: bug
    color:
      description: Hex colour string.
      type: str
      sample: "#E53935"
    project_id:
      description: ID of the owning project.
      type: int
      sample: 1
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)


def _find_label(labels, name):
    """Return the first label dict whose name matches, or None."""
    for lbl in labels:
        if lbl.get('name') == name:
            return lbl
    return None


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        name=dict(type='str', required=True),
        color=dict(type='str', required=False),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    project = module.params['project']
    name = module.params['name']
    color = module.params['color']
    state = module.params['state']

    try:
        client = WarmDeskClient.from_module(module)
        labels = client.get('/projects/%s/labels' % project)
    except WarmDeskAPIError as exc:
        module.fail_json(msg='Failed to list labels: %s (HTTP %s)' % (exc.message, exc.status))

    existing = _find_label(labels, name)

    # ------------------------------------------------------------------ absent
    if state == 'absent':
        if existing is None:
            module.exit_json(changed=False, label=None)
        if module.check_mode:
            module.exit_json(changed=True, label=None)
        try:
            client.delete('/projects/%s/labels/%d' % (project, existing['id']))
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to delete label "%s": %s (HTTP %s)' % (name, exc.message, exc.status)
            )
        module.exit_json(changed=True, label=None)

    # ----------------------------------------------------------------- present
    if existing is None:
        # Create — color is required for new labels
        if not color:
            module.fail_json(
                msg='Parameter "color" is required when creating a new label.'
            )
        if module.check_mode:
            module.exit_json(changed=True, label=None)
        try:
            created = client.post(
                '/projects/%s/labels' % project,
                {'name': name, 'color': color},
            )
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to create label "%s": %s (HTTP %s)' % (name, exc.message, exc.status)
            )
        module.exit_json(changed=True, label=created)

    # Update colour if it changed
    if color is not None and existing.get('color') != color:
        if module.check_mode:
            module.exit_json(changed=True, label=existing)
        try:
            updated = client.put(
                '/projects/%s/labels/%d' % (project, existing['id']),
                {'name': name, 'color': color},
            )
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to update label "%s": %s (HTTP %s)' % (name, exc.message, exc.status)
            )
        module.exit_json(changed=True, label=updated)

    # No change needed
    module.exit_json(changed=False, label=existing)


def main():
    run_module()


if __name__ == '__main__':
    main()
