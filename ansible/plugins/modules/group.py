# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: group
short_description: Manage WarmDesk user groups
version_added: "0.3.0"
author:
  - Ton Kersten (@tonk)
description:
  - Create, update, or delete a WarmDesk user group.
  - Optionally manage the group's member list and its access rights on projects
    and customers in a declarative, idempotent fashion.
  - When C(members), C(project_access), or C(customer_access) are provided the
    module reconciles the live state to match the desired list exactly — adding
    missing entries and removing entries that are no longer wanted.
  - Omitting any of those lists leaves the corresponding live state untouched,
    so you can use the module purely to create or rename a group without
    affecting its membership.
  - Only global admins may create, modify, or delete groups.
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
options:
  name:
    description:
      - Group name.  Used as the idempotency key to locate an existing group.
      - Must be unique across all groups in the WarmDesk instance.
    type: str
    required: true
  description:
    description:
      - Free-text description of the group.
      - When omitted the description is left unchanged on an existing group.
    type: str
  members:
    description:
      - Declarative list of WarmDesk usernames that should belong to the group.
      - When provided the module ensures the group has B(exactly) these
        members — adding any that are missing and removing any that are
        not in the list.
      - Omit (or pass C(null)) to leave current membership untouched.
    type: list
    elements: str
  project_access:
    description:
      - Declarative list of projects the group should have access to, each
        entry specifying the project slug and the desired role.
      - When provided the module reconciles live project-access rows to match
        the list exactly — adding or updating missing/wrong entries and
        removing entries that are not in the list.
      - Omit (or pass C(null)) to leave current project access untouched.
    type: list
    elements: dict
    suboptions:
      project:
        description: Project slug (the URL-safe identifier, e.g. C(edge-analytics)).
        type: str
        required: true
      role:
        description: Role to grant the group on this project.
        type: str
        choices: [viewer, member, owner]
        default: member
  customer_access:
    description:
      - Declarative list of customers the group should have access to, each
        entry specifying the customer name and the desired role.
      - When provided the module reconciles live customer-access rows to match
        the list exactly.
      - Omit (or pass C(null)) to leave current customer access untouched.
    type: list
    elements: dict
    suboptions:
      customer:
        description:
          - Customer name (the display name used as unique identifier,
            e.g. C(Acme Corporation)).
        type: str
        required: true
      role:
        description: Role to grant the group on this customer.
        type: str
        choices: [viewer, member, owner]
        default: member
  state:
    description:
      - C(present) ensures the group exists and matches the desired
        configuration.
      - C(absent) deletes the group together with all its members and access
        rows.
    type: str
    choices: [present, absent]
    default: present
'''

EXAMPLES = r'''
- name: Create a group (no members or access yet)
  ansilabnl.warmdesk.group:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: Frontend Team
    description: Web and mobile engineers
    state: present

- name: Create a group with members and project access
  ansilabnl.warmdesk.group:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: DevOps Team
    description: Infrastructure and site reliability engineers
    members:
      - alice
      - bob
      - carol
    project_access:
      - project: infra-prod
        role: owner
      - project: api-platform
        role: member
    state: present

- name: Grant a group viewer access to a customer
  ansilabnl.warmdesk.group:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: Acme Stakeholders
    members:
      - account-manager
      - project-lead
    customer_access:
      - customer: Acme Corporation
        role: viewer
    state: present

- name: Update description without changing membership
  ansilabnl.warmdesk.group:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: Frontend Team
    description: Web, mobile, and design engineers
    state: present

- name: Remove a stale group
  ansilabnl.warmdesk.group:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: Temporary Group
    state: absent

- name: Provision groups from a variable list
  ansilabnl.warmdesk.group:
    warmdesk_url: "{{ warmdesk_url }}"
    warmdesk_token: "{{ vault_wd_token }}"
    name: "{{ item.name }}"
    description: "{{ item.description | default(omit) }}"
    members: "{{ item.members | default(omit) }}"
    project_access: "{{ item.project_access | default(omit) }}"
    customer_access: "{{ item.customer_access | default(omit) }}"
    state: present
  loop: "{{ warmdesk_groups }}"
'''

RETURN = r'''
changed:
  description: Whether the module made any changes on the server.
  type: bool
  returned: always
  sample: true
group:
  description:
    - The full group detail object after the module ran.
    - C(null) when C(state=absent) and the group did not exist or was just
      deleted.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric group ID.
      type: int
      sample: 3
    name:
      description: Group name.
      type: str
      sample: Frontend Team
    description:
      description: Group description.
      type: str
      sample: Web and mobile engineers
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2026-01-15T10:00:00Z"
    updated_at:
      description: ISO-8601 last-update timestamp.
      type: str
      sample: "2026-04-01T08:30:00Z"
    members:
      description: Current list of group members.
      type: list
      elements: dict
      contains:
        group_id:
          description: Numeric group ID.
          type: int
          sample: 3
        user_id:
          description: Numeric user ID.
          type: int
          sample: 7
        user:
          description: Embedded user object (id, username, email, display_name, gravatar_url).
          type: dict
    project_access:
      description: Current project access rows for this group.
      type: list
      elements: dict
      contains:
        group_id:
          description: Numeric group ID.
          type: int
          sample: 3
        project_id:
          description: Numeric project ID.
          type: int
          sample: 12
        role:
          description: Role granted to the group on this project.
          type: str
          sample: member
        project:
          description: Embedded project object (id, name, slug, color, …).
          type: dict
    customer_access:
      description: Current customer access rows for this group.
      type: list
      elements: dict
      contains:
        group_id:
          description: Numeric group ID.
          type: int
          sample: 3
        customer_id:
          description: Numeric customer ID.
          type: int
          sample: 2
        role:
          description: Role granted to the group on this customer.
          type: str
          sample: viewer
        customer:
          description: Embedded customer object (id, name, description, …).
          type: dict
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Helpers — resource lookup
# ---------------------------------------------------------------------------

def _find_group(client, name):
    """Return the GroupListItem dict whose name matches, or None."""
    groups = client.get('/admin/groups')
    for g in groups:
        if g.get('name') == name:
            return g
    return None


def _find_user_id(client, username):
    """Return the numeric user ID for *username*, or None."""
    users = client.get('/users')
    for u in users:
        if u.get('username') == username:
            return u['id']
    return None


def _find_project_id(client, slug):
    """Return the numeric project ID for the given *slug*, or None."""
    projects = client.get('/projects')
    for p in projects:
        if p.get('slug') == slug:
            return p['id']
    return None


def _find_customer_id(client, name):
    """Return the numeric customer ID for the given *name*, or None."""
    customers = client.get('/customers')
    for c in customers:
        if c.get('name') == name:
            return c['id']
    return None


# ---------------------------------------------------------------------------
# Helpers — reconciliation
# ---------------------------------------------------------------------------

def _sync_members(client, group_id, current_members, desired_usernames, module):
    """Add missing members and remove extra members to reach *desired_usernames*."""
    current_by_username = {}
    for m in current_members:
        uname = m.get('user', {}).get('username')
        if uname:
            current_by_username[uname] = m['user_id']

    desired_set = set(desired_usernames)
    current_set = set(current_by_username.keys())

    for username in desired_set - current_set:
        user_id = _find_user_id(client, username)
        if user_id is None:
            module.fail_json(msg='User "%s" not found in WarmDesk.' % username)
        client.post('/admin/groups/%d/members' % group_id, {'user_id': user_id})

    for username in current_set - desired_set:
        user_id = current_by_username[username]
        client.delete('/admin/groups/%d/members/%d' % (group_id, user_id))


def _sync_project_access(client, group_id, current_access, desired_access, module):
    """Set/update/remove project access rows to match *desired_access*."""
    current_by_slug = {}
    for pa in current_access:
        slug = pa.get('project', {}).get('slug')
        if slug:
            current_by_slug[slug] = {'project_id': pa['project_id'], 'role': pa['role']}

    desired_by_slug = {item['project']: item.get('role', 'member') for item in desired_access}

    for slug, desired_role in desired_by_slug.items():
        current = current_by_slug.get(slug)
        if current is None or current['role'] != desired_role:
            if current is not None:
                project_id = current['project_id']
            else:
                project_id = _find_project_id(client, slug)
                if project_id is None:
                    module.fail_json(msg='Project with slug "%s" not found.' % slug)
            client.put(
                '/admin/groups/%d/projects/%d' % (group_id, project_id),
                {'role': desired_role},
            )

    for slug, current in current_by_slug.items():
        if slug not in desired_by_slug:
            client.delete('/admin/groups/%d/projects/%d' % (group_id, current['project_id']))


def _sync_customer_access(client, group_id, current_access, desired_access, module):
    """Set/update/remove customer access rows to match *desired_access*."""
    current_by_name = {}
    for ca in current_access:
        name = ca.get('customer', {}).get('name')
        if name:
            current_by_name[name] = {'customer_id': ca['customer_id'], 'role': ca['role']}

    desired_by_name = {item['customer']: item.get('role', 'member') for item in desired_access}

    for name, desired_role in desired_by_name.items():
        current = current_by_name.get(name)
        if current is None or current['role'] != desired_role:
            if current is not None:
                customer_id = current['customer_id']
            else:
                customer_id = _find_customer_id(client, name)
                if customer_id is None:
                    module.fail_json(msg='Customer "%s" not found.' % name)
            client.put(
                '/admin/groups/%d/customers/%d' % (group_id, customer_id),
                {'role': desired_role},
            )

    for name, current in current_by_name.items():
        if name not in desired_by_name:
            client.delete('/admin/groups/%d/customers/%d' % (group_id, current['customer_id']))


def _members_changed(current_members, desired_usernames):
    """Return True if the member list needs to change."""
    current_set = set(
        m.get('user', {}).get('username')
        for m in current_members
        if m.get('user', {}).get('username')
    )
    return current_set != set(desired_usernames)


def _project_access_changed(current_access, desired_access):
    """Return True if project access needs to change."""
    current = {
        pa.get('project', {}).get('slug'): pa.get('role')
        for pa in current_access
        if pa.get('project', {}).get('slug')
    }
    desired = {item['project']: item.get('role', 'member') for item in desired_access}
    return current != desired


def _customer_access_changed(current_access, desired_access):
    """Return True if customer access needs to change."""
    current = {
        ca.get('customer', {}).get('name'): ca.get('role')
        for ca in current_access
        if ca.get('customer', {}).get('name')
    }
    desired = {item['customer']: item.get('role', 'member') for item in desired_access}
    return current != desired


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(
        name=dict(type='str', required=True),
        description=dict(type='str'),
        members=dict(type='list', elements='str'),
        project_access=dict(
            type='list',
            elements='dict',
            options=dict(
                project=dict(type='str', required=True),
                role=dict(type='str', default='member',
                          choices=['viewer', 'member', 'owner']),
            ),
        ),
        customer_access=dict(
            type='list',
            elements='dict',
            options=dict(
                customer=dict(type='str', required=True),
                role=dict(type='str', default='member',
                          choices=['viewer', 'member', 'owner']),
            ),
        ),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']

    try:
        client = WarmDeskClient.from_module(module)

        # ------------------------------------------------------------------
        # Locate existing group by name
        # ------------------------------------------------------------------
        existing = _find_group(client, p['name'])

        # ------------------------------------------------------------------
        # state=absent
        # ------------------------------------------------------------------
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, group=None)
            if not module.check_mode:
                client.delete('/admin/groups/%d' % existing['id'])
            module.exit_json(changed=True, group=None)

        # ------------------------------------------------------------------
        # state=present — CREATE
        # ------------------------------------------------------------------
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, group=None)

            body = {'name': p['name']}
            if p.get('description') is not None:
                body['description'] = p['description']

            group = client.post('/admin/groups', body)
            group_id = group['id']

            if p.get('members') is not None:
                _sync_members(client, group_id, [], p['members'], module)

            if p.get('project_access') is not None:
                _sync_project_access(client, group_id, [], p['project_access'], module)

            if p.get('customer_access') is not None:
                _sync_customer_access(client, group_id, [], p['customer_access'], module)

            detail = client.get('/admin/groups/%d' % group_id)
            module.exit_json(changed=True, group=detail)

        # ------------------------------------------------------------------
        # state=present — UPDATE
        # ------------------------------------------------------------------
        group_id = existing['id']
        detail = client.get('/admin/groups/%d' % group_id)

        changed = False

        # Check description change
        desc_changed = (
            p.get('description') is not None
            and existing.get('description') != p['description']
        )
        if desc_changed:
            if not module.check_mode:
                client.patch('/admin/groups/%d' % group_id,
                             {'description': p['description']})
            changed = True

        # Check member list
        if p.get('members') is not None:
            if _members_changed(detail.get('members', []), p['members']):
                if not module.check_mode:
                    _sync_members(
                        client, group_id,
                        detail.get('members', []),
                        p['members'],
                        module,
                    )
                changed = True

        # Check project access
        if p.get('project_access') is not None:
            if _project_access_changed(detail.get('project_access', []),
                                        p['project_access']):
                if not module.check_mode:
                    _sync_project_access(
                        client, group_id,
                        detail.get('project_access', []),
                        p['project_access'],
                        module,
                    )
                changed = True

        # Check customer access
        if p.get('customer_access') is not None:
            if _customer_access_changed(detail.get('customer_access', []),
                                         p['customer_access']):
                if not module.check_mode:
                    _sync_customer_access(
                        client, group_id,
                        detail.get('customer_access', []),
                        p['customer_access'],
                        module,
                    )
                changed = True

        if not changed:
            module.exit_json(changed=False, group=detail)

        if not module.check_mode:
            detail = client.get('/admin/groups/%d' % group_id)

        module.exit_json(changed=True, group=detail)

    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='WarmDesk API error (HTTP %s): %s' % (exc.status, exc.message)
        )


def main():
    run_module()


if __name__ == '__main__':
    main()
