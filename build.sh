#!/usr/bin/env bash

SRC_DIR=.
DST_DIR=tob

mkdir -p "${DST_DIR}"

protoc -I "${SRC_DIR}/" "${SRC_DIR}/tob.proto" --csharp_out=${DST_DIR} --go_out=plugins=grpc:${DST_DIR}
