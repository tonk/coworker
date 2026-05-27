# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: user_options
short_description: Manage WarmDesk user preferences and UI options
version_added: "0.3.8"
author: "Ton Kersten (@tonk)"
description:
  - Update per-user preferences such as locale, timezone, theme options,
    and UI layout settings via the admin API.
  - Idempotent — a second run with the same parameters produces no change.
  - When updating your own options, any authenticated credentials work.
  - To update another user's options, admin credentials
    (C(global_role=admin)) are required.
  - This module only touches preference fields; it does not change core
    account attributes (email, password, role, active state).
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
options:
  username:
    description:
      - Login name of the user whose options should be updated.
    type: str
    required: true
  theme:
    description:
      - UI theme preference.
    type: str
    choices: [light, dark, system]
  show_breadcrumbs:
    description:
      - Whether to show breadcrumb navigation in the UI.
    type: bool
  email_notifications:
    description:
      - Whether the user receives email notifications.
    type: bool
  locale:
    description:
      - Preferred UI locale (e.g. C(en), C(nl), C(de), C(fr), C(es)).
    type: str
  date_time_format:
    description:
      - Preferred date/time display format string (e.g. C(YYYY-MM-DD HH:mm)).
    type: str
  timezone:
    description:
      - IANA timezone string (e.g. C(Europe/Amsterdam)).
    type: str
  font:
    description:
      - UI font family name (e.g. C(Inter), C(Roboto), C(system)).
    type: str
  font_size:
    description:
      - UI base font size (e.g. C(14)).
    type: str
  sidebar_position:
    description:
      - Sidebar placement.
    type: str
    choices: [left, right]
  accent_color:
    description:
      - UI accent colour theme.
    type: str
    choices: [blue, red, green, orange]
  time_notation:
    description:
      - How time values are displayed in reports.
    type: str
    choices: [decimal, hhmm]
  week_start:
    description:
      - First day of the week in time tracking views.
    type: str
    choices: [monday, sunday]
notes:
  - Check mode is fully supported.
  - Only fields explicitly provided are compared and updated. Omitted fields
    are left unchanged on the server.
seealso:
  - module: ansilabnl.warmdesk.user
  - module: ansilabnl.warmdesk.user_access
    description: Use user_access to manage feature flags (board, chat, helpdesk, time tracking) and global role.
'''

EXAMPLES = r'''
- name: Set Dutch locale and Amsterdam timezone for alice
  ansilabnl.warmdesk.user_options:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    username: alice
    locale: nl
    timezone: Europe/Amsterdam
    date_time_format: DD-MM-YYYY HH:mm

- name: Switch alice to dark theme and compact font
  ansilabnl.warmdesk.user_options:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    username: alice
    font: Roboto
    font_size: "13"

- name: Set sidebar to right side with red accent
  ansilabnl.warmdesk.user_options:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    username: bob
    sidebar_position: right
    accent_color: red

- name: Configure time tracking display preferences
  ansilabnl.warmdesk.user_options:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    username: charlie
    time_notation: hhmm
    week_start: monday

- name: Switch to dark theme with breadcrumbs hidden
  ansilabnl.warmdesk.user_options:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    username: alice
    theme: dark
    show_breadcrumbs: false
'''

RETURN = r'''
changed:
  description: Whether any option was updated.
  type: bool
  returned: always
user:
  description: The full user object as returned by the WarmDesk API.
  type: dict
  returned: always
'''

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)


def _find_user_id(client, username):
    """Return the numeric user ID for *username* (or None)."""
    try:
        users = client.get('/users')
    except WarmDeskAPIError:
        return None
    for u in users:
        if u.get('username') == username:
            return u['id']
    return None


_OPTION_FIELDS = (
    'theme',
    'show_breadcrumbs',
    'email_notifications',
    'locale',
    'date_time_format',
    'timezone',
    'font',
    'font_size',
    'sidebar_position',
    'accent_color',
    'time_notation',
    'week_start',
)


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        username=dict(type='str', required=True),
        theme=dict(type='str', choices=['light', 'dark', 'system']),
        show_breadcrumbs=dict(type='bool'),
        email_notifications=dict(type='bool'),
        locale=dict(type='str'),
        date_time_format=dict(type='str'),
        timezone=dict(type='str'),
        font=dict(type='str'),
        font_size=dict(type='str'),
        sidebar_position=dict(type='str', choices=['left', 'right']),
        accent_color=dict(type='str', choices=['blue', 'red', 'green', 'orange']),
        time_notation=dict(type='str', choices=['decimal', 'hhmm']),
        week_start=dict(type='str', choices=['monday', 'sunday']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    client = WarmDeskClient.from_module(module)

    # Find target user ID by username.
    try:
        user_id = _find_user_id(client, p['username'])
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Error looking up user "%s": %s (HTTP %s)' % (
                p['username'], exc.message, exc.status)
        )

    if user_id is None:
        module.fail_json(msg='User "%s" not found.' % p['username'])

    # Determine the API path: self-service via /auth/me, or admin via /admin/users/:id.
    try:
        me = client.get('/auth/me')
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Failed to fetch current user: %s (HTTP %s)' % (
                exc.message, exc.status)
        )

    is_self = me.get('username') == p['username']
    if is_self:
        user = me
        endpoint = '/auth/me'
    else:
        try:
            user = client.get('/admin/users/%d' % user_id)
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to fetch user "%s" (admin required): %s (HTTP %s)' % (
                    p['username'], exc.message, exc.status)
            )
        endpoint = '/admin/users/%d' % user_id

    # Build update body from supplied options that differ from current values.
    body = {}
    changed = False

    for key in _OPTION_FIELDS:
        desired = p.get(key)
        if desired is None:
            continue
        if user.get(key) != desired:
            body[key] = desired
            changed = True

    if not changed:
        module.exit_json(changed=False, user=user)

    if module.check_mode:
        module.exit_json(changed=True, user=user)

    try:
        user = client.put(endpoint, body)
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Failed to update options for user "%s": %s (HTTP %s)' % (
                p['username'], exc.message, exc.status)
        )

    module.exit_json(changed=True, user=user)


def main():
    run_module()


if __name__ == '__main__':
    main()
