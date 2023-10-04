#!/bin/bash
FILE=$1

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

if [ "$#" -ne 1 ]; then
    echo "Usage: ./mount_zns.sh [filename]"
	exit -1
fi

if [ ! -f $FILE ]; then
    echo "file does not exist"
    exit -1
fi

ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

./fiemap/fiemap $FILE
[ -f $FILE ] && watch -n 1 ../fiemap/fiemap $FILE