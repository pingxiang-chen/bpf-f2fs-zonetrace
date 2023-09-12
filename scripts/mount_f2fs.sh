#!/bin/bash

ZNS_DEVICE=nvme0n1
LOOP_DEV_NUM=77
REGULAR_DEVICE=loop$LOOP_DEV_NUM

OVERPROVISIONING_SECTIONS=3
MOUNT_POINT=/mnt/f2fs
METADATA_IMG=/tmp/f2fs_metadata.img

BLKADDR_FILE=$HOME/.config/zonetracer/f2fs_blkaddr.txt
MKFS_LOG_FILE=$HOME/.config/zonetracer/mkfs.f2fs

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

# 1. Create metadata image for loop device
if [ ! -f $METADATA_IMG ]
then
    echo "Create $METADATA_IMG"
    truncate -s 256M $METADATA_IMG
fi

# 2. Create loop device
if [ ! -b /dev/$REGULAR_DEVICE ]
then
    echo "Create Loop device"
    losetup /dev/$REGULAR_DEVICE $METADATA_IMG
fi

# 3. Create blkaddr file for recording main_blkaddr and start_blkaddr when mkfs.f2fs
if [ ! -d $(dirname $BLKADDR_FILE) ]
then
    mkdir -p $(dirname $BLKADDR_FILE)
fi

# 4. Create mount point
mkdir -p $MOUNT_POINT
umount -q $MOUNT_POINT

# 5. Format f2fs
mkfs.f2fs -d 1 -o $OVERPROVISIONING_SECTIONS -m -f -c /dev/$ZNS_DEVICE /dev/$REGULAR_DEVICE | tee $MKFS_LOG_FILE

if [ $? -ne 0 ]
then
    echo "Format f2fs failed"
    exit 1
fi

# 6. Record main_blkaddr and start_blkaddr
MAIN_BLKADDR=$(cat $MKFS_LOG_FILE | awk '/main_blkaddr/{print $NF}')
START_BLKADDR=$(cat $MKFS_LOG_FILE | awk '/start_blkaddr/ {print $2}')
echo "$MAIN_BLKADDR $START_BLKADDR $ZNS_DEVICE $REGULAR_DEVICE" > $BLKADDR_FILE

# 7. Mount f2fs
mount -t f2fs /dev/$REGULAR_DEVICE $MOUNT_POINT