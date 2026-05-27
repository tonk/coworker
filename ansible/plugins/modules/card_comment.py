# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: card_comment
short_description: Manage comments on a WarmDesk card
version_added: "0.3.3"
author: "Ton Kersten (@tonk)"
description:
  - Creates, updates, or deletes a comment on a WarmDesk card.
  - The card is located via I(project) + I(card_number) (e.g. C(EDA-42)).
  - When I(comment_id) is supplied the module is idempotent — it only updates
    the comment when I(body) or I(time_spent) differs from the stored
    value.
  - When I(comment_id) is omitted and I(state=present) a new comment is
    B(always) created. This is intentional for CI/CD pipelines that post a
    deployment notice without caring whether one already exists.
  - Requires at least I(member) role on the project (or global admin) for
    write operations.
options:
  project:
    description:
      - Slug of the WarmDesk project that owns the card (e.g. C(my-project)).
    type: str
    required: true
  card_number:
    description:
      - Card reference in C(<PREFIX>-<NUMBER>) format, e.g. C(EDA-42).
    type: str
    required: true
  body:
    description:
      - Text of the comment (Markdown supported).
      - Required when I(state=present).
    type: str
  comment_id:
    description:
      - Numeric ID of an existing comment.
      - Supply this to update or delete a specific comment.
      - When omitted with I(state=present) a new comment is always created.
      - Required when I(state=absent).
    type: int
  time_spent:
    description:
      - Time logged with this comment, in minutes.
      - Only used when I(state=present).
      - Pass C(0) to clear an existing time entry on update.
    type: int
    default: 0
  state:
    description:
      - C(present) — ensure the comment exists with the given I(body).
      - C(absent) — delete the comment identified by I(comment_id).
    type: str
    choices: [present, absent]
    default: present
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
notes:
  - Check mode is fully supported; no changes are made to the server.
  - Only the comment author can update a comment. Use the same API key for
    both create and update tasks.
  - Project owners and global admins can delete any comment regardless of
    authorship.
seealso:
  - module: ansilabnl.warmdesk.card
  - module: ansilabnl.warmdesk.checklist_item
'''

EXAMPLES = r'''
- name: Post a deployment notification comment
  ansilabnl.warmdesk.card_comment:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    body: "Deployed to production at {{ ansible_date_time.iso8601 }}"
  register: deploy_comment

- name: Show the new comment ID (for later update or delete)
  ansible.builtin.debug:
    msg: "Comment ID is {{ deploy_comment.comment.id }}"

- name: Log time with a comment
  ansilabnl.warmdesk.card_comment:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: ops-board
    card_number: OPS-7
    body: "Investigated network issue; root cause was a misconfigured firewall rule."
    time_spent: 90

- name: Update an existing comment by ID (idempotent)
  ansilabnl.warmdesk.card_comment:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    comment_id: "{{ deploy_comment.comment.id }}"
    body: "Deployed to production at {{ ansible_date_time.iso8601 }} — rollback available"

- name: Delete a comment
  ansilabnl.warmdesk.card_comment:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    comment_id: 17
    state: absent
'''

RETURN = r'''
changed:
  description: Whether the module made any change on the server.
  type: bool
  returned: always
comment:
  description:
    - The comment object as returned by the WarmDesk API.
    - C(null) when I(state=absent) and the comment did not exist, or after a
      successful deletion.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric database ID of the comment.
      type: int
      sample: 17
    card_id:
      description: ID of the card that owns this comment.
      type: int
      sample: 42
    body:
      description: Text of the comment.
      type: str
      sample: "Deployed to production"
    is_edited:
      description: Whether the comment has been edited after creation.
      type: bool
      sample: false
    time_spent:
      description: Time logged with this comment, in minutes.
      type: int
      sample: 0
    created_at:
      description: ISO 8601 creation timestamp.
      type: str
      sample: "2026-05-21T09:00:00Z"
    user:
      description: Author of the comment (subset of user object).
      type: dict
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_resolve import (
    find_card_by_number,
)


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        card_number=dict(type='str', required=True),
        body=dict(type='str'),
        comment_id=dict(type='int'),
        time_spent=dict(type='int', default=0),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ('state', 'present', ('body',)),
            ('state', 'absent', ('comment_id',)),
        ],
    )

    project = module.params['project']
    card_number = module.params['card_number']
    body = module.params['body']
    comment_id = module.params['comment_id']
    time_spent = module.params['time_spent']
    state = module.params['state']

    try:
        client = WarmDeskClient.from_module(module)
    except WarmDeskAPIError as exc:
        module.fail_json(msg=str(exc))

    # ------------------------------------------------------------------
    # Resolve card reference → card dict
    # ------------------------------------------------------------------
    try:
        card = find_card_by_number(client, project, card_number)
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Error looking up card "%s": %s (HTTP %s)' % (
                card_number, exc.message, exc.status)
        )

    if card is None:
        if state == 'absent':
            module.exit_json(changed=False, comment=None)
        module.fail_json(
            msg='Card "%s" not found in project "%s".' % (card_number, project)
        )

    card_id = card['id']
    base_url = '/projects/%s/cards/%d/comments' % (project, card_id)

    # ------------------------------------------------------------------ absent
    if state == 'absent':
        if module.check_mode:
            module.exit_json(changed=True, comment=None)
        try:
            client.delete('%s/%d' % (base_url, comment_id))
        except WarmDeskAPIError as exc:
            if exc.status == 404:
                module.exit_json(changed=False, comment=None)
            module.fail_json(
                msg='Failed to delete comment %d: %s (HTTP %s)' % (
                    comment_id, exc.message, exc.status)
            )
        module.exit_json(changed=True, comment=None)

    # ------------------------------------------------------------------ present
    if comment_id is None:
        # Create — always results in a new comment
        if module.check_mode:
            module.exit_json(changed=True, comment=None)
        try:
            created = client.post(base_url, {
                'body': body,
                'time_spent_minutes': time_spent,
            })
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to create comment: %s (HTTP %s)' % (exc.message, exc.status)
            )
        module.exit_json(changed=True, comment=created)

    # Update existing comment — fetch current state first
    try:
        comments = client.get(base_url)
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Failed to list comments for card "%s": %s (HTTP %s)' % (
                card_number, exc.message, exc.status)
        )

    existing = next((c for c in comments if c.get('id') == comment_id), None)
    if existing is None:
        module.fail_json(
            msg='Comment %d not found on card "%s".' % (comment_id, card_number)
        )

    if existing.get('body') == body and existing.get('time_spent_minutes', 0) == time_spent:
        module.exit_json(changed=False, comment=existing)

    if module.check_mode:
        module.exit_json(changed=True, comment=existing)

    try:
        updated = client.put('%s/%d' % (base_url, comment_id), {
            'body': body,
            'time_spent_minutes': time_spent,
        })
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Failed to update comment %d: %s (HTTP %s)' % (
                comment_id, exc.message, exc.status)
        )
    module.exit_json(changed=True, comment=updated)


def main():
    run_module()


if __name__ == '__main__':
    main()
