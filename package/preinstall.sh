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

# Install creates the dayzsa user and group using the
# name 'dayzsa'. The dayzsa user does not have a shell.
# This function can be called more than once as it is idempotent.
install() {
    username="dayzsa"

    if getent group "$username" &>/dev/null; then
        echo "Group ${username} already exists."
    else
        groupadd "$username"
    fi

    if id "$username" &>/dev/null; then
        echo "User ${username} already exists"
        exit 0
    else
        useradd --shell /sbin/nologin --system "$username" -g "$username"
    fi
}

# Upgrade should perform the same steps as install
upgrade() {
    install
}

action="$1"

echo "Running preinstall with action: $action"

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
