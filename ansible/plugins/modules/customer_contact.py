# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: customer_contact
short_description: Manage WarmDesk customer contact persons
version_added: "0.6.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete contact persons that belong to a WarmDesk customer.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - The idempotency key is the combination of C(customer) (name) and C(name).
    Contact names must therefore be unique within a customer.
  - At most one contact per customer may be marked as C(is_primary=true).  The
    API enforces this; use a separate task to demote an existing primary contact
    before promoting another one if needed.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  customer:
    description:
      - Name of the customer that owns this contact person.
      - The customer must already exist; use M(ansilabnl.warmdesk.customer)
        to create it first.
    type: str
    required: true

  name:
    description:
      - Full name of the contact person.  Together with C(customer) this forms
        the idempotency key.
    type: str
    required: true

  department:
    description:
      - Department or team within the customer organisation.
      - Omit to leave unset or unchanged.
    type: str

  phone:
    description:
      - Phone number for the contact person.
      - Omit to leave unset or unchanged.
    type: str

  email:
    description:
      - E-mail address for the contact person.
      - Omit to leave unset or unchanged.
    type: str

  is_primary:
    description:
      - When C(true), marks this contact as the primary contact for the customer.
      - Only one contact per customer may be primary at a time.
      - Omit to leave the flag unchanged on update, or unset on create.
    type: bool

  state:
    description:
      - C(present) ensures the contact person exists with the specified attributes.
      - C(absent) removes the contact person if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a primary contact for a customer
# ---------------------------------------------------------------------------
- name: Add primary contact for Acme Corp
  ansilabnl.warmdesk.customer_contact:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Jane Smith
    department: IT Operations
    phone: "+31 20 123 4567"
    email: jane.smith@acme.example.com
    is_primary: true
    state: present

# ---------------------------------------------------------------------------
# Create a secondary contact (no is_primary flag)
# ---------------------------------------------------------------------------
- name: Add billing contact for Acme Corp
  ansilabnl.warmdesk.customer_contact:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Bob Johnson
    department: Finance
    email: bob.johnson@acme.example.com

# ---------------------------------------------------------------------------
# Update an existing contact's phone number and department
# ---------------------------------------------------------------------------
- name: Update Jane Smith's details
  ansilabnl.warmdesk.customer_contact:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Jane Smith
    department: Cloud & Infrastructure
    phone: "+31 20 999 8888"

# ---------------------------------------------------------------------------
# Provision contacts from a variable list (idempotent loop)
# ---------------------------------------------------------------------------
- name: Ensure all contacts are present
  ansilabnl.warmdesk.customer_contact:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: "{{ item.customer }}"
    name: "{{ item.name }}"
    department: "{{ item.department | default(omit) }}"
    phone: "{{ item.phone | default(omit) }}"
    email: "{{ item.email | default(omit) }}"
    is_primary: "{{ item.is_primary | default(omit) }}"
  loop: "{{ customer_contacts }}"

# ---------------------------------------------------------------------------
# Delete a contact that is no longer employed
# ---------------------------------------------------------------------------
- name: Remove departed contact from Acme Corp
  ansilabnl.warmdesk.customer_contact:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    customer: Acme Corp
    name: Bob Johnson
    state: absent
"""

RETURN = r"""
contact:
  description:
    - The contact person object as returned by the WarmDesk API after the
      operation.
    - C(null) when C(state=absent) and the contact was deleted (or did not
      exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric contact person ID.
      returned: always
      type: int
      sample: 42
    customer_id:
      description: Numeric ID of the owning customer.
      returned: always
      type: int
      sample: 7
    name:
      description: Full name of the contact person.
      returned: always
      type: str
      sample: Jane Smith
    department:
      description: Department or team within the customer organisation, or null.
      returned: always
      type: str
      sample: IT Operations
    phone:
      description: Phone number of the contact person, or null.
      returned: always
      type: str
      sample: "+31 20 123 4567"
    email:
      description: E-mail address of the contact person, or null.
      returned: always
      type: str
      sample: jane.smith@acme.example.com
    is_primary:
      description: Whether this contact is the primary contact for the customer.
      returned: always
      type: bool
      sample: true
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2025-06-01T08:00:00Z"
    updated_at:
      description: ISO-8601 last-update timestamp.
      returned: always
      type: str
      sample: "2025-06-19T12:30:00Z"
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

_TEXT_PARAMS = ('name', 'department', 'phone', 'email')
_BOOL_PARAMS = ('is_primary',)


def _find_contact(client, customer_id, name):
    """Return the contact dict whose name matches within *customer_id*, or None."""
    contacts = client.get('/customers/%d/contacts' % customer_id)
    for c in contacts:
        if c.get('name') == name:
            return c
    return None


def _build_body(p):
    """Build a create/update body from module params, omitting None values.

    Empty strings are sent through unchanged so the caller can explicitly
    clear a text field.  Only Python None is treated as "not provided".
    """
    body = {}
    for key in _TEXT_PARAMS:
        if p.get(key) is not None:
            body[key] = p[key]
    for key in _BOOL_PARAMS:
        if p.get(key) is not None:
            body[key] = p[key]
    return body


def _fields_changed(existing, p):
    """Return True if any mutable field in *p* differs from *existing*.

    Text comparison is direct string equality.  The API may return None/null
    for unset fields; we treat that as an empty string for comparison so that
    an explicit empty-string param triggers a write.  Boolean fields compare
    directly (None in *p* means "not specified — leave unchanged").
    """
    for key in _TEXT_PARAMS:
        desired = p.get(key)
        if desired is None:
            continue
        # Normalise None in existing → empty string for comparison.
        current = existing.get(key) or ''
        if current != desired:
            return True

    for key in _BOOL_PARAMS:
        desired = p.get(key)
        if desired is None:
            continue
        if existing.get(key) != desired:
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
        department=dict(type='str'),
        phone=dict(type='str'),
        email=dict(type='str'),
        is_primary=dict(type='bool'),
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
        # the correct behaviour — a contact cannot exist without a customer.
        customer_id = resolve_customer_id(client, p['customer'])

        existing = _find_contact(client, customer_id, p['name'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, contact=None)
            if not module.check_mode:
                client.delete(
                    '/customers/%d/contacts/%d' % (customer_id, existing['id'])
                )
            module.exit_json(changed=True, contact=None)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, contact=None)
            contact = client.post(
                '/customers/%d/contacts' % customer_id,
                _build_body(p),
            )
            module.exit_json(changed=True, contact=contact)

        # ------------------------------------------------------------------ #
        # state=present — UPDATE                                               #
        # ------------------------------------------------------------------ #
        if not _fields_changed(existing, p):
            module.exit_json(changed=False, contact=existing)

        if module.check_mode:
            module.exit_json(changed=True, contact=existing)

        contact = client.put(
            '/customers/%d/contacts/%d' % (customer_id, existing['id']),
            _build_body(p),
        )
        module.exit_json(changed=True, contact=contact)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
