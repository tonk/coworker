# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: user_access
short_description: Manage WarmDesk user feature flags and global role
version_added: "0.4.0"
author: "Ton Kersten (@tonk)"
description:
  - Update feature access flags for an existing WarmDesk user account via the
    admin API.
  - Controls which features the user can access — board (Scrum/Kanban), team
    chat, time tracking, and helpdesk (tickets).
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires admin credentials (C(global_role=admin)) or a token obtained from
    an admin account.
  - This module only modifies feature flags; it does not create or delete user
    accounts (use the C(user) module for that).
  - Admin users bypass all feature flags except time tracking (the
    C(time_tracking_enabled) and C(time_tracking_viewer) flags still apply),
    so setting C(board_enabled=false) on an admin has no effect.
  - The username is used as the idempotency key.  If the user does not exist
    the module fails.

seealso:
  - module: ansilabnl.warmdesk.user
  - module: ansilabnl.warmdesk.user_options
    description: Use user_options to manage UI preferences (locale, theme, timezone, date format, etc.).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  username:
    description:
      - Login name of the user to update.  Used as the idempotency key.
    type: str
    required: true

  board_enabled:
    description:
      - Whether the user can access Scrum/Kanban boards and projects.
    type: bool

  chat_enabled:
    description:
      - Whether the user can access team chat and direct messages.
    type: bool

  helpdesk_enabled:
    description:
      - Whether the user can access the helpdesk (tickets).
    type: bool

  time_tracking_enabled:
    description:
      - Whether the user can log time entries and access time tracking
        features.
      - Unlike other feature flags, this flag also applies to admin users.
    type: bool

  time_tracking_viewer:
    description:
      - Whether the user can view time reports without being able to log time.
      - Only meaningful when C(time_tracking_enabled) is C(false).
    type: bool

  can_create_projects:
    description:
      - Whether the user can create new projects (in addition to global
        admins, who can always create projects).
    type: bool

  is_active:
    description:
      - Whether the account is enabled.  Set to C(false) to soft-disable a
        user without deleting the account.
    type: bool

  global_role:
    description:
      - System-wide role for the account.
    type: str
    choices: [admin, user, viewer, metrics, backup]
"""

EXAMPLES = r"""
- name: Grant board and chat access to a user
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    username: alice
    board_enabled: true
    chat_enabled: true

- name: Grant helpdesk access to a user
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    username: bob
    helpdesk_enabled: true

- name: Grant time tracking and viewer access
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    username: charlie
    time_tracking_enabled: true
    time_tracking_viewer: true

- name: Revoke all feature access (deactivate)
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ api_key }}"
    username: departed_user
    board_enabled: false
    chat_enabled: false
    helpdesk_enabled: false
    time_tracking_enabled: false

- name: Allow a user to create their own projects
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    username: alice
    can_create_projects: true

- name: Soft-disable a user account
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    username: bob
    is_active: false

- name: Promote user to admin
  ansilabnl.warmdesk.user_access:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_token: "{{ warmdesk_token }}"
    username: alice
    global_role: admin
"""

RETURN = r"""
user:
  description:
    - The user object as returned by the WarmDesk API after the update.
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
    board_enabled:
      description: Whether board access is enabled.
      returned: always
      type: bool
      sample: true
    chat_enabled:
      description: Whether chat access is enabled.
      returned: always
      type: bool
      sample: true
    helpdesk_enabled:
      description: Whether helpdesk access is enabled.
      returned: always
      type: bool
      sample: true
    time_tracking_enabled:
      description: Whether time tracking is enabled.
      returned: always
      type: bool
      sample: false
    time_tracking_viewer:
      description: Whether time tracking viewer is enabled.
      returned: always
      type: bool
      sample: false
    can_create_projects:
      description: Whether the user can create new projects.
      returned: always
      type: bool
      sample: false
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
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


def _find_user(client, username):
    users = client.get('/admin/users')
    for u in users:
        if u.get('username') == username:
            return u
    return None


def _build_body(p, existing):
    body = {}
    changed = False

    comparable = {
        'board_enabled': 'board_enabled',
        'chat_enabled': 'chat_enabled',
        'helpdesk_enabled': 'helpdesk_enabled',
        'time_tracking_enabled': 'time_tracking_enabled',
        'time_tracking_viewer': 'time_tracking_viewer',
        'can_create_projects': 'can_create_projects',
        'is_active': 'is_active',
        'global_role': 'global_role',
    }
    for param_key, api_key in comparable.items():
        desired = p.get(param_key)
        if desired is None:
            continue
        if existing.get(api_key) != desired:
            body[api_key] = desired
            changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        username=dict(type='str', required=True),
        board_enabled=dict(type='bool'),
        chat_enabled=dict(type='bool'),
        helpdesk_enabled=dict(type='bool'),
        time_tracking_enabled=dict(type='bool'),
        time_tracking_viewer=dict(type='bool'),
        can_create_projects=dict(type='bool'),
        is_active=dict(type='bool'),
        global_role=dict(
            type='str',
            choices=['admin', 'user', 'viewer', 'metrics', 'backup'],
        ),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    client = WarmDeskClient.from_module(module)

    try:
        existing = _find_user(client, p['username'])
        if existing is None:
            module.fail_json(
                msg='User "%s" not found. Use the user module to create it first.' % p['username']
            )

        update_body, changed = _build_body(p, existing)

        if not changed:
            module.exit_json(changed=False, user=existing)

        if module.check_mode:
            module.exit_json(changed=True, user=existing)

        user = client.put('/admin/users/%d' % existing['id'], update_body)
        module.exit_json(changed=True, user=user)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
