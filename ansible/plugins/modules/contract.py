# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: contract
short_description: Manage WarmDesk customer contracts
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete contracts that belong to a WarmDesk customer.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - The idempotency key is the combination of C(customer) (name) and C(name).
    Contract names must therefore be unique within a customer.
  - Deletion is an admin-only operation server-side.  Associated projects are
    detached (not deleted) by the API automatically.
  - Date strings must be in C(YYYY-MM-DD) format or omitted entirely.  Passing
    an empty string clears a previously set date.

extends_documentation_fragment:
  - ansilab.warmdesk.auth

options:
  customer:
    description:
      - Name of the customer that owns this contract.
      - The customer must already exist; use M(ansilab.warmdesk.customer)
        to create it first.
    type: str
    required: true

  name:
    description:
      - Contract name.  Together with C(customer) this forms the idempotency
        key.
    type: str
    required: true

  description:
    description:
      - Free-text description of the contract.
    type: str

  start_date:
    description:
      - Contract start date in C(YYYY-MM-DD) format.
      - Omit to leave unset or unchanged.
    type: str

  end_date:
    description:
      - Contract end date in C(YYYY-MM-DD) format.
      - Omit to leave unset or unchanged.
    type: str

  state:
    description:
      - C(present) ensures the contract exists with the specified attributes.
      - C(absent) removes the contract if it exists.  Requires admin
        credentials.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a simple contract (no dates)
# ---------------------------------------------------------------------------
- name: Create annual support contract for Acme Corp
  ansilab.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Annual Support 2025
    description: 12-month support and maintenance agreement.
    state: present

# ---------------------------------------------------------------------------
# Create a contract with explicit start and end dates
# ---------------------------------------------------------------------------
- name: Create time-boxed project contract
  ansilab.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Phase 2 Implementation
    description: Fixed-price implementation engagement.
    start_date: "2025-04-01"
    end_date: "2025-09-30"

# ---------------------------------------------------------------------------
# Update an existing contract's description and extend the end date
# ---------------------------------------------------------------------------
- name: Extend Phase 2 contract
  ansilab.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    customer: Acme Corp
    name: Phase 2 Implementation
    description: Fixed-price implementation engagement (scope extended).
    end_date: "2025-12-31"

# ---------------------------------------------------------------------------
# Provision contracts from a variable list (idempotent loop)
# ---------------------------------------------------------------------------
- name: Ensure all contracts are present
  ansilab.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: "{{ item.customer }}"
    name: "{{ item.name }}"
    description: "{{ item.description | default(omit) }}"
    start_date: "{{ item.start_date | default(omit) }}"
    end_date: "{{ item.end_date | default(omit) }}"
  loop: "{{ contracts }}"

# ---------------------------------------------------------------------------
# Delete a contract (admin operation)
# ---------------------------------------------------------------------------
- name: Remove expired contract
  ansilab.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Annual Support 2024
    state: absent
"""

RETURN = r"""
contract:
  description:
    - The contract object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the contract was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric contract ID.
      returned: always
      type: int
      sample: 15
    name:
      description: Contract name.
      returned: always
      type: str
      sample: Annual Support 2025
    description:
      description: Contract description.
      returned: always
      type: str
      sample: 12-month support and maintenance agreement.
    customer_id:
      description: Numeric ID of the owning customer.
      returned: always
      type: int
      sample: 7
    start_date:
      description: Contract start date (YYYY-MM-DD) or null.
      returned: always
      type: str
      sample: "2025-01-01"
    end_date:
      description: Contract end date (YYYY-MM-DD) or null.
      returned: always
      type: str
      sample: "2025-12-31"
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2025-01-15T10:30:00Z"
    updated_at:
      description: ISO-8601 last-update timestamp.
      returned: always
      type: str
      sample: "2025-03-20T14:00:00Z"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)
from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_resolve import (
    resolve_customer_id,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_DATE_PARAMS = ('start_date', 'end_date')
_TEXT_PARAMS = ('name', 'description')


def _find_contract(client, customer_id, name):
    """Return the contract dict whose name matches within *customer_id*, or None."""
    contracts = client.get('/customers/%d/contracts' % customer_id)
    for c in contracts:
        if c.get('name') == name:
            return c
    return None


def _build_body(p):
    """Build a create/update body from module params, omitting None values.

    Empty strings are sent through unchanged so the caller can explicitly
    clear a date field.  Only Python None is treated as "not provided".
    """
    body = {}
    for key in _TEXT_PARAMS + _DATE_PARAMS:
        if p.get(key) is not None:
            body[key] = p[key]
    return body


def _fields_changed(existing, p):
    """Return True if any mutable field in *p* differs from *existing*.

    Date comparison is string-based (YYYY-MM-DD).  The API may return
    None / null for unset dates; we treat that as an empty string for
    comparison purposes so that an explicit empty-string param causes a write.
    """
    for key in _TEXT_PARAMS:
        desired = p.get(key)
        if desired is None:
            continue
        if existing.get(key) != desired:
            return True

    for key in _DATE_PARAMS:
        desired = p.get(key)
        if desired is None:
            continue
        # Normalise None in existing → empty string for comparison.
        current = existing.get(key) or ''
        if current != desired:
            return True

    return False


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        customer=dict(type='str', required=True),
        name=dict(type='str', required=True),
        description=dict(type='str'),
        start_date=dict(type='str'),
        end_date=dict(type='str'),
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
        # Resolve the customer name to its numeric ID.  This raises
        # WarmDeskAPIError(404, …) if the customer does not exist, which is
        # the correct behaviour — a contract cannot exist without a customer.
        customer_id = resolve_customer_id(client, p['customer'])

        existing = _find_contract(client, customer_id, p['name'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, contract=None)
            if not module.check_mode:
                client.delete(
                    '/customers/%d/contracts/%d' % (customer_id, existing['id'])
                )
            module.exit_json(changed=True, contract=None)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, contract=None)
            contract = client.post(
                '/customers/%d/contracts' % customer_id,
                _build_body(p),
            )
            module.exit_json(changed=True, contract=contract)

        # ------------------------------------------------------------------ #
        # state=present — UPDATE                                               #
        # ------------------------------------------------------------------ #
        if not _fields_changed(existing, p):
            module.exit_json(changed=False, contract=existing)

        if module.check_mode:
            module.exit_json(changed=True, contract=existing)

        contract = client.put(
            '/customers/%d/contracts/%d' % (customer_id, existing['id']),
            _build_body(p),
        )
        module.exit_json(changed=True, contract=contract)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
