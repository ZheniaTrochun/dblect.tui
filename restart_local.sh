#!/usr/bin/env bash

go build

echo "build done"

pid=$(pgrep dblect)
kill -9 $pid

echo "previous process stopped"

./dblect > /dev/null 2>&1 &

echo "new ssh server started"
