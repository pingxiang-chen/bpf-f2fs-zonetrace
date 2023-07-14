#!/bin/bash

OVERPROVISIONING_SECTIONS=3
LOOP_DEV_NUM=77
MOUNT_POINT=/mnt/f2fs
METADATA_IMG=/tmp/f2fs_metadata.img

BLKADDR_FILE=$HOME/.config/zonetracer/f2fs_blkaddr.txt

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

# setup metadata loop image
if [ ! -f $METADATA_IMG ]
then
    echo "Create $METADATA_IMG"
    truncate -s 256M $METADATA_IMG
fi

# setup loop device
if [ ! -b /dev/loop$LOOP_DEV_NUM ]
then
    echo "Create Loop device"
    losetup /dev/loop$LOOP_DEV_NUM $METADATA_IMG
fi

if [ ! -d $(dirname $BLKADDR_FILE) ]
then
    mkdir -p $(dirname $BLKADDR_FILE)
fi

mkdir -p $MOUNT_POINT
umount -q $MOUNT_POINT
MKFS_OUTPUT=$(mkfs.f2fs -d 1 -o $OVERPROVISIONING_SECTIONS -m -f -c /dev/nvme0n1 /dev/loop$LOOP_DEV_NUM)
MAIN_BLKADDR=$(echo $MKFS_OUTPUT | awk '/main_blkaddr/ {print $2}')
START_BLKADDR=$(echo $MKFS_OUTPUT | awk '/start_blkaddr/ {print $2}')
echo "$MAIN_BLKADDR $START_BLKADDR" > $BLKADDR_FILE
mount -t f2fs /dev/loop$LOOP_DEV_NUM $MOUNT_POINT