#!/bin/bash
set -e

ls /app/public

cat etc/nginx/conf.d/default.conf

nginx -g 'daemon off;' &
dblect &
ORIGIN_HOST="https://dblect.fly.dev" node /app/web-be/app.js &

wait -n
exit $?
