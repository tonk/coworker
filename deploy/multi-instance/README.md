# Multi-instance deployment (3 nodes, 1 shared database)

This directory has templates for running 3 interchangeable WarmDesk app
instances behind a load balancer, all backed by one shared database, one
shared Redis, and one shared upload storage volume. Use this when a single
instance is no longer enough capacity for your team, or you want redundancy
if one app node goes down.

```
                      ┌──────────────────────┐
        HTTPS         │   Load balancer      │
   ───────────────►   │  (nginx-lb.conf)     │
                      └──────┬───┬───┬───────┘
                    round-robin, no sticky sessions needed
                             │   │   │
                 ┌───────────┘   │   └───────────┐
                 ▼               ▼               ▼
          warmdesk-1        warmdesk-2        warmdesk-3
        (warmdesk.service, identical warmdesk.yaml on each)
                 │                │                │
                 └───────┬────────┴────────┬───────┘
                         ▼                 ▼
                 PostgreSQL/MySQL        Redis
                  (db_dsn, TLS)      (redis_url — WS fan-out)
                         │
                 shared upload_dir (NFS/shared volume — fstab.example)
```

## Why each piece is required

- **Shared PostgreSQL or MySQL, not SQLite.** SQLite is a single-writer file
  database and cannot be shared across processes/hosts. All 3 instances point
  at the same `db_dsn`. GORM's `AutoMigrate` (which runs on every startup) is
  idempotent, so 3 instances starting up against the same schema is safe.
- **`redis_url` set on every instance.** Each instance's WebSocket hub is
  in-process memory only. Without Redis, a board/chat/presence update made by
  a user connected to instance 2 never reaches a teammate connected to
  instance 1 or 3. This is not optional in this topology.
- **Shared `upload_dir`.** Attachments, avatars, and logos are plain files on
  disk referenced by path — not stored in the database. All 3 instances must
  mount the *same* shared storage at the *same* path, or a file uploaded via
  one instance 404s when a different instance happens to serve the download.
  See `fstab.example` for an NFS mount.
- **One shared `jwt_secret` across all instances.** A token issued by
  instance 1 must validate on instances 2 and 3.
- **No sticky sessions needed at the load balancer.** Redis pub/sub is what
  makes any instance able to deliver any broadcast, so plain round-robin is
  fine — see `nginx-lb.conf`.

## Setup steps

1. **Provision the shared backing services** (not covered by these
   templates): a PostgreSQL or MySQL server, a Redis instance, and a shared
   filesystem mount (NFS export or equivalent) reachable from all 3 app hosts.
   If the database isn't on the same host as the app instances — which it
   won't be here — set `db_tls_mode: verify-full` with a CA cert; don't run a
   remote database over plaintext.

2. **Generate one JWT secret** and keep it somewhere you can copy from
   securely (a secrets manager, or at minimum copy it once and never
   regenerate it independently per host):
   ```bash
   openssl rand -hex 32
   ```

3. **Fill in `warmdesk.yaml.example`** in this directory (DB DSN, Redis URL,
   JWT secret, upload path, hostnames, trusted proxy). Deploy the *exact same
   file* as `/etc/warmdesk/warmdesk.yaml` on all 3 app hosts — there is no
   per-instance configuration in this setup. (It's named `.example` here only
   because `warmdesk.yaml` is git-ignored repo-wide, to keep real runtime
   configs with secrets out of version control.)

4. **Mount shared upload storage** on all 3 hosts at the path referenced by
   `upload_dir` in `warmdesk.yaml`. See `fstab.example`.

5. **Install `warmdesk.service`** on each of the 3 app hosts (same unit file
   everywhere), then on each host:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now warmdesk
   ```

6. **Configure the load balancer** using `nginx-lb.conf` — update the
   `upstream warmdesk_backend` block with your actual 3 hostnames/addresses,
   and set `trusted_proxies` in `warmdesk.yaml` to the load balancer's address
   so real client IPs are logged and rate-limited correctly instead of the
   LB's own IP.

7. **Scheduled backups — enable on exactly one instance.** `backup_schedule`
   is a runtime system setting (Admin → Backup / Restore), not a YAML key, so
   it isn't set by the shared config file. If you enable it on more than one
   instance, you'll get duplicate/overlapping backup runs. Pick one instance
   as the designated "backup node" and enable the schedule there only; the
   others can still take manual on-demand backups if needed.

## Known limitation: auth rate limiting is per-instance

The login/register/password-reset rate limiter (10 requests / 15 min per IP)
tracks attempts in-memory, per instance — there is currently no Redis-backed
shared limiter. With 3 instances behind round-robin, an attacker spreading
requests across all 3 effectively gets ~3x the intended budget before any
single instance's limiter engages. This is a reasonable tradeoff for an
internal team on a trusted network; if this deployment will be exposed to the
open internet at real scale, treat this as a gap worth closing (e.g. an
external WAF/rate-limiting layer in front of the load balancer) rather than
something these instances handle for you.

## Verifying it worked

- Log in as the same user from two different browsers/devices, connect to
  different instances (you can force this temporarily by pointing each
  browser at an instance's address directly instead of the LB), and confirm a
  board update made in one appears live in the other — this exercises the
  Redis WebSocket fan-out end to end.
- Upload an attachment via one browser session, then fetch its download URL
  from a session pinned to a different instance, to confirm the shared
  upload storage is mounted correctly everywhere.
