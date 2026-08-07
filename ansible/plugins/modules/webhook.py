# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: webhook
short_description: Manage incoming webhooks for a WarmDesk project
version_added: "0.1.0"
author: "Ton Kersten (@tonk)"
description:
  - Create or delete incoming webhook integrations on a WarmDesk project.
  - An incoming webhook posts a bot message to the project's chat channel when
    an HTTP POST is made to its unique URL.
  - Idempotency key is C(project) + C(name).  A second run with the same
    C(name) reports no change unless C(regenerate_token=true) is also set.
  - When C(regenerate_token=true) and the webhook already exists, a new token
    is generated and returned; the module always reports C(changed=true) in
    that case.
  - Supports check mode.
notes:
  - Requires the authenticated user to be a project owner or a global admin.
  - The plain-text token is only returned at creation time or after a token
    regeneration.  On subsequent idempotent runs C(token) is absent from the
    return value.
  - WarmDesk webhooks support the following types — C(generic) (default),
    C(gitea), C(github), C(gitlab).

extends_documentation_fragment:
  - ansilabnl.warmdesk.connection

options:
  project:
    description:
      - Slug of the project that owns this webhook.
    type: str
    required: true

  name:
    description:
      - Display name of the webhook.  Used as the idempotency key together
        with C(project).
    type: str
    required: true

  type:
    description:
      - Webhook flavour that controls how the incoming payload is interpreted.
      - C(generic) — plain JSON body with C(text) and optional C(username) fields.
      - C(gitea) / C(github) / C(gitlab) — parse the respective platform's
        event payload and format the chat message accordingly.
    type: str
    choices: [generic, gitea, github, gitlab]
    default: generic

  regenerate_token:
    description:
      - When C(true) and the webhook already exists, request a new secret
        token from the API and return it.
      - This is a one-shot action; the module always sets C(changed=true) when
        it executes the regeneration call.
      - Has no effect during creation (a fresh token is always returned).
    type: bool
    default: false

  state:
    description:
      - C(present) ensures the webhook exists with the specified name and type.
      - C(absent) removes the webhook if it exists.
    type: str
    choices: [present, absent]
    default: present
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Create a generic incoming webhook for CI/CD notifications
# ---------------------------------------------------------------------------
- name: Create CI notification webhook
  ansilabnl.warmdesk.webhook:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: eda-00
    name: CI Bot
    type: generic
    state: present
  register: hook_result

- name: Show the one-time token
  ansible.builtin.debug:
    msg: "Webhook token (save this now): {{ hook_result.token }}"
  when: hook_result.token is defined

# ---------------------------------------------------------------------------
# Create a Gitea webhook integration
# ---------------------------------------------------------------------------
- name: Ensure Gitea webhook exists on the platform project
  ansilabnl.warmdesk.webhook:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: platform
    name: Gitea Events
    type: gitea
    state: present

# ---------------------------------------------------------------------------
# Rotate the secret token for an existing webhook
# ---------------------------------------------------------------------------
- name: Regenerate token for the CI Bot webhook
  ansilabnl.warmdesk.webhook:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: eda-00
    name: CI Bot
    regenerate_token: true
  register: rotated

- name: Store the new token
  ansible.builtin.debug:
    msg: "New token: {{ rotated.token }}"

# ---------------------------------------------------------------------------
# Remove a webhook
# ---------------------------------------------------------------------------
- name: Delete the old deploy hook
  ansilabnl.warmdesk.webhook:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_api_key }}"
    project: eda-00
    name: Deploy Hook
    state: absent
"""

RETURN = r"""
changed:
  description: Whether the module made any change.
  returned: always
  type: bool

webhook:
  description:
    - The webhook object as returned by the WarmDesk API.
    - Contains C(id), C(name), C(type), C(token_hint), C(project_id), C(created_at).
    - C(null) when C(state=absent) and the webhook was deleted or did not exist.
  returned: always
  type: dict
  contains:
    id:
      description: Numeric webhook ID.
      returned: always
      type: int
      sample: 7
    name:
      description: Display name of the webhook.
      returned: always
      type: str
      sample: CI Bot
    type:
      description: Webhook type (generic, gitea, github, gitlab).
      returned: always
      type: str
      sample: generic
    token_hint:
      description: Last 8 characters of the current secret token (safe to log).
      returned: always
      type: str
      sample: a3f9c1d2
    project_id:
      description: ID of the owning project.
      returned: always
      type: int
      sample: 3
    created_at:
      description: ISO-8601 creation timestamp.
      returned: always
      type: str
      sample: "2025-03-10T14:22:00Z"

token:
  description:
    - The plain-text secret token for the webhook.
    - Only present immediately after creation or a successful token regeneration.
    - This value cannot be retrieved again — store it securely.
  returned: when webhook created or token regenerated
  type: str
  sample: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _find_webhook(client, project_slug, name):
    """Return the first webhook dict matching *name* in *project_slug*, or None."""
    hooks = client.get('/projects/%s/webhooks' % project_slug)
    for h in hooks:
        if h.get('name') == name:
            return h
    return None


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        project=dict(type='str', required=True),
        name=dict(type='str', required=True),
        type=dict(
            type='str',
            default='generic',
            choices=['generic', 'gitea', 'github', 'gitlab'],
        ),
        regenerate_token=dict(type='bool', default=False),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    state = p['state']
    project_slug = p['project']
    client = WarmDeskClient.from_module(module)

    try:
        existing = _find_webhook(client, project_slug, p['name'])

        # ------------------------------------------------------------------ #
        # state=absent                                                         #
        # ------------------------------------------------------------------ #
        if state == 'absent':
            if existing is None:
                module.exit_json(changed=False, webhook=None)
            if not module.check_mode:
                client.delete('/projects/%s/webhooks/%d' % (project_slug, existing['id']))
            module.exit_json(changed=True, webhook=None)

        # ------------------------------------------------------------------ #
        # state=present — CREATE                                               #
        # ------------------------------------------------------------------ #
        if existing is None:
            if module.check_mode:
                module.exit_json(changed=True, webhook=None)
            result = client.post(
                '/projects/%s/webhooks' % project_slug,
                dict(name=p['name'], type=p['type']),
            )
            # The creation response includes the full plain-text token once.
            plain_token = result.pop('token', None)
            ret = dict(changed=True, webhook=result)
            if plain_token:
                ret['token'] = plain_token
            module.exit_json(**ret)

        # ------------------------------------------------------------------ #
        # state=present — existing webhook                                     #
        # ------------------------------------------------------------------ #
        if p['regenerate_token']:
            if module.check_mode:
                module.exit_json(changed=True, webhook=existing)
            regen = client.post(
                '/projects/%s/webhooks/%d/regenerate' % (project_slug, existing['id'])
            )
            # /regenerate only returns {token, token_hint} — merge the fresh
            # hint into the pre-rotation webhook dict so callers reading
            # webhook.token_hint after rotation see the current value, not
            # the one that just became invalid.
            webhook = dict(existing)
            if regen and 'token_hint' in regen:
                webhook['token_hint'] = regen['token_hint']
            ret = dict(changed=True, webhook=webhook)
            if regen and 'token' in regen:
                ret['token'] = regen['token']
                module.warn(
                    'Webhook token has been regenerated. '
                    'Update any CI/CD systems that use the old token.'
                )
            module.exit_json(**ret)

        # Idempotent — no change.
        module.exit_json(changed=False, webhook=existing)

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
