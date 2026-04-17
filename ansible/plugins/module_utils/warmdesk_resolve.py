# -*- coding: utf-8 -*-
# GNU General Public License v2.0+ (see COPYING or https://www.gnu.org/licenses/gpl-2.0.txt)
from __future__ import absolute_import, division, print_function
__metaclass__ = type

"""
Name-to-ID resolvers for the ansilabnl.warmdesk collection.

Modules that need a numeric ID (e.g. customer_id for a contract) call these
helpers rather than doing the lookup inline.  All resolvers raise
WarmDeskAPIError(404, …) when the named resource cannot be found.

Resolvers are intentionally simple: they fetch the relevant list and scan it.
WarmDesk has no server-side search for most resources, and the lists are
small enough that a linear scan is fine.
"""

from ansible_collections.ansilabnl.warmdesk.plugins.module_utils.warmdesk_api import (
    WarmDeskAPIError,
)


def resolve_user_id(client, username):
    """Return the numeric ID for *username*.

    Uses the non-admin /users endpoint so that project members (not just
    global admins) can look up user IDs for membership operations.
    """
    users = client.get('/users')
    for u in users:
        if u['username'] == username:
            return u['id']
    raise WarmDeskAPIError(404, 'User not found: %s' % username)


def resolve_customer_id(client, name):
    """Return the numeric ID of the customer whose name equals *name*."""
    customers = client.get('/customers')
    for c in customers:
        if c['name'] == name:
            return c['id']
    raise WarmDeskAPIError(404, 'Customer not found: %s' % name)


def resolve_contract_id(client, customer_id, name):
    """Return the numeric ID of the contract with *name* under *customer_id*."""
    contracts = client.get('/customers/%d/contracts' % customer_id)
    for c in contracts:
        if c['name'] == name:
            return c['id']
    raise WarmDeskAPIError(404, 'Contract "%s" not found for customer %d' % (name, customer_id))


def resolve_column_id(client, project_slug, column_name):
    """Return the numeric ID of the column named *column_name* in *project_slug*."""
    columns = client.get('/projects/%s/columns' % project_slug)
    for col in columns:
        if col['name'] == column_name:
            return col['id']
    raise WarmDeskAPIError(404, 'Column "%s" not found in project %s' % (column_name, project_slug))


def resolve_label_id(client, project_slug, label_name):
    """Return the numeric ID of the label named *label_name* in *project_slug*."""
    labels = client.get('/projects/%s/labels' % project_slug)
    for lbl in labels:
        if lbl['name'] == label_name:
            return lbl['id']
    raise WarmDeskAPIError(404, 'Label "%s" not found in project %s' % (label_name, project_slug))


def find_card_by_number(client, project_slug, card_ref):
    """Return a card dict given a card reference like 'EDA-42'.

    Iterates all columns in the project to locate the card.  Returns None if
    not found (so callers can decide whether to fail or treat it as absent).
    """
    project = client.get('/projects/%s' % project_slug)
    for col in project.get('columns', []):
        for card in col.get('cards', []):
            ref = '%s-%d' % (project.get('key_prefix', ''), card['card_number'])
            if ref == card_ref:
                return client.get('/projects/%s/cards/%d' % (project_slug, card['id']))
    return None
