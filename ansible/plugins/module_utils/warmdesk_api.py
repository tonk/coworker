# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

"""
Shared HTTP client and argument-spec helper for the ansilab.warmdesk collection.

All WarmDesk modules import WarmDeskClient, WarmDeskAPIError and
warmdesk_argument_spec from here.

Auth priority (first non-empty wins):
  1. warmdesk_api_key  →  X-API-Key header  (ticket API + project key ops)
  2. warmdesk_token    →  Bearer token       (pre-obtained JWT)
  3. warmdesk_username + warmdesk_password   →  POST /auth/login, auto-refresh on 401

Environment-variable fallbacks (when the param is not set in the task):
  WARMDESK_URL, WARMDESK_TOKEN, WARMDESK_USERNAME, WARMDESK_PASSWORD,
  WARMDESK_API_KEY
"""

import json

from ansible.module_utils.basic import env_fallback
from ansible.module_utils.urls import open_url, ConnectionError, SSLValidationError
from ansible.module_utils.six.moves.urllib.error import HTTPError, URLError

_BASE = '/api/v1'


# ---------------------------------------------------------------------------
# Public helpers
# ---------------------------------------------------------------------------

def warmdesk_argument_spec():
    """Return the common connection arguments shared by all WarmDesk modules."""
    return dict(
        warmdesk_url=dict(
            type='str',
            required=True,
            aliases=['url'],
            fallback=(env_fallback, ['WARMDESK_URL']),
        ),
        warmdesk_token=dict(
            type='str',
            no_log=True,
            aliases=['token'],
            fallback=(env_fallback, ['WARMDESK_TOKEN']),
        ),
        warmdesk_username=dict(
            type='str',
            aliases=['username'],
            fallback=(env_fallback, ['WARMDESK_USERNAME']),
        ),
        warmdesk_password=dict(
            type='str',
            no_log=True,
            aliases=['password'],
            fallback=(env_fallback, ['WARMDESK_PASSWORD']),
        ),
        warmdesk_api_key=dict(
            type='str',
            no_log=True,
            aliases=['api_key'],
            fallback=(env_fallback, ['WARMDESK_API_KEY']),
        ),
        validate_certs=dict(type='bool', default=True),
    )


class WarmDeskAPIError(Exception):
    """Raised for any non-2xx response or connection failure."""

    def __init__(self, status, message):
        self.status = status
        self.message = message
        super(WarmDeskAPIError, self).__init__(message)


class WarmDeskClient(object):
    """Thin REST client for the WarmDesk API."""

    def __init__(self, url, username=None, password=None,
                 token=None, api_key=None, validate_certs=True):
        self.base_url = url.rstrip('/')
        self._username = username
        self._password = password
        self.token = token
        self.api_key = api_key
        self.validate_certs = validate_certs

    # ------------------------------------------------------------------
    # Construction helpers
    # ------------------------------------------------------------------

    @classmethod
    def from_module(cls, module):
        """Create a client from an AnsibleModule's params, failing fast on
        missing credentials."""
        p = module.params
        client = cls(
            url=p['warmdesk_url'],
            username=p.get('warmdesk_username'),
            password=p.get('warmdesk_password'),
            token=p.get('warmdesk_token'),
            api_key=p.get('warmdesk_api_key'),
            validate_certs=p.get('validate_certs', True),
        )
        has_creds = (
            client.api_key
            or client.token
            or (client._username and client._password)
        )
        if not has_creds:
            module.fail_json(
                msg='Provide warmdesk_api_key, warmdesk_token, '
                    'or warmdesk_username + warmdesk_password.'
            )
        return client

    # ------------------------------------------------------------------
    # Authentication
    # ------------------------------------------------------------------

    def _login(self):
        """Exchange username/password for a JWT access token."""
        data = json.dumps({
            'username': self._username,
            'password': self._password,
        }).encode('utf-8')
        try:
            resp = open_url(
                self.base_url + _BASE + '/auth/login',
                data=data,
                headers={'Content-Type': 'application/json',
                         'Accept': 'application/json'},
                method='POST',
                validate_certs=self.validate_certs,
            )
            body = json.loads(resp.read().decode('utf-8'))
            self.token = body['access_token']
        except HTTPError as exc:
            raise WarmDeskAPIError(exc.code, 'Login failed: ' + _http_error_msg(exc))

    def _auth_headers(self):
        headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
        }
        if self.api_key:
            headers['X-API-Key'] = self.api_key
        elif self.token:
            headers['Authorization'] = 'Bearer ' + self.token
        return headers

    # ------------------------------------------------------------------
    # Raw HTTP
    # ------------------------------------------------------------------

    def _raw(self, method, path, body=None, params=None):
        url = self.base_url + _BASE + path
        if params:
            from ansible.module_utils.six.moves.urllib.parse import urlencode
            url = url + '?' + urlencode(params)

        data = json.dumps(body).encode('utf-8') if body is not None else None

        try:
            resp = open_url(
                url,
                data=data,
                headers=self._auth_headers(),
                method=method,
                validate_certs=self.validate_certs,
            )
            raw = resp.read().decode('utf-8').strip()
            return json.loads(raw) if raw else None
        except HTTPError as exc:
            raise WarmDeskAPIError(exc.code, _http_error_msg(exc))
        except (ConnectionError, SSLValidationError, URLError) as exc:
            raise WarmDeskAPIError(0, str(exc))

    # ------------------------------------------------------------------
    # Public request method (handles auth + one-shot token refresh)
    # ------------------------------------------------------------------

    def request(self, method, path, body=None, params=None):
        if not self.api_key and not self.token:
            self._login()

        try:
            return self._raw(method, path, body, params)
        except WarmDeskAPIError as exc:
            if exc.status == 401 and self._username and self._password:
                self._login()
                return self._raw(method, path, body, params)
            raise

    # ------------------------------------------------------------------
    # Convenience wrappers
    # ------------------------------------------------------------------

    def get(self, path, params=None):
        return self.request('GET', path, params=params)

    def post(self, path, body=None):
        return self.request('POST', path, body)

    def put(self, path, body=None):
        return self.request('PUT', path, body)

    def patch(self, path, body=None):
        return self.request('PATCH', path, body)

    def delete(self, path):
        return self.request('DELETE', path)


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

def _http_error_msg(exc):
    """Extract the 'error' key from a WarmDesk JSON error body, or fall back
    to the HTTP reason string."""
    try:
        body = json.loads(exc.read().decode('utf-8'))
        return body.get('error', str(exc))
    except Exception:
        return str(exc)
