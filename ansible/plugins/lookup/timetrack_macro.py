# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: timetrack_macro
short_description: Look up time-tracking macros from a user's WarmDesk macro library
version_added: "0.6.0"
author: "Ton Kersten (@tonk)"
description:
  - Fetches the current user's time-tracking macro library via
    C(GET /api/v1/time-entries/macro-library) and returns the named macros
    as a list of dicts.
  - Raises C(AnsibleError) when a requested macro name is not found in the
    library.
notes:
  - Requires that C(time_tracking_enabled) is set for the authenticated user's
    account.  The server returns 403 otherwise.
  - Pass C(wantlist=true) in your task to always receive a list even when
    looking up a single macro name.

options:
  _terms:
    description:
      - One or more macro names to look up.
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
# Look up a single macro and print its rows
# ---------------------------------------------------------------------------
- name: Get the Teaching block macro
  ansible.builtin.debug:
    msg: "{{ lookup('ansilabnl.warmdesk.timetrack_macro', 'Teaching block',
              url='https://warmdesk.example.com',
              api_key=vault_api_key) }}"

# ---------------------------------------------------------------------------
# Look up multiple macros and iterate over them
# ---------------------------------------------------------------------------
- name: Fetch several macros
  ansible.builtin.set_fact:
    my_macros: "{{ lookup('ansilabnl.warmdesk.timetrack_macro',
                           'Teaching block', 'DevOps day',
                           url=warmdesk_url,
                           api_key=warmdesk_key,
                           wantlist=true) }}"

- name: Show macro names and row counts
  ansible.builtin.debug:
    msg: "{{ item.name }} has {{ item.rows | length }} rows"
  loop: "{{ my_macros }}"

# ---------------------------------------------------------------------------
# Use a macro's rows as a template for a time-tracking task
# ---------------------------------------------------------------------------
- name: Get macro for use in another task
  ansible.builtin.set_fact:
    macro_rows: "{{ lookup('ansilabnl.warmdesk.timetrack_macro',
                            'Teaching block',
                            url=warmdesk_url, token=warmdesk_token).rows }}"
"""

RETURN = r"""
_list:
  description:
    - List of macro dicts, one for each name in C(_terms).  The order matches
      the order of the input terms.
  type: list
  elements: dict
  contains:
    id:
      description: Numeric macro ID within the library.
      type: int
      sample: 1
    name:
      description: Display name.
      type: str
      sample: Teaching block
    apply_days:
      description: Number of days the macro covers.
      type: int
      sample: 5
    alternating:
      description: Whether the macro uses alternating day patterns.
      type: bool
      sample: false
    rows:
      description: List of time-entry row dicts.
      type: list
      elements: dict
      contains:
        customer_id:
          description: Customer ID (null if not set).
          type: int
          sample: null
        project_id:
          description: Project ID (null if not set).
          type: int
          sample: 7
        description:
          description: Free-text description for the time entry.
          type: str
          sample: Teaching
        day1_minutes:
          description: Minutes logged on day 1 (empty string when unset).
          type: str
          sample: "360"
        day1_start:
          description: Start time on day 1 (HH:MM or empty string).
          type: str
          sample: "09:00"
        day1_end:
          description: End time on day 1 (HH:MM or empty string).
          type: str
          sample: "15:00"
        day2_minutes:
          description: Minutes logged on day 2 when alternating is enabled.
          type: str
          sample: "360"
        day2_start:
          description: Start time on day 2.
          type: str
          sample: ""
        day2_end:
          description: End time on day 2.
          type: str
          sample: ""
        day1_distance:
          description: Travel distance on day 1 (empty string when unset).
          type: str
          sample: ""
        day2_distance:
          description: Travel distance on day 2 (empty string when unset).
          type: str
          sample: ""
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
        validate_certs = kwargs.get('validate_certs', True)

        if not url:
            raise AnsibleError(
                'timetrack_macro lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'timetrack_macro lookup requires authentication: '
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
            result = client.get('/time-entries/macro-library')
        except WarmDeskAPIError as e:
            raise AnsibleError(
                'WarmDesk API error %d fetching macro library: %s' % (e.status, e.message)
            )

        raw_library = result.get('library') if result else None
        if not raw_library or not isinstance(raw_library.get('macros'), list):
            macros = []
        else:
            macros = raw_library['macros']

        macro_index = {}
        for m in macros:
            mname = m.get('name', '')
            if mname:
                macro_index[mname] = m

        # Flatten nested list terms (e.g. when the caller passes a list variable).
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        results = []
        for name in flat_terms:
            if name not in macro_index:
                raise AnsibleError(
                    'timetrack_macro lookup: macro "%s" not found in library.' % name
                )
            results.append(macro_index[name])

        return results
