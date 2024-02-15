#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
shopt -s nullglob nocaseglob

# check for gh cli tool
if ! command -v gh &>/dev/null; then
  echo "gh command line not installed skipping machine-cfg repo clone"
  exit 0
fi

# if running on linux set BROWSER env var to false
#if [[ $(uname) == "Linux" ]]; then
  export BROWSER=false
#fi

# clone the repo from github.com using the gh command line tool
if [[ $MACHINECFG == "personal" ]]; then
  echo "cloning personal machine-cfg repo..."
  if [[ -d ~/code/machine-cfg ]]; then
    echo "machine-cfg repo already exists"
    git -C ~/code/machine-cfg pull
  else
    gh auth login
    gh auth setup-git
    gh repo clone trodemaster/machine-cfg ~/code/machine-cfg
  fi
fi

exit 0

if [[ $MACHINE == "work" ]]; then
  echo "cloning work machine-cfg repo..."
  gh auth login https://github.com/adobe/blake-machine-cfg ~/code/machine-cfg
fi

exit 0
