# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: card
short_description: Look up WarmDesk cards by their card reference (e.g. EDA-42)
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Fetches one or more WarmDesk card objects by their card reference string
    (e.g. C(EDA-42)) and returns them as a list of dicts.
  - Requires the C(project) keyword argument specifying the project slug.
  - Each reference is resolved by scanning the project's columns and cards via
    C(GET /api/v1/projects/:slug), then fetching the full card detail via
    C(GET /api/v1/projects/:slug/cards/:id).
  - Raises C(AnsibleError) when a requested card reference is not found in the
    project.
notes:
  - Card references are of the form C(<KEY_PREFIX>-<number>), e.g. C(EDA-42).
    The key prefix is the project's C(key_prefix) field (uppercase letters and
    digits).
  - Card references are case-sensitive; use the exact prefix as shown in the UI.
  - Closed (archived) cards are included in the search because the project
    endpoint returns all cards regardless of their state.
  - Pass C(wantlist=true) in your task to always receive a list even when
    looking up a single card.

options:
  _terms:
    description:
      - One or more card references to look up (e.g. C(EDA-42), C(PLAT-7)).
    required: true
    type: list
    elements: str

  project:
    description:
      - Slug of the project to search in (e.g. C(eda-00)).
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
# Look up a single card by reference
# ---------------------------------------------------------------------------
- name: Fetch card EDA-42
  ansible.builtin.set_fact:
    card: "{{ lookup('ansilabnl.warmdesk.card', 'EDA-42',
                     project='eda-00',
                     url='https://warmdesk.example.com',
                     api_key=vault_api_key) }}"

- name: Show card title and status
  ansible.builtin.debug:
    msg: "{{ card.card_ref }}: {{ card.title }} (closed={{ card.closed }})"

# ---------------------------------------------------------------------------
# Look up multiple cards from the same project
# ---------------------------------------------------------------------------
- name: Fetch several cards from the platform project
  ansible.builtin.set_fact:
    cards: "{{ lookup('ansilabnl.warmdesk.card',
                      'PLAT-1', 'PLAT-5', 'PLAT-12',
                      project='platform',
                      url='https://warmdesk.example.com',
                      api_key=vault_api_key,
                      wantlist=true) }}"

- name: Show card titles
  ansible.builtin.debug:
    msg: "{{ item.card_ref }}: {{ item.title }}"
  loop: "{{ cards }}"

# ---------------------------------------------------------------------------
# Use the card ID in a subsequent API call (via uri module)
# ---------------------------------------------------------------------------
- name: Get card details for a release card
  ansible.builtin.set_fact:
    release_card: "{{ lookup('ansilabnl.warmdesk.card', release_card_ref,
                             project=release_project,
                             url=warmdesk_url, api_key=warmdesk_key) }}"

- name: Post a comment on the release card
  ansible.builtin.uri:
    url: "{{ warmdesk_url }}/api/v1/projects/{{ release_project }}/cards/{{ release_card.id }}/comments"
    method: POST
    headers:
      X-API-Key: "{{ warmdesk_key }}"
    body_format: json
    body:
      body: "Release pipeline completed successfully."
"""

RETURN = r"""
_list:
  description:
    - List of card dicts, one per requested card reference, in input order.
    - Each dict is the full card detail object returned by
      C(GET /api/v1/projects/:slug/cards/:id).
  type: list
  elements: dict
  contains:
    id:
      description: Numeric card ID.
      type: int
      sample: 123
    card_number:
      description: Sequential card number within the project.
      type: int
      sample: 42
    card_ref:
      description: Full card reference string (key_prefix + card_number).
      type: str
      sample: EDA-42
    title:
      description: Card title.
      type: str
      sample: Implement OAuth2 login
    description:
      description: Card description (Markdown).
      type: str
      sample: "Add support for OAuth2 SSO via Google and GitHub."
    closed:
      description: Whether the card is closed/archived.
      type: bool
      sample: false
    project_id:
      description: Numeric ID of the owning project.
      type: int
      sample: 3
    column_id:
      description: Numeric ID of the column the card is currently in.
      type: int
      sample: 8
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2025-02-14T11:30:00Z"
"""

from ansible.errors import AnsibleError
from ansible.plugins.lookup import LookupBase
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
)
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_resolve import (
    find_card_by_number,
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
        project_slug = kwargs.get('project') or ''

        if not url:
            raise AnsibleError(
                'card lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'card lookup requires authentication: '
                'provide api_key, token, or username+password.'
            )

        if not project_slug:
            raise AnsibleError(
                'card lookup requires the "project" keyword '
                'argument (project slug).'
            )

        client = WarmDeskClient(
            url=url,
            username=username,
            password=password,
            token=token,
            api_key=api_key,
            validate_certs=validate_certs,
        )

        results = []
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        for card_ref in flat_terms:
            try:
                card = find_card_by_number(client, project_slug, card_ref)
            except WarmDeskAPIError as e:
                raise AnsibleError(
                    'WarmDesk API error %d looking up card "%s" in project "%s": %s'
                    % (e.status, card_ref, project_slug, e.message)
                )

            if card is None:
                raise AnsibleError(
                    'card lookup: card "%s" not found in project "%s".'
                    % (card_ref, project_slug)
                )

            results.append(card)

        return results
