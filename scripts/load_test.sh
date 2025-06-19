#!/bin/bash

# Load Testing Script for Prayog Serviceability Service
# Tests the service under high concurrent load

set -e

# Configuration
TARGET_URL="${TARGET_URL:-http://localhost:9022}"
HEALTH_ENDPOINT="$TARGET_URL/serviceability/health"
API_ENDPOINT="$TARGET_URL/serviceability/v1/status"

# Test parameters
CONCURRENT_USERS="${CONCURRENT_USERS:-1000}"
REQUESTS_PER_USER="${REQUESTS_PER_USER:-100}"
RAMP_UP_TIME="${RAMP_UP_TIME:-30s}"
TEST_DURATION="${TEST_DURATION:-300s}"

echo "🧪 Load Testing Prayog Serviceability Service"
echo "================================================"
echo "Target URL: $TARGET_URL"
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Requests per User: $REQUESTS_PER_USER"
echo "Ramp-up Time: $RAMP_UP_TIME"
echo "Test Duration: $TEST_DURATION"
echo "================================================"

# Check if required tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        echo "   Ubuntu/Debian: sudo apt-get install $1"
        echo "   macOS: brew install $1"
        exit 1
    fi
}

echo "🔍 Checking required tools..."
check_tool curl
check_tool ab  # Apache Bench

# Wait for service to be ready
echo "⏳ Waiting for service to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s -f "$HEALTH_ENDPOINT" > /dev/null 2>&1; then
        echo "✅ Service is ready!"
        break
    fi
    
    attempt=$((attempt + 1))
    echo "   Attempt $attempt/$max_attempts - Service not ready yet..."
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ Service failed to start within $(($max_attempts * 2)) seconds"
    exit 1
fi

# Function to run Apache Bench test
run_ab_test() {
    local name="$1"
    local url="$2"
    local concurrent="$3"
    local requests="$4"
    
    echo ""
    echo "🚀 Running $name..."
    echo "   URL: $url"
    echo "   Concurrent: $concurrent"
    echo "   Total Requests: $requests"
    
    ab -n $requests -c $concurrent -k -H "Accept: application/json" "$url" 2>&1 | \
    grep -E "(Requests per second|Time per request|Transfer rate|Connection Times|Percentage of the requests)" || true
}

# Basic performance tests
echo ""
echo "🏃‍♂️ Starting Load Tests..."

# Test 1: Health endpoint (lightweight)
run_ab_test "Health Endpoint Test" "$HEALTH_ENDPOINT" 100 10000

# Test 2: API endpoint (moderate load)
run_ab_test "API Endpoint - Moderate Load" "$API_ENDPOINT" 500 25000

# Test 3: API endpoint (high load)
run_ab_test "API Endpoint - High Load" "$API_ENDPOINT" 1000 50000

# Test 4: API endpoint (extreme load) - This is your 80K test
EXTREME_CONCURRENT=2000
EXTREME_REQUESTS=80000

echo ""
echo "🔥 EXTREME LOAD TEST - This is the big one!"
echo "   Simulating $EXTREME_REQUESTS requests with $EXTREME_CONCURRENT concurrent connections"
echo "   This test represents your 80K+ concurrent scenario"

run_ab_test "API Endpoint - EXTREME LOAD (80K)" "$API_ENDPOINT" $EXTREME_CONCURRENT $EXTREME_REQUESTS

# Function to run custom curl-based concurrent test
run_concurrent_curl_test() {
    local name="$1"
    local url="$2"
    local concurrent="$3"
    local total_requests="$4"
    
    echo ""
    echo "🔄 Running $name with curl..."
    
    requests_per_worker=$((total_requests / concurrent))
    
    # Create temporary directory for results
    temp_dir=$(mktemp -d)
    
    # Start concurrent workers
    for i in $(seq 1 $concurrent); do
        {
            start_time=$(date +%s.%N)
            success_count=0
            error_count=0
            
            for j in $(seq 1 $requests_per_worker); do
                if curl -s -f "$url" > /dev/null 2>&1; then
                    success_count=$((success_count + 1))
                else
                    error_count=$((error_count + 1))
                fi
            done
            
            end_time=$(date +%s.%N)
            duration=$(echo "$end_time - $start_time" | bc -l)
            
            echo "$i,$success_count,$error_count,$duration" >> "$temp_dir/results.csv"
        } &
    done
    
    # Wait for all workers to complete
    wait
    
    # Calculate results
    total_success=0
    total_errors=0
    total_time=0
    worker_count=0
    
    while IFS=',' read -r worker success errors duration; do
        total_success=$((total_success + success))
        total_errors=$((total_errors + errors))
        total_time=$(echo "$total_time + $duration" | bc -l)
        worker_count=$((worker_count + 1))
    done < "$temp_dir/results.csv"
    
    avg_time=$(echo "scale=3; $total_time / $worker_count" | bc -l)
    total_requests_actual=$((total_success + total_errors))
    success_rate=$(echo "scale=2; $total_success * 100 / $total_requests_actual" | bc -l)
    rps=$(echo "scale=2; $total_success / $avg_time" | bc -l)
    
    echo "   Results:"
    echo "   - Total Requests: $total_requests_actual"
    echo "   - Successful: $total_success"
    echo "   - Failed: $total_errors"
    echo "   - Success Rate: $success_rate%"
    echo "   - Average Duration: ${avg_time}s"
    echo "   - Requests/Second: $rps"
    
    # Cleanup
    rm -rf "$temp_dir"
}

# Additional concurrent test with curl
run_concurrent_curl_test "Concurrent Curl Test" "$API_ENDPOINT" 1000 50000

echo ""
echo "📊 Load Testing Complete!"
echo ""
echo "💡 Performance Tips:"
echo "   - Monitor CPU, memory, and network during tests"
echo "   - Check error logs for any issues"
echo "   - Use 'htop' or 'top' to monitor system resources"
echo "   - Use 'ss -tuln' to check connection states"
echo "   - Profile with pprof if enabled: http://localhost:6060/debug/pprof/"
echo ""
echo "🔧 If performance is not meeting expectations:"
echo "   1. Run the system optimization script: ./scripts/optimize_system.sh"
echo "   2. Increase database connection pool size"
echo "   3. Enable profiling and analyze bottlenecks"
echo "   4. Consider horizontal scaling (multiple instances)"
echo "   5. Use a load balancer (nginx, HAProxy, etc.)" 