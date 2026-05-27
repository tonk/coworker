# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: customer_member
short_description: Manage access to a WarmDesk customer
version_added: "0.1.0"
author:
  - Ton Kersten (@tonk)
description:
  - Add, update the role of, or remove a user's access to a WarmDesk customer.
  - Idempotent — adding a user who already has the correct role produces no change.
    Updating the role is also idempotent.
  - WarmDesk uses a strict allowlist model — a user who has no C(CustomerAccess)
    row cannot see the customer at all.  This module manages those rows.
  - The C(customer) parameter takes the customer B(name) (the display name used
    as the unique identifier).  Use the return value of
    M(ansilabnl.warmdesk.customer) to obtain it.
notes:
  - Only global admins and users who hold the C(admin) role for the specific
    customer may call the members endpoint.
  - A customer-admin cannot remove their own admin row via this module; the
    server silently preserves it to prevent self-lockout.
  - To add a user who is not yet a member the module resolves the username to an
    ID via C(GET /users).  The calling account must have at least authenticated
    access to the WarmDesk instance for that lookup to succeed.
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
options:
  customer:
    description:
      - Customer name (the unique display name, e.g. C(Acme Corp)).
      - Used as the idempotency key to look up the customer ID.
    type: str
    required: true
  username:
    description:
      - WarmDesk username of the user to grant, update, or revoke access for.
    type: str
    required: true
  role:
    description:
      - Access role to assign to the user.
      - C(member) — read-only visibility of the customer and its contracts.
      - C(admin) — can manage contracts and the customer's member list.
      - Ignored when C(state=absent).
    type: str
    choices: [member, admin]
    default: member
  state:
    description:
      - C(present) ensures the user has access to the customer with the
        specified C(role).  If the user already has access with a different
        role, the role is updated.
      - C(absent) removes the user's access to the customer.  If the user has
        no access row, no change is made.
    type: str
    choices: [present, absent]
    default: present
'''

EXAMPLES = r'''
- name: Grant a user read-only access to a customer
  ansilabnl.warmdesk.customer_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    customer: Acme Corp
    username: jdoe
    state: present

- name: Promote a member to customer-admin
  ansilabnl.warmdesk.customer_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    customer: Acme Corp
    username: jdoe
    role: admin
    state: present

- name: Ensure multiple users have access using a loop
  ansilabnl.warmdesk.customer_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    customer: "{{ item.customer }}"
    username: "{{ item.user }}"
    role: "{{ item.role | default('member') }}"
    state: present
  loop:
    - { customer: Acme Corp, user: alice, role: admin }
    - { customer: Acme Corp, user: bob }
    - { customer: Beta Ltd,  user: carol }

- name: Remove a user's access to a customer
  ansilabnl.warmdesk.customer_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    customer: Acme Corp
    username: former-employee
    state: absent

- name: Provision access from a register result
  ansilabnl.warmdesk.customer:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: New Client GmbH
    state: present
  register: new_customer

- name: Grant the account manager admin access
  ansilabnl.warmdesk.customer_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    customer: "{{ new_customer.customer.name }}"
    username: account-manager
    role: admin
    state: present
'''

RETURN = r'''
changed:
  description: Whether the module made any changes on the server.
  type: bool
  returned: always
  sample: true
member:
  description:
    - The access record for the user after the module ran.
    - C(null) when C(state=absent) and the access row did not exist or was
      just removed.
  type: dict
  returned: always
  contains:
    user_id:
      description: Numeric ID of the user.
      type: int
      sample: 5
    username:
      description: Login username.
      type: str
      sample: jdoe
    display_name:
      description: Friendly display name.
      type: str
      sample: Jane Doe
    email:
      description: User e-mail address.
      type: str
      sample: jdoe@example.com
    avatar_url:
      description: URL to the user's avatar image, if set.
      type: str
      sample: https://api.dicebear.com/9.x/avataaars/svg?seed=jdoe
    gravatar_url:
      description: Gravatar URL derived from the user's e-mail.
      type: str
      sample: https://www.gravatar.com/avatar/...
    role:
      description: The user's role for this customer.
      type: str
      sample: member
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
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


def _find_member(members, username):
    """Return the member dict with the given username, or None."""
    for m in members:
        if m.get('username') == username:
            return m
    return None


def _find_user_id(client, username):
    """Return the numeric user ID for *username*, or None."""
    users = client.get('/users')
    for u in users:
        if u.get('username') == username:
            return u['id']
    return None


def _members_payload(members):
    """Build the list of {user_id, role} dicts expected by the PUT endpoint."""
    return [{'user_id': m['user_id'], 'role': m['role']} for m in members]


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(
        customer=dict(type='str', required=True),
        username=dict(type='str', required=True),
        role=dict(type='str', default='member', choices=['member', 'admin']),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    params = module.params
    state = params['state']
    customer_name = params['customer']
    username = params['username']
    desired_role = params['role']

    try:
        client = WarmDeskClient.from_module(module)

        # ------------------------------------------------------------------
        # Resolve customer name → ID
        # ------------------------------------------------------------------
        customer = _find_customer(client, customer_name)
        if customer is None:
            module.fail_json(
                msg='Customer "%s" not found or not accessible.' % customer_name
            )

        customer_id = customer['id']

        # ------------------------------------------------------------------
        # Fetch current member list
        # ------------------------------------------------------------------
        try:
            members = client.get('/customers/%d/members' % customer_id)
        except WarmDeskAPIError as exc:
            if exc.status == 403:
                module.fail_json(
                    msg='Insufficient permissions to manage members of customer "%s".'
                    ' Requires global admin or customer-admin role.' % customer_name
                )
            raise

        existing = _find_member(members, username)

        # ------------------------------------------------------------------
        # state=absent
        # ------------------------------------------------------------------
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, member=None)

            if not module.check_mode:
                new_list = [m for m in members if m.get('username') != username]
                client.put(
                    '/customers/%d/members' % customer_id,
                    {'members': _members_payload(new_list)},
                )

            module.exit_json(changed=True, member=None)

        # ------------------------------------------------------------------
        # state=present — role already correct
        # ------------------------------------------------------------------
        if existing is not None and existing.get('role') == desired_role:
            module.exit_json(changed=False, member=existing)

        # ------------------------------------------------------------------
        # state=present — update role for existing member
        # ------------------------------------------------------------------
        if existing is not None:
            if module.check_mode:
                updated = dict(existing)
                updated['role'] = desired_role
                module.exit_json(changed=True, member=updated)

            new_list = []
            for m in members:
                if m.get('username') == username:
                    new_list.append({'user_id': m['user_id'], 'role': desired_role})
                else:
                    new_list.append({'user_id': m['user_id'], 'role': m['role']})

            client.put(
                '/customers/%d/members' % customer_id,
                {'members': new_list},
            )

            updated = dict(existing)
            updated['role'] = desired_role
            module.exit_json(changed=True, member=updated)

        # ------------------------------------------------------------------
        # state=present — add new member (user not yet in list)
        # ------------------------------------------------------------------
        user_id = _find_user_id(client, username)
        if user_id is None:
            module.fail_json(
                msg='User "%s" not found in this WarmDesk instance.' % username
            )

        if module.check_mode:
            module.exit_json(
                changed=True,
                member=dict(
                    user_id=user_id,
                    username=username,
                    display_name=None,
                    email=None,
                    avatar_url=None,
                    gravatar_url=None,
                    role=desired_role,
                ),
            )

        new_list = _members_payload(members)
        new_list.append({'user_id': user_id, 'role': desired_role})

        client.put(
            '/customers/%d/members' % customer_id,
            {'members': new_list},
        )

        # Re-fetch the member list to return the full member object.
        refreshed = client.get('/customers/%d/members' % customer_id)
        new_member = _find_member(refreshed, username)
        if new_member is None:
            # Fallback if the server silently rejected the row (e.g. self-lockout).
            new_member = dict(user_id=user_id, username=username, role=desired_role)

        module.exit_json(changed=True, member=new_member)

    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='WarmDesk API error (HTTP %s): %s' % (exc.status, exc.message)
        )


def main():
    run_module()


if __name__ == '__main__':
    main()
