#!/bin/bash

LOOP_DEV_NUM=77
MOUNT_POINT=/mnt/f2fs
METADATA_IMG=/tmp/f2fs_metadata.img

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

mkdir -p $MOUNT_POINT
umount -q $MOUNT_POINT
mkfs.f2fs -d 1 -o 3 -m -f -c /dev/nvme0n1 /dev/loop77
mount -t f2fs /dev/loop$LOOP_DEV_NUM $MOUNT_POINT