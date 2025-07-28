#!/bin/sh

set -eu

# POSTGRES_DSN="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DATABASE?sslmode=$POSTGRES_SSLMODE"
# echo "running migration..."
# migrate -source="file://migrations" -database="$POSTGRES_DSN" up

echo "running app..."
exec ./app

# echo "halting..."
# exec tail -f /dev/null
