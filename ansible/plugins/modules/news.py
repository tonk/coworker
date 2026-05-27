# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: news
short_description: Manage WarmDesk dashboard news items
version_added: "0.3.1"
author: "Ton Kersten (@tonk)"
description:
  - Create, update, or delete news items that are shown as dismissible tiles
    on the WarmDesk dashboard.
  - Each active news item whose current date falls within its C(start_date) /
    C(end_date) window is displayed to all users on their dashboard.
  - Idempotency key is C(title).  A second run with the same C(title) reports
    no change unless a field value differs.
  - Supports check mode.
notes:
  - Requires an admin account.  Use an API key belonging to an admin user, or
    supply admin credentials via I(warmdesk_username)/I(warmdesk_password).
  - Date strings must be valid ISO-8601 (e.g. C(2026-06-01T00:00:00Z) or
    C(2026-06-01T00:00:00)).  Both are compared after normalisation, so minor
    formatting differences (trailing Z vs. +00:00, omitted seconds) do not
    cause spurious changes.
  - If multiple news items share the same C(title) the module operates on the
    first one returned by the API (oldest by creation date).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  title:
    description:
      - Title of the news item.  Used as the idempotency key.
    type: str
    required: true

  text:
    description:
      - Body text of the news item.  Displayed below the title on the
        dashboard tile.
      - Required when C(state=present).
    type: str

  start_date:
    description:
      - ISO-8601 date/time from which the news item becomes visible.
      - Leave unset (or C(null)) to make the item visible immediately.
    type: str

  end_date:
    description:
      - ISO-8601 date/time after which the news item is no longer shown.
      - Leave unset (or C(null)) to display the item indefinitely.
    type: str

  active:
    description:
      - When C(false) the item is stored but never shown on the dashboard,
        regardless of its date window.
    type: bool
    default: true

  state:
    description:
      - C(present) ensures the news item exists with the given field values.
      - C(absent) removes the news item if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Publish a maintenance notice from midnight tonight until end of day
# ---------------------------------------------------------------------------
- name: Announce scheduled maintenance
  ansilabnl.warmdesk.news:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    title: Scheduled maintenance — 2026-06-15
    text: >-
      The server will be unavailable on 15 June 2026 from 02:00 to 04:00 UTC
      for a database upgrade.  No data will be lost.
    start_date: "2026-06-14T00:00:00Z"
    end_date: "2026-06-15T23:59:59Z"
    active: true
    state: present

# ---------------------------------------------------------------------------
# Create a permanent news item (no date window)
# ---------------------------------------------------------------------------
- name: Announce new feature
  ansilabnl.warmdesk.news:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    title: Time tracking export now supports XLSX
    text: You can now export your time reports directly to Excel format.
    state: present

# ---------------------------------------------------------------------------
# Disable a news item without deleting it
# ---------------------------------------------------------------------------
- name: Hide the maintenance notice early
  ansilabnl.warmdesk.news:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    title: Scheduled maintenance — 2026-06-15
    text: >-
      The server will be unavailable on 15 June 2026 from 02:00 to 04:00 UTC
      for a database upgrade.  No data will be lost.
    active: false
    state: present

# ---------------------------------------------------------------------------
# Remove a news item permanently
# ---------------------------------------------------------------------------
- name: Remove old announcement
  ansilabnl.warmdesk.news:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    title: Time tracking export now supports XLSX
    state: absent

# ---------------------------------------------------------------------------
# Use username/password instead of an API key
# ---------------------------------------------------------------------------
- name: Post release notes
  ansilabnl.warmdesk.news:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: admin
    warmdesk_password: "{{ vault_admin_password }}"
    title: "WarmDesk v0.9.38 released"
    text: >-
      Version 0.9.38 adds admin-managed dashboard news, a format-aware date
      input, and the WarmDesk logo on news tiles.
    state: present
"""

RETURN = r"""
changed:
  description: Whether the module made any change.
  returned: always
  type: bool

news_item:
  description:
    - The news item object as stored in WarmDesk.
    - C(null) when C(state=absent) and the item was deleted or did not exist.
  returned: always
  type: dict
  contains:
    id:
      description: Numeric news item ID.
      type: int
      sample: 3
    title:
      description: Title of the news item.
      type: str
      sample: Scheduled maintenance — 2026-06-15
    text:
      description: Body text.
      type: str
    start_date:
      description: ISO-8601 visibility start, or null.
      type: str
      sample: "2026-06-14T00:00:00Z"
    end_date:
      description: ISO-8601 visibility end, or null.
      type: str
      sample: "2026-06-15T23:59:59Z"
    active:
      description: Whether the item is active.
      type: bool
      sample: true
    created_at:
      description: ISO-8601 creation timestamp.
      type: str
      sample: "2026-05-12T10:00:00Z"
    updated_at:
      description: ISO-8601 last-update timestamp.
      type: str
      sample: "2026-05-12T10:00:00Z"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Date normalisation — avoids spurious changes from formatting differences
# ---------------------------------------------------------------------------

def _norm_date(value):
    """Return a normalised ISO-8601 string, or None if value is falsy."""
    if not value:
        return None
    # Strip trailing Z or offset for comparison purposes; keep as naive UTC.
    s = str(value).strip()
    # Accept both "Z" and "+00:00" suffixes.
    for suffix in ('+00:00', '-00:00', 'Z'):
        if s.endswith(suffix):
            s = s[:-len(suffix)]
            break
    # Normalise to second precision (drop fractional seconds).
    if '.' in s:
        s = s[:s.index('.')]
    return s


def _dates_equal(a, b):
    return _norm_date(a) == _norm_date(b)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _find_news(client, title):
    """Return the first news item dict whose title matches, or None."""
    items = client.get('/admin/news')
    for item in items:
        if item.get('title') == title:
            return item
    return None


def _needs_update(existing, params):
    """Return True if any managed field differs from the desired state."""
    if params['text'] is not None and existing.get('text') != params['text']:
        return True
    if not _dates_equal(existing.get('start_date'), params['start_date']):
        return True
    if not _dates_equal(existing.get('end_date'), params['end_date']):
        return True
    if existing.get('active') != params['active']:
        return True
    return False


def _build_payload(params):
    return dict(
        title=params['title'],
        text=params['text'] or '',
        start_date=params['start_date'] or None,
        end_date=params['end_date'] or None,
        active=params['active'],
    )


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        title=dict(type='str', required=True),
        text=dict(type='str'),
        start_date=dict(type='str'),
        end_date=dict(type='str'),
        active=dict(type='bool', default=True),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        required_if=[
            ('state', 'present', ('text',)),
        ],
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']
    client = WarmDeskClient.from_module(module)

    try:
        existing = _find_news(client, p['title'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, news_item=None)
            if not module.check_mode:
                client.delete('/admin/news/%d' % existing['id'])
            module.exit_json(changed=True, news_item=None)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, news_item=None)
            created = client.post('/admin/news', _build_payload(p))
            module.exit_json(changed=True, news_item=created)

        # ------------------------------------------------------------------ #
        # state=present — UPDATE if needed                                     #
        # ------------------------------------------------------------------ #
        if not _needs_update(existing, p):
            module.exit_json(changed=False, news_item=existing)

        if module.check_mode:
            module.exit_json(changed=True, news_item=existing)

        updated = client.put('/admin/news/%d' % existing['id'], _build_payload(p))
        module.exit_json(changed=True, news_item=updated)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
