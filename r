#!/bin/bash
#
# Small program to build, seed, training seed, configure
# and run WarmDesk.
#
#
# Screenshots:
#   Start the app with:
#    chromium-browser --window-size=1400,900 --app="http://localhost:8080" &>/dev/null &
#    chromium-browser --window-size=1400,900 --app="http://localhost:8080" --ozone-platform=x11 &>/dev/null &
#
#   Shoot with:
#     sleep 3; maim --hidecursor -u  -i "$(xdotool getactivewindow)" <filename>
#
#
# Just for development.
#

# Determine the program name and the 'running directory'
IAM="${0##*/}"
CRD="$(dirname $(realpath "${0}"))"

cd ${CRD} || {
	echo "Could not change to the WarmDesk directory"
	exit 1
}

if [[ x"${1:-}" == x"run" ]]
then
	# Run
	cd dist
	./warmdesk
else
	# Build all
	make all || {
		echo "Building WarmDesk failed" >&2
		exit 1
	}

	# Fill with demo data
	cd dist
	./warmdesk-seed --reset
	./warmdesk-training --reset
	#./warmdesk-training 4 Salami

	# Create simple config
	cat <<- '@EOF' > warmdesk.yaml
		---
		base_url: https://warmdesk.example.com
		port: "8080"
		db_driver: sqlite
		db_dsn: "./warmdesk.db"

		smtp:
		  host: "master.tonkersten.com"
		  port: 25
		  from: ""
		  username: ""
		  password: ""

		jwt_secret: "change-me-for-production-and-some-for-the-minimum-length"
		allowed_origins: "http://localhost:8080"
		#web_dir: "web"

		# All logging in debug mode
		gin_mode: "debug"
		db_log: "info"
		api_log: true
	@EOF

	# And run
	./warmdesk
fi
