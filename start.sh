#!/bin/bash
set -e

nginx -g 'daemon off;' &
dblect &
ORIGIN_HOST="https://dblect.fly.dev" node /app/web-be/app.js &

wait -n
exit $?
