#!/bin/bash

ZNS_DEVICE=nvme3n1
REGULAR_DEVICE=nvme4n1p1

OVERPROVISIONING_SECTIONS=3
MOUNT_POINT=/mnt/f2fs

BLKADDR_FILE=$HOME/.config/zonetracer/f2fs_blkaddr.txt
MKFS_LOG_FILE=$HOME/.config/zonetracer/mkfs.f2fs

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

if [ ! -d $(dirname $BLKADDR_FILE) ]
then
    mkdir -p $(dirname $BLKADDR_FILE)
fi

mkdir -p $MOUNT_POINT
umount -q $MOUNT_POINT

mkfs.f2fs -d 1 -o $OVERPROVISIONING_SECTIONS -m -f -c /dev/$ZNS_DEVICE /dev/$REGULAR_DEVICE | tee $MKFS_LOG_FILE

MAIN_BLKADDR=$(cat $MKFS_LOG_FILE | awk '/main_blkaddr/{print $NF}')
START_BLKADDR=$(cat $MKFS_LOG_FILE | awk '/start_blkaddr/ {print $2}')
echo "$MAIN_BLKADDR $START_BLKADDR $ZNS_DEVICE $REGULAR_DEVICE" > $BLKADDR_FILE

mount -t f2fs /dev/$REGULAR_DEVICE $MOUNT_POINT
