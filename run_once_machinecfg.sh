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


exit 0
