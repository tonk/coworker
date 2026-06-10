# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: ticket_checklist_template
short_description: Manage WarmDesk ticket checklist templates
version_added: "0.5.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete admin-managed ticket checklist templates.
  - Templates are reusable item lists that agents can apply to helpdesk tickets
    in one click.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires admin credentials (C(global_role=admin)) or a token obtained from
    an admin account.
  - The template name is used as the idempotency key.
  - Deleting a template that does not exist is a no-op (no error).
  - The C(items) list comparison is order-sensitive — reordering items counts as
    a change.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  name:
    description:
      - Display name for the template.  Used as the idempotency key.
    type: str
    required: true

  description:
    description:
      - Optional description explaining when to use this template.
    type: str

  items:
    description:
      - Ordered list of checklist item texts that are created when the template
        is applied to a ticket.
    type: list
    elements: str

  is_active:
    description:
      - Whether this template is visible in the apply-template dropdown.
    type: bool
    default: true

  sort_order:
    description:
      - Numeric sort order; lower values appear first in the template list.
    type: int
    default: 0

  state:
    description:
      - C(present) ensures the template exists with the specified attributes.
      - C(absent) removes the template if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create a deployment checklist template
  ansilabnl.warmdesk.ticket_checklist_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Deployment checklist
    description: Steps to verify after every production deployment.
    items:
      - Check health endpoints return 200
      - Verify error rate in Grafana
      - Confirm no new alerts firing
      - Update the deployment log
    is_active: true
    sort_order: 10

- name: Add an onboarding template
  ansilabnl.warmdesk.ticket_checklist_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: New customer onboarding
    items:
      - Send welcome email
      - Create user account
      - Assign customer success manager
      - Schedule kick-off call

- name: Deactivate a template without removing it
  ansilabnl.warmdesk.ticket_checklist_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Old incident template
    is_active: false

- name: Remove a template
  ansilabnl.warmdesk.ticket_checklist_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Deployment checklist
    state: absent
"""

RETURN = r"""
template:
  description:
    - The checklist template object as returned by the WarmDesk API after the
      operation.
    - C(null) when C(state=absent) and the template was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric template ID.
      returned: always
      type: int
      sample: 4
    name:
      description: Template name.
      returned: always
      type: str
      sample: Deployment checklist
    description:
      description: Description text.
      returned: always
      type: str
      sample: Steps to verify after every production deployment.
    items:
      description: Ordered list of checklist item strings.
      returned: always
      type: list
      elements: str
      sample:
        - Check health endpoints return 200
        - Verify error rate in Grafana
    is_active:
      description: Whether the template is active.
      returned: always
      type: bool
      sample: true
    sort_order:
      description: Sort position (lower = first).
      returned: always
      type: int
      sample: 10
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


def _find_template(client, name):
    """Return the template dict whose name matches, or None."""
    templates = client.get('/admin/ticket-checklist-templates')
    for t in templates:
        if t.get('name') == name:
            return t
    return None


def _build_update_body(p, existing):
    body = {}
    changed = False

    for param_key, api_key in (
        ('description', 'description'),
        ('is_active', 'is_active'),
        ('sort_order', 'sort_order'),
    ):
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            body[api_key] = desired
            changed = True

    if p.get('items') is not None:
        existing_items = existing.get('items') or []
        if list(existing_items) != list(p['items']):
            body['items'] = p['items']
            changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        name=dict(type='str', required=True),
        description=dict(type='str'),
        items=dict(type='list', elements='str'),
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
        existing = _find_template(client, p['name'])

        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, template=None)
            if not module.check_mode:
                client.delete('/admin/ticket-checklist-templates/%d' % existing['id'])
            module.exit_json(changed=True, template=None)

        # state=present — CREATE
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, template=None)
            body = dict(name=p['name'])
            for optional in ('description', 'sort_order'):
                if p.get(optional) is not None:
                    body[optional] = p[optional]
            body['items'] = p['items'] if p.get('items') is not None else []
            if p.get('is_active') is not None:
                body['is_active'] = p['is_active']
            tmpl = client.post('/admin/ticket-checklist-templates', body)
            module.exit_json(changed=True, template=tmpl)

        # state=present — UPDATE
        update_body, changed = _build_update_body(p, existing)

        if not changed:
            module.exit_json(changed=False, template=existing)
        if module.check_mode:
            module.exit_json(changed=True, template=existing)

        tmpl = client.put('/admin/ticket-checklist-templates/%d' % existing['id'], update_body)
        module.exit_json(changed=True, template=tmpl)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
