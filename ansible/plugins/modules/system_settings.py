# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

DOCUMENTATION = r"""
---
module: system_settings
short_description: Configure WarmDesk system-wide settings
version_added: "1.0.0"
author: "Ton Kersten (@tonk)"
description:
  - Read and update system-wide configuration for a WarmDesk instance via the
    admin API.
  - Accepts a C(settings) dict whose keys map directly to WarmDesk setting
    names.  Only the keys you supply are written; all others are left as-is.
  - Performs a GET before the PUT and only reports C(changed=true) when at
    least one of the requested values differs from the current value.
  - There is no C(state) parameter — this module sets values and never deletes
    individual keys (WarmDesk does not support key deletion).
  - Supports check mode.
notes:
  - Requires an admin account or a token obtained from an admin account.
  - The SMTP password is never returned by the API (only C(smtp_password_set)
    boolean is included).  Therefore, if you supply C(smtp_password) in
    C(settings), the module always treats it as changed because it cannot
    compare the current and desired ciphertext.
  - Boolean-like settings (e.g. C(registration_enabled)) must be passed as
    their native Python boolean type (C(true)/C(false)) not as strings.
  - Integer-like settings (e.g. C(smtp_port), C(session_timeout_minutes),
    C(password_min_length)) should be passed as integers.

extends_documentation_fragment:
  - ansilab.warmdesk.auth

options:
  settings:
    description:
      - Dict of setting key/value pairs to apply.
      - Supported keys and their expected types are listed below.
      - |
        B(String keys:)
        C(company_name), C(company_logo), C(default_columns) (newline-separated),
        C(default_labels) (newline-separated), C(default_timezone),
        C(default_date_time_format), C(default_theme) (C(light)|C(dark)|C(system)),
        C(default_font), C(default_font_size), C(default_locale)
        (C(en)|C(nl)|C(de)|C(fr)|C(es)), C(smtp_host), C(smtp_from),
        C(smtp_username), C(smtp_password).
      - |
        B(Integer keys:)
        C(smtp_port), C(session_timeout_minutes), C(password_min_length).
      - |
        B(Boolean keys:)
        C(registration_enabled), C(mfa_required), C(password_require_upper),
        C(password_require_lower), C(password_require_digit),
        C(password_require_special).
    type: dict
    required: true
"""

EXAMPLES = r"""
# ---------------------------------------------------------------------------
# Set basic branding and locale defaults
# ---------------------------------------------------------------------------
- name: Configure WarmDesk branding
  ansilab.warmdesk.system_settings:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    settings:
      company_name: Acme Corp
      default_locale: nl
      default_timezone: Europe/Amsterdam
      default_date_time_format: "DD-MM-YYYY HH:mm"
      default_theme: dark

# ---------------------------------------------------------------------------
# Configure SMTP (without touching password if already set)
# ---------------------------------------------------------------------------
- name: Configure SMTP relay
  ansilab.warmdesk.system_settings:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    settings:
      smtp_host: smtp.example.com
      smtp_port: 587
      smtp_from: warmdesk@example.com
      smtp_username: warmdesk-mailer
      smtp_password: "{{ vault_smtp_password }}"

# ---------------------------------------------------------------------------
# Set project defaults (columns and labels)
# ---------------------------------------------------------------------------
- name: Set default board columns and labels
  ansilab.warmdesk.system_settings:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    settings:
      default_columns: "Backlog\nIn Progress\nReview\nDone"
      default_labels: "Bug\nFeature\nImprovement\nSecurity"

# ---------------------------------------------------------------------------
# Harden security settings
# ---------------------------------------------------------------------------
- name: Enforce password policy and MFA
  ansilab.warmdesk.system_settings:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    settings:
      registration_enabled: false
      mfa_required: true
      password_min_length: 16
      password_require_upper: true
      password_require_lower: true
      password_require_digit: true
      password_require_special: true
      session_timeout_minutes: 30

# ---------------------------------------------------------------------------
# Check mode — preview what would change without applying
# ---------------------------------------------------------------------------
- name: Preview settings change
  ansilab.warmdesk.system_settings:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ vault_admin_api_key }}"
    settings:
      company_name: "New Corp Name"
  check_mode: true
  register: preview

- name: Show what would change
  ansible.builtin.debug:
    var: preview.diff
"""

RETURN = r"""
changed:
  description: Whether any setting value was updated.
  returned: always
  type: bool

settings:
  description:
    - The complete settings object as returned by the WarmDesk admin API
      after the update (or the current state in check mode / when unchanged).
    - The SMTP password is never included; C(smtp_password_set) (bool string)
      indicates whether one is configured.
  returned: always
  type: dict
  sample:
    company_name: Acme Corp
    default_locale: nl
    default_timezone: Europe/Amsterdam
    registration_enabled: "true"
    smtp_host: smtp.example.com
    smtp_port: "587"
    smtp_password_set: "false"

diff:
  description:
    - Human-readable summary of which keys changed and what the old/new values
      are.  Only present when C(changed=true).
  returned: when changed
  type: dict
  contains:
    before:
      description: Dict of keys that changed with their old values.
      type: dict
    after:
      description: Dict of keys that changed with their new values.
      type: dict
"""

from ansible.module_utils.basic import AnsibleModule
from ansible_collections.ansilab.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskClient,
    WarmDeskAPIError,
    warmdesk_argument_spec,
)

# Settings that are opaque (cannot be compared to desired value):
# the API returns a sentinel instead of the actual secret.
_OPAQUE_WRITE_ONLY = {'smtp_password'}

# Settings that the API accepts as boolean Python objects in the PUT body.
_BOOL_KEYS = {
    'registration_enabled',
    'mfa_required',
    'password_require_upper',
    'password_require_lower',
    'password_require_digit',
    'password_require_special',
}

# Settings that the API accepts as integers in the PUT body.
_INT_KEYS = {
    'smtp_port',
    'session_timeout_minutes',
    'password_min_length',
}


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _coerce(key, value):
    """Coerce a desired value to the type the API PUT endpoint expects."""
    if key in _BOOL_KEYS:
        if isinstance(value, bool):
            return value
        return str(value).lower() in ('true', '1', 'yes')
    if key in _INT_KEYS:
        return int(value)
    return value


def _api_repr(key, value):
    """Return the string representation the API stores for comparison.

    The GET endpoint returns all settings as strings (e.g. "true", "587").
    We normalise desired values to the same format for accurate diffing.
    """
    if key in _BOOL_KEYS:
        if isinstance(value, bool):
            return 'true' if value else 'false'
        return 'true' if str(value).lower() in ('true', '1', 'yes') else 'false'
    if key in _INT_KEYS:
        return str(int(value))
    return str(value)


def _diff(current, desired_settings):
    """Return (put_body, before_dict, after_dict).

    Compares each key in *desired_settings* against the current API state.
    Opaque/write-only keys are always included in the PUT body and marked as
    changed (since the current value is hidden from the API response).
    """
    put_body = {}
    before = {}
    after = {}

    for key, desired in desired_settings.items():
        if key in _OPAQUE_WRITE_ONLY:
            # Cannot compare — always treat as changed.
            put_body[key] = _coerce(key, desired)
            before[key] = '*** (hidden)'
            after[key] = '*** (hidden)'
            continue

        current_val = current.get(key, '')
        desired_str = _api_repr(key, desired)

        if current_val != desired_str:
            put_body[key] = _coerce(key, desired)
            before[key] = current_val
            after[key] = desired_str

    return put_body, before, after


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def run_module():
    argument_spec = warmdesk_argument_spec()
    argument_spec.update(dict(
        settings=dict(type='dict', required=True),
    ))

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    p = module.params
    client = WarmDeskClient.from_module(module)

    try:
        current = client.get('/admin/system')

        put_body, before, after = _diff(current, p['settings'])

        if not put_body:
            module.exit_json(changed=False, settings=current)

        if module.check_mode:
            module.exit_json(
                changed=True,
                settings=current,
                diff=dict(before=before, after=after),
            )

        updated = client.put('/admin/system', put_body)
        module.exit_json(
            changed=True,
            settings=updated,
            diff=dict(before=before, after=after),
        )

    except WarmDeskAPIError as e:
        module.fail_json(msg='WarmDesk API error %d: %s' % (e.status, e.message))


def main():
    run_module()


if __name__ == '__main__':
    main()
