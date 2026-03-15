#!/bin/sh
set -e

cd /app

if [ "$MIGRATE" = "true" ]; then
  echo "Running migrations..."
  ./migrate
fi

echo "Running api server..."
exec ./api
