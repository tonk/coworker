# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: contract
short_description: Manage WarmDesk customer contracts
version_added: "0.1.0"
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
  - The API's update endpoint overwrites C(description), C(start_date),
    C(end_date), C(price_per_hour), C(price_per_km), and C(time_slots) from
    whatever is in the request body on every call — an omitted field is
    indistinguishable, server-side, from an explicit clear.  This module
    works around that by fetching the current contract and resending its
    existing value for any of those fields you don't explicitly set, so a
    task that only changes C(name) (for example) won't wipe the others.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  customer:
    description:
      - Name of the customer that owns this contract.
      - The customer must already exist; use M(ansilabnl.warmdesk.customer)
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
      - Omit to leave unset or unchanged; pass an empty string to clear it.
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

  price_per_hour:
    description:
      - Hourly rate for this contract, as a decimal string (e.g. C("75.00")).
      - Omit to leave unset or unchanged; pass an empty string to clear it.
    type: str

  price_per_km:
    description:
      - Per-kilometre travel rate for this contract, as a decimal string.
      - Omit to leave unset or unchanged; pass an empty string to clear it.
    type: str

  currency:
    description:
      - ISO 4217-style currency code or symbol (e.g. C(EUR), C(€)).
      - The API only applies this when non-empty — there is no way to clear
        it back to the server default through this endpoint.
    type: str

  time_slots:
    description:
      - Alternative hourly rates for work performed outside standard hours.
      - Omit to leave the existing time slots unchanged; pass an empty list
        to remove all of them.  When supplied, this is a full replacement of
        the contract's time slots, not a merge.
    type: list
    elements: dict
    suboptions:
      label:
        description: Free-text label for the slot (e.g. C(Evening)).
        type: str
      start_time:
        description: Start time in C(HH:MM) format.
        type: str
        required: true
      end_time:
        description: End time in C(HH:MM) format.
        type: str
        required: true
      day_type:
        description:
          - C(all), C(weekdays), C(weekends), or a comma-separated list of
            day names (C(monday), C(tuesday), …, C(sunday)).
        type: str
        default: all
      end_day_offset:
        description:
          - Number of calendar days after the anchor day when C(end_time)
            applies (C(1) = next morning, C(2) = two days later, …). Used
            when C(end_time) is earlier than C(start_time) (an overnight
            slot).
        type: int
        default: 0
      multiplication_factor:
        description: Rate multiplier applied instead of a flat hourly rate.
        type: float
      hourly_rate:
        description: Flat hourly rate for this slot, overriding the contract rate.
        type: float

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
  ansilabnl.warmdesk.contract:
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
  ansilabnl.warmdesk.contract:
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
  ansilabnl.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    customer: Acme Corp
    name: Phase 2 Implementation
    description: Fixed-price implementation engagement (scope extended).
    end_date: "2025-12-31"

# ---------------------------------------------------------------------------
# Create a contract with pricing and an evening rate time slot
# ---------------------------------------------------------------------------
- name: Create contract with rates and an out-of-hours time slot
  ansilabnl.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Standby Support 2025
    price_per_hour: "95.00"
    price_per_km: "0.35"
    currency: EUR
    time_slots:
      - label: Evening
        start_time: "19:00"
        end_time: "07:00"
        day_type: all
        end_day_offset: 1
        multiplication_factor: 1.5

# ---------------------------------------------------------------------------
# Clear a contract's hourly rate and remove all time slots
# ---------------------------------------------------------------------------
- name: Clear pricing on Standby Support 2025
  ansilabnl.warmdesk.contract:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Standby Support 2025
    price_per_hour: ""
    time_slots: []

# ---------------------------------------------------------------------------
# Provision contracts from a variable list (idempotent loop)
# ---------------------------------------------------------------------------
- name: Ensure all contracts are present
  ansilabnl.warmdesk.contract:
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
  ansilabnl.warmdesk.contract:
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
    price_per_hour:
      description: Hourly rate, or null when unset.
      returned: always
      type: float
      sample: 95.0
    price_per_km:
      description: Per-kilometre travel rate, or null when unset.
      returned: always
      type: float
      sample: 0.35
    currency:
      description: Currency code or symbol.
      returned: always
      type: str
      sample: EUR
    time_slots:
      description: Alternative rate time slots configured for this contract.
      returned: always
      type: list
      elements: dict
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
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_resolve import (
    resolve_customer_id,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_DATE_PARAMS = ('start_date', 'end_date')
_TEXT_PARAMS = ('name', 'description')
# Fields the API unconditionally overwrites from the request body on every
# update — an omitted key is indistinguishable from an explicit clear
# server-side, so the module must resend the current value for any of these
# it isn't explicitly changing. 'name' and 'currency' are excluded: the API
# only touches them when the request value is non-empty.
_MERGE_ON_UPDATE = ('description',) + _DATE_PARAMS


def _find_contract(client, customer_id, name):
    """Return the contract dict whose name matches within *customer_id*, or None."""
    contracts = client.get('/customers/%d/contracts' % customer_id)
    for c in contracts:
        if c.get('name') == name:
            return c
    return None


def _resolve_price(raw):
    """Parse a price_per_* string param.

    Returns (value, was_given). was_given is False when raw is None (the
    param was omitted — leave the field untouched). An empty string is an
    explicit request to clear the price back to null.
    """
    if raw is None:
        return None, False
    if raw == '':
        return None, True
    try:
        return float(raw), True
    except ValueError:
        raise ValueError('must be a decimal number or an empty string, got %r' % raw)


def _normalize_time_slots(slots):
    """Return time-slot dicts in the API's canonical shape (order preserved).

    Strips server-only keys (id, contract_id) and fills in the same defaults
    the backend applies, so slots freshly returned by the API and slots
    freshly supplied by the caller compare equal.
    """
    normalized = []
    for s in slots:
        normalized.append(dict(
            label=s.get('label') or '',
            start_time=s.get('start_time') or '',
            end_time=s.get('end_time') or '',
            day_type=s.get('day_type') or 'all',
            end_day_offset=s.get('end_day_offset') or 0,
            multiplication_factor=s.get('multiplication_factor'),
            hourly_rate=s.get('hourly_rate'),
        ))
    return normalized


def _build_body(p, existing=None):
    """Build a create/update body from module params.

    On update (existing is not None), fields in _MERGE_ON_UPDATE plus
    price_per_hour/price_per_km/time_slots are resent with their current
    value whenever the caller didn't explicitly supply a new one — see
    _MERGE_ON_UPDATE's comment for why this is necessary.
    """
    body = {}

    if p.get('name') is not None:
        body['name'] = p['name']
    if p.get('currency') is not None:
        body['currency'] = p['currency']

    for key in _MERGE_ON_UPDATE:
        if p.get(key) is not None:
            body[key] = p[key]
        elif existing is not None:
            body[key] = existing.get(key) or ''

    price_per_hour, given = _resolve_price(p.get('price_per_hour'))
    if given:
        body['price_per_hour'] = price_per_hour
    elif existing is not None:
        body['price_per_hour'] = existing.get('price_per_hour')

    price_per_km, given = _resolve_price(p.get('price_per_km'))
    if given:
        body['price_per_km'] = price_per_km
    elif existing is not None:
        body['price_per_km'] = existing.get('price_per_km')

    if p.get('time_slots') is not None:
        body['time_slots'] = _normalize_time_slots(p['time_slots'])
    elif existing is not None:
        body['time_slots'] = _normalize_time_slots(existing.get('time_slots') or [])

    return body


def _fields_changed(existing, p):
    """Return True if any mutable field explicitly supplied in *p* differs
    from *existing*. Merged (preserved) fields are equal to *existing* by
    construction and never contribute a difference here.
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

    if p.get('currency') is not None and existing.get('currency') != p['currency']:
        return True

    for key in ('price_per_hour', 'price_per_km'):
        desired, given = _resolve_price(p.get(key))
        if not given:
            continue
        if existing.get(key) != desired:
            return True

    if p.get('time_slots') is not None:
        desired_slots = _normalize_time_slots(p['time_slots'])
        current_slots = _normalize_time_slots(existing.get('time_slots') or [])
        if desired_slots != current_slots:
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
        price_per_hour=dict(type='str'),
        price_per_km=dict(type='str'),
        currency=dict(type='str'),
        time_slots=dict(type='list', elements='dict', options=dict(
            label=dict(type='str'),
            start_time=dict(type='str', required=True),
            end_time=dict(type='str', required=True),
            day_type=dict(type='str', default='all'),
            end_day_offset=dict(type='int', default=0),
            multiplication_factor=dict(type='float'),
            hourly_rate=dict(type='float'),
        )),
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
            _build_body(p, existing),
        )
        module.exit_json(changed=True, contract=contract)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))
    except ValueError as e:
        module.fail_json(msg=str(e))


def main():
    run_module()


if __name__ == '__main__':
    main()
