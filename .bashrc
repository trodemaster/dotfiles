# Aliases
alias vautenv="set | grep VAULT"
alias java=""
alias lz="ls | fzf"
alias ls="ls -1 -p"
alias vaultcfg="set | grep VAULT | grep -v TOKEN"
alias vaultlogin="source ~/code/ops-tools/vaultlogin"
alias macports="~/code/macports/macports"
alias ykman="/Applications/YubiKey\ Manager.app/Contents/MacOS/ykman"
alias chrome="/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome"
alias lsusb="ioreg -p IOUSB"
alias ghc="GH_HOST=git.corp.adobe.com gh"
alias eaw="chmod +a \"user:$USER allow read,write,append,readattr,writeattr,readextattr,writeextattr,readsecurity\" " 

if [[ -f ~/code/weatherdesktop/wd ]]; then
  alias wd="~/code/weatherdesktop/wd"
fi

export HISTCONTROL=ignoredups:erasedups # no duplicate entries
export HISTSIZE=100000                  # big big history
export HISTFILESIZE=100000              # big big history
shopt -s histappend                     # append to history, don't overwrite it
PROMPT_COMMAND="history -a;$PROMPT_COMMAND"


# conditional shell stuff

if [[ $BASH != /opt/local/bin/bash ]]; then
  echo "Bash Version $BASH_VERSION"
fi

#shopt -q -s extglob

SHELL_TYPE="terminal"

if [[ ${GIT_ASKPASS} =~ "code-server" ]]; then
  SHELL_TYPE="code-server"
fi

if [[ ${GIT_ASKPASS} =~ "Visual" ]]; then
  SHELL_TYPE="vscode"
fi

# bash-completion
if [ -f /opt/local/etc/profile.d/bash_completion.sh ]; then
  . /opt/local/etc/profile.d/bash_completion.sh
fi

# prompt and friends
export PS1="\[\e[;34m\]\u\[\e[1;37m\]@\h\[\e[;32m\]:\W$ \[\e[0m\]"
export DISPLAY=:0
declare -x CLICOLOR=1
# colorize directorys

# macOS specific
if [[ $(uname) == "Darwin" ]]; then
  #echo "setting up macos"
  complete -C /opt/local/bin/vault vault

  # disable shell warning
  export BASH_SILENCE_DEPRECATION_WARNING=1
  #echo "path 0"
  #echo $PATH

  # Path
  export PATH="/opt/local/bin:/opt/local/sbin:/usr/local/bin:$PATH:/usr/local/go/bin:/$HOME/blakeconfig:/Applications/Visual Studio Code.app/Contents/Resources/app/bin:$HOME/Library/Python/3.9/bin"

  ## add gnu tools to path if available
  #if [[ -d /opt/local/libexec/gnubin ]]; then
  #  export PATH="/opt/local/libexec/gnubin:$PATH"
  #fi

  #echo "path 2"
  #echo $PATH

  # golang and hashicorp
  export GOPATH=${HOME}/code/go
  export PATH=$PATH:${HOME}/code/go/bin

  #echo "path 3"
  #echo $PATH

  # VMware tools
  if [[ -d Applications/VMware\ Fusion.app/Contents/Library ]]; then
    export PATH=$PATH:/Applications/VMware\ Fusion.app/Contents/Library/VMware\ OVF\ Tool
    alias vmrun="/Applications/VMware\ Fusion.app/Contents/Library/vmrun"
    alias vmware-vdiskmanager="/Applications/VMware\ Fusion.app/Contents/Library/vmware-vdiskmanager"
  fi

  # Packer
  #export PACKER_LOG=1
  #export PACKER_LOG_PATH=logs/packer-build.log
  #export PACKER_KEY_INTERVAL=10ms

  # Aliases
  alias fullres="~/code/blakeconfig/fullres"
  alias preso="~/code/blakeconfig/preso"
  alias remote="~/code/blakeconfig/mirrorscreen"
  alias encode="~/code/timelapse-tools/encode"
  alias crtchk="/opt/local/bin/openssl x509 -text -noout -dates -in"
  alias yolossh="ssh -o \"ControlPath ~/.ssh/controlmasters/%r@%h:%p\" -o \"ControlMaster auto\" -o \"ControlPersist 10m\" -o \"StrictHostKeyChecking no\" -o \"UserKnownHostsFile=/dev/null\""
  #alias killssh="/opt/local/libexec/gnubin/sed"
  #alias testbork="echo before $1 after"
  if [[ -f ${HOME}/code/macsl/macsl ]]; then
    alias macsl=${HOME}/code/macsl/macsl
  fi
  # python virtualenv
  #export VIRTUAL_ENV_DISABLE_PROMPT=1
  #export VIRTUALENVWRAPPER_PYTHON='/opt/local/bin/python3.7'
  #export VIRTUALENVWRAPPER_VIRTUALENV='/opt/local/bin/virtualenv-3.7'
  #export VIRTUALENVWRAPPER_VIRTUALENV_CLONE='/opt/local/bin/virtualenv-clone-3.7'
  #source /opt/local/bin/virtualenvwrapper.sh-3.7
  #source ~/.vpython/bin/activate

  # bash completion
  if [ -f /opt/local/etc/profile.d/bash_completion.sh ]; then
    . /opt/local/etc/profile.d/bash_completion.sh
  fi

fi

if [[ $(uname) == "Linux" ]]; then
  echo "linux"
  export SYSTEMD_EDITOR="/usr/bin/vi"
fi

# source the fzf config
[[ -f "${HOME}/.fzf_completion.bash" ]] && source "${HOME}/.fzf_completion.bash"
[[ -f "${HOME}/.fzf_completion.bash" ]] && source "${HOME}/.fzf_key-bindings.bash"

# set code as editor without wait
[[ -f "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code" ]] && export EDITOR="code"

# configure secretive ssh
if [ -f ${HOME}/Library/Containers/com.maxgoedjen.Secretive.SecretAgent ]; then
  export SSH_AUTH_SOCK=${HOME}/Library/Containers/com.maxgoedjen.Secretive.SecretAgent/Data/socket.ssh
fi

# source the azure cli config
[ -f /usr/local/lib/azure-cli/az.completion ] && source '/usr/local/lib/azure-cli/az.completion'

# add bash autocomplete for vault
[ -f /opt/local/bin/vault ] && complete -C /opt/local/bin/vault vault

# Do stuff for vscode shell
#if [[ $SHELL_TYPE =~ "code-server" ]];then
#  alias code="~/code/code-server/bin/code-server"
#  echo yep
#
#fi

# add kubeconfig
[[ -f ${HOME}/kubeconfig.yaml ]] && export KUBECONFIG="${HOME}/kubeconfig.yaml"

# Do stuff for vscode shell
if [[ $SHELL_TYPE =~ "vscode" ]]; then
  true
fi

# do stuff for terminal shell
#SHELL_TYPE=disablepowerline
if [[ $SHELL_TYPE =~ (terminal|vscode|code-server) ]]; then
  #alias code="/Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code"

  # seup powerline
  PYTHON_VERSION=$(python --version | /usr/bin/grep -oE "[1-9].*" | cut -d . -f -2)
  if POWERLINE_DAEMON=$(ls -1 /opt/local/Library/Frameworks/Python.framework/Versions/${PYTHON_VERSION}/bin/powerline-daemon); then
    export PATH="$PATH:/opt/local/Library/Frameworks/Python.framework/Versions/${PYTHON_VERSION}/bin"
    powerline-daemon -q
    export POWERLINE_BASH_CONTINUATION=1
    export POWERLINE_BASH_SELECT=1
    source /opt/local/Library/Frameworks/Python.framework/Versions/${PYTHON_VERSION}/lib/python${PYTHON_VERSION}/site-packages/powerline/bindings/bash/powerline.sh
  fi
fi
