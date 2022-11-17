#!/bin/bash

cd /ent && go get entgo.io/ent@v0.11.4
# go generate --feature upsert,execquery,namedges,lock,modifier,execquery $1

go run -mod=mod entgo.io/ent/cmd/ent generate --feature upsert,execquery,namedges,lock,modifier,execquery $1/schema