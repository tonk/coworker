# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: sla_policy
short_description: Manage WarmDesk SLA policies
version_added: "0.4.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete SLA (Service Level Agreement) policies via the
    admin API.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires admin credentials (C(global_role=admin)) or a token obtained from
    an admin account.
  - The policy name is used as the idempotency key.
  - Deleting an SLA policy that does not exist is a no-op (no error).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  name:
    description:
      - Display name for the SLA policy.  Used as the idempotency key.
    type: str
    required: true

  response_time_minutes:
    description:
      - Maximum time in minutes before a first response is due.
      - Set to C(0) for no response deadline.
    type: int
    default: 0

  resolution_time_minutes:
    description:
      - Maximum time in minutes by which the ticket must be resolved.
      - Set to C(0) for no resolution deadline.
    type: int
    default: 0

  priority_filter:
    description:
      - Comma-separated list of priorities this policy applies to
        (e.g. C(critical,high)).
      - Leave empty to match all priorities (catch-all).
    type: str

  is_active:
    description:
      - Whether this policy is active and should be applied to matching
        tickets.
    type: bool
    default: true

  state:
    description:
      - C(present) ensures the SLA policy exists with the specified settings.
      - C(absent) removes the policy if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create a critical-priority SLA policy
  ansilabnl.warmdesk.sla_policy:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Critical Incident Response
    response_time_minutes: 15
    resolution_time_minutes: 240
    priority_filter: critical
    is_active: true
    state: present

- name: Create a catch-all SLA policy
  ansilabnl.warmdesk.sla_policy:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    name: Standard Support
    response_time_minutes: 60
    resolution_time_minutes: 2880
    is_active: true

- name: Update an existing policy
  ansilabnl.warmdesk.sla_policy:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Critical Incident Response
    response_time_minutes: 10
    resolution_time_minutes: 120

- name: Disable a policy without deleting it
  ansilabnl.warmdesk.sla_policy:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    name: Old Policy
    is_active: false

- name: Delete an SLA policy
  ansilabnl.warmdesk.sla_policy:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Old Policy
    state: absent
"""

RETURN = r"""
sla_policy:
  description:
    - The SLA policy object as returned by the WarmDesk API after the
      operation.
    - C(null) when C(state=absent) and the policy was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric SLA policy ID.
      returned: always
      type: int
      sample: 3
    name:
      description: Display name.
      returned: always
      type: str
      sample: Critical Incident Response
    response_time_minutes:
      description: Response deadline in minutes.
      returned: always
      type: int
      sample: 15
    resolution_time_minutes:
      description: Resolution deadline in minutes.
      returned: always
      type: int
      sample: 240
    priority_filter:
      description: Comma-separated priority filter.
      returned: when set
      type: str
      sample: critical
    is_active:
      description: Whether the policy is active.
      returned: always
      type: bool
      sample: true
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2026-05-26T10:00:00Z"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


def _find_policy(client, name):
    policies = client.get('/admin/sla-policies')
    for p in policies:
        if p.get('name') == name:
            return p
    return None


def _build_body(p, existing):
    body = {}
    changed = False

    comparable = {
        'response_time_minutes': 'response_time_minutes',
        'resolution_time_minutes': 'resolution_time_minutes',
        'priority_filter': 'priority_filter',
    }
    for param_key, api_key in comparable.items():
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            body[api_key] = desired
            changed = True

    if p.get('is_active') is not None:
        if existing.get('is_active') != p['is_active']:
            body['is_active'] = p['is_active']
            changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        name=dict(type='str', required=True),
        response_time_minutes=dict(type='int'),
        resolution_time_minutes=dict(type='int'),
        priority_filter=dict(type='str'),
        is_active=dict(type='bool'),
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
        existing = _find_policy(client, p['name'])

        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, sla_policy=None)
            if not module.check_mode:
                client.delete('/admin/sla-policies/%d' % existing['id'])
            module.exit_json(changed=True, sla_policy=None)

        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, sla_policy=None)
            body = dict(name=p['name'])
            for opt in ('response_time_minutes', 'resolution_time_minutes', 'priority_filter'):
                if p.get(opt) is not None:
                    body[opt] = p[opt]
            body['is_active'] = p.get('is_active', True)
            policy = client.post('/admin/sla-policies', body)
            module.exit_json(changed=True, sla_policy=policy)

        update_body, changed = _build_body(p, existing)

        if not changed:
            module.exit_json(changed=False, sla_policy=existing)

        if module.check_mode:
            module.exit_json(changed=True, sla_policy=existing)

        policy = client.put('/admin/sla-policies/%d' % existing['id'], update_body)
        module.exit_json(changed=True, sla_policy=policy)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
