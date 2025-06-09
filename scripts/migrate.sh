#!/bin/bash

# Database Migration Script for Prayog Serviceability Service
# This script provides easy commands for database operations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  migrate     Run database migrations"
    echo "  seed        Seed database with initial data"
    echo "  drop        Drop all tables (WARNING: destructive)"
    echo "  reset       Drop tables, run migrations, and seed data"
    echo "  status      Check database connection status"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 migrate              # Run migrations only"
    echo "  $0 reset               # Complete database reset"
    echo "  $0 migrate seed         # Run migrations then seed"
}

# Function to check if migration binary exists
check_migration_binary() {
    if [ ! -f "./cmd/migrations/migrations" ]; then
        print_info "Building migration binary..."
        go build -o ./cmd/migrations/migrations ./cmd/migrations/
        if [ $? -ne 0 ]; then
            print_error "Failed to build migration binary"
            exit 1
        fi
    fi
}

# Function to run migration command
run_migration() {
    check_migration_binary
    print_info "Running: ./cmd/migrations/migrations $@"
    ./cmd/migrations/migrations "$@"
}

# Function to check database status
check_status() {
    print_info "Checking database connection..."
    if go run ./cmd/migrations/ 2>/dev/null; then
        print_info "Database connection successful"
    else
        print_error "Database connection failed"
        exit 1
    fi
}

# Main script logic
case "${1:-help}" in
    "migrate")
        print_info "Running database migrations..."
        run_migration -migrate
        print_info "Migrations completed successfully"
        ;;
    
    "seed")
        print_info "Seeding database..."
        run_migration -seed
        print_info "Database seeded successfully"
        ;;
    
    "drop")
        print_warning "This will DROP ALL TABLES in the database!"
        read -p "Are you sure? Type 'yes' to continue: " confirm
        if [ "$confirm" = "yes" ]; then
            print_info "Dropping all tables..."
            run_migration -drop
            print_info "All tables dropped successfully"
        else
            print_info "Operation cancelled"
        fi
        ;;
    
    "reset")
        print_warning "This will DROP ALL TABLES, run migrations, and seed data!"
        read -p "Are you sure? Type 'yes' to continue: " confirm
        if [ "$confirm" = "yes" ]; then
            print_info "Resetting database..."
            run_migration -drop -migrate -seed
            print_info "Database reset completed successfully"
        else
            print_info "Operation cancelled"
        fi
        ;;
    
    "status")
        check_status
        ;;
    
    "help"|"--help"|"-h")
        show_usage
        ;;
    
    *)
        # Handle multiple commands
        if [ $# -gt 0 ]; then
            flags=""
            for arg in "$@"; do
                case "$arg" in
                    "migrate") flags="$flags -migrate" ;;
                    "seed") flags="$flags -seed" ;;
                    "drop") flags="$flags -drop" ;;
                    *) 
                        print_error "Unknown command: $arg"
                        show_usage
                        exit 1
                        ;;
                esac
            done
            
            if [ -n "$flags" ]; then
                print_info "Running multiple operations..."
                run_migration $flags
                print_info "All operations completed successfully"
            fi
        else
            show_usage
        fi
        ;;
esac 