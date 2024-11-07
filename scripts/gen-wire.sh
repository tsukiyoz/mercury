#!/bin/bash
ROOT=$(cd "$(dirname "$0")/.." || exit ; pwd)

services=("user" "sms" "interactive" "comment" "oauth2" "crontask" "article" "captcha" "ranking" "bff" "follow")

echo "=======> check services..."
args=$@
if [ $# -gt 0 ]; then
  tmpSrv=()
  for target in "${args[@]}"; do
    found=false
    for service in "${services[@]}"; do
        if [[ $service == "$target" ]]; then
          tmpSrv+=("$service")
          found=true
          break
        fi
    done
    if ! $found; then
      echo "=======> service $target does not exist, ignore..."
    fi
  done
  services=("${tmpSrv[@]}")
fi

if [ ${#services[@]} -eq 0 ]; then
  echo "=======> no valid services specified, exiting..."
  exit 0
fi

echo "=======> wire service" "[${services[*]}]"

for service in "${services[@]}"
do
  service_dir="${ROOT}/${service}"
  if [ -d "$service_dir" ]; then
    cd "$service_dir" || exit
    wire
  else
    echo "$service_dir does not exist, ignore..."
  fi
done

cd "${ROOT}" || exit
echo "=======> finished wire"