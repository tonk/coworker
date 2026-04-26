# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: from_vars
short_description: Provision WarmDesk resources from YAML variable files
version_added: "1.0.0"
author:
  - Ton Kersten (@tonk)
description:
  - |
      Reads one or more YAML files and creates or updates the WarmDesk resources
      described in them.
  - |
      Resources are always processed in dependency order regardless of how they
      are declared in the files, so playbook task order does not matter.
  - |
      Processing order: users → customers (with contracts and customer members) →
      projects (with columns, labels, project members, and cards) →
      groups (with members, customer access, and project access) →
      system settings.
  - |
      All resources are treated as C(state=present). To remove a resource use
      the dedicated module (for example M(ansilabnl.warmdesk.project) with
      C(state=absent)).
notes:
  - |
      Resources are identified by name (or username for users). Renaming a
      resource in the var file creates a new resource — the old one is not
      removed.
  - |
      Multiple files are merged. Lists under the same top-level key are
      concatenated; scalar values from later files override earlier ones.
  - |
      Paths are relative to the directory from which Ansible runs. Use
      C({{ playbook_dir }}) to make paths absolute.
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
options:
  var_files:
    description:
      - One or more YAML files to load.
      - |
          Each file may contain any combination of the supported top-level keys
          (C(users), C(customers), C(groups), C(projects), C(system_settings)).
      - |
          Use C({{ playbook_dir }}/vars/warmdesk.yml) for paths relative to the
          playbook.
    type: list
    elements: path
    required: true
'''

EXAMPLES = r'''
# vars/warmdesk.yml ─────────────────────────────────────────────────────────
#
# users:
#   - username: alice
#     email: alice@example.com
#     first_name: Alice
#     last_name: Smith
#     global_role: user
#     customer_roles:
#       Acme Corp: admin
#
# customers:
#   - name: Acme Corp
#     description: Our main customer
#     contracts:
#       - name: Support 2026
#         start_date: "2026-01-01"
#         end_date: "2026-12-31"
#     members:
#       - username: alice
#         role: admin
#
# groups:
#   - name: DevOps
#     members: [alice, bob]
#     customer_access:
#       - customer: Acme Corp
#         role: member
#     project_access:
#       - project: Infrastructure
#         role: member
#
# projects:
#   - name: Infrastructure
#     customer: Acme Corp
#     contract: Support 2026
#     color: "#3b82f6"
#     key_prefix: INF
#     columns:
#       - name: Backlog
#       - name: In Progress
#         wip_limit: 3
#       - name: Done
#     labels:
#       - name: Bug
#         color: "#ef4444"
#     members:
#       - username: alice
#         role: owner
#     cards:
#       - title: Setup CI/CD pipeline
#         column: Backlog
#         priority: high
#         assignee: alice
#
# system_settings:
#   company_name: Acme Corp
#   locale: en
# ────────────────────────────────────────────────────────────────────────────

- name: Provision all WarmDesk resources from a single vars file
  ansilabnl.warmdesk.from_vars:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    var_files:
      - "{{ playbook_dir }}/vars/warmdesk.yml"

- name: Provision from multiple files (merged in order)
  ansilabnl.warmdesk.from_vars:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    var_files:
      - "{{ playbook_dir }}/vars/users.yml"
      - "{{ playbook_dir }}/vars/customers.yml"
      - "{{ playbook_dir }}/vars/projects.yml"
'''

RETURN = r'''
changed:
  description: Whether any resource was created or updated.
  type: bool
  returned: always
results:
  description: Per-resource-type summary of actions taken.
  type: dict
  returned: always
  contains:
    users:
      description: Counts for user operations.
      type: dict
      sample: {created: 2, updated: 1, unchanged: 3}
    customers:
      description: Counts for customer operations.
      type: dict
    contracts:
      description: Counts for contract operations.
      type: dict
    customer_members:
      description: Counts for customer membership operations.
      type: dict
    projects:
      description: Counts for project operations.
      type: dict
    columns:
      description: Counts for board column operations.
      type: dict
    labels:
      description: Counts for label operations.
      type: dict
    project_members:
      description: Counts for project membership operations.
      type: dict
    cards:
      description: Counts for card operations.
      type: dict
    groups:
      description: Counts for group operations.
      type: dict
    system_settings:
      description: Counts for system setting operations.
      type: dict
'''

import os

try:
    import yaml as _yaml
    HAS_YAML = True
except ImportError:
    HAS_YAML = False

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# YAML loading
# ---------------------------------------------------------------------------

def _load_var_files(paths):
    """Load and merge one or more YAML files.

    List values under the same top-level key are concatenated so that
    splitting resources across multiple files is additive. Scalar values
    from later files override earlier ones.
    """
    merged = {}
    for path in paths:
        path = os.path.expandvars(os.path.expanduser(str(path)))
        with open(path, 'r') as fh:
            data = _yaml.safe_load(fh) or {}
        if not isinstance(data, dict):
            raise ValueError('var file %s must be a YAML mapping' % path)
        for key, value in data.items():
            if key in merged and isinstance(merged[key], list) and isinstance(value, list):
                merged[key] = merged[key] + value
            elif key in merged and isinstance(merged[key], dict) and isinstance(value, dict):
                merged[key].update(value)
            else:
                merged[key] = value
    return merged


# ---------------------------------------------------------------------------
# Result counters
# ---------------------------------------------------------------------------

_RESOURCE_TYPES = (
    'users', 'customers', 'contracts', 'customer_members',
    'projects', 'columns', 'labels', 'project_members', 'cards',
    'groups', 'system_settings',
)


class _Counts(object):
    __slots__ = ('created', 'updated', 'unchanged')

    def __init__(self):
        self.created = 0
        self.updated = 0
        self.unchanged = 0

    def to_dict(self):
        return {'created': self.created, 'updated': self.updated, 'unchanged': self.unchanged}


# ---------------------------------------------------------------------------
# Provisioner
# ---------------------------------------------------------------------------

class Provisioner(object):
    """Drives all WarmDesk resource provisioning for a single from_vars run."""

    def __init__(self, client, module):
        self.client = client
        self.module = module
        self.check_mode = module.check_mode
        self._counts = {r: _Counts() for r in _RESOURCE_TYPES}

        # Lazy-loaded list caches (refreshed after mutations)
        self._users_cache = None       # list[dict] from GET /admin/users
        self._customers_cache = None   # list[dict] from GET /customers
        self._projects_cache = None    # list[dict] from GET /projects

        # Per-project detail cache: slug → full project dict (with cards)
        self._project_detail = {}

    # ── Results ──────────────────────────────────────────────────────────────

    @property
    def changed(self):
        return any(c.created or c.updated for c in self._counts.values())

    def results(self):
        return {k: v.to_dict() for k, v in self._counts.items()}

    # ── List caches ───────────────────────────────────────────────────────────

    def _users(self):
        if self._users_cache is None:
            self._users_cache = self.client.get('/admin/users') or []
        return self._users_cache

    def _customers(self):
        if self._customers_cache is None:
            self._customers_cache = self.client.get('/customers') or []
        return self._customers_cache

    def _projects(self):
        if self._projects_cache is None:
            self._projects_cache = self.client.get('/projects') or []
        return self._projects_cache

    def _project_detail_cached(self, slug):
        if slug not in self._project_detail:
            self._project_detail[slug] = self.client.get('/projects/%s' % slug) or {}
        return self._project_detail[slug]

    def _invalidate_project_detail(self, slug):
        self._project_detail.pop(slug, None)

    # ── Name/ID lookups ───────────────────────────────────────────────────────

    def _find_user(self, username):
        return next((u for u in self._users() if u.get('username') == username), None)

    def _find_customer(self, name):
        return next((c for c in self._customers() if c.get('name') == name), None)

    def _find_project(self, name):
        """Find a project by display name or slug."""
        return next(
            (p for p in self._projects()
             if p.get('name') == name or p.get('slug') == name),
            None,
        )

    def _require_user_id(self, username):
        u = self._find_user(username)
        if u is None:
            raise WarmDeskAPIError(404, 'user not found: %s' % username)
        return u['id']

    def _require_customer_id(self, name):
        c = self._find_customer(name)
        if c is None:
            raise WarmDeskAPIError(404, 'customer not found: %s' % name)
        return c['id']

    # ── Users ─────────────────────────────────────────────────────────────────

    def ensure_user(self, defn):
        username = defn.get('username')
        if not username:
            self.module.warn('user entry missing username, skipping')
            return

        existing = self._find_user(username)
        scalar_fields = ('email', 'first_name', 'last_name', 'display_name',
                         'global_role', 'is_active', 'locale', 'timezone',
                         'date_time_format', 'accent_color', 'sidebar_position')

        if existing is None:
            if not self.check_mode:
                body = {'username': username}
                for f in scalar_fields:
                    if defn.get(f) is not None:
                        body[f] = defn[f]
                if defn.get('password'):
                    body['password'] = defn['password']
                created = self.client.post('/admin/users', body)
                self._users_cache.append(created)
            self._counts['users'].created += 1
            return

        update = {}
        for f in scalar_fields:
            if defn.get(f) is not None and defn[f] != existing.get(f):
                update[f] = defn[f]
        if defn.get('password'):
            update['password'] = defn['password']

        if update:
            if not self.check_mode:
                updated = self.client.put('/admin/users/%d' % existing['id'], update)
                for i, u in enumerate(self._users_cache):
                    if u['id'] == existing['id']:
                        self._users_cache[i] = updated
                        break
            self._counts['users'].updated += 1
        else:
            self._counts['users'].unchanged += 1

    def apply_user_customer_roles(self, defn):
        """Set a user's customer access (called after customers are provisioned)."""
        customer_roles = defn.get('customer_roles')
        if not customer_roles:
            return
        username = defn.get('username')
        existing = self._find_user(username)
        if existing is None:
            self.module.warn('user not found when applying customer_roles: %s' % username)
            return

        customer_ids = []
        roles_by_id = {}
        for cust_name, role in customer_roles.items():
            try:
                cid = self._require_customer_id(str(cust_name))
                customer_ids.append(cid)
                roles_by_id[cid] = role
            except WarmDeskAPIError:
                self.module.warn(
                    'customer "%s" not found when setting customer_roles for user %s'
                    % (cust_name, username)
                )

        if not self.check_mode:
            self.client.put(
                '/admin/users/%d/customers' % existing['id'],
                {'customer_ids': customer_ids, 'customer_roles': roles_by_id},
            )

    # ── Customers ─────────────────────────────────────────────────────────────

    def ensure_customer(self, defn):
        name = defn.get('name')
        if not name:
            self.module.warn('customer entry missing name, skipping')
            return None

        existing = self._find_customer(name)

        if existing is None:
            if not self.check_mode:
                body = {'name': name}
                for f in ('description', 'logo_url'):
                    if defn.get(f) is not None:
                        body[f] = defn[f]
                created = self.client.post('/customers', body)
                self._customers_cache.append(created)
                existing = created
            self._counts['customers'].created += 1
        else:
            update = {}
            for f in ('description', 'logo_url'):
                if defn.get(f) is not None and defn[f] != existing.get(f):
                    update[f] = defn[f]
            if update:
                if not self.check_mode:
                    updated = self.client.put('/customers/%d' % existing['id'], update)
                    for i, c in enumerate(self._customers_cache):
                        if c['id'] == existing['id']:
                            self._customers_cache[i] = updated
                            existing = updated
                            break
                self._counts['customers'].updated += 1
            else:
                self._counts['customers'].unchanged += 1

        if existing is not None and not self.check_mode and 'starred' in defn:
            try:
                if defn['starred']:
                    self.client.post('/customers/%d/favorite' % existing['id'])
                else:
                    self.client.delete('/customers/%d/favorite' % existing['id'])
            except WarmDeskAPIError:
                pass

        return existing

    def ensure_contract(self, customer_name, defn):
        name = defn.get('name')
        if not name:
            self.module.warn('contract entry under customer "%s" missing name, skipping' % customer_name)
            return

        try:
            customer_id = self._require_customer_id(customer_name)
        except WarmDeskAPIError:
            self.module.warn('customer "%s" not found for contract "%s"' % (customer_name, name))
            return

        contracts = self.client.get('/customers/%d/contracts' % customer_id) or []
        existing = next((c for c in contracts if c.get('name') == name), None)

        if existing is None:
            if not self.check_mode:
                body = {'name': name}
                for f in ('description', 'start_date', 'end_date'):
                    if defn.get(f) is not None:
                        body[f] = str(defn[f])
                self.client.post('/customers/%d/contracts' % customer_id, body)
            self._counts['contracts'].created += 1
            return

        update = {}
        for f in ('description', 'start_date', 'end_date'):
            if defn.get(f) is not None:
                desired = str(defn[f])
                current = str(existing.get(f) or '')
                if desired != current:
                    update[f] = desired
        if update:
            if not self.check_mode:
                self.client.put(
                    '/customers/%d/contracts/%d' % (customer_id, existing['id']),
                    update,
                )
            self._counts['contracts'].updated += 1
        else:
            self._counts['contracts'].unchanged += 1

    def ensure_customer_member(self, customer_name, defn):
        username = defn.get('username')
        role = defn.get('role', 'member')
        if not username:
            self.module.warn('customer member entry under "%s" missing username, skipping' % customer_name)
            return

        try:
            customer_id = self._require_customer_id(customer_name)
        except WarmDeskAPIError:
            self.module.warn('customer "%s" not found for member "%s"' % (customer_name, username))
            return

        members = self.client.get('/customers/%d/members' % customer_id) or []
        existing = next((m for m in members if m.get('username') == username), None)

        if existing is None:
            try:
                user_id = self._require_user_id(username)
            except WarmDeskAPIError:
                self.module.warn('user "%s" not found for customer member' % username)
                return
            if not self.check_mode:
                new_list = [{'user_id': m['user_id'], 'role': m['role']} for m in members]
                new_list.append({'user_id': user_id, 'role': role})
                self.client.put('/customers/%d/members' % customer_id, {'members': new_list})
            self._counts['customer_members'].created += 1
        elif existing.get('role') != role:
            if not self.check_mode:
                new_list = [
                    {'user_id': m['user_id'],
                     'role': role if m.get('username') == username else m['role']}
                    for m in members
                ]
                self.client.put('/customers/%d/members' % customer_id, {'members': new_list})
            self._counts['customer_members'].updated += 1
        else:
            self._counts['customer_members'].unchanged += 1

    # ── Projects ──────────────────────────────────────────────────────────────

    def ensure_project(self, defn):
        name = defn.get('name')
        if not name:
            self.module.warn('project entry missing name, skipping')
            return None

        customer_name = defn.get('customer')
        if not customer_name:
            self.module.warn('project "%s" has no customer, skipping' % name)
            return None

        try:
            customer_id = self._require_customer_id(customer_name)
        except WarmDeskAPIError:
            self.module.warn('customer "%s" not found for project "%s", skipping' % (customer_name, name))
            return None

        contract_id = None
        if defn.get('contract'):
            try:
                contracts = self.client.get('/customers/%d/contracts' % customer_id) or []
                c = next((c for c in contracts if c.get('name') == defn['contract']), None)
                if c:
                    contract_id = c['id']
                else:
                    self.module.warn(
                        'contract "%s" not found for project "%s"' % (defn['contract'], name)
                    )
            except WarmDeskAPIError:
                pass

        existing = self._find_project(name)

        if existing is None:
            if not self.check_mode:
                body = {'name': name, 'customer_id': customer_id}
                if contract_id:
                    body['contract_id'] = contract_id
                for f in ('description', 'color', 'board_type', 'key_prefix'):
                    if defn.get(f) is not None:
                        body[f] = defn[f]
                created = self.client.post('/projects', body)
                self._projects_cache.append(created)
                existing = created
            self._counts['projects'].created += 1
        else:
            # Always send customer_id/contract_id on update (backend requires it)
            resolved_customer_id = customer_id
            resolved_contract_id = (
                contract_id if defn.get('contract') is not None
                else existing.get('contract_id')
            )

            update = {
                'customer_id': resolved_customer_id,
                'contract_id': resolved_contract_id,
            }
            changed_fields = False
            for f in ('description', 'color'):
                if defn.get(f) is not None and defn[f] != existing.get(f):
                    update[f] = defn[f]
                    changed_fields = True
            if defn.get('is_archived') is not None and defn['is_archived'] != existing.get('is_archived'):
                update['is_archived'] = defn['is_archived']
                changed_fields = True
            if resolved_customer_id != existing.get('customer_id') or resolved_contract_id != existing.get('contract_id'):
                changed_fields = True

            if changed_fields:
                if not self.check_mode:
                    updated = self.client.put('/projects/%s' % existing['slug'], update)
                    for i, p in enumerate(self._projects_cache):
                        if p.get('id') == existing['id']:
                            self._projects_cache[i] = updated
                            existing = updated
                            break
                    self._invalidate_project_detail(existing.get('slug', ''))
                self._counts['projects'].updated += 1
            else:
                self._counts['projects'].unchanged += 1

        if existing is not None and not self.check_mode and 'starred' in defn:
            try:
                if defn['starred']:
                    self.client.post('/projects/%s/star' % existing['slug'])
                else:
                    self.client.delete('/projects/%s/star' % existing['slug'])
            except WarmDeskAPIError:
                pass

        return existing

    def ensure_column(self, slug, defn):
        name = defn.get('name')
        if not name:
            self.module.warn('column entry in project "%s" missing name, skipping' % slug)
            return

        columns = self.client.get('/projects/%s/columns' % slug) or []
        existing = next((c for c in columns if c.get('name') == name), None)

        if existing is None:
            if not self.check_mode:
                body = {'name': name}
                if defn.get('color') is not None:
                    body['color'] = defn['color']
                if defn.get('wip_limit') is not None:
                    body['wip_limit'] = defn['wip_limit']
                self.client.post('/projects/%s/columns' % slug, body)
                self._invalidate_project_detail(slug)
            self._counts['columns'].created += 1
            return

        update = {'name': name}
        changed = False
        if defn.get('color') is not None and defn['color'] != existing.get('color'):
            update['color'] = defn['color']
            changed = True
        if defn.get('wip_limit') is not None and defn['wip_limit'] != existing.get('wip_limit'):
            update['wip_limit'] = defn['wip_limit']
            changed = True
        if changed:
            if not self.check_mode:
                self.client.put('/projects/%s/columns/%d' % (slug, existing['id']), update)
            self._counts['columns'].updated += 1
        else:
            self._counts['columns'].unchanged += 1

    def ensure_label(self, slug, defn):
        name = defn.get('name')
        if not name:
            self.module.warn('label entry in project "%s" missing name, skipping' % slug)
            return

        labels = self.client.get('/projects/%s/labels' % slug) or []
        existing = next((l for l in labels if l.get('name') == name), None)

        if existing is None:
            if not self.check_mode:
                body = {'name': name}
                if defn.get('color') is not None:
                    body['color'] = defn['color']
                self.client.post('/projects/%s/labels' % slug, body)
            self._counts['labels'].created += 1
            return

        if defn.get('color') is not None and defn['color'] != existing.get('color'):
            if not self.check_mode:
                self.client.put(
                    '/projects/%s/labels/%d' % (slug, existing['id']),
                    {'name': name, 'color': defn['color']},
                )
            self._counts['labels'].updated += 1
        else:
            self._counts['labels'].unchanged += 1

    def ensure_project_member(self, slug, defn):
        username = defn.get('username')
        role = defn.get('role', 'member')
        if not username:
            self.module.warn('project member entry in "%s" missing username, skipping' % slug)
            return

        members = self.client.get('/projects/%s/members' % slug) or []
        existing = next(
            (m for m in members if m.get('user', {}).get('username') == username),
            None,
        )

        if existing is None:
            if not self.check_mode:
                self.client.post('/projects/%s/members' % slug, {'login': username, 'role': role})
            self._counts['project_members'].created += 1
        elif existing.get('role') != role:
            if not self.check_mode:
                self.client.put(
                    '/projects/%s/members/%d/role' % (slug, existing['user_id']),
                    {'role': role},
                )
            self._counts['project_members'].updated += 1
        else:
            self._counts['project_members'].unchanged += 1

    def ensure_card(self, slug, defn):
        title = defn.get('title')
        column_name = defn.get('column')
        if not title:
            self.module.warn('card entry in project "%s" missing title, skipping' % slug)
            return
        if not column_name:
            self.module.warn('card "%s" in project "%s" has no column, skipping' % (title, slug))
            return

        # Fetch full project detail (includes columns + cards); use cache
        project_data = self._project_detail_cached(slug)
        all_columns = project_data.get('columns', [])

        # Resolve target column
        col = next((c for c in all_columns if c.get('name') == column_name), None)
        if col is None:
            self.module.warn(
                'column "%s" not found in project "%s" for card "%s"' % (column_name, slug, title)
            )
            return
        col_id = col['id']

        # Find existing card by title (anywhere in the project)
        existing = None
        for c in all_columns:
            for card in c.get('cards', []):
                if card.get('title') == title:
                    existing = card
                    break
            if existing:
                break

        assignee_id = None
        if defn.get('assignee'):
            try:
                assignee_id = self._require_user_id(defn['assignee'])
            except WarmDeskAPIError:
                self.module.warn(
                    'assignee "%s" not found for card "%s"' % (defn['assignee'], title)
                )

        if existing is None:
            if not self.check_mode:
                body = {'title': title}
                for f in ('description', 'priority', 'start_date', 'due_date'):
                    if defn.get(f) is not None:
                        body[f] = defn[f]
                if assignee_id is not None:
                    body['assignee_id'] = assignee_id
                self.client.post('/projects/%s/columns/%d/cards' % (slug, col_id), body)
                self._invalidate_project_detail(slug)
            self._counts['cards'].created += 1
            return

        update = {}
        for f in ('description', 'priority', 'start_date', 'due_date'):
            if defn.get(f) is not None and defn[f] != existing.get(f):
                update[f] = defn[f]
        if assignee_id is not None and assignee_id != existing.get('assignee_id'):
            update['assignee_id'] = assignee_id
        if update:
            if not self.check_mode:
                self.client.put('/projects/%s/cards/%d' % (slug, existing['id']), update)
                self._invalidate_project_detail(slug)
            self._counts['cards'].updated += 1
        else:
            self._counts['cards'].unchanged += 1

    # ── Groups ────────────────────────────────────────────────────────────────

    def ensure_group(self, defn):
        name = defn.get('name')
        if not name:
            self.module.warn('group entry missing name, skipping')
            return

        groups = self.client.get('/admin/groups') or []
        stub = next((g for g in groups if g.get('name') == name), None)

        if stub is None:
            if not self.check_mode:
                body = {'name': name}
                if defn.get('description') is not None:
                    body['description'] = defn['description']
                stub = self.client.post('/admin/groups', body)
            self._counts['groups'].created += 1
        else:
            if defn.get('description') is not None and defn['description'] != stub.get('description'):
                if not self.check_mode:
                    self.client.patch('/admin/groups/%d' % stub['id'],
                                      {'description': defn['description']})
                self._counts['groups'].updated += 1
            else:
                self._counts['groups'].unchanged += 1

        if self.check_mode or stub is None:
            return

        # Fetch full group detail (members, access lists)
        group = self.client.get('/admin/groups/%d' % stub['id'])
        group_id = group['id']

        # Members
        if 'members' in defn:
            self._sync_group_members(group_id, group.get('members', []), defn['members'])

        # Customer access
        if 'customer_access' in defn:
            self._sync_group_customer_access(
                group_id, group.get('customer_access', []), defn['customer_access']
            )

        # Project access — accepts project names; resolves to slug
        if 'project_access' in defn:
            self._sync_group_project_access(
                group_id, group.get('project_access', []), defn['project_access']
            )

    def _sync_group_members(self, group_id, current_members, desired_usernames):
        current_by_username = {
            m['user']['username']: m['user_id']
            for m in current_members
            if m.get('user', {}).get('username')
        }
        desired_set = set(desired_usernames)
        current_set = set(current_by_username)

        for username in desired_set - current_set:
            try:
                user_id = self._require_user_id(username)
            except WarmDeskAPIError:
                self.module.warn('user "%s" not found for group member' % username)
                continue
            self.client.post('/admin/groups/%d/members' % group_id, {'user_id': user_id})

        for username in current_set - desired_set:
            self.client.delete(
                '/admin/groups/%d/members/%d' % (group_id, current_by_username[username])
            )

    def _sync_group_customer_access(self, group_id, current_access, desired_access):
        current_by_name = {
            ca['customer']['name']: {'customer_id': ca['customer_id'], 'role': ca['role']}
            for ca in current_access
            if ca.get('customer', {}).get('name')
        }
        desired_by_name = {
            item['customer']: item.get('role', 'member')
            for item in desired_access
            if item.get('customer')
        }

        for cust_name, desired_role in desired_by_name.items():
            current = current_by_name.get(cust_name)
            if current is None or current['role'] != desired_role:
                try:
                    cid = current['customer_id'] if current else self._require_customer_id(cust_name)
                except WarmDeskAPIError:
                    self.module.warn('customer "%s" not found for group access' % cust_name)
                    continue
                self.client.put('/admin/groups/%d/customers/%d' % (group_id, cid), {'role': desired_role})

        for cust_name, current in current_by_name.items():
            if cust_name not in desired_by_name:
                self.client.delete('/admin/groups/%d/customers/%d' % (group_id, current['customer_id']))

    def _sync_group_project_access(self, group_id, current_access, desired_access):
        current_by_slug = {
            pa['project']['slug']: {'project_id': pa['project_id'], 'role': pa['role']}
            for pa in current_access
            if pa.get('project', {}).get('slug')
        }

        # Build desired map: resolve project names to slugs
        desired_by_slug = {}
        for item in desired_access:
            proj_name = item.get('project')
            if not proj_name:
                continue
            p = self._find_project(proj_name)
            if p is None:
                self.module.warn('project "%s" not found for group access' % proj_name)
                continue
            desired_by_slug[p['slug']] = {'project_id': p['id'], 'role': item.get('role', 'member')}

        for slug, desired in desired_by_slug.items():
            current = current_by_slug.get(slug)
            if current is None or current['role'] != desired['role']:
                self.client.put(
                    '/admin/groups/%d/projects/%d' % (group_id, desired['project_id']),
                    {'role': desired['role']},
                )

        for slug, current in current_by_slug.items():
            if slug not in desired_by_slug:
                self.client.delete('/admin/groups/%d/projects/%d' % (group_id, current['project_id']))

    # ── System settings ───────────────────────────────────────────────────────

    def ensure_system_settings(self, desired):
        if not desired:
            return
        current = self.client.get('/admin/system') or {}
        update = {}
        for key, value in desired.items():
            if str(current.get(key, '')) != str(value):
                update[key] = value
        if update:
            if not self.check_mode:
                self.client.put('/admin/system', update)
            self._counts['system_settings'].updated += 1
        else:
            self._counts['system_settings'].unchanged += 1

    # ── Main entry point ──────────────────────────────────────────────────────

    def run(self, data):
        # Phase 1 — users (basic fields; customer_roles applied in phase 2b)
        for defn in data.get('users', []):
            self.ensure_user(defn)

        # Phase 2a — customers (with nested contracts and customer members)
        for defn in data.get('customers', []):
            self.ensure_customer(defn)
            for contract_def in defn.get('contracts', []):
                self.ensure_contract(defn['name'], contract_def)
            for member_def in defn.get('members', []):
                self.ensure_customer_member(defn['name'], member_def)

        # Phase 2b — user customer_roles (customers now exist)
        for defn in data.get('users', []):
            if defn.get('customer_roles'):
                self.apply_user_customer_roles(defn)

        # Phase 3 — projects (with nested columns, labels, members, cards)
        for defn in data.get('projects', []):
            project = self.ensure_project(defn)
            if project is None:
                # Project could not be created (missing customer, check_mode, etc.)
                # Still count what would be created so check_mode is informative.
                if self.check_mode:
                    self._counts['columns'].created += len(defn.get('columns', []))
                    self._counts['labels'].created += len(defn.get('labels', []))
                    self._counts['project_members'].created += len(defn.get('members', []))
                    self._counts['cards'].created += len(defn.get('cards', []))
                continue

            slug = project.get('slug')
            if not slug:
                continue

            for col_def in defn.get('columns', []):
                self.ensure_column(slug, col_def)
            for label_def in defn.get('labels', []):
                self.ensure_label(slug, label_def)
            for member_def in defn.get('members', []):
                self.ensure_project_member(slug, member_def)
            for card_def in defn.get('cards', []):
                self.ensure_card(slug, card_def)

        # Phase 4 — groups (after users, customers, projects)
        for defn in data.get('groups', []):
            self.ensure_group(defn)

        # Phase 5 — system settings (admin only)
        self.ensure_system_settings(data.get('system_settings', {}))


# ---------------------------------------------------------------------------
# Module entry point
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(
        var_files=dict(type='list', elements='path', required=True),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    if not HAS_YAML:
        module.fail_json(msg='PyYAML is required for this module (pip install pyyaml)')

    try:
        data = _load_var_files(module.params['var_files'])
    except (IOError, OSError) as exc:
        module.fail_json(msg='Cannot read var file: %s' % str(exc))
    except Exception as exc:
        module.fail_json(msg='Failed to parse YAML var file: %s' % str(exc))

    try:
        client = WarmDeskClient.from_module(module)
        prov = Provisioner(client, module)
        prov.run(data)
        module.exit_json(changed=prov.changed, results=prov.results())
    except WarmDeskAPIError as exc:
        module.fail_json(msg='WarmDesk API error (HTTP %s): %s' % (exc.status, exc.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
