#!/bin/sh
# wait-for-postgres.sh
# Usage: wait-for-postgres.sh host:port [timeout]

set -e

HOST_PORT="$1"
TIMEOUT="${2:-30}"

if [ -z "$HOST_PORT" ]; then
  echo "Usage: $0 host:port [timeout]"
  exit 1
fi

HOST=$(echo $HOST_PORT | cut -d: -f1)
PORT=$(echo $HOST_PORT | cut -d: -f2)

for i in $(seq 1 $TIMEOUT); do
  if nc -z "$HOST" "$PORT"; then
    echo "Postgres is up on $HOST:$PORT"
    exit 0
  fi
  echo "Waiting for Postgres at $HOST:$PORT... ($i/$TIMEOUT)"
  sleep 1
done

echo "Timed out waiting for Postgres at $HOST:$PORT"
exit 1
