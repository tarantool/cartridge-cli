#!/bin/bash

if [[ "$1" ]]; then
    run_dir=$1
else
    run_dir=./dev
fi

mkdir -p ${run_dir}
cartridge start --cfg demo.yml --run_dir ${run_dir}
