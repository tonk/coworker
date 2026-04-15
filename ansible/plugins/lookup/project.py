# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: project
short_description: Look up WarmDesk projects by slug
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Fetches one or more WarmDesk project objects by their slug and returns
    them as a list of dicts.
  - Each slug is resolved by calling C(GET /api/v1/projects/:slug).
  - Raises C(AnsibleError) when a requested slug is not found or the caller
    does not have access to the project.
notes:
  - The response includes the full project object with columns, cards, labels,
    and members.  This can be a large payload for busy projects.
  - Pass C(wantlist=true) in your task to always receive a list even when
    looking up a single slug.

options:
  _terms:
    description:
      - One or more project slugs to fetch.
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
# Fetch a single project
# ---------------------------------------------------------------------------
- name: Get project metadata
  ansible.builtin.set_fact:
    project: "{{ lookup('ansilab.warmdesk.project', 'eda-00',
                        url='https://warmdesk.example.com',
                        api_key=vault_api_key) }}"

- name: Show project ID and key prefix
  ansible.builtin.debug:
    msg: "Project {{ project.name }} has ID {{ project.id }} and prefix {{ project.key_prefix }}"

# ---------------------------------------------------------------------------
# Fetch multiple projects and iterate over members
# ---------------------------------------------------------------------------
- name: Fetch all team projects
  ansible.builtin.set_fact:
    projects: "{{ lookup('ansilab.warmdesk.project',
                         'platform', 'eda-00', 'eda-01',
                         url='https://warmdesk.example.com',
                         api_key=vault_api_key,
                         wantlist=true) }}"

- name: Show member count per project
  ansible.builtin.debug:
    msg: "{{ item.name }}: {{ item.members | length }} members"
  loop: "{{ projects }}"

# ---------------------------------------------------------------------------
# Use project slug from a variable
# ---------------------------------------------------------------------------
- name: Look up a project by slug variable
  ansible.builtin.debug:
    msg: "{{ lookup('ansilab.warmdesk.project', my_project_slug,
                   url=warmdesk_url, api_key=warmdesk_key) }}"
"""

RETURN = r"""
_list:
  description:
    - List of project dicts, one per requested slug, in input order.
  type: list
  elements: dict
  contains:
    id:
      description: Numeric project ID.
      type: int
      sample: 3
    name:
      description: Project display name.
      type: str
      sample: EDA Platform
    slug:
      description: URL-safe project identifier.
      type: str
      sample: eda-00
    key_prefix:
      description: Card number prefix (e.g. EDA for cards EDA-1, EDA-2, …).
      type: str
      sample: EDA
    description:
      description: Project description text.
      type: str
      sample: Main development project for the EDA platform.
    color:
      description: Hex color used in the UI.
      type: str
      sample: "#4f46e5"
    members:
      description: List of project member objects.
      type: list
    columns:
      description: List of board column objects with their cards.
      type: list
    labels:
      description: List of label objects available in this project.
      type: list
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2025-01-10T10:00:00Z"
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
                'warmdesk_project lookup requires the "url" parameter '
                '(or WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleError(
                'warmdesk_project lookup requires authentication: '
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

        results = []
        flat_terms = []
        for term in terms:
            if isinstance(term, list):
                flat_terms.extend(term)
            else:
                flat_terms.append(term)

        for slug in flat_terms:
            try:
                project = client.get('/projects/%s' % slug)
            except WarmDeskAPIError as e:
                if e.status == 404:
                    raise AnsibleError(
                        'warmdesk_project lookup: project "%s" not found.' % slug
                    )
                raise AnsibleError(
                    'WarmDesk API error %d fetching project "%s": %s'
                    % (e.status, slug, e.message)
                )
            results.append(project)

        return results
