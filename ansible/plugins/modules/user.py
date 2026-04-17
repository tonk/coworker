# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: user
short_description: Manage WarmDesk user accounts
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete WarmDesk user accounts via the admin API.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires admin credentials (C(global_role=admin)) or a token obtained from
    an admin account.
  - Password changes are always applied on an update because the API returns a
    hashed password and there is no way to detect whether the plaintext has
    changed.  Supply C(password) only when you explicitly want to set/rotate it.
  - MFA can be forcibly disabled by setting C(mfa_disable=true).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  username:
    description:
      - Login name for the user.  Used as the idempotency key.
    type: str
    required: true

  email:
    description:
      - E-mail address.  Required when C(state=present) and the user does not
        yet exist.
    type: str

  password:
    description:
      - Plaintext password.
      - Required when creating a new account (C(state=present) and user absent).
      - When updating an existing account, setting this will always trigger a
        C(changed=true) and push the new password to the API.
      - Treated as sensitive — not echoed in output.
    type: str
    no_log: true

  first_name:
    description: User's given name.
    type: str

  last_name:
    description: User's family name.
    type: str

  display_name:
    description:
      - Friendly display name shown in the UI.  Defaults to
        C(first_name + last_name) server-side when omitted.
    type: str

  global_role:
    description:
      - System-wide role for the account.
    type: str
    choices: [admin, user, viewer, metrics]
    default: user

  is_active:
    description:
      - Whether the account is enabled.  Set to C(false) to soft-disable a
        user without deleting the account.
    type: bool

  locale:
    description:
      - Preferred UI locale (e.g. C(en), C(nl), C(de), C(fr), C(es)).
    type: str

  timezone:
    description:
      - IANA timezone string (e.g. C(Europe/Amsterdam)).
    type: str

  date_time_format:
    description:
      - Preferred date/time display format string understood by the frontend.
    type: str

  accent_color:
    description:
      - Hex colour string for the user's UI accent colour (e.g. C(#4f46e5)).
    type: str

  sidebar_position:
    description:
      - Sidebar placement preference.  Accepted values depend on the WarmDesk
        version (commonly C(left) or C(right)).
    type: str

  mfa_disable:
    description:
      - When C(true) and the user exists, forcibly disable MFA for the account.
      - This is a one-shot action; the module always reports C(changed=true)
        when it executes the disable call.
    type: bool
    default: false

  customer_roles:
    description:
      - Dict mapping customer B(name) to access role (C(member) or C(admin)).
      - WarmDesk uses a strict allowlist model — only explicitly assigned
        customers are visible to the user.
      - This parameter performs a B(full sync) — customers not listed are
        removed, customers listed are added or updated.
      - Pass an empty dict C({}) to remove all customer assignments (the user
        will see no customers).
      - Omit (or set to C(null)) to leave customer assignments unchanged.
      - Customer names that cannot be resolved are skipped with a warning.
    type: dict

  state:
    description:
      - C(present) ensures the account exists with the specified attributes.
      - C(absent) removes the account if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a new standard user
# ---------------------------------------------------------------------------
- name: Create user alice
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    username: alice
    email: alice@example.com
    password: "{{ vault_alice_password }}"
    first_name: Alice
    last_name: Wonderland
    global_role: user

# ---------------------------------------------------------------------------
# Create an admin account using an API key
# ---------------------------------------------------------------------------
- name: Ensure ops-bot admin account exists
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: ops-bot
    email: ops-bot@example.com
    password: "{{ vault_opsbot_password }}"
    display_name: Ops Bot
    global_role: admin
    state: present

# ---------------------------------------------------------------------------
# Update an existing user — change role and locale only
# ---------------------------------------------------------------------------
- name: Promote alice to admin and set locale
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: alice
    global_role: admin
    locale: nl
    timezone: Europe/Amsterdam

# ---------------------------------------------------------------------------
# Disable an account without deleting it
# ---------------------------------------------------------------------------
- name: Deactivate departed employee
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: bob
    is_active: false

# ---------------------------------------------------------------------------
# Forcibly disable MFA for a locked-out user
# ---------------------------------------------------------------------------
- name: Disable MFA for alice (admin reset)
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: alice
    mfa_disable: true

# ---------------------------------------------------------------------------
# Assign user to customers with specific roles
# ---------------------------------------------------------------------------
- name: Create alice and assign her to two customers
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: alice
    email: alice@example.com
    password: "{{ vault_alice_password }}"
    customer_roles:
      Acme Corp: admin
      Beta Ltd: member
    state: present

# ---------------------------------------------------------------------------
# Update customer assignments for an existing user
# ---------------------------------------------------------------------------
- name: Give bob admin access to Acme Corp, remove all other customer access
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: bob
    customer_roles:
      Acme Corp: admin

# ---------------------------------------------------------------------------
# Remove all customer assignments (user sees no customers)
# ---------------------------------------------------------------------------
- name: Strip all customer visibility from charlie
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: charlie
    customer_roles: {}

# ---------------------------------------------------------------------------
# Delete a user account
# ---------------------------------------------------------------------------
- name: Remove user bob
  ansilabnl.warmdesk.user:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    username: bob
    state: absent
"""

RETURN = r"""
user:
  description:
    - The user object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the user was deleted (or did not exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric user ID.
      returned: always
      type: int
      sample: 42
    username:
      description: Login name.
      returned: always
      type: str
      sample: alice
    email:
      description: E-mail address.
      returned: always
      type: str
      sample: alice@example.com
    first_name:
      description: Given name.
      returned: always
      type: str
      sample: Alice
    last_name:
      description: Family name.
      returned: always
      type: str
      sample: Wonderland
    display_name:
      description: Display name.
      returned: always
      type: str
      sample: Alice Wonderland
    global_role:
      description: System-wide role.
      returned: always
      type: str
      sample: user
    is_active:
      description: Whether the account is enabled.
      returned: always
      type: bool
      sample: true
    locale:
      description: Preferred locale.
      returned: when set
      type: str
      sample: nl
    timezone:
      description: IANA timezone.
      returned: when set
      type: str
      sample: Europe/Amsterdam
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2025-01-15T09:00:00Z"
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

def _find_user(client, username):
    """Return the user dict whose username matches, or None."""
    users = client.get('/admin/users')
    for u in users:
        if u.get('username') == username:
            return u
    return None


def _resolve_customer_roles(client, customer_roles_by_name):
    """Convert {customer_name: role} to ({customer_id (int): role}, [unresolved]).

    Roles that are not 'admin' or 'member' are coerced to 'member'.
    """
    if not customer_roles_by_name and customer_roles_by_name is not None:
        # Empty dict — caller wants to clear all assignments.
        return {}, []
    customers = client.get('/customers')
    by_name = {c['name']: c['id'] for c in customers}
    resolved = {}
    unresolved = []
    for name, role in customer_roles_by_name.items():
        if name in by_name:
            resolved[by_name[name]] = role if role in ('admin', 'member') else 'member'
        else:
            unresolved.append(name)
    return resolved, unresolved


def _customer_roles_changed(current_data, desired_by_id):
    """Return True when the desired customer assignments differ from current.

    current_data is the response from GET /admin/users/:id/customers:
      {"customer_ids": [...], "customer_roles": {"1": "admin", ...}}
    desired_by_id is {int_id: role}.
    """
    current_ids = set(int(x) for x in current_data.get('customer_ids') or [])
    current_roles = {
        int(k): v for k, v in (current_data.get('customer_roles') or {}).items()
    }
    desired_ids = set(desired_by_id.keys())
    if current_ids != desired_ids:
        return True
    for cid, role in desired_by_id.items():
        if current_roles.get(cid) != role:
            return True
    return False


def _apply_customer_roles(client, user_id, desired_by_id):
    """PUT the full customer assignment list for *user_id*."""
    ids = list(desired_by_id.keys())
    roles = {str(cid): role for cid, role in desired_by_id.items()}
    client.put('/admin/users/%d/customers' % user_id, {
        'customer_ids': ids,
        'customer_roles': roles,
    })


def _build_create_body(p):
    """Build the POST body from module params."""
    body = dict(
        username=p['username'],
        email=p['email'],
        password=p['password'],
        global_role=p.get('global_role') or 'user',
    )
    for optional in ('first_name', 'last_name', 'display_name'):
        if p.get(optional) is not None:
            body[optional] = p[optional]
    return body


def _build_update_body(p, existing, has_password):
    """Return (body, changed).

    Compares desired values against the existing user dict and builds a PUT
    body that only contains fields that differ.  Always includes a password
    when the caller supplied one (no comparison possible).
    """
    body = {}
    changed = False

    # Scalar fields that can be compared directly.
    comparable = {
        'global_role': 'global_role',
        'is_active': 'is_active',
        'first_name': 'first_name',
        'last_name': 'last_name',
        'display_name': 'display_name',
        'email': 'email',
        'locale': 'locale',
        'timezone': 'timezone',
        'date_time_format': 'date_time_format',
        'accent_color': 'accent_color',
        'sidebar_position': 'sidebar_position',
    }
    for param_key, api_key in comparable.items():
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            body[api_key] = desired
            changed = True

    # Password is always opaque — apply it unconditionally when given.
    if has_password:
        body['password'] = p['password']
        changed = True

    return body, changed


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        username=dict(type='str', required=True),
        email=dict(type='str'),
        password=dict(type='str', no_log=True),
        first_name=dict(type='str'),
        last_name=dict(type='str'),
        display_name=dict(type='str'),
        global_role=dict(
            type='str',
            default='user',
            choices=['admin', 'user', 'viewer', 'metrics'],
        ),
        is_active=dict(type='bool'),
        locale=dict(type='str'),
        timezone=dict(type='str'),
        date_time_format=dict(type='str'),
        accent_color=dict(type='str'),
        sidebar_position=dict(type='str'),
        mfa_disable=dict(type='bool', default=False),
        customer_roles=dict(type='dict'),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ('state', 'present', ('email',)),
        ],
    )

    p = module.params
    state = p['state']
    client = WarmDeskClient.from_module(module)

    try:
        existing = _find_user(client, p['username'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, user=None)
            if not module.check_mode:
                client.delete('/admin/users/%d' % existing['id'])
            module.exit_json(changed=True, user=None)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if existing is None:
            if not p.get('password'):
                module.fail_json(
                    msg='password is required when creating a new user.'
                )
            if module.check_mode:
                module.exit_json(changed=True, user=None)
            user = client.post('/admin/users', _build_create_body(p))
            # Optionally disable MFA right after creation (edge-case but
            # consistent with the idempotent contract).
            if p['mfa_disable']:
                user = client.post('/admin/users/%d/mfa/disable' % user['id'])
            if p.get('customer_roles') is not None:
                desired_by_id, unresolved = _resolve_customer_roles(client, p['customer_roles'])
                if unresolved:
                    module.warn('Unresolved customer names (skipped): %s' % ', '.join(unresolved))
                _apply_customer_roles(client, user['id'], desired_by_id)
            module.exit_json(changed=True, user=user)

        # ------------------------------------------------------------------ #
        # state=present — UPDATE                                               #
        # ------------------------------------------------------------------ #
        has_password = bool(p.get('password'))
        update_body, changed = _build_update_body(p, existing, has_password)

        mfa_changed = False
        if p['mfa_disable']:
            # We treat this as always-changed because we cannot inspect the
            # current MFA state reliably from the list endpoint.
            mfa_changed = True
            changed = True

        customers_changed = False
        desired_by_id = None
        if p.get('customer_roles') is not None:
            desired_by_id, unresolved = _resolve_customer_roles(client, p['customer_roles'])
            if unresolved:
                module.warn('Unresolved customer names (skipped): %s' % ', '.join(unresolved))
            current_cust = client.get('/admin/users/%d/customers' % existing['id'])
            if _customer_roles_changed(current_cust, desired_by_id):
                customers_changed = True
                changed = True

        if not changed:
            module.exit_json(changed=False, user=existing)

        if module.check_mode:
            module.exit_json(changed=True, user=existing)

        user = existing
        if update_body:
            user = client.put('/admin/users/%d' % existing['id'], update_body)
        if mfa_changed:
            user = client.post('/admin/users/%d/mfa/disable' % existing['id'])
        if customers_changed:
            _apply_customer_roles(client, existing['id'], desired_by_id)

        module.exit_json(changed=True, user=user)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
