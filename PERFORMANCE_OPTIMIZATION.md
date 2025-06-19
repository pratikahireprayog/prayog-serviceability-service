# Maximum Performance Configuration Guide

This document outlines all optimizations applied to maximize concurrent request handling capacity.

## Applied Optimizations

### 1. Database Connection Pool (UNLIMITED)

- **MaxOpenConns**: `0` (unlimited connections)
- **MaxIdleConns**: `1000` (high idle pool)
- **Expected Impact**: Removes database connection bottleneck

### 2. Fiber Server Configuration (1 MILLION CONCURRENT)

- **Concurrency**: `1,048,576` (1 million concurrent connections)
- **ReadTimeout**: `5s` (aggressive timeout)
- **WriteTimeout**: `5s` (aggressive timeout)
- **IdleTimeout**: `30s` (quick recycling)
- **BodyLimit**: `50MB` (large request support)
- **ReadBufferSize**: `16KB`
- **WriteBufferSize**: `16KB`
- **Performance Flags**: All disabled for maximum speed

### 3. Rate Limiting (PRACTICALLY UNLIMITED)

- **Standard API**: `1,000,000 requests/minute`
- **Bulk API**: `1,000,000 requests/minute`
- **Request Size Limit**: DISABLED
- **Content Type Validation**: DISABLED

### 4. HTTP Client Connection Pool (MAXIMUM)

- **MaxIdleConns**: `10,000` (10x increase)
- **MaxIdleConnsPerHost**: `1,000` (100x increase)
- **MaxConnsPerHost**: `0` (unlimited)
- **HTTP/2**: Enabled for better performance

## Expected Performance Capacity

### Conservative Estimates (Safe Operation)

- **Concurrent Connections**: 50,000-100,000
- **Requests per Second**: 10,000-50,000
- **Database Operations**: Limited only by PostgreSQL performance

### Aggressive Estimates (Maximum Load)

- **Concurrent Connections**: 500,000-1,000,000
- **Requests per Second**: 100,000-500,000
- **Memory Usage**: 8-16GB RAM
- **CPU Usage**: 16-32 cores fully utilized

## Required Environment Variables for Maximum Performance

```bash
# Database Configuration
export DB_MAX_OPEN_CONNS=0
export DB_MAX_IDLE_CONNS=1000
export DB_CONN_MAX_LIFETIME=300s

# Server Configuration
export SERVER_READ_TIMEOUT=5s
export SERVER_WRITE_TIMEOUT=5s
export SERVER_IDLE_TIMEOUT=30s

# Logging (minimal for performance)
export LOG_LEVEL=warn
```

## Required Operating System Optimizations

### Linux System Limits

```bash
# Increase file descriptor limits
ulimit -n 1000000

# Increase network connection limits
echo 65535 > /proc/sys/net/core/somaxconn
echo 65535 > /proc/sys/net/core/netdev_max_backlog

# TCP optimizations
echo 1 > /proc/sys/net/ipv4/tcp_tw_reuse
echo 15 > /proc/sys/net/ipv4/tcp_fin_timeout
echo 1 > /proc/sys/net/ipv4/tcp_tw_recycle

# Memory optimizations
echo 1 > /proc/sys/vm/overcommit_memory
```

### Go Runtime Optimizations

```bash
export GOGC=100          # Aggressive garbage collection
export GOMAXPROCS=0      # Use all available CPU cores
```

## Hardware Recommendations for Maximum Load

### Minimum Requirements

- **CPU**: 8 cores, 3.0GHz+
- **RAM**: 16GB
- **Network**: 10Gbps
- **Storage**: NVMe SSD

### Recommended for Maximum Performance

- **CPU**: 32+ cores, 3.5GHz+
- **RAM**: 64GB+
- **Network**: 25Gbps+
- **Storage**: High-performance NVMe SSD array

## Load Testing Commands

### Basic Load Test

```bash
# Test serviceability endpoint
ab -n 100000 -c 1000 http://localhost:9022/serviceability/v1/check/12345

# Test with Apache Bench
ab -n 1000000 -c 10000 -k http://localhost:9022/serviceability/v1/ping
```

### Advanced Load Testing with wrk

```bash
# Install wrk: sudo apt install wrk

# High concurrency test
wrk -t32 -c10000 -d30s --latency http://localhost:9022/serviceability/v1/ping

# Sustained load test
wrk -t64 -c50000 -d300s --latency http://localhost:9022/serviceability/v1/check/12345
```

## Monitoring During Load Tests

### Key Metrics to Monitor

1. **Connection Count**: `netstat -an | grep :9022 | wc -l`
2. **Memory Usage**: `free -h`
3. **CPU Usage**: `top` or `htop`
4. **Database Connections**: Check PostgreSQL pg_stat_activity
5. **File Descriptors**: `lsof -p <process_id> | wc -l`

### Warning Signs

- Memory usage > 90%
- CPU usage sustained at 100%
- Database connection pool exhaustion
- File descriptor limit reached
- Network interface saturation

## Safety Notes

⚠️ **WARNING**: These optimizations remove safety limits and are intended for load testing only.

1. **Database**: Unlimited connections can overwhelm PostgreSQL
2. **Memory**: High concurrency can cause OOM kills
3. **CPU**: Maximum load can make system unresponsive
4. **Network**: Can saturate network interfaces

## Rollback Configuration

To restore safe defaults, revert the following files:

- `internal/shared/config/app_config.go`
- `internal/infrastructure/api/http/server.go`
- `internal/infrastructure/api/http/v1/middleware/serviceability_middleware.go`
- `internal/services/v1/partner_http_client.go`

## Testing Your APIs

### API Endpoints Optimized

1. **Serviceability Check**: `GET /serviceability/v1/check/{postal_code}`
2. **Bulk Serviceability**: `POST /serviceability/v1/bulk-check`
3. **Location Management**: Various endpoints under `/serviceability/v1/`

### Sample Load Test

```bash
# Test single postal code check
curl -X GET http://localhost:9022/serviceability/v1/check/12345

# Test bulk check
curl -X POST http://localhost:9022/serviceability/v1/bulk-check \
  -H "Content-Type: application/json" \
  -d '{"postal_codes": ["12345", "67890", "11111"]}'
```

Start with lower concurrency and gradually increase to find your system's maximum capacity!
