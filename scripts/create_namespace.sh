#!/bin/sh

ZNS_SSD=/dev/nvme3
NUM_OF_NAMESPACE=$1
ZONE_SIZE=524288
NUM_OF_ZONE_EACH_NAMESPACE=0
GB=1073741824
CONVENTION_SIZE=$((2*GB/4096))
NUM_OF_ZONE=0
NUM_OF_ZONE_OF_EACH_NAMESPACE=$1

if [ "$#" -ne 1 ]; then
	echo "Usage: ./create_namespace.sh [# of zones]"
	exit -1
fi
nvme delete-ns $ZNS_SSD -n 0xffffffff

if [ $? -ne 0 ]
then
   exit 1
fi

nvme format $ZNS_SSD -n 0xffffffff -l 2 -f

if [ $? -ne 0 ]
then
   exit 1
fi

SIZE=$(nvme id-ctrl $ZNS_SSD | grep tnvmcap | sed 's/,//g'| sed 's/^tnvmcap   : //g' | awk '{print $1/4096}')

NUM_OF_ZONE=$((SIZE/ZONE_SIZE))
echo "Total zone in ZNS NVMe: $NUM_OF_ZONE"
NUM_OF_ZONE_OF_EACH_NAMESPACE=$1
echo "Number of zones of each namespace: $NUM_OF_ZONE_OF_EACH_NAMESPACE"

SIZE_OF_EACH_NAMESPACE=$((NUM_OF_ZONE_OF_EACH_NAMESPACE*ZONE_SIZE))

nvme create-ns -s $SIZE_OF_EACH_NAMESPACE -c $SIZE_OF_EACH_NAMESPACE -f 2 -d 0 --csi=2 $ZNS_SSD

if [ $? -ne 0 ]
then
	exit 1
fi

nvme attach-ns $ZNS_SSD -n 1 -c 0
