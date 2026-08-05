#!/usr/bin/env bash

SSHD_CONFIG_FILE="/etc/ssh/sshd_config.d/99-disable-password-auth.conf"

if [ ! -f "$SSHD_CONFIG_FILE" ] || ! grep -q "^PasswordAuthentication no" "$SSHD_CONFIG_FILE"; then
    printf 'PasswordAuthentication no\n' | sudo tee "$SSHD_CONFIG_FILE" > /dev/null
    sudo chmod 644 "$SSHD_CONFIG_FILE"
    echo "Created $SSHD_CONFIG_FILE"

    if command -v systemctl &> /dev/null; then
        sudo systemctl reload sshd
    elif command -v launchctl &> /dev/null; then
        sudo launchctl kickstart -k system/com.openssh.sshd
    fi
fi

exit 0
