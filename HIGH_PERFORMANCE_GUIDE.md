# High Performance Guide for Prayog Serviceability Service

This guide will help you configure and run your Go Fiber service to handle **80,000+ concurrent API calls** with maximum performance.

## 🚀 Quick Start

1. **Optimize your system** (run once):

   ```bash
   sudo ./scripts/optimize_system.sh
   ```

2. **Start the service in high-performance mode**:

   ```bash
   ./scripts/run_high_performance.sh
   ```

3. **Test the performance**:
   ```bash
   ./scripts/load_test.sh
   ```

## 📋 System Requirements

### Minimum Requirements

- **CPU**: 4+ cores
- **RAM**: 8GB+ (16GB recommended)
- **Network**: 1Gbps+
- **OS**: Linux (Ubuntu 20.04+ or CentOS 8+)

### Recommended for 80K+ Concurrent Requests

- **CPU**: 16+ cores
- **RAM**: 32GB+
- **Network**: 10Gbps+
- **SSD**: For database and logging

## ⚙️ Configuration Options

### Environment Variables

Copy and modify `scripts/high-performance.env`:

```bash
# Critical performance settings
GOGC=400                    # Reduce GC frequency
GOMEMLIMIT=16GiB           # Set based on available RAM
GOMAXPROCS=0               # Use all CPU cores

# Database connection pool
DB_MAX_OPEN_CONNS=2000     # Max concurrent DB connections
DB_MAX_IDLE_CONNS=1000     # Keep connections ready

# Server timeouts (aggressive for max throughput)
SERVER_READ_TIMEOUT=2s
SERVER_WRITE_TIMEOUT=2s
SERVER_IDLE_TIMEOUT=15s
```

### System Limits

The optimization script sets these limits:

- **File descriptors**: 1,048,576
- **Max processes**: 65,536
- **TCP socket buffer**: 268MB
- **Connection backlog**: 65,535

## 🔧 Performance Tuning

### 1. Database Optimization

```bash
# PostgreSQL configuration (add to postgresql.conf)
max_connections = 2000
shared_buffers = 8GB              # 25% of RAM
effective_cache_size = 24GB       # 75% of RAM
work_mem = 256MB
maintenance_work_mem = 2GB
checkpoint_completion_target = 0.9
wal_buffers = 64MB
default_statistics_target = 100
```

### 2. Application-Level Optimizations

Your service already includes these optimizations:

- **2M concurrent connections** supported
- **32KB buffers** for read/write operations
- **Aggressive timeouts** (2-5 seconds)
- **Disabled overhead features** (headers, startup messages)
- **Connection pooling** with keep-alive

### 3. Go Runtime Optimizations

```bash
# Set these environment variables
export GOGC=400              # Less frequent GC
export GOMEMLIMIT=16GiB      # Memory limit
export GOMAXPROCS=0          # Use all CPUs
```

## 📊 Monitoring & Profiling

### Enable Profiling

Set in your environment:

```bash
ENABLE_PPROF=true
PPROF_PORT=6060
```

### Access Profiling Endpoints

Once running, access these URLs:

- **CPU Profile**: http://localhost:6060/debug/pprof/profile
- **Memory Profile**: http://localhost:6060/debug/pprof/heap
- **Goroutines**: http://localhost:6060/debug/pprof/goroutine
- **All Profiles**: http://localhost:6060/debug/pprof/

### Command-Line Profiling

```bash
# CPU profile for 30 seconds
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### System Monitoring

Monitor these metrics during load testing:

```bash
# CPU and memory usage
htop

# Network connections
ss -tuln | grep :9022

# Network traffic
iftop

# I/O performance
iostat -x 1

# Real-time performance
vmstat 1
```

## 🧪 Load Testing

### Basic Load Test

```bash
# Test with default settings (50K requests)
./scripts/load_test.sh
```

### Custom Load Test

```bash
# Test 80K requests with 2K concurrent connections
TARGET_URL=http://localhost:9022 \
EXTREME_REQUESTS=80000 \
EXTREME_CONCURRENT=2000 \
./scripts/load_test.sh
```

### Using Apache Bench Directly

```bash
# 80K requests, 2K concurrent
ab -n 80000 -c 2000 -k http://localhost:9022/serviceability/v1/status
```

### Using wrk (if installed)

```bash
# Install wrk first
sudo apt-get install wrk  # Ubuntu
brew install wrk          # macOS

# Run load test
wrk -t12 -c2000 -d300s --latency http://localhost:9022/serviceability/v1/status
```

## 🎯 Expected Performance

With proper configuration, you should achieve:

- **Requests/Second**: 50,000-100,000+ RPS
- **Latency**: <10ms p99 for simple endpoints
- **Memory Usage**: 2-8GB (depending on load)
- **CPU Usage**: 70-90% under full load

## 🔍 Troubleshooting

### Common Issues

1. **"Too many open files" error**

   ```bash
   # Check current limit
   ulimit -n

   # Run system optimization script
   sudo ./scripts/optimize_system.sh
   ```

2. **High memory usage**

   ```bash
   # Reduce GOGC value
   export GOGC=200

   # Set memory limit
   export GOMEMLIMIT=8GiB
   ```

3. **Database connection errors**

   ```bash
   # Increase connection pool
   export DB_MAX_OPEN_CONNS=3000
   export DB_MAX_IDLE_CONNS=1500
   ```

4. **High CPU usage**
   ```bash
   # Enable profiling and check for hotspots
   go tool pprof http://localhost:6060/debug/pprof/profile
   ```

### Performance Debugging

```bash
# Check system resources
free -h                    # Memory usage
df -h                      # Disk usage
lscpu                      # CPU info

# Check network limits
sysctl net.core.somaxconn  # Connection backlog
sysctl net.core.rmem_max   # Receive buffer size

# Check file descriptor limits
cat /proc/sys/fs/file-max
ulimit -n
```

## 🏗️ Scaling Beyond Single Instance

### Horizontal Scaling

For loads beyond what a single instance can handle:

1. **Load Balancer** (nginx/HAProxy)
2. **Multiple service instances**
3. **Database read replicas**
4. **Redis caching layer**

### Example nginx configuration:

```nginx
upstream serviceability {
    server 127.0.0.1:9022;
    server 127.0.0.1:9023;
    server 127.0.0.1:9024;
    server 127.0.0.1:9025;
}

server {
    listen 80;
    location / {
        proxy_pass http://serviceability;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🛡️ Production Considerations

1. **Use a reverse proxy** (nginx, HAProxy)
2. **Enable compression** (gzip)
3. **Implement rate limiting**
4. **Set up monitoring** (Prometheus, Grafana)
5. **Configure log rotation**
6. **Use connection pooling** for external services
7. **Implement circuit breakers** for external dependencies

## 📚 Additional Resources

- [Go Performance Tuning](https://golang.org/doc/diagnostics)
- [Fiber Performance Guide](https://docs.gofiber.io/guide/performance)
- [Linux Performance Tools](http://www.brendangregg.com/linuxperf.html)
- [Database Connection Pooling](https://pkg.go.dev/database/sql#DB.SetMaxOpenConns)

---

**Happy high-performance serving! 🚀**
