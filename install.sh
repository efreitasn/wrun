#!/bin/bash

set -e
cp wrun /usr/local/bin
cp completion.sh /usr/share/bash-completion/completions/wrun
if [ -f ~/.zshrc ]; then
  echo -e "\nautoload bashcompinit\nbashcompinit\nsource /usr/share/bash-completion/completions/wrun" >> ~/.zshrc
fi
echo "Installation is complete"


