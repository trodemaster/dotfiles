#!/usr/bin/env bash

#if controlmasters directory does not exist, create it
if [ ! -d ~/.ssh/controlmasters ]; then
    mkdir -p ~/.ssh/controlmasters
fi

# if code directory does not exist, create it
if [ ! -d ~/code ]; then
    mkdir -p ~/code
fi

# config git
git config --global push.default current
git config --global pull.rebase false
git config --global user.name "Blake Garner"
git config --global user.email blake@netjibbing.com
git config --global alias.change-commits '!'"f() { VAR=\$1; OLD=\$2; NEW=\$3; shift 3; git filter-branch --env-filter \"if [[ \\\"\$\`echo \$VAR\`\\\" = '\$OLD' ]]; then export \$VAR='\$NEW'; fi\" \$@; }; f"
git config --unset commit.gpgsign

exit 0