#!/bin/bash

# use the following command to list all available distributions
# $ go tool dist list

DISTS=(
  linux/amd64
  windows/amd64
  linux/arm64
  linux/arm
)

OUTPUT_DIR=build

for dist in ${DISTS[@]}; do
  os=$(echo $dist | cut -d '/' -f 1)
  arch=$(echo ${dist} | cut -d '/' -f 2)

  output_file=${OUTPUT_DIR}/codefetcher_${os}_${arch}
  if [ "$os" = "windows" ]; then
    output_file+=".exe"
  fi

  echo -n "Building $(basename $output_file)... "
  env GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build -o ${output_file} cmd/codefetcher/main.go
  if [ $? -eq 0 ]; then
    echo "OK"
  fi

done
