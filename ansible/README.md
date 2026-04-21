# ansilabnl.warmdesk

Ansible collection for managing [WarmDesk](https://github.com/tonk/warmdesk) — a self-hosted project management tool with Kanban boards, team chat, customers, and contracts.

## Requirements

- Ansible >= 2.14
- Python >= 3.9
- A running WarmDesk instance (v0.7+)

## Installation

```bash
ansible-galaxy collection install ansilabnl.warmdesk
```

## Authentication

All modules and plugins accept the same connection parameters (or their environment variable equivalents):

| Parameter | Environment variable | Description |
|---|---|---|
| `warmdesk_url` | `WARMDESK_URL` | Base URL of the WarmDesk server |
| `warmdesk_api_key` | `WARMDESK_API_KEY` | API key (`X-API-Key` header) |
| `warmdesk_token` | `WARMDESK_TOKEN` | Pre-obtained JWT bearer token |
| `warmdesk_username` | `WARMDESK_USERNAME` | Username for password authentication |
| `warmdesk_password` | `WARMDESK_PASSWORD` | Password for password authentication |
| `validate_certs` | — | Set to `false` to disable SSL verification (default: `true`) |

Only one authentication method should be used at a time.

## Modules

| Module | Description |
|---|---|
| `ansilabnl.warmdesk.user` | Manage WarmDesk users |
| `ansilabnl.warmdesk.api_key` | Manage API keys for a user |
| `ansilabnl.warmdesk.project` | Manage projects |
| `ansilabnl.warmdesk.project_member` | Manage project membership and roles |
| `ansilabnl.warmdesk.column` | Manage Kanban columns in a project |
| `ansilabnl.warmdesk.label` | Manage card labels in a project |
| `ansilabnl.warmdesk.card` | Manage cards on a Kanban board |
| `ansilabnl.warmdesk.checklist_item` | Manage checklist items on a card |
| `ansilabnl.warmdesk.customer` | Manage customers |
| `ansilabnl.warmdesk.customer_member` | Manage customer access and roles |
| `ansilabnl.warmdesk.contract` | Manage contracts for a customer |
| `ansilabnl.warmdesk.group` | Manage user groups and their project/customer access (admin only) |
| `ansilabnl.warmdesk.webhook` | Manage webhooks |
| `ansilabnl.warmdesk.system_settings` | Manage global system settings (admin only) |

## Lookup plugins

| Plugin | Description |
|---|---|
| `ansilabnl.warmdesk.card` | Look up one or more cards by reference |
| `ansilabnl.warmdesk.project` | Look up projects by slug |
| `ansilabnl.warmdesk.customer` | Look up customers by name |
| `ansilabnl.warmdesk.contract` | Look up contracts by name |
| `ansilabnl.warmdesk.user` | Look up users by username |

## Inventory plugin

| Plugin | Description |
|---|---|
| `ansilabnl.warmdesk.warmdesk` | Dynamic inventory from WarmDesk projects and users |

## Example

```yaml
- name: Provision a new project with columns and members
  hosts: localhost
  gather_facts: false
  vars:
    warmdesk_url: https://warmdesk.example.com
    warmdesk_api_key: "{{ lookup('env', 'WARMDESK_API_KEY') }}"

  tasks:
    - name: Create project
      ansilabnl.warmdesk.project:
        warmdesk_url: "{{ warmdesk_url }}"
        warmdesk_api_key: "{{ warmdesk_api_key }}"
        name: My Project
        state: present
      register: project

    - name: Add columns
      ansilabnl.warmdesk.column:
        warmdesk_url: "{{ warmdesk_url }}"
        warmdesk_api_key: "{{ warmdesk_api_key }}"
        project: "{{ project.project.slug }}"
        name: "{{ item }}"
        state: present
      loop:
        - Backlog
        - In Progress
        - In Review
        - Done

    - name: Add a member
      ansilabnl.warmdesk.project_member:
        warmdesk_url: "{{ warmdesk_url }}"
        warmdesk_api_key: "{{ warmdesk_api_key }}"
        project: "{{ project.project.slug }}"
        username: jdoe
        role: member
        state: present
```

## License

GPL-2.0-or-later

## Author

Ton Kersten
