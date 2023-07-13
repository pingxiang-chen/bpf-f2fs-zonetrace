#!/bin/sh

set -e

ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

# run mount_f2fs.sh and get zoned start_blkaddr
START_BLKADDR = $(sudo ./scripts/mount_f2fs.sh | awk '/start_blkaddr/ {print $2}')

sudo ./src/bpf/f2fszonetracer nvme0n1 $START_BLKADDR | ./viewer/viewer