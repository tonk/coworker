# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: user
short_description: Look up WarmDesk user accounts by username
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Fetches one or more WarmDesk user accounts and returns them as a list of
    dicts.
  - Looks up each username in the result of C(GET /api/v1/users) and returns
    the matching user dict.
  - Raises C(AnsibleError) when a requested username is not found.
notes:
  - Uses the non-admin C(/users) endpoint, so any authenticated user can call
    this lookup.  The admin endpoint (C(/admin/users)) returns additional
    fields and is used by the C(ansilab.warmdesk.user) module.
  - Pass C(wantlist=true) in your task to always receive a list even when
    looking up a single username.

options:
  _terms:
    description:
      - One or more usernames to look up.
    required: true
    type: list
    elements: str

  url:
    description:
      - Base URL of the WarmDesk instance (e.g. C(https://warmdesk.example.com)).
    type: str
    required: true
    env:
      - name: WARMDESK_URL

  token:
    description:
      - Pre-obtained JWT access token.  Takes priority over C(username)/C(password).
    type: str
    no_log: true
    env:
      - name: WARMDESK_TOKEN

  username:
    description:
      - WarmDesk login name for password-based authentication.
    type: str
    env:
      - name: WARMDESK_USERNAME

  password:
    description:
      - WarmDesk password for password-based authentication.
    type: str
    no_log: true
    env:
      - name: WARMDESK_PASSWORD

  api_key:
    description:
      - WarmDesk API key (C(X-API-Key) header).  Takes priority over all
        other auth methods.
    type: str
    no_log: true
    env:
      - name: WARMDESK_API_KEY

  validate_certs:
    description:
      - Whether to validate TLS certificates.
    type: bool
    default: true
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Look up a single user and print the result
# ---------------------------------------------------------------------------
- name: Get alice's user record
  ansible.builtin.debug:
    msg: "{{ lookup('ansilab.warmdesk.user', 'alice',
              url='https://warmdesk.example.com',
              api_key=vault_api_key) }}"

# ---------------------------------------------------------------------------
# Look up multiple users at once
# ---------------------------------------------------------------------------
- name: Fetch several users
  ansible.builtin.set_fact:
    wd_users: "{{ lookup('ansilab.warmdesk.user', 'alice', 'bob', 'carol',
                         url='https://warmdesk.example.com',
                         api_key=vault_api_key,
                         wantlist=true) }}"

- name: Show user IDs
  ansible.builtin.debug:
    msg: "{{ item.username }} → {{ item.id }}"
  loop: "{{ wd_users }}"

# ---------------------------------------------------------------------------
# Use in a with_items style loop via wantlist
# ---------------------------------------------------------------------------
- name: Print email addresses for a list of users
  ansible.builtin.debug:
    msg: "{{ item.email }}"
  loop: "{{ lookup('ansilab.warmdesk.user', team_members,
                   url=warmdesk_url, api_key=warmdesk_key,
                   wantlist=true) }}"
  vars:
    team_members:
      - alice
      - bob
"""

RETURN = r"""
_list:
  description:
    - List of user dicts, one for each username in C(_terms).  The order
      matches the order of the input terms.
  type: list
  elements: dict
  contains:
    id:
      description: Numeric user ID.
      type: int
      sample: 42
    username:
      description: Login name.
      type: str
      sample: alice
    email:
      description: E-mail address.
      type: str
      sample: alice@example.com
    display_name:
      description: Display name shown in the UI.
      type: str
      sample: Alice Wonderland
    global_role:
      description: System-wide role (admin, user, viewer, metrics).
      type: str
      sample: user
    is_active:
      description: Whether the account is enabled.
      type: bool
      sample: true
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2025-01-15T09:00:00Z"
"""

from ansible.errors import AnsibleError
from ansible.plugins.lookup import LookupBase
from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
)


class LookupModule(LookupBase):

    def run(self, terms, variables=None, **kwargs):
        self.set_options(var_options=variables, direct=kwargs)

        url = kwargs.get('url') or ''
        token = kwargs.get('token')
        username = kwargs.get('username')
        password = kwargs.get('password')
        api_key = kwargs.get('api_key')
        validate_certs = kwargs.get('validate_certs', True)

        if not url:
            raise AnsibleError(
                'user lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'user lookup requires authentication: '
                'provide api_key, token, or username+password.'
            )

        client = WarmDeskClient(
            url=url,
            username=username,
            password=password,
            token=token,
            api_key=api_key,
            validate_certs=validate_certs,
        )

        try:
            all_users = client.get('/users')
        except WarmDeskAPIError as e:
            raise AnsibleError(
                'WarmDesk API error %d fetching users: %s' % (e.status, e.message)
            )

        # Build a lookup index for O(1) access per username.
        user_index = {}
        for u in all_users:
            uname = u.get('username', '')
            if uname:
                user_index[uname] = u

        results = []
        # terms may contain nested lists if the caller passes a list variable.
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        for uname in flat_terms:
            if uname not in user_index:
                raise AnsibleError(
                    'user lookup: user "%s" not found.' % uname
                )
            results.append(user_index[uname])

        return results
