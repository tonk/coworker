# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: customer
short_description: Look up WarmDesk customers by name
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Fetches one or more WarmDesk customer records by their display name and
    returns them as a list of dicts.
  - Resolves each name by scanning C(GET /api/v1/customers) and returning the
    matching entry.
  - Raises C(AnsibleError) when a requested customer name is not found.
  - The returned dicts include C(project_count) and C(contract_count) metadata
    fields from the API's list response.
notes:
  - Customer names are case-sensitive; use the exact name as it appears in the
    WarmDesk UI.
  - Pass C(wantlist=true) in your task to always receive a list even when
    looking up a single name.

options:
  _terms:
    description:
      - One or more customer names to look up.
    required: true
    type: list
    elements: str

  url:
    description:
      - Base URL of the WarmDesk instance.
    type: str
    required: true
    env:
      - name: WARMDESK_URL

  token:
    description:
      - Pre-obtained JWT access token.
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
      - WarmDesk API key (C(X-API-Key) header).
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
# Look up a single customer
# ---------------------------------------------------------------------------
- name: Get customer record for Acme Corp
  ansible.builtin.set_fact:
    acme: "{{ lookup('ansilabnl.warmdesk.customer', 'Acme Corp',
                     url='https://warmdesk.example.com',
                     api_key=vault_api_key) }}"

- name: Show customer ID
  ansible.builtin.debug:
    msg: "Acme Corp has ID {{ acme.id }}"

# ---------------------------------------------------------------------------
# Look up multiple customers at once
# ---------------------------------------------------------------------------
- name: Fetch several customers
  ansible.builtin.set_fact:
    customers: "{{ lookup('ansilabnl.warmdesk.customer',
                          'Acme Corp', 'Globex', 'Initech',
                          url='https://warmdesk.example.com',
                          api_key=vault_api_key,
                          wantlist=true) }}"

- name: Show all customer IDs and project counts
  ansible.builtin.debug:
    msg: "{{ item.name }} (id={{ item.id }}) has {{ item.project_count }} projects"
  loop: "{{ customers }}"

# ---------------------------------------------------------------------------
# Use a customer name from inventory variables
# ---------------------------------------------------------------------------
- name: Look up customer by variable
  ansible.builtin.debug:
    msg: "{{ lookup('ansilabnl.warmdesk.customer', customer_name,
                   url=warmdesk_url, api_key=warmdesk_key) }}"
"""

RETURN = r"""
_list:
  description:
    - List of customer dicts, one per requested name, in input order.
  type: list
  elements: dict
  contains:
    id:
      description: Numeric customer ID.
      type: int
      sample: 1
    name:
      description: Customer display name.
      type: str
      sample: Acme Corp
    description:
      description: Optional free-text description.
      type: str
      sample: Our biggest enterprise client.
    logo_url:
      description: URL to the customer's logo image.
      type: str
      sample: https://warmdesk.example.com/uploads/logo_acme.png
    project_count:
      description: Number of projects linked to this customer.
      type: int
      sample: 4
    contract_count:
      description: Number of contracts linked to this customer.
      type: int
      sample: 2
    is_favorite:
      description: Whether the calling user has starred this customer.
      type: bool
      sample: false
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2024-11-01T12:00:00Z"
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
                'customer lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'customer lookup requires authentication: '
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
            all_customers = client.get('/customers')
        except WarmDeskAPIError as e:
            raise AnsibleError(
                'WarmDesk API error %d fetching customers: %s' % (e.status, e.message)
            )

        # Build a name → customer dict for O(1) lookups.
        customer_index = {}
        for c in all_customers:
            cname = c.get('name', '')
            if cname:
                customer_index[cname] = c

        results = []
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        for name in flat_terms:
            if name not in customer_index:
                raise AnsibleError(
                    'customer lookup: customer "%s" not found.' % name
                )
            results.append(customer_index[name])

        return results
