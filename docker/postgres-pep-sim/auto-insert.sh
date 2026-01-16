#!/bin/bash
# Auto-insert script for postgres-pep-sim
# Runs in background and inserts a new death record every 2 minutes
# This simulates real hospital activity for demo purposes

set -e

# Wait for PostgreSQL to be fully ready
until pg_isready -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
    echo "Waiting for PostgreSQL to be ready..."
    sleep 2
done

# Additional wait for init scripts to complete
sleep 5

echo "Starting auto-insert daemon (interval: 120 seconds)"

# Background loop for auto-insertion
while true; do
    # Sleep first to allow manual testing before first auto-insert
    sleep 120

    # Insert a random death record
    PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT insert_random_obito();" 2>/dev/null || {
        echo "Warning: Failed to insert auto-obito, will retry..."
    }

    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Auto-inserted new obito record"
done &

echo "Auto-insert daemon started in background (PID: $!)"
