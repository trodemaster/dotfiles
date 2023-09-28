#!/usr/bin/env bash

{{ if eq .chezmoi.os "darwin" }}

# test for xcode command line tools by looking for /Library/Developer/CommandLineTools
if ! [[ -d /Library/Developer/CommandLineTools ]]; then
  echo "Installing xcode command line tools..."
  touch /tmp/.com.apple.dt.CommandLineTools.installondemand.in-progress
  CMDLINE_TOOLS=$(softwareupdate -l | grep "\*.*Command Line" | tail -n 1 | sed 's/^[^C]* //')
  softwareupdate -i "$CMDLINE_TOOLS" --verbose
fi

{{ end }}

exit 0
