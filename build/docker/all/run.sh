#!/bin/sh

trap 'pkill -P $$; exit 1;' TERM INT

echo "starting ui"

/bin/direktiv-ui &

echo "starting minio"
./bin/minio --config-dir=/tmp server /data &
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

/bin/direktiv -d -t wis &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start direktiv: $status"
  exit $status
fi


echo "ready to go"
while sleep 60; do
  ps aux |grep direktiv-ui |grep -q -v grep
  PROCESS_1_STATUS=$?
  ps aux |grep minio |grep -q -v grep
  PROCESS_2_STATUS=$?
  ps aux |grep direktiv |grep -q -v grep
  PROCESS_3_STATUS=$?
  if [ $PROCESS_1_STATUS -ne 0 -o $PROCESS_2_STATUS -ne 0 -o $PROCESS_3_STATUS -ne 0 ]; then
    echo "One of the processes has already exited."
    echo  $PROCESS_1_STATUS
    echo  $PROCESS_2_STATUS
    echo  $PROCESS_3_STATUS
    exit 1
  fi
done
