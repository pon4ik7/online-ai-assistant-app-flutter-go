#!/bin/sh

DB_URL="postgres://radat:radatSWP25@postgres:5432/radatDB?sslmode=disable"

wait_for_postgres() {
    until pg_isready -h postgres -p 5432 -U radat -d radatDB -q; do
        echo "Waiting for PostgreSQL to become available"
        sleep 2
    done
}
wait_for_postgres
echo "Applying database migrations"
if ! migrate -path /migrations -database "$DB_URL" up; then
    echo "Failed to apply migrations"
    exit 1
fi
echo "Starting application"
exec /app/main