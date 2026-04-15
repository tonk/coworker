# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: customer
short_description: Manage WarmDesk customers
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete WarmDesk customer records.
  - Optionally mark a customer as a favourite (starred) for the authenticating
    user.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Deletion is an admin-only operation server-side.  Associated projects are
    detached (not deleted) automatically by the API.
  - The C(starred) parameter controls whether the calling user has the customer
    starred.  It does not affect other users.

extends_documentation_fragment:
  - ansilab.warmdesk.auth

options:
  name:
    description:
      - Customer name.  Used as the idempotency key — must be unique within
        the WarmDesk instance.
    type: str
    required: true

  description:
    description:
      - Free-text description of the customer.
    type: str

  logo_url:
    description:
      - URL to the customer's logo image.
    type: str

  starred:
    description:
      - C(true) ensures the customer is in the calling user's favourites.
      - C(false) ensures the customer is I(not) in the calling user's
        favourites.
      - Omit (or C(null)) to leave the starred status unchanged.
    type: bool

  state:
    description:
      - C(present) ensures the customer exists with the specified attributes.
      - C(absent) removes the customer if it exists.  Requires admin
        credentials.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a new customer
# ---------------------------------------------------------------------------
- name: Create customer Acme Corp
  ansilab.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Acme Corp
    description: Our largest enterprise client.
    logo_url: https://cdn.example.com/logos/acme.png
    state: present

# ---------------------------------------------------------------------------
# Update description and mark as starred
# ---------------------------------------------------------------------------
- name: Update Acme Corp and star it
  ansilab.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    name: Acme Corp
    description: Largest enterprise client — Q1 contract renewed.
    starred: true

# ---------------------------------------------------------------------------
# Ensure a customer is unstarred (without changing other fields)
# ---------------------------------------------------------------------------
- name: Unstar Acme Corp
  ansilab.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Acme Corp
    starred: false

# ---------------------------------------------------------------------------
# Idempotent creation — safe to run in a provisioning loop
# ---------------------------------------------------------------------------
- name: Ensure all customers exist
  ansilab.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: "{{ item.name }}"
    description: "{{ item.description | default(omit) }}"
  loop: "{{ customers }}"

# ---------------------------------------------------------------------------
# Delete a customer (admin operation)
# ---------------------------------------------------------------------------
- name: Remove legacy customer
  ansilab.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: OldCo Inc
    state: absent
"""

RETURN = r"""
customer:
  description:
    - The customer object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the customer was deleted (or never
      existed).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric customer ID.
      returned: always
      type: int
      sample: 7
    name:
      description: Customer name.
      returned: always
      type: str
      sample: Acme Corp
    description:
      description: Customer description.
      returned: always
      type: str
      sample: Our largest enterprise client.
    logo_url:
      description: URL to the customer logo.
      returned: when set
      type: str
      sample: https://cdn.example.com/logos/acme.png
    is_favorite:
      description: Whether the customer is starred by the calling user.
      returned: always
      type: bool
      sample: true
    project_count:
      description: Number of projects linked to this customer.
      returned: always
      type: int
      sample: 3
    contract_count:
      description: Number of contracts linked to this customer.
      returned: always
      type: int
      sample: 2
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2025-03-01T12:00:00Z"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _find_customer(client, name):
    """Return the customer dict whose name matches, or None."""
    customers = client.get('/customers')
    for c in customers:
        if c.get('name') == name:
            return c
    return None


def _fetch_customer(client, customer_id):
    """Fetch the full CustomerDetailResponse for *customer_id*."""
    return client.get('/customers/%d' % customer_id)


def _build_body(p):
    """Build a create/update body from module params, omitting None values."""
    body = {}
    for key in ('name', 'description', 'logo_url'):
        if p.get(key) is not None:
            body[key] = p[key]
    return body


def _fields_changed(existing, p):
    """Return True if any mutable field in *p* differs from *existing*."""
    for api_key, param_key in (('name', 'name'),
                                ('description', 'description'),
                                ('logo_url', 'logo_url')):
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            return True
    return False


def _starred_action_needed(existing, desired_starred):
    """Return 'star', 'unstar', or None based on current and desired state."""
    if desired_starred is None:
        return None
    current = existing.get('is_favorite', False)
    if desired_starred and not current:
        return 'star'
    if not desired_starred and current:
        return 'unstar'
    return None


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        name=dict(type='str', required=True),
        description=dict(type='str'),
        logo_url=dict(type='str'),
        starred=dict(type='bool'),
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
        existing = _find_customer(client, p['name'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, customer=None)
            if not module.check_mode:
                client.delete('/customers/%d' % existing['id'])
            module.exit_json(changed=True, customer=None)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, customer=None)

            customer = client.post('/customers', _build_body(p))

            # Handle starring on a freshly created customer.
            star_action = _starred_action_needed(customer, p.get('starred'))
            if star_action == 'star':
                client.post('/customers/%d/favorite' % customer['id'])
                customer['is_favorite'] = True
            # 'unstar' on a brand-new customer is a no-op (default False).

            module.exit_json(changed=True, customer=customer)

        # ------------------------------------------------------------------ #
        # state=present — UPDATE                                               #
        # ------------------------------------------------------------------ #
        fields_differ = _fields_changed(existing, p)
        star_action = _starred_action_needed(existing, p.get('starred'))
        changed = fields_differ or (star_action is not None)

        if not changed:
            # Return the full detail object for a richer return value.
            customer = _fetch_customer(client, existing['id'])
            module.exit_json(changed=False, customer=customer)

        if module.check_mode:
            module.exit_json(changed=True, customer=existing)

        # Apply field updates when needed.
        if fields_differ:
            customer = client.put('/customers/%d' % existing['id'],
                                  _build_body(p))
        else:
            customer = _fetch_customer(client, existing['id'])

        # Apply starring change.
        if star_action == 'star':
            client.post('/customers/%d/favorite' % existing['id'])
            customer['is_favorite'] = True
        elif star_action == 'unstar':
            client.delete('/customers/%d/favorite' % existing['id'])
            customer['is_favorite'] = False

        module.exit_json(changed=True, customer=customer)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
