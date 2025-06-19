#!/bin/bash

# System optimization script for high concurrent connections
echo "🚀 Optimizing system for high concurrent connections..."

# Increase file descriptor limits
echo "📁 Setting file descriptor limits..."
sudo sysctl -w fs.file-max=2097152
echo "fs.file-max = 2097152" | sudo tee -a /etc/sysctl.conf

# Increase socket limits
echo "🔌 Optimizing socket settings..."
sudo sysctl -w net.core.somaxconn=65535
sudo sysctl -w net.core.netdev_max_backlog=5000
sudo sysctl -w net.core.rmem_default=262144
sudo sysctl -w net.core.rmem_max=268435456
sudo sysctl -w net.core.wmem_default=262144
sudo sysctl -w net.core.wmem_max=268435456

# TCP optimizations
echo "📡 Optimizing TCP settings..."
sudo sysctl -w net.ipv4.tcp_rmem="4096 65536 134217728"
sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 134217728"
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr
sudo sysctl -w net.ipv4.tcp_slow_start_after_idle=0
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_fin_timeout=30
sudo sysctl -w net.ipv4.tcp_keepalive_time=300
sudo sysctl -w net.ipv4.tcp_keepalive_probes=3
sudo sysctl -w net.ipv4.tcp_keepalive_intvl=30
sudo sysctl -w net.ipv4.tcp_max_syn_backlog=8192

# Memory optimizations
echo "💾 Optimizing memory settings..."
sudo sysctl -w vm.swappiness=10
sudo sysctl -w vm.dirty_ratio=15
sudo sysctl -w vm.dirty_background_ratio=5

# Apply all changes
sudo sysctl -p

# Set ulimits for current session
echo "⚙️ Setting ulimits..."
ulimit -n 1048576
ulimit -u 65536

# Create systemd service limits file
sudo mkdir -p /etc/systemd/system.conf.d/
cat << EOF | sudo tee /etc/systemd/system.conf.d/limits.conf
[Manager]
DefaultLimitNOFILE=1048576
DefaultLimitNPROC=65536
EOF

echo "✅ System optimization complete!"
echo "📝 Remember to restart your terminal/service to apply ulimit changes"
echo "🔄 For permanent changes, add ulimits to /etc/security/limits.conf" 