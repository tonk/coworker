# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: invoice_template
short_description: Manage WarmDesk invoice templates
version_added: "0.6.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete invoice templates via the admin API.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires admin credentials (C(global_role=admin)) or a token obtained from
    an admin account.
  - The template name is used as the idempotency key.
  - Deleting an invoice template that does not exist is a no-op (no error).
  - C(line_items) are compared by C(description), C(quantity), and
    C(unit_price) in order.  The module computes C(amount) and sets
    C(is_manual=true) automatically before sending to the API.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  name:
    description:
      - Display name for the invoice template.  Used as the idempotency key.
    type: str
    required: true

  line_items:
    description:
      - List of line items to include in the template.
      - Each item must have C(description), C(quantity), C(unit_price), and
        C(currency).
      - C(amount) is computed as C(quantity * unit_price) and C(is_manual) is
        always set to C(true) before the data is sent to the API.
    type: list
    elements: dict
    suboptions:
      description:
        description: Free-text description of the line item.
        type: str
        required: true
      quantity:
        description: Number of units.
        type: float
        required: true
      unit_price:
        description: Price per unit.
        type: float
        required: true
      currency:
        description: Currency symbol or code for this line item.
        type: str
        required: true

  default_vat_rate:
    description:
      - Default VAT percentage applied to the template totals.
      - Set to C(0.0) for no VAT.
    type: float
    default: 0.0

  default_currency:
    description:
      - Default currency symbol shown on the generated invoice.
    type: str
    default: "€"

  notes:
    description:
      - Free-text notes printed at the bottom of invoices generated from this
        template.
    type: str

  state:
    description:
      - C(present) ensures the invoice template exists with the specified
        settings.
      - C(absent) removes the template if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create a minimal invoice template
  ansilabnl.warmdesk.invoice_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Basic Support Contract
    default_currency: "€"
    default_vat_rate: 21.0
    state: present

- name: Create a template with line items
  ansilabnl.warmdesk.invoice_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Managed Services Monthly
    default_currency: "€"
    default_vat_rate: 21.0
    notes: "Payment due within 30 days."
    line_items:
      - description: "Managed hosting — monthly fee"
        quantity: 1
        unit_price: 499.00
        currency: "€"
      - description: "Additional storage (per GB)"
        quantity: 50
        unit_price: 0.10
        currency: "€"
    state: present

- name: Update the VAT rate and notes on an existing template
  ansilabnl.warmdesk.invoice_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    name: Managed Services Monthly
    default_vat_rate: 0.0
    notes: "VAT exempt — article 44 EU VAT Directive."
    state: present

- name: Replace line items on an existing template
  ansilabnl.warmdesk.invoice_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Basic Support Contract
    default_currency: "€"
    default_vat_rate: 21.0
    line_items:
      - description: "Support hours"
        quantity: 10
        unit_price: 85.00
        currency: "€"
    state: present

- name: Delete an invoice template
  ansilabnl.warmdesk.invoice_template:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    name: Basic Support Contract
    state: absent
"""

RETURN = r"""
invoice_template:
  description:
    - The invoice template object as returned by the WarmDesk API after the
      operation.
    - C(null) when C(state=absent) and the template was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric invoice template ID.
      returned: always
      type: int
      sample: 7
    name:
      description: Display name.
      returned: always
      type: str
      sample: Managed Services Monthly
    line_items:
      description: JSON string encoding the list of line items.
      returned: always
      type: str
      sample: '[{"description":"Hosting","quantity":1,"unit_price":499.0,"amount":499.0,"currency":"€","is_manual":true}]'
    default_vat_rate:
      description: Default VAT percentage.
      returned: always
      type: float
      sample: 21.0
    default_currency:
      description: Default currency symbol.
      returned: always
      type: str
      sample: "€"
    notes:
      description: Free-text notes for the template.
      returned: when set
      type: str
      sample: "Payment due within 30 days."
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2026-05-26T10:00:00Z"
    updated_at:
      description: ISO-8601 last-update timestamp.
      returned: always
      type: str
      sample: "2026-06-01T14:30:00Z"
"""

import json

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


def _find_template(client, name):
    templates = client.get('/invoice-templates')
    for t in templates:
        if t.get('name') == name:
            return t
    return None


def _build_line_items(raw_items):
    """Normalise the module's line_items list into API-ready dicts."""
    result = []
    for item in raw_items:
        qty = float(item['quantity'])
        price = float(item['unit_price'])
        result.append({
            'description': item['description'],
            'quantity': qty,
            'unit_price': price,
            'amount': round(qty * price, 10),
            'currency': item['currency'],
            'is_manual': True,
        })
    return result


def _parse_existing_line_items(existing):
    """Return a list of dicts from the stored JSON string (or list)."""
    raw = existing.get('line_items')
    if not raw:
        return []
    if isinstance(raw, list):
        return raw
    try:
        return json.loads(raw)
    except (ValueError, TypeError):
        return []


def _line_items_changed(desired_items, existing_items):
    """Compare by description/quantity/unit_price in order."""
    if len(desired_items) != len(existing_items):
        return True
    for d, e in zip(desired_items, existing_items):
        if d['description'] != e.get('description'):
            return True
        if float(d['quantity']) != float(e.get('quantity', 0)):
            return True
        if float(d['unit_price']) != float(e.get('unit_price', 0)):
            return True
    return False


def _build_update_body(p, existing, desired_items):
    """Return (body, changed) for a PUT call against an existing template.

    The backend requires ``name`` in every PUT body, so it is always included.
    """
    # name is always required by the API even on partial updates.
    body = {'name': p['name']}
    changed = False

    # Allow renaming (name change) to be detected via the name param itself.
    if p['name'] != existing.get('name'):
        changed = True

    # Scalar fields
    for param_key in ('default_vat_rate', 'default_currency', 'notes'):
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(param_key) != desired:
            body[param_key] = desired
            changed = True

    # line_items — only compare when the caller supplied them
    if p.get('line_items') is not None:
        existing_items = _parse_existing_line_items(existing)
        if _line_items_changed(desired_items, existing_items):
            body['line_items'] = json.dumps(desired_items)
            changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        name=dict(type='str', required=True),
        line_items=dict(
            type='list',
            elements='dict',
            options=dict(
                description=dict(type='str', required=True),
                quantity=dict(type='float', required=True),
                unit_price=dict(type='float', required=True),
                currency=dict(type='str', required=True),
            ),
        ),
        default_vat_rate=dict(type='float'),
        default_currency=dict(type='str'),
        notes=dict(type='str'),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']
    client = WarmDeskClient.from_module(module)

    # Pre-compute normalised line items once (avoids repetition below)
    desired_items = _build_line_items(p['line_items']) if p.get('line_items') is not None else []

    try:
        existing = _find_template(client, p['name'])

        # ------------------------------------------------------------------ #
        # state: absent
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, invoice_template=None)
            if not module.check_mode:
                client.delete('/admin/invoice-templates/%d' % existing['id'])
            module.exit_json(changed=True, invoice_template=None)

        # ------------------------------------------------------------------ #
        # state: present — create
        # ------------------------------------------------------------------ #
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, invoice_template=None)

            body = dict(name=p['name'])
            body['default_vat_rate'] = p['default_vat_rate'] if p.get('default_vat_rate') is not None else 0.0
            body['default_currency'] = p['default_currency'] if p.get('default_currency') is not None else '€'
            if p.get('notes') is not None:
                body['notes'] = p['notes']
            body['line_items'] = json.dumps(desired_items)

            template = client.post('/admin/invoice-templates', body)
            module.exit_json(changed=True, invoice_template=template)

        # ------------------------------------------------------------------ #
        # state: present — update
        # ------------------------------------------------------------------ #
        update_body, changed = _build_update_body(p, existing, desired_items)

        if not changed:
            module.exit_json(changed=False, invoice_template=existing)

        if module.check_mode:
            module.exit_json(changed=True, invoice_template=existing)

        template = client.put('/admin/invoice-templates/%d' % existing['id'], update_body)
        module.exit_json(changed=True, invoice_template=template)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
