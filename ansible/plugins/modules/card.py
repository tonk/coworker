# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r'''
---
module: card
short_description: Manage cards on a WarmDesk Kanban board
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Creates, updates, moves, or deletes a card on a WarmDesk Kanban board.
  - Idempotent for I(state=present) — the module first tries to locate the card
    by I(card_number) (preferred) or by I(project) + I(column) + I(title).
    If a matching card is found its fields are compared and only a PUT is
    issued when something differs.
  - For I(state=absent) and updates, I(card_number) (e.g. C(EDA-42)) is the
    preferred lookup key; I(title) is used as a fallback only for I(present).
  - Optional I(move_to_column) — if the card exists in a different column a
    PATCH move request is issued automatically.
  - Requires at least I(member) role on the project (or global admin).
options:
  project:
    description:
      - Slug of the WarmDesk project (e.g. C(my-project)).
    type: str
    required: true
  column:
    description:
      - Name of the target column (e.g. C(Backlog)).
      - Required when creating a card (I(state=present) and no existing card
        is found by I(card_number)).
      - Also used as the idempotency key when I(card_number) is not given.
    type: str
    required: false
  title:
    description:
      - Card title. Required for all operations.
      - Used as part of the idempotency key (with I(project) and I(column))
        when I(card_number) is not supplied.
    type: str
    required: true
  description:
    description:
      - Markdown body / description of the card.
    type: str
    required: false
  priority:
    description:
      - Priority level of the card.
    type: str
    choices: [none, low, medium, high, urgent]
    default: none
  start_date:
    description:
      - Start date in C(YYYY-MM-DD) format, or C(null) to clear.
    type: str
    required: false
  due_date:
    description:
      - Due date in C(YYYY-MM-DD) format, or C(null) to clear.
    type: str
    required: false
  assignee:
    description:
      - Username or email address of the user to assign to the card.
      - The module resolves the value to a numeric user ID.  It first
        queries C(GET /api/v1/users); if that returns 403 (project-scoped
        API key) it falls back to C(GET /api/v1/projects/{slug}/members),
        which is always accessible to project-scoped keys.
    type: str
    required: false
  card_number:
    description:
      - Existing card reference (e.g. C(EDA-42)).
      - When provided this is used as the primary lookup key, bypassing the
        title-based search.
    type: str
    required: false
  closed:
    description:
      - Whether the card is closed (done).
      - C(true) closes the card and sets C(closed_at); C(false) reopens it.
      - Omit the parameter to leave the closed state unchanged.
    type: bool
    required: false
  time_spent:
    description:
      - Total time logged against the card, in minutes (denormalized).
      - This directly sets the card-level counter, bypassing the
        comment-based auto-creation path.  Set to C(0) to clear it.
    type: int
    required: false
  story_points:
    description:
      - Story points for Scrum-board projects.
      - Not accepted by the create endpoint — when set together with card
        creation, this module applies it with an immediate follow-up update.
    type: int
    required: false
  external_issue_url:
    description:
      - URL of a linked external issue (e.g. a Jira or GitHub issue).
      - Pass an empty string to clear a previously set link.
      - Not accepted by the create endpoint — when set together with card
        creation, this module applies it with an immediate follow-up update.
    type: str
    required: false
  external_issue_ref:
    description:
      - Human-readable reference of a linked external issue (e.g. C(PROJ-123)).
      - Pass an empty string to clear a previously set reference.
      - Not accepted by the create endpoint — when set together with card
        creation, this module applies it with an immediate follow-up update.
    type: str
    required: false
  epic_id:
    description:
      - Numeric ID of the epic to link this card to, as a string (e.g.
        C("7")), or the literal string C(null) to clear it.
      - Not accepted by the create endpoint — when set together with card
        creation, this module applies it with an immediate follow-up update.
    type: str
    required: false
  move_to_column:
    description:
      - Name of the column the card should be moved to.
      - When set and the card is found in a different column, a PATCH move
        request is issued. Has no effect when the card is already in the
        target column.
    type: str
    required: false
  state:
    description:
      - C(present) — ensure the card exists and its fields match.
      - C(absent) — ensure the card does not exist; requires I(card_number)
        or a unique I(title) match.
    type: str
    choices: [present, absent]
    default: present
extends_documentation_fragment:
  - ansilabnl.warmdesk.connection
notes:
  - Check mode is fully supported; no API writes are issued.
  - Date fields accept C(null) as a string to explicitly clear the date on an
    existing card.
  - The move operation (PATCH) is performed I(after) any field update (PUT),
    so both may fire in a single module run.
  - C(description) cannot be cleared to empty through this API — the server
    treats an empty value the same as "not supplied" and leaves the stored
    description untouched. This module's idempotency check does the same, so
    it won't report C(changed) forever trying to apply a clear that can never
    take effect.
  - C(story_points), C(external_issue_url), C(external_issue_ref), and
    C(epic_id) are not accepted by the create endpoint. When any of them are
    set on a newly created card, this module issues an immediate follow-up
    update so the field still ends up applied — at the cost of one extra
    request on the creating run only.
  - Looking up a card by I(card_number) cannot find a sub-card. The project
    endpoint this module reads only returns top-level cards (those with no
    parent card), so a sub-card's reference is never present in that list to
    scan — this module has no way to locate one, whether by number or by
    title. Operate on the parent card, or use I(card_number) for a top-level
    card only.
seealso:
  - module: ansilabnl.warmdesk.column
  - module: ansilabnl.warmdesk.checklist_item
'''

EXAMPLES = r'''
- name: Create a card in the Backlog column
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    column: Backlog
    title: "Implement login page"
    description: "OAuth2 + local password fallback"
    priority: high
    due_date: "2026-05-01"

- name: Assign a card to a team member and set a start date
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    assignee: jdoe
    start_date: "2026-04-20"

- name: Move card EDA-42 to "In Review" column
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    column: In Progress
    title: Placeholder title
    card_number: EDA-42
    move_to_column: In Review

- name: Update priority of a known card by number
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    column: In Progress
    title: "Deploy monitoring stack"
    card_number: OPS-7
    priority: urgent
    due_date: "2026-04-30"

- name: Close a card
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    title: irrelevant
    card_number: EDA-42
    closed: true

- name: Reopen a card
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    title: irrelevant
    card_number: EDA-42
    closed: false

- name: Set logged time (in minutes) on a card
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    card_number: EDA-42
    time_spent: 120

- name: Link a card to an epic and an external tracker issue
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    title: irrelevant
    card_number: EDA-42
    epic_id: "7"
    external_issue_url: "https://github.com/example/repo/issues/123"
    external_issue_ref: "PROJ-123"

- name: Delete a card by number
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    title: irrelevant
    card_number: EDA-99
    state: absent

- name: Create a batch of cards from a list
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: ops-board
    column: Backlog
    title: "{{ item.title }}"
    priority: "{{ item.priority | default('none') }}"
  loop:
    - {title: "Setup CI pipeline",  priority: high}
    - {title: "Write runbook",      priority: medium}
    - {title: "Review access list", priority: low}
  register: card_results

- name: Create a card and use its reference in a follow-up task
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    column: Backlog
    title: "Deploy monitoring stack"
    priority: high
  register: new_card

- name: Move the newly created card to In Progress
  ansilabnl.warmdesk.card:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"
    project: my-project
    title: "Deploy monitoring stack"
    card_number: "{{ new_card.card.card_ref }}"
    move_to_column: In Progress
'''

RETURN = r'''
changed:
  description: Whether the module made any change on the server.
  type: bool
  returned: always
card:
  description:
    - The card object as returned by the WarmDesk API.
    - C(null) when I(state=absent) and the card was not found, or after a
      successful deletion.
  type: dict
  returned: always
  contains:
    id:
      description: Numeric database ID of the card.
      type: int
      sample: 42
    card_number:
      description: Sequential card number within the project.
      type: int
      sample: 42
    title:
      description: Card title.
      type: str
      sample: Implement login page
    description:
      description: Markdown body.
      type: str
      sample: "OAuth2 + local password fallback"
    priority:
      description: Priority level.
      type: str
      sample: high
    column_id:
      description: ID of the column that currently holds the card.
      type: int
      sample: 3
    project_id:
      description: ID of the owning project.
      type: int
      sample: 1
    assignee_id:
      description: ID of the assigned user, or C(null).
      type: int
      sample: 5
    start_date:
      description: Start date in ISO 8601 format, or C(null).
      type: str
      sample: "2026-04-20T00:00:00Z"
    due_date:
      description: Due date in ISO 8601 format, or C(null).
      type: str
      sample: "2026-05-01T00:00:00Z"
    time_spent:
      description: Total time logged against the card, in minutes.
      type: int
      sample: 120
    closed:
      description: Whether the card has been marked as done/closed.
      type: bool
      sample: false
    story_points:
      description: Story points, or C(null) when unset.
      type: int
      sample: 5
    external_issue_url:
      description: URL of a linked external issue, or empty when unset.
      type: str
      sample: "https://github.com/example/repo/issues/123"
    external_issue_ref:
      description: Human-readable external issue reference, or empty when unset.
      type: str
      sample: "PROJ-123"
    epic_id:
      description: ID of the linked epic, or C(null) when unset.
      type: int
      sample: 3
    position:
      description: Float used for ordering within the column.
      type: float
      sample: 65536.0
    card_ref:
      description: >
        Full card reference combining the project key prefix and the card
        number (e.g. C(GF00-4)).  Use this value as I(card_number) in
        subsequent module calls.
      type: str
      sample: "GF00-4"
'''

from ansible.module_utils.basic import AnsibleModule

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
    WarmDeskClient,
    warmdesk_argument_spec,
)
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_resolve import (
    find_card_by_number,
    resolve_column_id,
    resolve_user_id,
)

_PRIORITY_CHOICES = ('none', 'low', 'medium', 'high', 'urgent')

# Sentinel to distinguish "not supplied" from an explicit None / null.
_UNSET = object()


def _find_card_by_title(client, project, column_id, title):
    """Scan a column's cards and return the first match by title, or None."""
    try:
        cards = client.get('/projects/%s/columns/%d/cards' % (project, column_id))
    except WarmDeskAPIError:
        return None
    for card in cards:
        if card.get('title') == title:
            return card
    return None


def _card_needs_update(card, title, description, priority, start_date, due_date, assignee_id, closed, time_spent,
                        story_points, external_issue_url, external_issue_ref, epic_id):
    """Return True if any supplied field differs from the current card value."""
    if title and card.get('title') != title:
        return True
    # UpdateCard only applies description when it's non-empty (an empty
    # value can never be written, so treating it like "not supplied" here
    # matches actual server behaviour and avoids reporting changed=true on
    # every run for a clear that can never take effect).
    if description and card.get('description') != description:
        return True
    if priority and card.get('priority') != priority:
        return True
    # Dates: compare only the date part of the ISO string returned by the API
    if start_date is not _UNSET:
        api_val = (card.get('start_date') or '')[:10] or None
        desired = start_date  # 'YYYY-MM-DD' or None
        if api_val != desired:
            return True
    if due_date is not _UNSET:
        api_val = (card.get('due_date') or '')[:10] or None
        desired = due_date
        if api_val != desired:
            return True
    if assignee_id is not _UNSET:
        if card.get('assignee_id') != assignee_id:
            return True
    if closed is not _UNSET:
        if card.get('closed') != closed:
            return True
    if time_spent is not None and card.get('time_spent_minutes') != time_spent:
        return True
    if story_points is not None and card.get('story_points') != story_points:
        return True
    # external_issue_url/ref: the API applies these whenever supplied, even
    # empty, so an explicit "" is a real clear and must be compared as such.
    if external_issue_url is not None and (card.get('external_issue_url') or '') != external_issue_url:
        return True
    if external_issue_ref is not None and (card.get('external_issue_ref') or '') != external_issue_ref:
        return True
    if epic_id is not _UNSET and card.get('epic_id') != epic_id:
        return True
    return False


def _extra_fields_body(closed, time_spent, story_points, external_issue_url, external_issue_ref, epic_id):
    """Return a PUT-body fragment for the fields UpdateCard supports but
    CreateCard does not (closed, time_spent, story_points,
    external_issue_url/ref, epic_id), including only what was supplied."""
    body = {}
    if closed is not _UNSET:
        body['closed'] = closed
    if time_spent is not None:
        body['time_spent_minutes'] = time_spent
    if story_points is not None:
        body['story_points'] = story_points
    if external_issue_url is not None:
        body['external_issue_url'] = external_issue_url
    if external_issue_ref is not None:
        body['external_issue_ref'] = external_issue_ref
    if epic_id is not _UNSET:
        body['epic_id'] = epic_id
    return body


def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        column=dict(type='str', required=False),
        title=dict(type='str', required=True),
        description=dict(type='str', required=False),
        priority=dict(
            type='str',
            required=False,
            default=None,
            choices=list(_PRIORITY_CHOICES),
        ),
        start_date=dict(type='str', required=False),
        due_date=dict(type='str', required=False),
        assignee=dict(type='str', required=False),
        card_number=dict(type='str', required=False),
        closed=dict(type='bool', required=False),
        time_spent=dict(type='int', required=False),
        story_points=dict(type='int', required=False),
        external_issue_url=dict(type='str', required=False),
        external_issue_ref=dict(type='str', required=False),
        epic_id=dict(type='str', required=False),
        move_to_column=dict(type='str', required=False),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    project = module.params['project']
    column_name = module.params['column']
    title = module.params['title']
    description = module.params['description']
    priority = module.params['priority']
    start_date_raw = module.params['start_date']      # 'YYYY-MM-DD', 'null', or None (unset)
    due_date_raw = module.params['due_date']
    assignee_username = module.params['assignee']
    card_number = module.params['card_number']
    closed_raw = module.params['closed']
    closed = closed_raw if closed_raw is not None else _UNSET
    time_spent = module.params['time_spent']
    story_points = module.params['story_points']
    external_issue_url = module.params['external_issue_url']
    external_issue_ref = module.params['external_issue_ref']
    epic_id_raw = module.params['epic_id']
    move_to_column = module.params['move_to_column']
    state = module.params['state']

    # Normalise dates: 'null' string → None means "clear"; omitted stays _UNSET
    def _norm_date(raw):
        if raw is None:
            return _UNSET
        if raw.lower() == 'null':
            return None
        return raw

    start_date = _norm_date(start_date_raw)
    due_date = _norm_date(due_date_raw)

    # epic_id follows the same 'null'-string-clears / omitted-stays-_UNSET
    # convention as the date fields, but holds a numeric ID.
    if epic_id_raw is None:
        epic_id = _UNSET
    elif epic_id_raw.lower() == 'null':
        epic_id = None
    else:
        try:
            epic_id = int(epic_id_raw)
        except ValueError:
            module.fail_json(msg='epic_id must be an integer string or "null", got %r' % epic_id_raw)

    try:
        client = WarmDeskClient.from_module(module)
    except WarmDeskAPIError as exc:
        module.fail_json(msg=str(exc))

    # ----------------------------------------------------------------
    # Resolve assignee username → user ID (if provided)
    # ----------------------------------------------------------------
    assignee_id = _UNSET
    if assignee_username:
        try:
            assignee_id = resolve_user_id(client, assignee_username, project_slug=project)
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Cannot resolve assignee "%s": %s (HTTP %s)' % (
                    assignee_username, exc.message, exc.status)
            )

    # ----------------------------------------------------------------
    # Locate existing card
    # ----------------------------------------------------------------
    existing = None

    if card_number:
        try:
            existing = find_card_by_number(client, project, card_number)
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Error looking up card "%s": %s (HTTP %s)' % (
                    card_number, exc.message, exc.status)
            )
    elif column_name:
        try:
            col_id = resolve_column_id(client, project, column_name)
        except WarmDeskAPIError as exc:
            if state == 'absent':
                # Column gone → card is effectively absent
                module.exit_json(changed=False, card=None)
            module.fail_json(
                msg='Cannot resolve column "%s": %s (HTTP %s)' % (
                    column_name, exc.message, exc.status)
            )
        existing = _find_card_by_title(client, project, col_id, title)

    # ------------------------------------------------------------------ absent
    if state == 'absent':
        if existing is None:
            module.exit_json(changed=False, card=None)
        if module.check_mode:
            module.exit_json(changed=True, card=None)
        try:
            client.delete('/projects/%s/cards/%d' % (project, existing['id']))
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to delete card: %s (HTTP %s)' % (exc.message, exc.status)
            )
        module.exit_json(changed=True, card=None)

    # ----------------------------------------------------------------- present
    changed = False
    result_card = existing

    if existing is None:
        # ----- Create -----
        if not column_name:
            module.fail_json(
                msg='Parameter "column" is required when creating a new card.'
            )
        try:
            col_id = resolve_column_id(client, project, column_name)
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Cannot resolve column "%s": %s (HTTP %s)' % (
                    column_name, exc.message, exc.status)
            )

        if module.check_mode:
            module.exit_json(changed=True, card=None)

        create_body = {'title': title}
        if description is not None:
            create_body['description'] = description
        if priority:
            create_body['priority'] = priority
        if start_date is not _UNSET:
            create_body['start_date'] = start_date  # None → null in JSON
        if due_date is not _UNSET:
            create_body['due_date'] = due_date
        if assignee_id is not _UNSET:
            create_body['assignee_id'] = assignee_id

        try:
            result_card = client.post(
                '/projects/%s/columns/%d/cards' % (project, col_id), create_body
            )
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Failed to create card "%s": %s (HTTP %s)' % (title, exc.message, exc.status)
            )
        changed = True

        # closed/time_spent/story_points/external_issue_url/external_issue_ref/
        # epic_id are not accepted by the create endpoint — apply them with an
        # immediate follow-up update when requested.
        post_create_body = _extra_fields_body(closed, time_spent, story_points,
                                              external_issue_url, external_issue_ref, epic_id)
        if post_create_body:
            try:
                result_card = client.put(
                    '/projects/%s/cards/%d' % (project, result_card['id']), post_create_body
                )
            except WarmDeskAPIError as exc:
                module.fail_json(
                    msg='Card created but failed to apply closed/time_spent/story_points/'
                        'external_issue_url/external_issue_ref/epic_id: %s (HTTP %s)'
                        % (exc.message, exc.status)
                )

    else:
        # ----- Update if needed -----
        if _card_needs_update(existing, title, description, priority,
                              start_date, due_date, assignee_id, closed,
                              time_spent, story_points, external_issue_url,
                              external_issue_ref, epic_id):
            if module.check_mode:
                module.exit_json(changed=True, card=existing)

            body = {'title': title}
            if description is not None:
                body['description'] = description
            if priority:
                body['priority'] = priority
            if start_date is not _UNSET:
                body['start_date'] = start_date
            if due_date is not _UNSET:
                body['due_date'] = due_date
            if assignee_id is not _UNSET:
                body['assignee_id'] = assignee_id
            body.update(_extra_fields_body(closed, time_spent, story_points,
                                           external_issue_url, external_issue_ref, epic_id))

            try:
                result_card = client.put(
                    '/projects/%s/cards/%d' % (project, existing['id']), body
                )
            except WarmDeskAPIError as exc:
                module.fail_json(
                    msg='Failed to update card: %s (HTTP %s)' % (exc.message, exc.status)
                )
            changed = True

    # ----------------------------------------------------------------
    # Optional move
    # ----------------------------------------------------------------
    if move_to_column and result_card is not None:
        try:
            target_col_id = resolve_column_id(client, project, move_to_column)
        except WarmDeskAPIError as exc:
            module.fail_json(
                msg='Cannot resolve move_to_column "%s": %s (HTTP %s)' % (
                    move_to_column, exc.message, exc.status)
            )

        current_col_id = result_card.get('column_id')
        if current_col_id != target_col_id:
            if module.check_mode:
                module.exit_json(changed=True, card=result_card)
            try:
                result_card = client.patch(
                    '/projects/%s/cards/%d/move' % (project, result_card['id']),
                    {'column_id': target_col_id, 'position': 65536},
                )
            except WarmDeskAPIError as exc:
                module.fail_json(
                    msg='Failed to move card to "%s": %s (HTTP %s)' % (
                        move_to_column, exc.message, exc.status)
                )
            changed = True

    # Enrich the returned card with a card_ref string (e.g. "GF00-4") so
    # callers can pass it directly as card_number in follow-up tasks.
    if result_card is not None:
        try:
            proj = client.get('/projects/%s' % project)
            key_prefix = proj.get('key_prefix', '')
            if key_prefix and result_card.get('card_number') is not None:
                result_card = dict(result_card)
                result_card['card_ref'] = '%s-%d' % (key_prefix, result_card['card_number'])
        except WarmDeskAPIError:
            pass  # non-fatal; proceed without card_ref

    module.exit_json(changed=changed, card=result_card)


def main():
    run_module()


if __name__ == '__main__':
    main()
