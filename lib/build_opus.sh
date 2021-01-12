#!/bin/bash

if [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    echo "GNU/Linux"
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
    echo "MINGW64_NT"
    mingw64_params='--host=x86_64-w64-mingw32'
else
    echo "暂不支持从该环境编译 欢迎提交pr"
    exit
fi

build_path="$(pwd)/build"

cd opus

git clean -xfd
./autogen.sh

#make clean
./configure \
    --prefix="$build_path" \
    --enable-static \
    --disable-shared \
    --disable-extra-programs \
    --disable-doc \
    --enable-float-approx \
    $mingw64_params \
    CFLAGS="-pipe -Ofast -funroll-loops -ffast-math"
make -j$(nproc)
make install
