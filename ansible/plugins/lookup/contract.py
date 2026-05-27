# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: contract
short_description: Look up WarmDesk contracts by name within a customer
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Fetches one or more WarmDesk contract records by their display name and
    returns them as a list of dicts.
  - Requires the C(customer) keyword argument specifying the customer name.
    The customer name is resolved first (via C(GET /api/v1/customers)), then
    the contracts for that customer are fetched (via
    C(GET /api/v1/customers/:id/contracts)).
  - Raises C(AnsibleError) when the customer or any requested contract name is
    not found.
notes:
  - Contract names are scoped to a customer; the same contract name can appear
    under different customers without conflict.
  - Contract names are case-sensitive.
  - Pass C(wantlist=true) in your task to always receive a list even when
    looking up a single name.

options:
  _terms:
    description:
      - One or more contract names to look up within the specified customer.
    required: true
    type: list
    elements: str

  customer:
    description:
      - Display name of the customer that owns the contracts.
      - Required.
    type: str
    required: true

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
# Look up a single contract for a customer
# ---------------------------------------------------------------------------
- name: Get the 2025 SLA contract for Acme Corp
  ansible.builtin.set_fact:
    sla: "{{ lookup('ansilabnl.warmdesk.contract', 'SLA 2025',
                    customer='Acme Corp',
                    url='https://warmdesk.example.com',
                    api_key=vault_api_key) }}"

- name: Show contract details
  ansible.builtin.debug:
    msg: >-
      Contract {{ sla.name }} (id={{ sla.id }}) runs from
      {{ sla.start_date | default('N/A') }} to {{ sla.end_date | default('N/A') }}

# ---------------------------------------------------------------------------
# Look up multiple contracts for the same customer
# ---------------------------------------------------------------------------
- name: Fetch all active contracts for Globex
  ansible.builtin.set_fact:
    globex_contracts: "{{ lookup('ansilabnl.warmdesk.contract',
                                 'Support 2025', 'Development Q2',
                                 customer='Globex',
                                 url='https://warmdesk.example.com',
                                 api_key=vault_api_key,
                                 wantlist=true) }}"

- name: Print contract IDs
  ansible.builtin.debug:
    msg: "{{ item.name }}: {{ item.id }}"
  loop: "{{ globex_contracts }}"

# ---------------------------------------------------------------------------
# Use with dynamic variables
# ---------------------------------------------------------------------------
- name: Look up contract by variable
  ansible.builtin.debug:
    msg: "{{ lookup('ansilabnl.warmdesk.contract', contract_name,
                   customer=client_name,
                   url=warmdesk_url, api_key=warmdesk_key) }}"
"""

RETURN = r"""
_list:
  description:
    - List of contract dicts, one per requested name, in input order.
  type: list
  elements: dict
  contains:
    id:
      description: Numeric contract ID.
      type: int
      sample: 5
    name:
      description: Contract display name.
      type: str
      sample: "SLA 2025"
    description:
      description: Optional free-text description.
      type: str
      sample: Annual service level agreement.
    customer_id:
      description: ID of the owning customer.
      type: int
      sample: 1
    start_date:
      description: Contract start date (ISO-8601 or null).
      type: str
      sample: "2025-01-01T00:00:00Z"
    end_date:
      description: Contract end date (ISO-8601 or null).
      type: str
      sample: "2025-12-31T23:59:59Z"
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2024-12-01T09:00:00Z"
"""

from ansible.errors import AnsibleError
from ansible.plugins.lookup import LookupBase
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
)
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_resolve import (
    resolve_customer_id,
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
        customer_name = kwargs.get('customer') or ''

        if not url:
            raise AnsibleError(
                'contract lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'contract lookup requires authentication: '
                'provide api_key, token, or username+password.'
            )

        if not customer_name:
            raise AnsibleError(
                'contract lookup requires the "customer" keyword '
                'argument (customer display name).'
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
            customer_id = resolve_customer_id(client, customer_name)
        except WarmDeskAPIError as e:
            if e.status == 404:
                raise AnsibleError(
                    'contract lookup: customer "%s" not found.'
                    % customer_name
                )
            raise AnsibleError(
                'WarmDesk API error %d resolving customer "%s": %s'
                % (e.status, customer_name, e.message)
            )

        try:
            all_contracts = client.get('/customers/%d/contracts' % customer_id)
        except WarmDeskAPIError as e:
            raise AnsibleError(
                'WarmDesk API error %d fetching contracts for customer "%s": %s'
                % (e.status, customer_name, e.message)
            )

        # Build a name index.
        contract_index = {}
        for c in all_contracts:
            cname = c.get('name', '')
            if cname:
                contract_index[cname] = c

        results = []
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        for name in flat_terms:
            if name not in contract_index:
                raise AnsibleError(
                    'contract lookup: contract "%s" not found under '
                    'customer "%s".' % (name, customer_name)
                )
            results.append(contract_index[name])

        return results
