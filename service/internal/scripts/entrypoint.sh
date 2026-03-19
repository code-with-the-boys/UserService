#!/bin/sh
set -e

echo "=== Starting UserService ==="
echo "Applying database migrations..."

GOOSE_BIN="/usr/local/bin/goose"
if [ ! -x "$GOOSE_BIN" ]; then
  echo "ERROR: goose binary not found at $GOOSE_BIN"
  exit 1
fi

if [ -z "$POSTGRES_HOST" ] || [ -z "$POSTGRES_DB" ]; then
    echo "ERROR: Required PostgreSQL environment variables are not set"
    echo "Required: POSTGRES_HOST, POSTGRES_DB"
    exit 1
fi


POSTGRES_ADMIN_USER=${POSTGRES_ADMIN_USER:-postgres}
POSTGRES_ADMIN_PASSWORD=${POSTGRES_ADMIN_PASSWORD:-postgres}

DB_STRING="host=$POSTGRES_HOST port=5432 user=$POSTGRES_ADMIN_USER password=$POSTGRES_ADMIN_PASSWORD dbname=$POSTGRES_DB sslmode=disable"

echo "Connecting to PostgreSQL at $POSTGRES_HOST:5432 with user $POSTGRES_ADMIN_USER"


echo "Waiting for PostgreSQL to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0
until PGPASSWORD=$POSTGRES_ADMIN_PASSWORD psql -h "$POSTGRES_HOST" -p 5432 -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -c '\q' 2>/dev/null; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "ERROR: PostgreSQL not ready after $MAX_RETRIES attempts"
        exit 1
    fi
    echo "PostgreSQL is unavailable - sleeping (attempt $RETRY_COUNT/$MAX_RETRIES)"
    sleep 2
done

echo "PostgreSQL is up - executing migrations"

echo "Running migrations..."
"$GOOSE_BIN" -dir /app/migrations postgres "$DB_STRING" up

if [ $? -eq 0 ]; then
    echo "Migrations applied successfully"
    
    echo "Verifying user and role creation..."
    USER_CHECK=$(PGPASSWORD=$POSTGRES_ADMIN_PASSWORD psql -h "$POSTGRES_HOST" -p 5432 -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -t -c "SELECT 1 FROM pg_roles WHERE rolname='fitness_user_service_role';")
    if [ -n "$USER_CHECK" ]; then
        echo "✓ Role 'fitness_user_service_role' created successfully"
    else
        echo "⚠ Warning: Role 'fitness_user_service_role' not found"
    fi
    
    USER_CHECK=$(PGPASSWORD=$POSTGRES_ADMIN_PASSWORD psql -h "$POSTGRES_HOST" -p 5432 -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -t -c "SELECT 1 FROM pg_roles WHERE rolname='fitness_user';")
    if [ -n "$USER_CHECK" ]; then
        echo "✓ User 'fitness_user' created successfully"
    else
        echo "⚠ Warning: User 'fitness_user' not found"
    fi
else
    echo "Failed to apply migrations"
    exit 1
fi

echo "Starting application..."
exec /app/main