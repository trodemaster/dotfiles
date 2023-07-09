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
if [ "$(git config --global push.default)" != "current" ]; then
    git config --global push.default current
fi
if [ "$(git config --global pull.rebase)" != "false" ]; then
    git config --global pull.rebase false
fi
if [ -z "$(git config --global user.name)" ]; then
    git config --global user.name "Blake Garner"
fi
if [ -z "$(git config --global user.email)" ]; then
    git config --global user.email blake@netjibbing.com
fi
if [ -z "$(git config --global alias.change-commits)" ]; then
    git config --global alias.change-commits '!'"f() { VAR=\$1; OLD=\$2; NEW=\$3; shift 3; git filter-branch --env-filter \"if [[ \\\"\$\`echo \$VAR\`\\\" = '\$OLD' ]]; then export \$VAR='\$NEW'; fi\" \$@; }; f"
fi
if [ "$(git config --global commit.gpgsign)" != "false" ]; then
    git config --global commit.gpgsign false
fi

exit 0
