#!/bin/bash

# VitalConnect Database Migration Script
# Usage: ./migrate.sh [up|down|status]

set -e

# Default database URL
DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/vitalconnect?sslmode=disable}"

# Parse database URL to extract components
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\(.*\):.*/\1/p' | cut -d':' -f1)
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*\/\/\(.*\):.*@.*/\1/p')
DB_PASS=$(echo $DATABASE_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')

# Set PGPASSWORD for non-interactive use
export PGPASSWORD="$DB_PASS"

MIGRATIONS_DIR="$(dirname "$0")"
ACTION="${1:-up}"

echo "VitalConnect Database Migration"
echo "================================"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "Action: $ACTION"
echo ""

run_migration() {
    local file=$1
    echo "Running: $(basename $file)"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file" 2>&1 || {
        # Ignore "already exists" errors
        if [[ $? -ne 0 ]]; then
            echo "  Warning: Some statements may have failed (object may already exist)"
        fi
    }
    echo ""
}

case $ACTION in
    up)
        echo "Running UP migrations..."
        echo ""

        # Run init.sql first
        if [ -f "$MIGRATIONS_DIR/init.sql" ]; then
            run_migration "$MIGRATIONS_DIR/init.sql"
        fi

        # Run numbered migrations in order
        for migration in $(ls -1 "$MIGRATIONS_DIR"/*.sql 2>/dev/null | grep -E '^.*/[0-9]+_.*\.sql$' | sort); do
            run_migration "$migration"
        done

        echo "Migrations completed!"
        ;;

    status)
        echo "Checking database tables..."
        echo ""
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT table_name, pg_size_pretty(pg_total_relation_size(quote_ident(table_name))) as size
            FROM information_schema.tables
            WHERE table_schema = 'public'
            AND table_type = 'BASE TABLE'
            ORDER BY table_name;
        "
        echo ""
        echo "Checking indexes..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT tablename, indexname
            FROM pg_indexes
            WHERE schemaname = 'public'
            ORDER BY tablename, indexname;
        "
        ;;

    verify)
        echo "Verifying required tables..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'hospitals') THEN '[OK]'
                    ELSE '[MISSING]'
                END as hospitals,
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN '[OK]'
                    ELSE '[MISSING]'
                END as users,
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'obitos_simulados') THEN '[OK]'
                    ELSE '[MISSING]'
                END as obitos_simulados,
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'occurrences') THEN '[OK]'
                    ELSE '[MISSING]'
                END as occurrences,
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'occurrence_history') THEN '[OK]'
                    ELSE '[MISSING]'
                END as occurrence_history,
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'triagem_rules') THEN '[OK]'
                    ELSE '[MISSING]'
                END as triagem_rules,
                CASE
                    WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'notifications') THEN '[OK]'
                    ELSE '[MISSING]'
                END as notifications;
        "
        ;;

    *)
        echo "Usage: $0 [up|status|verify]"
        echo ""
        echo "Commands:"
        echo "  up      - Run all migrations"
        echo "  status  - Show current tables and indexes"
        echo "  verify  - Check if all required tables exist"
        exit 1
        ;;
esac
