#! /bin/bash

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi


if [ "$#" -ne 1 ]; then
    echo "Usage: ./set_cpu.sh [mode]"
	exit -1
fi

mode=$1

case $mode in

  powersave)
    ;;

  performance)
    ;;

  *)
    echo "Wrong mode"
    exit -1
esac

echo "Set all cpu to $1 mode"
for file in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do echo $1 > $file; done
