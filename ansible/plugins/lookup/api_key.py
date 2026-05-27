# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: api_key
short_description: List WarmDesk API keys and show their owner
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Returns all personal API keys for the authenticated user, or all
    project-scoped API keys for a given project.
  - Each returned dict includes C(username) and C(display_name) fields
    resolved from the user list so you can see who owns each key.
  - When C(_terms) is provided the results are filtered to keys whose
    C(name) matches one of the given strings.  Omit C(_terms) (or pass
    an empty list) to return every key.
notes:
  - Personal keys are scoped to the authenticated user; only that user's
    own keys are visible without admin access.
  - Project keys are visible to any project member.
  - The plain-text key value is B(never) returned by the API after creation;
    only the C(key_prefix) (first 12 characters) is available.

options:
  _terms:
    description:
      - Zero or more key names to filter on.  When empty every key is returned.
    required: false
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

  project:
    description:
      - Project slug.  When set, project-scoped API keys for that project are
        returned instead of the authenticated user's personal keys.
    type: str
    required: false

  validate_certs:
    description:
      - Whether to validate TLS certificates.
    type: bool
    default: true
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# List all personal API keys for the authenticated user
# ---------------------------------------------------------------------------
- name: Show my API keys
  ansible.builtin.debug:
    msg: "{{ lookup('ansilabnl.warmdesk.api_key', '',
              url='https://warmdesk.example.com',
              api_key=vault_api_key) }}"

# ---------------------------------------------------------------------------
# List all project-scoped keys for a project
# ---------------------------------------------------------------------------
- name: Show project API keys with owner info
  ansible.builtin.debug:
    msg: "key {{ item.name }} ({{ item.key_prefix }}…) owned by {{ item.username }}"
  loop: "{{ lookup('ansilabnl.warmdesk.api_key', '',
                   url='https://warmdesk.example.com',
                   api_key=vault_api_key,
                   project='edge-data-analytics',
                   wantlist=true) }}"

# ---------------------------------------------------------------------------
# Look up a specific key by name
# ---------------------------------------------------------------------------
- name: Get the ci-bot key details
  ansible.builtin.set_fact:
    ci_key: "{{ lookup('ansilabnl.warmdesk.api_key', 'ci-bot',
                        url='https://warmdesk.example.com',
                        api_key=vault_api_key,
                        project='edge-data-analytics') }}"
"""

RETURN = r"""
_list:
  description:
    - List of API key dicts.  When C(_terms) is given, only keys whose
      C(name) matches are included; otherwise all keys are returned.
  type: list
  elements: dict
  contains:
    id:
      description: Numeric key ID.
      type: int
      sample: 7
    name:
      description: Human-readable label for the key.
      type: str
      sample: ci-bot
    key_prefix:
      description: First 12 characters of the plain-text key (safe to log).
      type: str
      sample: cwk_abc12345ef
    user_id:
      description: Numeric ID of the user who owns this key.
      type: int
      sample: 3
    username:
      description: Login name of the user who owns this key.
      type: str
      sample: alice
    display_name:
      description: Display name of the user who owns this key.
      type: str
      sample: Alice Wonderland
    project_id:
      description: Numeric project ID for project-scoped keys, or null for personal keys.
      type: int
      sample: 12
    last_used_at:
      description: ISO-8601 timestamp of the last time this key was used, or null.
      type: str
      sample: "2026-03-10T14:22:00Z"
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2026-01-15T09:00:00Z"
"""

from ansible.errors import AnsibleError
from ansible.plugins.lookup import LookupBase
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
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
        project = kwargs.get('project') or ''
        validate_certs = kwargs.get('validate_certs', True)

        if not url:
            raise AnsibleError(
                'api_key lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )
        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'api_key lookup requires authentication: '
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

        # Fetch the keys
        try:
            if project:
                keys = client.get('/projects/%s/api-keys' % project)
            else:
                keys = client.get('/auth/api-keys')
        except WarmDeskAPIError as e:
            raise AnsibleError(
                'WarmDesk API error %d fetching API keys: %s' % (e.status, e.message)
            )

        # Enrich with user info (username + display_name)
        try:
            all_users = client.get('/users')
            user_index = {u['id']: u for u in all_users if 'id' in u}
        except WarmDeskAPIError:
            user_index = {}

        for key in keys:
            uid = key.get('user_id')
            user = user_index.get(uid, {})
            key['username'] = user.get('username', '')
            key['display_name'] = user.get('display_name') or user.get('username', '')

        # Flatten terms (caller may pass a list variable)
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        # Filter to requested names; empty / blank terms → return all
        filter_names = [t for t in flat_terms if t]
        if filter_names:
            keys = [k for k in keys if k.get('name') in filter_names]
            missing = set(filter_names) - {k.get('name') for k in keys}
            if missing:
                raise AnsibleError(
                    'api_key lookup: key(s) not found: %s' % ', '.join(sorted(missing))
                )

        return keys
