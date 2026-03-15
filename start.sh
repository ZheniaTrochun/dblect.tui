#!/bin/bash
set -e

cleanup() {
  echo "Shutting down..."
  kill -TERM "$nginx_pid" "$dblect_pid" "$node_pid" 2>/dev/null || true
  wait
  exit 0
}

trap cleanup SIGTERM SIGINT

nginx -g 'daemon off;' &
nginx_pid=$!

dblect &
dblect_pid=$!

ORIGIN_HOST="https://dblect.fly.dev" node /app/web-be/app.js &
node_pid=$!

wait -n
cleanup
