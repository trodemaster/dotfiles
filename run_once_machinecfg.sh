#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
shopt -s nullglob nocaseglob

echo "machineconfig from chezmoi data is $MACHINECFG"

# if running on linux set BROWSER env var to false
if [[ $(uname) == "Linux" ]]; then
  echo "setting BROWSER env var to false"
  export BROWSER=false
fi

# clone the repo from github.com using the gh command line tool
if [[ $MACHINECFG == "personal" ]]; then
  echo "cloning personal machine-cfg repo..."
  if [[ -d ~/code/machine-cfg ]]; then
    echo "machine-cfg repo already exists"
    git -C ~/code/machine-cfg pull
  else
    gh auth login
    git clone https://github.com/trodemaster/machine-cfg.git ~/code/machine-cfg
  fi
fi

exit 0

if [[ $MACHINE == "work" ]]; then
  echo "cloning work machine-cfg repo..."
  gh auth login https://github.com/adobe/blake-machine-cfg ~/code/machine-cfg
fi

exit 0