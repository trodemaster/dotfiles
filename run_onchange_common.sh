#!/usr/bin/env bash

#if controlmasters directory does not exist, create it
if [ ! -d ~/.ssh/controlmasters ]; then
    mkdir -p ~/.ssh/controlmasters
fi

# if code directory does not exist, create it
if [ ! -d ~/code ]; then
    mkdir -p ~/code
fi

# add ssh include if the directory exists
if [ -d ~/code/machine-cfg ]; then
    if [ -d /etc/ssh/sshd_config.d ]; then
        if [ ! -f /etc/ssh/sshd_config.d/206-machine-cfg.conf ]; then
            echo "Include ~/code/machine-cfg/ssh_config" | sudo tee /etc/ssh/sshd_config.d/206-machine-cfg.conf >/dev/null
        fi
    fi
fi

exit 0
