#!/bin/bash

cd /ent && go get entgo.io/ent@v0.11.4

go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/upsert,sql/execquery,sql/lock,sql/modifier,sql/execquery $1/schema