#!/usr/bin/env bash

usage="./start-doc-cache [testnet|testnet2|mainnet] [build|image]"
if [ $# -ne 2 ]; then
    echo $usage
    exit 1
fi

if [[ $1 != 'testnet' && $1 != 'testnet2' && $1 != 'mainnet' ]]; then
    echo $usage
    exit 1
elif [[ $2 != 'build' && $2 != 'image' ]]; then
    echo $usage
    exit 1
fi

script_path="$(dirname $(realpath ${BASH_SOURCE[0]}))"

pushd $script_path

cp .env.$1 .env

if [[ $2 = build ]]; then
  docker-compose -p doc-cache-go-$1 up --build -d
else
  docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p doc-cache-go-$1 up -d
fi

popd
