#!/usr/bin/env bash

#if controlmasters directory does not exist, create it
if [ ! -d ~/.ssh/controlmasters ]; then
    mkdir -p ~/.ssh/controlmasters
fi

# if code directory does not exist, create it
if [ ! -d ~/code ]; then
    mkdir -p ~/code
fi

exit 0