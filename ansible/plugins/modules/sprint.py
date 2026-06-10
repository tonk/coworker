# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: sprint
short_description: Manage WarmDesk Scrum sprints
version_added: "0.5.0"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete Scrum sprints on a WarmDesk project board.
  - Idempotent — a second run with the same parameters produces no change.
  - Supports check mode.
notes:
  - Requires at least I(member) role on the project (or global admin).
  - The sprint name is used as the idempotency key within a project.
  - Status transitions follow the sprint lifecycle C(planning) → C(active) →
    C(completed).  Only one sprint can be active at a time; the module will
    fail if another sprint is already active when trying to start this one.
  - Deleting a sprint that does not exist is a no-op (no error).
  - Completing a sprint moves all unfinished cards back to the backlog.

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  project:
    description:
      - Slug of the WarmDesk project the sprint belongs to (e.g. C(my-project)).
    type: str
    required: true

  name:
    description:
      - Display name of the sprint.  Used as the idempotency key within a project.
    type: str
    required: true

  goal:
    description:
      - Optional sprint goal text displayed at the top of the board.
    type: str

  start_date:
    description:
      - Planned start date in C(YYYY-MM-DD) format.
    type: str

  end_date:
    description:
      - Planned end date in C(YYYY-MM-DD) format.
    type: str

  status:
    description:
      - Desired sprint status.
      - C(planning) is the default for new sprints.
      - Setting C(active) on a C(planning) sprint calls C(POST .../start) on the
        server.  There must be no other active sprint in the project or the task
        fails.
      - Setting C(completed) on an C(active) sprint calls C(POST .../complete).
        Unfinished cards are returned to the backlog.
      - Transitions that skip a step (e.g. C(planning) → C(completed)) are not
        supported and will fail.
    type: str
    choices: [planning, active, completed]

  state:
    description:
      - C(present) ensures the sprint exists with the specified attributes.
      - C(absent) removes the sprint if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
- name: Create Sprint 1
  ansilabnl.warmdesk.sprint:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Sprint 1
    goal: Deliver the new login flow.
    start_date: "2026-07-01"
    end_date: "2026-07-14"

- name: Start Sprint 1
  ansilabnl.warmdesk.sprint:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Sprint 1
    status: active

- name: Complete Sprint 1
  ansilabnl.warmdesk.sprint:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Sprint 1
    status: completed

- name: Create and immediately start a sprint
  ansilabnl.warmdesk.sprint:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: frontend
    name: "{{ sprint_name }}"
    goal: "{{ sprint_goal }}"
    start_date: "{{ sprint_start }}"
    end_date: "{{ sprint_end }}"
    status: active

- name: Delete a sprint
  ansilabnl.warmdesk.sprint:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: backend
    name: Sprint 1
    state: absent
"""

RETURN = r"""
sprint:
  description:
    - The sprint object as returned by the WarmDesk API after the operation.
    - C(null) when C(state=absent) and the sprint was deleted (or did not exist).
  returned: always
  type: dict
  contains:
    id:
      description: Numeric sprint ID.
      returned: always
      type: int
      sample: 5
    project_id:
      description: Numeric ID of the owning project.
      returned: always
      type: int
      sample: 1
    name:
      description: Sprint name.
      returned: always
      type: str
      sample: Sprint 1
    goal:
      description: Sprint goal text.
      returned: always
      type: str
      sample: Deliver the new login flow.
    status:
      description: C(planning), C(active), or C(completed).
      returned: always
      type: str
      sample: planning
    start_date:
      description: Planned start date (ISO-8601) or null.
      returned: always
      type: str
      sample: "2026-07-01T00:00:00Z"
    end_date:
      description: Planned end date (ISO-8601) or null.
      returned: always
      type: str
      sample: "2026-07-14T00:00:00Z"
    card_count:
      description: Number of cards in the sprint.
      returned: always
      type: int
      sample: 8
    total_points:
      description: Sum of story points across all sprint cards.
      returned: always
      type: int
      sample: 21
    completed_points:
      description: Story points for done cards.
      returned: always
      type: int
      sample: 13
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2026-06-28T10:00:00Z"
"""

import re

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


_DATE_RE = re.compile(r'^\d{4}-\d{2}-\d{2}$')


def _date_to_iso(date_str):
    """Convert YYYY-MM-DD to a full ISO timestamp accepted by the server."""
    if not date_str or not _DATE_RE.match(date_str):
        return None
    return '%sT00:00:00Z' % date_str


def _date_field_changed(desired_str, existing_iso):
    """True when the desired YYYY-MM-DD differs from the stored ISO timestamp."""
    if desired_str is None:
        return False
    if not desired_str:
        return existing_iso is not None
    if existing_iso is None:
        return True
    return existing_iso[:10] != desired_str


def _find_sprint(client, project_slug, name):
    """Return the sprint dict whose name matches, or None."""
    sprints = client.get('/projects/%s/sprints' % project_slug)
    for s in sprints:
        if s.get('name') == name:
            return s
    return None


def _build_update_body(p, existing):
    body = {}
    changed = False

    if p.get('name') is not None and existing.get('name') != p['name']:
        body['name'] = p['name']
        changed = True

    if p.get('goal') is not None and existing.get('goal') != p['goal']:
        body['goal'] = p['goal']
        changed = True

    if _date_field_changed(p.get('start_date'), existing.get('start_date')):
        body['start_date'] = _date_to_iso(p['start_date'])
        changed = True

    if _date_field_changed(p.get('end_date'), existing.get('end_date')):
        body['end_date'] = _date_to_iso(p['end_date'])
        changed = True

    return body, changed


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        name=dict(type='str', required=True),
        goal=dict(type='str'),
        start_date=dict(type='str'),
        end_date=dict(type='str'),
        status=dict(type='str', choices=['planning', 'active', 'completed']),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']
    client = WarmDeskClient.from_module(module)

    try:
        existing = _find_sprint(client, p['project'], p['name'])

        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, sprint=None)
            if not module.check_mode:
                client.delete('/projects/%s/sprints/%d' % (p['project'], existing['id']))
            module.exit_json(changed=True, sprint=None)

        # state=present — CREATE
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, sprint=None)
            body = dict(name=p['name'])
            if p.get('goal') is not None:
                body['goal'] = p['goal']
            if p.get('start_date'):
                body['start_date'] = _date_to_iso(p['start_date'])
            if p.get('end_date'):
                body['end_date'] = _date_to_iso(p['end_date'])
            sprint = client.post('/projects/%s/sprints' % p['project'], body)

            # Apply status transition after creation if requested
            desired_status = p.get('status')
            if desired_status and desired_status != 'planning':
                if desired_status == 'active':
                    sprint = client.post(
                        '/projects/%s/sprints/%d/start' % (p['project'], sprint['id'])
                    )
                elif desired_status == 'completed':
                    module.fail_json(
                        msg='Cannot set status=completed on a newly created sprint; '
                            'start it first.'
                    )
            module.exit_json(changed=True, sprint=sprint)

        # state=present — UPDATE
        update_body, fields_changed = _build_update_body(p, existing)

        desired_status = p.get('status')
        current_status = existing.get('status', 'planning')
        status_changed = desired_status is not None and desired_status != current_status

        if status_changed:
            # Validate the transition
            valid_transitions = {
                ('planning', 'active'): 'start',
                ('active', 'completed'): 'complete',
            }
            transition = valid_transitions.get((current_status, desired_status))
            if transition is None:
                module.fail_json(
                    msg='Invalid sprint status transition: %s → %s. '
                        'Supported: planning → active, active → completed.'
                        % (current_status, desired_status)
                )

        if not fields_changed and not status_changed:
            module.exit_json(changed=False, sprint=existing)

        if module.check_mode:
            module.exit_json(changed=True, sprint=existing)

        sprint = existing
        if update_body:
            sprint = client.put(
                '/projects/%s/sprints/%d' % (p['project'], existing['id']), update_body
            )

        if status_changed:
            sprint = client.post(
                '/projects/%s/sprints/%d/%s' % (p['project'], existing['id'], transition)
            )

        module.exit_json(changed=True, sprint=sprint)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
