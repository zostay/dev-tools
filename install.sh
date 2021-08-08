#!/bin/zsh

if [[ -d "$HOME/.zx" ]]; then
    git --git-dir="$HOME/.zx/.git" pull
else
    git clone git@github.com:zostay/dev-tools.git $HOME/.zx
fi

cd ~/.zx

go build -o ~/.zx/bin ./cmd/zxconfig/...
./bin/zxinstall
