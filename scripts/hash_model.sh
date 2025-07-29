#!/usr/bin/env bash
set -e
FILE=$1  # path/to/model.onnx
sha256sum ${FILE} | awk '{print $1}' > ${FILE}.sha256
echo "SHA256:" $(cat ${FILE}.sha256)
