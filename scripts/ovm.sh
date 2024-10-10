#!/usr/bin/env bash


ovm_init(){
  echo ovm-arm64 $workdir machine init  $image $version
}

ovm_start(){
  echo ovm-arm64 $workdir machine start  $volume $external_disk $twinpid
}

ovm_restapi(){
  echo ovm-arm64 system service
}


check_environment() {
  # Test ovm-arm64 in PATH
  ovm-arm64 > /dev/null 2>&1
    ret=$?
    if [ $ret -ne 0 ]; then
      echo "ovm-arm64 not find !"
      exit $ret
    fi
}

main(){
  workdir=$1
  image=$2
  version=$3
  volume=$4
  external_disk=$5
  twinpid=$6
  ovm_init
  ovm_start
  ovm_restapi
}

#  ovm.sh \
#  --workdir=/Users/danhexon/myvm/ \
#  --images=/Users/danhexon/alpine_virt/bugbox-machine-default-arm64.raw.xz \
#  --image-version="1.0" \
#  --volume=/tmp/:/tmp/macos/tmp/ \
#  --external-disk=/Users/danhexon/alpine_virt/mydisk.raw  \
#  --twinpid=1000
main "$@"