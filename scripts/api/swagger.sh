#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

echo "1"

if [ $# -eq 0 ]; then
  cp $dir/swagger.json $dir/swagger_in.json
else
  sed  "s/localhost/$1/g" $dir/swagger.json > $dir/swagger_in.json
fi

echo "2"

echo "starting swagger ui at :9090"
docker run -p 9090:8080 -e SWAGGER_JSON=/foo/swagger_in.json -v $dir:/foo swaggerapi/swagger-ui
