#!/bin/sh

git clone github.com/zostay/dev-tools $HOME/.zx

cd ~/.zx

go build -o ~/.zx/bin/zxconfig ./cmd/zxconfig/...
./bin/zxinstall
