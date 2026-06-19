# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: invoice
short_description: Manage WarmDesk customer invoices
version_added: "0.6.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete invoices that belong to a WarmDesk customer.
  - When C(invoice_number) is provided the module is fully idempotent — it
    locates the invoice by number and only writes when mutable fields change.
  - When C(invoice_number) is omitted a new invoice is always created
    (the module warns about this non-idempotent behaviour).
  - Supports check mode.
notes:
  - Invoice numbers are assigned by the server on creation.  You cannot
    specify a number at creation time; use C(invoice_number) only to
    reference an already-existing invoice.
  - Deletion is only allowed for invoices whose status is C(draft) or
    C(credit_note).  Attempting to delete a C(sent) or C(paid) invoice
    causes the module to fail.
  - Date strings must be in C(YYYY-MM-DD) format.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  customer:
    description:
      - Name of the customer that owns this invoice.
      - The customer must already exist; use M(ansilabnl.warmdesk.customer)
        to create it first.
    type: str
    required: true

  invoice_number:
    description:
      - Server-assigned invoice number (e.g. C(INV-0042)).
      - When provided, acts as the idempotency key.  If an invoice with this
        number exists it is updated; if it does not exist the module fails
        with a clear error (the server generates numbers — you cannot create
        one with a specific number).
      - Required when C(state=absent).
      - Omit to always create a new invoice (non-idempotent; a warning is
        emitted).
    type: str

  period_start:
    description:
      - Billing period start date in C(YYYY-MM-DD) format.
      - Required when creating a new invoice (C(invoice_number) not provided
        and C(state=present)).
    type: str

  period_end:
    description:
      - Billing period end date in C(YYYY-MM-DD) format.
      - Required when creating a new invoice (C(invoice_number) not provided
        and C(state=present)).
    type: str

  line_items:
    description:
      - List of line items on the invoice.
      - Required when creating a new invoice (C(invoice_number) not provided
        and C(state=present)).
      - Each item is a dict with the fields below.  C(amount) is computed
        automatically as C(quantity * unit_price) and must not be supplied.
    type: list
    elements: dict
    suboptions:
      description:
        description: Human-readable line description.
        type: str
        required: true
      quantity:
        description: Number of units.
        type: float
        default: 1
      unit_price:
        description: Price per unit.
        type: float
        required: true
      currency:
        description: ISO-4217 currency code for this line (e.g. C(EUR)).
        type: str
      is_manual:
        description: When C(true) the line was entered manually rather than
          derived from time entries.
        type: bool

  currency:
    description:
      - Fallback ISO-4217 currency code for the invoice (e.g. C(EUR)).
    type: str

  vat_rate:
    description:
      - VAT percentage to apply (e.g. C(21.0) for 21 %).
    type: float
    default: 0.0

  due_date:
    description:
      - Payment due date in C(YYYY-MM-DD) format.
    type: str

  notes:
    description:
      - Free-text notes attached to the invoice.
    type: str

  status:
    description:
      - Target status for an existing invoice (update only).
      - Ignored when creating — newly created invoices always start as
        C(draft).
    type: str
    choices: [draft, sent, paid]

  payment_date:
    description:
      - Date payment was received, in C(YYYY-MM-DD) format.
      - Typically set together with C(status=paid).
    type: str

  payment_amount:
    description:
      - Amount actually received (may differ from invoice total).
    type: float

  payment_reference:
    description:
      - Bank reference, cheque number, or other payment identifier.
    type: str

  payment_method:
    description:
      - Method by which the invoice was settled.
    type: str
    choices: [bank, card, cash, other]

  state:
    description:
      - C(present) ensures the invoice exists with the given attributes.
      - C(absent) removes the invoice.  C(invoice_number) is required.
        Only C(draft) and C(credit_note) invoices may be deleted.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a new draft invoice (non-idempotent — a new invoice each run)
# ---------------------------------------------------------------------------
- name: Create draft invoice for Acme Corp
  ansilabnl.warmdesk.invoice:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    period_start: "2025-01-01"
    period_end: "2025-03-31"
    currency: EUR
    vat_rate: 21.0
    due_date: "2025-04-30"
    notes: Q1 2025 support services.
    line_items:
      - description: Monthly support — January
        quantity: 1
        unit_price: 1500.00
        currency: EUR
        is_manual: true
      - description: Monthly support — February
        quantity: 1
        unit_price: 1500.00
        currency: EUR
        is_manual: true
      - description: Monthly support — March
        quantity: 1
        unit_price: 1500.00
        currency: EUR
        is_manual: true
    state: present

# ---------------------------------------------------------------------------
# Update status to sent (idempotent via invoice_number)
# ---------------------------------------------------------------------------
- name: Mark invoice INV-0042 as sent
  ansilabnl.warmdesk.invoice:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    invoice_number: INV-0042
    status: sent
    state: present

# ---------------------------------------------------------------------------
# Record a bank payment against an invoice
# ---------------------------------------------------------------------------
- name: Record payment for INV-0042
  ansilabnl.warmdesk.invoice:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    invoice_number: INV-0042
    status: paid
    payment_date: "2025-05-03"
    payment_amount: 5445.00
    payment_reference: "WIRE-20250503-001"
    payment_method: bank
    state: present

# ---------------------------------------------------------------------------
# Idempotent update — only writes when notes or due_date actually change
# ---------------------------------------------------------------------------
- name: Ensure invoice notes and due date are up to date
  ansilabnl.warmdesk.invoice:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    invoice_number: INV-0042
    notes: Q1 2025 support services — amended scope.
    due_date: "2025-05-15"
    state: present

# ---------------------------------------------------------------------------
# Delete a draft invoice
# ---------------------------------------------------------------------------
- name: Remove draft invoice INV-0099
  ansilabnl.warmdesk.invoice:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    invoice_number: INV-0099
    state: absent

# ---------------------------------------------------------------------------
# Loop — generate one invoice per entry in a variable list
# ---------------------------------------------------------------------------
- name: Create invoices from billing data
  ansilabnl.warmdesk.invoice:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: "{{ item.customer }}"
    period_start: "{{ item.period_start }}"
    period_end: "{{ item.period_end }}"
    currency: "{{ item.currency | default('EUR') }}"
    vat_rate: "{{ item.vat_rate | default(0.0) }}"
    due_date: "{{ item.due_date | default(omit) }}"
    notes: "{{ item.notes | default(omit) }}"
    line_items: "{{ item.line_items }}"
    state: present
  loop: "{{ billing_runs }}"
"""

RETURN = r"""
invoice:
  description:
    - The invoice object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the invoice was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric invoice ID.
      returned: always
      type: int
      sample: 42
    invoice_number:
      description: Server-assigned invoice number.
      returned: always
      type: str
      sample: INV-0042
    customer_id:
      description: Numeric ID of the owning customer.
      returned: always
      type: int
      sample: 7
    period_start:
      description: Billing period start date (YYYY-MM-DD).
      returned: always
      type: str
      sample: "2025-01-01"
    period_end:
      description: Billing period end date (YYYY-MM-DD).
      returned: always
      type: str
      sample: "2025-03-31"
    status:
      description: Invoice status (draft, sent, paid, credit_note).
      returned: always
      type: str
      sample: draft
    currency:
      description: ISO-4217 currency code.
      returned: always
      type: str
      sample: EUR
    subtotal:
      description: Sum of all line-item amounts before VAT.
      returned: always
      type: float
      sample: 4500.00
    vat_rate:
      description: VAT percentage applied.
      returned: always
      type: float
      sample: 21.0
    vat_amount:
      description: Computed VAT amount.
      returned: always
      type: float
      sample: 945.00
    total:
      description: Total amount including VAT.
      returned: always
      type: float
      sample: 5445.00
    due_date:
      description: Payment due date (YYYY-MM-DD) or null.
      returned: always
      type: str
      sample: "2025-04-30"
    notes:
      description: Free-text notes or null.
      returned: always
      type: str
      sample: Q1 2025 support services.
    payment_date:
      description: Date payment was received (YYYY-MM-DD) or null.
      returned: always
      type: str
      sample: "2025-05-03"
    payment_amount:
      description: Amount actually received or null.
      returned: always
      type: float
      sample: 5445.00
    payment_reference:
      description: Payment reference string or null.
      returned: always
      type: str
      sample: WIRE-20250503-001
    payment_method:
      description: Payment method (bank, card, cash, other) or null.
      returned: always
      type: str
      sample: bank
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
# Constants
# ---------------------------------------------------------------------------

# Fields that are mutable after creation.
_UPDATE_FIELDS = (
    'status',
    'notes',
    'due_date',
    'vat_rate',
    'payment_date',
    'payment_amount',
    'payment_reference',
    'payment_method',
)

# line_items requires special handling (list comparison).

# Statuses that the server allows to be deleted.
_DELETABLE_STATUSES = ('draft', 'credit_note')


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _find_invoice(client, customer_id, invoice_number):
    """Return the invoice dict whose invoice_number matches, or None."""
    invoices = client.get('/customers/%d/invoices' % customer_id)
    for inv in invoices:
        if inv.get('invoice_number') == invoice_number:
            return inv
    return None


def _build_line_items(raw_items):
    """Normalise the module line_items param into API-ready dicts.

    Each item gets an explicit C(amount) computed as quantity * unit_price.
    Fields absent in the source dict are omitted from the output so the
    server applies its own defaults.
    """
    result = []
    for item in (raw_items or []):
        qty = float(item.get('quantity') or 1)
        unit_price = float(item['unit_price'])
        entry = {
            'description': item['description'],
            'quantity': qty,
            'unit_price': unit_price,
            'amount': round(qty * unit_price, 10),
        }
        if item.get('currency') is not None:
            entry['currency'] = item['currency']
        if item.get('is_manual') is not None:
            entry['is_manual'] = item['is_manual']
        result.append(entry)
    return result


def _build_create_body(p):
    """Build the POST body for a new invoice."""
    body = {
        'period_start': p['period_start'],
        'period_end': p['period_end'],
        'line_items': _build_line_items(p.get('line_items')),
    }
    for key in ('currency', 'due_date', 'notes'):
        if p.get(key) is not None:
            body[key] = p[key]
    if p.get('vat_rate') is not None:
        body['vat_rate'] = p['vat_rate']
    return body


def _build_update_body(p):
    """Build the PUT body for an existing invoice, omitting None values."""
    body = {}
    for key in _UPDATE_FIELDS:
        if p.get(key) is not None:
            body[key] = p[key]
    if p.get('line_items') is not None:
        body['line_items'] = _build_line_items(p['line_items'])
    return body


def _line_items_changed(existing_items, desired_raw):
    """Return True if the desired line items differ from the existing ones.

    Comparison is based on description, quantity, unit_price, and amount.
    Extra server-side fields (id, etc.) in existing_items are ignored.
    """
    desired = _build_line_items(desired_raw)
    if len(existing_items) != len(desired):
        return True
    for ex, de in zip(existing_items, desired):
        for field in ('description', 'quantity', 'unit_price', 'amount'):
            if ex.get(field) != de.get(field):
                return True
    return False


def _fields_changed(existing, p):
    """Return True if any mutable field in params differs from the existing invoice."""
    for key in _UPDATE_FIELDS:
        desired = p.get(key)
        if desired is None:
            continue
        current = existing.get(key)
        # Normalise None in existing to empty string for string fields so that
        # an explicit empty-string param correctly triggers a write.
        if isinstance(desired, str):
            current = current or ''
        if current != desired:
            return True

    # Check line_items separately.
    if p.get('line_items') is not None:
        if _line_items_changed(existing.get('line_items') or [], p['line_items']):
            return True

    return False


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        customer=dict(type='str', required=True),
        invoice_number=dict(type='str'),
        period_start=dict(type='str'),
        period_end=dict(type='str'),
        line_items=dict(type='list', elements='dict'),
        currency=dict(type='str'),
        vat_rate=dict(type='float'),
        due_date=dict(type='str'),
        notes=dict(type='str'),
        status=dict(type='str', choices=['draft', 'sent', 'paid']),
        payment_date=dict(type='str'),
        payment_amount=dict(type='float'),
        payment_reference=dict(type='str'),
        payment_method=dict(type='str', choices=['bank', 'card', 'cash', 'other']),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']
    invoice_number = p.get('invoice_number')
    client = WarmDeskClient.from_module(module)

    # Validate: state=absent requires invoice_number.
    if state == 'absent' and not invoice_number:
        module.fail_json(
            msg="invoice_number is required when state=absent."
        )

    # Validate: creating without invoice_number requires period_start,
    # period_end, and line_items.
    if state == 'present' and not invoice_number:
        missing = [
            f for f in ('period_start', 'period_end', 'line_items')
            if not p.get(f)
        ]
        if missing:
            module.fail_json(
                msg=(
                    "The following parameters are required when creating a new "
                    "invoice (invoice_number not provided): %s"
                ) % ', '.join(missing)
            )
        module.warn(
            "invoice_number not provided — a new invoice will be created on "
            "every run. Supply invoice_number for idempotent behaviour."
        )

    try:
        customer_id = resolve_customer_id(client, p['customer'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            existing = _find_invoice(client, customer_id, invoice_number)
            if existing is None:
                module.exit_json(changed=False, invoice=None)

            inv_status = existing.get('status', '')
            if inv_status not in _DELETABLE_STATUSES:
                module.fail_json(
                    msg=(
                        "Invoice %s cannot be deleted because its status is "
                        "'%s'. Only draft and credit_note invoices may be "
                        "deleted."
                    ) % (invoice_number, inv_status)
                )

            if not module.check_mode:
                client.delete(
                    '/customers/%d/invoices/%d' % (customer_id, existing['id'])
                )
            module.exit_json(changed=True, invoice=None)

        # ------------------------------------------------------------------ #
        # state=present, invoice_number given — idempotent find-or-fail       #
        # ------------------------------------------------------------------ #
        if invoice_number:
            existing = _find_invoice(client, customer_id, invoice_number)
            if existing is None:
                module.fail_json(
                    msg=(
                        "Invoice '%s' was not found for customer '%s'. "
                        "Invoice numbers are assigned by the server — you "
                        "cannot create an invoice with a specific number. "
                        "Omit invoice_number to create a new invoice."
                    ) % (invoice_number, p['customer'])
                )

            if not _fields_changed(existing, p):
                module.exit_json(changed=False, invoice=existing)

            if module.check_mode:
                module.exit_json(changed=True, invoice=existing)

            updated = client.put(
                '/customers/%d/invoices/%d' % (customer_id, existing['id']),
                _build_update_body(p),
            )
            module.exit_json(changed=True, invoice=updated)

        # ------------------------------------------------------------------ #
        # state=present, no invoice_number — always create                    #
        # ------------------------------------------------------------------ #
        if module.check_mode:
            module.exit_json(changed=True, invoice=None)

        invoice = client.post(
            '/customers/%d/invoices' % customer_id,
            _build_create_body(p),
        )
        module.exit_json(changed=True, invoice=invoice)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
