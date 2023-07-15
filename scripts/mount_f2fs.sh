#!/bin/bash

OVERPROVISIONING_SECTIONS=3
LOOP_DEV_NUM=77
MOUNT_POINT=/mnt/f2fs
METADATA_IMG=/tmp/f2fs_metadata.img

BLKADDR_FILE=$HOME/.config/zonetracer/f2fs_blkaddr.txt
MKFS_LOG_FILE=$HOME/.config/zonetracer/mkfs.f2fs

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
mkfs.f2fs -d 1 -o $OVERPROVISIONING_SECTIONS -m -f -c /dev/nvme0n1 /dev/loop$LOOP_DEV_NUM | tee $MKFS_LOG_FILE
MAIN_BLKADDR=$(cat $MKFS_LOG_FILE | awk '/main_blkaddr/{print $NF}')
START_BLKADDR=$(cat $MKFS_LOG_FILE | awk '/start_blkaddr/ {print $2}')
echo "$MAIN_BLKADDR $START_BLKADDR $LOOP_DEV_NUM" > $BLKADDR_FILE
mount -t f2fs /dev/loop$LOOP_DEV_NUM $MOUNT_POINT