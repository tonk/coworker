# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: project_member
short_description: Manage membership of a WarmDesk project
version_added: "1.0.0"
author:
  - Ton Kersten (@tonk)
description:
  - Add, update the role of, or remove a user from a WarmDesk project.
  - Idempotent — adding a user who is already a member with the correct role
    produces no change. Updating the role is also idempotent.
  - The C(project) parameter takes the project B(slug) (the server-generated
    URL-safe identifier). Use the return value of
    M(ansilabnl.warmdesk.project) to obtain the slug.
notes:
  - Only project owners (or global admins) may add members or change roles.
    Members may add other members but cannot promote to owner or change existing
    roles; viewers cannot modify membership at all.
  - A user can remove themselves from a project regardless of role, but only
    owners/admins may remove other members.
  - The API uses C(login) (username or e-mail) to add a member. This module
    always sends the username; the server accepts either.
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
options:
  project:
    description:
      - The project slug (e.g. C(edge-data-analytics)). This is the
        server-generated, URL-safe identifier returned by
        M(ansilabnl.warmdesk.project) as C(project.slug).
    type: str
    required: true
  username:
    description:
      - WarmDesk username of the user to add, update, or remove.
    type: str
    required: true
  role:
    description:
      - Project role to assign to the user.
      - C(owner) — full control including project settings and member management.
      - C(member) — can create and edit cards; default when omitted.
      - C(viewer) — read-only access.
      - Ignored when C(state=absent).
    type: str
    choices: [owner, member, viewer]
    default: member
  state:
    description:
      - C(present) ensures the user is a member of the project with the
        specified C(role). If the user is already a member with a different
        role, the role is updated.
      - C(absent) removes the user from the project. If the user is not a
        member, no change is made.
    type: str
    choices: [present, absent]
    default: present
'''

EXAMPLES = r'''
- name: Add a member to a project (default role)
  ansilabnl.warmdesk.project_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    project: edge-data-analytics
    username: jdoe
    state: present

- name: Add a viewer
  ansilabnl.warmdesk.project_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    project: edge-data-analytics
    username: auditor
    role: viewer
    state: present

- name: Promote an existing member to owner
  ansilabnl.warmdesk.project_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    project: edge-data-analytics
    username: jdoe
    role: owner
    state: present

- name: Ensure multiple users are members using a loop
  ansilabnl.warmdesk.project_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    project: "{{ project.slug }}"
    username: "{{ item.user }}"
    role: "{{ item.role | default('member') }}"
    state: present
  loop:
    - { user: alice, role: owner }
    - { user: bob }
    - { user: carol, role: viewer }

- name: Remove a user from a project
  ansilabnl.warmdesk.project_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    project: edge-data-analytics
    username: former-employee
    state: absent

- name: Add member using a project slug from warmdesk_project output
  ansilabnl.warmdesk.project:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    name: New Team Project
    state: present
  register: new_project

- name: Grant the service account viewer access
  ansilabnl.warmdesk.project_member:
    warmdesk_url: https://desk.example.com
    warmdesk_token: "{{ vault_wd_token }}"
    project: "{{ new_project.project.slug }}"
    username: ci-bot
    role: viewer
    state: present
'''

RETURN = r'''
changed:
  description: Whether the module made any changes on the server.
  type: bool
  returned: always
  sample: true
member:
  description:
    - The membership record after the module ran.
    - C(null) when C(state=absent) and the membership did not exist or was
      just removed.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric membership record ID.
      type: int
      sample: 17
    project_id:
      description: Numeric ID of the project.
      type: int
      sample: 42
    user_id:
      description: Numeric ID of the member user.
      type: int
      sample: 5
    role:
      description: The user's role in this project.
      type: str
      sample: member
    invited_by:
      description: Numeric ID of the user who added this member.
      type: int
      sample: 1
    created_at:
      description: ISO-8601 timestamp when the membership was created.
      type: str
      sample: "2026-01-20T14:00:00Z"
    updated_at:
      description: ISO-8601 timestamp of the last update to this record.
      type: str
      sample: "2026-03-10T09:15:00Z"
    user:
      description: Embedded user object.
      type: dict
      contains:
        id:
          description: Numeric user ID.
          type: int
          sample: 5
        username:
          description: Login username.
          type: str
          sample: jdoe
        email:
          description: User e-mail address.
          type: str
          sample: jdoe@example.com
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _find_member(members, username):
    """Return the member dict whose embedded user has the given username, or None."""
    for m in members:
        user = m.get('user', {})
        if user.get('username') == username:
            return m
    return None


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(
        project=dict(type='str', required=True),
        username=dict(type='str', required=True),
        role=dict(type='str', default='member', choices=['owner', 'member', 'viewer']),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    params = module.params
    state = params['state']
    slug = params['project']
    username = params['username']
    desired_role = params['role']

    try:
        client = WarmDeskClient.from_module(module)

        # ------------------------------------------------------------------
        # Fetch current members
        # ------------------------------------------------------------------
        try:
            members = client.get('/projects/%s/members' % slug)
        except WarmDeskAPIError as exc:
            if exc.status == 404:
                module.fail_json(
                    msg='Project "%s" not found or not accessible.' % slug
                )
            raise

        existing = _find_member(members, username)

        # ------------------------------------------------------------------
        # state=absent
        # ------------------------------------------------------------------
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, member=None)

            user_id = existing['user_id']

            if not module.check_mode:
                client.delete('/projects/%s/members/%d' % (slug, user_id))

            module.exit_json(changed=True, member=None)

        # ------------------------------------------------------------------
        # state=present — Add new member
        # ------------------------------------------------------------------
        if existing is None:
            if module.check_mode:
                module.exit_json(
                    changed=True,
                    member=dict(
                        user=dict(username=username),
                        role=desired_role,
                        project_id=None,
                        user_id=None,
                    ),
                )

            new_member = client.post(
                '/projects/%s/members' % slug,
                {'login': username, 'role': desired_role},
            )
            module.exit_json(changed=True, member=new_member)

        # ------------------------------------------------------------------
        # state=present — Update role if needed
        # ------------------------------------------------------------------
        current_role = existing.get('role')
        if current_role == desired_role:
            # Already correct — nothing to do
            module.exit_json(changed=False, member=existing)

        user_id = existing['user_id']

        if module.check_mode:
            updated = dict(existing)
            updated['role'] = desired_role
            module.exit_json(changed=True, member=updated)

        updated_member = client.put(
            '/projects/%s/members/%d/role' % (slug, user_id),
            {'role': desired_role},
        )
        module.exit_json(changed=True, member=updated_member)

    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='WarmDesk API error (HTTP %s): %s' % (exc.status, exc.message)
        )


def main():
    run_module()


if __name__ == '__main__':
    main()
