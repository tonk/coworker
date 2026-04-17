#!/bin/bash

# Determine the program name and the 'running directory'
IAM="${0##*/}"
CRD="$(dirname $(realpath "${0}"))"

cd ${CRD} || {
	echo "Could not change to the WarmDesk directory"
	exit 1
}

# Build all
make clean all
cd dist

# Fill with demo data
./warmdesk-seed --reset
./warmdesk-training --reset
./warmdesk-training 4 Salami

# Create simple config
cat << '@EOF'
---
port: "8080"
db_driver: sqlite
db_dsn: "./warmdesk.db"

smtp:
  host: ""             # e.g. mail.example.com or smtp.gmail.com
  port: 587            # 587 for STARTTLS (recommended), 465 for SSL, 25 for plaintext
  from: ""             # e.g. warmdesk@example.com
  username: ""         # SMTP login (often the same as the from address)
  password: ""         # SMTP password or app-specific password

jwt_secret: "change-me-for-production"
allowed_origins: "http://localhost:8080"
web_dir: "web"

# All logging in debug mode
gin_mode: "debug"
db_log: "info"
api_log: true
@EOF

# And run
./warmdesk

