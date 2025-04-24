#!/bin/bash

# apply_migrations.sh
# Script to apply PostgreSQL migrations
# Path configurations
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../" &> /dev/null && pwd)"
ENV_FILE="${PROJECT_ROOT}/.env" # Path to the .env file
MIGRATIONS_DIR="${PROJECT_ROOT}/db/migrations"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables from .env file
if [ -f "$ENV_FILE" ]; then
    echo -e "${GREEN}Loading configuration from .env file...${NC}"
    source "$ENV_FILE"
else
    echo -e "${YELLOW}No .env file found, using default values${NC}"
fi

# Set default values if not provided in .env
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-"strategy_user"}
DB_PASS=${DB_PASS:-"strategy_password"}
DB_NAME=${DB_NAME:-"strategy_wars"}

# Check if migration files exist
UP_MIGRATION="$MIGRATIONS_DIR/up.sql"
DOWN_MIGRATION="$MIGRATIONS_DIR/down.sql"

if [ ! -f "$UP_MIGRATION" ] || [ ! -f "$DOWN_MIGRATION" ]; then
    echo -e "${RED}Error: Migration files not found in $MIGRATIONS_DIR${NC}"
    echo -e "${YELLOW}Make sure up.sql and down.sql exist in the migrations folder${NC}"
    exit 1
fi

# Function to execute SQL in Docker container
execute_sql_file() {
    local file=$1
    local action=$2
    
    echo -e "${YELLOW}Applying $action migration...${NC}"
    
    # Use docker exec to run psql inside the container
    cat "$file" | docker exec -i strategy_postgres psql -U "$DB_USER" -d "$DB_NAME"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}$action migration completed successfully.${NC}"
        return 0
    else
        echo -e "${RED}$action migration failed.${NC}"
        return 1
    fi
}

# Main execution
echo "=== PostgreSQL Migration Tool ==="
echo "Database: $DB_NAME"
echo "Using Docker container: strategy_postgres"
echo

# Simple menu
if [ "$1" == "down" ]; then
    execute_sql_file "$DOWN_MIGRATION" "down"
else
    execute_sql_file "$UP_MIGRATION" "up"
fi

echo
echo "Migration completed."