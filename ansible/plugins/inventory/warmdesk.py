# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
name: warmdesk
short_description: Ansible inventory from WarmDesk project members
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Generates an Ansible inventory by fetching project member lists from a
    WarmDesk instance.
  - Each project member becomes a host named C(<username>).
  - Hosts are added to a top-level C(warmdesk) group and to a per-project
    group named C(warmdesk_<project_slug>).
  - Optional role-based groups can be configured (e.g. map C(owner) role to
    an C(owners) group).
  - Host variables include C(warmdesk_user_id), C(warmdesk_email),
    C(warmdesk_role), C(warmdesk_project), and C(warmdesk_display_name).
  - Supports caching.

notes:
  - The config file must be named C(warmdesk.yml) or C(warmdesk.yaml) for the
    plugin to activate automatically.
  - Authentication priority: C(api_key) → C(token) → C(username) + C(password).
  - A single user who is a member of multiple projects appears once as a host
    but accumulates host variables from the last project processed.  Use
    C(compose) and C(keyed_groups) for advanced variable merging.
  - Set C(validate_certs: false) only in development environments.

options:
  plugin:
    description:
      - Must be set to C(ansilab.warmdesk.warmdesk) to activate the plugin.
    required: true
    type: str
    choices: [ansilab.warmdesk.warmdesk]

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
      - Whether to validate TLS certificates when connecting to WarmDesk.
    type: bool
    default: true

  project:
    description:
      - Single project slug to include.  Mutually exclusive with C(projects).
    type: str

  projects:
    description:
      - List of project slugs to include.  Mutually exclusive with C(project).
    type: list
    elements: str

  groups:
    description:
      - Optional mapping of WarmDesk role names to Ansible group names.
      - Keys are WarmDesk role values (C(owner), C(member), C(viewer)).
      - Values are the Ansible group names to add hosts with that role to.
    type: dict
    default: {}
    sample:
      owners: owner
      members: member
      viewers: viewer
"""

EXAMPLES = r"""
# warmdesk.yml — single project
# ----------------------------------------------------------------------
plugin: ansilab.warmdesk.warmdesk
url: https://warmdesk.example.com
api_key: "cwk_a1b2c3d4e5f6"
project: eda-00
validate_certs: true
groups:
  owners: owner
  members: member
  viewers: viewer

# warmdesk.yml — multiple projects with role groups
# ----------------------------------------------------------------------
plugin: ansilab.warmdesk.warmdesk
url: https://warmdesk.example.com
token: "{{ lookup('env', 'WARMDESK_TOKEN') }}"
projects:
  - eda-00
  - eda-01
  - platform
validate_certs: true
groups:
  project_owners: owner

# warmdesk.yml — username/password auth (development)
# ----------------------------------------------------------------------
plugin: ansilab.warmdesk.warmdesk
url: http://localhost:8080
username: admin
password: "{{ lookup('env', 'WD_ADMIN_PASSWORD') }}"
project: dev-project
validate_certs: false

# Using the inventory in a playbook
# ----------------------------------------------------------------------
# ansible-inventory -i warmdesk.yml --list
#
# - name: Deploy to all WarmDesk project owners
#   hosts: owners
#   roles:
#     - deploy_app
#
# - name: Show all WarmDesk hosts in the EDA project
#   hosts: warmdesk_eda-00
#   tasks:
#     - debug:
#         msg: >-
#           User {{ warmdesk_display_name }} ({{ warmdesk_email }})
#           has role {{ warmdesk_role }} in {{ warmdesk_project }}
"""

import os

from ansible.errors import AnsibleError, AnsibleParserError
from ansible.plugins.inventory import BaseInventoryPlugin, Constructable, Cacheable

from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
)

NAME = 'ansilab.warmdesk.warmdesk'


class InventoryModule(BaseInventoryPlugin, Constructable, Cacheable):

    NAME = 'ansilab.warmdesk.warmdesk'

    # ------------------------------------------------------------------
    # verify_file
    # ------------------------------------------------------------------

    def verify_file(self, path):
        """Return True only for files named warmdesk.yml or warmdesk.yaml."""
        if super(InventoryModule, self).verify_file(path):
            basename = os.path.basename(path)
            if basename in ('warmdesk.yml', 'warmdesk.yaml'):
                return True
        return False

    # ------------------------------------------------------------------
    # parse
    # ------------------------------------------------------------------

    def parse(self, inventory, loader, path, cache=True):
        super(InventoryModule, self).parse(inventory, loader, path, cache)

        # Read the config file.
        self._read_config_data(path)

        url = self.get_option('url') or ''
        token = self.get_option('token')
        username = self.get_option('username')
        password = self.get_option('password')
        api_key = self.get_option('api_key')
        validate_certs = self.get_option('validate_certs')
        role_group_map = self.get_option('groups') or {}

        # Resolve authentication from environment variables when not set in config.
        if not url:
            url = os.environ.get('WARMDESK_URL', '')
        if not token:
            token = os.environ.get('WARMDESK_TOKEN', '')
        if not username:
            username = os.environ.get('WARMDESK_USERNAME', '')
        if not password:
            password = os.environ.get('WARMDESK_PASSWORD', '')
        if not api_key:
            api_key = os.environ.get('WARMDESK_API_KEY', '')

        if not url:
            raise AnsibleParserError(
                'warmdesk inventory: "url" is required (set in config or '
                'WARMDESK_URL environment variable).'
            )

        if not (api_key or token or (username and password)):
            raise AnsibleParserError(
                'warmdesk inventory: authentication is required. '
                'Provide api_key, token, or username + password in the '
                'config file or via environment variables.'
            )

        # Collect project slugs.
        project_option = self.get_option('project')
        projects_option = self.get_option('projects') or []

        if project_option and projects_option:
            raise AnsibleParserError(
                'warmdesk inventory: specify either "project" or "projects", not both.'
            )

        if project_option:
            project_slugs = [project_option]
        elif projects_option:
            project_slugs = list(projects_option)
        else:
            raise AnsibleParserError(
                'warmdesk inventory: at least one project must be specified '
                'via "project" or "projects".'
            )

        # Build the API client.
        client = WarmDeskClient(
            url=url,
            username=username or None,
            password=password or None,
            token=token or None,
            api_key=api_key or None,
            validate_certs=validate_certs,
        )

        # Caching.
        cache_key = self.get_cache_key(path)
        use_cache = self.get_option('cache') if hasattr(self, 'get_option') else cache
        attempt_to_read_cache = use_cache and cache
        cache_needs_update = False

        if attempt_to_read_cache:
            try:
                all_project_members = self._cache[cache_key]
            except KeyError:
                cache_needs_update = True
                all_project_members = None
        else:
            all_project_members = None

        if all_project_members is None:
            all_project_members = self._fetch_members(client, project_slugs)
            if cache_needs_update:
                self._cache[cache_key] = all_project_members

        # Ensure the top-level warmdesk group exists.
        self.inventory.add_group('warmdesk')

        # Populate inventory from the fetched data.
        for slug, members in all_project_members.items():
            self._populate_project(slug, members, role_group_map)

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _fetch_members(self, client, project_slugs):
        """Return {slug: [member_dict, …]} for each slug."""
        result = {}
        for slug in project_slugs:
            try:
                members = client.get('/projects/%s/members' % slug)
            except WarmDeskAPIError as e:
                raise AnsibleError(
                    'warmdesk inventory: error fetching members for project '
                    '"%s" (HTTP %d): %s' % (slug, e.status, e.message)
                )
            result[slug] = members or []
        return result

    def _populate_project(self, project_slug, members, role_group_map):
        """Add each member as a host and set host variables."""
        project_group = 'warmdesk_%s' % project_slug
        # Sanitise the group name: replace hyphens with underscores so that
        # Ansible does not reject the group name.
        project_group = project_group.replace('-', '_')

        self.inventory.add_group(project_group)
        self.inventory.add_child('warmdesk', project_group)

        for member in members:
            user = member.get('user') or {}
            username = user.get('username') or member.get('username') or ''
            if not username:
                continue

            hostname = username

            # Add the host to the inventory.
            self.inventory.add_host(hostname, group=project_group)

            # Set host variables from the member/user data.
            self.inventory.set_variable(
                hostname, 'warmdesk_user_id',
                user.get('id') or member.get('user_id'),
            )
            self.inventory.set_variable(
                hostname, 'warmdesk_email',
                user.get('email', ''),
            )
            self.inventory.set_variable(
                hostname, 'warmdesk_role',
                member.get('role', ''),
            )
            self.inventory.set_variable(
                hostname, 'warmdesk_project',
                project_slug,
            )
            self.inventory.set_variable(
                hostname, 'warmdesk_display_name',
                user.get('display_name') or user.get('username', username),
            )

            # Also add to the top-level warmdesk group.
            self.inventory.add_child('warmdesk', hostname)

            # Role-based groups.
            member_role = member.get('role', '')
            if member_role and member_role in role_group_map:
                role_group = role_group_map[member_role]
                self.inventory.add_group(role_group)
                self.inventory.add_host(hostname, group=role_group)
