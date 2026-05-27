# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: ticket
short_description: Manage WarmDesk helpdesk tickets
version_added: "0.4.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete helpdesk tickets via the customer-scoped API.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires a user with customer access (member or admin role) or an admin
    account.
  - The customer is identified by name (C(customer)).  If the customer does
    not exist the module fails.
  - Usernames passed to C(assigned_to) and C(owner) are resolved to numeric
    IDs server-side.
  - Attempting to delete a ticket that does not exist is a no-op (no error).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  customer:
    description:
      - Customer name the ticket belongs to.
    type: str
    required: true

  title:
    description:
      - Ticket title / subject line.
      - Used together with C(customer) as the idempotency key.
    type: str
    required: true

  description:
    description:
      - Detailed description of the issue or request.
    type: str

  type:
    description:
      - Ticket type classification.
    type: str
    choices: [incident, problem, service_request, change_request]
    default: incident

  priority:
    description:
      - Priority level.
    type: str
    choices: [low, medium, high, critical]
    default: medium

  status:
    description:
      - Current status.
      - New tickets are always created with C(new).
      - C(pending) puts the ticket on hold pending a reminder date;
        C(pending_close) schedules automatic closing.
    type: str
    choices: [new, open, pending, pending_close, closed]

  assigned_to:
    description:
      - Username or email of the agent assigned to this ticket.
      - Set to an empty string C("") to unassign.
    type: str

  owner:
    description:
      - Username or email of the ticket owner (the person ultimately
        responsible).
      - Set to an empty string C("") to clear.
    type: str

  state:
    description:
      - C(present) ensures the ticket exists with the specified attributes.
      - C(absent) removes the ticket if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create a critical incident ticket for Acme Corp
  ansilabnl.warmdesk.ticket:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    customer: Acme Corporation
    title: Login page returns 500 on Safari
    description: |
      Users on Safari are seeing a 500 error after the latest deployment.
    type: incident
    priority: critical
    assigned_to: marc
    state: present

- name: Reassign ticket to a different agent
  ansilabnl.warmdesk.ticket:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    customer: Acme Corporation
    title: Login page returns 500 on Safari
    assigned_to: sarah

- name: Unassign a ticket (set to no one)
  ansilabnl.warmdesk.ticket:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    customer: Acme Corporation
    title: Login page returns 500 on Safari
    assigned_to: ""

- name: Update ticket priority and reassign in one pass
  ansilabnl.warmdesk.ticket:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    customer: Acme Corporation
    title: Login page returns 500 on Safari
    priority: high
    assigned_to: sarah

- name: Close a resolved ticket
  ansilabnl.warmdesk.ticket:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    customer: Globex Systems
    title: Kubernetes cluster node drain failing
    status: closed

- name: Delete a ticket
  ansilabnl.warmdesk.ticket:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    customer: Acme Corporation
    title: Old ticket no longer needed
    state: absent
"""

RETURN = r"""
ticket:
  description:
    - The ticket object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the ticket was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric ticket ID.
      returned: always
      type: int
      sample: 42
    customer_id:
      description: Numeric customer ID the ticket belongs to.
      returned: always
      type: int
      sample: 7
    title:
      description: Ticket subject.
      returned: always
      type: str
      sample: Login page returns 500 on Safari
    type:
      description: Ticket type.
      returned: always
      type: str
      sample: incident
    status:
      description: Current status.
      returned: always
      type: str
      sample: new
    priority:
      description: Priority level.
      returned: always
      type: str
      sample: critical
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2026-05-26T07:19:58Z"
"""

import re

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_resolve import (
    resolve_customer_id,
    resolve_user_id,
)


_DATE_PREFIX_RE = re.compile(r'^\[\d{4}-\d{2}-\d{2}\]\s*')


def _strip_date_prefix(title):
    return _DATE_PREFIX_RE.sub('', title or '')


def _find_ticket(client, customer_id, title):
    tickets = client.get('/customers/%d/tickets' % customer_id)
    for t in tickets:
        if _strip_date_prefix(t.get('title', '')) == title:
            return client.get('/customers/%d/tickets/%d' % (customer_id, t['id']))
    return None


def _build_update_body(p, existing, user_resolve_errors):
    body = {}
    changed = False

    comparable = {
        'description': 'description',
        'type': 'type',
        'priority': 'priority',
        'status': 'status',
    }
    for param_key, api_key in comparable.items():
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            body[api_key] = desired
            changed = True

    if p.get('assigned_to') is not None:
        val = p['assigned_to']
        if val == '':
            desired_id = 0
        else:
            try:
                desired_id = resolve_user_id(client, val)
            except WarmDeskAPIError as e:
                user_resolve_errors.append('assigned_to: %s' % str(e))
                desired_id = None
        if desired_id is not None:
            current_id = 0
            if existing.get('assigned_to') and existing['assigned_to'].get('id'):
                current_id = existing['assigned_to']['id']
            if current_id != desired_id:
                body['assigned_to_id'] = desired_id if desired_id != 0 else None
                changed = True

    if p.get('owner') is not None:
        val = p['owner']
        if val == '':
            desired_id = 0
        else:
            try:
                desired_id = resolve_user_id(client, val)
            except WarmDeskAPIError as e:
                user_resolve_errors.append('owner: %s' % str(e))
                desired_id = None
        if desired_id is not None:
            current_id = 0
            if existing.get('owner') and existing['owner'].get('id'):
                current_id = existing['owner']['id']
            if current_id != desired_id:
                body['owner_id'] = desired_id if desired_id != 0 else None
                changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        customer=dict(type='str', required=True),
        title=dict(type='str', required=True),
        description=dict(type='str'),
        type=dict(type='str', choices=['incident', 'problem', 'service_request', 'change_request']),
        priority=dict(type='str', choices=['low', 'medium', 'high', 'critical']),
        status=dict(type='str', choices=['new', 'open', 'pending', 'pending_close', 'closed']),
        assigned_to=dict(type='str'),
        owner=dict(type='str'),
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
        customer_id = resolve_customer_id(client, p['customer'])

        existing = _find_ticket(client, customer_id, p['title'])

        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, ticket=None)
            if not module.check_mode:
                client.delete('/customers/%d/tickets/%d' % (customer_id, existing['id']))
            module.exit_json(changed=True, ticket=None)

        # state=present - CREATE
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, ticket=None)

            body = dict(
                title=p['title'],
                type=p.get('type') or 'incident',
                priority=p.get('priority') or 'medium',
            )
            if p.get('description') is not None:
                body['description'] = p['description']
            if p.get('assigned_to'):
                try:
                    body['assigned_to_id'] = resolve_user_id(client, p['assigned_to'])
                except WarmDeskAPIError as e:
                    module.fail_json(msg='Failed to resolve assigned_to: %s' % str(e))
            if p.get('owner'):
                try:
                    body['owner_id'] = resolve_user_id(client, p['owner'])
                except WarmDeskAPIError as e:
                    module.fail_json(msg='Failed to resolve owner: %s' % str(e))

            ticket = client.post('/customers/%d/tickets' % customer_id, body)
            module.exit_json(changed=True, ticket=ticket)

        # state=present - UPDATE
        user_errors = []
        update_body, changed = _build_update_body(p, existing, user_errors)

        if not changed:
            module.exit_json(changed=False, ticket=existing)

        if module.check_mode:
            module.exit_json(changed=True, ticket=existing)

        if user_errors:
            module.fail_json(msg='Failed to resolve users: %s' % '; '.join(user_errors))

        ticket = client.put('/customers/%d/tickets/%d' % (customer_id, existing['id']), update_body)
        module.exit_json(changed=True, ticket=ticket)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
