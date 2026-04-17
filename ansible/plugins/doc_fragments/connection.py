# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type


class ModuleDocFragment(object):

    DOCUMENTATION = r'''
options:
  warmdesk_url:
    description:
      - Base URL of the WarmDesk server (e.g. C(https://warmdesk.example.com)).
      - Can also be set via the C(WARMDESK_URL) environment variable.
    type: str
    required: true
    aliases: [url]
  warmdesk_api_key:
    description:
      - API key for authentication (C(X-API-Key) header).
      - Mutually exclusive with I(warmdesk_token) and
        I(warmdesk_username)/I(warmdesk_password).
      - Can also be set via the C(WARMDESK_API_KEY) environment variable.
    type: str
    no_log: true
    aliases: [api_key]
  warmdesk_token:
    description:
      - Pre-obtained JWT bearer token.
      - Mutually exclusive with I(warmdesk_api_key) and
        I(warmdesk_username)/I(warmdesk_password).
      - Can also be set via the C(WARMDESK_TOKEN) environment variable.
    type: str
    no_log: true
    aliases: [token]
  warmdesk_username:
    description:
      - Username for password authentication. Must be combined with
        I(warmdesk_password); the module performs a login and auto-refreshes
        on 401.
      - Can also be set via the C(WARMDESK_USERNAME) environment variable.
    type: str
    aliases: [username]
  warmdesk_password:
    description:
      - Password for password authentication. Must be combined with
        I(warmdesk_username).
      - Can also be set via the C(WARMDESK_PASSWORD) environment variable.
    type: str
    no_log: true
    aliases: [password]
  validate_certs:
    description:
      - If C(false), SSL certificate verification is disabled.
      - Only disable this for development or self-signed certificates.
    type: bool
    default: true
'''
