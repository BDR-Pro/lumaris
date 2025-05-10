#!/usr/bin/env bash
# Minimal wait-for-it using /dev/tcp

host="$1"
port="$2"
shift 2
cmd="$@"

echo "Waiting for $host:$port..."

while ! echo > /dev/tcp/$host/$port 2>/dev/null; do
  sleep 1
done

echo "$host:$port is up!"
exec $cmd
