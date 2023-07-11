#!/bin/sh
set -e

# FEMU Image directory
IMAGE_DIR=$HOME/image
# Virtual machine disk image
IMAGE_FILE=debian.qcow2


ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

if [ ! -d ./femu/build-femu ]
then
    echo "Installing FEMU..."
    # Install FEMU
    # Download and Extract
    wget -q --show-progress "https://github.com/vtess/FEMU/archive/refs/tags/femu-v8.0.0.tar.gz" -O "femu.tar.gz"
    mkdir -p femu
    tar zxf femu.tar.gz -C femu --strip-components=1

    # Build FEMU
    cd femu
    mkdir build-femu
    cd build-femu
    cp ../femu-scripts/femu-copy-scripts.sh .
    ./femu-copy-scripts.sh .
    sudo ./pkgdep.sh
    ./femu-compile.sh
else
    echo "FEMU is already installed."
fi

# update Image Path
sed -i "s|IMGDIR=.*|IMGDIR=$IMAGE_DIR|g" femu/build-femu/run-zns.sh
sed -i "s|OSIMGF=.*|OSIMGF=\$IMGDIR/$IMAGE_FILE|g" femu/build-femu/run-zns.sh

cd $ROOT/femu/build-femu
echo "Starting FEMU..."
./run-zns.sh
