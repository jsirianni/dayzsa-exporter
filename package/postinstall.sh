#!/bin/bash

# Install handles systemd service management for debian and rhel based
# platforms. This function can be called more than once as it is idempotent.
install() {
    # Set the permission of the prometheus directory to 0700
    chmod 0700 /var/lib/bindplane/prometheus

    # Set the permission of the prometheus.yml file to 0600
    chmod 0600 /var/lib/bindplane/prometheus/prometheus.yml

    # Change the owner/group of all prometheus files to bindplane:bindplane
    chown -R bindplane:bindplane /var/lib/bindplane/prometheus

    systemctl daemon-reload
}

# Upgrade performs the same steps as install.
upgrade() {
    install
}

action="$1"

echo "Running postinstall with action: $action"

case "$action" in
  "0" | "install")
    install
    ;;
  "1" | "upgrade")
    upgrade
    ;;
  *)
    echo "Unknown action: $action. Defaulting to install."
    install
    ;;
esac
