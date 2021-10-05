#!/bin/bash

echo "database:
  host: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "pgbouncer-host"}}' | base64 --decode)\"
  port: $(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "pgbouncer-port"}}' | base64 --decode)
  user: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "user"}}' | base64 --decode)\"
  password: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "password"}}' | base64 --decode)\"
  name: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "dbname"}}' | base64 --decode)\"
  sslmode: require"
