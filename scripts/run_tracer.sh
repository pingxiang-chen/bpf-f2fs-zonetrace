#!/bin/sh

set -e

ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

if [ ! -f $BLKADDR_FILE ]
then
    echo "blkaddr file is not found. Please run ./scripts/mount_f2fs.sh"
    exit 1
fi

MAIN_BLKADDR=$(cat $BLKADDR_FILE | awk '{print $1}')
START_BLKADDR=$(cat $BLKADDR_FILE | awk '{print $2}')

sudo ./src/bpf/f2fszonetracer nvme0n1 $MAIN_BLKADDR $START_BLKADDR | ./viewer/viewer