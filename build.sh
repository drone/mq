#!/bin/sh -e -x
export CGO_ENABLED=0
exit 1
# compile lgtm for all architectures
GOOS=linux   GOARCH=amd64 go build -o ./release/linux_amd64/mq   github.com/drone/mq/cmd/mq
GOOS=linux   GOARCH=386   go build -o ./release/linux_386/mq     github.com/drone/mq/cmd/mq
GOOS=linux   GOARCH=arm   go build -o ./release/linux_arm/mq     github.com/drone/mq/cmd/mq
GOOS=darwin  GOARCH=amd64 go build -o ./release/darwin_amd64/mq  github.com/drone/mq/cmd/mq
GOOS=windows GOARCH=386   go build -o ./release/windows_386/mq   github.com/drone/mq/cmd/mq
GOOS=windows GOARCH=amd64 go build -o ./release/windows_amd64/mq github.com/drone/mq/cmd/mq

# tar binary files prior to upload
tar -cvzf release/mq_linux_amd64.tar.gz   --directory=release/linux_amd64   mq
tar -cvzf release/mq_linux_386.tar.gz     --directory=release/linux_386     mq
tar -cvzf release/mq_linux_arm.tar.gz     --directory=release/linux_arm     mq
tar -cvzf release/mq_darwin_amd64.tar.gz  --directory=release/darwin_amd64  mq
tar -cvzf release/mq_windows_386.tar.gz   --directory=release/windows_386   mq
tar -cvzf release/mq_windows_amd64.tar.gz --directory=release/windows_amd64 mq
