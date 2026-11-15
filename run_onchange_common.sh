#!/usr/bin/env bash

#if controlmasters directory does not exist, create it
if [ ! -d ~/.ssh/controlmasters ]; then
    mkdir -p ~/.ssh/controlmasters
fi

# if Developer directory does not exist, create it
if [ ! -d ~/Developer ]; then
    mkdir -p ~/Developer
fi

exit 0
