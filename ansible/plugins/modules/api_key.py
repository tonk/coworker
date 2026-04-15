# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: api_key
short_description: Manage personal and project-scoped WarmDesk API keys
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Create or revoke personal API keys (C(scope=personal)) or project-scoped
    API keys (C(scope=project)).
  - Idempotency key is C(scope) + C(name) for personal keys, and C(scope) +
    C(project) + C(name) for project keys.
  - When a key with the given name already exists, the module reports
    C(changed=false) and does B(not) return the plain-text key — it is
    irrecoverable after initial creation.
  - The plain-text API key is only available in the C(api_key) return field
    immediately after creation.  Store it in a vault or secret manager;
    subsequent runs omit C(api_key) entirely.
  - Supports check mode.
notes:
  - Personal keys authenticate as the user whose credentials are used to call
    this module.  They carry the same permissions as that user.
  - Project keys are scoped to a single project and authenticate via the
    C(X-API-Key) header or C(?api_key=) query parameter on the Ticket API.
  - Revoking a key is permanent.  To rotate a key, delete it (C(state=absent))
    and re-create it in a subsequent task; the new plain-text key will be
    returned at creation time.

extends_documentation_fragment:
  - ansilab.warmdesk.auth

options:
  scope:
    description:
      - C(personal) — key belongs to the currently authenticated user and is
        not tied to any specific project.
      - C(project) — key is scoped to the project identified by C(project).
    type: str
    choices: [personal, project]
    required: true

  project:
    description:
      - Slug of the project when C(scope=project).
      - Ignored when C(scope=personal).
    type: str

  name:
    description:
      - Human-readable label for the key.
      - Used as the idempotency key together with C(scope) (and C(project) for
        project-scoped keys).
    type: str
    required: true

  description:
    description:
      - Optional free-text description of what this key is used for.
      - Stored as the key's C(name) field in the API (WarmDesk uses a single
        C(name) field; the C(description) option is mapped to it as an
        additional label prefix when both are supplied, or used directly when
        only description is set).
      - Note: The backend currently stores only the C(name) field; this
        parameter is accepted for forward compatibility but may not appear in
        the API response depending on the WarmDesk version.
    type: str

  state:
    description:
      - C(present) ensures the key exists.
      - C(absent) revokes the key if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a personal API key and save the token to a variable
# ---------------------------------------------------------------------------
- name: Create personal API key for CI pipeline
  ansilab.warmdesk.api_key:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_username: "{{ wd_admin_user }}"
    warmdesk_password: "{{ vault_admin_password }}"
    scope: personal
    name: ci-pipeline-key
    description: Used by the GitLab CI runner
    state: present
  register: key_result

- name: Store the API key (only available now!)
  ansible.builtin.debug:
    msg: "New key: {{ key_result.api_key }}"
  when: key_result.api_key is defined

# ---------------------------------------------------------------------------
# Create a project-scoped API key for ticket automation
# ---------------------------------------------------------------------------
- name: Create project API key for ticket bot
  ansilab.warmdesk.api_key:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_key }}"
    scope: project
    project: eda-00
    name: ticket-bot
    state: present
  register: project_key

- name: Configure the ticket bot with its new key
  ansible.builtin.template:
    src: ticket-bot.conf.j2
    dest: /etc/ticket-bot/config.conf
  vars:
    bot_api_key: "{{ project_key.api_key | default('') }}"
  when: project_key.api_key is defined

# ---------------------------------------------------------------------------
# Idempotent — second run does not change anything and api_key is absent
# ---------------------------------------------------------------------------
- name: Ensure project key exists (no-op if already present)
  ansilab.warmdesk.api_key:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_key }}"
    scope: project
    project: eda-00
    name: ticket-bot
    state: present
  # api_key will NOT be in the return value — the key already existed.

# ---------------------------------------------------------------------------
# Revoke a personal key
# ---------------------------------------------------------------------------
- name: Revoke the old CI key
  ansilab.warmdesk.api_key:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_key }}"
    scope: personal
    name: old-ci-key
    state: absent

# ---------------------------------------------------------------------------
# Revoke a project-scoped key
# ---------------------------------------------------------------------------
- name: Revoke project key
  ansilab.warmdesk.api_key:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_key }}"
    scope: project
    project: platform
    name: deploy-bot
    state: absent
"""

RETURN = r"""
changed:
  description: Whether the module made any change.
  returned: always
  type: bool

key:
  description:
    - The API key metadata object as returned by the WarmDesk API.
    - C(null) when C(state=absent) and the key was deleted or did not exist.
    - Does B(not) include the plain-text key value — see C(api_key) for that.
  returned: always
  type: dict
  contains:
    id:
      description: Numeric key ID.
      returned: always
      type: int
      sample: 12
    name:
      description: Human-readable label for the key.
      returned: always
      type: str
      sample: ci-pipeline-key
    key_prefix:
      description: First 12 characters of the key (safe to log; used for identification).
      returned: always
      type: str
      sample: "cwk_a1b2c3d4e5"
    project_id:
      description: Project ID for project-scoped keys; null for personal keys.
      returned: always
      type: int
      sample: 5
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2025-06-01T08:00:00Z"

api_key:
  description:
    - The full plain-text API key (e.g. C(cwk_a1b2c3d4…)).
    - B(Only present immediately after creation.)  This value is not stored by
      WarmDesk and cannot be retrieved again.  Save it to a vault immediately.
  returned: when key is newly created
  type: str
  sample: "cwk_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _list_path(scope, project_slug):
    if scope == 'project':
        return '/projects/%s/api-keys' % project_slug
    return '/auth/api-keys'


def _delete_path(scope, project_slug, key_id):
    if scope == 'project':
        return '/projects/%s/api-keys/%d' % (project_slug, key_id)
    return '/auth/api-keys/%d' % key_id


def _find_key(client, scope, project_slug, name):
    """Return the key dict matching *name*, or None."""
    keys = client.get(_list_path(scope, project_slug))
    for k in keys:
        if k.get('name') == name:
            return k
    return None


def _build_name(p):
    """Combine name and optional description into the API 'name' field."""
    if p.get('description'):
        return '%s — %s' % (p['name'], p['description'])
    return p['name']


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        scope=dict(type='str', required=True, choices=['personal', 'project']),
        project=dict(type='str'),
        name=dict(type='str', required=True),
        description=dict(type='str'),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ('scope', 'project', ('project',)),
        ],
    )

    p = module.params
    scope = p['scope']
    project_slug = p.get('project') or ''
    state = p['state']
    client = WarmDeskClient.from_module(module)

    try:
        existing = _find_key(client, scope, project_slug, p['name'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, key=None)
            if not module.check_mode:
                client.delete(_delete_path(scope, project_slug, existing['id']))
            module.exit_json(changed=True, key=None)

        # ------------------------------------------------------------------ #
        # state=present — already exists                                       #
        # ------------------------------------------------------------------ #
        if existing is not None:
            # The plain-text key is irrecoverable — do not include it.
            module.exit_json(changed=False, key=existing)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if module.check_mode:
            module.exit_json(changed=True, key=None)

        body = dict(name=_build_name(p))
        result = client.post(_list_path(scope, project_slug), body)

        # The creation response includes the plain-text key once.
        plain_key = result.pop('key', None)

        ret = dict(changed=True, key=result)
        if plain_key:
            ret['api_key'] = plain_key
            module.warn(
                'API key created. The plain-text key is in the "api_key" '
                'return field and cannot be retrieved again — store it '
                'securely now.'
            )

        module.exit_json(**ret)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
