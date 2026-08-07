# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: customer
short_description: Manage WarmDesk customers
version_added: "0.1.0"
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
  - C(GET /customers) excludes hidden customers by default; this module
    always queries with C(include_hidden=true) so a customer previously
    hidden via the admin UI is found instead of creating a duplicate.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

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

  color:
    description:
      - Hex colour code for the customer's UI colour swatch (e.g. C(#3b82f6)).
      - Omit to leave unset or unchanged. The API only applies non-empty
        values, so this cannot be cleared back to empty through this module.
    type: str

  billing_street:
    description: Billing address street line. Same omit/no-clear behaviour as C(color).
    type: str

  billing_city:
    description: Billing address city. Same omit/no-clear behaviour as C(color).
    type: str

  billing_postal_code:
    description: Billing address postal code. Same omit/no-clear behaviour as C(color).
    type: str

  billing_country:
    description: Billing address country. Same omit/no-clear behaviour as C(color).
    type: str

  vat_number:
    description: VAT/tax identification number. Same omit/no-clear behaviour as C(color).
    type: str

  po_reference:
    description: Default purchase-order reference. Same omit/no-clear behaviour as C(color).
    type: str

  is_hidden:
    description:
      - Whether the customer is hidden from the default (non-admin) customer
        list. Requires admin credentials to take effect; silently ignored
        server-side otherwise.
      - Omit to leave unchanged.
    type: bool

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
  ansilabnl.warmdesk.customer:
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
  ansilabnl.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    name: Acme Corp
    description: Largest enterprise client — Q1 contract renewed.
    starred: true

# ---------------------------------------------------------------------------
# Set billing details and VAT/PO reference
# ---------------------------------------------------------------------------
- name: Set Acme Corp billing details
  ansilabnl.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Acme Corp
    color: "#3b82f6"
    billing_street: "1 Main St"
    billing_city: Amsterdam
    billing_postal_code: "1011 AB"
    billing_country: Netherlands
    vat_number: NL123456789B01
    po_reference: PO-2025-042

# ---------------------------------------------------------------------------
# Hide a customer from the default customer list (admin only)
# ---------------------------------------------------------------------------
- name: Hide legacy customer without deleting it
  ansilabnl.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    name: OldCo Inc
    is_hidden: true

# ---------------------------------------------------------------------------
# Ensure a customer is unstarred (without changing other fields)
# ---------------------------------------------------------------------------
- name: Unstar Acme Corp
  ansilabnl.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: Acme Corp
    starred: false

# ---------------------------------------------------------------------------
# Idempotent creation — safe to run in a provisioning loop
# ---------------------------------------------------------------------------
- name: Ensure all customers exist
  ansilabnl.warmdesk.customer:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    name: "{{ item.name }}"
    description: "{{ item.description | default(omit) }}"
  loop: "{{ customers }}"

# ---------------------------------------------------------------------------
# Delete a customer (admin operation)
# ---------------------------------------------------------------------------
- name: Remove legacy customer
  ansilabnl.warmdesk.customer:
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
    color:
      description: Hex colour code for the customer's UI swatch.
      returned: when set
      type: str
      sample: "#3b82f6"
    billing_street:
      description: Billing address street line.
      returned: when set
      type: str
      sample: "1 Main St"
    billing_city:
      description: Billing address city.
      returned: when set
      type: str
      sample: Amsterdam
    billing_postal_code:
      description: Billing address postal code.
      returned: when set
      type: str
      sample: "1011 AB"
    billing_country:
      description: Billing address country.
      returned: when set
      type: str
      sample: Netherlands
    vat_number:
      description: VAT/tax identification number.
      returned: when set
      type: str
      sample: NL123456789B01
    po_reference:
      description: Default purchase-order reference.
      returned: when set
      type: str
      sample: PO-2025-042
    is_hidden:
      description: Whether the customer is hidden from the default customer list.
      returned: always
      type: bool
      sample: false
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
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_TEXT_FIELDS = (
    'name', 'description', 'logo_url', 'color',
    'billing_street', 'billing_city', 'billing_postal_code', 'billing_country',
    'vat_number', 'po_reference',
)


def _find_customer(client, name):
    """Return the customer dict whose name matches, or None.

    Includes hidden customers (GET /customers excludes them by default) so a
    customer hidden via the admin UI is found instead of creating a
    duplicate.
    """
    customers = client.get('/customers', params={'include_hidden': 'true'})
    for c in customers:
        if c.get('name') == name:
            return c
    return None


def _fetch_customer(client, customer_id):
    """Fetch the flat customer object for *customer_id*.

    GET /customers/:id returns {customer: {...}, contracts: [...], ...} —
    unwrap it so the return value matches the flat shape POST/PUT return.
    """
    return client.get('/customers/%d' % customer_id)['customer']


def _build_body(p):
    """Build a create/update body from module params, omitting None values."""
    body = {}
    for key in _TEXT_FIELDS:
        if p.get(key) is not None:
            body[key] = p[key]
    if p.get('is_hidden') is not None:
        body['is_hidden'] = p['is_hidden']
    return body


def _fields_changed(existing, p):
    """Return True if any mutable field in *p* differs from *existing*."""
    for key in _TEXT_FIELDS:
        desired = p.get(key)
        if desired is None:
            continue
        if existing.get(key) != desired:
            return True
    if p.get('is_hidden') is not None and existing.get('is_hidden', False) != p['is_hidden']:
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
        color=dict(type='str'),
        billing_street=dict(type='str'),
        billing_city=dict(type='str'),
        billing_postal_code=dict(type='str'),
        billing_country=dict(type='str'),
        vat_number=dict(type='str'),
        po_reference=dict(type='str'),
        is_hidden=dict(type='bool'),
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

            # is_hidden is not accepted by the create endpoint — apply it
            # with a follow-up update when explicitly requested as true
            # (new customers already default to false).
            if p.get('is_hidden'):
                customer = client.put('/customers/%d' % customer['id'], {'is_hidden': True})

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
