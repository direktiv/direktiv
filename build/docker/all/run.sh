#!/bin/sh

echo "starting minio"

./bin/minio --config-dir=/tmp --certs-dir=/miniocerts server /data &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start minio: $status"
  exit $status
fi

./usr/local/bin/docker-entrypoint.sh postgres &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start postgres: $status"
  exit $status
fi


echo "waiting for minio"

while ! echo exit | nc localhost 9000; do sleep 3; done

echo "starting direktiv"

/bin/direktiv -d -t wis -c /etc/conf.toml &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start direktiv: $status"
  exit $status
fi

while sleep 60; do
  ps aux |grep direktiv |grep -q -v grep
  PROCESS_1_STATUS=$?
  ps aux |grep minio |grep -q -v grep
  PROCESS_2_STATUS=$?
  if [ $PROCESS_1_STATUS -ne 0 -o $PROCESS_2_STATUS -ne 0 ]; then
    echo "One of the processes has already exited."
    echo  $PROCESS_1_STATUS
    echo  $PROCESS_2_STATUS
    exit 1
  fi
done
