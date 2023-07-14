#!/bin/sh

set -e

BLKADDR_FILE=$HOME/.config/zonetracer/f2fs_blkaddr.txt

if [ ! -d $(dirname $BLKADDR_FILE) ]
then
    mkdir -p $(dirname $BLKADDR_FILE)
fi

ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

if [ ! -f $BLKADDR_FILE ]
then
    echo "blkaddr is not found. formatting f2fs..."
    MOUNT_F2FS_OUTPUT=$(sudo ./scripts/mount_f2fs.sh)
    # run mount_f2fs.sh and get zoned start_blkaddr
    START_BLKADDR=$(echo $MOUNT_F2FS_OUTPUT | awk '/start_blkaddr/ {print $2}')
    MAIN_BLKADDR=$(echo $MOUNT_F2FS_OUTPUT | awk '/main_blkaddr/ {print $2}')
    echo "$MAIN_BLKADDR $START_BLKADDR" > $BLKADDR_FILE
fi

MAIN_BLKADDR=$(cat $BLKADDR_FILE | awk '{print $1}')
START_BLKADDR=$(cat $BLKADDR_FILE | awk '{print $2}')

sudo ./src/bpf/f2fszonetracer nvme0n1 $MAIN_BLKADDR $START_BLKADDR | ./viewer/viewer