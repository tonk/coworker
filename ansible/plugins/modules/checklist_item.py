# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: checklist_item
short_description: Manage checklist items on a WarmDesk card
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Creates, updates, or deletes a single checklist item on a WarmDesk card.
  - Idempotent — the item is identified by I(project) + I(card_number) + I(body)
    text. Running the module again with unchanged parameters produces no change.
  - The card is located via I(card_number) (e.g. C(EDA-42)) using
    C(find_card_by_number) from the shared resolver utilities.
  - Requires at least I(member) role on the project (or global admin) for
    write operations; viewer role for reads.
options:
  project:
    description:
      - Slug of the WarmDesk project that owns the card (e.g. C(my-project)).
    type: str
    required: true
  card_number:
    description:
      - Card reference in C(<PREFIX>-<NUMBER>) format, e.g. C(EDA-42).
      - Used to locate the card before operating on its checklist.
    type: str
    required: true
  body:
    description:
      - Text of the checklist item. Used as the idempotency key together with
        I(project) and I(card_number).
    type: str
    required: true
  is_completed:
    description:
      - Whether the checklist item should be marked as completed.
      - Defaults to C(false) on creation; updated when it differs from the
        existing value.
    type: bool
    default: false
  state:
    description:
      - C(present) — ensure the checklist item exists with the given I(body)
        and I(is_completed) value.
      - C(absent) — ensure the checklist item does not exist.
    type: str
    choices: [present, absent]
    default: present
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
notes:
  - Check mode is fully supported; no changes are made to the server.
  - If two items with identical body text exist on the same card (which the UI
    allows), only the first match is operated on. Use distinct body text to
    avoid ambiguity.
seealso:
  - module: ansilabnl.warmdesk.card
  - module: ansilabnl.warmdesk.column
'''

EXAMPLES = r'''
- name: Add an uncompleted checklist item to card EDA-42
  ansilabnl.warmdesk.checklist_item:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    body: "Write unit tests"

- name: Add a pre-completed item (e.g. already done in a previous sprint)
  ansilabnl.warmdesk.checklist_item:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    body: "Design database schema"
    is_completed: true

- name: Mark an existing checklist item as completed
  ansilabnl.warmdesk.checklist_item:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    body: "Write unit tests"
    is_completed: true

- name: Provision a full checklist from a variable
  ansilabnl.warmdesk.checklist_item:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: ops-board
    card_number: OPS-7
    body: "{{ item.body }}"
    is_completed: "{{ item.done | default(false) }}"
  loop:
    - {body: "Provision server",      done: true}
    - {body: "Configure firewall",    done: true}
    - {body: "Deploy application",    done: false}
    - {body: "Run smoke tests",       done: false}
    - {body: "Notify stakeholders",   done: false}

- name: Remove a checklist item
  ansilabnl.warmdesk.checklist_item:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    body: "Obsolete step"
    state: absent

- name: Add item and capture result
  ansilabnl.warmdesk.checklist_item:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    project: dev-project
    card_number: DEV-3
    body: "Code review"
  register: item_result

- name: Show the checklist item
  ansible.builtin.debug:
    var: item_result.checklist_item
'''

RETURN = r'''
changed:
  description: Whether the module made any change on the server.
  type: bool
  returned: always
checklist_item:
  description:
    - The checklist item object as returned by the WarmDesk API.
    - C(null) when I(state=absent) and the item did not exist, or after a
      successful deletion.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric database ID of the checklist item.
      type: int
      sample: 15
    card_id:
      description: ID of the card that owns this item.
      type: int
      sample: 42
    body:
      description: Text of the checklist item.
      type: str
      sample: Write unit tests
    is_completed:
      description: Whether the item has been ticked off.
      type: bool
      sample: false
    position:
      description: Float used for ordering items within the checklist.
      type: float
      sample: 2000.0
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


def _find_item(items, body):
    """Return the first checklist item whose body matches, or None."""
    for item in items:
        if item.get('body') == body:
            return item
    return None


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        card_number=dict(type='str', required=True),
        body=dict(type='str', required=True),
        is_completed=dict(type='bool', default=False),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    project = module.params['project']
    card_number = module.params['card_number']
    body = module.params['body']
    is_completed = module.params['is_completed']
    state = module.params['state']

    try:
        client = WarmDeskClient.from_module(module)
    except WarmDeskAPIError as exc:
        module.fail_json(msg=str(exc))

    # ----------------------------------------------------------------
    # Resolve card reference → card dict
    # ----------------------------------------------------------------
    try:
        card = find_card_by_number(client, project, card_number)
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Error looking up card "%s": %s (HTTP %s)' % (
                card_number, exc.message, exc.status)
        )

    if card is None:
        if state == 'absent':
            module.exit_json(changed=False, checklist_item=None)
        module.fail_json(
            msg='Card "%s" not found in project "%s".' % (card_number, project)
        )

    card_id = card['id']

    # ----------------------------------------------------------------
    # Fetch checklist
    # ----------------------------------------------------------------
    try:
        items = client.get('/projects/%s/cards/%d/checklist' % (project, card_id))
    except WarmDeskAPIError as exc:
        module.fail_json(
            msg='Failed to list checklist for card "%s": %s (HTTP %s)' % (
                card_number, exc.message, exc.status)
        )

    existing = _find_item(items, body)

    # ------------------------------------------------------------------ absent
    if state == 'absent':
        if existing is None:
            module.exit_json(changed=False, checklist_item=None)
        if module.check_mode:
            module.exit_json(changed=True, checklist_item=None)
        try:
            client.delete(
                '/projects/%s/cards/%d/checklist/%d' % (project, card_id, existing['id'])
            )
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to delete checklist item: %s (HTTP %s)' % (exc.message, exc.status)
            )
        module.exit_json(changed=True, checklist_item=None)

    # ----------------------------------------------------------------- present
    if existing is None:
        # Create
        if module.check_mode:
            module.exit_json(changed=True, checklist_item=None)
        try:
            created = client.post(
                '/projects/%s/cards/%d/checklist' % (project, card_id),
                {'body': body, 'is_completed': is_completed},
            )
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to create checklist item: %s (HTTP %s)' % (exc.message, exc.status)
            )
        module.exit_json(changed=True, checklist_item=created)

    # Update if is_completed differs
    if existing.get('is_completed') != is_completed:
        if module.check_mode:
            module.exit_json(changed=True, checklist_item=existing)
        try:
            updated = client.put(
                '/projects/%s/cards/%d/checklist/%d' % (project, card_id, existing['id']),
                {'body': body, 'is_completed': is_completed},
            )
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to update checklist item: %s (HTTP %s)' % (exc.message, exc.status)
            )
        module.exit_json(changed=True, checklist_item=updated)

    # No change needed
    module.exit_json(changed=False, checklist_item=existing)


def main():
    run_module()


if __name__ == '__main__':
    main()
