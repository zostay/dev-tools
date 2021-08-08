#!/bin/sh

git clone git@github.com:zostay/dev-tools.git $HOME/.zx

cd ~/.zx

go build -o ~/.zx/bin/zxconfig ./cmd/zxconfig/...
./bin/zxinstall
