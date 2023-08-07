#!/bin/bash

set -e

BLKADDR_FILE=$HOME/.config/zonetracer/f2fs_blkaddr.txt

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

if [ ! -f $BLKADDR_FILE ]
then
    echo "blkaddr file is not found. Please run ./scripts/mount_f2fs.sh"
    exit
fi

MAIN_BLKADDR=$(cat $BLKADDR_FILE | awk '{print $1}')
START_BLKADDR=$(cat $BLKADDR_FILE | awk '{print $2}')
ZNS_DEVICE=$(cat $BLKADDR_FILE | awk '{print $3}')
REGULAR_DEVICE=$(cat $BLKADDR_FILE | awk '{print $4}')

./src/bpf/f2fszonetracer $ZNS_DEVICE $REGULAR_DEVICE $MAIN_BLKADDR $START_BLKADDR | ./viewer/viewer