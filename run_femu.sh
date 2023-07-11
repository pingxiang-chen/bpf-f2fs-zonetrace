#!/bin/sh

set -e

ROOT=$(pwd)

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
    # update OSIMGF
    sed -i 's/u20s.qcow2/debian.qcow2/g' run-zns.sh
    sudo ./pkgdep.sh
    ./femu-compile.sh
fi

echo $ROOT
cd $ROOT/femu/build-femu
./run-zns.sh
