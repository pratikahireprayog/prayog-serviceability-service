#!/bin/bash

# High Performance Startup Script for Prayog Serviceability Service
# Optimized for handling 80K+ concurrent requests

set -e

echo "🚀 Starting Prayog Serviceability Service in High Performance Mode..."

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load high-performance environment
if [ -f "$SCRIPT_DIR/high-performance.env" ]; then
    echo "📋 Loading high-performance configuration..."
    export $(cat "$SCRIPT_DIR/high-performance.env" | grep -v '^#' | xargs)
else
    echo "⚠️  High-performance config not found, using defaults..."
fi

# Set additional Go runtime optimizations
export GOGC=${GOGC:-400}                    # Reduce GC frequency
export GOMEMLIMIT=${GOMEMLIMIT:-8GiB}       # Memory limit
export GOMAXPROCS=${GOMAXPROCS:-0}          # Use all CPUs

# Disable CGO for better performance (if possible)
export CGO_ENABLED=${CGO_ENABLED:-0}

# Build optimizations
export GOOS=${GOOS:-linux}
export GOARCH=${GOARCH:-amd64}

# Set ulimits for current process
echo "⚙️ Setting resource limits..."
ulimit -n 1048576  # File descriptors
ulimit -u 65536    # Max user processes
ulimit -s 8192     # Stack size (KB)

# Check if binary exists or needs building
BINARY_PATH="$PROJECT_ROOT/bin/server"
if [ ! -f "$BINARY_PATH" ] || [ "$PROJECT_ROOT/cmd/server/main.go" -nt "$BINARY_PATH" ]; then
    echo "🔨 Building optimized binary..."
    mkdir -p "$PROJECT_ROOT/bin"
    
    cd "$PROJECT_ROOT"
    
    # Build with optimizations
    go build \
        -ldflags="-s -w" \
        -gcflags="-B -C" \
        -tags netgo \
        -installsuffix netgo \
        -o "$BINARY_PATH" \
        ./cmd/server/main.go
        
    echo "✅ Binary built successfully"
else
    echo "✅ Using existing optimized binary"
fi

# Pre-flight checks
echo "🔍 Running pre-flight checks..."

# Check available memory
AVAILABLE_MEM=$(free -m | awk '/^Mem:/{print $7}')
if [ "$AVAILABLE_MEM" -lt 4096 ]; then
    echo "⚠️  Warning: Available memory is ${AVAILABLE_MEM}MB. Recommended: 4GB+"
fi

# Check file descriptor limits
FD_LIMIT=$(ulimit -n)
if [ "$FD_LIMIT" -lt 65536 ]; then
    echo "⚠️  Warning: File descriptor limit is $FD_LIMIT. Recommended: 65536+"
fi

# Check port availability
if netstat -ln | grep -q ":${SERVER_PORT:-9022} "; then
    echo "❌ Port ${SERVER_PORT:-9022} is already in use!"
    exit 1
fi

# Enable profiling if requested
if [ "${ENABLE_PPROF:-false}" = "true" ]; then
    echo "📊 Profiling enabled on port ${PPROF_PORT:-6060}"
    export ENABLE_PPROF=true
fi

# Print startup configuration
echo "📊 High Performance Configuration:"
echo "   - Port: ${SERVER_PORT:-9022}"
echo "   - GOMAXPROCS: $(go env GOMAXPROCS 2>/dev/null || echo $GOMAXPROCS)"
echo "   - GOGC: $GOGC"
echo "   - Memory Limit: $GOMEMLIMIT"
echo "   - File Descriptors: $FD_LIMIT"
echo "   - DB Max Connections: ${DB_MAX_OPEN_CONNS:-2000}"
echo "   - Profiling: ${ENABLE_PPROF:-false}"

# Function to handle cleanup on exit
cleanup() {
    echo "🛑 Shutting down gracefully..."
    if [ ! -z "$SERVER_PID" ]; then
        kill -TERM "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    echo "✅ Shutdown complete"
}

# Set trap for graceful shutdown
trap cleanup SIGTERM SIGINT EXIT

# Start the server
echo "🏁 Starting server..."
cd "$PROJECT_ROOT"

# Run with profiling if enabled
if [ "${ENABLE_PPROF:-false}" = "true" ]; then
    # Enable pprof endpoints
    "$BINARY_PATH" &
    SERVER_PID=$!
    
    echo "📊 Profiling endpoints available:"
    echo "   - http://localhost:${PPROF_PORT:-6060}/debug/pprof/"
    echo "   - http://localhost:${PPROF_PORT:-6060}/debug/pprof/goroutine"
    echo "   - http://localhost:${PPROF_PORT:-6060}/debug/pprof/heap"
    echo "   - http://localhost:${PPROF_PORT:-6060}/debug/pprof/profile"
else
    "$BINARY_PATH" &
    SERVER_PID=$!
fi

echo "✅ Server started with PID: $SERVER_PID"
echo "🌐 Service available at http://localhost:${SERVER_PORT:-9022}"
echo "❤️  Health check: http://localhost:${SERVER_PORT:-9022}/serviceability/health"

# Wait for server to finish
wait "$SERVER_PID" 