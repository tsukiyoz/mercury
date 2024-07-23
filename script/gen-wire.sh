#!/bin/bash
ROOT=$(cd `dirname $0`/..; pwd)

services=" follow user sms interactive comment oauth2 crontask article captcha ranking bff "

args=$@
if [ $# -gt 0 ]; then
  tmpSrv=" "
  for target in ${args}
  do
    if [[ ${services} == *" $target "* ]]; then
      tmpSrv+="$target "
    fi
  done
  services=$tmpSrv
fi

echo "=======> wire service [${services}]"

for service in ${services}
do
  cd ${ROOT}/${service}
  wire
done

cd ${ROOT}
echo "=======> finished wire"