#!/bin/bash

cd /ent && go get entgo.io/ent@v0.11.4
go generate $1
