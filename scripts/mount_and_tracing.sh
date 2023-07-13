#!/bin/sh

set -e

BLKADDR_FILE=/tmp/f2fs_start_blkaddr.txt

ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

if [ ! -f $BLKADDR_FILE ]
then
    echo "start_blkaddr is not found. formatting f2fs..."
    # run mount_f2fs.sh and get zoned start_blkaddr
    START_BLKADDR=$(sudo ./scripts/mount_f2fs.sh | awk '/start_blkaddr/ {print $2}')
    echo $START_BLKADDR > $BLKADDR_FILE
fi

START_BLKADDR=$(cat $START_BLKADDR)
sudo ./src/bpf/f2fszonetracer nvme0n1 $START_BLKADDR | ./viewer/viewer