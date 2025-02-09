#!/bin/bash

# This script must work on the following:
#  debian 10
#  debian 11
#  ubuntu 18.04
#  ubuntu 20.04
#  ubuntu 22.04
#  centos / rhel 7
#  centos / rhel 8
#  SLES 12
#  SLES 15

set -e

# Remove deletes the dayzsa systemd service file
# and reloads systemd. If the file does not exist, return early.
remove() {
    rm -f /usr/lib/systemd/system/dayzsa.service || return
    systemctl daemon-reload
}

# Upgrade performs a no-op and is included here for future use.
upgrade() {
    return
}

action="$1"

echo "Running postremove with action: $action"

case "$action" in
  "0" | "remove")
    remove
    ;;
  "1" | "upgrade")
    upgrade
    ;;
  *)
    echo "Unknown action: $action"
    ;;
esac
